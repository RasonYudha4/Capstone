package repositories

import (
	"capstone/app/schemas"
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepo struct {
	db *pgxpool.Pool
}

func NewAuditRepo(db *pgxpool.Pool) *AuditRepo {
	return &AuditRepo{
		db: db,
	}
}

func (a *AuditRepo) GetAudit() ([]schemas.AuditResponse, error) {
	rows, err := a.db.Query(context.Background(), `
	SELECT 
		a.audit_id, 
		a.action,
		a.description,
		u.email AS userName,
		a.document_id AS documentId,
		d.filename AS documentName,
		a.created_at,
		a.updated_at
	FROM audit a
	JOIN users u ON u.user_id = a.user_id
	LEFT JOIN documents d ON d.document_id = a.document_id
	WHERE a.source = 'client'
	`)
	if err != nil {
		log.Print("Error fetching data from db :", err)
		return nil, err
	}
	defer rows.Close()

	var audits []schemas.AuditResponse

	for rows.Next() {
		var audit schemas.AuditResponse

		err := rows.Scan(
			&audit.AuditId,
			&audit.Action,
			&audit.Description,
			&audit.UserName,
			&audit.DocumentId,
			&audit.DocumentName,
			&audit.CreatedAt,
			&audit.UpdatedAt,
		)
		if err != nil {
			log.Print("error scanning row :", err)
			return nil, err
		}

		audits = append(audits, audit)
	}

	return audits, err
}

func (a *AuditRepo) SaveAudit(action, description string, userId, documentId uuid.UUID, source string, createdAt, updatedAt time.Time) (string, error) {
	var filename string

	insertQuery := `
	INSERT INTO audit
	(action,description,user_id,document_id,source,created_at,updated_at) VALUES
	($1,$2,$3,$4,$5,$6,$7)
	returning
		(SELECT filename FROM documents WHERE document_id = $4)
	`

	err := a.db.QueryRow(context.Background(), insertQuery, action, description, userId, documentId, source, createdAt, updatedAt).Scan(&filename)
	if err != nil {
		log.Print("Error Insert Audit log: ", err)
		return "", err
	}

	return filename, nil
}
