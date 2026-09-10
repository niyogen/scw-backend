package handler

import (
	"net/http"

	"delivery-backend/internal/dto"
	"delivery-backend/internal/middleware"
	"delivery-backend/internal/service"
	"delivery-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LocationHandler struct {
	locationService service.LocationService
}

func NewLocationHandler(locationService service.LocationService) *LocationHandler {
	return &LocationHandler{locationService: locationService}
}

// GetLocations godoc
// @Summary List saved user locations
// @Tags Location
// @Security BearerAuth
// @Produce json
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/user/locations [get]
func (h *LocationHandler) GetLocations(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.JSONUnauthorized(c, "User not authenticated")
		return
	}

	locations, err := h.locationService.GetUserLocations(userID)
	if err != nil {
		utils.JSONInternalServerError(c, "Failed to retrieve locations", err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Locations retrieved successfully", locations)
}

// AddLocation godoc
// @Summary Add a new address/location for user
// @Tags Location
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.SaveLocationRequest true "Location payload"
// @Success 201 {object} utils.StandardResponse
// @Router /api/v1/user/locations [post]
func (h *LocationHandler) AddLocation(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.JSONUnauthorized(c, "User not authenticated")
		return
	}

	var req dto.SaveLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONBadRequest(c, "Invalid location payload", err.Error())
		return
	}

	res, err := h.locationService.AddLocation(userID, &req)
	if err != nil {
		utils.JSONBadRequest(c, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusCreated, "Location added successfully", res)
}

// SetDefaultLocation godoc
// @Summary Set an address as default
// @Tags Location
// @Security BearerAuth
// @Produce json
// @Param id path string true "Location UUID"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/user/locations/{id}/default [put]
func (h *LocationHandler) SetDefaultLocation(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.JSONUnauthorized(c, "User not authenticated")
		return
	}

	idStr := c.Param("id")
	locID, err := uuid.Parse(idStr)
	if err != nil {
		utils.JSONBadRequest(c, "Invalid location ID format", err.Error())
		return
	}

	if err := h.locationService.SetDefaultLocation(userID, locID); err != nil {
		utils.JSONBadRequest(c, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Default location updated successfully", nil)
}

// DeleteLocation godoc
// @Summary Delete a saved address
// @Tags Location
// @Security BearerAuth
// @Produce json
// @Param id path string true "Location UUID"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/user/locations/{id} [delete]
func (h *LocationHandler) DeleteLocation(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.JSONUnauthorized(c, "User not authenticated")
		return
	}

	idStr := c.Param("id")
	locID, err := uuid.Parse(idStr)
	if err != nil {
		utils.JSONBadRequest(c, "Invalid location ID format", err.Error())
		return
	}

	if err := h.locationService.DeleteLocation(userID, locID); err != nil {
		utils.JSONBadRequest(c, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Location deleted successfully", nil)
}
