package repositories

import (
	"capstone/app/schemas"
	"mime/multipart"

	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
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

type filter string

const (
	filterByService filter = "d.service_id"
	filterByStandard filter = "d.standard_id"
	filterByAssessment filter = "d.assessment_id"
)


const baseQuery = `
	SELECT 
		d.document_id,
		d.filename,
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



func (s *DocumentRepo) filterFetch(field filter, fieldId, createdById uuid.UUID, limit, offset int, role string)([]schemas.DocumentResponse, error){

	if role == "master-admin"{
		return s.fetchingData(fmt.Sprintf("%s WHERE %s = $1 AND is_deleted = false ORDER BY d.updated_at DESC LIMIT $2 OFFSET $3",
			baseQuery,field), fieldId, limit,offset)
	}
	return s.fetchingData(fmt.Sprintf("%s WHERE %s = $1 AND d.created_by = $2 AND is_deleted = false ORDER BY d.updated_at DESC LIMIT $3 OFFSET $4", baseQuery, field), fieldId, createdById, limit, offset)

}


func(s *DocumentRepo) GetDocuments(limit, offset int)([]schemas.DocumentResponse,error)  {
	query := baseQuery + "WHERE is_deleted = false AND status = 'approved' ORDER BY d.updated_at DESC LIMIT $1 OFFSET $2 "
	return s.fetchingData(query, limit, offset)
}

func(s *DocumentRepo) Get_document_by_status(status string, limit, offset int)([]schemas.DocumentResponse,error){
	query := baseQuery + "WHERE d.status = $1 AND is_deleted = false ORDER BY d.updated_at DESC LIMIT $2 OFFSET $3"
	return s.fetchingData(query, status,limit,offset)
}

func (s *DocumentRepo) Get_documents_by_type(limit, offset int) ([]schemas.DocumentResponse, error) {
		query := baseQuery + `
		WHERE dt.is_public = true
		  AND d.is_deleted = false
		ORDER BY d.updated_at DESC
		LIMIT $1 OFFSET $2`

	return s.fetchingData(query, limit, offset)
}

func(s *DocumentRepo) Get_document_by_group(groupId uuid.UUID,limit,offset int)([]schemas.DocumentResponse,error){
	query := baseQuery + "WHERE d.group_id = $1 AND is_deleted = false ORDER BY d.updated_at DESC LIMIT $2 OFFSET $3"
	return s.fetchingData(query, groupId, limit,offset)
}

func(s *DocumentRepo) Get_document_by_service(serviceId,createdById uuid.UUID,limit,offset int,role string)([]schemas.DocumentResponse,error){
	return s.filterFetch(filterByService,serviceId,createdById,limit, offset, role)
}

func(s *DocumentRepo) Get_document_by_standard(standardId, createdById uuid.UUID,limit,offset int, role string)([]schemas.DocumentResponse,error){
	return s.filterFetch(filterByStandard, standardId, createdById, limit, offset, role)
}

func(s *DocumentRepo) Get_document_by_assessment(assessmentId,createdById uuid.UUID,limit,offset int, role string)([]schemas.DocumentResponse,error){
	return s.filterFetch(filterByAssessment, assessmentId, createdById, limit, offset, role)
}

func(s *DocumentRepo) Get_document_by_createdBy(createdById uuid.UUID,limit,offset int)([]schemas.DocumentResponse,error){
	query := baseQuery + "WHERE d.created_by = $1 AND is_deleted = false ORDER BY d.updated_at DESC LIMIT $2 OFFSET $3"
	return s.fetchingData(query, createdById,limit,offset)
}

func(s *DocumentRepo) Get_document_by_id(documentId, createdById uuid.UUID, role string)(string,schemas.UploadResponse, error){
	var query string
	var filePath string
	
	query = `
	Select 
		filepath
		FROM documents 
	WHERE document_id = $1
	`
	switch role{
	case "master-admin":
		query = query + "AND is_deleted = false"
			err := s.db.QueryRow(context.Background(), query, documentId).Scan(&filePath)
		if err != nil{
			if err == pgx.ErrNoRows {
        		return "",schemas.UploadResponse{} ,nil 
    			}
				return filePath,schemas.UploadResponse{} ,err
		}
		return filePath,schemas.UploadResponse{
			Status: true,
			Message: "Getting the file path",
			FileName: filePath,
			FileSize: 0,
		} ,nil
	
	case "admin":
		query = query + "AND created_by = $2 AND is_deleted = false"
		err := s.db.QueryRow(context.Background(), query, documentId,createdById).Scan(&filePath)
		if err != nil{
			if err == pgx.ErrNoRows {
        		return "",schemas.UploadResponse{} ,nil 
    		}
			return filePath,schemas.UploadResponse{} ,err
	}
		return filePath,schemas.UploadResponse{
			Status: true,
			Message: "Getting the file path",
			FileName: filePath,
			FileSize: 0,
		} ,nil
	}
	
	return filePath,schemas.UploadResponse{} ,nil
}

func (s *DocumentRepo) Create_document(assessmeentId,documentTypeId,createdById,groupId,standardId,serviceId uuid.UUID, filename, filepath string, role string)(string,bool,bool,error){
	
	if role != "master-admin" {
		var authorized bool
		err := s.db.QueryRow(context.Background(), `
			SELECT EXISTS (
				SELECT 1
				FROM services s
				JOIN standard st ON st.service_id = s.service_id
				JOIN assessment a ON a.standard_id = st.standard_id
				JOIN users u ON u.group_id = s.group_id
				WHERE s.service_id = $1
				AND st.standard_id = $2
				AND a.assessment_id = $3
				AND u.user_id = $4
			)`, serviceId, standardId, assessmeentId, createdById).Scan(&authorized)
		if err != nil {
			return "", false, false, err
		}
		if !authorized {
			return "", false, false, nil
		}
	}
	
	var serviceGroupId uuid.UUID
	err := s.db.QueryRow(context.Background(),
	`SELECT group_id from services
		WHERE service_id = $1 `, serviceId).Scan(&serviceGroupId)
	if err!= nil{
		log.Print("error getting group id from service", err)
		return "", false,false,nil
	}
	
	var document_id string
	err = s.db.QueryRow(context.Background(), `
	INSERT INTO documents
	(assessment_id,filename,filepath,document_type_id,status,created_at,updated_at,created_by,group_id,service_id,standard_id,is_deleted) 
	VALUES
	($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) RETURNING document_id`,assessmeentId,filename,filepath,documentTypeId,"pending",time.Now(),time.Now(),createdById,serviceGroupId,serviceId,standardId,false).Scan(&document_id)

	if err != nil{
		log.Print("error inserting to db", err)
		return "",false ,false,err
	}

	return document_id,true,true,nil
}

func (s *DocumentRepo) Update_document(documentId, createdById uuid.UUID, filename, filepath, description string) (int64, error) {
	toText := func(s string) pgtype.Text {
		if strings.TrimSpace(s) == "" {
			return pgtype.Text{Valid: false}
		}
		return pgtype.Text{String: s, Valid: true}
	}

	result, err := s.db.Exec(context.Background(), `
		UPDATE documents SET
			filename = COALESCE($1, filename),
			filepath = COALESCE($2, filepath),
			updated_at = NOW()
		WHERE document_id = $3
		AND is_deleted = false
	`, toText(filename), toText(filepath), documentId)

	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
func (s *DocumentRepo) Delete_document(documentId, userId uuid.UUID, role string)(string,bool ,error){
	var filename string
	if role == "master-admin"{
	deleteQuery := `
	UPDATE documents SET 
		is_deleted = true,
		updated_at = NOW()
		WHERE document_id = $1
			AND is_deleted = false
		RETURNING filename`

	err := s.db.QueryRow(context.Background(),deleteQuery, documentId).Scan(&filename)
	if err != nil {
		if err == pgx.ErrNoRows{
		return "",false, nil
		}
		return "",false, err
	}
	}else {
		deleteQuery := `
	UPDATE documents SET 
		is_deleted = true 
		WHERE document_id = $1
			AND created_by = $2
			AND is_deleted = false
		RETURNING filename`

	err := s.db.QueryRow(context.Background(), deleteQuery, documentId, userId).Scan(&filename)
	if err != nil {
		if err == pgx.ErrNoRows{
		return "",false, nil
		}
		return "",false, err
	}
	}
	return filename, true ,nil
}


func (s *DocumentRepo) Approval_document(documentId uuid.UUID, status string, filePath string, file multipart.File) (int64, error) {

    var result pgconn.CommandTag
    var err error

    if status == "approved" && file != nil {
        query := `
            UPDATE documents SET
                status = $1,
                filepath = $2,
                updated_at = NOW()
            WHERE document_id = $3
            AND is_deleted = false
        `
        result, err = s.db.Exec(context.Background(), query, status, filePath, documentId)
    }else if status == "approve" && file == nil{
		 query := `
            UPDATE documents SET
                status = $1,
                updated_at = NOW()
            WHERE document_id = $2
            AND is_deleted = false
        `
        result, err = s.db.Exec(context.Background(), query, status, documentId)
	}else {
        query := `
            UPDATE documents SET
                status = $1,
                updated_at = NOW()
            WHERE document_id = $2
            AND is_deleted = false
        `
        result, err = s.db.Exec(context.Background(), query, status, documentId)
    }

    if err != nil {
        log.Print("Error update status", err)
        return 0, err
    }

    return result.RowsAffected(), nil
}

func (s *DocumentRepo) Get_document_owner_email(documentId uuid.UUID) (string,uuid.UUID, error){
	query := `SELECT u.email, d.created_by from documents d JOIN users u ON d.created_by=u.user_id WHERE document_id = $1`

	var ownerEmail string
	var ownerId uuid.UUID
	err := s.db.QueryRow(context.Background(), query, documentId).Scan(&ownerEmail,&ownerId)
	if err != nil {
		return "",uuid.Nil,err
	}
	return ownerEmail,ownerId,nil
}

func (s *DocumentRepo) Get_stats(userId uuid.UUID, role string) ([]schemas.GroupStat, schemas.StatusStat, error) {
    groupQuery := `
        SELECT 
            g.group_id,
            g.group_name,
            COUNT(d.document_id) FILTER (WHERE d.is_deleted = false) AS total_files,
            0 AS empty_sections  
        FROM groups g
        LEFT JOIN documents d ON d.group_id = g.group_id
        GROUP BY g.group_id, g.group_name
        ORDER BY g.group_name
    `
    rows, err := s.db.Query(context.Background(), groupQuery)
    if err != nil {
        return nil, schemas.StatusStat{}, err
    }
    defer rows.Close()

    var groups []schemas.GroupStat
    for rows.Next() {
        var group schemas.GroupStat
        if err := rows.Scan(&group.GroupId, &group.GroupName, &group.TotalFiles, &group.EmptySections); err != nil {
            return nil, schemas.StatusStat{}, err
        }
        groups = append(groups, group)
    }

    var stats schemas.StatusStat
	switch role {
	case "master-admin":
		err = s.db.QueryRow(context.Background(), `
        	SELECT
            	COUNT(*) FILTER (WHERE status = 'approved' AND is_deleted = false),
            	COUNT(*) FILTER (WHERE status = 'pending'  AND is_deleted = false),
            	COUNT(*) FILTER (WHERE status = 'rejected' AND is_deleted = false)
        	FROM documents
    	`,).Scan(&stats.Approved, &stats.Pending, &stats.Rejected)
    	if err != nil {
        	return nil, schemas.StatusStat{}, err
    	}

	case "admin":
    	err = s.db.QueryRow(context.Background(), `
        	SELECT
            	COUNT(*) FILTER (WHERE status = 'approved' AND is_deleted = false),
            	COUNT(*) FILTER (WHERE status = 'pending'  AND is_deleted = false),
            	COUNT(*) FILTER (WHERE status = 'rejected' AND is_deleted = false)
        	FROM documents
				WHERE created_by = $1
    	`,userId).Scan(&stats.Approved, &stats.Pending, &stats.Rejected)
    	if err != nil {
        	return nil, schemas.StatusStat{}, err
    	}
	}
    return groups, stats, nil
}

func (s *DocumentRepo) Get_document_fileName(documentId, createdById uuid.UUID)(string, error){
	var fileName string 

	query := `
	SELECT
		filename
		FROM documents
		WHERE document_id = $1
		AND created_by = $2
		AND is_deleted = false
	`

	err := s.db.QueryRow(context.Background(),query,documentId,createdById).Scan(&fileName)
	if err != nil{
		log.Print("Error fetching filepath: ", err)
		return "", err
	}

	return fileName, nil
}

func (s *DocumentRepo) Get_admin_email()([]string,[]uuid.UUID, error){
	query := `
	SELECT 
		user_id,
		email
	FROM users
		WHERE role = 'master-admin'`

	rows, err := s.db.Query(context.Background(),query)
	if err != nil{
		return nil,nil,err
	}
	defer rows.Close()

	var masterAdminEmail []string
	var masterAdminId	[]uuid.UUID
		for rows.Next(){
			var email string
			var id uuid.UUID
			if err := rows.Scan(&id,&email); err != nil {
				log.Print("Error scanning row", err)
				return nil,nil,err
			}
			masterAdminEmail = append(masterAdminEmail, email)
			masterAdminId = append(masterAdminId, id)
		}
	return masterAdminEmail,masterAdminId, nil
}

func (s *DocumentRepo) Check_document_owner(documentId, userId uuid.UUID) (bool, error) {
	var authorized bool
	var status string
	err := s.db.QueryRow(context.Background(), `
		SELECT EXISTS (
			SELECT 1
			FROM documents doc
			JOIN services sv ON sv.service_id = doc.service_id
			JOIN users u ON u.group_id = sv.group_id
			WHERE doc.document_id = $1
			AND u.user_id = $2
			AND doc.created_by = $2
			AND doc.is_deleted = false
			AND doc.status = 'pending'
		),
		(SELECT status FROM documents WHERE document_id = $1)	
		`, documentId, userId).Scan(&authorized, &status)

	if err != nil {
		return false, err
	}

	if status == "approved"{
		return false, fmt.Errorf("Cannnot edit an approve document")
	}

	return authorized, nil
}