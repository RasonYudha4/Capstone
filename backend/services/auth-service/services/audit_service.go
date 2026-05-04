package services

import (
	"database/sql"
	"log"
)

// logs authentication events to the `audit` table
type AuditService struct {
	db *sql.DB
}

// creates an AuditService backed by the given database.
func NewAuditService(db *sql.DB) *AuditService {
	return &AuditService{db: db}
}

// records an authentication event in the audit table.
func (s *AuditService) Log(auditType, action string, userID *string, source string) {
	_, err := s.db.Exec(
		`INSERT INTO audit (audit_id, type, action, user_id, source, created_at, updated_at)
		 VALUES (gen_random_uuid(), $1::audit_type, $2, $3, $4::audit_source, NOW(), NOW())`,
		auditType, action, userID, source,
	)
	if err != nil {
		log.Printf("⚠️  Audit log failed (type=%s, action=%s): %v", auditType, action, err)
	}
}
