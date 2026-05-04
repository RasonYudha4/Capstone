package middleware

import (
	"net/http"

	"auth-service/config"
	"auth-service/models"
	"auth-service/services"

	"github.com/gin-gonic/gin"
)

// RequireRoles returns a Gin middleware that enforces role-based access control.
// It accepts a list of roles that are allowed to access the endpoint.
//
// This middleware MUST be placed AFTER JWTAuth in the middleware chain,
// because it reads the authenticated user's claims from the Gin context.
//
// Usage:
//
//	router.POST("/upload", middleware.RequireRoles(config.RoleAdmin, config.RoleMasterAdmin))
//
// How it works:
//  1. Retrieves the claims stored by JWTAuth from the Gin context.
//  2. Checks if the user's role is in the list of allowed roles.
//  3. If not → aborts with 403 Forbidden.
//  4. If yes → passes control to the next handler.
func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	// Build a set for O(1) lookup.
	roleSet := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		roleSet[role] = struct{}{}
	}

	return func(c *gin.Context) {
		// Retrieve claims from context (set by JWTAuth middleware).
		value, exists := c.Get(config.ContextKeyUser)
		if !exists {
			// This should never happen if JWTAuth is applied first,
			// but we guard against misconfiguration.
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Authentication required.",
			})
			return
		}

		claims, ok := value.(*services.Claims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "Failed to read user claims.",
			})
			return
		}

		// Check if the user's role is in the allowed set.
		if _, allowed := roleSet[claims.Role]; !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, models.APIResponse{
				Success: false,
				Message: "Access denied. Required role(s): " + formatRoles(allowedRoles),
			})
			return
		}

		c.Next()
	}
}

// formatRoles joins role names for human-readable error messages.
func formatRoles(roles []string) string {
	result := ""
	for i, r := range roles {
		if i > 0 {
			result += ", "
		}
		result += r
	}
	return result
}
