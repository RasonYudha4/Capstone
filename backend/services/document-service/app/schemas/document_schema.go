package schemas

import (
	"github.com/google/uuid"
	"time"
)
type Response struct {
	Status bool	`json:"status"`
	Message string	`json:"message"`
}
type DocumentResponse struct {
	DocumentId uuid.UUID `json:"document_id"`
	Filename string		 `json:"filename"`
	DocumentType string	 `json:"document_type"` 	
	CreatedBy string	 `json:"created_by"`
	UpdatedAt time.Time	 `json:"updated_at"`
	Status string		 `json:"status"`
}

type DocumentRequest struct {
	ServicesId string	`form:"service_id"`
	StandardId string 	`form:"standard_id"`
	AssessmentId string	`form:"assessment_id"`
	FileName string 		`form:"filename"`
	DocumentTypeId string`form:"document_type_id"`
	Description string 		`form:"description"`
}


type UploadResponse struct {
	Status bool 	`json:"status"`
	Message string 	`json:"message"`
	FileName string `json:"filename"`
	FileSize int64 	`json:"filesize"`
}

type UpdateRequest struct {
	DocumentId string		`form:"document_id"`
	FileName string 		`form:"filename"`
	Description string 		`form:"description"`
}

type ApprovalRequest struct {
	DocumentId string `json:"document_id"`
	Status string `json:"status"`
}

type GroupStat struct {
    GroupId      uuid.UUID `json:"group_id"`
    GroupName    string    `json:"group_name"`
    TotalFiles   int       `json:"total_files"`
    EmptySections int      `json:"empty_sections"`
}

type StatusStat struct {
    Approved int `json:"approved"`
    Pending  int `json:"pending"`
    Rejected int `json:"rejected"`
}

type StatsResponse struct {
	Status  bool        `json:"status"`
	Groups  []GroupStat `json:"groups"`
	Stats   StatusStat  `json:"stats"`
	Total   int         `json:"total"`
}

