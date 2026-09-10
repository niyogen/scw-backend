package dto

import (
	"time"

	"github.com/google/uuid"
)

type ValidateVoucherRequest struct {
	Code     string  `json:"code" binding:"required"`
	CartCost float64 `json:"cart_cost" binding:"required,min=0"`
}

type VoucherValidationResponse struct {
	Valid          bool    `json:"valid"`
	Code           string  `json:"code"`
	Title          string  `json:"title"`
	DiscountAmount float64 `json:"discount_amount"`
	FinalAmount    float64 `json:"final_amount"`
	Message        string  `json:"message"`
}

type VoucherResponse struct {
	ID              uuid.UUID `json:"id"`
	Code            string    `json:"code"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	DiscountPercent float64   `json:"discount_percent"`
	DiscountAmount  float64   `json:"discount_amount"`
	MinSpend        float64   `json:"min_spend"`
	MaxDiscount     float64   `json:"max_discount"`
	ValidUntil      time.Time `json:"valid_until"`
	IsActive        bool      `json:"is_active"`
}
