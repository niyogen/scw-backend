package repository

import (
	"delivery-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookingRepository interface {
	Create(booking *models.Booking) error
	FindByID(id uuid.UUID) (*models.Booking, error)
	FindByBookingNumber(num string) (*models.Booking, error)
	FindByUserID(userID uuid.UUID) ([]models.Booking, error)
	UpdateStatus(id uuid.UUID, status string, paymentStatus string) error
	FindAll() ([]models.Booking, error)
}

type bookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) Create(booking *models.Booking) error {
	return r.db.Create(booking).Error
}

func (r *bookingRepository) FindByID(id uuid.UUID) (*models.Booking, error) {
	var booking models.Booking
	err := r.db.Preload("User").Preload("Worker").Preload("Service").Preload("Service.Options").First(&booking, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *bookingRepository) FindByBookingNumber(num string) (*models.Booking, error) {
	var booking models.Booking
	err := r.db.Preload("User").Preload("Worker").Preload("Service").First(&booking, "booking_number = ?", num).Error
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (r *bookingRepository) FindByUserID(userID uuid.UUID) ([]models.Booking, error) {
	var bookings []models.Booking
	err := r.db.Preload("Worker").Preload("Service").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&bookings).Error
	return bookings, err
}

func (r *bookingRepository) UpdateStatus(id uuid.UUID, status string, paymentStatus string) error {
	updates := map[string]interface{}{}
	if status != "" {
		updates["status"] = status
	}
	if paymentStatus != "" {
		updates["payment_status"] = paymentStatus
	}
	return r.db.Model(&models.Booking{}).Where("id = ?", id).Updates(updates).Error
}

func (r *bookingRepository) FindAll() ([]models.Booking, error) {
	var bookings []models.Booking
	err := r.db.Preload("User").Preload("Worker").Preload("Service").Order("created_at DESC").Find(&bookings).Error
	return bookings, err
}
