package models

import "time"

// OTPEntry represents a single OTP record stored in the in-memory store.
// Each entry is keyed by the user's email address.
type OTPEntry struct {
	// Code is the 6-digit OTP string (e.g. "048312").
	Code string

	// Purpose indicates why this OTP was generated.
	// "login" for login verification, "action_confirm" for sensitive actions.
	Purpose string

	// Attempts tracks how many times the user has tried to verify this OTP.
	Attempts int

	// ExpiresAt marks when this OTP becomes invalid.
	ExpiresAt time.Time

	// CreatedAt records when this OTP was generated (for resend cooldown).
	CreatedAt time.Time

	PreAuthToken string  // Token bukti password sudah diverifikasi
}

// IsExpired returns true if the OTP has passed its expiration time.
func (o *OTPEntry) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}
