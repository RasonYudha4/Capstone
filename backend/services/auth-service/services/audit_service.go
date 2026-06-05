package services

import (
	"log"

	"auth-service/repositories"
)

// logs authentication events to the `audit` table
type AuditService struct {
	auditRepo *repositories.AuditRepository
}

// creates an AuditService backed by the given AuditRepository.
func NewAuditService(auditRepo *repositories.AuditRepository) *AuditService {
	return &AuditService{auditRepo: auditRepo}
}

// records an authentication event in the audit table.
func (s *AuditService) Log(auditType, action string, userID *string, source string) {
	err := s.auditRepo.Save(auditType, action, userID, source)
	if err != nil {
		log.Printf("⚠️  Audit log failed (type=%s, action=%s): %v", auditType, action, err)
	}
}
