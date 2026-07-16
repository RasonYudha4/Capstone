package repositories

import (
	"auth-service/models"
	"database/sql"
	"fmt"
	"time"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// shared column list for all user queries.
const userColumns = `user_id, group_id, email, verified, role,
	verification_code, password_hash, account_status, invitation_token,
	token_expires_at, failed_attempts, locked_until, created_at, updated_at`

// scans a sql.Row into a User struct.
func scanUser(row *sql.Row) (*models.User, error) {
	user := &models.User{}
	err := row.Scan(
		&user.UserID, &user.GroupID, &user.Email, &user.Verified,
		&user.Role, &user.VerificationCode, &user.PasswordHash,
		&user.AccountStatus, &user.InvitationToken, &user.TokenExpiresAt,
		&user.FailedAttempts, &user.LockedUntil,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	row := r.db.QueryRow(
		`SELECT `+userColumns+` FROM users WHERE email = $1`, email,
	)
	user, err := scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("query user by email: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetByID(id string) (*models.User, error) {
	row := r.db.QueryRow(
		`SELECT `+userColumns+` FROM users WHERE user_id = $1`, id,
	)
	user, err := scanUser(row)
	if err != nil {
		return nil, fmt.Errorf("query user by id: %w", err)
	}
	return user, nil
}

func (r *UserRepository) IncrementFailedAttempts(userID string) (int, error) {
	var count int
	err := r.db.QueryRow(
		`UPDATE users SET failed_attempts = failed_attempts + 1, updated_at = NOW()
		 WHERE user_id = $1 RETURNING failed_attempts`,
		userID,
	).Scan(&count)
	return count, err
}

func (r *UserRepository) UpdateLock(userID string, lockedUntil time.Time) error {
	_, err := r.db.Exec(
		`UPDATE users SET locked_until = $1, updated_at = NOW() WHERE user_id = $2`,
		lockedUntil, userID,
	)
	return err
}

func (r *UserRepository) ResetFailedAttempts(userID string) error {
	_, err := r.db.Exec(
		`UPDATE users SET failed_attempts = 0, locked_until = NULL, updated_at = NOW()
		 WHERE user_id = $1`,
		userID,
	)
	return err
}

func (r *UserRepository) MarkAsVerified(userID string) error {
	_, err := r.db.Exec(
		`UPDATE users SET verified = true, updated_at = NOW() WHERE user_id = $1`,
		userID,
	)
	return err
}

func (r *UserRepository) UpdatePasswordHash(userID string, passwordHash string) error {
	_, err := r.db.Exec(
		`UPDATE users SET password_hash = $1, updated_at = NOW() WHERE user_id = $2`,
		passwordHash, userID,
	)
	return err
}

func (r *UserRepository) UpdateRole(userID string, role string) error {
	_, err := r.db.Exec(
		`UPDATE users SET role = $1, updated_at = NOW() WHERE user_id = $2`,
		role, userID,
	)
	return err
}

func (r *UserRepository) CreateInvitedUser(email, role, token string, expiresAt time.Time) error {
	_, err := r.db.Exec(
		`INSERT INTO users (email, role, account_status, invitation_token, token_expires_at, verified, created_at, updated_at)
		 VALUES ($1, $2, 'invited', $3, $4, false, NOW(), NOW())`,
		email, role, token, expiresAt,
	)
	return err
}

func (r *UserRepository) GetByInvitationToken(token string) (*models.User, error) {
	row := r.db.QueryRow(
		`SELECT `+userColumns+` FROM users WHERE invitation_token = $1`, token,
	)
	return scanUser(row)
}

func (r *UserRepository) CompleteInvitation(userID, passwordHash string) error {
	_, err := r.db.Exec(
		`UPDATE users SET password_hash = $1, account_status = 'active', 
		 invitation_token = NULL, token_expires_at = NULL, verified = true, updated_at = NOW() 
		 WHERE user_id = $2`,
		passwordHash, userID,
	)
	return err
}

// GetAllUsers returns all users except master-admins, ordered by created_at desc.
// Only returns the columns needed for the admin panel (no sensitive fields).
func (r *UserRepository) GetAllUsers() ([]models.UserListItem, error) {
	rows, err := r.db.Query(
		`SELECT user_id, email, role, account_status, verified
		 FROM users
		 WHERE role != 'master-admin'
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query all users: %w", err)
	}
	defer rows.Close()

	var users []models.UserListItem
	for rows.Next() {
		var u models.UserListItem
		if err := rows.Scan(&u.UserID, &u.Email, &u.Role, &u.AccountStatus, &u.Verified); err != nil {
			return nil, fmt.Errorf("scan user row: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user rows: %w", err)
	}
	return users, nil
}

// UpdateStatus sets the account_status of a user.
func (r *UserRepository) UpdateStatus(userID, status string) error {
	_, err := r.db.Exec(
		`UPDATE users SET account_status = $1, updated_at = NOW() WHERE user_id = $2`,
		status, userID,
	)
	return err
}

// DeleteUser permanently removes a user record.
// Should only be called for users with account_status = 'invited'.
func (r *UserRepository) DeleteUser(userID string) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE user_id = $1`, userID)
	return err
}
