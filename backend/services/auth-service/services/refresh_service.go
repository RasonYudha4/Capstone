package services

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"auth-service/config"
)

// manages refresh token creation, validation, and revocation.
// raw token is only ever held by the client.
type RefreshService struct {
	db *sql.DB
}

// creates a RefreshService backed by the given database.
func NewRefreshService(db *sql.DB) *RefreshService {
	return &RefreshService{db: db}
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

	_, err := s.db.Exec(
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("store refresh token: %w", err)
	}

	return rawToken, nil
}

// checks if a raw refresh token is valid.
func (s *RefreshService) ValidateToken(rawToken string) (userID string, err error) {
	tokenHash := hashToken(rawToken)

	var expiresAt time.Time
	var revoked bool

	err = s.db.QueryRow(
		`SELECT user_id, expires_at, revoked
		 FROM refresh_tokens WHERE token_hash = $1`,
		tokenHash,
	).Scan(&userID, &expiresAt, &revoked)

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

	err = s.db.QueryRow(
		`UPDATE refresh_tokens SET revoked = TRUE
		 WHERE token_hash = $1 AND revoked = FALSE
		 RETURNING user_id`,
		tokenHash,
	).Scan(&userID)

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
	_, err := s.db.Exec(
		`UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1 AND revoked = FALSE`,
		userID,
	)
	return err
}

// atomically revokes the old refresh token and issues a new one.
// prevents stolen tokens from being reused indefinitely.
func (s *RefreshService) RotateToken(rawToken string) (newRawToken string, userID string, err error) {
    tx, err := s.db.Begin()
    if err != nil {
        return "", "", fmt.Errorf("rotate: begin tx: %w", err)
    }
    defer tx.Rollback() // no-op if committed

    // 1. Revoke old token within transaction
    oldHash := hashToken(rawToken)
    err = tx.QueryRow(
        `UPDATE refresh_tokens SET revoked = TRUE
         WHERE token_hash = $1 AND revoked = FALSE
         RETURNING user_id`,
        oldHash,
    ).Scan(&userID)
    if err == sql.ErrNoRows {
        return "", "", fmt.Errorf("token not found or already revoked")
    }
    if err != nil {
        return "", "", fmt.Errorf("rotate: revoke: %w", err)
    }

    // 2. Create new token within same transaction
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", "", fmt.Errorf("rotate: generate token: %w", err)
    }
    newRawToken = base64.URLEncoding.EncodeToString(b)
    newHash := hashToken(newRawToken)
    expiresAt := time.Now().Add(config.RefreshTokenExpiry)

    _, err = tx.Exec(
        `INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
         VALUES ($1, $2, $3)`,
        userID, newHash, expiresAt,
    )
    if err != nil {
        return "", "", fmt.Errorf("rotate: create: %w", err)
    }

    // 3. Commit both operations atomically
    if err := tx.Commit(); err != nil {
        return "", "", fmt.Errorf("rotate: commit: %w", err)
    }

    return newRawToken, userID, nil
}

// produces a hex-encoded SHA-256 hash of the raw token.
func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
