package main

import (
	"log"

	"auth-service/config"
	"auth-service/handlers"
	"auth-service/middleware"
	"auth-service/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// ── Database ──────────────────────────────────────────────────────
	services.InitDB()
	defer services.DB.Close()

	// ── Services ─────────────────────────────────────────────────────
	userService := services.NewUserService(services.DB)
	otpService := services.NewOTPService()
	jwtService := services.NewJWTService()

	// Seed sample users for development/testing.
	// In production, remove this and use a migration tool.
	if err := userService.SeedUsers(); err != nil {
		log.Fatalf("❌ Failed to seed users: %v", err)
	}

	// ── Handlers ─────────────────────────────────────────────────────
	authHandler := handlers.NewAuthHandler(userService, otpService, jwtService)
	documentHandler := handlers.NewDocumentHandler()

	// ── Router ───────────────────────────────────────────────────────
	router := gin.Default()

	// PUBLIC ROUTES — no authentication required.
	auth := router.Group("/auth")
	{
		// Step 1: email + password login.
		// Returns JWT directly for "user" role.
		// Returns requires_otp: true for "admin" and "master_admin".
		auth.POST("/login", authHandler.Login)

		// Step 2: OTP verification (only for admin / master_admin).
		// Completes the two-step login and returns a JWT.
		auth.POST("/verify-otp", authHandler.VerifyOTP)
	}

	// PROTECTED ROUTES — JWT authentication required.
	// All routes in this group pass through the JWTAuth middleware first.
	// The middleware validates the Bearer token and injects the user's claims
	// into the Gin context, making them available to all downstream handlers.
	protected := router.Group("/")
	protected.Use(middleware.JWTAuth(jwtService))
	{
		// Current user profile — accessible by ALL authenticated roles.
		protected.GET("/auth/me", authHandler.Me)

		// Document listing — accessible by ALL authenticated roles.
		protected.GET("/documents", documentHandler.ListDocuments)

		// Document upload — only "admin" and "master_admin" can upload.
		protected.POST("/upload",
			middleware.RequireRoles(config.RoleAdmin, config.RoleMasterAdmin),
			documentHandler.UploadDocument,
		)

		// Document approval — only "master_admin" can approve.
		// This is the most restricted endpoint in the system.
		protected.POST("/approve",
			middleware.RequireRoles(config.RoleMasterAdmin),
			documentHandler.ApproveDocument,
		)
	}

	// ── Start Server ─────────────────────────────────────────────────
	log.Printf("🚀 Auth service starting on port %s", config.ServerPort)
	if err := router.Run(config.ServerPort); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
