package api

import (
	"capstone/app/services"
	"log"

	"github.com/gin-gonic/gin"
)

type FormOptionsHandler struct {
	service services.FormOptionsService
}

func NewFormOptionsHandler(service services.FormOptionsService) *FormOptionsHandler {
	return &FormOptionsHandler{service: service}
}

func (h *FormOptionsHandler) GetFormOptions(c *gin.Context) {
	userId := c.GetString("user_id")
	log.Print(userId)

	data, err := h.service.GetFormOptions(c.Request.Context(), userId)
	if err != nil {
		RespondSuccess(c, 200, "Internal Server Error", nil)
		return
	}
	RespondSuccess(c, 200, "Success Getting Form Options", data)
}