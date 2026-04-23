package services

import (
	"capstone/app/repositories"
	"capstone/app/schemas"
	"mime/multipart"
	"log"
	"time"

	"github.com/google/uuid"
)
type DocumentService struct {
	repo *repositories.DocumentRepo
	audit *repositories.AuditRepo
	storage *repositories.StorageRepo
}

func NewDocumentService(repo *repositories.DocumentRepo, audit *repositories.AuditRepo, storage *repositories.StorageRepo) *DocumentService{
	return &DocumentService{
		repo : repo,
		audit : audit,
		storage : storage,
	}
}

func pageLimit(page, limit int)(int, int, int){
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100{
		limit = 20
	}
	offset := (page - 1 ) *  limit

	return page, limit, offset
}

func (s *DocumentService) Get_Document(page,limit int)([]schemas.DocumentResponse,error){
	page, limit, offset := pageLimit(page, limit)
	
	docs, err := s.repo.GetDocuments(limit, offset)
	if err != nil{
		return nil, err
	}
	return docs, nil
}

func (s *DocumentService) Get_document_by_id(id uuid.UUID)(schemas.DocumentResponse, error){
	docs, err := s.repo.Get_document_by_id(id)
	if err != nil{
		return schemas.DocumentResponse{}, err
	}
	return docs, nil
}

func (s *DocumentService) Get_document_by_status(status string, page,limit int)([]schemas.DocumentResponse, int, int, error){
	page,limit,offset:= pageLimit(page, limit)
	docs, err := s.repo.Get_document_by_status(status, limit, offset)
	if err != nil {
		return []schemas.DocumentResponse{}, 0, 0, nil
	}

	return docs, page, limit, nil
}

func (s *DocumentService) Get_document_by_type(typesId uuid.UUID, page, limit int)([]schemas.DocumentResponse, int, int, error){
	page, limit, offset := pageLimit(page, limit)
	docs, err := s.repo.Get_document_by_type(typesId, page, offset)
	if err != nil {
		return []schemas.DocumentResponse{}, 0, 0, nil
	}
	return docs, page,limit, nil
}

func (s *DocumentService) Get_document_by_group(groupId uuid.UUID, page, limit int)([]schemas.DocumentResponse, int, int, error){
	page, limit, offset := pageLimit(page,limit)
	docs, err := s.repo.Get_document_by_group(groupId, limit, offset)
	if err != nil {
		return 	[]schemas.DocumentResponse{}, 0,0, nil
	}
	return docs, page, limit, nil
}

func(s *DocumentService) Get_document_by_standard(standardId uuid.UUID, page, limit int)([]schemas.DocumentResponse, int, int, error){
	page, limit, offset := pageLimit(page,limit)
	docs, err := s.repo.Get_document_by_standard(standardId, limit,offset)
	if err != nil {
		return []schemas.DocumentResponse{}, 0,0, err
	}
	return docs,page,limit, nil
}

func (s *DocumentService) Get_document_by_service(serviceId uuid.UUID, page,limit int)([]schemas.DocumentResponse, int, int, error){
	page, limit, offset := pageLimit(page,limit)
	docs, err := s.repo.Get_document_by_service(serviceId, limit, offset)
	if err != nil {
		return []schemas.DocumentResponse{}, 0, 0, err
	}
	return docs, 0,0,nil
}

func (s *DocumentService) Get_document_by_assessment(assessmentId uuid.UUID, page, limit int)([]schemas.DocumentResponse, int,int,error){
	page, limit, offset := pageLimit(page, limit)
	docs, err := s.repo.Get_document_by_assessment(assessmentId, limit, offset)
	if err != nil{
		return []schemas.DocumentResponse{}, 0,0, err
	}
	return docs, 0,0, nil
}

func(s *DocumentService) Get_document_by_createdBy(createdById uuid.UUID, page, limit int)([]schemas.DocumentResponse, int, int, error){
	page, limit, offset := pageLimit(page,limit)
	docs, err := s.repo.Get_document_by_createdBy(createdById, limit, offset)
	if err != nil {
		return []schemas.DocumentResponse{}, 0, 0 , err
	}
	return docs, 0,0, nil
}

func (s *DocumentService) Create_document(req schemas.DocumentRequest, file multipart.File, header *multipart.FileHeader, createdById uuid.UUID)(schemas.Response,error){
	filename, filepath, err := s.storage.Upload_document(file, header, req.FileName)
	if err != nil{
		return schemas.Response{}, err
	}

	assessmentId, _ := uuid.Parse(req.AssessmentId)
	documentTypeId,_ := uuid.Parse(req.DocumentTypeId)
	groupId, _ := uuid.Parse(req.GroupId)
	serviceId, _ := uuid.Parse(req.ServicesId)
	standardId, _ := uuid.Parse(req.StandardId)

	document_Id, err := s.repo.Create_document(
		assessmentId, 
		documentTypeId,
		createdById,
		groupId,
		standardId,
		serviceId,
		filename,
		filepath,
	)
	if err != nil {
		s.storage.Delete_document(filepath)
		return schemas.Response{},err	
	}

	documentId, _:= uuid.Parse(document_Id)

	err = s.audit.SaveAudit("activity","insert",createdById,documentId,"client",time.Now(),time.Now())
	if err != nil{
		log.Print("Error adding create log: ",err)
	}

	return schemas.Response{
		Status: true,
		Message: "Successfull adding document",
	}, nil
}

func(s *DocumentService) Update_document(req schemas.UpdateRequest, file multipart.File, header *multipart.FileHeader,userId uuid.UUID)(schemas.Response, error){
	filename, filepath, err := s.storage.Upload_document(file, header, req.FileName)
	if err != nil {
		return schemas.Response{}, err
	}

	documentId, _ := uuid.Parse(req.DocumentId)
	assessmentId, _ := uuid.Parse(req.AssessmentId)
	documentTypeId,_ := uuid.Parse(req.DocumentTypeId)
	groupId, _ := uuid.Parse(req.GroupId)
	serviceId, _ := uuid.Parse(req.ServicesId)
	standardId, _ := uuid.Parse(req.StandardId)

	rows, err := s.repo.Update_document(
		documentId,
		assessmentId,
		documentTypeId,
		userId,
		groupId,
		standardId,
		serviceId,
		filename,
		filepath,
	)

	if err != nil {
		s.storage.Delete_document(filepath)
		log.Print("alamak ", err)
		return schemas.Response{}, err
	}

	if rows == 0 {
		return schemas.Response{
			Status: false,
			Message: "Document not found",
		}, nil
	}

	err = s.audit.SaveAudit("activity", "update", userId, documentId,"client",time.Now(),time.Now())
	if err != nil {
		log.Print("Error adding update log: ", err)
	}
	return schemas.Response{
		Status: true,
		Message: "sucess update document",
	},  nil
}



func (s *DocumentService) Delete_document(documentId uuid.UUID, userId uuid.UUID)(schemas.Response, error){
	rows, err := s.repo.Delete_document(documentId, userId)
	if err != nil{
		log.Print("error",err)
		return schemas.Response{},err
	}

	if rows == 0 {
		return schemas.Response{
			Status: false,
			Message: "Document Not Found",
		}, nil
	}

	err = s.audit.SaveAudit("activity","delete",userId,documentId,"client",time.Now(),time.Now())
	if err != nil{
		log.Print("error adding delete log: ", err)
	}

	return schemas.Response{
		Status: true,
		Message: "Sucess deleteing  document",
	}, nil
}

func (s *DocumentService) Approval_document(documentId, userId uuid.UUID,status string)(schemas.Response, error){
	rows, err := s.repo.Approval_document(documentId,userId,status)
	if err != nil{
		return schemas.Response{}, err
	}

	if rows == 0 {
		return schemas.Response{
			Status: false,
			Message: "error updating document status or document not found",
		}, nil
	}
	err = s.audit.SaveAudit("activity","delete", userId,documentId,"client",time.Now(),time.Now())
	if err != nil{
		log.Print("error adding approval log: ", err)
	}
	return schemas.Response{
		Status: true,
		Message: "Success updating document status",
	}, nil
}