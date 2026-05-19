package main

import (
	"log"
	"time"

	"auth-service/config"
	"auth-service/handlers"
	"auth-service/middleware"
	"auth-service/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// database
	services.InitDB()
	defer services.DB.Close()

	// services
	userService := services.NewUserService(services.DB)
	otpService := services.NewOTPService()
	jwtService := services.NewJWTService()
	refreshService := services.NewRefreshService(services.DB) // Phase 2
	auditService := services.NewAuditService(services.DB)     // Phase 2

	// set passwords on existing seed users 

	// handlers
	authHandler := handlers.NewAuthHandler(
		userService, otpService, jwtService, refreshService, auditService,
	)
	documentHandler := handlers.NewDocumentHandler()

	// router
	router := gin.Default()

	// CORS — restrict origins in production.
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// public routes
	auth := router.Group("/auth")
	{
		// email + password login (staff → JWT, admin → OTP required).
		auth.POST("/login", authHandler.Login)

		// otp verification (completes admin/master-admin login).
		auth.POST("/verify-otp", authHandler.VerifyOTP)

		// resend OTP (only if login was already initiated via /auth/login).
		auth.POST("/resend-otp", authHandler.ResendOTP)

		// get a new access token using a refresh token.
		auth.POST("/refresh", authHandler.Refresh)

		// revoke a refresh token (logout).
		auth.POST("/logout", authHandler.Logout)
	}

	// protected routes
	protected := router.Group("/")
	protected.Use(middleware.JWTAuth(jwtService))
	{
		// current user profile — accessible by ALL authenticated roles.
		protected.GET("/auth/me", authHandler.Me)

		// document listing — accessible by ALL authenticated roles.
		protected.GET("/documents", documentHandler.ListDocuments)

		// document upload — only "admin" and "master-admin" can upload.
		protected.POST("/upload",
			middleware.RequireRoles(config.RoleAdmin, config.RoleMasterAdmin),
			documentHandler.UploadDocument,
		)

		// document approval — only "master-admin" can approve.
		protected.POST("/approve",
			middleware.RequireRoles(config.RoleMasterAdmin),
			documentHandler.ApproveDocument,
		)
	}

	// start server
	log.Printf("Auth service starting on port %s", config.ServerPort)
	if err := router.Run(config.ServerPort); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
