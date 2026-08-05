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
	userId := c.GetString("user_id")
	scope := c.Query("scope")

	var data any
	var err error

	if scope == "group" {
		// Filtered by user's group — used for upload form
		data, err = h.service.GetFormOptionsByGroup(c.Request.Context(), userId)
	} else {
		// All options — used for breadcrumbs / navigation
		data, err = h.service.GetFormOptions(c.Request.Context())
	}

	if err != nil {
		RespondError(c, 500, "Internal Server Error")
		return
	}
	RespondSuccess(c, 200, "Success Getting Form Options", data)
}