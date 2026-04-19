package models

import (
	"time"

	"github.com/google/uuid"
)

type Audit struct{
	AuditId uuid.UUID
	Type string
	Action string 
	UserId uuid.UUID
	DocumentId uuid.UUID
	Source string
	CreatedAt time.Time
	UpdatedAt time.Time
}

