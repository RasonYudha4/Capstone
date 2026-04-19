package repositories

import (
	"capstone/app/core/db"
	"capstone/app/schemas"
	"context"
	"log"
)

func GetAudit()([]schemas.AuditResponse, error){
	rows, err := db.DB.Query(context.Background(), `
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

