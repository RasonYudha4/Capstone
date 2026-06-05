package services

import (
	"capstone/app/repositories"
	"capstone/app/schemas"

	"fmt"
	"log"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
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
func (d *DocumentService) Get_Document(page,limit int)([]schemas.DocumentResponse,error){
	page, limit, offset := pageLimit(page, limit)
	
	docs, err := d.repo.GetDocuments(limit, offset)
	if err != nil{
		return nil, err
	}
	return docs, nil
}

func (d *DocumentService) Get_document_by_id(documentId, createdById uuid.UUID, role string)(string,schemas.UploadResponse, error){
	filePath, status ,err := d.repo.Get_document_by_id(documentId,createdById, role)
	if err != nil {
    	return "", schemas.UploadResponse{Status: false, Message: "Error fetching document"}, err
	}
	if filePath == "" {
    	return "", schemas.UploadResponse{Status: false, Message: "Document not found"}, nil
	}

	_, err = d.audit.SaveAudit("open","Opening File ",createdById,documentId,"client",time.Now(),time.Now())
	if err != nil {
		log.Print("Error saving open audit activty")
	}
	
	return filePath,status, nil
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

func (d *DocumentService) Create_document(req schemas.DocumentRequest, file multipart.File, header *multipart.FileHeader, createdById uuid.UUID,role string)(schemas.Response,error){
	filename, filepath, err := d.storage.Upload_document(file, header, req.FileName)
	if err != nil{
		log.Print("error upload document to storage", err)
		return schemas.Response{}, err
	}

	assessmentId, _ := uuid.Parse(req.AssessmentId)
	documentTypeId,_ := uuid.Parse(req.DocumentTypeId)
	groupId, _ := uuid.Parse(req.GroupId)
	serviceId, _ := uuid.Parse(req.ServicesId)
	standardId, _ := uuid.Parse(req.StandardId)
	document_Id,isCreated, isAuthorized,err := d.repo.Create_document(
		assessmentId, 
		documentTypeId,
		createdById,
		groupId,
		standardId,
		serviceId,
		filename,
		filepath,
		role,
	)
	
	documentId, _:= uuid.Parse(document_Id)
	
	if err != nil {
		d.storage.Delete_document(filepath)
		log.Print("error delete document at storage", err)
		return schemas.Response{},err	
	}
	
	if !isAuthorized {
		return schemas.Response{
			Status: false,
			Message: "Not Authorized, service/standard/assessment is not under the current group",
		}, nil
	}
	
	if !isCreated {	
		d.storage.Delete_document(filepath)
		_, err = d.audit.SaveAudit("error","error inserting data into database",createdById,documentId,"system",time.Now(),time.Now())

		return schemas.Response{
			Status: false,
			Message: "Error insert data",
		}, nil
	
	}
	
	filename, err  = d.audit.SaveAudit("insert","Upload new document",createdById,documentId,"client",time.Now(),time.Now())
	if err != nil{
		log.Print("Error adding create log: ",err)
	}

	adminEmail, adminId, err := d.repo.Get_admin_email()
	if err != nil {
		log.Print("Error getting master admin email")
	}

	go func(){
		 log.Printf("[email] NotifyDeptHead done for document %s", documentId)
		d.notification.NotifyDeptHead(adminEmail, documentId.String())
	}()

	go d.notification.NotifySSE(adminId, SSEEvent{
		Type:		"new_document",
		DocumentId:	documentId.String(),
		Status:		"Pending",
		Message:	fmt.Sprintf("A new document (%s) requires your approval", filename),	
	})

	return schemas.Response{
		Status: true,
		Message: "Successfull adding document",
	}, nil
}

func(d *DocumentService) Update_document(req schemas.UpdateRequest, file multipart.File, header *multipart.FileHeader,createdById uuid.UUID, userRole string)(schemas.Response, error){
	documentId, err := uuid.Parse(req.DocumentId)
	if err != nil {
		return schemas.Response{Status: false, Message: "Invalid document ID"}, err
	}

	if userRole != "master_admin" {
		authorized, err := d.repo.Check_document_owner(documentId, createdById)
		if err != nil {
			return schemas.Response{Status: false, Message: "Authorization check failed"}, err
		}
		if !authorized {
			return schemas.Response{Status: false, Message: "Unauthorized: cannot edit this document"}, nil
		}
	}

	oldFileName, err := d.repo.Get_document_fileName(documentId, createdById)
	if err != nil{
		return schemas.Response{}, err
	}

	hasfile := file != nil
	hasfileName := req.FileName != ""

	var newFileName, filePath string
	if hasfile && hasfileName{
		newFileName, filePath, err = d.storage.Update_document(file,header,req.FileName)
		if err != nil{
			return schemas.Response{}, err
		}
	}else if hasfile{
		newFileName, filePath, err = d.storage.Update_documentFile(file,header,oldFileName)
		if err != nil {
			return schemas.Response{}, err
		}
	}else if hasfileName{
		newFileName, filePath, err = d.storage.Update_documentName(oldFileName, req.FileName)
		if err != nil{
			return schemas.Response{}, err 
		}
	}else {
		newFileName = ""
		filePath = ""
	}

 	result, err := d.repo.Update_document(documentId,createdById,newFileName,filePath, req.Description)
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

func (d *DocumentService) Delete_document(documentId uuid.UUID, userId uuid.UUID, userRole string)(schemas.Response, error){
	filename, isDeleted ,err := d.repo.Delete_document(documentId, userId, userRole)
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

	err = d.storage.Delete_document(filename) 
	if err != nil{
		log.Print("Error deleting document at the storage", err)
	}

	_, err = d.audit.SaveAudit("delete",fmt.Sprintf("Deleting file %s",documentId),userId,documentId,"client",time.Now(),time.Now())
	if err != nil{
		log.Print("error adding delete log: ", err)
	}

	return schemas.Response{
		Status: true,
		Message: "Sucess deleting document",
	}, nil
}

func (d *DocumentService) Approval_document(documentId, userId uuid.UUID, status string, file multipart.File, fileHeader *multipart.FileHeader) (schemas.Response, error) {

	var rows int64
	var err error
	if status == "approved" && file != nil && fileHeader != nil {
    
		var signedDocPath string
    	_, signedDocPath, err = d.storage.Upload_document(file, fileHeader, fileHeader.Filename)
    
		if err != nil {
        	return schemas.Response{}, err
    	}
    		rows, err = d.repo.Approval_document(documentId, status, signedDocPath, file)
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

	filename, err := d.audit.SaveAudit("update","Updating document status",userId, documentId, "client", time.Now(), time.Now())
	if err != nil {
		log.Print("error adding approval log: ", err)
	}

	ownerEmail, ownerId, err := d.repo.Get_document_owner_email(documentId)
	if err != nil {
		log.Print("failed to get document owner email:", err)
	} else {
		msg := fmt.Sprintf("Your Document (%s) is now %s", filename, status)

		go d.notification.NotifyOwner(ownerEmail, documentId, msg)

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

func (s *DocumentService) GetStats(userId uuid.UUID, role string) (schemas.StatsResponse, error) {
	groups, stat, err := s.repo.Get_stats(userId, role)
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

