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

func (r *UserRepository) UpdateRole(userID string, role string, groupID *string) error {
	_, err := r.db.Exec(
		`UPDATE users SET role = $1, group_id = $2, updated_at = NOW() WHERE user_id = $3`,
		role, groupID, userID,
	)
	return err
}

func (r *UserRepository) CreateInvitedUser(email, role string, groupID *string, tokenHash string, expiresAt time.Time) (string, error) {
	var userID string
	err := r.db.QueryRow(
		`INSERT INTO users (email, role, group_id, account_status, invitation_token, token_expires_at, verified, created_at, updated_at)
		 VALUES ($1, $2, $3, 'invited', $4, $5, false, NOW(), NOW())
		 RETURNING user_id`,
		email, role, groupID, tokenHash, expiresAt,
	).Scan(&userID)
	return userID, err
}

func (r *UserRepository) GetByInvitationToken(tokenHash string) (*models.User, error) {
	row := r.db.QueryRow(
		`SELECT `+userColumns+` FROM users WHERE invitation_token = $1`, tokenHash,
	)
	return scanUser(row)
}

func (r *UserRepository) CompleteInvitation(userID, passwordHash string) error {
	// Keep verified=false for admin/master-admin so first login still requires OTP.
	// Staff do not use the OTP branch, so mark them verified immediately.
	_, err := r.db.Exec(
		`UPDATE users SET password_hash = $1, account_status = 'active',
		 invitation_token = NULL, token_expires_at = NULL,
		 verified = CASE WHEN role IN ('admin', 'master-admin') THEN false ELSE true END,
		 updated_at = NOW()
		 WHERE user_id = $2`,
		passwordHash, userID,
	)
	return err
}

func (r *UserRepository) SetResetToken(userID, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(
		`UPDATE users SET invitation_token = $1, token_expires_at = $2, updated_at = NOW()
		 WHERE user_id = $3`,
		tokenHash, expiresAt, userID,
	)
	return err
}

func (r *UserRepository) UpdateInvitationToken(userID, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(
		`UPDATE users SET invitation_token = $1, token_expires_at = $2, updated_at = NOW()
		 WHERE user_id = $3 AND account_status = 'invited'`,
		tokenHash, expiresAt, userID,
	)
	return err
}

func (r *UserRepository) CompletePasswordReset(userID, passwordHash string) error {
	_, err := r.db.Exec(
		`UPDATE users SET password_hash = $1, invitation_token = NULL, token_expires_at = NULL,
		 failed_attempts = 0, locked_until = NULL, updated_at = NOW()
		 WHERE user_id = $2`,
		passwordHash, userID,
	)
	return err
}

// GetAllUsers returns all users except master-admins, ordered by created_at desc.
// Only returns the columns needed for the admin panel (no sensitive fields).
func (r *UserRepository) GetAllUsers() ([]models.UserListItem, error) {
	rows, err := r.db.Query(
		`SELECT u.user_id, u.email, u.role, u.group_id, g.group_name, u.account_status, u.verified
		 FROM users u
		 LEFT JOIN groups g ON u.group_id = g.group_id
		 WHERE u.role != 'master-admin'
		 ORDER BY u.created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query all users: %w", err)
	}
	defer rows.Close()

	var users []models.UserListItem
	for rows.Next() {
		var u models.UserListItem
		if err := rows.Scan(&u.UserID, &u.Email, &u.Role, &u.GroupID, &u.GroupName, &u.AccountStatus, &u.Verified); err != nil {
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

// DeleteUser permanently removes an invited user record.
// The account_status guard is enforced in SQL to close TOCTOU races.
func (r *UserRepository) DeleteUser(userID string) error {
	result, err := r.db.Exec(
		`DELETE FROM users WHERE user_id = $1 AND account_status = 'invited'`,
		userID,
	)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("user not found or not in invited status")
	}
	return nil
}
