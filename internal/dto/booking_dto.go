package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateBookingRequest struct {
	WorkerID          *uuid.UUID             `json:"worker_id"`
	ServiceID         *uuid.UUID             `json:"service_id"`
	LocationTitle     string                 `json:"location_title" binding:"required"`
	LocationSubtitle  string                 `json:"location_subtitle" binding:"required"`
	SelectedItems     map[string]interface{} `json:"selected_items"`
	TotalCost         float64                `json:"total_cost" binding:"required"`
	VoucherCode       string                 `json:"voucher_code"`
	PaymentMethod     string                 `json:"payment_method"` // promptpay, credit_card, cash
	Notes             string                 `json:"notes"`
	ScheduledAtString string                 `json:"scheduled_at"` // ISO string
}

type BookingResponse struct {
	ID               uuid.UUID              `json:"id"`
	BookingNumber    string                 `json:"booking_number"`
	UserID           uuid.UUID              `json:"user_id"`
	WorkerID         *uuid.UUID             `json:"worker_id,omitempty"`
	Worker           *WorkerResponse        `json:"worker,omitempty"`
	ServiceID        *uuid.UUID             `json:"service_id,omitempty"`
	Service          *ServiceResponse       `json:"service,omitempty"`
	LocationTitle    string                 `json:"location_title"`
	LocationSubtitle string                 `json:"location_subtitle"`
	SelectedItems    map[string]interface{} `json:"selected_items"`
	TotalCost        float64                `json:"total_cost"`
	DiscountAmount   float64                `json:"discount_amount"`
	FinalAmount      float64                `json:"final_amount"`
	VoucherCode      string                 `json:"voucher_code,omitempty"`
	PaymentMethod    string                 `json:"payment_method"`
	PaymentStatus    string                 `json:"payment_status"`
	Status           string                 `json:"status"`
	Notes            string                 `json:"notes,omitempty"`
	ScheduledAt      *time.Time             `json:"scheduled_at,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
}

type UpdateBookingStatusRequest struct {
	Status        string `json:"status" binding:"required"` // pending, confirmed, in_progress, completed, cancelled
	PaymentStatus string `json:"payment_status"`            // pending, paid, refunded, failed
}
