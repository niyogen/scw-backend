package handler

import (
	"net/http"

	"delivery-backend/internal/dto"
	"delivery-backend/internal/middleware"
	"delivery-backend/internal/service"
	"delivery-backend/internal/utils"

	"github.com/gin-gonic/gin"
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
