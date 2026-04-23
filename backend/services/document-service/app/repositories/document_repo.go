package repositories

import (
	"capstone/app/schemas"

	"context"
	"log"
	"time"


	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DocumentRepo struct{
	db *pgxpool.Pool
}

func NewDocumentRepo(db *pgxpool.Pool) *DocumentRepo{
	return &DocumentRepo{
		db: db,
	}
}

const baseQuery = `
	SELECT 
		d.document_id,
		d.filename,
		d.filepath,
		dt.name,
		u.email AS created_by,
		d.updated_at,
		d.status
	FROM documents d
	JOIN document_types dt ON d.document_type_id = dt.document_type_id
	JOIN users u ON d.created_by = u.user_id
	`

func(s *DocumentRepo) fetchingData(query string, args ...any)([]schemas.DocumentResponse, error){
	rows, err := s.db.Query(context.Background(), query, args...)
	if err != nil{
		return nil, err
	}

	defer rows.Close()
	var documents []schemas.DocumentResponse
	for rows.Next() {
		var doc schemas.DocumentResponse

		err := rows.Scan(
			&doc.DocumentId,
			&doc.Filename,
			&doc.FilePath,
			&doc.DocumentType,
			&doc.CreatedBy,
			&doc.UpdatedAt,
			&doc.Status,
		)

		if err != nil {
			log.Print("Error scanning row", err)
			return nil, err
		}
		documents = append(documents, doc)
	}
	
	return documents, nil
}

func(s *DocumentRepo) GetDocuments(limit, offset int)([]schemas.DocumentResponse,error)  {
	query := baseQuery + "ORDER BY d.updated_at DESC LIMIT $1 OFFSET $2"
	return s.fetchingData(query, limit, offset)
}

func(s *DocumentRepo) Get_document_by_status(status string, limit, offset int)([]schemas.DocumentResponse,error){
	query := baseQuery + "WHERE d.status = $1 ORDER BY d.updated_at DESC LIMIT $2 OFFSET $3"
	return s.fetchingData(query, status,limit,offset)
}

func(s *DocumentRepo) Get_document_by_type(typeId uuid.UUID, limit,offset int)([]schemas.DocumentResponse,error){
	query := baseQuery + "WHERE d.document_type_id = $1 ORDER BY d.updated_at DESC LIMIT $2 OFFSET $3"
	return s.fetchingData(query, typeId,limit,offset)
}

func(s *DocumentRepo) Get_document_by_group(groupId uuid.UUID,limit,offset int)([]schemas.DocumentResponse,error){
	query := baseQuery + "WHERE d.group_id = $1 ORDER BY d.updated_at DESC LIMIT $2 OFFSET $3"
	return s.fetchingData(query, groupId, limit,offset)
}

func(s *DocumentRepo) Get_document_by_service(serviceId uuid.UUID,limit,offset int)([]schemas.DocumentResponse,error){
	query := baseQuery + "WHERE d.service_id = $1 ORDER BY d.updated_at DESC LIMIT $2 OFFSET $3"
	return s.fetchingData(query,serviceId, limit, offset)
}

func(s *DocumentRepo) Get_document_by_standard(standardId uuid.UUID,limit,offset int)([]schemas.DocumentResponse,error){
	query := baseQuery + "WHERE d.standard_id = $1 ORDER BY d.updated_at DESC LIMIT $2 OFFSET $3"
	return s.fetchingData(query, standardId, limit,offset)
}

func(s *DocumentRepo) Get_document_by_assessment(assessmentId uuid.UUID,limit,offset int)([]schemas.DocumentResponse,error){
	query := baseQuery + "WHERE d.assessment_id = $1 ORDER BY d.updated_at DESC LIMIT $2 OFFSET $3"
	return s.fetchingData(query, assessmentId,limit,offset)
}

func(s *DocumentRepo) Get_document_by_createdBy(createdById uuid.UUID,limit,offset int)([]schemas.DocumentResponse,error){
	query := baseQuery + "WHERE d.created_by = $1 ORDER BY d.updated_at DESC LIMIT $2 OFFSET $3"
	return s.fetchingData(query, createdById,limit,offset)
}

func(s *DocumentRepo) Get_document_by_id(id uuid.UUID)(schemas.DocumentResponse, error){
	var document schemas.DocumentResponse
	err := s.db.QueryRow(context.Background(),`
	Select 
		d.document_id,
		d.filename,
		d.filepath,
		dt.name,
		u.email AS created_by,
		d.updated_at,
		d.status
	FROM documents d
	JOIN document_types dt ON d.document_type_id = dt.document_type_id
	JOIN users u ON d.created_by = u.user_id
	WHERE d.document_id = $1
	`,id).Scan(
			&document.DocumentId,
			&document.Filename,
			&document.FilePath,
			&document.DocumentType,
			&document.CreatedBy,
			&document.UpdatedAt,
			&document.Status,
		)
		if err != nil{
			return document, err
		}
	return document, nil
}

func (s *DocumentRepo) Create_document(assessmeentId,documentTypeId,createdById,groupId,standardId,serviceId uuid.UUID, filename, filepath string)(string ,error){
	var document_id string
	err := s.db.QueryRow(context.Background(), `
	INSERT INTO documents
	(assessment_id,filename,filepath,document_type_id,status,created_at,updated_at,created_by,group_id,service_id,standard_id) 
	VALUES
	($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING document_id`,assessmeentId,filename,filepath,documentTypeId,"pending",time.Now(),time.Now(),createdById,groupId,serviceId,standardId).Scan(&document_id)

	if err != nil{
		log.Print("Error insert to db: ", err)
		return "",err
	}

	return document_id, nil
}

func (s *DocumentRepo)Update_document(documentId, assessmentId, documentTypeId,createdById, groupId,StandardId, serviceId uuid.UUID, filename, filepath string)(int64, error){
	updateQuery := `
	UPDATE documents SET
		assessment_id = $1,
		filename = $2,
		filepath = $3,
		document_type_id = $4,
		updated_at =$5,
		group_id = $6,
		service_id = $7, 
		standard_id = $8
	WHERE document_id = $9
		AND created_by = $10
		`

	result, err := s.db.Exec(context.Background(), updateQuery, 
	assessmentId,filename,filepath, documentTypeId,time.Now() ,groupId, serviceId, StandardId, documentId, createdById)
	if err != nil{
		log.Print("hidup jokowi: ", err)
		return 0, err
	}
	log.Print("mana ERRORNYA: ", result)
	log.Print("berapa row wok:",result.RowsAffected())
	return result.RowsAffected(), nil
}

func (s *DocumentRepo) Delete_document(documentId, userId uuid.UUID)(int64, error){
	deleteQuery := `
	DELETE FROM documents 
		WHERE document_id = $1
			AND created_by = $2`

	result, err := s.db.Exec(context.Background(), deleteQuery, documentId, userId)
	if err != nil {
		log.Print("error delet ", err)
		return 0, nil
	}

	return result.RowsAffected(), nil
}


func(s *DocumentRepo) Approval_document(documentId, userId uuid.UUID,status string)(int64, error){
	updateStatusQuery:= `
	UPDATE documents SET
		status = $1
			WHERE document_id = $3
				AND created_by = $4
	` 

	result,err := s.db.Exec(context.Background(),updateStatusQuery, status, documentId, userId)
	if err != nil{
		log.Print("Error update status", err)
		return 0, nil
	}

	return result.RowsAffected(),nil
}