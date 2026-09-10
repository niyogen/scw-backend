package service

import (
	"errors"

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
	GetProfile(userID uuid.UUID) (*dto.UserResponse, error)
	UpdateProfile(userID uuid.UUID, req *dto.UpdateProfileRequest) (*dto.UserResponse, error)
	ChangePassword(userID uuid.UUID, req *dto.ChangePasswordRequest) error
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

	user := models.User{
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: hashedPassword,
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
