package api

import (
	"capstone/app/repositories"
	"errors"
	"log"
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func Get_all_documents_Handler(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    
    if page < 1 { page = 1 }
    if limit < 1 || limit > 100 { limit = 20 }
    
    offset := (page - 1) * limit
    docs, err := repositories.GetDocuments(limit, offset)
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

func Get_document_by_id_handler(r *gin.Context){
	document_id := r.Param("id")
	id, err := uuid.Parse(document_id)
	if err != nil{
		log.Print("error to parse id")
	}

	docs, err := repositories.Get_document_by_id(id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows){
			r.JSON(404,"Document Not Found")
			return
		}
		r.JSON(500,"Internal error")
	}

	r.JSON(200,docs)
}

func Get_document_by_status_handler(r *gin.Context){
	statusParam := r.Param("status")
	page, _ := strconv.Atoi(r.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(r.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }
    if limit < 1 || limit > 100 { limit = 20 }
    
    offset := (page - 1) * limit
	status, err := repositories.Get_document_by_status(statusParam,limit,offset)
	if err != nil {
		r.JSON(500,"internal server error")
		return
	}
	r.JSON(200,status)
}

func Get_document_by_type_handler(r *gin.Context){
	types := r.Param("type")
	typeId, err := uuid.Parse(types)
	if err != nil{
		log.Print("Error parse type into uuid; ", err)
		return
	}
	page, _ := strconv.Atoi(r.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(r.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }
    if limit < 1 || limit > 100 { limit = 20 }
    
    offset := (page - 1) * limit
	Type, err := repositories.Get_document_by_type(typeId,limit,offset)
	if err != nil{
		log.Print("Error terus wok", err)
		r.JSON(500,"Internal Server Error")
		return
	}
	r.JSON(200,Type)
}

func Get_document_by_group_handler(r *gin.Context){
	groupIdParam := r.Param("group")
	groupId, err := uuid.Parse(groupIdParam)
	if err != nil{
		log.Print("Error parse groupid uuid: ", err)
		return
	}
	page, _ := strconv.Atoi(r.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(r.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }
    if limit < 1 || limit > 100 { limit = 20 }
    
    offset := (page - 1) * limit
	
	group, err:= repositories.Get_document_by_group(groupId,limit,offset)
	if err != nil{
		r.JSON(500,"Internal Server Error")
		return
	}
	r.JSON(200, group)
}

func Get_document_by_standard_handler(r *gin.Context){
	standardParam := r.Param("standard")
	standardId, err := uuid.Parse(standardParam)
	if err != nil {
		log.Print("error Parse standard uui: ",err)
		return
	}
	page, _ := strconv.Atoi(r.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(r.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }
    if limit < 1 || limit > 100 { limit = 20 }
    
    offset := (page - 1) * limit

	standard, err := repositories.Get_document_by_standard(standardId,limit,offset)
	if err != nil{
		r.JSON(500,"internal server error")
		return 
	}
	r.JSON(200,standard)
}

func Get_document_by_service_handler(r *gin.Context){
	serviceParam := r.Param("service")
	serviceId, err := uuid.Parse(serviceParam)
	if err != nil {
		log.Print("error parsing service uuid: ", err)
		return
	}
	page, _ := strconv.Atoi(r.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(r.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }
    if limit < 1 || limit > 100 { limit = 20 }
    
    offset := (page - 1) * limit

	service, err := repositories.Get_document_by_service(serviceId,limit,offset)
	if err != nil {
		r.JSON(500, "internal server error")
		return
	}
	r.JSON(200,service)
}

func Get_document_by_assessment_handler(r *gin.Context){
	assessmentParam := r.Param("assessment")
	assessmentId, err := uuid.Parse(assessmentParam)
	if err !=nil {	
		log.Print("error parse assessment uuid: ", err)
		return
	}

	page, _ := strconv.Atoi(r.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(r.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }
    if limit < 1 || limit > 100 { limit = 20 }
    
    offset := (page - 1) * limit
	assessment, err := repositories.Get_document_by_assessment(assessmentId,limit,offset)
	if err != nil {
		r.JSON(500, "internal server error")
		return
	}
	r.JSON(200,assessment)
}

func Get_document_by_createdBy_handler(r *gin.Context){
	createdByParam := r.Param("createdBy")
	createdById, err := uuid.Parse(createdByParam)
	if err != nil {
		log.Print("error parsing createdBy uuid: ", err)
		return
	}
	page, _ := strconv.Atoi(r.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(r.DefaultQuery("limit", "20"))
	if page < 1 { page = 1 }
    if limit < 1 || limit > 100 { limit = 20 }
    
    offset := (page - 1) * limit
	createdBy, err := repositories.Get_document_by_createdBy(createdById,limit,offset)
	if err != nil {
		r.JSON(500, "internal server error")
		return
	}
	r.JSON(200, createdBy)
}

func Create_document_handler(r *gin.Context){
	docs, err := repositories.Create_document(r)
	if err != nil {
		log.Print("ERROR WOK:", err)
		r.JSON(500,"internal server error")
	}
	r.JSON(200,docs)
}
 