package repositories

import (
	"auth-service/models"
	"database/sql"
	"time"
)

type OTPRepository struct {
	db *sql.DB
}

func NewOTPRepository(db *sql.DB) *OTPRepository {
	return &OTPRepository{db: db}
}

func (r *OTPRepository) DeleteByEmail(email string) error {
	_, err := r.db.Exec(`DELETE FROM otp_entries WHERE email = $1`, email)
	return err
}

func (r *OTPRepository) Save(email, code, preAuthToken string, expiresAt time.Time) error {
	_, err := r.db.Exec(
		`INSERT INTO otp_entries (email, otp_code, pre_auth_token, failed_attempts, expires_at)
		 VALUES ($1, $2, $3, 0, $4)`,
		email, code, preAuthToken, expiresAt,
	)
	return err
}

func (r *OTPRepository) Get(email, preAuthToken string) (*models.OTPEntry, error) {
	var entry models.OTPEntry
	err := r.db.QueryRow(
		`SELECT id, email, otp_code, pre_auth_token, failed_attempts, expires_at, created_at
		 FROM otp_entries WHERE email = $1 AND pre_auth_token = $2`,
		email, preAuthToken,
	).Scan(&entry.ID, &entry.Email, &entry.Code, &entry.PreAuthToken, &entry.FailedAttempts, &entry.ExpiresAt, &entry.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *OTPRepository) DeleteByID(id int) error {
	_, err := r.db.Exec(`DELETE FROM otp_entries WHERE id = $1`, id)
	return err
}

func (r *OTPRepository) IncrementFailedAttempts(id int) error {
	_, err := r.db.Exec(`UPDATE otp_entries SET failed_attempts = failed_attempts + 1 WHERE id = $1`, id)
	return err
}

func (r *OTPRepository) HasPending(email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM otp_entries WHERE email = $1 AND expires_at > NOW())`,
		email,
	).Scan(&exists)
	return exists, err
}

func (r *OTPRepository) GetLastCreatedAt(email string) (time.Time, error) {
	var createdAt time.Time
	err := r.db.QueryRow(
		`SELECT created_at FROM otp_entries WHERE email = $1 ORDER BY created_at DESC LIMIT 1`,
		email,
	).Scan(&createdAt)
	return createdAt, err
}

func (r *OTPRepository) DeleteExpired() (int64, error) {
	result, err := r.db.Exec(`DELETE FROM otp_entries WHERE expires_at < NOW()`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
