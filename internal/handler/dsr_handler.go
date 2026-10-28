package handler

import (
	"net/http"

	"delivery-backend/internal/middleware"
	"delivery-backend/internal/service"
	"delivery-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type DSRHandler struct {
	dsrSvc service.DSRService
}

func NewDSRHandler(dsrSvc service.DSRService) *DSRHandler {
	return &DSRHandler{dsrSvc: dsrSvc}
}

// Export godoc
// @Summary Export every piece of personal data the authenticated user has in the system.
// @Description Satisfies PDPA §20 right of access + §22 right of data portability.
// @Description Returns JSON (machine-readable); downstream can offer a PDF variant later if demand.
// @Tags DSR
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/dsr/export [get]
func (h *DSRHandler) Export(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.JSONUnauthorized(c, "Authentication required")
		return
	}
	export, err := h.dsrSvc.Export(userID)
	if err != nil {
		utils.JSONInternalServerError(c, "Failed to build data export", err.Error())
		return
	}
	utils.JSONSuccess(c, http.StatusOK, "Data export", export)
}

// RequestErasure godoc
// @Summary Request deletion of the authenticated user's account.
// @Description Starts a 7-day grace window; cancelable via DELETE /dsr/erasure.
// @Description During the grace window the account is marked inactive; after it, a background job (manual in v1) will hard-delete related rows and anonymise reviews. Consent records are preserved indefinitely.
// @Tags DSR
// @Produce json
// @Security BearerAuth
// @Success 202 {object} utils.StandardResponse
// @Router /api/v1/dsr/erasure [post]
func (h *DSRHandler) RequestErasure(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.JSONUnauthorized(c, "Authentication required")
		return
	}
	resp, err := h.dsrSvc.RequestErasure(userID)
	if err != nil {
		utils.JSONBadRequest(c, err.Error(), nil)
		return
	}
	utils.JSONSuccess(c, http.StatusAccepted, "Account scheduled for deletion", resp)
}

// CancelErasure godoc
// @Summary Cancel a pending account deletion.
// @Description Only works during the 7-day grace window. After it closes, the deletion proceeds and cannot be cancelled.
// @Tags DSR
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/dsr/erasure [delete]
func (h *DSRHandler) CancelErasure(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.JSONUnauthorized(c, "Authentication required")
		return
	}
	resp, err := h.dsrSvc.CancelErasure(userID)
	if err != nil {
		utils.JSONBadRequest(c, err.Error(), nil)
		return
	}
	utils.JSONSuccess(c, http.StatusOK, "Account deletion cancelled", resp)
}
