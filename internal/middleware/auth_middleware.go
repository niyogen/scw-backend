package middleware

import (
	"net/http"
	"strings"

	"delivery-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthMiddleware validates JWT Bearer token
func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.JSONUnauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.JSONUnauthorized(c, "Invalid Authorization header format. Expected 'Bearer <token>'")
			c.Abort()
			return
		}

		claims, err := utils.ValidateJWT(parts[1], secret)
		if err != nil {
			utils.JSONUnauthorized(c, "Invalid or expired authorization token")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", claims.Role)
		c.Next()
	}
}

// GetCurrentUserID retrieves the authenticated user's ID from gin context
func GetCurrentUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, false
	}
	userID, ok := val.(uuid.UUID)
	return userID, ok
}

// GetCurrentUserRole retrieves the user's role from context
func GetCurrentUserRole(c *gin.Context) string {
	val, exists := c.Get("userRole")
	if !exists {
		return "customer"
	}
	if role, ok := val.(string); ok {
		return role
	}
	return "customer"
}

// RequireRole enforces specific roles (e.g. admin)
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetCurrentUserRole(c)
		for _, r := range allowedRoles {
			if r == role {
				c.Next()
				return
			}
		}
		utils.JSONForbidden(c, "You do not have permission to access this resource")
		c.Abort()
	}
}

// CORS returns CORS middleware
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
