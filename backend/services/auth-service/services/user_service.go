package services

import (
	"errors"
	"fmt"
	"log"
	"time"
	"unicode"

	"auth-service/config"
	"auth-service/models"
	"auth-service/repositories"

	"golang.org/x/crypto/bcrypt"
)

// Password policy errors returned by ValidatePasswordPolicy.
// Handlers should use IsPasswordPolicyError rather than comparing error strings.
var (
	ErrPasswordTooShort = fmt.Errorf(
		"password must be at least %d characters",
		config.PasswordMinLength,
	)
	ErrPasswordComplexity = errors.New(
		"password must contain at least one uppercase letter, one lowercase letter, and one digit",
	)
)

// IsPasswordPolicyError reports whether err is a password-policy validation failure.
func IsPasswordPolicyError(err error) bool {
	return errors.Is(err, ErrPasswordTooShort) || errors.Is(err, ErrPasswordComplexity)
}

// UserService handles all user-related operations.
type UserService struct {
	userRepo *repositories.UserRepository
}

// NewUserService creates a UserService backed by the given UserRepository.
func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// GetByEmail looks up a user by email. Returns (nil, nil) if not found.
func (s *UserService) GetByEmail(email string) (*models.User, error) {
	return s.userRepo.GetByEmail(email)
}

// GetByID looks up a user by UUID. Returns (nil, nil) if not found.
func (s *UserService) GetByID(id string) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

// Authenticate verifies email + password against the database.
// Returns the user if credentials are valid, nil if invalid.
// Does NOT check lockout — the caller should check user.IsLocked() first.
func (s *UserService) Authenticate(email, password string) (*models.User, error) {
	user, err := s.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if !s.CheckPassword(user, password) {
		return nil, nil
	}
	return user, nil
}

// CheckPassword verifies a plaintext password against an already-loaded user.
// Returns false when the user is nil, has no password, or the password does not match.
func (s *UserService) CheckPassword(user *models.User, password string) bool {
	if user == nil || user.PasswordHash == nil || *user.PasswordHash == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)) == nil
}

// IncrementFailedAttempts adds 1 to the failed login counter and returns the new count.
func (s *UserService) IncrementFailedAttempts(userID string) (int, error) {
	return s.userRepo.IncrementFailedAttempts(userID)
}

// LockAccount sets locked_until to NOW (UTC) + LockDuration.
func (s *UserService) LockAccount(userID string) error {
	lockUntil := time.Now().UTC().Add(config.LockDuration)
	return s.userRepo.UpdateLock(userID, lockUntil)
}

// ResetFailedAttempts clears the failed attempt counter and lock after a successful login.
func (s *UserService) ResetFailedAttempts(userID string) error {
	return s.userRepo.ResetFailedAttempts(userID)
}

// MarkAsVerified marks a user as verified in the database.
func (s *UserService) MarkAsVerified(userID string) error {
	return s.userRepo.MarkAsVerified(userID)
}

// UpdateRole updates a user's role and group assignment in the database.
func (s *UserService) UpdateRole(userID string, role string, groupID *string) error {
	return s.userRepo.UpdateRole(userID, role, groupID)
}

// GetAllUsers returns all non-master-admin users ordered by creation date.
func (s *UserService) GetAllUsers() ([]models.UserListItem, error) {
	return s.userRepo.GetAllUsers()
}

// UpdateStatus sets a user's account_status to 'active' or 'suspended'.
func (s *UserService) UpdateStatus(userID, status string) error {
	return s.userRepo.UpdateStatus(userID, status)
}

// DeleteUser permanently deletes a user record. Caller must validate the
// user is in 'invited' status before calling this.
func (s *UserService) DeleteUser(userID string) error {
	return s.userRepo.DeleteUser(userID)
}

// HashPassword produces a bcrypt hash of the given plaintext password.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), config.BcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// ValidatePasswordPolicy checks minimum length and character-class requirements.
func ValidatePasswordPolicy(password string) error {
	if len(password) < config.PasswordMinLength {
		return ErrPasswordTooShort
	}

	var hasUpper, hasLower, hasDigit bool
	for _, c := range password {
		if unicode.IsUpper(c) {
			hasUpper = true
		}
		if unicode.IsLower(c) {
			hasLower = true
		}
		if unicode.IsDigit(c) {
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return ErrPasswordComplexity
	}
	return nil
}

// SeedPasswords sets bcrypt-hashed passwords on seed users from migration 0002.
func (s *UserService) SeedPasswords() error {
	seeds := []struct {
		Email, Password string
	}{
		{config.MasterAdminEmail, config.MasterAdminPassword},
		{config.AdminEmail, config.AdminPassword},
		{config.StaffEmail, config.StaffPassword},
	}

	for _, seed := range seeds {
		user, err := s.GetByEmail(seed.Email)
		if err != nil {
			return err
		}
		if user == nil {
			log.Printf("⚠️  Seed user %s not found in database, skipping", seed.Email)
			continue
		}

		if user.PasswordHash != nil && *user.PasswordHash != "" {
			continue
		}

		hash, err := HashPassword(seed.Password)
		if err != nil {
			return err
		}

		if err := s.userRepo.UpdatePasswordHash(user.UserID, hash); err != nil {
			return fmt.Errorf("set password for %s: %w", seed.Email, err)
		}
		log.Printf("🔑 Set password for: %s (role: %s)", seed.Email, user.Role)
	}

	return nil
}

// InviteUser creates a new user in 'invited' status and returns a secure token.
func (s *UserService) InviteUser(email, role string, groupID *string) (string, error) {
	token, err := generateSecureToken()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(config.InvitationTokenExpiry)
	if err := s.userRepo.CreateInvitedUser(email, role, groupID, token, expiresAt); err != nil {
		return "", err
	}

	return token, nil
}

// ResendInvitation refreshes the invitation token for an invited user.
func (s *UserService) ResendInvitation(userID string) (string, error) {
	user, err := s.GetByID(userID)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", fmt.Errorf("user not found")
	}
	if user.AccountStatus != models.AccountStatusInvited {
		return "", fmt.Errorf("only invited users can have their invitation resent")
	}

	token, err := generateSecureToken()
	if err != nil {
		return "", fmt.Errorf("generate invitation token: %w", err)
	}
	expiresAt := time.Now().Add(config.InvitationTokenExpiry)

	if err := s.userRepo.UpdateInvitationToken(userID, token, expiresAt); err != nil {
		return "", err
	}

	return token, nil
}

// GetByInvitationToken finds a user by token and checks expiration.
func (s *UserService) GetByInvitationToken(token string) (*models.User, error) {
	user, err := s.userRepo.GetByInvitationToken(token)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	if user.TokenExpiresAt != nil && time.Now().After(*user.TokenExpiresAt) {
		return nil, fmt.Errorf("invitation token expired")
	}

	return user, nil
}

// CompleteInvitation sets the password and activates the user.
func (s *UserService) CompleteInvitation(userID, password string) error {
	if err := ValidatePasswordPolicy(password); err != nil {
		return err
	}

	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	return s.userRepo.CompleteInvitation(userID, hash)
}

// RequestPasswordReset generates a reset token for an active user with a password set.
// Returns the raw token and user when a reset email should be sent.
// Returns ("", nil, nil) when no action is taken (unknown/inactive/no-password user).
func (s *UserService) RequestPasswordReset(email string) (token string, user *models.User, err error) {
	user, err = s.GetByEmail(email)
	if err != nil {
		return "", nil, err
	}
	if user == nil {
		return "", nil, nil
	}
	if user.AccountStatus != models.AccountStatusActive {
		return "", nil, nil
	}
	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return "", nil, nil
	}

	token, err = generateSecureToken()
	if err != nil {
		return "", nil, fmt.Errorf("generate reset token: %w", err)
	}
	expiresAt := time.Now().Add(config.ResetTokenExpiry)

	if err := s.userRepo.SetResetToken(user.UserID, token, expiresAt); err != nil {
		return "", nil, err
	}

	return token, user, nil
}

// GetByResetToken finds an active user by reset token and checks expiration.
func (s *UserService) GetByResetToken(token string) (*models.User, error) {
	user, err := s.userRepo.GetByInvitationToken(token)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	if user.AccountStatus != models.AccountStatusActive {
		return nil, nil
	}
	if user.TokenExpiresAt != nil && time.Now().After(*user.TokenExpiresAt) {
		return nil, fmt.Errorf("reset token expired")
	}
	return user, nil
}

// CompletePasswordReset validates and stores a new password, clearing the reset token.
func (s *UserService) CompletePasswordReset(userID, password string) error {
	if err := ValidatePasswordPolicy(password); err != nil {
		return err
	}

	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	return s.userRepo.CompletePasswordReset(userID, hash)
}
