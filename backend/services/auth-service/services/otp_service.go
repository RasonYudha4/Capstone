package services

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"auth-service/config"
	"auth-service/models"
)

// OTPService manages OTP generation, in-memory storage, and verification.
// The store is protected by a mutex for concurrent safety.
// This will be swapped for Redis or a database table in later phases.
type OTPService struct {
	mu    sync.RWMutex
	store map[string]*models.OTPEntry // email → OTP entry
}

// NewOTPService creates a ready-to-use OTPService instance.
func NewOTPService() *OTPService {
	return &OTPService{
		store: make(map[string]*models.OTPEntry),
	}
}

// GenerateAndStore creates a cryptographically random OTP for the given email,
// stores it with an expiration timestamp, and logs it to the console.
//
// Parameters:
//   - email:   the user's email (used as the storage key)
//   - purpose: why this OTP is being generated ("login" or "action_confirm")
func (s *OTPService) GenerateAndStore(email, purpose string) (string, error) {
	// Generate a secure random 6-digit code using crypto/rand.
	// (math/rand is NOT suitable for security-sensitive codes.)
	code, err := generateSecureOTP(config.OTPLength)
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}

	entry := &models.OTPEntry{
		Code:      code,
		Purpose:   purpose,
		Attempts:  0,
		ExpiresAt: time.Now().Add(config.OTPExpiration),
	}

	// Store the OTP. Any previous OTP for this email is overwritten,
	// which implicitly invalidates it.
	s.mu.Lock()
	s.store[email] = entry
	s.mu.Unlock()

	// Simulate sending the OTP via email.
	// In production, replace this with an actual email service (e.g. SendGrid, SES).
	log.Printf("📧 [OTP] Code for %s: %s (purpose: %s, expires: %s)",
		email, code, purpose, entry.ExpiresAt.Format(time.RFC3339))

	return code, nil
}

// Verify checks the submitted OTP against the stored entry.
// Returns:
//   - valid:  true if the OTP matches and hasn't expired or exceeded max attempts
//   - errMsg: human-readable reason for rejection (empty on success)
//   - err:    non-nil only on unexpected internal failures
func (s *OTPService) Verify(email, code string) (valid bool, errMsg string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.store[email]

	// No OTP was ever requested for this email.
	if !exists {
		return false, "No OTP requested for this email.", nil
	}

	// OTP has expired — clean up and reject.
	if entry.IsExpired() {
		delete(s.store, email)
		return false, "OTP has expired. Please request a new one.", nil
	}

	// Too many failed attempts — invalidate and reject.
	if entry.Attempts >= config.OTPMaxAttempts {
		delete(s.store, email)
		return false, "Too many failed attempts. Please request a new OTP.", nil
	}

	// Code mismatch — increment attempt counter.
	if entry.Code != code {
		entry.Attempts++
		remaining := config.OTPMaxAttempts - entry.Attempts
		return false, fmt.Sprintf("Invalid OTP. %d attempt(s) remaining.", remaining), nil
	}

	// Valid OTP — delete it so it can't be reused (single-use).
	delete(s.store, email)
	return true, "", nil
}

// generateSecureOTP produces a random numeric string of the given length
// using crypto/rand for unpredictability.
func generateSecureOTP(length int) (string, error) {
	// The upper bound is 10^length (e.g. 1_000_000 for 6 digits).
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	// Zero-pad to the required length (e.g. 42 → "000042").
	format := fmt.Sprintf("%%0%dd", length)
	return fmt.Sprintf(format, n), nil
}
