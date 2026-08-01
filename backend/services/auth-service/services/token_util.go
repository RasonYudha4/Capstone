package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"auth-service/config"
)

// generateSecureToken returns a cryptographically random hex token.
// On failure it returns the underlying crypto/rand error unmodified so
// callers can wrap it with their existing error messages.
func generateSecureToken() (string, error) {
	b := make([]byte, config.SecureTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// hashToken returns a hex-encoded SHA-256 digest of raw.
// Used for refresh tokens, OTP codes, and invitation/reset tokens at rest.
func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
