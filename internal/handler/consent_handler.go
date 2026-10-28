package handler

import (
	"net/http"
	"strings"

	"delivery-backend/internal/dto"
	"delivery-backend/internal/middleware"
	"delivery-backend/internal/service"
	"delivery-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type ConsentHandler struct {
	consentSvc service.ConsentService
}

func NewConsentHandler(consentSvc service.ConsentService) *ConsentHandler {
	return &ConsentHandler{consentSvc: consentSvc}
}

// Grant godoc
// @Summary Record a new consent (grant) for the authenticated user.
// @Description Used for re-consent on version bump, local-ledger replay from mobile, and settings re-grant.
// @Description Idempotent on `consent_id`: replaying the same consent_id is a no-op.
// @Tags Consent
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ConsentGrantRequest true "Consent payload"
// @Success 201 {object} utils.StandardResponse
// @Router /api/v1/consent/grant [post]
func (h *ConsentHandler) Grant(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.JSONUnauthorized(c, "Authentication required")
		return
	}
	var req dto.ConsentGrantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONBadRequest(c, "Invalid consent payload", err.Error())
		return
	}
	if req.Channel == "" {
		req.Channel = detectChannel(c)
	}
	record, err := h.consentSvc.Grant(userID, req.ConsentInput)
	if err != nil {
		utils.JSONBadRequest(c, err.Error(), nil)
		return
	}
	utils.JSONSuccess(c, http.StatusCreated, "Consent recorded", record)
}

// Withdraw godoc
// @Summary Withdraw a previously granted consent.
// @Description Creates a new append-only row with granted=false, withdrawn=true.
// @Description The original grant row remains for audit purposes.
// @Tags Consent
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.ConsentWithdrawRequest true "Withdrawal payload"
// @Success 201 {object} utils.StandardResponse
// @Router /api/v1/consent/withdraw [post]
func (h *ConsentHandler) Withdraw(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.JSONUnauthorized(c, "Authentication required")
		return
	}
	var req dto.ConsentWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONBadRequest(c, "Invalid withdrawal payload", err.Error())
		return
	}
	record, err := h.consentSvc.Withdraw(userID, req)
	if err != nil {
		utils.JSONBadRequest(c, err.Error(), nil)
		return
	}
	utils.JSONSuccess(c, http.StatusCreated, "Consent withdrawn", record)
}

// Status godoc
// @Summary Current consent state + live document versions.
// @Description Lets mobile detect when re-consent is required after a version bump.
// @Tags Consent
// @Produce json
// @Security BearerAuth
// @Param types query string false "Comma-separated consent_type list (default: user_tc,user_privacy,age_confirmation)"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/consent/status [get]
func (h *ConsentHandler) Status(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.JSONUnauthorized(c, "Authentication required")
		return
	}
	var types []string
	raw := strings.TrimSpace(c.Query("types"))
	if raw != "" {
		for _, t := range strings.Split(raw, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				types = append(types, t)
			}
		}
	}
	resp, err := h.consentSvc.Status(userID, types)
	if err != nil {
		utils.JSONInternalServerError(c, "Failed to load consent data", err.Error())
		return
	}
	utils.JSONSuccess(c, http.StatusOK, "Consent status", resp)
}

// History godoc
// @Summary Full append-only consent history for the authenticated user.
// @Description Used by Settings → Privacy and by DSR data export.
// @Tags Consent
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/consent/history [get]
func (h *ConsentHandler) History(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.JSONUnauthorized(c, "Authentication required")
		return
	}
	resp, err := h.consentSvc.History(userID)
	if err != nil {
		utils.JSONInternalServerError(c, "Failed to load consent data", err.Error())
		return
	}
	utils.JSONSuccess(c, http.StatusOK, "Consent history", resp)
}

// detectChannel inspects the User-Agent to classify the request as
// mobile or web. Not security-critical — it's just a label on the
// ledger row — but useful for the DPO dashboard.
func detectChannel(c *gin.Context) string {
	ua := strings.ToLower(c.GetHeader("User-Agent"))
	switch {
	case strings.Contains(ua, "dart") || strings.Contains(ua, "okhttp") ||
		strings.Contains(ua, "cfnetwork") || strings.Contains(ua, "smart sewa"):
		return "mobile"
	default:
		return "web"
	}
}
