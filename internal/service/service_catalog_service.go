package service

import (
	"delivery-backend/internal/dto"
	"delivery-backend/internal/models"
	"delivery-backend/internal/repository"
)

type ServiceCatalogService interface {
	GetAllServices(category string) ([]dto.ServiceResponse, error)
	GetServiceByIDOrCode(idOrCode string) (*dto.ServiceResponse, error)
}

type serviceCatalogService struct {
	serviceRepo repository.ServiceRepository
}

func NewServiceCatalogService(serviceRepo repository.ServiceRepository) ServiceCatalogService {
	return &serviceCatalogService{serviceRepo: serviceRepo}
}

func (s *serviceCatalogService) GetAllServices(category string) ([]dto.ServiceResponse, error) {
	services, err := s.serviceRepo.FindAll(category)
	if err != nil {
		return nil, err
	}

	res := make([]dto.ServiceResponse, 0, len(services))
	for _, srv := range services {
		res = append(res, mapServiceToDTO(&srv))
	}
	return res, nil
}

func (s *serviceCatalogService) GetServiceByIDOrCode(idOrCode string) (*dto.ServiceResponse, error) {
	srv, err := s.serviceRepo.FindByIDOrCode(idOrCode)
	if err != nil {
		return nil, err
	}
	dtoRes := mapServiceToDTO(srv)
	return &dtoRes, nil
}

func mapServiceToDTO(s *models.ServiceItem) dto.ServiceResponse {
	options := make([]dto.ServiceOptionResponse, 0, len(s.Options))
	for _, opt := range s.Options {
		options = append(options, dto.ServiceOptionResponse{
			ID:          opt.ID,
			Name:        opt.Name,
			Description: opt.Description,
			Price:       opt.Price,
			Unit:        opt.Unit,
		})
	}

	return dto.ServiceResponse{
		ID:                  s.ID,
		Code:                s.Code,
		Title:               s.Title,
		HighlightedSubtitle: s.HighlightedSubtitle,
		Description:         s.Description,
		IsNew:               s.IsNew,
		IconType:            s.IconType,
		StartingPrice:       s.StartingPrice,
		BasePrice:           s.BasePrice,
		Category:            s.Category,
		Options:             options,
	}
}
