package api

import(
	"capstone/app/services"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	service *services.AuditService
}

func NewAuditHandler(service *services.AuditService) *AuditHandler{
	return &AuditHandler {
		service: service,
	}
}

func (a *AuditHandler)GetAudit(r *gin.Context){
	audit, err := a.service.Get_audit()
	if err != nil{
		r.JSON(500, "Internal error")
		return
	}

	r.JSON(200, audit)
}