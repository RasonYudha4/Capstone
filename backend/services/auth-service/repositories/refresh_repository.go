package repositories

import (
	"database/sql"
	"fmt"
	"time"
)

type RefreshRepository struct {
	db *sql.DB
}

func NewRefreshRepository(db *sql.DB) *RefreshRepository {
	return &RefreshRepository{db: db}
}

func (r *RefreshRepository) Save(userID, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	return err
}

func (r *RefreshRepository) GetByHash(tokenHash string) (userID string, expiresAt time.Time, revoked bool, err error) {
	err = r.db.QueryRow(
		`SELECT user_id, expires_at, revoked
		 FROM refresh_tokens WHERE token_hash = $1`,
		tokenHash,
	).Scan(&userID, &expiresAt, &revoked)
	return
}

func (r *RefreshRepository) RevokeByHash(tokenHash string) (userID string, err error) {
	err = r.db.QueryRow(
		`UPDATE refresh_tokens SET revoked = TRUE
		 WHERE token_hash = $1 AND revoked = FALSE
		 RETURNING user_id`,
		tokenHash,
	).Scan(&userID)
	return
}

func (r *RefreshRepository) RevokeAllByUserID(userID string) error {
	_, err := r.db.Exec(
		`UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1 AND revoked = FALSE`,
		userID,
	)
	return err
}

func (r *RefreshRepository) RotateToken(oldHash, newHash string, expiresAt time.Time) (userID string, err error) {
	tx, err := r.db.Begin()
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// 1. Revoke old token within transaction
	err = tx.QueryRow(
		`UPDATE refresh_tokens SET revoked = TRUE
		 WHERE token_hash = $1 AND revoked = FALSE
		 RETURNING user_id`,
		oldHash,
	).Scan(&userID)
	if err != nil {
		return "", err
	}

	// 2. Create new token within same transaction
	_, err = tx.Exec(
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)`,
		userID, newHash, expiresAt,
	)
	if err != nil {
		return "", err
	}

	// 3. Commit both operations atomically
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit: %w", err)
	}

	return userID, nil
}
