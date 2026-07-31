package services

import (
	"crypto/rand"
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
