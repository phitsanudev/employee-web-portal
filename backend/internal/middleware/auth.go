package middleware

import (
	"strings"

	"employee-portal/backend/internal/domain"
	"employee-portal/backend/internal/response"
	"employee-portal/backend/internal/service"
	"github.com/gin-gonic/gin"
	"log/slog"
)

const UserIDKey = "user_id"
const UserRoleKey = "user_role"

func Auth(authService *service.AuthService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			response.Fail(c, logger, domain.NewAppError(domain.ErrUnauthorized, "UNAUTHORIZED", "Missing bearer token"))
			c.Abort()
			return
		}
		claims, err := authService.Verify(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			response.Fail(c, logger, err)
			c.Abort()
			return
		}
		c.Set(UserIDKey, claims.UserID)
		c.Set(UserRoleKey, claims.Role)
		c.Next()
	}
}

func RequireRole(role string, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if UserRole(c) != role {
			response.Fail(c, logger, domain.NewAppError(domain.ErrForbidden, "FORBIDDEN", "This action requires admin permission"))
			c.Abort()
			return
		}
		c.Next()
	}
}

func UserID(c *gin.Context) uint {
	value, exists := c.Get(UserIDKey)
	if !exists {
		return 0
	}
	id, _ := value.(uint)
	return id
}

func UserRole(c *gin.Context) string {
	value, exists := c.Get(UserRoleKey)
	if !exists {
		return ""
	}
	role, _ := value.(string)
	return role
}
