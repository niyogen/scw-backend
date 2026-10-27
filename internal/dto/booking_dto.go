package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateBookingRequest struct {
	WorkerID          *string                `json:"worker_id"`
	ServiceID         *string                `json:"service_id"`
	LocationTitle     string                 `json:"location_title" binding:"required"`
	LocationSubtitle  string                 `json:"location_subtitle" binding:"required"`
	Latitude          float64                `json:"latitude"`
	Longitude         float64                `json:"longitude"`
	SelectedItems     map[string]interface{} `json:"selected_items"`
	TotalCost         float64                `json:"total_cost" binding:"required"`
	VoucherCode       string                 `json:"voucher_code"`
	PaymentMethod     string                 `json:"payment_method"` // promptpay, credit_card, cash
	Notes             string                 `json:"notes"`
	JobTitle          string                 `json:"job_title"`
	JobDescription    string                 `json:"job_description"`
	JobPhotos         []string               `json:"job_photos"`
	ScheduledAtString string                 `json:"scheduled_at"` // ISO string
}

type BookingResponse struct {
	ID               uuid.UUID              `json:"id"`
	BookingNumber    string                 `json:"booking_number"`
	UserID           uuid.UUID              `json:"user_id"`
	CustomerName     string                 `json:"customer_name,omitempty"`
	CustomerPhone    string                 `json:"customer_phone,omitempty"`
	WorkerID         *uuid.UUID             `json:"worker_id,omitempty"`
	Worker           *WorkerResponse        `json:"worker,omitempty"`
	ServiceID        *uuid.UUID             `json:"service_id,omitempty"`
	Service          *ServiceResponse       `json:"service,omitempty"`
	LocationTitle    string                 `json:"location_title"`
	LocationSubtitle string                 `json:"location_subtitle"`
	Latitude         float64                `json:"latitude"`
	Longitude        float64                `json:"longitude"`
	DistanceKm       float64                `json:"distance_km,omitempty"`
	SelectedItems    map[string]interface{} `json:"selected_items"`
	TotalCost        float64                `json:"total_cost"`
	DiscountAmount   float64                `json:"discount_amount"`
	FinalAmount      float64                `json:"final_amount"`
	VoucherCode      string                 `json:"voucher_code,omitempty"`
	PaymentMethod    string                 `json:"payment_method"`
	PaymentStatus    string                 `json:"payment_status"`
	Status           string                 `json:"status"`
	Notes            string                 `json:"notes,omitempty"`
	JobTitle         string                 `json:"job_title,omitempty"`
	JobDescription   string                 `json:"job_description,omitempty"`
	JobPhotos        []string               `json:"job_photos,omitempty"`
	ScheduledAt      *time.Time             `json:"scheduled_at,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
}

type RespondJobRequest struct {
	WorkerID string `json:"worker_id" binding:"required"`
	Action   string `json:"action" binding:"required"` // "accept" or "reject"
}

type UpdateBookingStatusRequest struct {
	Status        string `json:"status" binding:"required"` // pending, confirmed, in_progress, completed, cancelled
	PaymentStatus string `json:"payment_status"`            // pending, paid, refunded, failed
}
