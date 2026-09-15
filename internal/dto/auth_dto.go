package dto

import (
	"time"

	"github.com/google/uuid"
)

type RegisterRequest struct {
	Name            string `json:"name" binding:"required,min=2,max=100"`
	Email           string `json:"email" binding:"required,email"`
	Phone           string `json:"phone" binding:"required,min=8,max=20"`
	Password        string `json:"password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
	AvatarURL       string `json:"avatar_url"`
}

type LoginRequest struct {
	PhoneOrEmail string `json:"phone_or_email" binding:"required"`
	Password     string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	AvatarURL      string    `json:"avatar_url"`
	Role           string    `json:"role"`
	IsActive       bool      `json:"is_active"`
	BookingsCount  int       `json:"bookings_count"`
	AddressesCount int       `json:"addresses_count"`
	TotalSpent     float64   `json:"total_spent"`
	CreatedAt      time.Time `json:"created_at"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UpdateProfileRequest struct {
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	AvatarURL string `json:"avatar_url"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type ForgotPasswordRequest struct {
	PhoneOrEmail    string `json:"phone_or_email" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}

type CreateCustomerRequest struct {
	Name      string `json:"name" binding:"required,min=2,max=100"`
	Email     string `json:"email" binding:"required,email"`
	Phone     string `json:"phone" binding:"required,min=8,max=20"`
	Password  string `json:"password"`
	AvatarURL string `json:"avatar_url"`
	IsActive  *bool  `json:"is_active"`
	Role      string `json:"role"`
}

type UpdateCustomerRequest struct {
	Name      *string `json:"name"`
	Email     *string `json:"email"`
	Phone     *string `json:"phone"`
	Password  *string `json:"password"`
	AvatarURL *string `json:"avatar_url"`
	IsActive  *bool   `json:"is_active"`
	Role      *string `json:"role"`
}

type ToggleCustomerStatusRequest struct {
	IsActive bool `json:"is_active"`
}
