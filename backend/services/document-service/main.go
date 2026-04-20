package main

import (
	"capstone/app/core/db"
	"capstone/app/core/objectStorage"
	"capstone/app/routes"
	"log"
	"github.com/gin-gonic/gin"
)


func main(){
	gin.SetMode(gin.DebugMode)
	r := gin.Default()
	r.HandleMethodNotAllowed = true 
	
	db.InitDB()
	objectStorage.InitMinio()
	routes.DocumentRoute(r)
	
	log.Fatal(r.Run(":8081"))	
}