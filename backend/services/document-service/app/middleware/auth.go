package middleware

import (
	"capstone/app/schemas"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Extract_JWT_data(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := accessTokenFromRequest(c)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "Authentication required",
			})
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &schemas.Claims{}, func(t *jwt.Token) (any, error) {
			// Prevent algorithm confusion attacks (e.g. "none" or RSA→HMAC).
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "invalid or expired token",
			})
			return
		}
		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "token validation failed",
			})
			return
		}
		claims := token.Claims.(*schemas.Claims)
		c.Set("user_id", claims.UserId)
		c.Set("role", claims.Role)

		c.Next()
	}
}

func AllowedRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("role")

		for _, role := range allowedRoles {
			if userRole == role {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"message": "Unauthorized"})
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
