package api

import (
	"capstone/app/services"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	service *services.AuditService
}

func NewAuditHandler(service *services.AuditService) *AuditHandler {
	return &AuditHandler{
		service: service,
	}
}

func (a *AuditHandler) GetAudit(c *gin.Context) {
	audit, err := a.service.Get_audit()
	if err != nil {
		RespondSuccess(c, 500, "Internal Server Error", nil)
		return
	}

	RespondSuccess(c, 200, "Success Getting Audit log", audit)
}