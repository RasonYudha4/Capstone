package repositories

import (
	"capstone/app/core/db"
	"capstone/app/core/objectStorage"
	"capstone/app/schemas"

	"strings"
	"context"
	"log"
	"time"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/minio/minio-go/v7"
)

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

func fetchingData(query string, args ...any)([]schemas.DocumentResponse, error){
	rows, err := db.DB.Query(context.Background(), query, args...)
	if err != nil{
		log.Print("Error fetching data from db: ",err)
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
	if len(documents) == 0 {
		return nil, pgx.ErrNoRows
	}
	return documents, nil
}

func GetDocuments(limit, offset int)([]schemas.DocumentResponse,error)  {
	query := baseQuery + "ORDER BY d.updated_at DESC LIMIT $1 OFFSET $2"
	return fetchingData(query, limit, offset)
}

func Get_document_by_status(status string, limit, offset int)([]schemas.DocumentResponse,error){
	query := baseQuery + "WHERE d.status = $1 ORDER BY d.updated_at DESC LIMIT $2 OFFSET $3"
	return fetchingData(query, status,limit,offset)
}

func Get_document_by_type(typeId uuid.UUID, limit,offset int)([]schemas.DocumentResponse,error){
	query := baseQuery + "WHERE d.document_type_id = $1 ORDER BY d.updated_at DESC LIMIT $2 OFFSET $3"
	return fetchingData(query, typeId,limit,offset)
}

func Get_document_by_group(groupId uuid.UUID,limit,offset int)([]schemas.DocumentResponse,error){
	query := baseQuery + "WHERE d.group_id = $1 ORDER BY d.updated_at DESC LIMIT $2 OFFSET $3"
	return fetchingData(query, groupId, limit,offset)
}

func Get_document_by_service(serviceId uuid.UUID,limit,offset int)([]schemas.DocumentResponse,error){
	query := baseQuery + "WHERE d.service_id = $1 ORDER BY d.updated_at DESC LIMIT $2 OFFSET $3"
	return fetchingData(query,serviceId, limit, offset)
}

func Get_document_by_standard(standardId uuid.UUID,limit,offset int)([]schemas.DocumentResponse,error){
	query := baseQuery + "WHERE d.standard_id = $1 ORDER BY d.updated_at DESC LIMIT $2 OFFSET $3"
	return fetchingData(query, standardId, limit,offset)
}

func Get_document_by_assessment(assessmentId uuid.UUID,limit,offset int)([]schemas.DocumentResponse,error){
	query := baseQuery + "WHERE d.assessment_id = $1 ORDER BY d.updated_at DESC LIMIT $2 OFFSET $3"
	return fetchingData(query, assessmentId,limit,offset)
}

func Get_document_by_createdBy(createdById uuid.UUID,limit,offset int)([]schemas.DocumentResponse,error){
	query := baseQuery + "WHERE d.created_by = $1 ORDER BY d.updated_at DESC LIMIT $2 OFFSET $3"
	return fetchingData(query, createdById,limit,offset)
}

func Get_document_by_id(id uuid.UUID)(schemas.DocumentResponse, error){
	var document schemas.DocumentResponse
	err := db.DB.QueryRow(context.Background(),`
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


func Create_document(r *gin.Context) (schemas.UploadResult,error){
	if err := r.Request.ParseMultipartForm(10 << 20); err != nil {
		log.Print("Error parsing")
		return schemas.UploadResult{}, nil
	}

	var documentRequest schemas.DocumentRequest
	if err := r.ShouldBind(&documentRequest); err != nil {
		log.Print("Error receive data", err )
		return schemas.UploadResult{}, nil
	}

	src, err := r.FormFile("uploadedFile")
	if err != nil {
		log.Print("Error receive file", err)
		return schemas.UploadResult{}, nil
	}

	file, err := src.Open()
	if err != nil {
		log.Print("Error open file", err)
		return schemas.UploadResult{}, nil
	}
	defer file.Close()

	s := strings.Fields(src.Filename)
	formattedFileName := strings.Join(s,"_")
	bucketName := "testing"
	objectName := formattedFileName
	_, err = objectStorage.MinioClient.PutObject(context.Background(), bucketName, objectName,file,src.Size, minio.PutObjectOptions{
		ContentType: src.Header.Get("Content-type"),
		ContentDisposition: "inline",
	})

	if err != nil{
		log.Print(err)
		return schemas.UploadResult{}, nil
	}
	log.Print("Success Upload to document file to bucket")

	filePath := fmt.Sprintf("http://localhost:9000/testing/%s", formattedFileName)

	insertQuery := `
	INSERT INTO documents
		(document_id,assessment_id, filename, filepath, document_type_id, status, created_at, updated_at, created_by, group_id, service_id, standard_id) VALUES
		($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)	
	`
	document_id := uuid.New()
	assessmentId, err := uuid.Parse(documentRequest.AssessmentId)
	documentTypeId, err := uuid.Parse(documentRequest.DocumentTypeId)
	userId, err := uuid.Parse(documentRequest.UserId)
	GroupId, err := uuid.Parse(documentRequest.GroupId)
	serviceId, err := uuid.Parse(documentRequest.ServicesId)
	standardId, err := uuid.Parse(documentRequest.StandardId)

	_, err = db.DB.Exec(context.Background(),insertQuery,document_id.String(), assessmentId, documentRequest.FileName,filePath, documentTypeId, "pending",time.Now(), time.Now(),userId, GroupId, serviceId, standardId)
	if err != nil{
		log.Print("Eror wok: ",err )
		return schemas.UploadResult{}, nil
	}
	
	url, err := objectStorage.MinioClient.PresignedGetObject(
    context.Background(),
    "testing",
    formattedFileName,
    time.Hour * 1,
    nil,
	)
		
	log.Print(url)
	result := schemas.UploadResult{
		Status: true,
		Message: "Upload successfull",
		FileName: src.Filename,
		FileSize: src.Size,
	}
	return result, nil
}

func Update_document(r *gin.Context)(schemas.UpdateResponse, error){



	return schemas.UpdateResponse{}, nil
}

func approval_document(){

}

