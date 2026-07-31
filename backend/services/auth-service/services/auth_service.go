package services

import (
	"fmt"
	"log"
	"time"

	"auth-service/config"
	"auth-service/models"
)

// Domain result status codes mapped to HTTP by handlers.
type ResultStatus int

const (
	StatusOK ResultStatus = iota
	StatusBadRequest
	StatusUnauthorized
	StatusForbidden
	StatusNotFound
	StatusConflict
	StatusTooManyRequests
	StatusInternalError
)

// Audit action / description constants (consistent wording).
const (
	AuditActionError   = "error"
	AuditActionLogin   = "login"
	AuditActionLoginFail = "login_fail"
	AuditActionLockout = "lockout"
	AuditActionUpdate  = "update"
	AuditActionInsert  = "insert"
	AuditActionDelete  = "delete"

	AuditDescLoginFailUnknown  = "login failed: unknown email"
	AuditDescLoginFailPassword = "login failed: invalid password"
	AuditDescLoginLocked       = "login failed: account locked"
	AuditDescLoginBlockedPref  = "login failed: account blocked ("
	AuditDescLockout           = "account locked due to too many failed attempts"
	AuditDescLoginSuccess      = "user logged in"
	AuditDescLoginOTPSuccess   = "user logged in via OTP"
	AuditDescOTPResend         = "otp resent"
)

// LoginResult is the outcome of AuthService.Login.
type LoginResult struct {
	Status       ResultStatus
	Message      string
	RequiresOTP  bool
	PreAuthToken string
	AccessToken  string
	RefreshToken string
	ExpiresIn    string
}

// VerifyOTPResult is the outcome of AuthService.VerifyOTPLogin.
type VerifyOTPResult struct {
	Status       ResultStatus
	Message      string
	AccessToken  string
	RefreshToken string
	ExpiresIn    string
}

// ResendOTPResult is the outcome of AuthService.ResendLoginOTP.
type ResendOTPResult struct {
	Status  ResultStatus
	Message string
}

// AuthService owns authentication use-cases (login, lockout, OTP branching).
type AuthService struct {
	users   UserManager
	otp     OTPManager
	jwt     TokenIssuer
	refresh RefreshManager
	audit   Auditor
}

// NewAuthService wires the dependencies required for auth flows.
func NewAuthService(
	users UserManager,
	otp OTPManager,
	jwt TokenIssuer,
	refresh RefreshManager,
	audit Auditor,
) *AuthService {
	return &AuthService{
		users:   users,
		otp:     otp,
		jwt:     jwt,
		refresh: refresh,
		audit:   audit,
	}
}

// Login authenticates email + password, applying lockout and OTP rules.
func (s *AuthService) Login(email, password string) LoginResult {
	user, err := s.users.GetByEmail(email)
	if err != nil {
		return LoginResult{Status: StatusInternalError, Message: "Internal server error."}
	}
	if user == nil {
		s.audit.Log(AuditActionError, AuditDescLoginFailUnknown, nil, config.AuditSourceClient)
		return LoginResult{Status: StatusUnauthorized, Message: "Invalid email or password."}
	}

	if user.IsLocked() {
		remaining := time.Until(*user.LockedUntil).Round(time.Second)
		s.audit.Log(AuditActionError, AuditDescLoginLocked, &user.UserID, config.AuditSourceClient)
		return LoginResult{
			Status:  StatusTooManyRequests,
			Message: fmt.Sprintf("Account is locked. Try again in %s.", remaining),
		}
	}

	if blockReason := user.LoginBlockReason(); blockReason != "" {
		s.audit.Log(AuditActionError, AuditDescLoginBlockedPref+user.AccountStatus+")", &user.UserID, config.AuditSourceClient)
		return LoginResult{Status: StatusForbidden, Message: blockReason}
	}

	if user.LockedUntil != nil && !user.IsLocked() {
		if err := s.users.ResetFailedAttempts(user.UserID); err != nil {
			log.Printf("⚠️  Failed to reset failed attempts for user %s: %v", user.UserID, err)
		}
	}

	if !s.users.CheckPassword(user, password) {
		return s.handleFailedLogin(user)
	}

	if err := s.users.ResetFailedAttempts(user.UserID); err != nil {
		log.Printf("⚠️  Failed to reset failed attempts for user %s: %v", user.UserID, err)
	}

	if requiresLoginOTP(user) {
		return s.startOTPLogin(user)
	}

	return s.completeLogin(user)
}

func (s *AuthService) handleFailedLogin(user *models.User) LoginResult {
	count, _ := s.users.IncrementFailedAttempts(user.UserID)
	s.audit.Log(AuditActionLoginFail, AuditDescLoginFailPassword, &user.UserID, config.AuditSourceClient)

	if count >= config.MaxLoginAttempts {
		if lockErr := s.users.LockAccount(user.UserID); lockErr != nil {
			log.Printf("⚠️  Failed to lock account for user %s: %v", user.UserID, lockErr)
		}
		if revokeErr := s.refresh.RevokeAllUserTokens(user.UserID); revokeErr != nil {
			log.Printf("⚠️  Failed to revoke tokens for user %s: %v", user.UserID, revokeErr)
		}
		s.audit.Log(AuditActionLockout, AuditDescLockout, &user.UserID, config.AuditSourceSystem)
		return LoginResult{
			Status: StatusTooManyRequests,
			Message: fmt.Sprintf(
				"Account locked for %s due to too many failed attempts.",
				config.LockDuration,
			),
		}
	}

	remaining := config.MaxLoginAttempts - count
	return LoginResult{
		Status:  StatusUnauthorized,
		Message: fmt.Sprintf("Invalid email or password. %d attempt(s) remaining.", remaining),
	}
}

func requiresLoginOTP(user *models.User) bool {
	return (user.Role == config.RoleAdmin || user.Role == config.RoleMasterAdmin) && !user.Verified
}

func (s *AuthService) startOTPLogin(user *models.User) LoginResult {
	preAuthToken, err := s.otp.GenerateAndStore(user.Email, config.OTPPurposeLogin)
	if err != nil {
		return LoginResult{Status: StatusInternalError, Message: "Failed to generate OTP."}
	}

	return LoginResult{
		Status:       StatusOK,
		Message:      "OTP has been sent to your email. Please verify to complete login.",
		RequiresOTP:  true,
		PreAuthToken: preAuthToken,
	}
}

func (s *AuthService) completeLogin(user *models.User) LoginResult {
	accessToken, refreshToken, err := s.issueTokens(user)
	if err != nil {
		return LoginResult{Status: StatusInternalError, Message: "Failed to generate tokens."}
	}

	s.audit.Log(AuditActionLogin, AuditDescLoginSuccess, &user.UserID, config.AuditSourceClient)
	return LoginResult{
		Status:       StatusOK,
		Message:      "Login successful.",
		RequiresOTP:  false,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    config.AccessTokenExpiry.String(),
	}
}

// VerifyOTPLogin completes admin/master-admin login after OTP verification.
func (s *AuthService) VerifyOTPLogin(email, otp, preAuthToken string) VerifyOTPResult {
	valid, errMsg, err := s.otp.Verify(email, otp, preAuthToken)
	if err != nil {
		return VerifyOTPResult{Status: StatusInternalError, Message: "Internal error during OTP verification."}
	}
	if !valid {
		return VerifyOTPResult{Status: StatusUnauthorized, Message: errMsg}
	}

	user, err := s.users.GetByEmail(email)
	if err != nil || user == nil {
		return VerifyOTPResult{Status: StatusInternalError, Message: "Failed to retrieve user information."}
	}

	if blockReason := user.LoginBlockReason(); blockReason != "" {
		return VerifyOTPResult{Status: StatusForbidden, Message: blockReason}
	}

	if !user.Verified {
		if err := s.users.MarkAsVerified(user.UserID); err != nil {
			log.Printf("⚠️  Failed to mark user %s as verified: %v", user.Email, err)
			return VerifyOTPResult{Status: StatusInternalError, Message: "Failed to update verification status."}
		}
		user.Verified = true
	}

	accessToken, refreshToken, err := s.issueTokens(user)
	if err != nil {
		return VerifyOTPResult{Status: StatusInternalError, Message: "Failed to generate tokens."}
	}

	s.audit.Log(AuditActionLogin, AuditDescLoginOTPSuccess, &user.UserID, config.AuditSourceClient)
	return VerifyOTPResult{
		Status:       StatusOK,
		Message:      "Login successful.",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    config.AccessTokenExpiry.String(),
	}
}

// ResendLoginOTP issues a new OTP when a login OTP is already pending.
func (s *AuthService) ResendLoginOTP(email string) ResendOTPResult {
	if !s.otp.HasPending(email) {
		return ResendOTPResult{
			Status:  StatusBadRequest,
			Message: "No pending OTP for this email. Please log in first.",
		}
	}

	if !s.otp.CanResend(email) {
		return ResendOTPResult{
			Status:  StatusTooManyRequests,
			Message: "Please wait before requesting a new OTP.",
		}
	}

	if _, err := s.otp.GenerateAndStore(email, config.OTPPurposeLogin); err != nil {
		return ResendOTPResult{Status: StatusInternalError, Message: "Failed to generate OTP."}
	}

	s.audit.Log(AuditActionUpdate, AuditDescOTPResend, nil, config.AuditSourceClient)
	return ResendOTPResult{
		Status:  StatusOK,
		Message: "A new OTP has been sent to your email.",
	}
}

func (s *AuthService) issueTokens(user *models.User) (accessToken, refreshToken string, err error) {
	accessToken, err = s.jwt.GenerateToken(user.UserID, user.Email, user.Role)
	if err != nil {
		return "", "", fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err = s.refresh.CreateToken(user.UserID)
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}
