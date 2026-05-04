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

// manages OTP generation, in-memory storage, and verification.
// The store is protected by a mutex for concurrent safety.
type OTPService struct {
	mu    sync.RWMutex
	store map[string]*models.OTPEntry
}

// creates a ready-to-use OTPService instance.
func NewOTPService() *OTPService {
	return &OTPService{
		store: make(map[string]*models.OTPEntry),
	}
}

// creates a cryptographically random OTP for the given email,
// stores it with an expiration timestamp, and logs it to the console.
//   - email:   the user's email (used as the storage key)
//   - purpose: why this OTP is being generated ("login" or "action_confirm")
func (s *OTPService) GenerateAndStore(email, purpose string) (string, error) {
	// secure random 6-digit code using crypto/rand.
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

	// store the OTP. overwrites the previous OTP for this email
	s.mu.Lock()
	s.store[email] = entry
	s.mu.Unlock()

	// send OTP via email
	log.Printf("📧 [OTP] Code for %s: %s (purpose: %s, expires: %s)",
		email, code, purpose, entry.ExpiresAt.Format(time.RFC3339))

	return code, nil
}

// checks the submitted OTP against the stored entry.
func (s *OTPService) Verify(email, code string) (valid bool, errMsg string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.store[email]

	// OTP was never requested for this email.
	if !exists {
		return false, "No OTP requested for this email.", nil
	}

	// OTP has expired
	if entry.IsExpired() {
		delete(s.store, email)
		return false, "OTP has expired. Please request a new one.", nil
	}

	// Too many failed attempts 
	if entry.Attempts >= config.OTPMaxAttempts {
		delete(s.store, email)
		return false, "Too many failed attempts. Please request a new OTP.", nil
	}

	// Code mismatch, increment attempt counter.
	if entry.Code != code {
		entry.Attempts++
		remaining := config.OTPMaxAttempts - entry.Attempts
		return false, fmt.Sprintf("Invalid OTP. %d attempt(s) remaining.", remaining), nil
	}

	// delete valid otp so it can't be reused (single-use).
	delete(s.store, email)
	return true, "", nil
}

// checks whether an OTP entry exists for the given email.
func (s *OTPService) HasPending(email string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.store[email]
	return exists
}

// produces a random numeric string of the given length
func generateSecureOTP(length int) (string, error) {
	// upper bound is 10^length (e.g. 1_000_000 for 6 digits).
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	// zero-pad to the required length (e.g. 42 → "000042").
	format := fmt.Sprintf("%%0%dd", length)
	return fmt.Sprintf(format, n), nil
}
