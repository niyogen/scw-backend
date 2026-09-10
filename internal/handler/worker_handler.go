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

type WorkerHandler struct {
	workerService service.WorkerService
}

func NewWorkerHandler(workerService service.WorkerService) *WorkerHandler {
	return &WorkerHandler{workerService: workerService}
}

// GetAllWorkers godoc
// @Summary List workers with filters
// @Tags Workers
// @Produce json
// @Param service_type query string false "Filter by service type (e.g. plumbing, electricity)"
// @Param available_only query bool false "Filter only currently available workers"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/workers [get]
func (h *WorkerHandler) GetAllWorkers(c *gin.Context) {
	serviceType := c.Query("service_type")
	availableOnly := c.Query("available_only") == "true"

	workers, err := h.workerService.GetAllWorkers(serviceType, availableOnly)
	if err != nil {
		utils.JSONInternalServerError(c, "Failed to fetch workers", err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Workers retrieved successfully", workers)
}

// GetWorkerByID godoc
// @Summary Get worker details by ID or code
// @Tags Workers
// @Produce json
// @Param id path string true "Worker UUID or code (e.g. wrk_1)"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/workers/{id} [get]
func (h *WorkerHandler) GetWorkerByID(c *gin.Context) {
	idOrCode := c.Param("id")
	worker, err := h.workerService.GetWorkerByIDOrCode(idOrCode)
	if err != nil {
		utils.JSONNotFound(c, "Worker not found")
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Worker retrieved successfully", worker)
}

// CreateWorker godoc
// @Summary Create a new service worker
// @Tags Workers
// @Accept json
// @Produce json
// @Param request body dto.CreateWorkerRequest true "Create Worker Payload"
// @Success 201 {object} utils.StandardResponse
// @Router /api/v1/workers [post]
func (h *WorkerHandler) CreateWorker(c *gin.Context) {
	var req dto.CreateWorkerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONBadRequest(c, "Invalid worker data", err.Error())
		return
	}

	worker, err := h.workerService.CreateWorker(&req)
	if err != nil {
		utils.JSONInternalServerError(c, "Failed to create worker", err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusCreated, "Worker created successfully", worker)
}

// UpdateWorker godoc
// @Summary Update an existing service worker
// @Tags Workers
// @Accept json
// @Produce json
// @Param id path string true "Worker UUID or code"
// @Param request body dto.UpdateWorkerRequest true "Update Worker Payload"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/workers/{id} [put]
func (h *WorkerHandler) UpdateWorker(c *gin.Context) {
	idOrCode := c.Param("id")
	var req dto.UpdateWorkerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONBadRequest(c, "Invalid worker update data", err.Error())
		return
	}

	worker, err := h.workerService.UpdateWorker(idOrCode, &req)
	if err != nil {
		utils.JSONBadRequest(c, "Failed to update worker", err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Worker updated successfully", worker)
}

// DeleteWorker godoc
// @Summary Delete a service worker
// @Tags Workers
// @Produce json
// @Param id path string true "Worker UUID or code"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/workers/{id} [delete]
func (h *WorkerHandler) DeleteWorker(c *gin.Context) {
	idOrCode := c.Param("id")
	if err := h.workerService.DeleteWorker(idOrCode); err != nil {
		utils.JSONNotFound(c, "Worker not found or failed to delete")
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Worker deleted successfully", nil)
}

// AddWorkerReview godoc
// @Summary Add a review for a worker
// @Tags Workers
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Worker UUID"
// @Param request body dto.CreateWorkerReviewRequest true "Review Payload"
// @Success 201 {object} utils.StandardResponse
// @Router /api/v1/workers/{id}/reviews [post]
func (h *WorkerHandler) AddWorkerReview(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.JSONUnauthorized(c, "User not authenticated")
		return
	}

	workerIDStr := c.Param("id")
	workerID, err := uuid.Parse(workerIDStr)
	if err != nil {
		// Try looking up by code
		w, err := h.workerService.GetWorkerByIDOrCode(workerIDStr)
		if err != nil {
			utils.JSONNotFound(c, "Worker not found")
			return
		}
		workerID = w.ID
	}

	var req dto.CreateWorkerReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONBadRequest(c, "Invalid review payload", err.Error())
		return
	}

	review, err := h.workerService.AddWorkerReview(workerID, userID, &req)
	if err != nil {
		utils.JSONBadRequest(c, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusCreated, "Review submitted successfully", review)
}
