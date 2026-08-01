package middleware

import (
	"log"
	"net/http"
	"strings"

	"auth-service/config"
	"auth-service/models"
	"auth-service/services"

	"github.com/gin-gonic/gin"
)

// JWTAuth returns a Gin middleware that authenticates requests using JWT.
//
// How it works:
//  1. Reads the "Authorization" header and expects the format "Bearer <token>".
//  2. Passes the raw token to JWTService.ValidateToken for signature + expiry checks.
//  3. Reloads the user from the database and rejects inactive/locked accounts.
//  4. Overwrites claims.Role/Email from the database so RBAC reflects demotions immediately.
//  5. Stores the claims in the Gin context under config.ContextKeyUser.
func JWTAuth(jwtService services.TokenValidator, userService services.UserLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Authorization header is required.",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Authorization header must be in the format: Bearer <token>.",
			})
			return
		}

		tokenString := parts[1]

		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			log.Printf("⚠️  JWT validation failed: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Authentication failed.",
			})
			return
		}

		user, err := userService.GetByID(claims.UserID)
		if err != nil {
			log.Printf("⚠️  Failed to load user %s during JWT auth: %v", claims.UserID, err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Authentication failed.",
			})
			return
		}
		if user == nil || user.AccountStatus != models.AccountStatusActive {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Account is no longer active.",
			})
			return
		}
		if user.IsLocked() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, models.APIResponse{
				Success: false,
				Message: "Account is locked. Try again later.",
			})
			return
		}

		// Prefer authoritative DB values so role demotion takes effect immediately.
		claims.Role = user.Role
		claims.Email = user.Email
		c.Set(config.ContextKeyUser, claims)

		c.Next()
	}
}
