package services

import (
	"database/sql"
	"log"
	"time"

	"auth-service/config"

	_ "github.com/lib/pq"
)


var DB *sql.DB

// opens a connection to PostgreSQL and ensures the schema is ready.
func InitDB() {
	var err error
	DB, err = sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to open database connection: %v", err)
	}

	// Connection pool settings
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(15)
	DB.SetConnMaxLifetime(5 * time.Minute)

	if err = DB.Ping(); err != nil {
		log.Fatalf("❌ Failed to ping database: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL")

	migrateAuth()
}

// applies auth-specific schema changes on top of the existing tables.
func migrateAuth() {
	migrations := []string{
		// password column
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);`,

		// account lockout columns
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS failed_attempts INT DEFAULT 0;`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_until TIMESTAMP;`,

		// refresh tokens table
		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id    UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
			token_hash VARCHAR(64) NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			revoked    BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT NOW()
		);`,

		// Index
		`CREATE INDEX IF NOT EXISTS idx_refresh_token_hash ON refresh_tokens(token_hash);`,

		// OTP entries table
		`CREATE TABLE IF NOT EXISTS otp_entries (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) NOT NULL,
			otp_code VARCHAR(10) NOT NULL,
			pre_auth_token VARCHAR(64) NOT NULL,
			failed_attempts INT DEFAULT 0,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,

		// OTP indexes
		`CREATE INDEX IF NOT EXISTS idx_otp_entries_email ON otp_entries(email);`,
		`CREATE INDEX IF NOT EXISTS idx_otp_entries_pre_auth_token ON otp_entries(pre_auth_token);`,
	}

	for _, m := range migrations {
		if _, err := DB.Exec(m); err != nil {
			log.Fatalf("❌ Auth migration failed: %v", err)
		}
	}

	log.Println("✅ Auth schema ready (Phase 2)")
}
