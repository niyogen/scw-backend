package repository

import (
	"regexp"
	"strings"

	"delivery-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByID(id uuid.UUID) (*models.User, error)
	FindByEmailOrPhone(input string) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindByPhone(phone string) (*models.User, error)
	FindAll(role string) ([]models.User, error)
	Update(user *models.User) error
	UpdatePassword(id uuid.UUID, newHash string) error
	Delete(id uuid.UUID) error
	ToggleActive(id uuid.UUID, isActive bool) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func extractDigits(s string) string {
	reg := regexp.MustCompile(`[^0-9]`)
	return reg.ReplaceAllString(s, "")
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Addresses").First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmailOrPhone(input string) (*models.User, error) {
	var user models.User
	cleanInput := strings.TrimSpace(input)
	if cleanInput == "" {
		return nil, gorm.ErrRecordNotFound
	}

	// 1. Direct match on email (case-insensitive) or exact phone
	err := r.db.Where("LOWER(email) = LOWER(?) OR phone = ?", cleanInput, cleanInput).First(&user).Error
	if err == nil {
		return &user, nil
	}

	// 2. Flexible phone number matching across formatting variations
	digits := extractDigits(cleanInput)
	if len(digits) >= 8 {
		last9 := digits
		if len(digits) > 9 {
			last9 = digits[len(digits)-9:]
		}

		err = r.db.Where(
			"REGEXP_REPLACE(phone, '[^0-9]', '', 'g') = ? OR RIGHT(REGEXP_REPLACE(phone, '[^0-9]', '', 'g'), 9) = ?",
			digits, last9,
		).First(&user).Error
		if err == nil {
			return &user, nil
		}
	}

	return nil, err
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("LOWER(email) = LOWER(?)", strings.TrimSpace(email)).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByPhone(phone string) (*models.User, error) {
	var user models.User
	cleanPhone := strings.TrimSpace(phone)
	if cleanPhone == "" {
		return nil, gorm.ErrRecordNotFound
	}

	err := r.db.Where("phone = ?", cleanPhone).First(&user).Error
	if err == nil {
		return &user, nil
	}

	digits := extractDigits(cleanPhone)
	if len(digits) >= 8 {
		last9 := digits
		if len(digits) > 9 {
			last9 = digits[len(digits)-9:]
		}

		err = r.db.Where(
			"REGEXP_REPLACE(phone, '[^0-9]', '', 'g') = ? OR RIGHT(REGEXP_REPLACE(phone, '[^0-9]', '', 'g'), 9) = ?",
			digits, last9,
		).First(&user).Error
		if err == nil {
			return &user, nil
		}
	}

	return nil, err
}

func (r *userRepository) FindAll(role string) ([]models.User, error) {
	var users []models.User
	query := r.db.Preload("Addresses").Preload("Bookings").Order("created_at DESC")
	if role != "" {
		query = query.Where("role = ?", role)
	}
	err := query.Find(&users).Error
	return users, err
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) UpdatePassword(id uuid.UUID, newHash string) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("password_hash", newHash).Error
}

func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}

func (r *userRepository) ToggleActive(id uuid.UUID, isActive bool) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("is_active", isActive).Error
}
