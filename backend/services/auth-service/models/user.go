package models

import "time"

// Account status values — must match the account_status ENUM in the database.
const (
	AccountStatusInvited   = "invited"
	AccountStatusActive    = "active"
	AccountStatusSuspended = "suspended"
)

// User represents a row in the `users` table.
type User struct {
	UserID           string     `json:"user_id"`
	GroupID          *string    `json:"group_id"`       // nullable UUID FK → groups
	Email            string     `json:"email"`
	Verified         bool       `json:"verified"`
	Role             string     `json:"role"` // user_role ENUM: 'staff', 'admin', 'master-admin'
	VerificationCode *string    `json:"-"`    // hidden from JSON
	PasswordHash     *string    `json:"-"`    // hidden from JSON
	AccountStatus    string     `json:"account_status"` // invited | active | suspended
	InvitationToken  *string    `json:"-"`              // hidden from JSON
	TokenExpiresAt   *time.Time `json:"-"`              // hidden from JSON
	FailedAttempts   int        `json:"-"`              // failed login counter
	LockedUntil      *time.Time `json:"-"`              // account lock expiry
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// IsLocked returns true if the account is currently locked.
func (u *User) IsLocked() bool {
	return u.LockedUntil != nil && time.Now().Before(*u.LockedUntil)
}

// LoginBlockReason returns a user-facing message when login/token refresh is blocked.
// Empty string means the account may authenticate.
func (u *User) LoginBlockReason() string {
	switch u.AccountStatus {
	case AccountStatusSuspended:
		return "Account is suspended. Contact your administrator."
	case AccountStatusInvited:
		return "Account is not activated. Please complete your invitation setup."
	default:
		return ""
	}
}
