package dto

import (
	"time"

	"github.com/google/uuid"
)

// DSRExportResponse is what GET /dsr/export returns. All personal data
// the user has in the system, machine-readable (JSON) — satisfies PDPA
// §20 right of access + §22 right of data portability.
type DSRExportResponse struct {
	ExportedAt time.Time                 `json:"exported_at"`
	Subject    DSRExportSubject          `json:"subject"`
	Account    *UserResponse             `json:"account"`
	Bookings   []DSRExportBooking        `json:"bookings"`
	Addresses  []DSRExportAddress        `json:"saved_addresses"`
	Reviews    []DSRExportReview         `json:"reviews_given"`
	Consents   []ConsentRecordResponse   `json:"consent_history"`
	Media      []DSRExportMediaReference `json:"uploaded_media"`

	// Note: raw ID is not re-included under Subject so the hash remains
	// the privacy boundary. The user can see all their data, but the
	// export doesn't make it easier for a leaked copy to be correlated
	// against the subject_id_hash in consent_records.
}

type DSRExportSubject struct {
	UserID        uuid.UUID `json:"user_id"`
	SubjectIDHash string    `json:"subject_id_hash"`
	ExportVersion string    `json:"export_version"`
}

type DSRExportBooking struct {
	BookingID         uuid.UUID  `json:"booking_id"`
	BookingNumber     string     `json:"booking_number"`
	ServiceID         *uuid.UUID `json:"service_id,omitempty"`
	WorkerID          *uuid.UUID `json:"worker_id,omitempty"`
	LocationTitle     string     `json:"location_title"`
	LocationSubtitle  string     `json:"location_subtitle"`
	Status            string     `json:"status"`
	JobTitle          string     `json:"job_title,omitempty"`
	JobDescription    string     `json:"job_description,omitempty"`
	JobPhotos         []string   `json:"job_photos,omitempty"`
	ScheduledAt       *time.Time `json:"scheduled_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type DSRExportAddress struct {
	AddressID uuid.UUID `json:"address_id"`
	Title     string    `json:"title"`
	Subtitle  string    `json:"subtitle"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
}

type DSRExportReview struct {
	ReviewID  uuid.UUID  `json:"review_id"`
	WorkerID  uuid.UUID  `json:"worker_id"`
	BookingID *uuid.UUID `json:"booking_id,omitempty"`
	Rating    float64    `json:"rating"`
	Comment   string     `json:"comment"`
	CreatedAt time.Time  `json:"created_at"`
}

type DSRExportMediaReference struct {
	Context  string `json:"context"` // "job_photo", "avatar", etc.
	URL      string `json:"url"`
	Uploaded string `json:"uploaded_at,omitempty"`
}

// DSRErasureRequest is empty — the user ID comes from the auth token.
// Server derives everything else; no client-provided fields prevent
// accidental erasure of a different account.
type DSRErasureRequest struct{}

// DSRErasureResponse tells the user when their account will be deleted
// and how to cancel.
type DSRErasureResponse struct {
	Status              string    `json:"status"` // "scheduled"
	RequestedAt         time.Time `json:"requested_at"`
	ScheduledDeletionAt time.Time `json:"scheduled_deletion_at"`
	GraceDays           int       `json:"grace_days"`
	CancelableUntil     time.Time `json:"cancelable_until"`
	Message             string    `json:"message"`
}

// DSRErasureCancelResponse acknowledges cancellation during the grace window.
type DSRErasureCancelResponse struct {
	Status      string    `json:"status"` // "cancelled"
	CancelledAt time.Time `json:"cancelled_at"`
	Message     string    `json:"message"`
}
