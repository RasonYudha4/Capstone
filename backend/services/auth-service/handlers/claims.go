package handlers

import (
	"net/http"

	"auth-service/config"
	"auth-service/services"

	"github.com/gin-gonic/gin"
)

// requireClaims returns claims or writes 401/500 and returns nil.
func requireClaims(c *gin.Context) *services.Claims {
	value, exists := c.Get(config.ContextKeyUser)
	if !exists {
		respondError(c, http.StatusUnauthorized, "Not authenticated.")
		return nil
	}
	claims, ok := value.(*services.Claims)
	if !ok {
		respondError(c, http.StatusInternalServerError, "Failed to read user claims.")
		return nil
	}
	return claims
}

// httpStatus maps AuthService domain statuses to HTTP status codes.
func httpStatus(status services.ResultStatus) int {
	switch status {
	case services.StatusOK:
		return http.StatusOK
	case services.StatusBadRequest:
		return http.StatusBadRequest
	case services.StatusUnauthorized:
		return http.StatusUnauthorized
	case services.StatusForbidden:
		return http.StatusForbidden
	case services.StatusNotFound:
		return http.StatusNotFound
	case services.StatusConflict:
		return http.StatusConflict
	case services.StatusTooManyRequests:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
