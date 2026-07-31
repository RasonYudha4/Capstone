package schemas

import(
	"github.com/google/uuid"
	"time"
)

type AuditResponse struct{
	AuditId uuid.UUID 	`json:"audit_id"`
	Action string	  	`json:"action"`
	Description string	`json:"description"`
	UserName string	  	`json:"username"`
	DocumentName *string `json:"document_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time	`json:"updated_at"`
}