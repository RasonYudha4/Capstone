package config

import "time"

// Application-wide configuration constants.
// In production, these would come from environment variables or a config file.
const (
	// port the HTTP server listens on.
	ServerPort = ":8089"

	// JWTSecret is the signing key for JWT tokens.
	// In production, load this from an environment variable (e.g. JWT_SECRET).
	JWTSecret = "super-secret-key-change-in-production"

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

	// load from DATABASE_URL environment variable.
	DatabaseURL = "postgresql://capstone:capstoneboi@127.0.0.1:5432/capstone_db?sslmode=disable"
)
