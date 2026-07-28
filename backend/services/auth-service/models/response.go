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

// for PUT /auth/users/:id/role.
// Role hanya boleh 'staff' atau 'admin'; 'master-admin' tidak bisa di-assign via API.
// group_id wajib diisi jika role = 'admin'.
type UpdateRoleRequest struct {
	Role    string  `json:"role" binding:"required,oneof=staff admin"`
	GroupID *string `json:"group_id"`
}

// for PUT /auth/users/:id/status.
type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active suspended"`
}

// for POST /auth/invite
type InviteRequest struct {
	Email   string  `json:"email" binding:"required,email"`
	Role    string  `json:"role"  binding:"required,oneof=admin staff"`
	GroupID *string `json:"group_id"`
}

// for POST /auth/complete-invitation
type CompleteInvitationRequest struct {
	Token    string `json:"token"    binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

// for POST /auth/forgot-password
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// for POST /auth/reset-password
type ResetPasswordRequest struct {
	Token    string `json:"token"    binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
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

// represents a single user row returned in GET /auth/users.
type UserListItem struct {
	UserID        string  `json:"user_id"`
	Email         string  `json:"email"`
	Role          string  `json:"role"`
	GroupID       *string `json:"group_id,omitempty"`
	GroupName     *string `json:"group_name,omitempty"`
	AccountStatus string  `json:"account_status"`
	Verified      bool    `json:"verified"`
}

// wraps the user list returned by GET /auth/users.
type ListUsersResponse struct {
	Users []UserListItem `json:"users"`
	Total int            `json:"total"`
}
