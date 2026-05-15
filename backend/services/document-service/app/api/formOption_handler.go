package api

import (
	"capstone/app/services"
	"github.com/gin-gonic/gin"
)

type FormOptionsHandler struct {
	service services.FormOptionsService
}

func NewFormOptionsHandler(service services.FormOptionsService) *FormOptionsHandler {
	return &FormOptionsHandler{service: service}
}

func (h *FormOptionsHandler) GetFormOptions(c *gin.Context) {
	data, err := h.service.GetFormOptions(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"data": data})
}