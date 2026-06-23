-- Migration 0003: Add invitation-based onboarding columns to users table
-- account_status: tracks the lifecycle of a user account
--   'invited'  → admin has sent an invitation, password not yet set
--   'active'   → user has completed setup and can log in
--   'suspended'→ account is disabled by admin
-- invitation_token: cryptographically random token embedded in the invitation URL
-- token_expires_at: token is only valid for 24 hours

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS account_status   VARCHAR(20)  NOT NULL DEFAULT 'active',
    ADD COLUMN IF NOT EXISTS invitation_token VARCHAR(64)  NULL,
    ADD COLUMN IF NOT EXISTS token_expires_at TIMESTAMP    NULL;

-- Index for fast lookup when karyawan clicks the invitation link
CREATE INDEX IF NOT EXISTS idx_users_invitation_token ON users(invitation_token);
