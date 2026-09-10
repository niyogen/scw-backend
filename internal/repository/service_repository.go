package repository

import (
	"delivery-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ServiceRepository interface {
	FindAll(category string) ([]models.ServiceItem, error)
	FindByIDOrCode(idOrCode string) (*models.ServiceItem, error)
	Create(service *models.ServiceItem) error
}

type serviceRepository struct {
	db *gorm.DB
}

func NewServiceRepository(db *gorm.DB) ServiceRepository {
	return &serviceRepository{db: db}
}

func (r *serviceRepository) FindAll(category string) ([]models.ServiceItem, error) {
	var services []models.ServiceItem
	query := r.db.Preload("Options").Order("created_at ASC")
	if category != "" {
		query = query.Where("category = ?", category)
	}
	err := query.Find(&services).Error
	return services, err
}

func (r *serviceRepository) FindByIDOrCode(idOrCode string) (*models.ServiceItem, error) {
	var service models.ServiceItem
	parsedUUID, err := uuid.Parse(idOrCode)
	if err == nil {
		if err := r.db.Preload("Options").First(&service, "id = ?", parsedUUID).Error; err == nil {
			return &service, nil
		}
	}
	// Fallback to code
	if err := r.db.Preload("Options").First(&service, "code = ?", idOrCode).Error; err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *serviceRepository) Create(service *models.ServiceItem) error {
	return r.db.Create(service).Error
}
