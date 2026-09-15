package handler

import (
	"net/http"

	"delivery-backend/internal/dto"
	"delivery-backend/internal/service"
	"delivery-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type ServiceHandler struct {
	serviceCatalog service.ServiceCatalogService
}

func NewServiceHandler(serviceCatalog service.ServiceCatalogService) *ServiceHandler {
	return &ServiceHandler{serviceCatalog: serviceCatalog}
}

// GetAllServices godoc
// @Summary List all services with sub-options
// @Tags Services
// @Produce json
// @Param category query string false "Filter by category"
// @Param all query bool false "Include inactive services (Admin)"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/services [get]
func (h *ServiceHandler) GetAllServices(c *gin.Context) {
	category := c.Query("category")
	all := c.Query("all") == "true"
	services, err := h.serviceCatalog.GetAllServices(category, all)
	if err != nil {
		utils.JSONInternalServerError(c, "Failed to fetch services", err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Services retrieved successfully", services)
}

// GetServiceByID godoc
// @Summary Get service details by ID or code
// @Tags Services
// @Produce json
// @Param id path string true "Service UUID or code (e.g. srv_plumbing)"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/services/{id} [get]
func (h *ServiceHandler) GetServiceByID(c *gin.Context) {
	idOrCode := c.Param("id")
	serviceItem, err := h.serviceCatalog.GetServiceByIDOrCode(idOrCode)
	if err != nil {
		utils.JSONNotFound(c, "Service not found")
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Service retrieved successfully", serviceItem)
}

// CreateService godoc
// @Summary Create a new service offering
// @Tags Services
// @Accept json
// @Produce json
// @Param request body dto.CreateServiceRequest true "Create Service Payload"
// @Success 201 {object} utils.StandardResponse
// @Router /api/v1/services [post]
func (h *ServiceHandler) CreateService(c *gin.Context) {
	var req dto.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONBadRequest(c, "Invalid service data", err.Error())
		return
	}

	serviceItem, err := h.serviceCatalog.CreateService(&req)
	if err != nil {
		utils.JSONInternalServerError(c, "Failed to create service", err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusCreated, "Service created successfully", serviceItem)
}

// UpdateService godoc
// @Summary Update an existing service offering
// @Tags Services
// @Accept json
// @Produce json
// @Param id path string true "Service UUID or code"
// @Param request body dto.UpdateServiceRequest true "Update Service Payload"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/services/{id} [put]
func (h *ServiceHandler) UpdateService(c *gin.Context) {
	idOrCode := c.Param("id")
	var req dto.UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONBadRequest(c, "Invalid service update data", err.Error())
		return
	}

	serviceItem, err := h.serviceCatalog.UpdateService(idOrCode, &req)
	if err != nil {
		utils.JSONBadRequest(c, "Failed to update service", err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Service updated successfully", serviceItem)
}

// ToggleServiceStatus godoc
// @Summary Toggle service active/disabled status
// @Tags Services
// @Accept json
// @Produce json
// @Param id path string true "Service UUID or code"
// @Param request body dto.ToggleServiceStatusRequest true "Toggle payload"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/services/{id}/status [patch]
func (h *ServiceHandler) ToggleServiceStatus(c *gin.Context) {
	idOrCode := c.Param("id")
	var req dto.ToggleServiceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONBadRequest(c, "Invalid status payload", err.Error())
		return
	}

	if err := h.serviceCatalog.ToggleServiceActive(idOrCode, req.IsActive); err != nil {
		utils.JSONInternalServerError(c, "Failed to update service status", err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Service status updated successfully", nil)
}

// DeleteService godoc
// @Summary Delete a service offering
// @Tags Services
// @Produce json
// @Param id path string true "Service UUID or code"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/services/{id} [delete]
func (h *ServiceHandler) DeleteService(c *gin.Context) {
	idOrCode := c.Param("id")
	if err := h.serviceCatalog.DeleteService(idOrCode); err != nil {
		utils.JSONNotFound(c, "Service not found or failed to delete")
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Service deleted successfully", nil)
}
