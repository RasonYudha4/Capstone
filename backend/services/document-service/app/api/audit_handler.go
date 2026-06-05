package api

import (
	"capstone/app/services"
	"net/http"

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

func (a *AuditHandler) GetAudit(r *gin.Context) {
	audit, err := a.service.Get_audit()
	if err != nil {
		r.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error"})
		return
	}

	r.JSON(http.StatusOK, gin.H{
		"data": audit,
	})
}