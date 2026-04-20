package config

import "time"

// Application-wide configuration constants.
// In production, these would come from environment variables or a config file.
const (
	// ServerPort is the port the HTTP server listens on.
	ServerPort = ":8080"

	// JWTSecret is the signing key for JWT tokens.
	// IMPORTANT: In production, load this from an environment variable (e.g. JWT_SECRET).
	JWTSecret = "super-secret-key-change-in-production"

	// JWTExpiration is how long a JWT token remains valid after issuance.
	JWTExpiration = 1 * time.Hour

	// OTPExpiration is how long an OTP code remains valid after generation.
	OTPExpiration = 5 * time.Minute

	// OTPLength is the number of digits in a generated OTP code.
	OTPLength = 6

	// DefaultRole is the role assigned to every authenticated user for now.
	// This will be replaced with a proper role system once the database is integrated.
	DefaultRole = "master-admin"
)
