package services

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"time"

	"auth-service/config"
	"auth-service/repositories"
)

// OTPService manages OTP generation, storage, and verification.
type OTPService struct {
	otpRepo      *repositories.OTPRepository
	emailService *EmailService
}

// NewOTPService creates a ready-to-use OTPService instance.
func NewOTPService(otpRepo *repositories.OTPRepository, emailService *EmailService) *OTPService {
	return &OTPService{
		otpRepo:      otpRepo,
		emailService: emailService,
	}
}

// GenerateAndStore creates a cryptographically random OTP for the given email,
// stores a SHA-256 hash of the code (never the plaintext), and returns a pre-auth token.
func (s *OTPService) GenerateAndStore(email, purpose string) (string, error) {
	code, err := generateSecureOTP(config.OTPLength)
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}

	preAuthToken, err := generateSecureToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate pre-auth token: %w", err)
	}

	expiresAt := time.Now().Add(config.OTPExpiration)

	if err := s.otpRepo.DeleteByEmail(email); err != nil {
		return "", fmt.Errorf("failed to clear existing OTP: %w", err)
	}

	codeHash := hashToken(code)
	if err := s.otpRepo.Save(email, codeHash, preAuthToken, expiresAt); err != nil {
		return "", fmt.Errorf("failed to store OTP: %w", err)
	}

	if config.IsDevMode {
		log.Printf("📧 [OTP][DEV] Code generated for %s (purpose: %s, expires: %s) -> CODE: %s",
			email, purpose, expiresAt.Format(time.RFC3339), code)
	} else {
		log.Printf("📧 [OTP] Code generated for %s (purpose: %s, expires: %s)",
			email, purpose, expiresAt.Format(time.RFC3339))
	}

	if s.emailService != nil {
		go func() {
			if err := s.emailService.SendOTP(email, code); err != nil {
				log.Printf("⚠️  Failed to send OTP email to %s: %v", email, err)
			}
		}()
	}

	return preAuthToken, nil
}

// Verify checks the submitted OTP against the stored hash.
func (s *OTPService) Verify(email, code, preAuthToken string) (valid bool, errMsg string, err error) {
	entry, err := s.otpRepo.Get(email, preAuthToken)
	if err != nil {
		return false, "", fmt.Errorf("failed to query OTP: %w", err)
	}
	if entry == nil {
		return false, "No OTP requested for this email.", nil
	}

	if entry.IsExpired() {
		if delErr := s.otpRepo.DeleteByID(entry.ID); delErr != nil {
			log.Printf("⚠️  Failed to delete expired OTP for %s: %v", email, delErr)
		}
		return false, "OTP has expired. Please request a new one.", nil
	}

	if entry.FailedAttempts >= config.OTPMaxAttempts {
		if delErr := s.otpRepo.DeleteByID(entry.ID); delErr != nil {
			log.Printf("⚠️  Failed to delete OTP after max attempts for %s: %v", email, delErr)
		}
		return false, "Too many failed attempts. Please request a new OTP.", nil
	}

	codeHash := hashToken(code)
	if subtle.ConstantTimeCompare([]byte(entry.Code), []byte(codeHash)) != 1 {
		if updErr := s.otpRepo.IncrementFailedAttempts(entry.ID); updErr != nil {
			log.Printf("⚠️  Failed to update failed_attempts for %s: %v", email, updErr)
		}
		remaining := config.OTPMaxAttempts - entry.FailedAttempts - 1
		return false, fmt.Sprintf("Invalid OTP. %d attempt(s) remaining.", remaining), nil
	}

	if delErr := s.otpRepo.DeleteByID(entry.ID); delErr != nil {
		log.Printf("⚠️  Failed to delete verified OTP for %s: %v", email, delErr)
	}
	return true, "", nil
}

// HasPending checks whether an OTP entry exists for the given email.
func (s *OTPService) HasPending(email string) bool {
	exists, err := s.otpRepo.HasPending(email)
	if err != nil {
		log.Printf("⚠️  Failed to check pending OTP for %s: %v", email, err)
		return false
	}
	return exists
}

// CanResend checks if enough time has passed since the last OTP was generated.
func (s *OTPService) CanResend(email string) bool {
	const resendCooldown = 60 * time.Second
	createdAt, err := s.otpRepo.GetLastCreatedAt(email)
	if err != nil {
		if err == sql.ErrNoRows {
			return true
		}
		return false
	}
	return time.Since(createdAt) >= resendCooldown
}

// StartCleanup runs a background goroutine that deletes expired OTP entries every 2 minutes.
func (s *OTPService) StartCleanup(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("🛑 OTP cleanup goroutine stopped")
				return
			case <-ticker.C:
				rows, err := s.otpRepo.DeleteExpired()
				if err != nil {
					log.Printf("⚠️  OTP cleanup failed: %v", err)
					continue
				}
				if rows > 0 {
					log.Printf("🧹 OTP cleanup: removed %d expired entries", rows)
				}
			}
		}
	}()
}

func generateSecureOTP(length int) (string, error) {
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	format := fmt.Sprintf("%%0%dd", length)
	return fmt.Sprintf(format, n), nil
}
