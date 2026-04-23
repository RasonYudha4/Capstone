package api

import (
	"capstone/app/schemas"
	"capstone/app/services"
	"errors"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type DocumentHandler struct{
	service *services.DocumentService
}

func NewDocumentHandler(service *services.DocumentService) *DocumentHandler{
	return &DocumentHandler{
		service : service,
	}
}


func (d *DocumentHandler)Get_all_documents_Handler(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    
    docs,err := d.service.Get_Document(page, limit)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{
        "data":   docs,
        "page":   page,
        "limit":  limit,
    })
}

func (d *DocumentHandler) Get_document_by_id_handler(r *gin.Context){
	document_id := r.Param("id")
	id, err := uuid.Parse(document_id)
	if err != nil{
		log.Print("error to parse id")
	}

	docs, err := d.service.Get_document_by_id(id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows){
			r.JSON(404,"Document Not Found")
			return
		}
		r.JSON(500,"Internal error")
	}

	r.JSON(200,docs)
}

func(d *DocumentHandler) Get_document_by_status_handler(r *gin.Context){
	statusParam := r.Param("status")
	page, _ := strconv.Atoi(r.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(r.DefaultQuery("limit", "20"))
	
	status,page, limit, err := d.service.Get_document_by_status(statusParam,page,limit)
	if err != nil {
		r.JSON(500,"internal server error")
		return
	}
	 r.JSON(200, gin.H{
        "data":   status,
        "page":   page,
        "limit":  limit,
    })
}

func (d *DocumentHandler)Get_document_by_type_handler(r *gin.Context){
	types := r.Param("type")
	typeId, err := uuid.Parse(types)
	if err != nil{
		log.Print("Error parse type into uuid; ", err)
		return
	}
	page, _ := strconv.Atoi(r.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(r.DefaultQuery("limit", "20"))
	
	Type,page, limit, err := d.service.Get_document_by_type(typeId,page,limit)
	if err != nil{
		log.Print("Error terus wok", err)
		r.JSON(500,"Internal Server Error")
		return
	}
	r.JSON(200, gin.H{
		"data": Type,
		"page": page,
		"limit": limit,
	})
}

func(d *DocumentHandler) Get_document_by_group_handler(r *gin.Context){
	groupIdParam := r.Param("group")
	groupId, err := uuid.Parse(groupIdParam)
	if err != nil{
		log.Print("Error parse groupid uuid: ", err)
		return
	}
	page, _ := strconv.Atoi(r.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(r.DefaultQuery("limit", "20"))
	
	group,page,limit, err:= d.service.Get_document_by_group(groupId,page,limit)
	if err != nil{
		r.JSON(500,"Internal Server Error")
		return
	}
	r.JSON(200, gin.H{
		"data": group,
		"page": page,
		"limit": limit,
	})
}

func(d *DocumentHandler) Get_document_by_standard_handler(r *gin.Context){
	standardParam := r.Param("standard")
	standardId, err := uuid.Parse(standardParam)
	if err != nil {
		log.Print("error Parse standard uui: ",err)
		return
	}
	page, _ := strconv.Atoi(r.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(r.DefaultQuery("limit", "20"))
	

	standard, page, limit, err := d.service.Get_document_by_standard(standardId,page, limit)
	if err != nil{
		r.JSON(500,"internal server error")
		return 
	}
	r.JSON(200, gin.H{
		"data": standard,
		"page": page,
		"limit": limit,
	})
}

func(d *DocumentHandler) Get_document_by_service_handler(r *gin.Context){
	serviceParam := r.Param("service")
	serviceId, err := uuid.Parse(serviceParam)
	if err != nil {
		log.Print("error parsing service uuid: ", err)
		return
	}
	page, _ := strconv.Atoi(r.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(r.DefaultQuery("limit", "20"))

	service,page,limit, err := d.service.Get_document_by_service(serviceId,page,limit)
	if err != nil {
		r.JSON(500, "internal server error")
		return
	}
	r.JSON(200, gin.H{
		"data": service,
		"page": page,
		"limit": limit,
	})
}

func(d *DocumentHandler) Get_document_by_assessment_handler(r *gin.Context){
	assessmentParam := r.Param("assessment")
	assessmentId, err := uuid.Parse(assessmentParam)
	if err !=nil {	
		log.Print("error parse assessment uuid: ", err)
		return
	}

	page, _ := strconv.Atoi(r.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(r.DefaultQuery("limit", "20"))
	assessment,page,limit,err := d.service.Get_document_by_assessment(assessmentId,page,limit)
	if err != nil {
		r.JSON(500, "internal server error")
		return
	}
	r.JSON(200, gin.H{
		"data": assessment, 
		"page": page,
		"limit": limit,
	})
}

func(d *DocumentHandler) Get_document_by_createdBy_handler(r *gin.Context){
	createdByParam := r.Param("createdBy")
	createdById, err := uuid.Parse(createdByParam)
	if err != nil {
		log.Print("error parsing createdBy uuid: ", err)
		return
	}
	page, _ := strconv.Atoi(r.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(r.DefaultQuery("limit", "20"))
    
	createdBy,page, limit, err := d.service.Get_document_by_createdBy(createdById,page,limit)
	if err != nil {
		r.JSON(500, "internal server error")
		return
	}

	r.JSON(200, gin.H{
		"data" : createdBy,
		"page": page,
		"limit": limit,
	})
}

func(d *DocumentHandler) Create_document_handler(r *gin.Context){
	var req schemas.DocumentRequest
	if err := r.ShouldBind(&req); err != nil{
		log.Print("error parse json: ", err)
		r.JSON(400, gin.H{"error": err.Error()})
		return
	}
	fileHeader, err := r.FormFile("uploadedFile")
	if err != nil {
		log.Print("Error getting file: ", err)
		r.JSON(400, gin.H{"Error": err.Error()})
		return
	}

	file, _ := fileHeader.Open()
	defer file.Close()

	userId,err := uuid.Parse(r.GetString("user_id"))
	if err != nil{
		r.JSON(400, gin.H{"errir": err.Error()})
		return
	}

	result, err := d.service.Create_document(req,file,fileHeader,userId)
	if err != nil{
		r.JSON(500,"Internal server error")
		return
	}

	if !result.Status{
		r.JSON(500, "internal server error")
	}

	r.JSON(200,result)

}
 
func(d *DocumentHandler) Update_document_handler(r *gin.Context){
	var req schemas.UpdateRequest
	if err := r.ShouldBind(&req); err != nil{
		log.Print("Eror wok", err)
		r.JSON(400, gin.H{"error: ": err.Error()})
		return
	}

	fileHeader, err := r.FormFile("uploadedFile")
	if err != nil {
		log.Print("aduhai error: ",err)
		r.JSON(400, gin.H{"error: ": err.Error()})
		return
	}

	file, _ := fileHeader.Open()
	defer file.Close()

	userId, err := uuid.Parse(r.GetString("user_id"))
	if err != nil{
		log.Print("Failed parse useri: ", err)
		return
	}

	result, err := d.service.Update_document(req, file, fileHeader, userId)
	if err != nil {
		log.Print("Failed update document: ", err)
		r.JSON(500,"internal server error")
	}

	if !result.Status {
    r.JSON(500, result)
    return
}

	r.JSON(200, result)
}


func (d *DocumentHandler) Delete_document_handler(r *gin.Context){
	var req schemas.DeleteRequest
	if err := r.ShouldBindBodyWithJSON(&req); err != nil {
		r.JSON(400, gin.H{"error": err.Error()})
		return
	}

	documentId,err := uuid.Parse(req.DocumentId)
	if err != nil {
		log.Print("error ", err)
		return
	}
	userId,err := uuid.Parse(r.GetString("user_id"))
	if err != nil{
		log.Print("error", err)
	}	

	result, err:= d.service.Delete_document(documentId,userId)
	if err != nil{
		r.JSON(500, "internal server error")
	}

	if !result.Status{
		r.JSON(500,"internal server error")
		return
	}

	r.JSON(200,result)
}


func (d *DocumentHandler) Approval_document_handler(r *gin.Context){
	var req schemas.ApprovalRequest
	if err := r.ShouldBindBodyWithJSON(&req); err != nil {
		log.Print("errir json: ", err)
		r.JSON(400, gin.H{"error": err.Error()})
		return
	}
	documentId, err := uuid.Parse(req.DocumentId)
	if err != nil {
		log.Print("error parsing document id: ", err)
		return 
	}

	userId, err := uuid.Parse(r.GetString("user_id"))
	if err != nil{
		log.Print("error parsing document id: ", err)
	}

	result, err := d.service.Approval_document(documentId,userId ,req.Status)
	if err != nil{
		log.Print("error ", err)
		r.JSON(500,"internal server error")
		return
	}

	if !result.Status{
		r.JSON(500,"internal server error")
	}

	r.JSON(200, result)
}

