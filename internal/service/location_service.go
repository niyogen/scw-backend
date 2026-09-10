package service

import (
	"delivery-backend/internal/dto"
	"delivery-backend/internal/models"
	"delivery-backend/internal/repository"

	"github.com/google/uuid"
)

type LocationService interface {
	GetUserLocations(userID uuid.UUID) ([]dto.LocationResponse, error)
	AddLocation(userID uuid.UUID, req *dto.SaveLocationRequest) (*dto.LocationResponse, error)
	SetDefaultLocation(userID, locationID uuid.UUID) error
	DeleteLocation(userID, locationID uuid.UUID) error
}

type locationService struct {
	locationRepo repository.LocationRepository
}

func NewLocationService(locationRepo repository.LocationRepository) LocationService {
	return &locationService{locationRepo: locationRepo}
}

func (s *locationService) GetUserLocations(userID uuid.UUID) ([]dto.LocationResponse, error) {
	locations, err := s.locationRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	res := make([]dto.LocationResponse, 0, len(locations))
	for _, l := range locations {
		res = append(res, dto.LocationResponse{
			ID:        l.ID,
			Title:     l.Title,
			Subtitle:  l.Subtitle,
			Latitude:  l.Latitude,
			Longitude: l.Longitude,
			IsDefault: l.IsDefault,
		})
	}
	return res, nil
}

func (s *locationService) AddLocation(userID uuid.UUID, req *dto.SaveLocationRequest) (*dto.LocationResponse, error) {
	location := models.UserLocation{
		UserID:    userID,
		Title:     req.Title,
		Subtitle:  req.Subtitle,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		IsDefault: req.IsDefault,
	}

	if err := s.locationRepo.Create(&location); err != nil {
		return nil, err
	}

	return &dto.LocationResponse{
		ID:        location.ID,
		Title:     location.Title,
		Subtitle:  location.Subtitle,
		Latitude:  location.Latitude,
		Longitude: location.Longitude,
		IsDefault: location.IsDefault,
	}, nil
}

func (s *locationService) SetDefaultLocation(userID, locationID uuid.UUID) error {
	return s.locationRepo.SetDefault(userID, locationID)
}

func (s *locationService) DeleteLocation(userID, locationID uuid.UUID) error {
	return s.locationRepo.Delete(userID, locationID)
}
