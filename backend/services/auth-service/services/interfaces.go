package services

import "auth-service/models"

// Interfaces describe service contracts so callers (handlers, middleware, tests)
// can depend on behavior rather than concrete types.

// UserLookup is the subset of user operations needed by JWT middleware.
type UserLookup interface {
	GetByID(id string) (*models.User, error)
}

// TokenValidator validates access tokens.
type TokenValidator interface {
	ValidateToken(tokenString string) (*Claims, error)
}

// UserManager covers user persistence operations used by handlers and AuthService.
type UserManager interface {
	GetByEmail(email string) (*models.User, error)
	GetByID(id string) (*models.User, error)
	CheckPassword(user *models.User, password string) bool
	IncrementFailedAttempts(userID string) (int, error)
	LockAccount(userID string) error
	ResetFailedAttempts(userID string) error
	MarkAsVerified(userID string) error
	UpdateRole(userID string, role string, groupID *string) error
	GetAllUsers() ([]models.UserListItem, error)
	UpdateStatus(userID, status string) error
	DeleteUser(userID string) error
	InviteUser(email, role string, groupID *string) (rawToken, userID string, err error)
	ResendInvitation(userID string) (string, error)
	GetByInvitationToken(token string) (*models.User, error)
	CompleteInvitation(userID, password string) error
	RequestPasswordReset(email string) (token string, user *models.User, err error)
	GetByResetToken(token string) (*models.User, error)
	CompletePasswordReset(userID, password string) error
}

// GroupManager looks up admin groups.
type GroupManager interface {
	GetAll() ([]models.Group, error)
	Exists(groupID string) (bool, error)
}

// OTPManager generates and verifies one-time passwords.
type OTPManager interface {
	GenerateAndStore(email, purpose string) (string, error)
	Verify(email, code, preAuthToken string) (valid bool, errMsg string, err error)
	HasPending(email string) bool
	CanResend(email string) bool
}

// TokenIssuer creates access tokens.
type TokenIssuer interface {
	GenerateToken(userID, email, role string) (string, error)
}

// RefreshManager creates and revokes refresh tokens.
type RefreshManager interface {
	CreateToken(userID string) (string, error)
	RevokeToken(rawToken string) (userID string, err error)
	RevokeAllUserTokens(userID string) error
	RotateToken(rawToken string) (newRawToken string, userID string, err error)
}

// Auditor records authentication and admin events.
type Auditor interface {
	Log(action, description string, userID *string, source string)
}

// Mailer sends transactional auth emails.
type Mailer interface {
	SendInvitation(to, role, token string) error
	SendPasswordReset(to, token string) error
	SendPasswordChangedNotice(to string) error
}

// AuthFlow is the login / OTP authentication use-case surface.
type AuthFlow interface {
	Login(email, password string) LoginResult
	VerifyOTPLogin(email, otp, preAuthToken string) VerifyOTPResult
	ResendLoginOTP(email string) ResendOTPResult
}
