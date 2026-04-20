package models

import "time"

// OTPEntry represents a single OTP record stored in the in-memory store.
// Each entry is keyed by the user's email address.
type OTPEntry struct {
	// Code is the 6-digit OTP string (e.g. "048312").
	Code string

	// ExpiresAt marks when this OTP becomes invalid.
	ExpiresAt time.Time
}

// IsExpired returns true if the OTP has passed its expiration time.
func (o *OTPEntry) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}
