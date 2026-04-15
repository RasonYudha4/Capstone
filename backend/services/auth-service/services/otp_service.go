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

// OTPService manages OTP generation, storage, and verification.
// The in-memory store is protected by a mutex for concurrent safety —
// this will be swapped for Redis or a database in later phases.
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
// stores it with an expiration timestamp, and "sends" it (console log for now).
func (s *OTPService) GenerateAndStore(email string) (string, error) {
	// Generate a secure random 6-digit code using crypto/rand
	// (math/rand is NOT suitable for security-sensitive codes).
	code, err := generateSecureOTP(config.OTPLength)
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}

	entry := &models.OTPEntry{
		Code:      code,
		ExpiresAt: time.Now().Add(config.OTPExpiration),
	}

	// Store the OTP — any previous OTP for this email is overwritten,
	// which implicitly invalidates it.
	s.mu.Lock()
	s.store[email] = entry
	s.mu.Unlock()

	// Simulate sending the OTP via email.
	// Replace this with an actual email service (e.g. SendGrid, SES) in production.
	log.Printf("📧 [OTP] Sending OTP to %s: %s (expires at %s)",
		email, code, entry.ExpiresAt.Format(time.RFC3339))

	return code, nil
}

// Verify checks the submitted OTP against the stored entry.
// On success the entry is deleted (one-time use).
// Returns (true, nil) if valid, (false, nil) if invalid, or (false, err) on unexpected failure.
func (s *OTPService) Verify(email, code string) (bool, error) {
	s.mu.RLock()
	entry, exists := s.store[email]
	s.mu.RUnlock()

	// No OTP was ever requested for this email.
	if !exists {
		return false, nil
	}

	// OTP has expired — clean it up and reject.
	if entry.IsExpired() {
		s.mu.Lock()
		delete(s.store, email)
		s.mu.Unlock()
		return false, nil
	}

	// Code mismatch.
	if entry.Code != code {
		return false, nil
	}

	// Valid OTP — delete it so it can't be reused.
	s.mu.Lock()
	delete(s.store, email)
	s.mu.Unlock()

	return true, nil
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
