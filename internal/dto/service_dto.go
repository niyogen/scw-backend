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
	HighlightedSubtitle string                  `json:"highlighted_subtitle"`
	Description         string                  `json:"description"`
	IsNew               bool                    `json:"is_new"`
	IconType            string                  `json:"icon_type"`
	StartingPrice       string                  `json:"starting_price"`
	BasePrice           float64                 `json:"base_price"`
	Category            string                  `json:"category"`
	Options             []ServiceOptionResponse `json:"options,omitempty"`
}

type CreateServiceRequest struct {
	Code                string                  `json:"code" binding:"required"`
	Title               string                  `json:"title" binding:"required"`
	HighlightedSubtitle string                  `json:"highlighted_subtitle"`
	Description         string                  `json:"description"`
	IsNew               bool                    `json:"is_new"`
	IconType            string                  `json:"icon_type" binding:"required"`
	StartingPrice       string                  `json:"starting_price" binding:"required"`
	BasePrice           float64                 `json:"base_price"`
	Category            string                  `json:"category"`
	Options             []ServiceOptionResponse `json:"options"`
}
