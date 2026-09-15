package dto

import (
	"github.com/google/uuid"
)

type ServiceOptionResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Unit        string    `json:"unit"`
}

type ServiceResponse struct {
	ID                  uuid.UUID               `json:"id"`
	Code                string                  `json:"code"`
	Title               string                  `json:"title"`
	ImageURL            string                  `json:"image_url"`
	HighlightedSubtitle string                  `json:"highlighted_subtitle"`
	Description         string                  `json:"description"`
	IsNew               bool                    `json:"is_new"`
	IsActive            bool                    `json:"is_active"`
	IconType            string                  `json:"icon_type"`
	StartingPrice       string                  `json:"starting_price"`
	BasePrice           float64                 `json:"base_price"`
	Category            string                  `json:"category"`
	Options             []ServiceOptionResponse `json:"options,omitempty"`
}

type CreateServiceOptionInput struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Unit        string  `json:"unit"`
}

type CreateServiceRequest struct {
	Code                string                     `json:"code"`
	Title               string                     `json:"title" binding:"required"`
	ImageURL            string                     `json:"image_url"`
	HighlightedSubtitle string                     `json:"highlighted_subtitle"`
	Description         string                     `json:"description"`
	IsNew               bool                       `json:"is_new"`
	IsActive            *bool                      `json:"is_active"`
	IconType            string                     `json:"icon_type"`
	StartingPrice       string                     `json:"starting_price"`
	BasePrice           float64                    `json:"base_price"`
	Category            string                     `json:"category"`
	Options             []CreateServiceOptionInput `json:"options"`
}

type UpdateServiceRequest struct {
	Title               *string                     `json:"title"`
	ImageURL            *string                     `json:"image_url"`
	HighlightedSubtitle *string                     `json:"highlighted_subtitle"`
	Description         *string                     `json:"description"`
	IsNew               *bool                       `json:"is_new"`
	IsActive            *bool                       `json:"is_active"`
	IconType            *string                     `json:"icon_type"`
	StartingPrice       *string                     `json:"starting_price"`
	BasePrice           *float64                    `json:"base_price"`
	Category            *string                     `json:"category"`
	Options             *[]CreateServiceOptionInput `json:"options"`
}

type ToggleServiceStatusRequest struct {
	IsActive bool `json:"is_active"`
}
