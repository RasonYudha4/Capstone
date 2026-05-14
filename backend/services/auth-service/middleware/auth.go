package middleware

import (
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
//  3. On success, stores the parsed claims in the Gin context under config.ContextKeyUser
//     so downstream handlers can access the authenticated user's email and role.
//  4. On failure, aborts the request with 401 Unauthorized.
func JWTAuth(jwtService *services.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// --- Step 1: Extract the Authorization header ---
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Authorization header is required.",
			})
			return
		}

		// --- Step 2: Validate "Bearer <token>" format ---
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Authorization header must be in the format: Bearer <token>.",
			})
			return
		}

		tokenString := parts[1]

		// --- Step 3: Validate the JWT ---
		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Invalid or expired token: " + err.Error(),
			})
			return
		}

		// --- Step 4: Store user info in context for downstream handlers ---
		c.Set(config.ContextKeyUser, claims)

		// Proceed to the next handler in the chain.
		c.Next()
	}
}
