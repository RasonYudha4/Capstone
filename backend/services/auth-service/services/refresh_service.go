package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"auth-service/config"
	"auth-service/repositories"
)

// manages refresh token creation, validation, and revocation.
// raw token is only ever held by the client.
type RefreshService struct {
	refreshRepo *repositories.RefreshRepository
}

// creates a RefreshService backed by the given RefreshRepository.
func NewRefreshService(refreshRepo *repositories.RefreshRepository) *RefreshService {
	return &RefreshService{refreshRepo: refreshRepo}
}

// generates a new refresh token for the given user.
// only the SHA-256 hash is stored in the database.
func (s *RefreshService) CreateToken(userID string) (string, error) {
	// Generate 32 random bytes → base64url encode → raw token.
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	rawToken := base64.URLEncoding.EncodeToString(b)

	// hash before storing (so a DB breach doesn't expose tokens).
	tokenHash := hashToken(rawToken)
	expiresAt := time.Now().Add(config.RefreshTokenExpiry)

	err := s.refreshRepo.Save(userID, tokenHash, expiresAt)
	if err != nil {
		return "", fmt.Errorf("store refresh token: %w", err)
	}

	return rawToken, nil
}

// checks if a raw refresh token is valid.
func (s *RefreshService) ValidateToken(rawToken string) (userID string, err error) {
	tokenHash := hashToken(rawToken)

	userID, expiresAt, revoked, err := s.refreshRepo.GetByHash(tokenHash)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("invalid refresh token")
	}
	if err != nil {
		return "", fmt.Errorf("query refresh token: %w", err)
	}

	if revoked {
		return "", fmt.Errorf("refresh token has been revoked")
	}

	if time.Now().After(expiresAt) {
		return "", fmt.Errorf("refresh token has expired")
	}

	return userID, nil
}

// marks a specific refresh token as revoked (used for logout).
func (s *RefreshService) RevokeToken(rawToken string) (userID string, err error) {
	tokenHash := hashToken(rawToken)

	userID, err = s.refreshRepo.RevokeByHash(tokenHash)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("token not found or already revoked")
	}
	if err != nil {
		return "", fmt.Errorf("revoke refresh token: %w", err)
	}

	return userID, nil
}

// revokes all refresh tokens for a user
func (s *RefreshService) RevokeAllUserTokens(userID string) error {
	return s.refreshRepo.RevokeAllByUserID(userID)
}

// atomically revokes the old refresh token and issues a new one.
// prevents stolen tokens from being reused indefinitely.
func (s *RefreshService) RotateToken(rawToken string) (newRawToken string, userID string, err error) {
	// 1. Generate new token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("rotate: generate token: %w", err)
	}
	newRawToken = base64.URLEncoding.EncodeToString(b)

	oldHash := hashToken(rawToken)
	newHash := hashToken(newRawToken)
	expiresAt := time.Now().Add(config.RefreshTokenExpiry)

	// 2. Perform rotation atomically in the repository
	userID, err = s.refreshRepo.RotateToken(oldHash, newHash, expiresAt)
	if err == sql.ErrNoRows {
		return "", "", fmt.Errorf("token not found or already revoked")
	}
	if err != nil {
		return "", "", fmt.Errorf("rotate: %w", err)
	}

	return newRawToken, userID, nil
}
