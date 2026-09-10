package service

import (
	"errors"
	"fmt"
	"time"

	"delivery-backend/internal/dto"
	"delivery-backend/internal/models"
	"delivery-backend/internal/repository"

	"github.com/google/uuid"
)

type WorkerService interface {
	GetAllWorkers(serviceType string, availableOnly bool) ([]dto.WorkerResponse, error)
	GetWorkerByIDOrCode(idOrCode string) (*dto.WorkerResponse, error)
	CreateWorker(req *dto.CreateWorkerRequest) (*dto.WorkerResponse, error)
	UpdateWorker(idOrCode string, req *dto.UpdateWorkerRequest) (*dto.WorkerResponse, error)
	DeleteWorker(idOrCode string) error
	AddWorkerReview(workerID uuid.UUID, userID uuid.UUID, req *dto.CreateWorkerReviewRequest) (*dto.WorkerReviewResponse, error)
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

func (s *workerService) GetAllWorkers(serviceType string, availableOnly bool) ([]dto.WorkerResponse, error) {
	workers, err := s.workerRepo.FindAll(serviceType, availableOnly)
	if err != nil {
		return nil, err
	}

	res := make([]dto.WorkerResponse, 0, len(workers))
	for _, w := range workers {
		res = append(res, mapWorkerToDTO(&w))
	}
	return res, nil
}

func (s *workerService) GetWorkerByIDOrCode(idOrCode string) (*dto.WorkerResponse, error) {
	worker, err := s.workerRepo.FindByIDOrCode(idOrCode)
	if err != nil {
		return nil, errors.New("worker not found")
	}
	dtoRes := mapWorkerToDTO(worker)
	return &dtoRes, nil
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
	if rateValue == 0 {
		rateValue = 500.00
	}
	if hourlyRate == "" {
		hourlyRate = fmt.Sprintf("฿%.0f/hr", rateValue)
	}

	isVerified := true
	if req.IsVerified != nil {
		isVerified = *req.IsVerified
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
			worker.HourlyRate = fmt.Sprintf("฿%.0f/hr", *req.RateValue)
		}
	}
	if req.IsVerified != nil {
		worker.IsVerified = *req.IsVerified
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

	review := models.WorkerReview{
		WorkerID:   workerID,
		UserID:     &userID,
		AuthorName: user.Name,
		Rating:     req.Rating,
		Comment:    req.Comment,
	}

	if err := s.workerRepo.AddReview(&review); err != nil {
		return nil, err
	}

	// Recalculate average rating
	_ = s.workerRepo.UpdateWorkerRating(workerID)

	return &dto.WorkerReviewResponse{
		ID:         review.ID,
		AuthorName: review.AuthorName,
		Rating:     review.Rating,
		Comment:    review.Comment,
		CreatedAt:  review.CreatedAt,
	}, nil
}

func mapWorkerToDTO(w *models.Worker) dto.WorkerResponse {
	reviews := make([]dto.WorkerReviewResponse, 0, len(w.Reviews))
	for _, r := range w.Reviews {
		reviews = append(reviews, dto.WorkerReviewResponse{
			ID:         r.ID,
			AuthorName: r.AuthorName,
			Rating:     r.Rating,
			Comment:    r.Comment,
			CreatedAt:  r.CreatedAt,
		})
	}

	return dto.WorkerResponse{
		ID:               w.ID,
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
		IsVerified:       w.IsVerified,
		IsAvailableToday: w.IsAvailableToday,
		Badges:           w.Badges,
		Reviews:          reviews,
	}
}
