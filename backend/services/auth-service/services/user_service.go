package services

import (
	"database/sql"
	"fmt"
	"log"

	"auth-service/config"
	"auth-service/models"

	"golang.org/x/crypto/bcrypt"
)

// UserService handles all user-related database operations.
type UserService struct {
	db *sql.DB
}

// NewUserService creates a UserService backed by the given database.
func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

// =====================================================================
//  Queries
// =====================================================================

// GetByEmail looks up a user by their email address.
// Returns (nil, nil) if not found — not an error, just "no match".
func (s *UserService) GetByEmail(email string) (*models.User, error) {
	user := &models.User{}
	err := s.db.QueryRow(
		`SELECT id, email, password_hash, name, role, is_active, created_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash,
		&user.Name, &user.Role, &user.IsActive, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query user by email: %w", err)
	}
	return user, nil
}

// GetByID looks up a user by their UUID.
// Returns (nil, nil) if not found.
func (s *UserService) GetByID(id string) (*models.User, error) {
	user := &models.User{}
	err := s.db.QueryRow(
		`SELECT id, email, password_hash, name, role, is_active, created_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash,
		&user.Name, &user.Role, &user.IsActive, &user.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query user by id: %w", err)
	}
	return user, nil
}

// =====================================================================
//  Authentication
// =====================================================================

// Authenticate verifies email + password against the database.
// Returns the user if credentials are valid, nil if invalid.
// A nil return with a nil error means "wrong email or password" (not a server error).
func (s *UserService) Authenticate(email, password string) (*models.User, error) {
	user, err := s.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil // email not found
	}
	if !user.IsActive {
		return nil, nil // account deactivated
	}

	// Compare the plaintext password against the stored bcrypt hash.
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash), []byte(password),
	); err != nil {
		return nil, nil // wrong password
	}

	return user, nil
}

// =====================================================================
//  Password Hashing
// =====================================================================

// HashPassword produces a bcrypt hash of the given plaintext password.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), config.BcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// =====================================================================
//  Seeding (for development / testing)
// =====================================================================

// SeedUsers creates sample users if they don't already exist.
// Passwords are hashed with bcrypt before insertion.
func (s *UserService) SeedUsers() error {
	seeds := []struct {
		Email, Password, Name, Role string
	}{
		{"staff@hospital.com", "password123", "Staff User", config.RoleStaff},
		{"admin@hospital.com", "password123", "Admin User", config.RoleAdmin},
		{"master@hospital.com", "password123", "Master Admin", config.RoleMasterAdmin},
	}

	for _, seed := range seeds {
		// Skip if user already exists.
		existing, _ := s.GetByEmail(seed.Email)
		if existing != nil {
			continue
		}

		hash, err := HashPassword(seed.Password)
		if err != nil {
			return err
		}

		_, err = s.db.Exec(
			`INSERT INTO users (email, password_hash, name, role)
			 VALUES ($1, $2, $3, $4)`,
			seed.Email, hash, seed.Name, seed.Role,
		)
		if err != nil {
			return fmt.Errorf("seed user %s: %w", seed.Email, err)
		}
		log.Printf("🌱 Seeded user: %s (role: %s)", seed.Email, seed.Role)
	}

	return nil
}
