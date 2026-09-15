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

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register Payload"
// @Success 201 {object} utils.StandardResponse
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONBadRequest(c, "Invalid registration input", err.Error())
		return
	}

	res, err := h.authService.Register(&req)
	if err != nil {
		utils.JSONBadRequest(c, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusCreated, "User registered successfully", res)
}

// Login godoc
// @Summary Login user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login Payload"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONBadRequest(c, "Invalid login credentials format", err.Error())
		return
	}

	res, err := h.authService.Login(&req)
	if err != nil {
		utils.JSONUnauthorized(c, err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Login successful", res)
}

// ForgotPassword handles customer password recovery
// POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req dto.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONBadRequest(c, "Invalid input data", err.Error())
		return
	}

	if err := h.authService.ForgotPassword(&req); err != nil {
		utils.JSONBadRequest(c, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Password reset successfully. You can now log in with your new password.", nil)
}

// GetProfile godoc
// @Summary Get current user profile
// @Tags User
// @Security BearerAuth
// @Produce json
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/user/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.JSONUnauthorized(c, "User not authenticated")
		return
	}

	res, err := h.authService.GetProfile(userID)
	if err != nil {
		utils.JSONNotFound(c, err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Profile retrieved successfully", res)
}

// UpdateProfile godoc
// @Summary Update current user profile
// @Tags User
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.UpdateProfileRequest true "Update Profile Payload"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/user/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.JSONUnauthorized(c, "User not authenticated")
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONBadRequest(c, "Invalid update profile payload", err.Error())
		return
	}

	res, err := h.authService.UpdateProfile(userID, &req)
	if err != nil {
		utils.JSONBadRequest(c, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Profile updated successfully", res)
}

// ChangePassword godoc
// @Summary Change current user password
// @Tags User
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.ChangePasswordRequest true "Change Password Payload"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/user/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.JSONUnauthorized(c, "User not authenticated")
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONBadRequest(c, "Invalid change password payload", err.Error())
		return
	}

	if err := h.authService.ChangePassword(userID, &req); err != nil {
		utils.JSONBadRequest(c, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Password changed successfully", nil)
}

// GetAllCustomers godoc
// @Summary List all customers (Admin)
// @Tags Admin - Customers
// @Produce json
// @Param role query string false "Filter by role (customer, worker, admin)"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/admin/customers [get]
func (h *AuthHandler) GetAllCustomers(c *gin.Context) {
	role := c.Query("role")
	customers, err := h.authService.GetAllCustomers(role)
	if err != nil {
		utils.JSONInternalServerError(c, "Failed to retrieve customers", err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Customers retrieved successfully", customers)
}

// GetCustomerByID godoc
// @Summary Get customer by ID (Admin)
// @Tags Admin - Customers
// @Produce json
// @Param id path string true "Customer UUID"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/admin/customers/{id} [get]
func (h *AuthHandler) GetCustomerByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.JSONBadRequest(c, "Invalid customer ID format", err.Error())
		return
	}

	customer, err := h.authService.GetCustomerByID(id)
	if err != nil {
		utils.JSONNotFound(c, "Customer not found")
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Customer retrieved successfully", customer)
}

// CreateCustomer godoc
// @Summary Create a new customer (Admin)
// @Tags Admin - Customers
// @Accept json
// @Produce json
// @Param request body dto.CreateCustomerRequest true "Create Customer Payload"
// @Success 201 {object} utils.StandardResponse
// @Router /api/v1/admin/customers [post]
func (h *AuthHandler) CreateCustomer(c *gin.Context) {
	var req dto.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONBadRequest(c, "Invalid customer data", err.Error())
		return
	}

	customer, err := h.authService.CreateCustomer(&req)
	if err != nil {
		utils.JSONBadRequest(c, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusCreated, "Customer created successfully", customer)
}

// UpdateCustomer godoc
// @Summary Update customer details (Admin)
// @Tags Admin - Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer UUID"
// @Param request body dto.UpdateCustomerRequest true "Update Customer Payload"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/admin/customers/{id} [put]
func (h *AuthHandler) UpdateCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.JSONBadRequest(c, "Invalid customer ID format", err.Error())
		return
	}

	var req dto.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONBadRequest(c, "Invalid customer update data", err.Error())
		return
	}

	customer, err := h.authService.UpdateCustomer(id, &req)
	if err != nil {
		utils.JSONBadRequest(c, err.Error(), nil)
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Customer updated successfully", customer)
}

// ToggleCustomerStatus godoc
// @Summary Toggle customer active/disabled status (Admin)
// @Tags Admin - Customers
// @Accept json
// @Produce json
// @Param id path string true "Customer UUID"
// @Param request body dto.ToggleCustomerStatusRequest true "Toggle status"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/admin/customers/{id}/status [patch]
func (h *AuthHandler) ToggleCustomerStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.JSONBadRequest(c, "Invalid customer ID format", err.Error())
		return
	}

	var req dto.ToggleCustomerStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.JSONBadRequest(c, "Invalid status payload", err.Error())
		return
	}

	if err := h.authService.ToggleCustomerActive(id, req.IsActive); err != nil {
		utils.JSONInternalServerError(c, "Failed to update customer status", err.Error())
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Customer status updated successfully", nil)
}

// DeleteCustomer godoc
// @Summary Delete customer account (Admin)
// @Tags Admin - Customers
// @Produce json
// @Param id path string true "Customer UUID"
// @Success 200 {object} utils.StandardResponse
// @Router /api/v1/admin/customers/{id} [delete]
func (h *AuthHandler) DeleteCustomer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.JSONBadRequest(c, "Invalid customer ID format", err.Error())
		return
	}

	if err := h.authService.DeleteCustomer(id); err != nil {
		utils.JSONNotFound(c, "Customer not found or failed to delete")
		return
	}

	utils.JSONSuccess(c, http.StatusOK, "Customer deleted successfully", nil)
}
