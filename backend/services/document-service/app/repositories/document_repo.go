package repositories

import (
	"capstone/app/schemas"
	"mime/multipart"

	"context"
	"fmt"
	"strings"
	"time"
	"log"
	"path/filepath"

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
        dt.name AS document_type,
        u.email AS created_by,
        d.updated_at,
        d.status,
        s.service_code,
        st.standard_code,
        a.assessment_code
    FROM documents d
    JOIN document_types dt ON d.document_type_id = dt.document_type_id
    JOIN users u ON d.created_by = u.user_id
    LEFT JOIN services s ON d.service_id = s.service_id
    LEFT JOIN standard st ON d.standard_id = st.standard_id
    LEFT JOIN assessment a ON d.assessment_id = a.assessment_id
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
			&doc.ServiceCode,
			&doc.StandardCode,
			&doc.AssessmentCode,
		)

		if err != nil {
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

	default:
		query = query + " AND is_deleted = false"
    	err := s.db.QueryRow(context.Background(), query, documentId).Scan(&filePath)
    	if err != nil {
        	if err == pgx.ErrNoRows {
            return "", schemas.UploadResponse{}, nil
        }
        return "", schemas.UploadResponse{}, err
    	}
    	return filePath, schemas.UploadResponse{}, nil
	}

	return filePath,schemas.UploadResponse{
		Status: true,
		Message: "Getting the file path",
		FileName: filePath,
		FileSize: 0,
		} ,nil
}
	

func (s *DocumentRepo) Create_document(assessmentId, documentTypeId, createdById, standardId, serviceId uuid.UUID, filename, filepath, fileHash, objectId string, role string, isPublic bool) (string, bool, bool, error) {

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
			)`, serviceId, standardId, assessmentId, createdById).Scan(&authorized)
		if err != nil {
			return "", false, false, err
		}
		if !authorized {
			return "", false, false, nil
		}
	}

	var serviceGroupId uuid.UUID
	err := s.db.QueryRow(context.Background(),
		`SELECT group_id FROM services WHERE service_id = $1`, serviceId).Scan(&serviceGroupId)
	if err != nil {
		return "", false, false, fmt.Errorf("error getting group_id: %w", err)
	}

	status := "pending"
	if isPublic {
		status = "approved"
	}

	var documentId string
	err = s.db.QueryRow(context.Background(), `
		INSERT INTO documents
			(assessment_id, filename, filepath,filehash, object_id, document_type_id, status, created_at, updated_at, created_by, group_id, service_id, standard_id, is_deleted)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,$13,$14)
		RETURNING document_id`,
		assessmentId, filename, filepath,fileHash,objectId, documentTypeId, status,
		time.Now(), time.Now(), createdById, serviceGroupId, serviceId, standardId, false,
	).Scan(&documentId)
	if err != nil {
		return "", false, false, err
	}

	return documentId, true, true, nil
}

func (s *DocumentRepo) Update_document(documentId, createdById uuid.UUID, filename, filepath, description, objectHash string) (int64, error) {
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
			filehash = COALESCE($3, filehash),
			updated_at = NOW(),
			status = CASE
				WHEN status = 'rejected' AND $3 IS NOT NULL AND $3 != (
            SELECT filehash FROM documents WHERE document_id = $4
        	) THEN 'pending'
        		ELSE status
			END	
		WHERE document_id = $4
		AND is_deleted = false
	`, toText(filename), toText(filepath), toText(objectHash) ,documentId)

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
    }else if status == "approved" && file == nil{
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
        return 0, err
    }

    return result.RowsAffected(), nil
}

func (s *DocumentRepo) Get_document_owner_email(documentId uuid.UUID) (string,uuid.UUID,error){
	query := `SELECT u.email, created_by  from documents d JOIN users u ON d.created_by=u.user_id WHERE document_id = $1`

	var ownerEmail string
	var ownerId uuid.UUID
	err := s.db.QueryRow(context.Background(), query, documentId).Scan(&ownerEmail,&ownerId)
	if err != nil {
		return "",uuid.Nil,err
	}
	return ownerEmail,ownerId, nil
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

func (s *DocumentRepo) Get_document_filePath(documentId uuid.UUID)(string, error){
	var filePath string 

	query := `
	SELECT
		filepath FROM documents
		WHERE document_id = $1
		AND is_deleted = false
	`

	err := s.db.QueryRow(context.Background(),query,documentId).Scan(&filePath)
	if err != nil{
		return "Error fetching filepath", err
	}
	log.Print("filenya jalan", filePath)
	return filePath, nil
}

func (s *DocumentRepo) Get_admin_email()(string,uuid.UUID, error){
	var masterAdminEmail string
	var masterAdminId uuid.UUID
	query := `
	SELECT 
		user_id,
		email
	FROM users
		WHERE role = 'master-admin'`

	err := s.db.QueryRow(context.Background(),query).Scan(&masterAdminId, &masterAdminEmail)
	if err != nil{
		return "", uuid.Nil,err
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

func (s *DocumentRepo) Get_group_name_by_serviceId(serviceId string) (string, error) {
	var groupName string
	err := s.db.QueryRow(context.Background(), `
		SELECT g.group_name 
		FROM groups g
		JOIN services sv ON sv.group_id = g.group_id
		WHERE sv.service_id = $1
	`, serviceId).Scan(&groupName)
	if err != nil {
		return "", err
	}
	return groupName, nil
}

func (s *DocumentRepo) Get_service_name_byid(serviceId string) (string, error) {
	var description string
	err := s.db.QueryRow(context.Background(), `
		SELECT description FROM services WHERE service_id = $1
	`, serviceId).Scan(&description)
	return description, err
}

func (s *DocumentRepo) Get_standard_name_byid(standardId string) (string, string, error) {
    var description, code string
    err := s.db.QueryRow(context.Background(), `
        SELECT description, standard_code FROM standard WHERE standard_id = $1
    `, standardId).Scan(&description, &code)
    return description, code, err
}

func (s *DocumentRepo) Get_assessment_name_byid(assessmentId string) (string, string, error) {
    var description, code string
    err := s.db.QueryRow(context.Background(), `
        SELECT description, assessment_code FROM assessment WHERE assessment_id = $1
    `, assessmentId).Scan(&description, &code)
    return description, code, err
}

func (s *DocumentRepo) Get_document_type_byid(documentTypeId string) (string, error) {
	var name string
	err := s.db.QueryRow(context.Background(), `
		SELECT name FROM document_types WHERE document_type_id = $1
	`, documentTypeId).Scan(&name)
	return name, err
}

func (s *DocumentRepo) Is_public_document(documentTypeId uuid.UUID)(bool, error){
	var isPublic bool
	err := s.db.QueryRow(context.Background(),
		`
		SELECT is_public FROM document_types
			WHERE document_type_id = $1
		`,documentTypeId).Scan(&isPublic)
	if err != nil {
		return false, err
	}
	return isPublic, nil
}

func(s *DocumentRepo) Check_document_name(filename string, header *multipart.FileHeader)(bool, error){
	var isSameName bool
	ext := filepath.Ext(header.Filename)
	err := s.db.QueryRow(context.Background(),
		`
		SELECT EXISTS(
			SELECT FROM documents
				WHERE filename = $1
				AND is_deleted = false)
		`,filename + ext).Scan(&isSameName)
	if err != nil{
		return false, err
	}

	return isSameName, nil
}

func (s *DocumentRepo) Document_is_approved(documentId uuid.UUID)(bool, error){
	var isApproved bool
	err := s.db.QueryRow(context.Background(),
	`
	SELECT EXISTS(
		SELECT FROM documents
			WHERE document_id = $1
			AND status = 'approved'
	)
	`, documentId).Scan(&isApproved)
	if err != nil{
		log.Print("error", err)
		return false, err
	}
	log.Print("is approve :", isApproved)
	return isApproved, nil
}

func(s *DocumentRepo) Check_document_hash(documentId uuid.UUID)(string, error){
	var hash string
	err := s.db.QueryRow(context.Background(), 
	`
		SELECT filehash FROM documents
			WHERE document_id = $1
	`, documentId).Scan(&hash)
	if err != nil{
		return "", err
	}

	return hash, nil
}

func (s *DocumentRepo) Get_object_id(documentId uuid.UUID)(string, error){
	var objectId string 
	err := s.db.QueryRow(context.Background(), `
		SELECT object_id FROM documents
			WHERE document_id = $1 
	`, documentId).Scan(&objectId)

	if err != nil {
		return "", err 
	}
	return objectId, nil
}