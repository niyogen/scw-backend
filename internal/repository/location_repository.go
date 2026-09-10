package repository

import (
	"delivery-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LocationRepository interface {
	FindByUserID(userID uuid.UUID) ([]models.UserLocation, error)
	Create(location *models.UserLocation) error
	SetDefault(userID, locationID uuid.UUID) error
	Delete(userID, locationID uuid.UUID) error
}

type locationRepository struct {
	db *gorm.DB
}

func NewLocationRepository(db *gorm.DB) LocationRepository {
	return &locationRepository{db: db}
}

func (r *locationRepository) FindByUserID(userID uuid.UUID) ([]models.UserLocation, error) {
	var locations []models.UserLocation
	err := r.db.Where("user_id = ?", userID).Order("is_default DESC, created_at DESC").Find(&locations).Error
	return locations, err
}

func (r *locationRepository) Create(location *models.UserLocation) error {
	if location.IsDefault {
		_ = r.db.Model(&models.UserLocation{}).Where("user_id = ?", location.UserID).Update("is_default", false).Error
	}
	return r.db.Create(location).Error
}

func (r *locationRepository) SetDefault(userID, locationID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.UserLocation{}).Where("user_id = ?", userID).Update("is_default", false).Error; err != nil {
			return err
		}
		return tx.Model(&models.UserLocation{}).Where("id = ? AND user_id = ?", locationID, userID).Update("is_default", true).Error
	})
}

func (r *locationRepository) Delete(userID, locationID uuid.UUID) error {
	return r.db.Where("id = ? AND user_id = ?", locationID, userID).Delete(&models.UserLocation{}).Error
}
