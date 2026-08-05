package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"fmt"
	"io"
	"log"
	"time"

	"capstone/app/core/utils"
	"capstone/app/repositories"
	"capstone/app/schemas"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type DocumentService struct {
	repo         *repositories.DocumentRepo
	audit        *repositories.AuditRepo
	storage      *repositories.StorageRepo
	notification *NotificationService
}

func NewDocumentService(repo *repositories.DocumentRepo, audit *repositories.AuditRepo, storage *repositories.StorageRepo, notification *NotificationService) *DocumentService {
	return &DocumentService{
		repo:         repo,
		audit:        audit,
		storage:      storage,
		notification: notification,
	}
}

func pageLimit(page, limit int) (int, int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 10 {
		limit = 10
	}
	offset := (page - 1) * limit
	return page, limit, offset
}

func return_Internal_error() (schemas.Response, error) {
	return schemas.Response{
		Status:  false,
		Message: "Internal Server Error",
	}, nil
}


func (d *DocumentService) Get_public_document_by_id(documentId uuid.UUID, role string) (schemas.Response,string, string, error) {
	storedHash, err := d.repo.Check_document_hash(documentId)
	if err != nil {
		return schemas.Response{},"","", err
	}

	objectId, err := d.repo.Get_object_id(documentId)
	if err != nil {
		return schemas.Response{},"","", err
	}

	filepath, err := d.repo.Get_document_filePath(documentId)
	if err != nil {
		return schemas.Response{},"","", err
	}

	objectHash, err := d.storage.GenerateObjectHMAC(objectId, filepath)
	if err != nil {
		return schemas.Response{},"","", err
	}

	if !hmac.Equal([]byte(storedHash), []byte(objectHash)) {
		return schemas.Response{
			Status:  false,
			Message: "Document integrity check failed. The file hash does not match the original document fingerprint",
		}, "","", nil
	}

	objectId, isPublic, err := d.repo.Get_document_by_id(documentId, uuid.Nil, role)
	if err != nil {
		return schemas.Response{}, "", "", err
	}

	url, contentType, err := d.storage.Get_document_presign(objectId, isPublic)
	if err != nil {
		log.Print("error getting presigned url: ", err)
		return schemas.Response{}, "", "", err
	}
	return schemas.Response{Status: true, Message: "Success", }, url, contentType, nil
}

func (d *DocumentService) Get_document_by_id(ctx context.Context, documentId, createdById uuid.UUID, role string) (schemas.Response,io.ReadCloser, *minio.ObjectInfo, error) {
	storedHash, err := d.repo.Check_document_hash(documentId)
	if err != nil {
		return schemas.Response{}, nil, nil, err
	}

	objectId, err := d.repo.Get_object_id(documentId)
	if err != nil {
		return schemas.Response{}, nil, nil, err
	}

	filepath, err := d.repo.Get_document_filePath(documentId)
	if err != nil {
		return schemas.Response{},	nil,nil, err
	}

	objectHash, err := d.storage.GenerateObjectHMAC(objectId, filepath)
	if err != nil {
		return schemas.Response{},nil,nil, err
	}

	if !hmac.Equal([]byte(storedHash), []byte(objectHash)) {
		return schemas.Response{
			Status:  false,
			Message: "Document integrity check failed. The file hash does not match the original document fingerprint",
		}, nil,nil, nil
	}
	
	_, _, err = d.repo.Get_document_by_id(documentId, createdById, role)
	if err != nil {
		return schemas.Response{}, nil, nil, err
	}

	filepath, err = d.repo.Get_document_filePath(documentId)
	if err != nil {
		return	schemas.Response{}, nil, nil, fmt.Errorf("error getting filepath for document %s: %w", documentId, err)
	}

	if _, err := d.audit.SaveAudit("open", "Opening File", createdById, documentId, "client", time.Now(), time.Now()); err != nil {
		log.Print("error saving open audit activity: ", err)
	}

	object, stat, err := d.storage.GetMinioObject(ctx, filepath)
	if err != nil {
		log.Print("error getting document stream: ", err)
		return schemas.Response{},nil, nil, err
	}
	return schemas.Response{
		Status: true,
		Message: "Success",
	},object, stat, nil
}

func (d *DocumentService) Get_document_by_status(status string, page, limit int) ([]schemas.DocumentResponse, int, int, error) {
	page, limit, offset := pageLimit(page, limit)
	docs, err := d.repo.Get_document_by_status(status, limit, offset)
	if err != nil {
		return []schemas.DocumentResponse{}, page, limit, err
	}
	if docs == nil {
		docs = []schemas.DocumentResponse{}
	}
	return docs, page, limit, nil
}

func (d *DocumentService) Get_documents_by_type(page, limit int) ([]schemas.DocumentResponse, int, int, error) {
	page, limit, offset := pageLimit(page, limit)
	docs, err := d.repo.Get_documents_by_type(limit, offset)
	if err != nil {
		return []schemas.DocumentResponse{}, 0, 0, err
	}
	return docs, page, limit, nil
}


func (d *DocumentService) Get_document_by_standard(standardId, createdById uuid.UUID, page, limit int, role string) ([]schemas.DocumentResponse, int, int, error) {
	page, limit, offset := pageLimit(page, limit)
	docs, err := d.repo.Get_document_by_standard(standardId, createdById, limit, offset, role)
	if err != nil {
		return []schemas.DocumentResponse{}, 0, 0, err
	}
	return docs, page, limit, nil
}

func (d *DocumentService) Get_document_by_service(serviceId, createdById uuid.UUID, page, limit int, role string) ([]schemas.DocumentResponse, int, int, error) {
	page, limit, offset := pageLimit(page, limit)
	docs, err := d.repo.Get_document_by_service(serviceId, createdById, limit, offset, role)
	if err != nil {
		return []schemas.DocumentResponse{}, 0, 0, err
	}
	return docs, page, limit, nil
}

func (d *DocumentService) Get_document_by_assessment(assessmentId, createdById uuid.UUID, page, limit int, role string) ([]schemas.DocumentResponse, int, int, error) {
	page, limit, offset := pageLimit(page, limit)
	docs, err := d.repo.Get_document_by_assessment(assessmentId, createdById, limit, offset, role)
	if err != nil {
		return []schemas.DocumentResponse{}, 0, 0, err
	}
	return docs, page, limit, nil
}

func (d *DocumentService) Get_document_by_createdBy(createdById uuid.UUID, page, limit int) ([]schemas.DocumentResponse, int, int, error) {
	page, limit, offset := pageLimit(page, limit)
	docs, err := d.repo.Get_document_by_createdBy(createdById, limit, offset)
	if err != nil {
		return []schemas.DocumentResponse{}, 0, 0, err
	}
	return docs, page, limit, nil
}

func (d *DocumentService) Create_document(req schemas.DocumentRequest, file io.Reader, fileName string, fileSize int64, contentType string, createdById uuid.UUID, role string) (schemas.Response, error) {
	documentTypeId, err := uuid.Parse(req.DocumentTypeId)
	if err != nil {
		log.Print("error parsing document type id: ", err)
		return return_Internal_error()
	}
	assessmentId, err := uuid.Parse(req.AssessmentId)
	if err != nil {
		log.Print("error parsing assessment id: ", err)
		return return_Internal_error()
	}
	serviceId, err := uuid.Parse(req.ServicesId)
	if err != nil {
		log.Print("error parsing service id: ", err)
		return return_Internal_error()
	}
	standardId, err := uuid.Parse(req.StandardId)
	if err != nil {
		log.Print("error parsing standard id: ", err)
		return return_Internal_error()
	}

	groupId, err := d.repo.Get_group_id_by_serviceId(req.ServicesId)
	if err != nil {
		log.Print("error getting group id: ", err)
		return return_Internal_error()
	}

	isSameName, err := d.repo.Check_document_name(req.FileName, groupId)
	if err != nil {
		log.Print("error checking same file name: ", err)
		return return_Internal_error()
	}
	if isSameName {
		return schemas.Response{
			Status:  false,
			Message: "filename already exist",
		}, nil
	}

	isPublic, err := d.repo.Is_public_document(documentTypeId)
	if err != nil {
		log.Print("error checking if document type is public: ", err)
		return return_Internal_error()
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Print("error reading uploaded file: ", err)
		return return_Internal_error()
	}
	fileHash := utils.GenerateHMAC(fileBytes)
	objectId := uuid.NewString()

	filepath, err := d.storage.Upload_document(bytes.NewReader(fileBytes), objectId, isPublic, fileSize, contentType)
	if err != nil {
		log.Print("error uploading document to object storage: ", err)
		return return_Internal_error()
	}

	documentIdStr, isCreated, isAuthorized, err := d.repo.Create_document(
		assessmentId,
		documentTypeId,
		createdById,
		standardId,
		serviceId,
		req.FileName,
		filepath,
		fileHash,
		objectId,
		role,
		isPublic,
	)
	if err != nil {
		d.storage.Delete_document(filepath)
		log.Print("error inserting document into database: ", err)
		return return_Internal_error()
	}

	if !isAuthorized {
		d.storage.Delete_document(filepath)
		return schemas.Response{
			Status:  false,
			Message: "Not Authorized, service/standard/assessment is not under the current group",
		}, nil
	}

	if !isCreated {
		d.storage.Delete_document(filepath)
		if _, auditErr := d.audit.SaveAudit("error", "error inserting data into database", createdById, uuid.Nil, "system", time.Now(), time.Now()); auditErr != nil {
			log.Print("error saving audit log: ", auditErr)
		}
		return return_Internal_error()
	}

	documentId, err := uuid.Parse(documentIdStr)
	if err != nil {
		log.Print("error parsing created document id: ", err)
		return return_Internal_error()
	}

	filename, err := d.audit.SaveAudit("insert", "Upload new document", createdById, documentId, "client", time.Now(), time.Now())
	if err != nil {
		log.Print("error saving create audit log: ", err)
	}

	adminEmail, adminId, err := d.repo.Get_admin_email()
	if err != nil {
		log.Print("error getting master admin email: ", err)
	}

	uploaderEmail, _, _ := d.repo.Get_document_owner_email(documentId)

	go d.TriggerIngestEvidence(bytes.NewReader(fileBytes), fileName, req, documentIdStr)

	go func() {
		d.notification.NotifyDeptHead(adminEmail, filename)
		log.Printf("[email] NotifyDeptHead done for document %s", filename)
	}()

	go d.notification.NotifySSE([]uuid.UUID{adminId}, SSEEvent{
		Type:       "new_document",
		DocumentId: documentId.String(),
		Status:     "Pending",
		Message:    fmt.Sprintf("A new document (%s by %s) requires your approval", req.FileName, uploaderEmail),
	})

	return schemas.Response{
		Status:  true,
		Message: "Successfully added document",
	}, nil
}

func (d *DocumentService) Update_document(req schemas.UpdateRequest, file io.Reader, fileSize int64, contentType string, createdById uuid.UUID, userRole string) (schemas.Response, error) {
	documentId, err := uuid.Parse(req.DocumentId)
	if err != nil {
		return schemas.Response{Status: false, Message: "Invalid document ID"}, err
	}

	status, err := d.repo.Document_is_approved(documentId)
	if err != nil {
		return schemas.Response{}, err
	}

	if status == "approved" || status == "pending" {
		return schemas.Response{
			Status:  false,
			Message: "Cannot Edit Approved/pending Document",
		}, nil
	}

	if userRole == "admin" {
		authorized, err := d.repo.Check_document_owner(documentId, createdById)
		if err != nil {
			return schemas.Response{Status: false, Message: "Authorization check failed"}, err
		}
		if !authorized {
			return schemas.Response{Status: false, Message: "Unauthorized: cannot edit this document"}, nil
		}
	}

	oldFilePath, err := d.repo.Get_document_filePath(documentId)
	if err != nil {
		return schemas.Response{}, err
	}

	objectId, err := d.repo.Get_object_id(documentId)
	if err != nil {
		return schemas.Response{}, err
	}

	var updatedHash, filePath string
	hasFile := file != nil
	hasFileName := req.FileName != ""
	newFileName := req.FileName

	switch {
	case hasFile && hasFileName:
		newFileName, updatedHash, filePath, err = d.storage.Update_document(file, fileSize, contentType, oldFilePath, req.FileName, objectId)
	case hasFile:
		newFileName, updatedHash, filePath, err = d.storage.Update_documentFile(file, fileSize, contentType, oldFilePath, objectId)
	}
	if err != nil {
		return schemas.Response{}, err
	}

	result, err := d.repo.Update_document(documentId, createdById, newFileName, filePath, updatedHash)
	if err != nil {
		log.Print("error updating document: ", err)
		return schemas.Response{Status: false, Message: "Error Updating document"}, nil
	}
	if result == 0 {
		return schemas.Response{Status: false, Message: "Document not Found"}, nil
	}
	if _, err := d.audit.SaveAudit("edit", fmt.Sprintf("Updating file %s", documentId), createdById, documentId, "client", time.Now(), time.Now()); err != nil {
		log.Print("error adding update log: ", err)
	}

	return schemas.Response{Status: true, Message: "Success updating document"}, nil
}

func (d *DocumentService) Delete_document(documentId, createdById uuid.UUID, userRole string) (schemas.Response, error) {
	filepath, isDeleted, err := d.repo.Delete_document(documentId, createdById, userRole)
	if err != nil {
		log.Print("error deleting document data: ", err)
		return schemas.Response{}, err
	}

	if !isDeleted {
		return schemas.Response{
			Status:  false,
			Message: "Document Not found",
		}, nil
	}

	if err := d.storage.Delete_document(filepath); err != nil {
		log.Print("error deleting document at the storage: ", err)
	}

	if _, err := d.audit.SaveAudit("delete", fmt.Sprintf("Deleting file %s", documentId), createdById, documentId, "client", time.Now(), time.Now()); err != nil {
		log.Print("error adding delete log: ", err)
	}

	return schemas.Response{
		Status:  true,
		Message: "Success deleting document",
	}, nil
}

func (d *DocumentService) Approval_document(documentId, createdById uuid.UUID, status string, file io.Reader, fileSize int64, contentType string) (schemas.Response, error) {
	storedHash, err := d.repo.Check_document_hash(documentId)
	if err != nil {
		return schemas.Response{}, err
	}

	objectId, err := d.repo.Get_object_id(documentId)
	if err != nil {
		return schemas.Response{}, err
	}

	filepath, err := d.repo.Get_document_filePath(documentId)
	if err != nil {
		return schemas.Response{}, err
	}

	objectHash, err := d.storage.GenerateObjectHMAC(objectId, filepath)
	if err != nil {
		return schemas.Response{}, err
	}

	if !hmac.Equal([]byte(storedHash), []byte(objectHash)) {
		return schemas.Response{
			Status:  false,
			Message: "Document integrity check failed. The file hash does not match the original document fingerprint",
		}, nil
	}

	storedStatus, err := d.repo.Document_is_approved(documentId)
	if err != nil {
		return schemas.Response{}, err
	}
	if storedStatus == "approved" {
		return schemas.Response{
			Status:  false,
			Message: "Cannot change status for approved document",
		}, nil
	}

	var rows int64
	if status == "approved" && file != nil {
		uploadedPath, err := d.storage.Upload_document(file, objectId, false, fileSize, contentType)
		if err != nil {
			return schemas.Response{}, err
		}
		rows, err = d.repo.Approval_document(documentId, status, uploadedPath, file)
		if err != nil {
			return schemas.Response{}, err
		}
	} else {
		rows, err = d.repo.Approval_document(documentId, status, "", file)
		if err != nil {
			return schemas.Response{}, err
		}
	}

	if rows == 0 {
		return schemas.Response{
			Status:  false,
			Message: "error updating document status or document not found",
		}, nil
	}

	filename, err := d.audit.SaveAudit("update", "Updating document status", createdById, documentId, "client", time.Now(), time.Now())
	if err != nil {
		log.Print("error saving approval audit log: ", err)
	}

	ownerEmail, ownerId, err := d.repo.Get_document_owner_email(documentId)
	if err != nil {
		log.Print("failed to get document owner email: ", err)
	} else {
		msg := fmt.Sprintf("Your Document (%s by %s) is now %s", filename, ownerEmail, status)

		go d.notification.NotifyOwner(ownerEmail, filename, msg)
		go d.notification.NotifySSE([]uuid.UUID{ownerId}, SSEEvent{
			Type:       "document_status_update",
			DocumentId: documentId.String(),
			Status:     status,
			Message:    msg,
		})
	}

	return schemas.Response{
		Status:  true,
		Message: "Success updating document status",
	}, nil
}

func (s *DocumentService) GetStats(createdById uuid.UUID, role string) (schemas.StatsResponse, error) {
	groups, stat, err := s.repo.Get_stats(createdById, role)
	if err != nil {
		return schemas.StatsResponse{}, fmt.Errorf("get document stats: %w", err)
	}

	total := stat.Approved + stat.Pending + stat.Rejected

	return schemas.StatsResponse{
		Groups: groups,
		Stats:  stat,
		Total:  total,
	}, nil
}
