package models

// =====================================================================
//  Request DTOs
// =====================================================================

// LoginRequest is the JSON body for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// OTPVerifyRequest is the JSON body for POST /auth/verify-otp.
type OTPVerifyRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp"   binding:"required,len=6"`
}

// =====================================================================
//  Response DTOs
// =====================================================================

// APIResponse is a generic envelope for all JSON responses.
// Using a consistent shape makes life easier for frontend consumers.
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"` // omitted when nil
}

// LoginResponse is returned by POST /auth/login.
// If RequiresOTP is true, the client must proceed to /auth/verify-otp.
type LoginResponse struct {
	RequiresOTP bool   `json:"requires_otp"`
	Token       string `json:"token,omitempty"`
	ExpiresIn   string `json:"expires_in,omitempty"`
}

// TokenResponse carries the JWT returned after successful OTP verification.
type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn string `json:"expires_in"`
}

// UserResponse is the public representation of a user (GET /auth/me).
type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}
