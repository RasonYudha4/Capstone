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
//  1. Reads access token from HttpOnly cookie "access_token", or Authorization: Bearer.
//  2. Passes the raw token to JWTService.ValidateToken for signature + expiry checks.
//  3. Reloads the user from the database and rejects inactive/locked accounts.
//  4. Overwrites claims.Role/Email from the database so RBAC reflects demotions immediately.
//  5. Stores the claims in the Gin context under config.ContextKeyUser.
func JWTAuth(jwtService services.TokenValidator, userService services.UserLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := accessTokenFromRequest(c)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.APIResponse{
				Success: false,
				Message: "Authentication required.",
			})
			return
		}

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

// accessTokenFromRequest prefers the HttpOnly cookie, then Authorization Bearer.
func accessTokenFromRequest(c *gin.Context) string {
	if token, err := c.Cookie("access_token"); err == nil && token != "" {
		return token
	}

	authHeader := c.GetHeader("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}
	return ""
}
