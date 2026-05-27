package config

import (
	"log"
	"os"
	"time"
)

// getEnv reads an environment variable or returns a fallback default.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Loaded from environment — NEVER hardcode secrets in production.
var (
	JWTSecret   = requireEnv("JWT_SECRET")
	DatabaseURL = getEnv("DB_URL", "postgresql://capstone:capstoneboi@127.0.0.1:5432/capstone_db?sslmode=disable")

	// Seeder credentials
	MasterAdminEmail    = getEnv("MASTER_ADMIN_EMAIL", "masteradmin@gmail.com")
	MasterAdminPassword = getEnv("MASTER_ADMIN_PASSWORD", "password123")
	AdminEmail          = getEnv("ADMIN_EMAIL", "admin@gmail.com")
	AdminPassword       = getEnv("ADMIN_PASSWORD", "password123")
	StaffEmail          = getEnv("STAFF_EMAIL", "staff@gmail.com")
	StaffPassword       = getEnv("STAFF_PASSWORD", "password123")
)

// requireEnv reads an environment variable or terminates.
func requireEnv(key string) string {
    v := os.Getenv(key)
    if v == "" {
        log.Fatalf("❌ Required environment variable %s is not set", key)
    }
    return v
}

// Application-wide configuration constants.
const (
	// port the HTTP server listens on.
	ServerPort = ":8089"

	// Token Lifetimes

	// how long an access token (JWT) remains valid.
	// short-lived to limit damage if stolen.
	AccessTokenExpiry = 15 * time.Minute

	// how long a refresh token remains valid.
	// long-lived to obtain new access tokens without re-login.
	RefreshTokenExpiry = 7 * 24 * time.Hour // 7 days

	// how long an OTP code remains valid after generation.
	OTPExpiration = 5 * time.Minute

	// number of digits in a generated OTP code.
	OTPLength = 6

	// maximum failed OTP verification attempts before the code is invalidated.
	OTPMaxAttempts = 3

	// bcrypt cost factor for password hashing.
	BcryptCost = 10

	// Account Lockout

	// before the account is temporarily locked.
	MaxLoginAttempts = 5

	// how long an account stays locked after max failed attempts.
	LockDuration = 15 * time.Minute

	// Password Policy

	// minimum password length for new passwords.
	PasswordMinLength = 8

	// RBAC Roles
	// must match the user_role ENUM defined in the database migration.
	// Role hierarchy (least → most privileged): staff < admin < master-admin.

	RoleStaff       = "staff"
	RoleAdmin       = "admin"
	RoleMasterAdmin = "master-admin"

	// context key where the authenticated user's
	// claims are stored after passing through the JWT middleware.
	ContextKeyUser = "authenticated_user"
)
