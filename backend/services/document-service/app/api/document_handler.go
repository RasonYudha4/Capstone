package api

import (
	"capstone/app/schemas"
	"capstone/app/services"
	"log"
	"strconv"
	"mime/multipart"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

)

type DocumentHandler struct{
	documentService *services.DocumentService
	notificationService *services.NotificationService
}

func NewDocumentHandler(document *services.DocumentService, notification *services.NotificationService) *DocumentHandler{
	return &DocumentHandler{
		documentService : document,
		notificationService : notification,
	}
}

/*
func (d *DocumentHandler)Get_all_documents_Handler(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
    
    docs,err := d.documentService.Get_Document(page, limit)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{
        "data":   docs,
        "page":   page,
        "limit":  limit,
    })
}*/

func (d *DocumentHandler) Get_document_by_id_handler(c *gin.Context){
	documentId, err := uuid.Parse(c.Param("id"))	
	if err != nil{
		log.Print("error to parse document id")
	}

	createdById, err := uuid.Parse(c.GetString("user_id"))
	if err != nil{
		log.Print("error parse user id ", err)
	}

	role := c.GetString("role")
	docs,status ,err := d.documentService.Get_document_by_id(documentId,createdById, role)
	if err != nil {
    	c.JSON(500, gin.H{"message": "Internal error"})
    	return
	}
	if !status.Status {
    	c.JSON(500, gin.H{"message": "Internal server error"})
    	return
	}
	c.JSON(200,gin.H{"Data":docs})
}

func(d *DocumentHandler) Get_document_by_status_handler(c *gin.Context){
	statusParam := c.Param("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	
	status,page, limit, err := d.documentService.Get_document_by_status(statusParam,page,limit)
	if err != nil {
		c.JSON(500,"internal server error")
		return
	}
	 c.JSON(200, gin.H{
        "data":   status,
        "page":   page,
        "limit":  limit,
    })
}

func (d *DocumentHandler)Get_document_by_type_handler(c *gin.Context){
	typeId, err := uuid.Parse(c.Param("type"))
	if err != nil{
		log.Print("Error parse type into uuid; ", err)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	
	Type,page, limit, err := d.documentService.Get_document_by_type(typeId,page,limit)
	if err != nil{
		c.JSON(500,"Internal Server Error")
		return
	}
	c.JSON(200, gin.H{
		"data": Type,
		"page": page,
		"limit": limit,
	})
}

func(d *DocumentHandler) Get_document_by_group_handler(c *gin.Context){
	groupId, err :=  uuid.Parse(c.Param("group"))
	if err != nil{
		log.Print("Error parse groupid uuid: ", err)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	
	group,page,limit, err:= d.documentService.Get_document_by_group(groupId,page,limit)
	if err != nil{
		c.JSON(500,"Internal Server Error")
		return
	}
	c.JSON(200, gin.H{
		"data": group,
		"page": page,
		"limit": limit,
	})
}

func(d *DocumentHandler) Get_document_by_standard_handler(c *gin.Context){
	standardId, err := uuid.Parse(c.Param("standard"))
	if err != nil {
		log.Print("error Parse standard uui: ",err)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	
	userId, err := uuid.Parse(c.GetString("user_id"))
	if err != nil{
		log.Print("Error Parse user id: ", err)
		c.JSON(500,"Internal server error")
		return
	}

	role := c.GetString("role")
	standard, page, limit, err := d.documentService.Get_document_by_standard(standardId,userId,page, limit,role)
	if err != nil{
		c.JSON(500,"internal server error")
		return 
	}
	c.JSON(200, gin.H{
		"data": standard,
		"page": page,
		"limit": limit,
	})
}

func(d *DocumentHandler) Get_document_by_service_handler(c *gin.Context){
	serviceId, err := uuid.Parse(c.Param("service"))
	if err != nil {
		log.Print("error parsing documentService uuid: ", err)
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	userId, err := uuid.Parse(c.GetString("user_id"))
	if err != nil{
		log.Print("Error Parse user id: ", err)
		c.JSON(500,"Internal server error")
		return
	}

	role := c.GetString("role")
	service,page,limit, err := d.documentService.Get_document_by_service(serviceId,userId,page,limit,role)
	if err != nil {
		c.JSON(500, "internal server error")
		return
	}
	c.JSON(200, gin.H{
		"data": service,
		"page": page,
		"limit": limit,
	})
}

func(d *DocumentHandler) Get_document_by_assessment_handler(c *gin.Context){
	assessmentId, err := uuid.Parse(c.Param("assessment"))
	if err !=nil {	
		log.Print("error parse assessment uuid: ", err)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	userId, err := uuid.Parse(c.GetString("user_id"))
	if err != nil{
		c.JSON(500,"Internal server error")
		return
	}

	role := c.GetString("role")
	assessment,page,limit,err := d.documentService.Get_document_by_assessment(assessmentId,userId,page,limit, role)
	if err != nil {
		c.JSON(500, "internal server error")
		return
	}
	c.JSON(200, gin.H{
		"data": assessment, 
		"page": page,
		"limit": limit,
	})
}

func (d *DocumentHandler) Get_document_by_createdBy_handler(c *gin.Context) {
	userId, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		log.Print("error parsing user_id from token: ", err)
		c.JSON(401, gin.H{"error": "unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	docs, page, limit, err := d.documentService.Get_document_by_createdBy(userId, page, limit)
	if err != nil {
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}

	if docs == nil {
		c.JSON(200, gin.H{
			"data":  []interface{}{},
			"page":  page,
			"limit": limit,
		})
		return
	}

	c.JSON(200, gin.H{
		"data":  docs,
		"page":  page,
		"limit": limit,
	})
}

func(d *DocumentHandler) Create_document_handler(c *gin.Context){
	var req schemas.DocumentRequest
	if err := c.ShouldBind(&req); err != nil{
		log.Print("error parse json: ", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	fileHeader, err := c.FormFile("uploadedFile")
	if err != nil {
		log.Print("Error getting file: ", err)
		c.JSON(400, gin.H{"Error": err.Error()})
		return
	}

	file, _ := fileHeader.Open()
	defer file.Close()

	userId,err := uuid.Parse(c.GetString("user_id"))
	if err != nil{
		c.JSON(400, gin.H{"errir": err.Error()})
		return
	}

	result, err := d.documentService.Create_document(req,file,fileHeader,userId)
	if err != nil{
		c.JSON(500,"Internal server error")
		return
	}

	if !result.Status{
		c.JSON(500, "internal server error")
	}

	c.JSON(200,result)

}
 
func (d *DocumentHandler) Update_document_handler(c *gin.Context) {
    var req schemas.UpdateRequest

    if err := c.ShouldBind(&req); err != nil {
        log.Print("Error binding request: ", err)
        c.JSON(400, gin.H{
            "status": false,
            "message": err.Error(),
        })
        return
    }

    fileHeader, err := c.FormFile("uploadedFile")

    var file multipart.File

    if err != nil {
        if err == http.ErrMissingFile {
            file = nil
            fileHeader = nil
        } else {
            log.Print("Error reading file: ", err)
            c.JSON(400, gin.H{
                "status": false,
                "message": "Failed to read file",
            })
            return
        }
    } else {
        file, err = fileHeader.Open()
        if err != nil {
            log.Print("Error opening file: ", err)
            c.JSON(500, gin.H{
                "status": false,
                "message": "Failed to open file",
            })
            return
        }
        defer file.Close()
    }

    
    userId, err := uuid.Parse(c.GetString("user_id"))
    if err != nil {
        log.Print("Failed parse user id: ", err)
        c.JSON(400, gin.H{
            "status": false,
            "message": "Invalid user",
        })
        return
    }
    
    result, err := d.documentService.Update_document(req, file, fileHeader, userId)
    if err != nil {
        log.Print("Failed update document: ", err)
        c.JSON(500, gin.H{
            "status": false,
            "message": "Internal server error",
        })
        return
    }

    if !result.Status {
        c.JSON(400, result)
        return
    }

    c.JSON(200, result)
}


func (d *DocumentHandler) Delete_document_handler(c *gin.Context){
	documentIdParam := c.Param("documentId")
	documentId,err := uuid.Parse(documentIdParam)
	if err != nil {
		log.Print("error parse documentid", err)
		return
	}
	userId,err := uuid.Parse(c.GetString("user_id"))
	if err != nil{
		log.Print("error", err)
	}	

	result, err:= d.documentService.Delete_document(documentId,userId)
	if err != nil{
		c.JSON(500, result)
	}

	if !result.Status{
		c.JSON(500, result)
		return
	}

	c.JSON(200,result)
}


func (d *DocumentHandler) Approval_document_handler(c *gin.Context) {
	var req schemas.ApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Print("error json: ", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	documentId, err := uuid.Parse(req.DocumentId)
	if err != nil {
		log.Print("error parsing document id: ", err)
		c.JSON(400, gin.H{"error": "invalid document id"})
		return
	}

	userId, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		log.Print("error parsing user id: ", err)
		c.JSON(400, gin.H{"error": "invalid user id"})
		return
	}

	var file multipart.File
	var fileHeader *multipart.FileHeader

	uploadedFile, header, err := c.Request.FormFile("document")
	if err == nil {
		file = uploadedFile
		fileHeader = header
		defer file.Close()
	}

	result, err := d.documentService.Approval_document(documentId, userId, req.Status, file, fileHeader)
	if err != nil {
		log.Print("error ", err)
		c.JSON(500, gin.H{"error": "internal server error"})
		return
	}

	if !result.Status {
		c.JSON(400, result)
		return
	}

	c.JSON(200, result)
}

func (h *DocumentHandler) GetStats(c *gin.Context) {
	userId, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(500,"Internal server error")
		return
	}

	role := c.GetString("role")


	result, err := h.documentService.GetStats(userId,role)
	if err != nil {
		c.JSON(500, gin.H{
			"status":  false,
			"message": "Failed to fetch statistics",
		})
		return
	}
	c.JSON(200, result)
}

