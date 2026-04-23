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

func NewAuditRepo(db *pgxpool.Pool) *AuditRepo{
	return &AuditRepo{
		db : db,
	}
}

func(a *AuditRepo) GetAudit()([]schemas.AuditResponse, error){
	rows, err := a.db.Query(context.Background(), `
	SELECT 
		a.audit_id, 
		a.type,
		a.action,
		u.email AS userName,
		d.filename AS documentName,
		a.source,
		a.created_at,
		a.updated_at
	FROM audit a
	JOIN users u ON u.user_id = a.user_id
	JOIN documents d ON d.document_id = a.document_id
	`)
	if err != nil{
		log.Print("Error fetching data from db :", err)
		return nil, err
	}
	defer rows.Close()

	var audits []schemas.AuditResponse

	for rows.Next(){
		var audit schemas.AuditResponse

		err := rows.Scan(
			&audit.AuditId,
			&audit.AuditType,
			&audit.Action,
			&audit.UserName,
			&audit.DocumentName,
			&audit.Source,
			&audit.CreatedAt,
			&audit.UpdatedAt,
		)
		if err != nil {
			log.Print("error scanning row :",err)
			return nil, err
		}
	
		audits = append(audits, audit)
	}

	return audits, err
}

func (a *AuditRepo)SaveAudit(types,action string, userId, documentId uuid.UUID, source string, createdAt, updatedAt time.Time)(error){
	insertQuery := `
	INSERT INTO audit
	(type,action,user_id,document_id,source,created_at,updated_at) VALUES
	($1,$2,$3,$4,$5,$6,$7)
	`
	
	_, err := a.db.Exec(context.Background(),insertQuery,types,action,userId,documentId,source,createdAt,updatedAt)
	if err != nil {
		log.Print("Error Insert Audit log: ", err)
		return err
	}

	return nil
}
