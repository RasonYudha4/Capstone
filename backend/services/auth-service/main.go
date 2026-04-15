package main

import (
	"log"

	"auth-service/config"
	"auth-service/handlers"
	"auth-service/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// --- Initialise services ---
	// OTPService: generates, stores, and verifies one-time passwords.
	otpService := services.NewOTPService()
	// JWTService: creates signed JSON Web Tokens after successful authentication.
	jwtService := services.NewJWTService()

	// --- Initialise handler with injected dependencies ---
	authHandler := handlers.NewAuthHandler(otpService, jwtService)

	// --- Set up Gin router ---
	router := gin.Default()

	// Auth routes — grouped under /auth for clarity.
	auth := router.Group("/auth")
	{
		// POST /auth/request-otp  → generates and "sends" an OTP for the given email.
		auth.POST("/request-otp", authHandler.RequestOTP)

		// POST /auth/verify-otp   → validates the OTP and returns a JWT on success.
		auth.POST("/verify-otp", authHandler.VerifyOTP)
	}

	// --- Start server ---
	log.Printf("🚀 Auth service starting on port %s", config.ServerPort)
	if err := router.Run(config.ServerPort); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
