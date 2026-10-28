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

	// DateOfBirth is required for PDPA signup (server computes age
	// and rejects <18). Format: ISO-8601 YYYY-MM-DD. The field is
	// required at the application layer, not the binding layer, so
	// existing legacy clients get a clearer error message
	// ("date_of_birth is required") rather than a generic 400.
	DateOfBirth string `json:"date_of_birth"`

	// Consents is the list of consent records captured at signup
	// (typically user_tc, user_privacy, age_confirmation). Required
	// at the application layer for the same reason.
	Consents []ConsentInput `json:"consents"`
}

type LoginRequest struct {
	PhoneOrEmail string `json:"phone_or_email" binding:"required"`
	Password     string `json:"password" binding:"required"`
	Role         string `json:"role"` // "customer" or "worker"
}

type UserResponse struct {
	ID             uuid.UUID       `json:"id"`
	Name           string          `json:"name"`
	Email          string          `json:"email"`
	Phone          string          `json:"phone"`
	AvatarURL      string          `json:"avatar_url"`
	Role           string          `json:"role"`
	IsActive       bool            `json:"is_active"`
	WorkerProfile  *WorkerResponse `json:"worker_profile,omitempty"`
	BookingsCount  int             `json:"bookings_count"`
	AddressesCount int             `json:"addresses_count"`
	TotalSpent     float64         `json:"total_spent"`
	CreatedAt      time.Time       `json:"created_at"`
}

type AuthResponse struct {
	Token         string          `json:"token"`
	User          UserResponse    `json:"user"`
	WorkerProfile *WorkerResponse `json:"worker_profile,omitempty"`
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

type GoogleLoginRequest struct {
	// IDToken is the Google ID token obtained from the native sign-in
	// flow (Flutter google_sign_in package). Required in production;
	// optional in non-prod where a mock fallback uses the email + name
	// fields below. See service.GoogleLogin for the resolution order.
	IDToken     string `json:"id_token"`
	AccessToken string `json:"access_token"`
	Email       string `json:"email"` // mock-fallback only; verified claims from IDToken take precedence
	Name        string `json:"name"`
	AvatarURL   string `json:"avatar_url"`
	Role        string `json:"role"` // "customer" or "worker"
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
