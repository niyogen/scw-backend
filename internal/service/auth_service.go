package service

import (
	"errors"
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
	userRepo repository.UserRepository
	cfg      *config.Config
}

func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) AuthService {
	return &authService{
		userRepo: userRepo,
		cfg:      cfg,
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

	return s.userRepo.UpdatePassword(user.ID, newHash)
}

func (s *authService) GetProfile(userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
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
