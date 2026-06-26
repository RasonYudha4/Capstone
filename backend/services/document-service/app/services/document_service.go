package services

import (
	"bytes"
	"capstone/app/core/utils"
	"capstone/app/repositories"
	"capstone/app/schemas"
	"context"
	"crypto/hmac"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)
type DocumentService struct {
	repo *repositories.DocumentRepo
	audit *repositories.AuditRepo
	storage *repositories.StorageRepo
	notification *NotificationService
}

func NewDocumentService(repo *repositories.DocumentRepo, audit *repositories.AuditRepo, storage *repositories.StorageRepo, notification *NotificationService) *DocumentService{
	return &DocumentService{
		repo : repo,
		audit : audit,
		storage : storage,
		notification: notification,
	}
}

func pageLimit(page, limit int)(int, int, int){
	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 10 {
		limit = 10
	}

	offset := (page - 1) * limit

	return page, limit, offset
}

func return_Internal_error()(schemas.Response, error){
	return schemas.Response{
		Status: false,
		Message: "Internal Server Error",
	}, nil
}

func (d *DocumentService) Get_Document(page,limit int)([]schemas.DocumentResponse,error){
	page, limit, offset := pageLimit(page, limit)
	
	docs, err := d.repo.GetDocuments(limit, offset)
	if err != nil{
		return nil, err
	}
	return docs, nil
}

func (d *DocumentService) Get_public_document_by_id(documentId uuid.UUID, role string)(string, error){
	filepath,_, err := d.repo.Get_document_by_id(documentId,uuid.Nil, role)
	if err != nil{
		return "", err
	}
	return filepath, nil
}

func (d *DocumentService) Get_document_by_id(ctx context.Context, documentId, createdById uuid.UUID, role string)(*minio.Object, *minio.ObjectInfo, error){
	filePath, _ ,err := d.repo.Get_document_by_id(documentId,createdById, role)
	if err != nil {
    	return nil,nil, err
	}
	if filePath == "" {
    	return nil, nil, err
	}

	log.Printf("StreamDocument — filePath: '%s'", filePath)
	_, err = d.audit.SaveAudit("open","Opening File ",createdById,documentId,"client",time.Now(),time.Now())
	if err != nil {
		log.Print("Error saving open audit activty")
	}
	
	object, stat, err := d.storage.GetMinioObject(ctx, filePath)
    if err != nil {
        return nil, nil, err
    }
	return object,stat, nil
}

func (d *DocumentService) Get_document_by_status(status string, page,limit int)([]schemas.DocumentResponse, int, int, error){
	page,limit,offset:= pageLimit(page, limit)
	docs, err := d.repo.Get_document_by_status(status, limit, offset)
	if err != nil {
		return []schemas.DocumentResponse{}, 0, 0, err
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

func (d *DocumentService) Get_document_by_group(groupId uuid.UUID, page, limit int)([]schemas.DocumentResponse, int, int, error){
	page, limit, offset := pageLimit(page,limit)
	docs, err := d.repo.Get_document_by_group(groupId, limit, offset)
	if err != nil {
		return 	[]schemas.DocumentResponse{}, 0,0, err
	}
	return docs, page, limit, nil
}

func(d *DocumentService) Get_document_by_standard(standardId,createdById uuid.UUID, page, limit int, role string)([]schemas.DocumentResponse, int, int, error){
	page, limit, offset := pageLimit(page,limit)
	docs, err := d.repo.Get_document_by_standard(standardId,createdById ,limit,offset, role)
	if err != nil {
		return []schemas.DocumentResponse{}, 0,0, err
	}
	return docs,page,limit, nil
}

func (d *DocumentService) Get_document_by_service(serviceId,createdById uuid.UUID, page,limit int, role string)([]schemas.DocumentResponse, int, int, error){
	page, limit, offset := pageLimit(page,limit)
	docs, err := d.repo.Get_document_by_service(serviceId,createdById,limit, offset, role)
	if err != nil {
		return []schemas.DocumentResponse{}, 0, 0, err
	}
	return docs, page,limit,nil
}

func (d *DocumentService) Get_document_by_assessment(assessmentId,createdById uuid.UUID, page, limit int, role string)([]schemas.DocumentResponse, int,int,error){
	page, limit, offset := pageLimit(page, limit)
	docs, err := d.repo.Get_document_by_assessment(assessmentId,createdById ,limit, offset, role )
	if err != nil{
		return []schemas.DocumentResponse{}, 0,0, err
	}
	return docs, page,limit, nil
}

func(d *DocumentService) Get_document_by_createdBy(createdById uuid.UUID, page, limit int)([]schemas.DocumentResponse, int, int, error){
	page, limit, offset := pageLimit(page,limit)
	docs, err := d.repo.Get_document_by_createdBy(createdById, limit, offset)
	if err != nil {
		return []schemas.DocumentResponse{}, 0, 0 , err
	}
	return docs, page,limit, nil
}

func (d *DocumentService) Create_document(req schemas.DocumentRequest, file multipart.File, header *multipart.FileHeader, createdById uuid.UUID, role string) (schemas.Response, error) {
	documentTypeId, err := uuid.Parse(req.DocumentTypeId)
	if err != nil{
		log.Print("Error Parsing document type id: ", err)
		return_Internal_error()
	}

	groupId, err := d.repo.Get_group_id_by_serviceId(req.ServicesId)
	if err != nil{
		log.Print("error getting group id: ", err)
		return_Internal_error()
	}

	isSameName, err := d.repo.Check_document_name(req.FileName, groupId)
	if err != nil {
		log.Print("error checking same file name: ", err)
		return_Internal_error()
	}

	if isSameName{
		return schemas.Response{
			Status : false,
			Message: "filename already exist",
		}, nil
	}

	isPublic, err := d.repo.Is_public_document(documentTypeId)
	if err != nil{
		log.Print("error document is public check ", err)
		return_Internal_error()
	}

	fileBytes, _ := io.ReadAll(file)
	fileHash := utils.GenerateHMAC(fileBytes)
	
	var filepath string 
	objectId := uuid.NewString()
	fileSize := header.Size
	contentType := header.Header.Get("Content-Type")
	
	if isPublic {
		filepath, err = d.storage.Upload_document(bytes.NewReader(fileBytes), objectId, isPublic, fileSize, contentType)
	}else{
		filepath, err = d.storage.Upload_document(bytes.NewReader(fileBytes), objectId, isPublic, fileSize, contentType)
	}

	if err != nil {
		log.Print("error upload document do object storage: ", err)
		return_Internal_error()
	}
	
	assessmentId, _ := uuid.Parse(req.AssessmentId)
	serviceId, _ := uuid.Parse(req.ServicesId)
	standardId, _ := uuid.Parse(req.StandardId)
	document_Id, isCreated, isAuthorized, err := d.repo.Create_document(
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
	documentId, _ := uuid.Parse(document_Id)

	if err != nil {
		d.storage.Delete_document(filepath)
		log.Print("error inerting data into database: ", err)
		return_Internal_error()
	}

	if !isAuthorized {
		return schemas.Response{
			Status:  false,
			Message: "Not Authorized, service/standard/assessment is not under the current group",
		}, nil
	}

	if !isCreated {
		d.storage.Delete_document(filepath)
		_, err = d.audit.SaveAudit("error", "error inserting data into database", createdById, documentId, "system", time.Now(), time.Now())
		log.Print("Error inserting data, data is not inserted: ", err)
		return_Internal_error()
	}
		
	filename, err := d.audit.SaveAudit("insert", "Upload new document", createdById, documentId, "client", time.Now(), time.Now())
	if err != nil {
		log.Print("Error adding create log: ", err)
	}
	
	adminEmail, adminId, err := d.repo.Get_admin_email()
	if err != nil {
		log.Print("Error getting master admin email", err)
	}

	go d.TriggerIngestEvidence(bytes.NewReader(fileBytes), header.Filename, req, document_Id)

	go func() {
		d.notification.NotifyDeptHead(adminEmail, filename)
		log.Printf("[email] NotifyDeptHead done for document %s", filename)
	}()

	go d.notification.NotifySSE([]uuid.UUID{adminId}, SSEEvent{
		Type:       "new_document",
		DocumentId: documentId.String(),
		Status:     "Pending",
		Message:    fmt.Sprintf("A new document (%s) requires your approval", req.FileName),
	})

	return schemas.Response{
		Status:  true,
		Message: "Successfull adding document",
	}, nil
}

func(d *DocumentService) Update_document(req schemas.UpdateRequest, file multipart.File, header *multipart.FileHeader,createdById uuid.UUID, userRole string)(schemas.Response, error){
	documentId, err := uuid.Parse(req.DocumentId)
	if err != nil {
		return schemas.Response{Status: false, Message: "Invalid document ID"}, err
	}

	isApproved, err := d.repo.Document_is_approved(documentId)
	if err != nil {
		log.Print("error", err)
		return schemas.Response{}, err
	}

	if isApproved {
		return schemas.Response{
			Status: false,
			Message: "Cannot Edit Approved Document",
		}, nil
	}
	
	if userRole != "master-admin" {
		authorized, err := d.repo.Check_document_owner(documentId, createdById)
		if err != nil {
			return schemas.Response{Status: false, Message: "Authorization check failed"}, err
		}
		if !authorized {
			return schemas.Response{Status: false, Message: "Unauthorized: cannot edit this document"}, nil
		}
	}

	oldFilePath, err := d.repo.Get_document_filePath(documentId)
	if err != nil{
		return schemas.Response{}, err
	}

	hasfile := file != nil
	hasfileName := req.FileName != ""
	
	objectId, _  := d.repo.Get_object_id(documentId)

	var newFileName,updatedHash, filePath string
	if hasfile && hasfileName{	
		newFileName, updatedHash,filePath, err = d.storage.Update_document(file,header,oldFilePath, req.FileName, objectId)
		if err != nil{
			return schemas.Response{}, err
		}
	}else if hasfile{
		newFileName,updatedHash, filePath, err = d.storage.Update_documentFile(file,header,oldFilePath, objectId)
		if err != nil {
			return schemas.Response{}, err
		}
	}else if hasfileName{
		newFileName, filePath, err = d.storage.Update_documentName(oldFilePath, req.FileName)
		if err != nil{
			return schemas.Response{}, err 
		}
	}else {
		newFileName = ""
		filePath = ""
	}
	log.Print(updatedHash, "update")

 	result, err := d.repo.Update_document(documentId,createdById,newFileName,filePath, req.Description, updatedHash)
	if err != nil{
		log.Print("Error updating file :", err)
		return schemas.Response{Status: false, Message: "Error Updating document"}, nil
	}

	if result == 0 {
		log.Print("Now row affected for document update :", err )
		return schemas.Response{Status: false, Message: "Document not Found"}, nil
	}

	_, err = d.audit.SaveAudit("edit",fmt.Sprintf("Updating file %s", documentId), createdById, documentId,"client",time.Now(),time.Now())
	if err != nil {
		log.Print("Error adding update log: ", err)
	}
	return schemas.Response{Status: true, Message: "sucess update document"},  nil
}


func (d *DocumentService) Delete_document(documentId uuid.UUID, createdById uuid.UUID, userRole string)(schemas.Response, error){
	filepath, isDeleted ,err := d.repo.Delete_document(documentId, createdById, userRole)
	if err != nil{
		log.Print("error delete document data",err)
		return schemas.Response{},err
	}
	
	log.Print(isDeleted)
	if !isDeleted{
		return schemas.Response{
			Status: false,
			Message: "Document Not found",
		}, nil
	}

	err = d.storage.Delete_document(filepath) 
	log.Print("error gak nih ", err)
	if err != nil{
		log.Print("Error deleting document at the storage", err)
	}

	_, err = d.audit.SaveAudit("delete",fmt.Sprintf("Deleting file %s",documentId),createdById,documentId,"client",time.Now(),time.Now())
	if err != nil{
		log.Print("error adding delete log: ", err)
	}

	return schemas.Response{
		Status: true,
		Message: "Sucess deleting document",
	}, nil
}

func (d *DocumentService) Approval_document(documentId, createdById uuid.UUID, status string, file multipart.File, fileHeader *multipart.FileHeader) (schemas.Response, error) {

	storedHash, err := d.repo.Check_document_hash(documentId)
	if err != nil{
		return schemas.Response{}, err
	}

	objectId, _  := d.repo.Get_object_id(documentId)
	filepath, err := d.repo.Get_document_filePath(documentId)
	objectHash, _ := d.storage.GenerateObjectHMAC(objectId, filepath)

	if !hmac.Equal(
		[]byte(storedHash),
    	[]byte(objectHash),
	){
		return schemas.Response{
			Status: false,
			Message: "Document integrity check failed. The file hash does not match the original document fingerprint",
		}, err
	}

	isApproved, err := d.repo.Document_is_approved(documentId)
	if err != nil{
		return schemas.Response{}, err
	}

	if isApproved{
		return schemas.Response{
			Status: false,
			Message: "Cannot change status for approved document",
		}, nil
	}

	var rows int64
	if status == "approved" && file != nil && fileHeader != nil {
		fileSize := fileHeader.Size
		contentType := fileHeader.Header.Get("Content-Type") 

    	filepath, _ := d.storage.Upload_document(file,objectId, false, fileSize, contentType)

    	rows, err = d.repo.Approval_document(documentId, status, filepath, file)
	}else {
		
		rows, err = d.repo.Approval_document(documentId, status, "",file)
	}
	
	if err != nil {
		return schemas.Response{}, err
	}

	if rows == 0 {	
		return schemas.Response{
			Status:  false,
			Message: "error updating document status or document not found",
		}, nil
	}

	filename, err := d.audit.SaveAudit("update","Updating document status",createdById, documentId, "client", time.Now(), time.Now())
	if err != nil {
		log.Print("error adding approval log: ", err)
	}
	ownerEmail, ownerId, err := d.repo.Get_document_owner_email(documentId)
	if err != nil {
		log.Print("failed to get document owner email:", err)
	} else {
		msg := fmt.Sprintf("Your Document (%s) is now %s", filename, status)

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
		return schemas.StatsResponse{Status: false}, err
	}
 
	total := stat.Approved + stat.Pending + stat.Rejected
 
	return schemas.StatsResponse{
		Status: true,
		Groups: groups,
		Stats:  stat,
		Total:  total,
	}, nil
}

