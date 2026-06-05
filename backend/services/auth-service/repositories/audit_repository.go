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

func (r *AuditRepository) Save(auditType, action string, userID *string, source string) error {
	_, err := r.db.Exec(
		`INSERT INTO audit (audit_id, type, action, user_id, source, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1::audit_type, $2, $3, $4::audit_source, NOW(), NOW())`,
		auditType, action, userID, source,
	)
	return err
}
