package handler

import (
	"net/http"

	"delivery-backend/internal/dto"
	"delivery-backend/internal/service"
	"delivery-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type VoucherHandler struct {
	voucherService service.VoucherService
}

func NewVoucherHandler(voucherService service.VoucherService) *VoucherHandler {
	return &VoucherHandler{voucherService: voucherService}
}

// GetActiveVouchers godoc
// @Summary List active promotional vouchers
// @Tags Vouchers
// @Produce json
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/vouchers [get]
func (h *VoucherHandler) GetActiveVouchers(c *gin.Context) {
	vouchers, err := h.voucherService.GetActiveVouchers()
	if err != nil {
		utils.JSONInternalServerError(c, "Failed to retrieve vouchers", err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Active vouchers retrieved successfully", vouchers)
}

// ValidateVoucher godoc
// @Summary Validate and calculate discount for a voucher code
// @Tags Vouchers
// @Accept json
// @Produce json
// @Param request body dto.ValidateVoucherRequest true "Validation payload"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/vouchers/validate [post]
func (h *VoucherHandler) ValidateVoucher(c *gin.Context) {
	var req dto.ValidateVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONBadRequest(c, "Invalid voucher validation payload", err.Error())
		return
	}

	res, err := h.voucherService.ValidateVoucher(&req)
	if err != nil {
		utils.JSONBadRequest(c, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Voucher validated", res)
}
