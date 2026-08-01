-- Finish 0004 on an already partially-migrated database.
-- Safe to re-run pieces that already succeeded (IF NOT EXISTS / exception handling).

-- 1) Re-point FKs from duplicate users to the keeper, then delete duplicates.
WITH ranked AS (
    SELECT
        user_id,
        email,
        ROW_NUMBER() OVER (
            PARTITION BY email
            ORDER BY
                CASE role
                    WHEN 'master-admin' THEN 0
                    WHEN 'admin' THEN 1
                    ELSE 2
                END,
                created_at ASC NULLS LAST,
                user_id::text ASC
        ) AS rn
    FROM users
),
keepers AS (
    SELECT user_id, email FROM ranked WHERE rn = 1
),
dupes AS (
    SELECT r.user_id, r.email, k.user_id AS keep_id
    FROM ranked r
    JOIN keepers k ON k.email = r.email
    WHERE r.rn > 1
)
UPDATE documents d
SET created_by = dupes.keep_id
FROM dupes
WHERE d.created_by = dupes.user_id;

WITH ranked AS (
    SELECT
        user_id,
        email,
        ROW_NUMBER() OVER (
            PARTITION BY email
            ORDER BY
                CASE role
                    WHEN 'master-admin' THEN 0
                    WHEN 'admin' THEN 1
                    ELSE 2
                END,
                created_at ASC NULLS LAST,
                user_id::text ASC
        ) AS rn
    FROM users
),
keepers AS (
    SELECT user_id, email FROM ranked WHERE rn = 1
),
dupes AS (
    SELECT r.user_id, r.email, k.user_id AS keep_id
    FROM ranked r
    JOIN keepers k ON k.email = r.email
    WHERE r.rn > 1
)
UPDATE audit a
SET user_id = dupes.keep_id
FROM dupes
WHERE a.user_id = dupes.user_id;

WITH ranked AS (
    SELECT
        user_id,
        email,
        ROW_NUMBER() OVER (
            PARTITION BY email
            ORDER BY
                CASE role
                    WHEN 'master-admin' THEN 0
                    WHEN 'admin' THEN 1
                    ELSE 2
                END,
                created_at ASC NULLS LAST,
                user_id::text ASC
        ) AS rn
    FROM users
),
keepers AS (
    SELECT user_id, email FROM ranked WHERE rn = 1
),
dupes AS (
    SELECT r.user_id, r.email, k.user_id AS keep_id
    FROM ranked r
    JOIN keepers k ON k.email = r.email
    WHERE r.rn > 1
)
UPDATE notifications n
SET user_id = dupes.keep_id
FROM dupes
WHERE n.user_id = dupes.user_id;

WITH ranked AS (
    SELECT
        user_id,
        email,
        ROW_NUMBER() OVER (
            PARTITION BY email
            ORDER BY
                CASE role
                    WHEN 'master-admin' THEN 0
                    WHEN 'admin' THEN 1
                    ELSE 2
                END,
                created_at ASC NULLS LAST,
                user_id::text ASC
        ) AS rn
    FROM users
)
DELETE FROM users u
USING ranked r
WHERE u.user_id = r.user_id
  AND r.rn > 1;

-- 2) Unique email (skip if already present).
DO $$
BEGIN
    ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email);
EXCEPTION
    WHEN duplicate_object THEN NULL;
    WHEN unique_violation THEN
        RAISE EXCEPTION 'Duplicate emails still exist after cleanup';
END $$;

-- 3) OTP column width (idempotent).
ALTER TABLE otp_entries
    ALTER COLUMN otp_code TYPE VARCHAR(64);

-- 4) Unique invitation token index.
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_invitation_token_unique
    ON users (invitation_token)
    WHERE invitation_token IS NOT NULL;
