package models

// Request DTOs

// JSON body for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// for POST /auth/verify-otp.
type OTPVerifyRequest struct {
	Email        string `json:"email" binding:"required,email"`
	OTP          string `json:"otp"   binding:"required,len=6"`
	PreAuthToken string `json:"pre_auth_token" binding:"required"`
}

// for POST /auth/refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// for POST /auth/logout.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// for POST /auth/resend-otp.
type ResendOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// Response DTOs

// generic envelope for all JSON responses.
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"` // omitted when nil
}

// returned by POST /auth/login.
// if true, the client must proceed to /auth/verify-otp.
type LoginResponse struct {
	RequiresOTP  bool   `json:"requires_otp"`
	AccessToken  string `json:"access_token,omitempty"`
	PreAuthToken string `json:"pre_auth_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    string `json:"expires_in,omitempty"`
}

// carries the tokens returned after successful OTP verification.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    string `json:"expires_in"`
}

// carries the new access token from POST /auth/refresh.
type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   string `json:"expires_in"`
}

// public representation of a user (GET /auth/me).
type UserResponse struct {
	UserID  string `json:"user_id"`
	Email   string `json:"email"`
	Role    string `json:"role"`
	GroupID string `json:"group_id,omitempty"`
}
