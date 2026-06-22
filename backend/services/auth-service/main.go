package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"auth-service/config"
	"auth-service/handlers"
	"auth-service/middleware"
	"auth-service/repositories"
	"auth-service/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func main() {
	// database
	services.InitDB()
	defer services.DB.Close()

	// repositories
	userRepo := repositories.NewUserRepository(services.DB)
	otpRepo := repositories.NewOTPRepository(services.DB)
	refreshRepo := repositories.NewRefreshRepository(services.DB)
	auditRepo := repositories.NewAuditRepository(services.DB)

	// services
	userService := services.NewUserService(userRepo)

	
	emailService := services.NewEmailService()
	otpService := services.NewOTPService(otpRepo, emailService)
	jwtService := services.NewJWTService()
	refreshService := services.NewRefreshService(refreshRepo) // Phase 2
	auditService := services.NewAuditService(auditRepo)       // Phase 2

	// Start OTP cleanup goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	otpService.StartCleanup(ctx)

	// set passwords on existing seed users 
	if err := userService.SeedPasswords(); err != nil {
		log.Printf("⚠️  Failed to seed passwords: %v", err)
	}
	authHandler := handlers.NewAuthHandler(
		userService, otpService, jwtService, refreshService, auditService,
	)
	documentHandler := handlers.NewDocumentHandler()

	// router
	router := gin.Default()

	// Rate limiter for sensitive auth endpoints (5 requests per minute, burst of 5)
	authRateLimiter := middleware.RateLimit(rate.Every(12*time.Second), 5)

	// public routes
	auth := router.Group("/auth")
	{
		// email + password login (staff → JWT, admin → OTP required).
		auth.POST("/login", authRateLimiter, authHandler.Login)

		// otp verification (completes admin/master-admin login).
		auth.POST("/verify-otp", authRateLimiter, authHandler.VerifyOTP)

		// resend OTP (only if login was already initiated via /auth/login).
		auth.POST("/resend-otp", authRateLimiter, authHandler.ResendOTP)

		// get a new access token using a refresh token.
		auth.POST("/refresh", authHandler.Refresh)

		// revoke a refresh token (logout).
		auth.POST("/logout", authHandler.Logout)

		// health check
		auth.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "UP"})
		})
	}

	// protected routes
	protected := router.Group("/")
	protected.Use(middleware.JWTAuth(jwtService))
	{
		// current user profile — accessible by ALL authenticated roles.
		protected.GET("/auth/me", authHandler.Me)

		// assign admin role — only "master-admin" can assign admin role.
		protected.POST("/auth/assign-admin",
			middleware.RequireRoles(config.RoleMasterAdmin),
			authHandler.AssignAdmin,
		)

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
