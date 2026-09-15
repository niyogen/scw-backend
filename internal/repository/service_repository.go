package repository

import (
	"delivery-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ServiceRepository interface {
	FindAll(category string, all bool) ([]models.ServiceItem, error)
	FindByIDOrCode(idOrCode string) (*models.ServiceItem, error)
	Create(service *models.ServiceItem) error
	Update(service *models.ServiceItem, options []models.ServiceOption) error
	ToggleActive(id uuid.UUID, isActive bool) error
	Delete(id uuid.UUID) error
}

type serviceRepository struct {
	db *gorm.DB
}

func NewServiceRepository(db *gorm.DB) ServiceRepository {
	return &serviceRepository{db: db}
}

func (r *serviceRepository) FindAll(category string, all bool) ([]models.ServiceItem, error) {
	var services []models.ServiceItem
	query := r.db.Preload("Options").Order("created_at ASC")
	if !all {
		query = query.Where("is_active = ?", true)
	}
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

func (r *serviceRepository) Update(service *models.ServiceItem, options []models.ServiceOption) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(service).Error; err != nil {
			return err
		}
		// If options were supplied to update
		if options != nil {
			if err := tx.Where("service_id = ?", service.ID).Delete(&models.ServiceOption{}).Error; err != nil {
				return err
			}
			for i := range options {
				options[i].ServiceID = service.ID
				if err := tx.Create(&options[i]).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *serviceRepository) ToggleActive(id uuid.UUID, isActive bool) error {
	return r.db.Model(&models.ServiceItem{}).Where("id = ?", id).Update("is_active", isActive).Error
}

func (r *serviceRepository) Delete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("service_id = ?", id).Delete(&models.ServiceOption{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.ServiceItem{}, "id = ?", id).Error
	})
}
