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
	groupService   *services.GroupService
	otpService     *services.OTPService
	emailService   *services.EmailService
	jwtService     *services.JWTService
	refreshService *services.RefreshService
	auditService   *services.AuditService
}

// creates an AuthHandler with all required services.
func NewAuthHandler(
	userService *services.UserService,
	groupService *services.GroupService,
	otpService *services.OTPService,
	emailService *services.EmailService,
	jwtService *services.JWTService,
	refreshService *services.RefreshService,
	auditService *services.AuditService,
) *AuthHandler {
	return &AuthHandler{
		userService:    userService,
		groupService:   groupService,
		otpService:     otpService,
		emailService:   emailService,
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
		h.auditService.Log("login_fail","user login error" ,&user.UserID, "client")

		if count >= config.MaxLoginAttempts {
			if lockErr := h.userService.LockAccount(user.UserID); lockErr != nil {
				log.Printf("⚠️  Failed to lock account for user %s: %v", user.UserID, lockErr)
			}
			if revokeErr := h.refreshService.RevokeAllUserTokens(user.UserID); revokeErr != nil {
				log.Printf("⚠️  Failed to revoke tokens for user %s: %v", user.UserID, revokeErr)
			}
			h.auditService.Log("lockout", "Error lock out, user got lockout ", &user.UserID, "system")
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
	// Admin and Master Admin require OTP verification if they are not verified yet.
	// Staff role bypasses OTP verification entirely (optional verification).
	if (user.Role == config.RoleAdmin || user.Role == config.RoleMasterAdmin) && !user.Verified {
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
				RequiresOTP:  true,
				PreAuthToken: preAuthToken,
			},
		})
		return
	}

	// For verified admins/master-admins and all staff members: log in immediately.
	accessToken, refreshToken, err := h.issueTokens(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to generate tokens.",
		})
		return
	}

	h.auditService.Log("login", "User has login", &user.UserID, "client")
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

	// If the user has not been marked as verified yet, mark them as verified now!
	if !user.Verified {
		if err := h.userService.MarkAsVerified(user.UserID); err != nil {
			log.Printf("⚠️  Failed to mark user %s as verified: %v", user.Email, err)
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "Failed to update verification status.",
			})
			return
		}
		user.Verified = true // update local state
	}

	accessToken, refreshToken, err := h.issueTokens(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to generate tokens.",
		})
		return
	}

	h.auditService.Log("login", "user has otp and enter dashboard page", &user.UserID, "client")
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

	h.auditService.Log("delete", "user logout", &userID, "client")
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

// GET /auth/users

// ListUsers returns all non-master-admin users.
// Accessible only by master-admin.
func (h *AuthHandler) ListUsers(c *gin.Context) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve user list.",
		})
		return
	}

	// Return empty array instead of null when no users found.
	if users == nil {
		users = []models.UserListItem{}
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "User list retrieved.",
		Data: models.ListUsersResponse{
			Users: users,
			Total: len(users),
		},
	})
}

// PUT /auth/users/:id/role

// UpdateRole changes the role of a target user to 'staff' or 'admin'.
// Security rules:
//   - Only master-admin can call this.
//   - Cannot target another master-admin.
//   - Cannot target themselves.
func (h *AuthHandler) UpdateRole(c *gin.Context) {
	targetID := c.Param("id")

	var req models.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// Get caller identity from JWT claims.
	claims := h.callerClaims(c)
	if claims == nil {
		return // response already written by callerClaims
	}

	// Guard: cannot modify self.
	if claims.UserID == targetID {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Message: "Cannot modify your own role.",
		})
		return
	}

	// Verify target exists.
	targetUser, err := h.userService.GetByID(targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Internal server error.",
		})
		return
	}
	if targetUser == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "User not found.",
		})
		return
	}

	// Guard: cannot target master-admin.
	if targetUser.Role == config.RoleMasterAdmin {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Message: "Cannot modify a master-admin's role.",
		})
		return
	}

	// Validate group assignment for admin role.
	var groupID *string
	if req.Role == config.RoleAdmin {
		if req.GroupID == nil || *req.GroupID == "" {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: "group_id is required when role is 'admin'.",
			})
			return
		}
		exists, err := h.groupService.Exists(*req.GroupID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "Failed to validate group.",
			})
			return
		}
		if !exists {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: "Invalid group_id.",
			})
			return
		}
		groupID = req.GroupID
	}

	// Perform role update.
	if err := h.userService.UpdateRole(targetUser.UserID, req.Role, groupID); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to update role.",
		})
		return
	}

	h.auditService.Log("update", "role_changed_to_"+req.Role, &targetUser.UserID, "system")
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "User role updated to '" + req.Role + "' successfully.",
	})
}

// PUT /auth/users/:id/status

// UpdateStatus suspends or re-activates a target user.
// Security rules:
//   - Only master-admin can call this.
//   - Cannot target another master-admin.
//   - Cannot target themselves.
func (h *AuthHandler) UpdateStatus(c *gin.Context) {
	targetID := c.Param("id")

	var req models.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// Get caller identity from JWT claims.
	claims := h.callerClaims(c)
	if claims == nil {
		return
	}

	// Guard: cannot modify self.
	if claims.UserID == targetID {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Message: "Cannot modify your own account status.",
		})
		return
	}

	// Verify target exists.
	targetUser, err := h.userService.GetByID(targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Internal server error.",
		})
		return
	}
	if targetUser == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "User not found.",
		})
		return
	}

	// Guard: cannot target master-admin.
	if targetUser.Role == config.RoleMasterAdmin {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Message: "Cannot modify a master-admin's status.",
		})
		return
	}

	if err := h.userService.UpdateStatus(targetUser.UserID, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to update account status.",
		})
		return
	}

	// If suspending, revoke all active refresh tokens immediately.
	if req.Status == "suspended" {
		if err := h.refreshService.RevokeAllUserTokens(targetUser.UserID); err != nil {
			log.Printf("⚠️  Failed to revoke tokens for suspended user %s: %v", targetUser.UserID, err)
		}
	}

	h.auditService.Log("update", "status_changed_to_"+req.Status, &targetUser.UserID, "system")
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "User account status updated to '" + req.Status + "' successfully.",
	})
}

// DELETE /auth/users/:id

// DeleteUser permanently removes a user who has not yet accepted their invitation.
// Only users with account_status = 'invited' may be deleted.
// Active or suspended users must be managed via UpdateStatus.
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	targetID := c.Param("id")

	// Get caller identity from JWT claims.
	claims := h.callerClaims(c)
	if claims == nil {
		return
	}

	// Guard: cannot delete self.
	if claims.UserID == targetID {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Message: "Cannot delete your own account.",
		})
		return
	}

	// Verify target exists.
	targetUser, err := h.userService.GetByID(targetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Internal server error.",
		})
		return
	}
	if targetUser == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "User not found.",
		})
		return
	}

	// Guard: cannot target master-admin.
	if targetUser.Role == config.RoleMasterAdmin {
		c.JSON(http.StatusForbidden, models.APIResponse{
			Success: false,
			Message: "Cannot delete a master-admin.",
		})
		return
	}

	// Guard: only 'invited' users may be deleted.
	if targetUser.AccountStatus != "invited" {
		c.JSON(http.StatusConflict, models.APIResponse{
			Success: false,
			Message: "Only users with 'invited' status can be deleted. Use status update to suspend active users.",
		})
		return
	}

	if err := h.userService.DeleteUser(targetUser.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to delete user.",
		})
		return
	}

	h.auditService.Log("delete", "invited_user_deleted", &targetUser.UserID, "system")
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Invited user deleted successfully.",
	})
}

// POST /auth/invite
func (h *AuthHandler) Invite(c *gin.Context) {
	var req models.InviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// 1. Check if user already exists
	existing, _ := h.userService.GetByEmail(req.Email)
	if existing != nil {
		c.JSON(http.StatusConflict, models.APIResponse{
			Success: false,
			Message: "User with this email already exists.",
		})
		return
	}

	// Validate group assignment for admin invites.
	var groupID *string
	if req.Role == config.RoleAdmin {
		if req.GroupID == nil || *req.GroupID == "" {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: "group_id is required when role is 'admin'.",
			})
			return
		}
		exists, err := h.groupService.Exists(*req.GroupID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "Failed to validate group.",
			})
			return
		}
		if !exists {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: "Invalid group_id.",
			})
			return
		}
		groupID = req.GroupID
	}

	// 2. Create invitation in DB
	token, err := h.userService.InviteUser(req.Email, req.Role, groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to create invitation: " + err.Error(),
		})
		return
	}

	// 3. Send email
	err = h.emailService.SendInvitation(req.Email, req.Role, token)
	if err != nil {
		log.Printf("⚠️  Failed to send invitation email to %s: %v", req.Email, err)
		// We don't fail the request because the user is already created in DB.
		// Admin might need a way to resend.
	}

	h.auditService.Log("insert", "user_invited", nil, "system")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Invitation sent successfully.",
	})
}

// POST /auth/complete-invitation
func (h *AuthHandler) CompleteInvitation(c *gin.Context) {
	var req models.CompleteInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// 1. Find user by token
	user, err := h.userService.GetByInvitationToken(req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "Invalid or already used invitation token.",
		})
		return
	}
	if user.AccountStatus != "invited" {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "Invalid or already used invitation token.",
		})
		return
	}

	// 2. Complete setup
	err = h.userService.CompleteInvitation(user.UserID, req.Password)
	if err != nil {
		// Check if the error is a password policy validation error
		if err.Error() == "password must be at least 8 characters" || err.Error() == "password must contain at least one uppercase letter, one lowercase letter, and one digit" {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to complete invitation: " + err.Error(),
		})
		return
	}

	h.auditService.Log("update", "invitation_completed", &user.UserID, "client")

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Account setup complete. You can now log in.",
	})
}

// POST /auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	const genericMessage = "If an account with that email exists, a password reset link has been sent."

	token, err := h.userService.RequestPasswordReset(req.Email)
	if err != nil {
		log.Printf("⚠️  Failed to create password reset token for %s: %v", req.Email, err)
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Message: genericMessage,
		})
		return
	}

	if token != "" {
		user, _ := h.userService.GetByEmail(req.Email)
		if user != nil {
			if err := h.refreshService.RevokeAllUserTokens(user.UserID); err != nil {
				log.Printf("⚠️  Failed to revoke tokens during password reset request for %s: %v", req.Email, err)
			}
			h.auditService.Log("update", "password_reset_requested", &user.UserID, "client")
		}

		if err := h.emailService.SendPasswordReset(req.Email, token); err != nil {
			log.Printf("⚠️  Failed to send password reset email to %s: %v", req.Email, err)
		}
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: genericMessage,
	})
}

// POST /auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	user, err := h.userService.GetByResetToken(req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "Invalid or expired reset link.",
		})
		return
	}

	err = h.userService.CompletePasswordReset(user.UserID, req.Password)
	if err != nil {
		if err.Error() == "password must be at least 8 characters" || err.Error() == "password must contain at least one uppercase letter, one lowercase letter, and one digit" {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to reset password: " + err.Error(),
		})
		return
	}

	if err := h.refreshService.RevokeAllUserTokens(user.UserID); err != nil {
		log.Printf("⚠️  Failed to revoke tokens after password reset for %s: %v", user.UserID, err)
	}

	if err := h.emailService.SendPasswordChangedNotice(user.Email); err != nil {
		log.Printf("⚠️  Failed to send password changed notice to %s: %v", user.Email, err)
	}

	h.auditService.Log("update", "password_reset_completed", &user.UserID, "client")
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Password reset successful. You can now log in.",
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

// callerClaims retrieves the authenticated caller's JWT claims from the Gin context.
// Returns nil and writes a 401 response if claims are not found (misconfiguration guard).
func (h *AuthHandler) callerClaims(c *gin.Context) *services.Claims {
	value, exists := c.Get(config.ContextKeyUser)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Not authenticated.",
		})
		return nil
	}
	claims, ok := value.(*services.Claims)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to read user claims.",
		})
		return nil
	}
	return claims
}

