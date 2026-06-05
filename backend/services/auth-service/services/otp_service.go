package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"time"

	"auth-service/config"
	"auth-service/models"
)

// manages OTP generation, PostgreSQL storage, and verification.
type OTPService struct {
	db *sql.DB
}

// creates a ready-to-use OTPService instance backed by PostgreSQL.
func NewOTPService(db *sql.DB) *OTPService {
	return &OTPService{
		db: db,
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

	// Generate pre-auth token (32 bytes hex encoded = 64 chars)
	preAuthBytes := make([]byte, 32)
	if _, err := rand.Read(preAuthBytes); err != nil {
		return "", fmt.Errorf("failed to generate pre-auth token: %w", err)
	}
	preAuthToken := hex.EncodeToString(preAuthBytes)

	expiresAt := time.Now().Add(config.OTPExpiration)

	// Delete any existing OTP for this email first
	if _, err := s.db.Exec(`DELETE FROM otp_entries WHERE email = $1`, email); err != nil {
		return "", fmt.Errorf("failed to clear existing OTP: %w", err)
	}

	// INSERT the new OTP entry into the database
	_, err = s.db.Exec(
		`INSERT INTO otp_entries (email, otp_code, pre_auth_token, failed_attempts, expires_at)
		 VALUES ($1, $2, $3, 0, $4)`,
		email, code, preAuthToken, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("failed to store OTP: %w", err)
	}

	// In production, send OTP via email service (SendGrid/SES/etc.).
	// DO NOT log the actual code in production.
	log.Printf("📧 [OTP] Code generated for %s (purpose: %s, expires: %s) -> CODE: %s",
		email, purpose, expiresAt.Format(time.RFC3339), code)

	return preAuthToken, nil // Return pre-auth token, BUKAN OTP code
}

// checks the submitted OTP against the stored entry.
func (s *OTPService) Verify(email, code, preAuthToken string) (valid bool, errMsg string, err error) {
	var entry models.OTPEntry
	err = s.db.QueryRow(
		`SELECT id, email, otp_code, pre_auth_token, failed_attempts, expires_at, created_at
		 FROM otp_entries WHERE email = $1 AND pre_auth_token = $2`,
		email, preAuthToken,
	).Scan(&entry.ID, &entry.Email, &entry.Code, &entry.PreAuthToken, &entry.FailedAttempts, &entry.ExpiresAt, &entry.CreatedAt)

	if err == sql.ErrNoRows {
		return false, "No OTP requested for this email.", nil
	}
	if err != nil {
		return false, "", fmt.Errorf("failed to query OTP: %w", err)
	}

	// OTP has expired
	if entry.IsExpired() {
		if _, delErr := s.db.Exec(`DELETE FROM otp_entries WHERE id = $1`, entry.ID); delErr != nil {
			log.Printf("⚠️  Failed to delete expired OTP for %s: %v", email, delErr)
		}
		return false, "OTP has expired. Please request a new one.", nil
	}

	// Too many failed attempts
	if entry.FailedAttempts >= config.OTPMaxAttempts {
		if _, delErr := s.db.Exec(`DELETE FROM otp_entries WHERE id = $1`, entry.ID); delErr != nil {
			log.Printf("⚠️  Failed to delete OTP after max attempts for %s: %v", email, delErr)
		}
		return false, "Too many failed attempts. Please request a new OTP.", nil
	}

	// Code mismatch, increment attempt counter.
	if entry.Code != code {
		if _, updErr := s.db.Exec(`UPDATE otp_entries SET failed_attempts = failed_attempts + 1 WHERE id = $1`, entry.ID); updErr != nil {
			log.Printf("⚠️  Failed to update failed_attempts for %s: %v", email, updErr)
		}
		remaining := config.OTPMaxAttempts - entry.FailedAttempts - 1
		return false, fmt.Sprintf("Invalid OTP. %d attempt(s) remaining.", remaining), nil
	}

	// delete valid OTP so it can't be reused (single-use).
	if _, delErr := s.db.Exec(`DELETE FROM otp_entries WHERE id = $1`, entry.ID); delErr != nil {
		log.Printf("⚠️  Failed to delete verified OTP for %s: %v", email, delErr)
	}
	return true, "", nil
}

// checks whether an OTP entry exists for the given email.
func (s *OTPService) HasPending(email string) bool {
	var exists bool
	err := s.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM otp_entries WHERE email = $1 AND expires_at > NOW())`,
		email,
	).Scan(&exists)
	if err != nil {
		log.Printf("⚠️  Failed to check pending OTP for %s: %v", email, err)
		return false
	}
	return exists
}

// CanResend checks if enough time has passed since the last OTP was generated.
// Returns true if at least 60 seconds have elapsed (prevents email flooding).
func (s *OTPService) CanResend(email string) bool {
	const resendCooldown = 60 * time.Second
	var createdAt time.Time
	err := s.db.QueryRow(
		`SELECT created_at FROM otp_entries WHERE email = $1 ORDER BY created_at DESC LIMIT 1`,
		email,
	).Scan(&createdAt)
	if err != nil {
		return false
	}
	return time.Since(createdAt) >= resendCooldown
}

// StartCleanup runs a background goroutine that deletes expired OTP entries every 2 minutes.
func (s *OTPService) StartCleanup(ctx context.Context) {*
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("🛑 OTP cleanup goroutine stopped")
				return
			case <-ticker.C:
				result, err := s.db.Exec(`DELETE FROM otp_entries WHERE expires_at < NOW()`)
				if err != nil {
					log.Printf("⚠️  OTP cleanup failed: %v", err)
					continue
				}
				if rows, _ := result.RowsAffected(); rows > 0 {
					log.Printf("🧹 OTP cleanup: removed %d expired entries", rows)
				}
			}
		}
	}()
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
