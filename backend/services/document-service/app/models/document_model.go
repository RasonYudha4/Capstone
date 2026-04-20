package models

import (
	"github.com/google/uuid"
	"time"
)

type Document struct {
	DocumentId uuid.UUID
	AssessmentId uuid.UUID
	Filename string 
	FilePath string
	DocumentTypeId int
	Status string
	CreatedAt time.Time
	UpdatedAt time.Time
	CreatedBy uuid.UUID
	GroupId uuid.UUID
	ServiceId uuid.UUID
	StandardId uuid.UUID
}


