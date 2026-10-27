package service

import (
	"errors"
	"fmt"
	"time"

	"delivery-backend/internal/dto"
	"delivery-backend/internal/models"
	"delivery-backend/internal/repository"
	"delivery-backend/internal/utils"

	"github.com/google/uuid"
)

type WorkerService interface {
	GetAllWorkers(serviceType string, availableOnly bool, all bool) ([]dto.WorkerResponse, error)
	GetNearbyWorkers(lat, lng float64, serviceType string, radiusKm float64) ([]dto.WorkerResponse, error)
	GetWorkerByIDOrCode(idOrCode string) (*dto.WorkerResponse, error)
	GetWorkerByUserID(userID uuid.UUID) (*dto.WorkerResponse, error)
	CreateWorker(req *dto.CreateWorkerRequest) (*dto.WorkerResponse, error)
	UpdateWorker(idOrCode string, req *dto.UpdateWorkerRequest) (*dto.WorkerResponse, error)
	UpdateLocationStatus(workerID uuid.UUID, req *dto.UpdateWorkerLocationStatusRequest) (*dto.WorkerResponse, error)
	ToggleWorkerActive(idOrCode string, isActive bool) error
	DeleteWorker(idOrCode string) error
	AddWorkerReview(workerID uuid.UUID, userID uuid.UUID, req *dto.CreateWorkerReviewRequest) (*dto.WorkerReviewResponse, error)
	GetAdminReviews(status string) ([]dto.AdminReviewResponse, error)
	UpdateReviewApproval(id uuid.UUID, isApproved bool) error
	DeleteReview(id uuid.UUID) error
}

type workerService struct {
	workerRepo repository.WorkerRepository
	userRepo   repository.UserRepository
}

func NewWorkerService(workerRepo repository.WorkerRepository, userRepo repository.UserRepository) WorkerService {
	return &workerService{
		workerRepo: workerRepo,
		userRepo:   userRepo,
	}
}

func (s *workerService) GetAllWorkers(serviceType string, availableOnly bool, all bool) ([]dto.WorkerResponse, error) {
	workers, err := s.workerRepo.FindAll(serviceType, availableOnly, all)
	if err != nil {
		return nil, err
	}

	res := make([]dto.WorkerResponse, 0, len(workers))
	for _, w := range workers {
		res = append(res, mapWorkerToDTO(&w))
	}
	return res, nil
}

func (s *workerService) GetNearbyWorkers(lat, lng float64, serviceType string, radiusKm float64) ([]dto.WorkerResponse, error) {
	if radiusKm <= 0 {
		radiusKm = 10.0 // default 10 km
	}

	// Fetch active workers who are open for work
	workers, err := s.workerRepo.FindAll(serviceType, false, false)
	if err != nil {
		return nil, err
	}

	res := make([]dto.WorkerResponse, 0)
	for _, w := range workers {
		// Worker must be open for work / available today
		if !w.IsOpenForWork && !w.IsAvailableToday {
			continue
		}

		// Calculate distance from requested location
		dist := 0.0
		if lat != 0 && lng != 0 && w.Latitude != 0 && w.Longitude != 0 {
			dist = utils.HaversineDistance(lat, lng, w.Latitude, w.Longitude)
			if dist > radiusKm {
				continue
			}
		}

		dtoW := mapWorkerToDTO(&w)
		dtoW.DistanceKm = dist
		res = append(res, dtoW)
	}

	return res, nil
}

func (s *workerService) GetWorkerByUserID(userID uuid.UUID) (*dto.WorkerResponse, error) {
	worker, err := s.workerRepo.FindByUserID(userID)
	if err != nil {
		return nil, errors.New("worker profile not found for user")
	}
	dtoRes := mapWorkerToDTO(worker)
	return &dtoRes, nil
}

func (s *workerService) UpdateLocationStatus(workerID uuid.UUID, req *dto.UpdateWorkerLocationStatusRequest) (*dto.WorkerResponse, error) {
	if err := s.workerRepo.UpdateLocationStatus(workerID, req.Latitude, req.Longitude, req.OperatingAddress, req.IsOpenForWork, req.RadiusKm); err != nil {
		return nil, err
	}

	worker, err := s.workerRepo.FindByIDOrCode(workerID.String())
	if err != nil {
		return nil, err
	}

	dtoRes := mapWorkerToDTO(worker)
	return &dtoRes, nil
}

func (s *workerService) GetWorkerByIDOrCode(idOrCode string) (*dto.WorkerResponse, error) {
	worker, err := s.workerRepo.FindByIDOrCode(idOrCode)
	if err == nil && worker != nil {
		dtoRes := mapWorkerToDTO(worker)
		return &dtoRes, nil
	}

	// Check if idOrCode corresponds to a User ID
	if parsedUUID, err := uuid.Parse(idOrCode); err == nil && s.userRepo != nil {
		if user, err := s.userRepo.FindByID(parsedUUID); err == nil && user != nil {
			// Auto-provision worker profile for this user
			newWorker := models.Worker{
				UserID:           &user.ID,
				WorkerCode:       fmt.Sprintf("wrk_%s", user.ID.String()[:8]),
				Name:             user.Name,
				Email:            user.Email,
				Phone:            user.Phone,
				AvatarURL:        user.AvatarURL,
				Age:              30,
				ExperienceYears:  5,
				Rating:           5.0,
				ReviewsCount:     0,
				ServiceTypes:     []string{"all", "other", "plumbing", "electricity", "cleaning", "building_work", "carpenter", "cook", "care_taking"},
				Specializations:  []string{"General Specialist", "All-Round Technician"},
				HourlyRate:       "Rs. 1,500 / hr",
				RateValue:        1500.0,
				Latitude:         6.9271,
				Longitude:        79.8612,
				OperatingAddress: "Colombo, Sri Lanka",
				IsOpenForWork:    true,
				RadiusKm:         20.0,
				IsVerified:       true,
				IsActive:         true,
				IsAvailableToday: true,
				Badges:           []string{"Verified Pro", "Active Partner"},
			}
			if err := s.workerRepo.Create(&newWorker); err == nil {
				dtoRes := mapWorkerToDTO(&newWorker)
				return &dtoRes, nil
			}
		}
	}

	return nil, errors.New("worker not found")
}

func (s *workerService) CreateWorker(req *dto.CreateWorkerRequest) (*dto.WorkerResponse, error) {
	workerCode := req.WorkerCode
	if workerCode == "" {
		workerCode = fmt.Sprintf("wrk_%d", time.Now().UnixNano()%1000000)
	}

	avatarURL := req.AvatarURL
	if avatarURL == "" {
		avatarURL = "https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=150&auto=format&fit=crop&q=80"
	}

	hourlyRate := req.HourlyRate
	rateValue := req.RateValue
	if rateValue == 0 && hourlyRate != "" {
		// Try to parse numeric rate if hourlyRate was provided without rateValue
		var parsed float64
		if _, err := fmt.Sscanf(hourlyRate, "Rs. %f", &parsed); err == nil {
			rateValue = parsed
		}
	}
	if rateValue == 0 {
		rateValue = 1500.00
	}
	if hourlyRate == "" {
		hourlyRate = fmt.Sprintf("Rs. %.0f/hr", rateValue)
	}

	isVerified := true
	if req.IsVerified != nil {
		isVerified = *req.IsVerified
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	isAvailableToday := true
	if req.IsAvailableToday != nil {
		isAvailableToday = *req.IsAvailableToday
	}

	specializations := req.Specializations
	if specializations == nil {
		specializations = []string{}
	}

	badges := req.Badges
	if badges == nil {
		badges = []string{"Verified Pro", "Fast Response"}
	}

	worker := models.Worker{
		WorkerCode:       workerCode,
		Name:             req.Name,
		AvatarURL:        avatarURL,
		Age:              req.Age,
		ExperienceYears:  req.ExperienceYears,
		Rating:           5.00,
		ReviewsCount:     0,
		ServiceTypes:     req.ServiceTypes,
		Specializations:  specializations,
		Phone:            req.Phone,
		Email:            req.Email,
		Bio:              req.Bio,
		CompletedJobs:    0,
		HourlyRate:       hourlyRate,
		RateValue:        rateValue,
		IsVerified:       isVerified,
		IsActive:         isActive,
		IsAvailableToday: isAvailableToday,
		Badges:           badges,
	}

	if err := s.workerRepo.Create(&worker); err != nil {
		return nil, err
	}

	res := mapWorkerToDTO(&worker)
	return &res, nil
}

func (s *workerService) UpdateWorker(idOrCode string, req *dto.UpdateWorkerRequest) (*dto.WorkerResponse, error) {
	worker, err := s.workerRepo.FindByIDOrCode(idOrCode)
	if err != nil {
		return nil, errors.New("worker not found")
	}

	if req.Name != nil {
		worker.Name = *req.Name
	}
	if req.AvatarURL != nil {
		worker.AvatarURL = *req.AvatarURL
	}
	if req.Age != nil {
		worker.Age = *req.Age
	}
	if req.ExperienceYears != nil {
		worker.ExperienceYears = *req.ExperienceYears
	}
	if req.ServiceTypes != nil {
		worker.ServiceTypes = *req.ServiceTypes
	}
	if req.Specializations != nil {
		worker.Specializations = *req.Specializations
	}
	if req.Phone != nil {
		worker.Phone = *req.Phone
	}
	if req.Email != nil {
		worker.Email = *req.Email
	}
	if req.Bio != nil {
		worker.Bio = *req.Bio
	}
	if req.CompletedJobs != nil {
		worker.CompletedJobs = *req.CompletedJobs
	}
	if req.HourlyRate != nil {
		worker.HourlyRate = *req.HourlyRate
	}
	if req.RateValue != nil {
		worker.RateValue = *req.RateValue
		if req.HourlyRate == nil {
			worker.HourlyRate = fmt.Sprintf("Rs. %.0f/hr", *req.RateValue)
		}
	}
	if req.IsVerified != nil {
		worker.IsVerified = *req.IsVerified
	}
	if req.IsActive != nil {
		worker.IsActive = *req.IsActive
	}
	if req.IsAvailableToday != nil {
		worker.IsAvailableToday = *req.IsAvailableToday
	}
	if req.Badges != nil {
		worker.Badges = *req.Badges
	}

	if err := s.workerRepo.Update(worker); err != nil {
		return nil, err
	}

	res := mapWorkerToDTO(worker)
	return &res, nil
}

func (s *workerService) ToggleWorkerActive(idOrCode string, isActive bool) error {
	worker, err := s.workerRepo.FindByIDOrCode(idOrCode)
	if err != nil {
		return errors.New("worker not found")
	}
	return s.workerRepo.ToggleActive(worker.ID, isActive)
}

func (s *workerService) DeleteWorker(idOrCode string) error {
	worker, err := s.workerRepo.FindByIDOrCode(idOrCode)
	if err != nil {
		return errors.New("worker not found")
	}
	return s.workerRepo.Delete(worker.ID)
}

func (s *workerService) AddWorkerReview(workerID uuid.UUID, userID uuid.UUID, req *dto.CreateWorkerReviewRequest) (*dto.WorkerReviewResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	worker, err := s.workerRepo.FindByIDOrCode(workerID.String())
	if err != nil {
		return nil, errors.New("worker not found")
	}

	review := models.WorkerReview{
		WorkerID:   workerID,
		UserID:     &userID,
		BookingID:  req.BookingID,
		AuthorName: user.Name,
		Rating:     req.Rating,
		Comment:    req.Comment,
		IsApproved: false, // Pending admin review
	}

	if err := s.workerRepo.AddReview(&review); err != nil {
		return nil, err
	}

	return &dto.WorkerReviewResponse{
		ID:         review.ID,
		WorkerID:   workerID,
		WorkerName: worker.Name,
		AuthorName: review.AuthorName,
		Rating:     review.Rating,
		Comment:    review.Comment,
		IsApproved: review.IsApproved,
		CreatedAt:  review.CreatedAt,
	}, nil
}

func (s *workerService) GetAdminReviews(status string) ([]dto.AdminReviewResponse, error) {
	reviews, err := s.workerRepo.GetAllReviews(status)
	if err != nil {
		return nil, err
	}

	res := make([]dto.AdminReviewResponse, 0, len(reviews))
	for _, r := range reviews {
		var workerName, workerCode, userEmail string
		if r.Worker != nil {
			workerName = r.Worker.Name
			workerCode = r.Worker.WorkerCode
		}
		if r.User != nil {
			userEmail = r.User.Email
		}
		res = append(res, dto.AdminReviewResponse{
			ID:         r.ID,
			WorkerID:   r.WorkerID,
			WorkerName: workerName,
			WorkerCode: workerCode,
			UserID:     r.UserID,
			UserEmail:  userEmail,
			AuthorName: r.AuthorName,
			Rating:     r.Rating,
			Comment:    r.Comment,
			IsApproved: r.IsApproved,
			CreatedAt:  r.CreatedAt,
		})
	}
	return res, nil
}

func (s *workerService) UpdateReviewApproval(id uuid.UUID, isApproved bool) error {
	review, err := s.workerRepo.FindReviewByID(id)
	if err != nil {
		return errors.New("review not found")
	}

	if err := s.workerRepo.UpdateReviewStatus(id, isApproved); err != nil {
		return err
	}

	// Recalculate average rating & review count for the worker
	return s.workerRepo.UpdateWorkerRating(review.WorkerID)
}

func (s *workerService) DeleteReview(id uuid.UUID) error {
	review, err := s.workerRepo.FindReviewByID(id)
	if err != nil {
		return errors.New("review not found")
	}

	workerID := review.WorkerID
	if err := s.workerRepo.DeleteReview(id); err != nil {
		return err
	}

	// Recalculate worker rating
	return s.workerRepo.UpdateWorkerRating(workerID)
}

func mapWorkerToDTO(w *models.Worker) dto.WorkerResponse {
	reviews := make([]dto.WorkerReviewResponse, 0, len(w.Reviews))
	for _, r := range w.Reviews {
		if r.IsApproved {
			reviews = append(reviews, dto.WorkerReviewResponse{
				ID:         r.ID,
				WorkerID:   w.ID,
				WorkerName: w.Name,
				AuthorName: r.AuthorName,
				Rating:     r.Rating,
				Comment:    r.Comment,
				IsApproved: r.IsApproved,
				CreatedAt:  r.CreatedAt,
			})
		}
	}

	return dto.WorkerResponse{
		ID:               w.ID,
		UserID:           w.UserID,
		WorkerCode:       w.WorkerCode,
		Name:             w.Name,
		AvatarURL:        w.AvatarURL,
		Age:              w.Age,
		ExperienceYears:  w.ExperienceYears,
		Rating:           w.Rating,
		ReviewsCount:     w.ReviewsCount,
		ServiceTypes:     w.ServiceTypes,
		Specializations:  w.Specializations,
		Phone:            w.Phone,
		Email:            w.Email,
		Bio:              w.Bio,
		CompletedJobs:    w.CompletedJobs,
		HourlyRate:       w.HourlyRate,
		RateValue:        w.RateValue,
		Latitude:         w.Latitude,
		Longitude:        w.Longitude,
		OperatingAddress: w.OperatingAddress,
		IsOpenForWork:    w.IsOpenForWork,
		RadiusKm:         w.RadiusKm,
		IsVerified:       w.IsVerified,
		IsActive:         w.IsActive,
		IsAvailableToday: w.IsAvailableToday,
		Badges:           w.Badges,
		Reviews:          reviews,
	}
}
