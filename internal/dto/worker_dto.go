package dto

import (
	"time"

	"github.com/google/uuid"
)

type WorkerReviewResponse struct {
	ID         uuid.UUID `json:"id"`
	AuthorName string    `json:"author"`
	Rating     float64   `json:"rating"`
	Comment    string    `json:"comment"`
	CreatedAt  time.Time `json:"created_at"`
}

type WorkerResponse struct {
	ID               uuid.UUID              `json:"id"`
	WorkerCode       string                 `json:"worker_code"`
	Name             string                 `json:"name"`
	AvatarURL        string                 `json:"avatar_url"`
	Age              int                    `json:"age"`
	ExperienceYears  int                    `json:"experience_years"`
	Rating           float64                `json:"rating"`
	ReviewsCount     int                    `json:"reviews_count"`
	ServiceTypes     []string               `json:"service_types"`
	Specializations  []string               `json:"specializations"`
	Phone            string                 `json:"phone"`
	Email            string                 `json:"email"`
	Bio              string                 `json:"bio"`
	CompletedJobs    int                    `json:"completed_jobs"`
	HourlyRate       string                 `json:"hourly_rate"`
	RateValue        float64                `json:"rate_value"`
	IsVerified       bool                   `json:"is_verified"`
	IsAvailableToday bool                   `json:"is_available_today"`
	Badges           []string               `json:"badges"`
	Reviews          []WorkerReviewResponse `json:"reviews,omitempty"`
}

type CreateWorkerReviewRequest struct {
	Rating  float64 `json:"rating" binding:"required,min=1,max=5"`
	Comment string  `json:"comment" binding:"required,min=3"`
}

type CreateWorkerRequest struct {
	WorkerCode       string   `json:"worker_code"`
	Name             string   `json:"name" binding:"required"`
	AvatarURL        string   `json:"avatar_url"`
	Age              int      `json:"age" binding:"required,min=18,max=99"`
	ExperienceYears  int      `json:"experience_years" binding:"min=0"`
	ServiceTypes     []string `json:"service_types" binding:"required"`
	Specializations  []string `json:"specializations"`
	Phone            string   `json:"phone" binding:"required"`
	Email            string   `json:"email" binding:"required,email"`
	Bio              string   `json:"bio"`
	HourlyRate       string   `json:"hourly_rate"`
	RateValue        float64  `json:"rate_value"`
	IsVerified       *bool    `json:"is_verified"`
	IsAvailableToday *bool    `json:"is_available_today"`
	Badges           []string `json:"badges"`
}

type UpdateWorkerRequest struct {
	Name             *string   `json:"name"`
	AvatarURL        *string   `json:"avatar_url"`
	Age              *int      `json:"age"`
	ExperienceYears  *int      `json:"experience_years"`
	ServiceTypes     *[]string `json:"service_types"`
	Specializations  *[]string `json:"specializations"`
	Phone            *string   `json:"phone"`
	Email            *string   `json:"email"`
	Bio              *string   `json:"bio"`
	CompletedJobs    *int      `json:"completed_jobs"`
	HourlyRate       *string   `json:"hourly_rate"`
	RateValue        *float64  `json:"rate_value"`
	IsVerified       *bool     `json:"is_verified"`
	IsAvailableToday *bool     `json:"is_available_today"`
	Badges           *[]string `json:"badges"`
}
