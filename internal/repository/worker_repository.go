package repository

import (
	"delivery-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkerRepository interface {
	FindAll(serviceType string, availableOnly bool) ([]models.Worker, error)
	FindByIDOrCode(idOrCode string) (*models.Worker, error)
	Create(worker *models.Worker) error
	Update(worker *models.Worker) error
	Delete(id uuid.UUID) error
	AddReview(review *models.WorkerReview) error
	GetReviews(workerID uuid.UUID) ([]models.WorkerReview, error)
	UpdateWorkerRating(workerID uuid.UUID) error
}

type workerRepository struct {
	db *gorm.DB
}

func NewWorkerRepository(db *gorm.DB) WorkerRepository {
	return &workerRepository{db: db}
}

func (r *workerRepository) FindAll(serviceType string, availableOnly bool) ([]models.Worker, error) {
	var workers []models.Worker
	query := r.db.Preload("Reviews").Order("rating DESC, completed_jobs DESC")

	if availableOnly {
		query = query.Where("is_available_today = ?", true)
	}

	if serviceType != "" {
		// Postgres JSON serializer check for array inclusion or ILIKE
		query = query.Where("service_types::text ILIKE ?", "%"+serviceType+"%")
	}

	err := query.Find(&workers).Error
	return workers, err
}

func (r *workerRepository) FindByIDOrCode(idOrCode string) (*models.Worker, error) {
	var worker models.Worker
	parsedUUID, err := uuid.Parse(idOrCode)
	if err == nil {
		if err := r.db.Preload("Reviews").First(&worker, "id = ?", parsedUUID).Error; err == nil {
			return &worker, nil
		}
	}
	// Fallback to worker_code
	if err := r.db.Preload("Reviews").First(&worker, "worker_code = ?", idOrCode).Error; err != nil {
		return nil, err
	}
	return &worker, nil
}

func (r *workerRepository) Create(worker *models.Worker) error {
	return r.db.Create(worker).Error
}

func (r *workerRepository) Update(worker *models.Worker) error {
	return r.db.Save(worker).Error
}

func (r *workerRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Worker{}, "id = ?", id).Error
}

func (r *workerRepository) AddReview(review *models.WorkerReview) error {
	return r.db.Create(review).Error
}

func (r *workerRepository) GetReviews(workerID uuid.UUID) ([]models.WorkerReview, error) {
	var reviews []models.WorkerReview
	err := r.db.Where("worker_id = ?", workerID).Order("created_at DESC").Find(&reviews).Error
	return reviews, err
}

func (r *workerRepository) UpdateWorkerRating(workerID uuid.UUID) error {
	var result struct {
		AvgRating float64
		Count     int64
	}
	err := r.db.Model(&models.WorkerReview{}).
		Where("worker_id = ?", workerID).
		Select("AVG(rating) as avg_rating, COUNT(id) as count").
		Scan(&result).Error
	if err != nil {
		return err
	}

	return r.db.Model(&models.Worker{}).Where("id = ?", workerID).Updates(map[string]interface{}{
		"rating":        result.AvgRating,
		"reviews_count": result.Count,
	}).Error
}
