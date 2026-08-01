DROP INDEX IF EXISTS idx_users_invitation_token_unique;

ALTER TABLE otp_entries
    ALTER COLUMN otp_code TYPE VARCHAR(10);

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_email_unique;

ALTER TABLE users
    ALTER COLUMN email DROP NOT NULL;
