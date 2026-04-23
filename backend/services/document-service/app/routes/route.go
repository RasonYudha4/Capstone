package routes

import (
	"capstone/app/api"
	"capstone/app/middleware"

	"github.com/gin-gonic/gin"
)

func DocumentRoute(r *gin.Engine, documentHandler *api.DocumentHandler) {

	documentsRoute := r.Group("/")

	documentsRoute.GET("/documents", documentHandler.Get_all_documents_Handler)
	documentsRoute.GET("/documents/:id", documentHandler.Get_document_by_id_handler)
	documentsRoute.GET("/documents/type/:type", documentHandler.Get_document_by_type_handler)
	documentsRoute.GET("/documents/groups/:group", documentHandler.Get_document_by_group_handler)
	documentsRoute.GET("/documents/standards/:standard", documentHandler.Get_document_by_standard_handler)
	documentsRoute.GET("/documents/createdBy/:createdBy", documentHandler.Get_document_by_createdBy_handler)
	documentsRoute.GET("/documents/assessments/:assessment", documentHandler.Get_document_by_assessment_handler)

	documentsRoute.GET(
		"/masterAdmin/documents/status/:status",
		middleware.Extract_JWT_data("super-secret-key-change-in-production"),
		middleware.AllowedRole("master-admin"),
		documentHandler.Get_document_by_status_handler,
	)

	documentsRoute.PATCH(
		"/documents/update",
		middleware.Extract_JWT_data("super-secret-key-change-in-production"),
		middleware.AllowedRole("master-admin", "admin"),
		documentHandler.Update_document_handler,
	)

	documentsRoute.POST(
		"/documents/upload",
		middleware.Extract_JWT_data("super-secret-key-change-in-production"),
		middleware.AllowedRole("master-admin","admin"),
		documentHandler.Create_document_handler,
	)

	documentsRoute.POST(
		"/documents/status/update",
		middleware.Extract_JWT_data("super-secret-key-change-in-production"),
		middleware.AllowedRole("master-admin"),
		documentHandler.Approval_document_handler,
	)

	documentsRoute.DELETE(
		"/documents/delete",
		middleware.Extract_JWT_data("super-secret-key-change-in-production"),
		middleware.AllowedRole("master-admin"),
		documentHandler.Delete_document_handler,
	)
}

func AuditRoute(r *gin.Engine, auditHandler *api.AuditHandler) {

	r.GET(
		"/audit",
		middleware.Extract_JWT_data("super-secret-key-change-in-production"),
		middleware.AllowedRole("master-admin"),
		auditHandler.GetAudit,
	)
}