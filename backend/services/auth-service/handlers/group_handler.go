package handlers

import (
	"auth-service/models"
	"auth-service/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	groupService *services.GroupService
}

func NewGroupHandler(groupService *services.GroupService) *GroupHandler {
	return &GroupHandler{groupService: groupService}
}

// GET /groups
func (h *GroupHandler) ListGroups(c *gin.Context) {
	groups, err := h.groupService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to fetch groups.",
		})
		return
	}

	if groups == nil {
		groups = []models.Group{}
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Groups retrieved.",
		Data:    groups,
	})
}
