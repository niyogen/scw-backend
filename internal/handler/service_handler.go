package handler

import (
	"net/http"

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
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/services [get]
func (h *ServiceHandler) GetAllServices(c *gin.Context) {
	category := c.Query("category")
	services, err := h.serviceCatalog.GetAllServices(category)
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
