CREATE TABLE IF NOT EXISTS otp_entries (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    otp_code VARCHAR(10) NOT NULL,
    pre_auth_token VARCHAR(64) NOT NULL,
    failed_attempts INT DEFAULT 0,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_otp_entries_email ON otp_entries(email);
CREATE INDEX idx_otp_entries_pre_auth_token ON otp_entries(pre_auth_token);
