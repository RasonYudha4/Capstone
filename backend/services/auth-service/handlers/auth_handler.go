package handlers

import (
	"net/http"

	"auth-service/config"
	"auth-service/models"
	"auth-service/services"

	"github.com/gin-gonic/gin"
)

// AuthHandler groups the HTTP handlers for authentication endpoints.
// Dependencies (OTP and JWT services) are injected through the constructor,
// which makes testing straightforward.
type AuthHandler struct {
	otpService *services.OTPService
	jwtService *services.JWTService
}

// NewAuthHandler creates an AuthHandler with all required services.
func NewAuthHandler(otpService *services.OTPService, jwtService *services.JWTService) *AuthHandler {
	return &AuthHandler{
		otpService: otpService,
		jwtService: jwtService,
	}
}

// RequestOTP handles POST /auth/request-otp
//
// Flow:
//  1. Parse & validate the request body (must contain a valid email).
//  2. Generate a secure OTP and store it with an expiration.
//  3. "Send" the OTP (logged to console for Phase 1).
//  4. Return a success message to the caller.
func (h *AuthHandler) RequestOTP(c *gin.Context) {
	var req models.OTPRequest

	// Bind and validate the JSON body.
	// Gin's binding tags (required, email) handle validation automatically.
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// Generate OTP, store it, and simulate sending.
	_, err := h.otpService.GenerateAndStore(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to generate OTP. Please try again.",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "OTP has been sent to " + req.Email,
	})
}

// VerifyOTP handles POST /auth/verify-otp
//
// Flow:
//  1. Parse & validate the request body (email + 6-digit OTP).
//  2. Look up the stored OTP for this email and verify it hasn't expired.
//  3. If valid → generate a JWT and return it.
//  4. If invalid or expired → return an appropriate error.
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req models.OTPVerifyRequest

	// Bind and validate the JSON body.
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// Verify the OTP against the in-memory store.
	valid, err := h.otpService.Verify(req.Email, req.OTP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Internal error during OTP verification.",
		})
		return
	}

	if !valid {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Invalid or expired OTP.",
		})
		return
	}

	// OTP is valid — issue a JWT.
	token, err := h.jwtService.GenerateToken(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to generate authentication token.",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Authentication successful.",
		Data: models.TokenResponse{
			Token:     token,
			ExpiresIn: config.JWTExpiration.String(),
		},
	})
}
