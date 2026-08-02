package handlers

import (
	"fmt"
	"log"
	"net/http"

	"auth-service/config"
	"auth-service/models"
	"auth-service/services"

	"github.com/gin-gonic/gin"
)

const forgotPasswordGenericMessage = "If an account with that email exists, a password reset link has been sent."

// AuthHandler groups the HTTP handlers for authentication endpoints.
type AuthHandler struct {
	auth           services.AuthFlow
	userService    services.UserManager
	groupService   services.GroupManager
	emailService   services.Mailer
	jwtService     services.TokenIssuer
	refreshService services.RefreshManager
	auditService   services.Auditor
}

// NewAuthHandler creates an AuthHandler with all required services.
func NewAuthHandler(
	auth services.AuthFlow,
	userService services.UserManager,
	groupService services.GroupManager,
	emailService services.Mailer,
	jwtService services.TokenIssuer,
	refreshService services.RefreshManager,
	auditService services.Auditor,
) *AuthHandler {
	return &AuthHandler{
		auth:           auth,
		userService:    userService,
		groupService:   groupService,
		emailService:   emailService,
		jwtService:     jwtService,
		refreshService: refreshService,
		auditService:   auditService,
	}
}

// Login handles POST /auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if !bindJSON(c, &req) {
		return
	}

	result := h.auth.Login(req.Email, req.Password)
	if result.Status != services.StatusOK {
		respondError(c, httpStatus(result.Status), result.Message)
		return
	}

	respondSuccess(c, result.Message, models.LoginResponse{
		RequiresOTP:  result.RequiresOTP,
		PreAuthToken: result.PreAuthToken,
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	})
}

// VerifyOTP handles POST /auth/verify-otp.
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req models.OTPVerifyRequest
	if !bindJSON(c, &req) {
		return
	}

	result := h.auth.VerifyOTPLogin(req.Email, req.OTP, req.PreAuthToken)
	if result.Status != services.StatusOK {
		respondError(c, httpStatus(result.Status), result.Message)
		return
	}

	respondSuccess(c, result.Message, models.TokenResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
	})
}

// ResendOTP handles POST /auth/resend-otp.
func (h *AuthHandler) ResendOTP(c *gin.Context) {
	var req models.ResendOTPRequest
	if !bindJSON(c, &req) {
		return
	}

	result := h.auth.ResendLoginOTP(req.Email)
	if result.Status != services.StatusOK {
		respondError(c, httpStatus(result.Status), result.Message)
		return
	}

	respondSuccess(c, result.Message, models.LoginResponse{
		RequiresOTP:  true,
		PreAuthToken: result.PreAuthToken,
	})
}

// Refresh handles POST /auth/refresh.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req models.RefreshRequest
	if !bindJSON(c, &req) {
		return
	}

	newRefreshToken, userID, err := h.refreshService.RotateToken(req.RefreshToken)
	if err != nil {
		log.Printf("⚠️  Refresh token rotation failed: %v", err)
		respondError(c, http.StatusUnauthorized, "Invalid or expired refresh token.")
		return
	}

	user, err := h.userService.GetByID(userID)
	if err != nil || user == nil {
		respondError(c, http.StatusUnauthorized, "User not found.")
		return
	}

	if blockReason := user.LoginBlockReason(); blockReason != "" {
		if revokeErr := h.refreshService.RevokeAllUserTokens(user.UserID); revokeErr != nil {
			log.Printf("⚠️  Failed to revoke tokens for inactive user %s: %v", user.UserID, revokeErr)
		}
		respondError(c, http.StatusUnauthorized, blockReason)
		return
	}
	if user.IsLocked() {
		if revokeErr := h.refreshService.RevokeAllUserTokens(user.UserID); revokeErr != nil {
			log.Printf("⚠️  Failed to revoke tokens for locked user %s: %v", user.UserID, revokeErr)
		}
		respondError(c, http.StatusTooManyRequests, "Account is locked. Try again later.")
		return
	}

	accessToken, err := h.jwtService.GenerateToken(user.UserID, user.Email, user.Role)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to generate access token.")
		return
	}

	h.auditService.Log(services.AuditActionUpdate, "token refreshed", &user.UserID, config.AuditSourceClient)
	respondSuccess(c, "Token refreshed.", models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    config.AccessTokenExpiry.String(),
	})
}

// Logout handles POST /auth/logout.
func (h *AuthHandler) Logout(c *gin.Context) {
	var req models.LogoutRequest
	if !bindJSON(c, &req) {
		return
	}

	userID, err := h.refreshService.RevokeToken(req.RefreshToken)
	if err != nil {
		log.Printf("⚠️  Token revocation failed: %v", err)
		respondError(c, http.StatusBadRequest, "Invalid or expired token.")
		return
	}

	h.auditService.Log(services.AuditActionDelete, "user logged out", &userID, config.AuditSourceClient)
	respondSuccess(c, "Logged out successfully.", nil)
}

// Me handles GET /auth/me.
func (h *AuthHandler) Me(c *gin.Context) {
	claims := requireClaims(c)
	if claims == nil {
		return
	}

	user, err := h.userService.GetByID(claims.UserID)
	if err != nil || user == nil {
		respondError(c, http.StatusNotFound, "User not found.")
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

	respondSuccess(c, "User info retrieved.", resp)
}

// ListUsers handles GET /auth/users (master-admin only).
func (h *AuthHandler) ListUsers(c *gin.Context) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to retrieve user list.")
		return
	}

	if users == nil {
		users = []models.UserListItem{}
	}

	respondSuccess(c, "User list retrieved.", models.ListUsersResponse{
		Users: users,
		Total: len(users),
	})
}

// UpdateRole handles PUT /auth/users/:id/role.
func (h *AuthHandler) UpdateRole(c *gin.Context) {
	targetID := c.Param("id")

	var req models.UpdateRoleRequest
	if !bindJSON(c, &req) {
		return
	}

	targetUser, ok := h.requireManageableTarget(c, targetID,
		"Cannot modify your own role.",
		"Cannot modify a master-admin's role.",
	)
	if !ok {
		return
	}

	groupID, ok := h.resolveAdminGroupID(c, req.Role, req.GroupID)
	if !ok {
		return
	}

	if err := h.userService.UpdateRole(targetUser.UserID, req.Role, groupID); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to update role.")
		return
	}

	// Invalidate existing sessions so clients must re-authenticate with the new role.
	if err := h.refreshService.RevokeAllUserTokens(targetUser.UserID); err != nil {
		log.Printf("⚠️  Failed to revoke tokens after role change for user %s: %v", targetUser.UserID, err)
	}

	h.auditService.Log(services.AuditActionUpdate, "role_changed_to_"+req.Role, &targetUser.UserID, config.AuditSourceSystem)
	respondSuccess(c, "User role updated to '"+req.Role+"' successfully.", nil)
}

// UpdateStatus handles PUT /auth/users/:id/status.
func (h *AuthHandler) UpdateStatus(c *gin.Context) {
	targetID := c.Param("id")

	var req models.UpdateStatusRequest
	if !bindJSON(c, &req) {
		return
	}

	targetUser, ok := h.requireManageableTarget(c, targetID,
		"Cannot modify your own account status.",
		"Cannot modify a master-admin's status.",
	)
	if !ok {
		return
	}

	if err := h.userService.UpdateStatus(targetUser.UserID, req.Status); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to update account status.")
		return
	}

	if req.Status == models.AccountStatusSuspended {
		if err := h.refreshService.RevokeAllUserTokens(targetUser.UserID); err != nil {
			log.Printf("⚠️  Failed to revoke tokens for suspended user %s: %v", targetUser.UserID, err)
		}
	}

	h.auditService.Log(services.AuditActionUpdate, "status_changed_to_"+req.Status, &targetUser.UserID, config.AuditSourceSystem)
	respondSuccess(c, "User account status updated to '"+req.Status+"' successfully.", nil)
}

// DeleteUser handles DELETE /auth/users/:id.
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	targetID := c.Param("id")

	targetUser, ok := h.requireManageableTarget(c, targetID,
		"Cannot delete your own account.",
		"Cannot delete a master-admin.",
	)
	if !ok {
		return
	}

	if targetUser.AccountStatus != models.AccountStatusInvited {
		respondError(c, http.StatusConflict,
			"Only users with 'invited' status can be deleted. Use status update to suspend active users.")
		return
	}

	if err := h.userService.DeleteUser(targetUser.UserID); err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to delete user.")
		return
	}

	h.auditService.Log(services.AuditActionDelete, "invited_user_deleted", &targetUser.UserID, config.AuditSourceSystem)
	respondSuccess(c, "Invited user deleted successfully.", nil)
}

// ResendInvitation handles POST /auth/users/:id/resend-invitation.
func (h *AuthHandler) ResendInvitation(c *gin.Context) {
	targetUser, ok := h.requireManageableTarget(c, c.Param("id"),
		"Cannot resend invitation to your own account.",
		"Cannot resend invitation for a master-admin.",
	)
	if !ok {
		return
	}

	if targetUser.AccountStatus != models.AccountStatusInvited {
		respondError(c, http.StatusConflict,
			"Only users with 'invited' status can have their invitation resent.")
		return
	}

	token, err := h.userService.ResendInvitation(targetUser.UserID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to resend invitation.")
		return
	}

	if config.IsDevMode {
		log.Printf("🔗 [DEV] Invitation link for %s: %s/setup-password?token=%s", targetUser.Email, config.FrontendURL, token)
	}

	h.sendInvitationEmailAsync(targetUser.Email, targetUser.Role, token)

	h.auditService.Log(services.AuditActionUpdate, "invitation_resent", &targetUser.UserID, config.AuditSourceSystem)
	respondSuccess(c, "Invitation resent successfully.", nil)
}

// Invite handles POST /auth/invite.
func (h *AuthHandler) Invite(c *gin.Context) {
	var req models.InviteRequest
	if !bindJSON(c, &req) {
		return
	}

	existing, _ := h.userService.GetByEmail(req.Email)
	if existing != nil {
		respondError(c, http.StatusConflict, "User with this email already exists.")
		return
	}

	groupID, ok := h.resolveAdminGroupID(c, req.Role, req.GroupID)
	if !ok {
		return
	}

	token, userID, err := h.userService.InviteUser(req.Email, req.Role, groupID)
	if err != nil {
		if services.IsUniqueViolation(err) {
			respondError(c, http.StatusConflict, "User with this email already exists.")
			return
		}
		respondError(c, http.StatusInternalServerError, "Failed to create invitation.")
		return
	}

	h.sendInvitationEmailAsync(req.Email, req.Role, token)

	h.auditService.Log(services.AuditActionInsert, "user_invited", &userID, config.AuditSourceSystem)
	respondSuccess(c, "Invitation sent successfully.", nil)
}

// CompleteInvitation handles POST /auth/complete-invitation.
func (h *AuthHandler) CompleteInvitation(c *gin.Context) {
	var req models.CompleteInvitationRequest
	if !bindJSON(c, &req) {
		return
	}

	user, err := h.userService.GetByInvitationToken(req.Token)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if user == nil || user.AccountStatus != models.AccountStatusInvited {
		respondError(c, http.StatusNotFound, "Invalid or already used invitation token.")
		return
	}

	if err := h.userService.CompleteInvitation(user.UserID, req.Password); err != nil {
		h.respondPasswordOrInternalError(c, err, "Failed to complete invitation: ")
		return
	}

	h.auditService.Log(services.AuditActionUpdate, "invitation_completed", &user.UserID, config.AuditSourceClient)
	respondSuccess(c, "Account setup complete. You can now log in.", nil)
}

// ForgotPassword handles POST /auth/forgot-password.
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if !bindJSON(c, &req) {
		return
	}

	token, user, err := h.userService.RequestPasswordReset(req.Email)
	if err != nil {
		log.Printf("⚠️  Failed to create password reset token for %s: %v", req.Email, err)
		respondSuccess(c, forgotPasswordGenericMessage, nil)
		return
	}

	if token != "" && user != nil {
		if err := h.refreshService.RevokeAllUserTokens(user.UserID); err != nil {
			log.Printf("⚠️  Failed to revoke tokens during password reset request for %s: %v", req.Email, err)
		}
		h.auditService.Log(services.AuditActionUpdate, "password_reset_requested", &user.UserID, config.AuditSourceClient)

		if config.IsDevMode {
			resetLink := fmt.Sprintf("%s/reset-password?token=%s", config.FrontendURL, token)
			log.Printf("🔗 [DEV] Password reset link for %s: %s", req.Email, resetLink)
		}

		go func(email, linkToken string) {
			if err := h.emailService.SendPasswordReset(email, linkToken); err != nil {
				log.Printf("⚠️  Failed to send password reset email to %s: %v", email, err)
			}
		}(req.Email, token)
	}

	respondSuccess(c, forgotPasswordGenericMessage, nil)
}

// ResetPassword handles POST /auth/reset-password.
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if !bindJSON(c, &req) {
		return
	}

	user, err := h.userService.GetByResetToken(req.Token)
	if err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	if user == nil {
		respondError(c, http.StatusNotFound, "Invalid or expired reset link.")
		return
	}

	if err := h.userService.CompletePasswordReset(user.UserID, req.Password); err != nil {
		h.respondPasswordOrInternalError(c, err, "Failed to reset password: ")
		return
	}

	if err := h.refreshService.RevokeAllUserTokens(user.UserID); err != nil {
		log.Printf("⚠️  Failed to revoke tokens after password reset for %s: %v", user.UserID, err)
	}

	go func(email string) {
		if err := h.emailService.SendPasswordChangedNotice(email); err != nil {
			log.Printf("⚠️  Failed to send password changed notice to %s: %v", email, err)
		}
	}(user.Email)

	h.auditService.Log(services.AuditActionUpdate, "password_reset_completed", &user.UserID, config.AuditSourceClient)
	respondSuccess(c, "Password reset successful. You can now log in.", nil)
}

// --- helpers ---

func (h *AuthHandler) loadTargetUser(c *gin.Context, targetID string) (*models.User, bool) {
	targetUser, err := h.userService.GetByID(targetID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Internal server error.")
		return nil, false
	}
	if targetUser == nil {
		respondError(c, http.StatusNotFound, "User not found.")
		return nil, false
	}
	return targetUser, true
}

func (h *AuthHandler) requireManageableTarget(
	c *gin.Context,
	targetID, selfForbiddenMsg, masterAdminForbiddenMsg string,
) (*models.User, bool) {
	claims := requireClaims(c)
	if claims == nil {
		return nil, false
	}

	if claims.UserID == targetID {
		respondError(c, http.StatusForbidden, selfForbiddenMsg)
		return nil, false
	}

	targetUser, ok := h.loadTargetUser(c, targetID)
	if !ok {
		return nil, false
	}

	if targetUser.Role == config.RoleMasterAdmin {
		respondError(c, http.StatusForbidden, masterAdminForbiddenMsg)
		return nil, false
	}

	return targetUser, true
}

func (h *AuthHandler) resolveAdminGroupID(c *gin.Context, role string, reqGroupID *string) (*string, bool) {
	if role != config.RoleAdmin {
		return nil, true
	}

	if reqGroupID == nil || *reqGroupID == "" {
		respondError(c, http.StatusBadRequest, "group_id is required when role is 'admin'.")
		return nil, false
	}

	exists, err := h.groupService.Exists(*reqGroupID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Failed to validate group.")
		return nil, false
	}
	if !exists {
		respondError(c, http.StatusBadRequest, "Invalid group_id.")
		return nil, false
	}

	return reqGroupID, true
}

func (h *AuthHandler) respondPasswordOrInternalError(c *gin.Context, err error, internalPrefix string) {
	if services.IsPasswordPolicyError(err) {
		respondError(c, http.StatusBadRequest, err.Error())
		return
	}
	respondError(c, http.StatusInternalServerError, internalPrefix+err.Error())
}

func (h *AuthHandler) sendInvitationEmailAsync(email, role, inviteToken string) {
	go func(email, role, inviteToken string) {
		if err := h.emailService.SendInvitation(email, role, inviteToken); err != nil {
			log.Printf("⚠️  Failed to send invitation email to %s: %v", email, err)
		}
	}(email, role, inviteToken)
}
