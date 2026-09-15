package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base contains common columns for all models
type Base struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// User model represents customer accounts
type User struct {
	Base
	Name         string         `gorm:"size:255;not null" json:"name"`
	Email        string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Phone        string         `gorm:"size:50;uniqueIndex;not null" json:"phone"`
	PasswordHash string         `gorm:"size:255;not null" json:"-"`
	AvatarURL    string         `gorm:"size:500" json:"avatar_url"`
	Role         string         `gorm:"size:50;default:'customer'" json:"role"` // customer, admin, worker
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	Bookings     []Booking      `gorm:"foreignKey:UserID" json:"bookings,omitempty"`
	Addresses    []UserLocation `gorm:"foreignKey:UserID" json:"addresses,omitempty"`
}

// ServiceItem represents available household & on-demand services
type ServiceItem struct {
	Base
	Code                string          `gorm:"size:100;uniqueIndex;not null" json:"code"` // e.g. srv_plumbing
	Title               string          `gorm:"size:255;not null" json:"title"`
	ImageURL            string          `gorm:"size:500" json:"image_url"`
	HighlightedSubtitle string          `gorm:"size:255" json:"highlighted_subtitle"`
	Description         string          `gorm:"type:text" json:"description"`
	IsNew               bool            `gorm:"default:false" json:"is_new"`
	IsActive            bool            `gorm:"default:true" json:"is_active"`
	IconType            string          `gorm:"size:100;not null" json:"icon_type"` // plumbing, electricity, cleaning, etc.
	StartingPrice       string          `gorm:"size:100;not null" json:"starting_price"`
	BasePrice           float64         `gorm:"type:decimal(10,2);default:0.0" json:"base_price"`
	Category            string          `gorm:"size:100;default:'general'" json:"category"`
	Options             []ServiceOption `gorm:"foreignKey:ServiceID" json:"options,omitempty"`
}

// ServiceOption represents sub-options (e.g., Wall-Mounted AC, Ceiling Cassette AC, deep cleaning)
type ServiceOption struct {
	Base
	ServiceID   uuid.UUID `gorm:"type:uuid;not null;index" json:"service_id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Price       float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	Unit        string    `gorm:"size:50;default:'unit'" json:"unit"` // per unit, per hour, per visit
}

// Worker represents skilled service technicians & professionals
type Worker struct {
	Base
	WorkerCode       string         `gorm:"size:100;uniqueIndex;not null" json:"worker_code"` // e.g. wrk_1
	Name             string         `gorm:"size:255;not null" json:"name"`
	AvatarURL        string         `gorm:"size:500;not null" json:"avatar_url"`
	Age              int            `gorm:"not null" json:"age"`
	ExperienceYears  int            `gorm:"not null" json:"experience_years"`
	Rating           float64        `gorm:"type:decimal(3,2);default:5.00" json:"rating"`
	ReviewsCount     int            `gorm:"default:0" json:"reviews_count"`
	ServiceTypes     []string       `gorm:"type:jsonb;serializer:json" json:"service_types"` // ['plumbing', 'other']
	Specializations  []string       `gorm:"type:jsonb;serializer:json" json:"specializations"`
	Phone            string         `gorm:"size:50;not null" json:"phone"`
	Email            string         `gorm:"size:255;not null" json:"email"`
	Bio              string         `gorm:"type:text" json:"bio"`
	CompletedJobs    int            `gorm:"default:0" json:"completed_jobs"`
	HourlyRate       string         `gorm:"size:100;not null" json:"hourly_rate"`
	RateValue        float64        `gorm:"type:decimal(10,2);default:500.00" json:"rate_value"`
	IsVerified       bool           `json:"is_verified"`
	IsActive         bool           `json:"is_active"`
	IsAvailableToday bool           `gorm:"default:true" json:"is_available_today"`
	Badges           []string       `gorm:"type:jsonb;serializer:json" json:"badges"`
	Reviews          []WorkerReview `gorm:"foreignKey:WorkerID" json:"reviews,omitempty"`
}

// WorkerReview holds customer reviews for workers
type WorkerReview struct {
	Base
	WorkerID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"worker_id"`
	Worker     *Worker    `gorm:"foreignKey:WorkerID" json:"worker,omitempty"`
	UserID     *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	User       *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	BookingID  *uuid.UUID `gorm:"type:uuid;index" json:"booking_id,omitempty"`
	AuthorName string     `gorm:"size:255;not null" json:"author"`
	Rating     float64    `gorm:"type:decimal(3,2);not null" json:"rating"`
	Comment    string     `gorm:"type:text;not null" json:"comment"`
	IsApproved bool       `gorm:"default:false;index" json:"is_approved"`
}

// Booking represents a customer service booking/order
type Booking struct {
	Base
	BookingNumber     string       `gorm:"size:100;uniqueIndex;not null" json:"booking_number"`
	UserID            uuid.UUID    `gorm:"type:uuid;not null;index" json:"user_id"`
	User              *User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	WorkerID          *uuid.UUID   `gorm:"type:uuid;index" json:"worker_id,omitempty"`
	Worker            *Worker      `gorm:"foreignKey:WorkerID" json:"worker,omitempty"`
	ServiceID         *uuid.UUID   `gorm:"type:uuid;index" json:"service_id,omitempty"`
	Service           *ServiceItem `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	LocationTitle     string       `gorm:"size:255;not null" json:"location_title"`
	LocationSubtitle  string       `gorm:"type:text;not null" json:"location_subtitle"`
	SelectedItemsJSON string       `gorm:"type:jsonb" json:"selected_items"` // map of item names to count/prices
	TotalCost         float64      `gorm:"type:decimal(10,2);not null" json:"total_cost"`
	DiscountAmount    float64      `gorm:"type:decimal(10,2);default:0" json:"discount_amount"`
	FinalAmount       float64      `gorm:"type:decimal(10,2);not null" json:"final_amount"`
	VoucherCode       string       `gorm:"size:50" json:"voucher_code,omitempty"`
	PaymentMethod     string       `gorm:"size:50;default:'promptpay'" json:"payment_method"` // promptpay, credit_card, cash
	PaymentStatus     string       `gorm:"size:50;default:'pending'" json:"payment_status"`   // pending, paid, refunded, failed
	Status            string       `gorm:"size:50;default:'pending'" json:"status"`           // pending, confirmed, in_progress, completed, cancelled
	Notes             string       `gorm:"type:text" json:"notes,omitempty"`
	JobTitle          string       `gorm:"size:255" json:"job_title,omitempty"`
	JobDescription    string       `gorm:"type:text" json:"job_description,omitempty"`
	JobPhotos         []string     `gorm:"type:jsonb;serializer:json" json:"job_photos,omitempty"`
	ScheduledAt       *time.Time   `json:"scheduled_at,omitempty"`
}

// Voucher represents discount codes
type Voucher struct {
	Base
	Code            string    `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Title           string    `gorm:"size:255;not null" json:"title"`
	Description     string    `gorm:"type:text" json:"description"`
	DiscountPercent float64   `gorm:"type:decimal(5,2);default:0" json:"discount_percent"`
	DiscountAmount  float64   `gorm:"type:decimal(10,2);default:0" json:"discount_amount"`
	MinSpend        float64   `gorm:"type:decimal(10,2);default:0" json:"min_spend"`
	MaxDiscount     float64   `gorm:"type:decimal(10,2);default:0" json:"max_discount"`
	ValidUntil      time.Time `json:"valid_until"`
	IsActive        bool      `gorm:"default:true" json:"is_active"`
}

// UserLocation represents saved customer addresses
type UserLocation struct {
	Base
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Title     string    `gorm:"size:255;not null" json:"title"`
	Subtitle  string    `gorm:"type:text;not null" json:"subtitle"`
	Latitude  float64   `gorm:"type:decimal(10,8)" json:"latitude"`
	Longitude float64   `gorm:"type:decimal(11,8)" json:"longitude"`
	IsDefault bool      `gorm:"default:false" json:"is_default"`
}
