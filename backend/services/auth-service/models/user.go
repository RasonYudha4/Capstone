package models

import "time"

// represents a row in the `users` table.
type User struct {
	UserID           string     `json:"user_id"`
	GroupID          *string    `json:"group_id"`           // nullable UUID FK → groups
	Email            string     `json:"email"`
	Verified         bool       `json:"verified"`
	Role             string     `json:"role"`               // user_role ENUM: 'staff', 'admin', 'master-admin'
	VerificationCode *string    `json:"-"`                  // hidden from JSON
	PasswordHash     *string    `json:"-"`                  // hidden from JSON (Phase 1)
	AccountStatus    string     `json:"account_status"`     // 'invited', 'active', 'suspended'
	InvitationToken  *string    `json:"-"`                  // hidden from JSON
	TokenExpiresAt   *time.Time `json:"-"`                  // hidden from JSON
	FailedAttempts   int        `json:"-"`                  // Phase 2: failed login counter
	LockedUntil      *time.Time `json:"-"`                  // Phase 2: account lock expiry
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// returns true if the account is currently locked.
func (u *User) IsLocked() bool {
	return u.LockedUntil != nil && time.Now().Before(*u.LockedUntil)
}
