package config

import "time"

// Application-wide configuration constants.
// In production, these would come from environment variables or a config file.
const (
	// port the HTTP server listens on.
	ServerPort = ":8080"

	// JWTSecret is the signing key for JWT tokens.
	// In production, load this from an environment variable (e.g. JWT_SECRET).
	JWTSecret = "super-secret-key-change-in-production"

	// how long a JWT token remains valid after issuance.
	JWTExpiration = 1 * time.Hour

	// how long an OTP code remains valid after generation.
	OTPExpiration = 5 * time.Minute

	// number of digits in a generated OTP code.
	OTPLength = 6

	// DefaultRole is the role assigned to every authenticated user for now.
	// This will be replaced with a proper role system once the database is integrated.
	DefaultRole = "master-admin"
	// maximum failed OTP verification attempts before the code is invalidated.
	OTPMaxAttempts = 3

	// bcrypt cost factor for password hashing.
	// 10 is a good balance between security and speed.
	BcryptCost = 10

	// --- RBAC Roles ---
	// Role hierarchy (least → most privileged): staff < admin < master_admin.

	RoleStaff       = "staff"
	RoleAdmin       = "admin"
	RoleMasterAdmin = "master_admin"

	// Gin context key where the authenticated user's
	// claims are stored after passing through the JWT middleware.
	ContextKeyUser = "authenticated_user"

	// PostgreSQL connection string.
	// In production, load from DATABASE_URL environment variable.
	DatabaseURL = "postgres://postgres:postgres@localhost:5432/hospital_auth?sslmode=disable"
)
