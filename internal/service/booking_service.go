package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"delivery-backend/internal/dto"
	"delivery-backend/internal/models"
	"delivery-backend/internal/repository"
	"delivery-backend/internal/utils"

	"github.com/google/uuid"
)

type BookingService interface {
	CreateBooking(userID uuid.UUID, req *dto.CreateBookingRequest) (*dto.BookingResponse, error)
	GetUserBookings(userID uuid.UUID) ([]dto.BookingResponse, error)
	GetWorkerBookings(workerID uuid.UUID) ([]dto.BookingResponse, error)
	GetAvailableJobsForWorker(workerID uuid.UUID, lat, lng float64, radiusKm float64) ([]dto.BookingResponse, error)
	RespondToJob(bookingID uuid.UUID, req *dto.RespondJobRequest) (*dto.BookingResponse, error)
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
	emailSvc    EmailService
}

func NewBookingService(
	bookingRepo repository.BookingRepository,
	voucherRepo repository.VoucherRepository,
	workerRepo repository.WorkerRepository,
	serviceRepo repository.ServiceRepository,
	emailSvc EmailService,
) BookingService {
	return &bookingService{
		bookingRepo: bookingRepo,
		voucherRepo: voucherRepo,
		workerRepo:  workerRepo,
		serviceRepo: serviceRepo,
		emailSvc:    emailSvc,
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

	lat := req.Latitude
	lng := req.Longitude
	if lat == 0 && lng == 0 {
		lat = 6.9271
		lng = 79.8612
	}

	var workerUUID *uuid.UUID
	if req.WorkerID != nil && strings.TrimSpace(*req.WorkerID) != "" {
		workerIDStr := strings.TrimSpace(*req.WorkerID)
		if worker, err := s.workerRepo.FindByIDOrCode(workerIDStr); err == nil && worker != nil {
			workerUUID = &worker.ID
		} else if parsedUUID, err := uuid.Parse(workerIDStr); err == nil {
			workerUUID = &parsedUUID
		}
	}

	var serviceUUID *uuid.UUID
	if req.ServiceID != nil && strings.TrimSpace(*req.ServiceID) != "" {
		serviceIDStr := strings.TrimSpace(*req.ServiceID)
		if service, err := s.serviceRepo.FindByIDOrCode(serviceIDStr); err == nil && service != nil {
			serviceUUID = &service.ID
		} else if parsedUUID, err := uuid.Parse(serviceIDStr); err == nil {
			serviceUUID = &parsedUUID
		}
	}

	booking := models.Booking{
		BookingNumber:     bookingNum,
		UserID:            userID,
		WorkerID:          workerUUID,
		ServiceID:         serviceUUID,
		LocationTitle:     req.LocationTitle,
		LocationSubtitle:  req.LocationSubtitle,
		Latitude:          lat,
		Longitude:         lng,
		SelectedItemsJSON: string(selectedJSONBytes),
		TotalCost:         req.TotalCost,
		DiscountAmount:    discount,
		FinalAmount:       finalAmount,
		VoucherCode:       req.VoucherCode,
		PaymentMethod:     req.PaymentMethod,
		PaymentStatus:     "pending",
		Status:            "pending", // First job status is always pending for worker acceptance
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

	// Trigger transactional emails
	if s.emailSvc != nil {
		s.emailSvc.SendBookingCreatedCustomer(saved)
		if saved.Worker != nil {
			s.emailSvc.SendJobAlertWorker(saved, saved.Worker)
		}
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

func (s *bookingService) GetWorkerBookings(workerID uuid.UUID) ([]dto.BookingResponse, error) {
	bookings, err := s.bookingRepo.FindByWorkerID(workerID)
	if err != nil {
		return nil, err
	}

	res := make([]dto.BookingResponse, 0, len(bookings))
	for _, b := range bookings {
		res = append(res, mapBookingToDTO(&b))
	}
	return res, nil
}

func (s *bookingService) GetAvailableJobsForWorker(workerID uuid.UUID, lat, lng float64, radiusKm float64) ([]dto.BookingResponse, error) {
	if radiusKm <= 0 {
		radiusKm = 20.0 // Default 20km radius
	}

	worker, err := s.workerRepo.FindByIDOrCode(workerID.String())
	if err != nil {
		return nil, errors.New("worker not found")
	}

	// Use worker's stored location if query params are not provided
	if lat == 0 && lng == 0 {
		lat = worker.Latitude
		lng = worker.Longitude
	}
	if lat == 0 && lng == 0 {
		lat = 6.9271
		lng = 79.8612
	}

	pendingBookings, err := s.bookingRepo.FindAvailablePendingJobs()
	if err != nil {
		return nil, err
	}

	res := make([]dto.BookingResponse, 0)
	workerIDStr := workerID.String()

	for _, b := range pendingBookings {
		// 1. Skip if already rejected by this worker
		rejected := false
		for _, rej := range b.RejectedWorkers {
			if rej == workerIDStr || (worker.WorkerCode != "" && rej == worker.WorkerCode) {
				rejected = true
				break
			}
		}
		if rejected {
			continue
		}

		// 2. If targeted directly to this worker, include immediately
		isDirectlyAssigned := b.WorkerID != nil && *b.WorkerID != uuid.Nil && *b.WorkerID == workerID
		if b.WorkerID != nil && *b.WorkerID != uuid.Nil && !isDirectlyAssigned {
			continue
		}

		jobLat := b.Latitude
		jobLng := b.Longitude
		if jobLat == 0 && jobLng == 0 {
			jobLat = 6.9271
			jobLng = 79.8612
		}

		dist := utils.HaversineDistance(lat, lng, jobLat, jobLng)

		if !isDirectlyAssigned {
			// 3. Check if category matches worker's service types
			matchesCategory := false
			if len(worker.ServiceTypes) == 0 {
				matchesCategory = true
			} else {
				for _, st := range worker.ServiceTypes {
					stLower := strings.ToLower(strings.TrimSpace(st))
					if stLower == "other" || stLower == "general" || stLower == "all" || stLower == "custom" {
						matchesCategory = true
						break
					}
					if b.Service != nil {
						iconLower := strings.ToLower(b.Service.IconType)
						catLower := strings.ToLower(b.Service.Category)
						codeLower := strings.ToLower(b.Service.Code)
						if strings.Contains(iconLower, stLower) || strings.Contains(stLower, iconLower) ||
							strings.Contains(catLower, stLower) || strings.Contains(stLower, catLower) ||
							strings.Contains(codeLower, stLower) || strings.Contains(stLower, codeLower) {
							matchesCategory = true
							break
						}
					}
					if b.JobTitle != "" {
						titleLower := strings.ToLower(b.JobTitle)
						if strings.Contains(titleLower, stLower) || strings.Contains(stLower, titleLower) {
							matchesCategory = true
							break
						}
					}
				}
			}
			if !matchesCategory && b.Service != nil {
				continue
			}

			// 4. Compute distance in km
			if dist > radiusKm {
				continue
			}
		}

		dtoB := mapBookingToDTO(&b)
		dtoB.DistanceKm = dist
		res = append(res, dtoB)
	}

	return res, nil
}

func (s *bookingService) RespondToJob(bookingID uuid.UUID, req *dto.RespondJobRequest) (*dto.BookingResponse, error) {
	booking, err := s.bookingRepo.FindByID(bookingID)
	if err != nil {
		return nil, errors.New("booking not found")
	}

	worker, err := s.workerRepo.FindByIDOrCode(req.WorkerID)
	var workerID uuid.UUID
	if err == nil && worker != nil {
		workerID = worker.ID
	} else if parsed, err := uuid.Parse(req.WorkerID); err == nil {
		workerID = parsed
	} else {
		return nil, errors.New("worker not found")
	}

	action := strings.ToLower(strings.TrimSpace(req.Action))

	if action == "accept" || action == "approve" || action == "accepted" || action == "approved" {
		if booking.WorkerID != nil && *booking.WorkerID != uuid.Nil && *booking.WorkerID != workerID {
			return nil, errors.New("this job has already been accepted by another worker")
		}

		if err := s.bookingRepo.AssignWorker(bookingID, workerID, "confirmed"); err != nil {
			return nil, err
		}

		updated, err := s.bookingRepo.FindByID(bookingID)
		if err != nil {
			return nil, err
		}

		if s.emailSvc != nil {
			s.emailSvc.SendBookingApprovedCustomer(updated, updated.Worker)
			s.emailSvc.SendBookingApprovedWorker(updated, updated.Worker)
		}

		dtoRes := mapBookingToDTO(updated)
		return &dtoRes, nil
	} else if action == "reject" || action == "decline" || action == "rejected" || action == "declined" {
		if err := s.bookingRepo.RejectJobForWorker(bookingID, workerID.String()); err != nil {
			return nil, err
		}

		if s.emailSvc != nil {
			s.emailSvc.SendBookingRejectedCustomer(booking, booking.Worker)
		}

		dtoRes := mapBookingToDTO(booking)
		return &dtoRes, nil
	}

	return nil, errors.New("invalid action: must be 'approve' (accept) or 'reject' (decline)")
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
	booking, err := s.bookingRepo.FindByID(id)
	if err != nil {
		return errors.New("booking not found")
	}

	status := strings.ToLower(strings.TrimSpace(req.Status))
	paymentStatus := req.PaymentStatus

	if status == "completed" {
		if paymentStatus == "" || paymentStatus == "pending" {
			paymentStatus = "paid"
		}
		// Increment completed job count for worker
		if booking.WorkerID != nil && *booking.WorkerID != uuid.Nil {
			_ = s.workerRepo.IncrementCompletedJobs(*booking.WorkerID)
		}
	} else if status == "cancelled" || status == "rejected" {
		if paymentStatus == "" {
			paymentStatus = "refunded"
		}
	}

	if err := s.bookingRepo.UpdateStatus(id, status, paymentStatus); err != nil {
		return err
	}

	if s.emailSvc != nil {
		s.emailSvc.SendBookingStatusUpdate(booking, status)
	}

	return nil
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

	if err := s.bookingRepo.UpdateStatus(id, "cancelled", "refunded"); err != nil {
		return err
	}

	if s.emailSvc != nil {
		s.emailSvc.SendBookingStatusUpdate(booking, "cancelled")
	}

	return nil
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

	var customerName, customerPhone string
	if b.User != nil {
		customerName = b.User.Name
		customerPhone = b.User.Phone
	}

	return dto.BookingResponse{
		ID:               b.ID,
		BookingNumber:    b.BookingNumber,
		UserID:           b.UserID,
		CustomerName:     customerName,
		CustomerPhone:    customerPhone,
		WorkerID:         b.WorkerID,
		Worker:           workerDTO,
		ServiceID:        b.ServiceID,
		Service:          serviceDTO,
		LocationTitle:    b.LocationTitle,
		LocationSubtitle: b.LocationSubtitle,
		Latitude:         b.Latitude,
		Longitude:        b.Longitude,
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
