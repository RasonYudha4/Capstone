-- Remove seeded users
DELETE FROM users WHERE email IN (
  'masteradmin@gmail.com',
  'staff@gmail.com',
  'admin.manajemen@gmail.com',
  'admin.pelayanan@gmail.com',
  'admin.keselamatan@gmail.com',
  'admin.nasional@gmail.com'
);

-- Remove all seeded reference data (order matters due to foreign keys)
TRUNCATE assessment CASCADE;
TRUNCATE standard CASCADE;
TRUNCATE services CASCADE;
TRUNCATE groups CASCADE;
TRUNCATE document_types CASCADE;

-- Remove indexes
DROP INDEX IF EXISTS idx_otp_entries_email;
DROP INDEX IF EXISTS idx_otp_entries_pre_auth_token;