package dto

import (
	"github.com/google/uuid"
)

type SaveLocationRequest struct {
	Title     string  `json:"title" binding:"required"`
	Subtitle  string  `json:"subtitle" binding:"required"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	IsDefault bool    `json:"is_default"`
}

type LocationResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Subtitle  string    `json:"subtitle"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	IsDefault bool      `json:"is_default"`
}
