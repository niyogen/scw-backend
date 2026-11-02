package dto

import (
	"time"

	"github.com/google/uuid"
)

type WorkerReviewResponse struct {
	ID         uuid.UUID `json:"id"`
	WorkerID   uuid.UUID `json:"worker_id,omitempty"`
	WorkerName string    `json:"worker_name,omitempty"`
	AuthorName string    `json:"author"`
	Rating     float64   `json:"rating"`
	Comment    string    `json:"comment"`
	IsApproved bool      `json:"is_approved"`
	CreatedAt  time.Time `json:"created_at"`
}

type AdminReviewResponse struct {
	ID         uuid.UUID  `json:"id"`
	WorkerID   uuid.UUID  `json:"worker_id"`
	WorkerName string     `json:"worker_name"`
	WorkerCode string     `json:"worker_code"`
	UserID     *uuid.UUID `json:"user_id,omitempty"`
	UserEmail  string     `json:"user_email,omitempty"`
	AuthorName string     `json:"author"`
	Rating     float64    `json:"rating"`
	Comment    string     `json:"comment"`
	IsApproved bool       `json:"is_approved"`
	CreatedAt  time.Time  `json:"created_at"`
}

type UpdateReviewStatusRequest struct {
	IsApproved bool `json:"is_approved"`
}

type WorkerResponse struct {
	ID               uuid.UUID              `json:"id"`
	UserID           *uuid.UUID             `json:"user_id,omitempty"`
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
	Latitude         float64                `json:"latitude"`
	Longitude        float64                `json:"longitude"`
	OperatingAddress string                 `json:"operating_address"`
	IsOpenForWork    bool                   `json:"is_open_for_work"`
	RadiusKm         float64                `json:"radius_km"`
	DistanceKm       float64                `json:"distance_km,omitempty"`
	IsVerified       bool                   `json:"is_verified"`
	IsActive         bool                   `json:"is_active"`
	IsAvailableToday bool                   `json:"is_available_today"`
	Badges           []string               `json:"badges"`
	Reviews          []WorkerReviewResponse `json:"reviews,omitempty"`

	// Task #25 — admin + applicant visibility of onboarding state.
	Status         string              `json:"status,omitempty"` // pending|approved|rejected|banned
	NICNumber      string              `json:"nic_number,omitempty"`
	City           string              `json:"city,omitempty"`
	ServiceAreas   []string            `json:"service_areas,omitempty"`
	KYCDocuments   *KYCDocumentsOutput `json:"kyc_documents,omitempty"`
	ApprovedAt     *time.Time          `json:"approved_at,omitempty"`
	ApprovedBy     *uuid.UUID          `json:"approved_by,omitempty"`
	RejectedAt     *time.Time          `json:"rejected_at,omitempty"`
	RejectedReason string              `json:"rejected_reason,omitempty"`
	BannedAt       *time.Time          `json:"banned_at,omitempty"`
	BannedReason   string              `json:"banned_reason,omitempty"`
	BannedBy       *uuid.UUID          `json:"banned_by,omitempty"`
}

// KYCDocumentsOutput mirrors models.KYCDocs but is a DTO-side struct so
// the HTTP contract isn't coupled to the storage shape.
type KYCDocumentsOutput struct {
	NICFront  string `json:"nic_front,omitempty"`
	NICBack   string `json:"nic_back,omitempty"`
	Selfie    string `json:"selfie,omitempty"`
	TradeCert string `json:"trade_cert,omitempty"`
}

type UpdateWorkerLocationStatusRequest struct {
	Latitude         float64  `json:"latitude"`
	Longitude        float64  `json:"longitude"`
	OperatingAddress string   `json:"operating_address"`
	IsOpenForWork    *bool    `json:"is_open_for_work"`
	RadiusKm         *float64 `json:"radius_km"`
}

type CreateWorkerReviewRequest struct {
	BookingID *uuid.UUID `json:"booking_id"`
	Rating    float64    `json:"rating" binding:"required,min=1,max=5"`
	Comment   string     `json:"comment" binding:"required,min=3"`
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
	IsActive         *bool    `json:"is_active"`
	IsAvailableToday *bool    `json:"is_available_today"`
	Badges           []string `json:"badges"`

	// Task #25 — optional onboarding fields the Vercel site + rebuild
	// form both send. Legacy clients that omit these still work.
	NICNumber    string   `json:"nic_number,omitempty"`
	City         string   `json:"city,omitempty"`
	ServiceAreas []string `json:"service_areas,omitempty"`
}

// RegisterWorkerRequest is the public self-service onboarding payload.
// This mirrors the live Vercel form shape (see FUNCTIONALITY_AUDIT §5)
// so the production frontend keeps working while the admin approval
// loop is wired. Fields that are present on CreateWorkerRequest but
// missing here (badges, hourly_rate, etc.) are server-derived.
type RegisterWorkerRequest struct {
	Name            string   `json:"name" binding:"required"`
	AvatarURL       string   `json:"avatar_url"`
	Age             int      `json:"age"`
	ExperienceYears int      `json:"experience_years"`
	ServiceTypes    []string `json:"service_types"`
	Specializations []string `json:"specializations"`
	Phone           string   `json:"phone" binding:"required"`
	Email           string   `json:"email" binding:"required,email"`
	Bio             string   `json:"bio"`
	HourlyRate      string   `json:"hourly_rate"`
	RateValue       float64  `json:"rate_value"`
	Badges          []string `json:"badges"`
	NICNumber       string   `json:"nic_number"`
	City            string   `json:"city"`
	ServiceAreas    []string `json:"service_areas"`
}

// RegisterWorkerResponse is the self-service onboarding response. The
// JWT lets the applicant return later to check status without logging
// in via a separate flow.
type RegisterWorkerResponse struct {
	WorkerID uuid.UUID `json:"worker_id"`
	UserID   uuid.UUID `json:"user_id"`
	Status   string    `json:"status"`
	Token    string    `json:"token,omitempty"`
}

// KYCUploadURLRequest asks for a signed GCS upload URL for one document.
type KYCUploadURLRequest struct {
	// DocumentType is one of nic_front, nic_back, selfie, trade_cert.
	DocumentType string `json:"document_type" binding:"required"`
	// ContentType hints what MIME the client will PUT; defaults to
	// application/octet-stream if omitted.
	ContentType string `json:"content_type"`
	// Filename is used only to pick an extension in the GCS object path.
	Filename string `json:"filename"`
}

// KYCUploadURLResponse carries the signed URL + the final object URI
// so the client can PUT the file and then PUT /workers/:id/kyc with the
// resulting URI.
type KYCUploadURLResponse struct {
	UploadURL  string    `json:"upload_url"`
	ObjectURI  string    `json:"object_uri"`  // gs://bucket/path form, stored on the worker row
	MediaURL   string    `json:"media_url"`   // https URL for admin read-back via backend proxy
	ExpiresAt  time.Time `json:"expires_at"`
	ObjectPath string    `json:"object_path"`
}

// UpdateKYCRequest records the uploaded KYC document URIs. All fields
// are optional; the applicant can upload one document at a time and
// call this endpoint repeatedly as each upload completes.
type UpdateKYCRequest struct {
	NICFront  string `json:"nic_front"`
	NICBack   string `json:"nic_back"`
	Selfie    string `json:"selfie"`
	TradeCert string `json:"trade_cert"`
}

// UpdateMyWorkerRequest lets a worker edit their own application while
// status=pending. Only a subset of fields can be edited by the applicant
// themselves; the full CreateWorkerRequest is admin-only.
type UpdateMyWorkerRequest struct {
	Name            *string   `json:"name"`
	AvatarURL       *string   `json:"avatar_url"`
	Age             *int      `json:"age"`
	ExperienceYears *int      `json:"experience_years"`
	ServiceTypes    *[]string `json:"service_types"`
	Specializations *[]string `json:"specializations"`
	Phone           *string   `json:"phone"`
	Bio             *string   `json:"bio"`
	City            *string   `json:"city"`
	ServiceAreas    *[]string `json:"service_areas"`
	NICNumber       *string   `json:"nic_number"`
}

// WorkerStatusResponse is what /workers/me/status returns to the
// applicant. Deliberately small — no PII leaked back beyond what the
// applicant already provided.
type WorkerStatusResponse struct {
	Status         string `json:"status"`
	KYCStatus      string `json:"kyc_status"` // one of: none, partial, complete
	RejectedReason string `json:"rejected_reason,omitempty"`
	BannedReason   string `json:"banned_reason,omitempty"`
}

// AdminWorkerListItem is a compact row for the admin dashboard list.
type AdminWorkerListItem struct {
	ID          uuid.UUID `json:"id"`
	UserID      *uuid.UUID `json:"user_id,omitempty"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	City        string    `json:"city,omitempty"`
	Status      string    `json:"status"`
	NICNumber   string    `json:"nic_number,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	ServiceTypes []string `json:"service_types,omitempty"`
}

// AdminWorkerDecisionRequest is the body for reject/ban. Approve + unban
// take no body.
type AdminWorkerDecisionRequest struct {
	Reason string `json:"reason" binding:"required,min=3,max=500"`
}

// AdminInviteRequest creates a new admin via /auth/admin-invite. The
// invited admin gets a password-reset email (never a password in the
// response). Email is required and must be unique.
type AdminInviteRequest struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name" binding:"required,min=2,max=100"`
	Phone string `json:"phone"`
}

// AdminInviteResponse echoes the created admin's identity. No password
// / reset-token leaks here — the reset link is emailed out-of-band.
type AdminInviteResponse struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	InviteSent bool     `json:"invite_sent"`
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
	IsActive         *bool     `json:"is_active"`
	IsAvailableToday *bool     `json:"is_available_today"`
	Badges           *[]string `json:"badges"`
}

type ToggleWorkerStatusRequest struct {
	IsActive bool `json:"is_active"`
}
