package handler

import (
	"net/http"
	"time"

	"delivery-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// HealthCheck godoc
// @Summary API and DB health status
// @Tags System
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	dbStatus := "healthy"
	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		dbStatus = "unhealthy"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "UP",
		"timestamp": time.Now().Format(time.RFC3339),
		"database":  dbStatus,
		"service":   "sewa-backend-api",
		"version":   "1.0.0",
	})
}

// RootHandler welcome endpoint
func (h *HealthHandler) RootHandler(c *gin.Context) {
	utils.JSONSuccess(c, http.StatusOK, "Sewa Household & Maintenance Service API is running smoothly", gin.H{
		"app":     "Sewa",
		"version": "1.0.0",
		"docs":    "/api/v1",
		"health":  "/health",
	})
}
