package main

import (
	"capstone/app/core/db"
	"capstone/app/core/objectStorage"
	"capstone/app/handler"
	"capstone/app/repositories"
	"capstone/app/routes"
	"capstone/app/services"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("❌ JWT_SECRET environment variable is required")
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.HandleMethodNotAllowed = true
	r.RedirectTrailingSlash = false

	dbConn, err := db.InitDB()
	if err != nil {
		log.Fatal("Error init db: ", err)
	}
	objectStorageConn, err := objectStorage.InitMinio()
	if err != nil {
		log.Fatal("Error init objectStorage: ", err)
	}

	if err := objectStorage.CreateBuckets(objectStorageConn); err != nil {
		log.Fatal("Error creating buckets: ", err)
	}

	repo := repositories.NewDocumentRepo(dbConn)
	audit := repositories.NewAuditRepo(dbConn)
	storage := repositories.NewStorageRepo(objectStorageConn)
	notificationRepo := repositories.NewNotificationRepository(dbConn)
	formOptionRepo := repositories.NewFormOptionsRepository(dbConn)

	notification := services.NewNotificationService(notificationRepo)
	service := services.NewDocumentService(repo, audit, storage, notification)
	auditService := services.NewAuditService(audit)
	formOptionService := services.NewFormOptionsService(formOptionRepo)

	documentHandler := api.NewDocumentHandler(service, notification)
	auditHandler := api.NewAuditHandler(auditService)
	notificationHandler := api.NewNotificationHandler(notification, notificationRepo)

	formOptionHandler := api.NewFormOptionsHandler(formOptionService)

	routes.DocumentRoute(r, documentHandler, jwtSecret)
	routes.AuditRoute(r, auditHandler, jwtSecret)

	routes.NotificationRoute(r, notificationHandler, jwtSecret)
	routes.FormOptionRoute(r, formOptionHandler, jwtSecret)

	log.Fatal(r.Run(":8081"))

}
