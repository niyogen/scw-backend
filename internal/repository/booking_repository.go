package repository

import (
	"encoding/json"

	"delivery-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookingRepository interface {
	Create(booking *models.Booking) error
	FindByID(id uuid.UUID) (*models.Booking, error)
	FindByBookingNumber(num string) (*models.Booking, error)
	FindByUserID(userID uuid.UUID) ([]models.Booking, error)
	FindByWorkerID(workerID uuid.UUID) ([]models.Booking, error)
	FindAvailablePendingJobs() ([]models.Booking, error)
	AssignWorker(bookingID uuid.UUID, workerID uuid.UUID, status string) error
	RejectJobForWorker(bookingID uuid.UUID, workerIDStr string) error
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

func (r *bookingRepository) FindByWorkerID(workerID uuid.UUID) ([]models.Booking, error) {
	var bookings []models.Booking
	err := r.db.Preload("User").Preload("Worker").Preload("Service").
		Where("worker_id = ?", workerID).
		Order("created_at DESC").
		Find(&bookings).Error
	return bookings, err
}

func (r *bookingRepository) FindAvailablePendingJobs() ([]models.Booking, error) {
	var bookings []models.Booking
	// All pending jobs waiting for worker acceptance
	err := r.db.Preload("User").Preload("Worker").Preload("Service").Preload("Service.Options").
		Where("status IN ?", []string{"pending", "requested"}).
		Order("created_at DESC").
		Find(&bookings).Error
	return bookings, err
}

func (r *bookingRepository) AssignWorker(bookingID uuid.UUID, workerID uuid.UUID, status string) error {
	if status == "" {
		status = "confirmed"
	}
	return r.db.Model(&models.Booking{}).Where("id = ?", bookingID).Updates(map[string]interface{}{
		"worker_id": workerID,
		"status":    status,
	}).Error
}

func (r *bookingRepository) RejectJobForWorker(bookingID uuid.UUID, workerIDStr string) error {
	var booking models.Booking
	if err := r.db.First(&booking, "id = ?", bookingID).Error; err != nil {
		return err
	}
	// Append worker ID to rejected list
	alreadyRejected := false
	for _, id := range booking.RejectedWorkers {
		if id == workerIDStr {
			alreadyRejected = true
			break
		}
	}
	if !alreadyRejected {
		booking.RejectedWorkers = append(booking.RejectedWorkers, workerIDStr)
	}

	rejectedJSON, _ := json.Marshal(booking.RejectedWorkers)
	updates := map[string]interface{}{
		"rejected_workers": string(rejectedJSON),
	}

	// If this booking was specifically assigned to this worker, clear worker_id so other nearby workers can pick it up
	parsedWorkerID, err := uuid.Parse(workerIDStr)
	if err == nil && booking.WorkerID != nil && *booking.WorkerID == parsedWorkerID {
		updates["worker_id"] = nil
	}

	return r.db.Model(&models.Booking{}).Where("id = ?", bookingID).Updates(updates).Error
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
