package repository

import (
	"time"

	"delivery-backend/internal/models"

	"gorm.io/gorm"
)

type VoucherRepository interface {
	FindAllActive() ([]models.Voucher, error)
	FindByCode(code string) (*models.Voucher, error)
}

type voucherRepository struct {
	db *gorm.DB
}

func NewVoucherRepository(db *gorm.DB) VoucherRepository {
	return &voucherRepository{db: db}
}

func (r *voucherRepository) FindAllActive() ([]models.Voucher, error) {
	var vouchers []models.Voucher
	err := r.db.Where("is_active = ? AND valid_until >= ?", true, time.Now()).Order("discount_percent DESC").Find(&vouchers).Error
	return vouchers, err
}

func (r *voucherRepository) FindByCode(code string) (*models.Voucher, error) {
	var voucher models.Voucher
	err := r.db.Where("LOWER(code) = LOWER(?) AND is_active = ? AND valid_until >= ?", code, true, time.Now()).First(&voucher).Error
	if err != nil {
		return nil, err
	}
	return &voucher, nil
}
