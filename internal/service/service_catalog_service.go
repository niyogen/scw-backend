package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"delivery-backend/internal/dto"
	"delivery-backend/internal/models"
	"delivery-backend/internal/repository"
)

type ServiceCatalogService interface {
	GetAllServices(category string, all bool) ([]dto.ServiceResponse, error)
	GetServiceByIDOrCode(idOrCode string) (*dto.ServiceResponse, error)
	CreateService(req *dto.CreateServiceRequest) (*dto.ServiceResponse, error)
	UpdateService(idOrCode string, req *dto.UpdateServiceRequest) (*dto.ServiceResponse, error)
	ToggleServiceActive(idOrCode string, isActive bool) error
	DeleteService(idOrCode string) error
}

type serviceCatalogService struct {
	serviceRepo repository.ServiceRepository
}

func NewServiceCatalogService(serviceRepo repository.ServiceRepository) ServiceCatalogService {
	return &serviceCatalogService{serviceRepo: serviceRepo}
}

func (s *serviceCatalogService) GetAllServices(category string, all bool) ([]dto.ServiceResponse, error) {
	services, err := s.serviceRepo.FindAll(category, all)
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

func (s *serviceCatalogService) CreateService(req *dto.CreateServiceRequest) (*dto.ServiceResponse, error) {
	code := strings.TrimSpace(req.Code)
	if code == "" {
		slug := strings.ToLower(strings.ReplaceAll(req.Title, " ", "_"))
		code = fmt.Sprintf("srv_%s_%d", slug, time.Now().Unix()%10000)
	}

	iconType := req.IconType
	if iconType == "" {
		iconType = "other"
	}

	category := req.Category
	if category == "" {
		category = "general"
	}

	startingPrice := req.StartingPrice
	if startingPrice == "" {
		startingPrice = fmt.Sprintf("Rs. %.0f", req.BasePrice)
	}

	var options []models.ServiceOption
	for _, opt := range req.Options {
		unit := opt.Unit
		if unit == "" {
			unit = "unit"
		}
		options = append(options, models.ServiceOption{
			Name:        opt.Name,
			Description: opt.Description,
			Price:       opt.Price,
			Unit:        unit,
		})
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	serviceItem := models.ServiceItem{
		Code:                code,
		Title:               req.Title,
		ImageURL:            req.ImageURL,
		HighlightedSubtitle: req.HighlightedSubtitle,
		Description:         req.Description,
		IsNew:               req.IsNew,
		IsActive:            isActive,
		IconType:            iconType,
		StartingPrice:       startingPrice,
		BasePrice:           req.BasePrice,
		Category:            category,
		Options:             options,
	}

	if err := s.serviceRepo.Create(&serviceItem); err != nil {
		return nil, err
	}

	saved, err := s.serviceRepo.FindByIDOrCode(serviceItem.ID.String())
	if err != nil {
		res := mapServiceToDTO(&serviceItem)
		return &res, nil
	}

	res := mapServiceToDTO(saved)
	return &res, nil
}

func (s *serviceCatalogService) UpdateService(idOrCode string, req *dto.UpdateServiceRequest) (*dto.ServiceResponse, error) {
	srv, err := s.serviceRepo.FindByIDOrCode(idOrCode)
	if err != nil {
		return nil, errors.New("service not found")
	}

	if req.Title != nil {
		srv.Title = *req.Title
	}
	if req.ImageURL != nil {
		srv.ImageURL = *req.ImageURL
	}
	if req.HighlightedSubtitle != nil {
		srv.HighlightedSubtitle = *req.HighlightedSubtitle
	}
	if req.Description != nil {
		srv.Description = *req.Description
	}
	if req.IsNew != nil {
		srv.IsNew = *req.IsNew
	}
	if req.IsActive != nil {
		srv.IsActive = *req.IsActive
	}
	if req.IconType != nil {
		srv.IconType = *req.IconType
	}
	if req.StartingPrice != nil {
		srv.StartingPrice = *req.StartingPrice
	}
	if req.BasePrice != nil {
		srv.BasePrice = *req.BasePrice
	}
	if req.Category != nil {
		srv.Category = *req.Category
	}

	var options []models.ServiceOption
	if req.Options != nil {
		options = make([]models.ServiceOption, 0, len(*req.Options))
		for _, opt := range *req.Options {
			unit := opt.Unit
			if unit == "" {
				unit = "unit"
			}
			options = append(options, models.ServiceOption{
				ServiceID:   srv.ID,
				Name:        opt.Name,
				Description: opt.Description,
				Price:       opt.Price,
				Unit:        unit,
			})
		}
	}

	if err := s.serviceRepo.Update(srv, options); err != nil {
		return nil, err
	}

	saved, err := s.serviceRepo.FindByIDOrCode(srv.ID.String())
	if err != nil {
		res := mapServiceToDTO(srv)
		return &res, nil
	}

	res := mapServiceToDTO(saved)
	return &res, nil
}

func (s *serviceCatalogService) ToggleServiceActive(idOrCode string, isActive bool) error {
	srv, err := s.serviceRepo.FindByIDOrCode(idOrCode)
	if err != nil {
		return errors.New("service not found")
	}
	return s.serviceRepo.ToggleActive(srv.ID, isActive)
}

func (s *serviceCatalogService) DeleteService(idOrCode string) error {
	srv, err := s.serviceRepo.FindByIDOrCode(idOrCode)
	if err != nil {
		return errors.New("service not found")
	}
	return s.serviceRepo.Delete(srv.ID)
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
		ImageURL:            s.ImageURL,
		HighlightedSubtitle: s.HighlightedSubtitle,
		Description:         s.Description,
		IsNew:               s.IsNew,
		IsActive:            s.IsActive,
		IconType:            s.IconType,
		StartingPrice:       s.StartingPrice,
		BasePrice:           s.BasePrice,
		Category:            s.Category,
		Options:             options,
	}
}
