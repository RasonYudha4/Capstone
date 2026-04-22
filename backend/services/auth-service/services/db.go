package services

import (
	"database/sql"
	"log"

	"auth-service/config"

	_ "github.com/lib/pq"
)

// DB is the package-level database connection pool.
// Initialised once at startup and shared across all services.
var DB *sql.DB

// InitDB opens a connection to PostgreSQL and creates required tables.
// Fatals on failure — the server cannot run without a database.
func InitDB() {
	var err error
	DB, err = sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to open database connection: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("❌ Failed to ping database: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL")

	createTables()
}

// createTables ensures the required tables exist.
// Uses IF NOT EXISTS so it's safe to call on every startup.
func createTables() {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email         VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		name          VARCHAR(255) NOT NULL,
		role          VARCHAR(20)  NOT NULL DEFAULT 'staff'
		              CHECK (role IN ('staff', 'admin', 'master_admin')),
		is_active     BOOLEAN NOT NULL DEFAULT TRUE,
		created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);`

	if _, err := DB.Exec(query); err != nil {
		log.Fatalf("❌ Failed to create tables: %v", err)
	}

	log.Println("✅ Database tables ready")
}
