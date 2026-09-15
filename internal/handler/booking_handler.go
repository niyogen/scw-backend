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

type BookingHandler struct {
	bookingService service.BookingService
}

func NewBookingHandler(bookingService service.BookingService) *BookingHandler {
	return &BookingHandler{bookingService: bookingService}
}

// CreateBooking godoc
// @Summary Create a service booking
// @Tags Bookings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateBookingRequest true "Booking Payload"
// @Success 201 {object} utils.StandardResponse
// @Router /api/v1/bookings [post]
func (h *BookingHandler) CreateBooking(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.JSONUnauthorized(c, "User not authenticated")
		return
	}

	var req dto.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONBadRequest(c, "Invalid booking payload", err.Error())
		return
	}

	booking, err := h.bookingService.CreateBooking(userID, &req)
	if err != nil {
		utils.JSONBadRequest(c, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusCreated, "Booking created successfully", booking)
}

// GetMyBookings godoc
// @Summary Get all bookings for authenticated user
// @Tags Bookings
// @Security BearerAuth
// @Produce json
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/bookings [get]
func (h *BookingHandler) GetMyBookings(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.JSONUnauthorized(c, "User not authenticated")
		return
	}

	bookings, err := h.bookingService.GetUserBookings(userID)
	if err != nil {
		utils.JSONInternalServerError(c, "Failed to retrieve bookings", err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Bookings retrieved successfully", bookings)
}

// GetAllBookings godoc
// @Summary List all platform bookings (Admin)
// @Tags Bookings
// @Produce json
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/admin/bookings [get]
func (h *BookingHandler) GetAllBookings(c *gin.Context) {
	bookings, err := h.bookingService.GetAllBookings()
	if err != nil {
		utils.JSONInternalServerError(c, "Failed to retrieve bookings", err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Bookings retrieved successfully", bookings)
}

// GetBookingByID godoc
// @Summary Get booking details by ID
// @Tags Bookings
// @Security BearerAuth
// @Produce json
// @Param id path string true "Booking UUID"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/bookings/{id} [get]
func (h *BookingHandler) GetBookingByID(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.JSONUnauthorized(c, "User not authenticated")
		return
	}
	role := middleware.GetCurrentUserRole(c)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.JSONBadRequest(c, "Invalid booking ID format", err.Error())
		return
	}

	booking, err := h.bookingService.GetBookingByID(id, userID, role)
	if err != nil {
		utils.JSONNotFound(c, err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Booking retrieved successfully", booking)
}

// CancelBooking godoc
// @Summary Cancel a booking
// @Tags Bookings
// @Security BearerAuth
// @Produce json
// @Param id path string true "Booking UUID"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/bookings/{id}/cancel [post]
func (h *BookingHandler) CancelBooking(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.JSONUnauthorized(c, "User not authenticated")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.JSONBadRequest(c, "Invalid booking ID format", err.Error())
		return
	}

	if err := h.bookingService.CancelBooking(id, userID); err != nil {
		utils.JSONBadRequest(c, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Booking cancelled successfully", nil)
}

// UpdateBookingStatus godoc
// @Summary Update booking status (Admin / Provider)
// @Tags Bookings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Booking UUID"
// @Param request body dto.UpdateBookingStatusRequest true "Status update payload"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/bookings/{id}/status [patch]
func (h *BookingHandler) UpdateBookingStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.JSONBadRequest(c, "Invalid booking ID format", err.Error())
		return
	}

	var req dto.UpdateBookingStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONBadRequest(c, "Invalid status update payload", err.Error())
		return
	}

	if err := h.bookingService.UpdateBookingStatus(id, &req); err != nil {
		utils.JSONBadRequest(c, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Booking status updated successfully", nil)
}
