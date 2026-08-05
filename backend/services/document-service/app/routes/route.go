package routes

import (
	api "capstone/app/handler"
	"capstone/app/middleware"

	"github.com/gin-gonic/gin"
)

func DocumentRoute(r *gin.Engine, documentHandler *api.DocumentHandler, jwtSecret string) {

	documentsRoute := r.Group("/")

	// Static / multi-segment routes first so they never lose to /documents/:id.
	documentsRoute.GET(
		"/documents/public",
		documentHandler.Get_document_by_type_handler,
	)

	documentsRoute.GET(
		"/documents/public/:id",
		documentHandler.Get_public_document_by_id_handler,
	)

	documentsRoute.GET(
		"/documents/stats",
		middleware.Extract_JWT_data(jwtSecret),
		middleware.AllowedRole("master-admin", "admin"),
		documentHandler.GetStats,
	)

	documentsRoute.GET(
		"/documents/createdBy/my-document",
		middleware.Extract_JWT_data(jwtSecret),
		middleware.AllowedRole("master-admin", "admin"),
		documentHandler.Get_document_by_createdBy_handler,
	)

	documentsRoute.GET(
		"/documents/masterAdmin/status/:status",
		middleware.Extract_JWT_data(jwtSecret),
		middleware.AllowedRole("master-admin"),
		documentHandler.Get_document_by_status_handler,
	)

	documentsRoute.GET(
		"/documents/services/:service",
		middleware.Extract_JWT_data(jwtSecret),
		middleware.AllowedRole("master-admin", "admin"),
		documentHandler.Get_document_by_service_handler,
	)

	documentsRoute.GET(
		"/documents/standards/:standard",
		middleware.Extract_JWT_data(jwtSecret),
		middleware.AllowedRole("master-admin", "admin"),
		documentHandler.Get_document_by_standard_handler,
	)

	documentsRoute.GET(
		"/documents/assessments/:assessment",
		middleware.Extract_JWT_data(jwtSecret),
		middleware.AllowedRole("master-admin", "admin"),
		documentHandler.Get_document_by_assessment_handler,
	)

	// Param route last among GETs under /documents/*
	documentsRoute.GET(
		"/documents/:id",
		middleware.Extract_JWT_data(jwtSecret),
		middleware.AllowedRole("master-admin", "admin"),
		documentHandler.Get_document_by_id_handler,
	)

	documentsRoute.PATCH(
		"/documents/edit",
		middleware.Extract_JWT_data(jwtSecret),
		middleware.AllowedRole("master-admin", "admin"),
		documentHandler.Update_document_handler,
	)

	documentsRoute.POST(
		"/documents/upload",
		middleware.Extract_JWT_data(jwtSecret),
		middleware.AllowedRole("master-admin", "admin"),
		documentHandler.Create_document_handler,
	)

	documentsRoute.POST(
		"/documents/status/update",
		middleware.Extract_JWT_data(jwtSecret),
		middleware.AllowedRole("master-admin"),
		documentHandler.Approval_document_handler,
	)

	documentsRoute.DELETE(
		"/documents/delete/:documentId",
		middleware.Extract_JWT_data(jwtSecret),
		middleware.AllowedRole("master-admin", "admin"),
		documentHandler.Delete_document_handler,
	)
}

func AuditRoute(r *gin.Engine, auditHandler *api.AuditHandler, jwtSecret string) {

	r.GET(
		"/audit",
		middleware.Extract_JWT_data(jwtSecret),
		middleware.AllowedRole("master-admin"),
		auditHandler.GetAudit,
	)
}

func NotificationRoute(r *gin.Engine, notificationHandler *api.NotificationHandler, jwtSecret string) {

	notification := r.Group("/notifications", middleware.Extract_JWT_data(jwtSecret), middleware.AllowedRole("master-admin", "admin"))
	{
		notification.GET("/events", notificationHandler.SSEHandler)
		notification.GET("", notificationHandler.GetAll)
		notification.PATCH("/:id/read", notificationHandler.MarkRead)
		notification.PATCH("/read-all", notificationHandler.MarkAllRead)
	}
}

func FormOptionRoute(r *gin.Engine, formOptionHandler *api.FormOptionsHandler, jwtSecret string) {

	r.GET("/form-option", middleware.Extract_JWT_data(jwtSecret), formOptionHandler.GetFormOptions)
}
