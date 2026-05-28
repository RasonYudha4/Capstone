package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"auth-service/config"
	"auth-service/models"
	"auth-service/services"

	"github.com/gin-gonic/gin"
)

// groups the HTTP handlers for authentication endpoints.
type AuthHandler struct {
	userService    *services.UserService
	otpService     *services.OTPService
	jwtService     *services.JWTService
	refreshService *services.RefreshService
	auditService   *services.AuditService
}

// creates an AuthHandler with all required services.
func NewAuthHandler(
	userService *services.UserService,
	otpService *services.OTPService,
	jwtService *services.JWTService,
	refreshService *services.RefreshService,
	auditService *services.AuditService,
) *AuthHandler {
	return &AuthHandler{
		userService:    userService,
		otpService:     otpService,
		jwtService:     jwtService,
		refreshService: refreshService,
		auditService:   auditService,
	}
}

//  POST /auth/login

// login handles email + password authentication
// - account lockout after MaxLoginAttempts failed attempts
// - audit logging for login success, failure, and lockout
// - refresh token issued alongside access token (for staff)
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// look up the user first (to check lockout before bcrypt).
	user, err := h.userService.GetByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Internal server error.",
		})
		return
	}

	if user == nil {
		// audit: failed login (unknown email).
		h.auditService.Log("error", "login_fail", nil, "client")
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Invalid email or password.",
		})
		return
	}

	// phase 2: account lockout check
	if user.IsLocked() {
		remaining := time.Until(*user.LockedUntil).Round(time.Second)
		h.auditService.Log("error", "login_locked", &user.UserID, "client")
		c.JSON(http.StatusTooManyRequests, models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("Account is locked. Try again in %s.", remaining),
		})
		return
	}

	// if lock has expired, reset the counter.
	if user.LockedUntil != nil && !user.IsLocked() {
		if err := h.userService.ResetFailedAttempts(user.UserID); err != nil {
			log.Printf("⚠️  Failed to reset failed attempts for user %s: %v", user.UserID, err)
		}
	}

	// verify password.
	authenticated, err := h.userService.Authenticate(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Internal server error.",
		})
		return
	}

	if authenticated == nil {
		// wrong password, increment failed attempts.
		count, _ := h.userService.IncrementFailedAttempts(user.UserID)
		h.auditService.Log("error", "login_fail", &user.UserID, "client")

		if count >= config.MaxLoginAttempts {
			if lockErr := h.userService.LockAccount(user.UserID); lockErr != nil {
				log.Printf("⚠️  Failed to lock account for user %s: %v", user.UserID, lockErr)
			}
			if revokeErr := h.refreshService.RevokeAllUserTokens(user.UserID); revokeErr != nil {
				log.Printf("⚠️  Failed to revoke tokens for user %s: %v", user.UserID, revokeErr)
			}
			h.auditService.Log("error", "lockout", &user.UserID, "system")
			c.JSON(http.StatusTooManyRequests, models.APIResponse{
				Success: false,
				Message: fmt.Sprintf("Account locked for %s due to too many failed attempts.", config.LockDuration),
			})
			return
		}

		remaining := config.MaxLoginAttempts - count
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid email or password. %d attempt(s) remaining.", remaining),
		})
		return
	}

	// password correct, reset failed attempts.
	if err := h.userService.ResetFailedAttempts(user.UserID); err != nil {
		log.Printf("⚠️  Failed to reset failed attempts for user %s: %v", user.UserID, err)
	}

	// role-based login branching

	// for "staff" role: issue tokens immediately.
	if user.Role == config.RoleStaff {
		accessToken, refreshToken, err := h.issueTokens(user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "Failed to generate tokens.",
			})
			return
		}

		h.auditService.Log("insert", "login", &user.UserID, "client")
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Message: "Login successful.",
			Data: models.LoginResponse{
				RequiresOTP:  false,
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
				ExpiresIn:    config.AccessTokenExpiry.String(),
			},
		})
		return
	}

	// for "admin" and "master-admin", require OTP as a second factor.
	preAuthToken, err := h.otpService.GenerateAndStore(user.Email, "login")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to generate OTP.",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "OTP has been sent to your email. Please verify to complete login.",
		Data: models.LoginResponse{
			RequiresOTP: true,
			PreAuthToken: preAuthToken,
		},
	})
}

// POST /auth/verify-otp

// verifyotp handles the second step of admin/master-admin login.
// issues both access and refresh tokens, with audit logging.
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req models.OTPVerifyRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	valid, errMsg, err := h.otpService.Verify(req.Email, req.OTP, req.PreAuthToken)
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
			Message: errMsg,
		})
		return
	}

	// OTP is valid, look up user and issue tokens.
	user, err := h.userService.GetByEmail(req.Email)
	if err != nil || user == nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve user information.",
		})
		return
	}

	accessToken, refreshToken, err := h.issueTokens(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to generate tokens.",
		})
		return
	}

	h.auditService.Log("insert", "login", &user.UserID, "client")
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Login successful.",
		Data: models.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    config.AccessTokenExpiry.String(),
		},
	})
}

// POST /auth/resend-otp

// resendotp generates a new OTP for users who already passed password verification.
// users don't need to re-enter their password.
// security:
// - only works if an OTP was already requested via /auth/login (password was verified).
// - overwrites the previous OTP (the old one becomes invalid).
// - does NOT reveal whether the email exists (generic error on failure).
func (h *AuthHandler) ResendOTP(c *gin.Context) {
	var req models.ResendOTPRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// only allow resend if an OTP login was already initiated.
	if !h.otpService.HasPending(req.Email) {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "No pending OTP for this email. Please log in first.",
		})
		return
	}

	// rate-limit: enforce cooldown between resend requests.
	if !h.otpService.CanResend(req.Email) {
		c.JSON(http.StatusTooManyRequests, models.APIResponse{
			Success: false,
			Message: "Please wait before requesting a new OTP.",
		})
		return
	}

	// generate a fresh OTP (overwrites the old one).
	_, err := h.otpService.GenerateAndStore(req.Email, "login")
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to generate OTP.",
		})
		return
	}

	h.auditService.Log("update", "otp_resend", nil, "client")
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "A new OTP has been sent to your email.",
	})
}

// POST /auth/refresh

// refresh issues a new access token and rotates the refresh token.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req models.RefreshRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// rotate: revoke old token and issue a new one.
	newRefreshToken, userID, err := h.refreshService.RotateToken(req.RefreshToken)
	if err != nil {
		log.Printf("⚠️  Refresh token rotation failed: %v", err)
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Invalid or expired refresh token.",
		})
		return
	}

	// get user details for the new access token.
	user, err := h.userService.GetByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "User not found.",
		})
		return
	}

	// generate a new access token.
	accessToken, err := h.jwtService.GenerateToken(user.UserID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to generate access token.",
		})
		return
	}

	h.auditService.Log("update", "token_refresh", &user.UserID, "client")
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Token refreshed.",
		Data: models.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: newRefreshToken,
			ExpiresIn:    config.AccessTokenExpiry.String(),
		},
	})
}

// POST /auth/logout

// logout revokes a refresh token so it can no longer be used.
func (h *AuthHandler) Logout(c *gin.Context) {
	var req models.LogoutRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	userID, err := h.refreshService.RevokeToken(req.RefreshToken)
	if err != nil {
		log.Printf("⚠️  Token revocation failed: %v", err)
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid or expired token.",
		})
		return
	}

	h.auditService.Log("delete", "logout", &userID, "client")
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Logged out successfully.",
	})
}

// GET /auth/me

// me returns the authenticated user's profile information.
func (h *AuthHandler) Me(c *gin.Context) {
	value, exists := c.Get(config.ContextKeyUser)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Not authenticated.",
		})
		return
	}

	claims := value.(*services.Claims)

	user, err := h.userService.GetByID(claims.UserID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "User not found.",
		})
		return
	}

	resp := models.UserResponse{
		UserID: user.UserID,
		Email:  user.Email,
		Role:   user.Role,
	}
	if user.GroupID != nil {
		resp.GroupID = *user.GroupID
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "User info retrieved.",
		Data:    resp,
	})
}

// helpers

// generates both an access token (JWT) and a refresh token.
func (h *AuthHandler) issueTokens(user *models.User) (accessToken, refreshToken string, err error) {
	accessToken, err = h.jwtService.GenerateToken(user.UserID, user.Email, user.Role)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err = h.refreshService.CreateToken(user.UserID)
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}
