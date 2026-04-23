package main

import (
	"capstone/app/core/db"
	"capstone/app/core/objectStorage"
	"capstone/app/repositories"
	"capstone/app/routes"
	"capstone/app/services"
	"capstone/app/api"
	"log"

	"github.com/gin-gonic/gin"
)


func main(){
	gin.SetMode(gin.DebugMode)
	r := gin.Default()
	r.HandleMethodNotAllowed = true 
	
	dbConn, err := db.InitDB()
	if err != nil{
		log.Fatal("Error init db: ", err)
	}
	objectStorageConn, err := objectStorage.InitMinio()
	if err != nil{
		log.Fatal("Error init objectStorage: ", err)
	}
	repo := repositories.NewDocumentRepo(dbConn)
	audit := repositories.NewAuditRepo(dbConn)
	storage := repositories.NewStorageRepo(objectStorageConn)
	service := services.NewDocumentService(repo, audit, storage)
	auditService := services.NewAuditService(audit)
	documentHandler := api.NewDocumentHandler(service)
	auditHandler := api.NewAuditHandler(auditService)

	routes.DocumentRoute(r,documentHandler)
	routes.AuditRoute(r,auditHandler)
	
	log.Fatal(r.Run(":8081"))	
}