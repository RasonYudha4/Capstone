package schemas

import(
	"github.com/google/uuid"
	"time"
)

type AuditResponse struct{
	AuditId uuid.UUID 	`json:"audit_id"`
	AuditType string  	`json:"audit_type"`
	Action string	  	`json:"action"`
	UserName string	  	`json:"username"`
	DocumentName string `json:"document_name"`
	Source string		`json:"source"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time	`json:"updated_at"`
}