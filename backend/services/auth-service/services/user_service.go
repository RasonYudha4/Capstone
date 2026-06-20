package services

import (
	"fmt"
	"log"
	"time"
	"unicode"

	"auth-service/config"
	"auth-service/models"
	"auth-service/repositories"

	"golang.org/x/crypto/bcrypt"
)

// handles all user-related operations.
type UserService struct {
	userRepo *repositories.UserRepository
}

// creates a UserService backed by the given UserRepository.
func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// looks up a user by their email address.
// returns (nil, nil) if not found.
func (s *UserService) GetByEmail(email string) (*models.User, error) {
	return s.userRepo.GetByEmail(email)
}

// looks up a user by their UUID.
// returns (nil, nil) if not found.
func (s *UserService) GetByID(id string) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

//  Authentication

// verifies email + password against the database.
// returns the user if credentials are valid, nil if invalid.
// does NOT check lockout — the caller (handler) should check user.IsLocked() first.
func (s *UserService) Authenticate(email, password string) (*models.User, error) {
	user, err := s.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil // email not found
	}

	// check that the user has a password set.
	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return nil, nil // no password set for this user
	}

	// compare the plaintext password against the stored bcrypt hash.
	if err := bcrypt.CompareHashAndPassword(
		[]byte(*user.PasswordHash), []byte(password),
	); err != nil {
		return nil, nil // wrong password
	}

	return user, nil
}

//  account lockout

// adds 1 to the failed login counter.
// returns the new count.
func (s *UserService) IncrementFailedAttempts(userID string) (int, error) {
	return s.userRepo.IncrementFailedAttempts(userID)
}

// sets the locked_until timestamp to NOW (UTC) + LockDuration.
// uses UTC because the database column is TIMESTAMP (without time zone),
// and lib/pq reads it back as UTC.
func (s *UserService) LockAccount(userID string) error {
	lockUntil := time.Now().UTC().Add(config.LockDuration)
	return s.userRepo.UpdateLock(userID, lockUntil)
}

// clears the failed attempt counter and lock.
// called after a successful login.
func (s *UserService) ResetFailedAttempts(userID string) error {
	return s.userRepo.ResetFailedAttempts(userID)
}

// MarkAsVerified marks a user as verified in the database.
func (s *UserService) MarkAsVerified(userID string) error {
	return s.userRepo.MarkAsVerified(userID)
}

// UpdateRole updates a user's role in the database.
func (s *UserService) UpdateRole(userID string, role string) error {
	return s.userRepo.UpdateRole(userID, role)
}

// password hashing & validation

// produces a bcrypt hash of the given plaintext password.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), config.BcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// checks if a password meets the requirements:
//   - minimum 8 characters
//   - at least one uppercase letter
//   - at least one lowercase letter
//   - at least one digit
// returns nil if valid, or a descriptive error.
func ValidatePasswordPolicy(password string) error {
	if len(password) < config.PasswordMinLength {
		return fmt.Errorf("password must be at least %d characters", config.PasswordMinLength)
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
		return fmt.Errorf("password must contain at least one uppercase letter, one lowercase letter, and one digit")
	}
	return nil
}

//  seeding (for development / testing)

// sets bcrypt-hashed passwords on the existing seed users
// that were created by migration 0002_seed.up.sql.
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

		// Skip if password is already set.
		if user.PasswordHash != nil && *user.PasswordHash != "" {
			continue
		}

		hash, err := HashPassword(seed.Password)
		if err != nil {
			return err
		}

		err = s.userRepo.UpdatePasswordHash(user.UserID, hash)
		if err != nil {
			return fmt.Errorf("set password for %s: %w", seed.Email, err)
		}
		log.Printf("🔑 Set password for: %s (role: %s)", seed.Email, user.Role)
	}

	return nil
}
