package routes

import (
	"capstone/app/api"
	"capstone/app/middleware"

	"github.com/gin-gonic/gin"
)

func DocumentRoute(r *gin.Engine){
	{
	documentsRoute := r.Group("/")
	documentsRoute.Use(middleware.Extract_JWT_data("super-secret-key-change-in-production"))	
	documentsRoute.GET("/documents", middleware.AllowedRole(), api.Get_all_documents_Handler)
	documentsRoute.GET("/documents/:id", api.Get_document_by_id_handler)
	documentsRoute.GET("/documents/type/:type", api.Get_document_by_type_handler)
	documentsRoute.GET("/documents/groups/:group", api.Get_document_by_group_handler)
	documentsRoute.GET("/documents/standards/:standard", api.Get_document_by_standard_handler)
	documentsRoute.GET("/documents/createdBy/:createdBy", api.Get_document_by_createdBy_handler)
	documentsRoute.GET("/documents/assessments/:assessment", api.Get_document_by_assessment_handler)
	documentsRoute.GET("/documents/audit", api.GetAudit)
	documentsRoute.POST("/documents/upload", middleware.Extract_JWT_data("super-secret-key-change-in-production"), 	
	api.Create_document_handler)

	documentsRoute.GET("/masterAdmin/documents/status/:status", middleware.AllowedRole(), api.Get_document_by_status_handler)
	}

	
}