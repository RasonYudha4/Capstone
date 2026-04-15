package models

// --- Request DTOs ---

// OTPRequest is the JSON body for POST /auth/request-otp.
type OTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// OTPVerifyRequest is the JSON body for POST /auth/verify-otp.
type OTPVerifyRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp"   binding:"required,len=6"`
}

// --- Response DTOs ---

// APIResponse is a generic envelope for all JSON responses.
// Using a consistent shape makes life easier for frontend consumers.
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"` // omitted when nil
}

// TokenResponse carries the JWT returned after successful OTP verification.
type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn string `json:"expires_in"`
}
