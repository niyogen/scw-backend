package repository

import (
	"delivery-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkerRepository interface {
	FindAll(serviceType string, availableOnly bool, all bool) ([]models.Worker, error)
	FindByIDOrCode(idOrCode string) (*models.Worker, error)
	Create(worker *models.Worker) error
	Update(worker *models.Worker) error
	ToggleActive(id uuid.UUID, isActive bool) error
	Delete(id uuid.UUID) error
	AddReview(review *models.WorkerReview) error
	GetReviews(workerID uuid.UUID) ([]models.WorkerReview, error)
	GetAllReviews(status string) ([]models.WorkerReview, error)
	FindReviewByID(id uuid.UUID) (*models.WorkerReview, error)
	UpdateReviewStatus(id uuid.UUID, isApproved bool) error
	DeleteReview(id uuid.UUID) error
	UpdateWorkerRating(workerID uuid.UUID) error
}

type workerRepository struct {
	db *gorm.DB
}

func NewWorkerRepository(db *gorm.DB) WorkerRepository {
	return &workerRepository{db: db}
}

func (r *workerRepository) FindAll(serviceType string, availableOnly bool, all bool) ([]models.Worker, error) {
	var workers []models.Worker
	query := r.db.Preload("Reviews", "is_approved = ?", true).Order("rating DESC, completed_jobs DESC")

	if !all {
		query = query.Where("is_active = ?", true)
	}

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
		if err := r.db.Preload("Reviews", "is_approved = ?", true).First(&worker, "id = ?", parsedUUID).Error; err == nil {
			return &worker, nil
		}
	}
	// Fallback to worker_code
	if err := r.db.Preload("Reviews", "is_approved = ?", true).First(&worker, "worker_code = ?", idOrCode).Error; err != nil {
		return nil, err
	}
	return &worker, nil
}

func (r *workerRepository) Create(worker *models.Worker) error {
	return r.db.Select("*").Create(worker).Error
}

func (r *workerRepository) Update(worker *models.Worker) error {
	return r.db.Save(worker).Error
}

func (r *workerRepository) ToggleActive(id uuid.UUID, isActive bool) error {
	return r.db.Model(&models.Worker{}).Where("id = ?", id).Update("is_active", isActive).Error
}

func (r *workerRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Worker{}, "id = ?", id).Error
}

func (r *workerRepository) AddReview(review *models.WorkerReview) error {
	return r.db.Create(review).Error
}

func (r *workerRepository) GetReviews(workerID uuid.UUID) ([]models.WorkerReview, error) {
	var reviews []models.WorkerReview
	err := r.db.Where("worker_id = ? AND is_approved = ?", workerID, true).Order("created_at DESC").Find(&reviews).Error
	return reviews, err
}

func (r *workerRepository) GetAllReviews(status string) ([]models.WorkerReview, error) {
	var reviews []models.WorkerReview
	query := r.db.Preload("Worker").Preload("User").Order("created_at DESC")
	if status == "pending" {
		query = query.Where("is_approved = ?", false)
	} else if status == "approved" {
		query = query.Where("is_approved = ?", true)
	}
	err := query.Find(&reviews).Error
	return reviews, err
}

func (r *workerRepository) FindReviewByID(id uuid.UUID) (*models.WorkerReview, error) {
	var review models.WorkerReview
	err := r.db.Preload("Worker").Preload("User").First(&review, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *workerRepository) UpdateReviewStatus(id uuid.UUID, isApproved bool) error {
	return r.db.Model(&models.WorkerReview{}).Where("id = ?", id).Update("is_approved", isApproved).Error
}

func (r *workerRepository) DeleteReview(id uuid.UUID) error {
	return r.db.Delete(&models.WorkerReview{}, "id = ?", id).Error
}

func (r *workerRepository) UpdateWorkerRating(workerID uuid.UUID) error {
	var result struct {
		AvgRating float64
		Count     int64
	}
	err := r.db.Model(&models.WorkerReview{}).
		Where("worker_id = ? AND is_approved = ?", workerID, true).
		Select("AVG(rating) as avg_rating, COUNT(id) as count").
		Scan(&result).Error
	if err != nil {
		return err
	}

	rating := result.AvgRating
	if result.Count == 0 {
		rating = 5.0
	}

	return r.db.Model(&models.Worker{}).Where("id = ?", workerID).Updates(map[string]interface{}{
		"rating":        rating,
		"reviews_count": result.Count,
	}).Error
}
