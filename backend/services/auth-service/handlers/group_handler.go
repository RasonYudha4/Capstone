package handlers

import (
	"net/http"

	"auth-service/models"
	"auth-service/services"

	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	groupService *services.GroupService
}

func NewGroupHandler(groupService *services.GroupService) *GroupHandler {
	return &GroupHandler{groupService: groupService}
}

// ListGroups handles GET /groups.
func (h *GroupHandler) ListGroups(c *gin.Context) {
	groups, err := h.groupService.GetAll()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to fetch groups.")
		return
	}

	if groups == nil {
		groups = []models.Group{}
	}

	respondSuccess(c, "Groups retrieved.", groups)
}
