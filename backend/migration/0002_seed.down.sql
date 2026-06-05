-- Remove all seeded reference data (order matters due to foreign keys)
TRUNCATE assessment CASCADE;
TRUNCATE standard CASCADE;
TRUNCATE services CASCADE;
TRUNCATE groups CASCADE;
TRUNCATE document_types CASCADE;

-- Remove indexes
DROP INDEX IF EXISTS idx_otp_entries_email;
DROP INDEX IF EXISTS idx_otp_entries_pre_auth_token;