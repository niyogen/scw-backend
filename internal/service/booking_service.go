package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"delivery-backend/internal/dto"
	"delivery-backend/internal/models"
	"delivery-backend/internal/repository"

	"github.com/google/uuid"
)

type BookingService interface {
	CreateBooking(userID uuid.UUID, req *dto.CreateBookingRequest) (*dto.BookingResponse, error)
	GetUserBookings(userID uuid.UUID) ([]dto.BookingResponse, error)
	GetAllBookings() ([]dto.BookingResponse, error)
	GetBookingByID(id uuid.UUID, userID uuid.UUID, role string) (*dto.BookingResponse, error)
	UpdateBookingStatus(id uuid.UUID, req *dto.UpdateBookingStatusRequest) error
	CancelBooking(id uuid.UUID, userID uuid.UUID) error
}

type bookingService struct {
	bookingRepo repository.BookingRepository
	voucherRepo repository.VoucherRepository
	workerRepo  repository.WorkerRepository
	serviceRepo repository.ServiceRepository
}

func NewBookingService(
	bookingRepo repository.BookingRepository,
	voucherRepo repository.VoucherRepository,
	workerRepo repository.WorkerRepository,
	serviceRepo repository.ServiceRepository,
) BookingService {
	return &bookingService{
		bookingRepo: bookingRepo,
		voucherRepo: voucherRepo,
		workerRepo:  workerRepo,
		serviceRepo: serviceRepo,
	}
}

func (s *bookingService) CreateBooking(userID uuid.UUID, req *dto.CreateBookingRequest) (*dto.BookingResponse, error) {
	discount := 0.0
	if req.VoucherCode != "" {
		if voucher, err := s.voucherRepo.FindByCode(req.VoucherCode); err == nil && voucher != nil {
			if req.TotalCost >= voucher.MinSpend {
				if voucher.DiscountPercent > 0 {
					discount = (req.TotalCost * voucher.DiscountPercent) / 100.0
					if voucher.MaxDiscount > 0 && discount > voucher.MaxDiscount {
						discount = voucher.MaxDiscount
					}
				} else if voucher.DiscountAmount > 0 {
					discount = voucher.DiscountAmount
				}
			}
		}
	}

	finalAmount := req.TotalCost - discount
	if finalAmount < 0 {
		finalAmount = 0
	}

	selectedJSONBytes, _ := json.Marshal(req.SelectedItems)

	var scheduledAt *time.Time
	if req.ScheduledAtString != "" {
		if t, err := time.Parse(time.RFC3339, req.ScheduledAtString); err == nil {
			scheduledAt = &t
		}
	}

	bookingNum := fmt.Sprintf("BK-%s-%04d", time.Now().Format("20060102"), rand.Intn(10000))

	booking := models.Booking{
		BookingNumber:     bookingNum,
		UserID:            userID,
		WorkerID:          req.WorkerID,
		ServiceID:         req.ServiceID,
		LocationTitle:     req.LocationTitle,
		LocationSubtitle:  req.LocationSubtitle,
		SelectedItemsJSON: string(selectedJSONBytes),
		TotalCost:         req.TotalCost,
		DiscountAmount:    discount,
		FinalAmount:       finalAmount,
		VoucherCode:       req.VoucherCode,
		PaymentMethod:     req.PaymentMethod,
		PaymentStatus:     "pending",
		Status:            "confirmed",
		Notes:             req.Notes,
		JobTitle:          req.JobTitle,
		JobDescription:    req.JobDescription,
		JobPhotos:         req.JobPhotos,
		ScheduledAt:       scheduledAt,
	}

	if err := s.bookingRepo.Create(&booking); err != nil {
		return nil, err
	}

	// Reload with relationships
	saved, err := s.bookingRepo.FindByID(booking.ID)
	if err != nil {
		return nil, err
	}

	dtoRes := mapBookingToDTO(saved)
	return &dtoRes, nil
}

func (s *bookingService) GetUserBookings(userID uuid.UUID) ([]dto.BookingResponse, error) {
	bookings, err := s.bookingRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	res := make([]dto.BookingResponse, 0, len(bookings))
	for _, b := range bookings {
		res = append(res, mapBookingToDTO(&b))
	}
	return res, nil
}

func (s *bookingService) GetAllBookings() ([]dto.BookingResponse, error) {
	bookings, err := s.bookingRepo.FindAll()
	if err != nil {
		return nil, err
	}

	res := make([]dto.BookingResponse, 0, len(bookings))
	for _, b := range bookings {
		res = append(res, mapBookingToDTO(&b))
	}
	return res, nil
}

func (s *bookingService) GetBookingByID(id uuid.UUID, userID uuid.UUID, role string) (*dto.BookingResponse, error) {
	booking, err := s.bookingRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("booking not found")
	}

	if role != "admin" && booking.UserID != userID {
		return nil, errors.New("unauthorized to view this booking")
	}

	dtoRes := mapBookingToDTO(booking)
	return &dtoRes, nil
}

func (s *bookingService) UpdateBookingStatus(id uuid.UUID, req *dto.UpdateBookingStatusRequest) error {
	return s.bookingRepo.UpdateStatus(id, req.Status, req.PaymentStatus)
}

func (s *bookingService) CancelBooking(id uuid.UUID, userID uuid.UUID) error {
	booking, err := s.bookingRepo.FindByID(id)
	if err != nil {
		return errors.New("booking not found")
	}

	if booking.UserID != userID {
		return errors.New("unauthorized to cancel this booking")
	}

	if booking.Status == "completed" || booking.Status == "cancelled" {
		return fmt.Errorf("cannot cancel booking in %s status", booking.Status)
	}

	return s.bookingRepo.UpdateStatus(id, "cancelled", "refunded")
}

func mapBookingToDTO(b *models.Booking) dto.BookingResponse {
	var selectedItems map[string]interface{}
	if b.SelectedItemsJSON != "" {
		_ = json.Unmarshal([]byte(b.SelectedItemsJSON), &selectedItems)
	}

	var workerDTO *dto.WorkerResponse
	if b.Worker != nil {
		w := mapWorkerToDTO(b.Worker)
		workerDTO = &w
	}

	var serviceDTO *dto.ServiceResponse
	if b.Service != nil {
		s := mapServiceToDTO(b.Service)
		serviceDTO = &s
	}

	return dto.BookingResponse{
		ID:               b.ID,
		BookingNumber:    b.BookingNumber,
		UserID:           b.UserID,
		WorkerID:         b.WorkerID,
		Worker:           workerDTO,
		ServiceID:        b.ServiceID,
		Service:          serviceDTO,
		LocationTitle:    b.LocationTitle,
		LocationSubtitle: b.LocationSubtitle,
		SelectedItems:    selectedItems,
		TotalCost:        b.TotalCost,
		DiscountAmount:   b.DiscountAmount,
		FinalAmount:      b.FinalAmount,
		VoucherCode:      b.VoucherCode,
		PaymentMethod:    b.PaymentMethod,
		PaymentStatus:    b.PaymentStatus,
		Status:           b.Status,
		Notes:            b.Notes,
		JobTitle:         b.JobTitle,
		JobDescription:   b.JobDescription,
		JobPhotos:        b.JobPhotos,
		ScheduledAt:      b.ScheduledAt,
		CreatedAt:        b.CreatedAt,
	}
}
