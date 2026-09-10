package service

import (
	"fmt"
	"math"

	"delivery-backend/internal/dto"
	"delivery-backend/internal/models"
	"delivery-backend/internal/repository"
)

type VoucherService interface {
	GetActiveVouchers() ([]dto.VoucherResponse, error)
	ValidateVoucher(req *dto.ValidateVoucherRequest) (*dto.VoucherValidationResponse, error)
}

type voucherService struct {
	voucherRepo repository.VoucherRepository
}

func NewVoucherService(voucherRepo repository.VoucherRepository) VoucherService {
	return &voucherService{voucherRepo: voucherRepo}
}

func (s *voucherService) GetActiveVouchers() ([]dto.VoucherResponse, error) {
	vouchers, err := s.voucherRepo.FindAllActive()
	if err != nil {
		return nil, err
	}

	res := make([]dto.VoucherResponse, 0, len(vouchers))
	for _, v := range vouchers {
		res = append(res, mapVoucherToDTO(&v))
	}
	return res, nil
}

func (s *voucherService) ValidateVoucher(req *dto.ValidateVoucherRequest) (*dto.VoucherValidationResponse, error) {
	voucher, err := s.voucherRepo.FindByCode(req.Code)
	if err != nil {
		return &dto.VoucherValidationResponse{
			Valid:   false,
			Code:    req.Code,
			Message: "Invalid or expired promo code",
		}, nil
	}

	if req.CartCost < voucher.MinSpend {
		return &dto.VoucherValidationResponse{
			Valid:   false,
			Code:    voucher.Code,
			Title:   voucher.Title,
			Message: fmt.Sprintf("Minimum spend of %.2f ฿ required to use this voucher", voucher.MinSpend),
		}, nil
	}

	var discount float64
	if voucher.DiscountPercent > 0 {
		discount = (req.CartCost * voucher.DiscountPercent) / 100.0
		if voucher.MaxDiscount > 0 && discount > voucher.MaxDiscount {
			discount = voucher.MaxDiscount
		}
	} else if voucher.DiscountAmount > 0 {
		discount = voucher.DiscountAmount
	}

	finalAmount := math.Max(0, req.CartCost-discount)

	return &dto.VoucherValidationResponse{
		Valid:          true,
		Code:           voucher.Code,
		Title:          voucher.Title,
		DiscountAmount: discount,
		FinalAmount:    finalAmount,
		Message:        "Voucher applied successfully",
	}, nil
}

func mapVoucherToDTO(v *models.Voucher) dto.VoucherResponse {
	return dto.VoucherResponse{
		ID:              v.ID,
		Code:            v.Code,
		Title:           v.Title,
		Description:     v.Description,
		DiscountPercent: v.DiscountPercent,
		DiscountAmount:  v.DiscountAmount,
		MinSpend:        v.MinSpend,
		MaxDiscount:     v.MaxDiscount,
		ValidUntil:      v.ValidUntil,
		IsActive:        v.IsActive,
	}
}
