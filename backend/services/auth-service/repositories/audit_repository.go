package repositories

import (
	"database/sql"
)

type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Save(action, description string, userID *string, source string) error {
	_, err := r.db.Exec(
		`INSERT INTO audit (audit_id, action, description, user_id, source, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1, $2, $3, $4::audit_source, NOW(), NOW())`,
		action, description, userID, source,
	)
	return err
}
