package handlers

import (
	"net/http"

	"auth-service/models"

	"github.com/gin-gonic/gin"
)

func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, models.APIResponse{
		Success: false,
		Message: message,
	})
}

func respondSuccess(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// bindJSON binds the request body into dest. On failure it writes a 400 response
// and returns false.
func bindJSON(c *gin.Context, dest any) bool {
	if err := c.ShouldBindJSON(dest); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return false
	}
	return true
}
