-- Rollback migration 0003
ALTER TABLE users
    DROP COLUMN IF EXISTS account_status,
    DROP COLUMN IF EXISTS invitation_token,
    DROP COLUMN IF EXISTS token_expires_at;

DROP INDEX IF EXISTS idx_users_invitation_token;
