package service

import (
	"errors"
	"fmt"
	"strings"

	"delivery-backend/internal/config"
	"delivery-backend/internal/dto"
	"delivery-backend/internal/models"
	"delivery-backend/internal/repository"
	"delivery-backend/internal/utils"

	"github.com/google/uuid"
)

type AuthService interface {
	Register(req *dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(req *dto.LoginRequest) (*dto.AuthResponse, error)
	GoogleLogin(req *dto.GoogleLoginRequest) (*dto.AuthResponse, error)
	ForgotPassword(req *dto.ForgotPasswordRequest) error
	GetProfile(userID uuid.UUID) (*dto.UserResponse, error)
	UpdateProfile(userID uuid.UUID, req *dto.UpdateProfileRequest) (*dto.UserResponse, error)
	ChangePassword(userID uuid.UUID, req *dto.ChangePasswordRequest) error
	GetAllCustomers(role string) ([]dto.UserResponse, error)
	GetCustomerByID(id uuid.UUID) (*dto.UserResponse, error)
	CreateCustomer(req *dto.CreateCustomerRequest) (*dto.UserResponse, error)
	UpdateCustomer(id uuid.UUID, req *dto.UpdateCustomerRequest) (*dto.UserResponse, error)
	ToggleCustomerActive(id uuid.UUID, isActive bool) error
	DeleteCustomer(id uuid.UUID) error
}

type authService struct {
	userRepo   repository.UserRepository
	workerRepo repository.WorkerRepository
	emailSvc   EmailService
	cfg        *config.Config
}

func NewAuthService(userRepo repository.UserRepository, workerRepo repository.WorkerRepository, emailSvc EmailService, cfg *config.Config) AuthService {
	return &authService{
		userRepo:   userRepo,
		workerRepo: workerRepo,
		emailSvc:   emailSvc,
		cfg:        cfg,
	}
}

func (s *authService) Register(req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	// Check if email or phone already registered
	if existing, _ := s.userRepo.FindByEmail(req.Email); existing != nil {
		return nil, errors.New("email is already registered")
	}
	if existing, _ := s.userRepo.FindByPhone(req.Phone); existing != nil {
		return nil, errors.New("phone number is already registered")
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	avatarURL := req.AvatarURL
	if avatarURL == "" {
		avatarURL = "https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=200&auto=format&fit=crop&q=80"
	}

	user := models.User{
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: hashedPassword,
		AvatarURL:    avatarURL,
		Role:         "customer",
		IsActive:     true,
	}

	if err := s.userRepo.Create(&user); err != nil {
		return nil, err
	}

	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role, s.cfg.JWTSecret, s.cfg.JWTExpiresIn)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Phone:     user.Phone,
			AvatarURL: user.AvatarURL,
			Role:      user.Role,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

func (s *authService) Login(req *dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.FindByEmailOrPhone(req.PhoneOrEmail)
	if err != nil || user == nil {
		return nil, errors.New("invalid email/phone or password")
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid email/phone or password")
	}

	if !user.IsActive {
		return nil, errors.New("user account is deactivated")
	}

	// Any registered user (customer, worker, admin) is permitted to access the platform.

	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role, s.cfg.JWTSecret, s.cfg.JWTExpiresIn)
	if err != nil {
		return nil, err
	}

	var workerProfile *dto.WorkerResponse
	if req.Role == "worker" || user.Role == "worker" || user.Role == "admin" {
		if s.workerRepo != nil {
			worker, err := s.workerRepo.FindByUserID(user.ID)
			if err != nil || worker == nil {
				worker, err = s.workerRepo.FindByEmail(user.Email)
			}
			if worker == nil && req.Role == "worker" {
				// Auto-provision worker profile for this registered user
				newWorker := models.Worker{
					UserID:           &user.ID,
					WorkerCode:       fmt.Sprintf("wrk_%s", user.ID.String()[:8]),
					Name:             user.Name,
					Email:            user.Email,
					Phone:            user.Phone,
					AvatarURL:        user.AvatarURL,
					Age:              30,
					ExperienceYears:  5,
					Rating:           5.0,
					ReviewsCount:     0,
					ServiceTypes:     []string{"all", "other", "plumbing", "electricity", "cleaning", "building_work", "carpenter", "cook", "care_taking"},
					Specializations:  []string{"General Specialist", "All-Round Technician"},
					HourlyRate:       "Rs. 1,500 / hr",
					RateValue:        1500.0,
					Latitude:         6.9271,
					Longitude:        79.8612,
					OperatingAddress: "Colombo, Sri Lanka",
					IsOpenForWork:    true,
					RadiusKm:         20.0,
					IsVerified:       true,
					IsActive:         true,
					IsAvailableToday: true,
					Badges:           []string{"Verified Pro", "Active Partner"},
				}
				if err := s.workerRepo.Create(&newWorker); err == nil {
					worker = &newWorker
				}
			}
			if worker != nil {
				w := mapWorkerToDTO(worker)
				workerProfile = &w
			}
		}
	}

	userResp := dto.UserResponse{
		ID:            user.ID,
		Name:          user.Name,
		Email:         user.Email,
		Phone:         user.Phone,
		AvatarURL:     user.AvatarURL,
		Role:          user.Role,
		IsActive:      user.IsActive,
		WorkerProfile: workerProfile,
		CreatedAt:     user.CreatedAt,
	}

	return &dto.AuthResponse{
		Token:         token,
		User:          userResp,
		WorkerProfile: workerProfile,
	}, nil
}

func (s *authService) GoogleLogin(req *dto.GoogleLoginRequest) (*dto.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" {
		return nil, errors.New("email is required for Google login")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		parts := strings.Split(email, "@")
		name = parts[0]
	}

	avatarURL := req.AvatarURL
	if avatarURL == "" {
		avatarURL = "https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=200&auto=format&fit=crop&q=80"
	}

	targetRole := "customer"
	if req.Role == "worker" {
		targetRole = "worker"
	}

	// 1. Find or create user by email
	user, err := s.userRepo.FindByEmail(email)
	if err != nil || user == nil {
		randPass := uuid.New().String()
		hashedPassword, err := utils.HashPassword(randPass)
		if err != nil {
			return nil, errors.New("failed to secure account")
		}

		placeholderPhone := "+94 77 " + strings.ReplaceAll(uuid.New().String()[:7], "-", "9")

		newUser := models.User{
			Name:         name,
			Email:        email,
			Phone:        placeholderPhone,
			PasswordHash: hashedPassword,
			AvatarURL:    avatarURL,
			Role:         targetRole,
			IsActive:     true,
		}

		if err := s.userRepo.Create(&newUser); err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
		user = &newUser
	} else {
		if !user.IsActive {
			return nil, errors.New("user account is deactivated")
		}

		// Update name/avatar if empty or placeholder
		updates := make(map[string]interface{})
		if user.AvatarURL == "" && avatarURL != "" {
			updates["avatar_url"] = avatarURL
			user.AvatarURL = avatarURL
		}
		if (user.Name == "" || strings.Contains(user.Name, "@")) && name != "" {
			updates["name"] = name
			user.Name = name
		}
		if len(updates) > 0 {
			_ = s.userRepo.Update(user)
		}
	}

	// 2. Generate JWT
	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role, s.cfg.JWTSecret, s.cfg.JWTExpiresIn)
	if err != nil {
		return nil, err
	}

	// 3. Resolve Worker Profile if needed
	var workerProfile *dto.WorkerResponse
	if targetRole == "worker" || user.Role == "worker" {
		worker, _ := s.workerRepo.FindByUserID(user.ID)
		if worker == nil {
			worker, _ = s.workerRepo.FindByEmail(user.Email)
		}
		if worker == nil {
			newWorker := models.Worker{
				UserID:           &user.ID,
				WorkerCode:       "wrk_" + strings.ToLower(uuid.New().String()[:6]),
				Name:             user.Name,
				AvatarURL:        user.AvatarURL,
				Email:            user.Email,
				Phone:            user.Phone,
				ServiceTypes:     []string{"plumbing", "electricity", "carpenter", "cleaning", "other"},
				Specializations:  []string{"General Maintenance", "Emergency Repairs", "Certified Technician"},
				HourlyRate:       "Rs. 1,800 / visit",
				RateValue:        1800.0,
				Latitude:         6.9080,
				Longitude:        79.8510,
				OperatingAddress: "Galle Road, Kollupitiya, Colombo 03",
				RadiusKm:         10.0,
				IsVerified:       true,
				IsActive:         true,
				IsAvailableToday: true,
				Badges:           []string{"Verified Pro", "Google Sign-In Partner"},
			}
			if err := s.workerRepo.Create(&newWorker); err == nil {
				worker = &newWorker
			}
		}
		if worker != nil {
			w := mapWorkerToDTO(worker)
			workerProfile = &w
		}
	}

	userResp := dto.UserResponse{
		ID:            user.ID,
		Name:          user.Name,
		Email:         user.Email,
		Phone:         user.Phone,
		AvatarURL:     user.AvatarURL,
		Role:          user.Role,
		IsActive:      user.IsActive,
		WorkerProfile: workerProfile,
		CreatedAt:     user.CreatedAt,
	}

	return &dto.AuthResponse{
		Token:         token,
		User:          userResp,
		WorkerProfile: workerProfile,
	}, nil
}

func (s *authService) ForgotPassword(req *dto.ForgotPasswordRequest) error {
	var user *models.User
	var err error

	if strings.Contains(req.PhoneOrEmail, "@") {
		user, err = s.userRepo.FindByEmail(req.PhoneOrEmail)
	} else {
		user, err = s.userRepo.FindByPhone(req.PhoneOrEmail)
	}

	if err != nil || user == nil {
		return errors.New("no account found matching that email or phone number")
	}

	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("failed to process password")
	}

	if err := s.userRepo.UpdatePassword(user.ID, newHash); err != nil {
		return err
	}

	// Send password reset security notification
	if s.emailSvc != nil {
		s.emailSvc.SendPasswordResetConfirmation(user)
	}

	return nil
}

func (s *authService) GetProfile(userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	var workerProfile *dto.WorkerResponse
	if user.Role == "worker" || user.Role == "admin" {
		if s.workerRepo != nil {
			if worker, err := s.workerRepo.FindByUserID(user.ID); err == nil && worker != nil {
				w := mapWorkerToDTO(worker)
				workerProfile = &w
			} else if worker, err := s.workerRepo.FindByEmail(user.Email); err == nil && worker != nil {
				w := mapWorkerToDTO(worker)
				workerProfile = &w
			}
		}
	}

	return &dto.UserResponse{
		ID:            user.ID,
		Name:          user.Name,
		Email:         user.Email,
		Phone:         user.Phone,
		AvatarURL:     user.AvatarURL,
		Role:          user.Role,
		IsActive:      user.IsActive,
		WorkerProfile: workerProfile,
		CreatedAt:     user.CreatedAt,
	}, nil
}

func (s *authService) UpdateProfile(userID uuid.UUID, req *dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return &dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Phone:     user.Phone,
		AvatarURL: user.AvatarURL,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	}, nil
}

func (s *authService) ChangePassword(userID uuid.UUID, req *dto.ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if !utils.CheckPasswordHash(req.OldPassword, user.PasswordHash) {
		return errors.New("current password is incorrect")
	}

	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	return s.userRepo.UpdatePassword(userID, newHash)
}

func (s *authService) GetAllCustomers(role string) ([]dto.UserResponse, error) {
	users, err := s.userRepo.FindAll(role)
	if err != nil {
		return nil, err
	}

	res := make([]dto.UserResponse, 0, len(users))
	for _, u := range users {
		res = append(res, mapUserToDTO(&u))
	}
	return res, nil
}

func (s *authService) GetCustomerByID(id uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("customer not found")
	}
	dtoRes := mapUserToDTO(user)
	return &dtoRes, nil
}

func (s *authService) CreateCustomer(req *dto.CreateCustomerRequest) (*dto.UserResponse, error) {
	if existing, _ := s.userRepo.FindByEmail(req.Email); existing != nil {
		return nil, errors.New("email is already registered")
	}
	if existing, _ := s.userRepo.FindByPhone(req.Phone); existing != nil {
		return nil, errors.New("phone number is already registered")
	}

	rawPassword := req.Password
	if rawPassword == "" {
		rawPassword = "Password123!"
	}

	hashedPassword, err := utils.HashPassword(rawPassword)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	avatarURL := req.AvatarURL
	if avatarURL == "" {
		avatarURL = "https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=200&auto=format&fit=crop&q=80"
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	role := req.Role
	if role == "" {
		role = "customer"
	}

	user := models.User{
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: hashedPassword,
		AvatarURL:    avatarURL,
		Role:         role,
		IsActive:     isActive,
	}

	if err := s.userRepo.Create(&user); err != nil {
		return nil, err
	}

	dtoRes := mapUserToDTO(&user)
	return &dtoRes, nil
}

func (s *authService) UpdateCustomer(id uuid.UUID, req *dto.UpdateCustomerRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("customer not found")
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Email != nil && *req.Email != user.Email {
		if existing, _ := s.userRepo.FindByEmail(*req.Email); existing != nil {
			return nil, errors.New("email is already taken by another account")
		}
		user.Email = *req.Email
	}
	if req.Phone != nil && *req.Phone != user.Phone {
		if existing, _ := s.userRepo.FindByPhone(*req.Phone); existing != nil {
			return nil, errors.New("phone number is already taken by another account")
		}
		user.Phone = *req.Phone
	}
	if req.Password != nil && *req.Password != "" {
		hashed, err := utils.HashPassword(*req.Password)
		if err != nil {
			return nil, errors.New("failed to hash password")
		}
		user.PasswordHash = hashed
	}
	if req.AvatarURL != nil {
		user.AvatarURL = *req.AvatarURL
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	dtoRes := mapUserToDTO(user)
	return &dtoRes, nil
}

func (s *authService) ToggleCustomerActive(id uuid.UUID, isActive bool) error {
	return s.userRepo.ToggleActive(id, isActive)
}

func (s *authService) DeleteCustomer(id uuid.UUID) error {
	return s.userRepo.Delete(id)
}

func mapUserToDTO(u *models.User) dto.UserResponse {
	totalSpent := 0.0
	for _, b := range u.Bookings {
		totalSpent += b.FinalAmount
	}

	return dto.UserResponse{
		ID:             u.ID,
		Name:           u.Name,
		Email:          u.Email,
		Phone:          u.Phone,
		AvatarURL:      u.AvatarURL,
		Role:           u.Role,
		IsActive:       u.IsActive,
		BookingsCount:  len(u.Bookings),
		AddressesCount: len(u.Addresses),
		TotalSpent:     totalSpent,
		CreatedAt:      u.CreatedAt,
	}
}
