package models

import "time"

// OTPEntry represents a single OTP record stored in the otp_entries table.
// Each entry is keyed by the user's email address.
type OTPEntry struct {
	// ID is the auto-incremented primary key.
	ID int `db:"id"`

	// Email is the user's email address.
	Email string `db:"email"`

	// Code holds the SHA-256 hex digest of the OTP (plaintext is never stored).
	Code string `db:"otp_code"`

	// PreAuthToken is a random token proving the user passed password verification.
	PreAuthToken string `db:"pre_auth_token"`

	// FailedAttempts tracks how many times the user has tried to verify this OTP.
	FailedAttempts int `db:"failed_attempts"`

	// ExpiresAt marks when this OTP becomes invalid.
	ExpiresAt time.Time `db:"expires_at"`

	// CreatedAt records when this OTP was generated (for resend cooldown).
	CreatedAt time.Time `db:"created_at"`
}

// IsExpired returns true if the OTP has passed its expiration time.
func (o *OTPEntry) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}
