package middleware

import (
	"capstone/app/schemas"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
)

func Extract_JWT_data(secretKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "Missing authorization header",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.ParseWithClaims(tokenString, &schemas.Claims{}, func(t *jwt.Token) (any, error) {
			return []byte(secretKey), nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "invalid or expired token",
			})
			return
		}
		claims := token.Claims.(*schemas.Claims)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

func AllowedRole() gin.HandlerFunc{
	return func(c *gin.Context){
		userRole := c.GetString("role")
		if userRole != "master-admin"{
			c.AbortWithStatusJSON(403, gin.H{
			"message": "Not Authorized",
		})
		return
		}

		c.Next()
	}
}
