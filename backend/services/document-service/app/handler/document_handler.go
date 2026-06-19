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
	"io"

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
func(d *DocumentHandler) Get_public_document_by_id_handler(c *gin.Context){
	documentId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error" :"invalid Document ID"})
		return
	}

	filepath, err := d.documentService.Get_public_document_by_id(documentId, "public")
	if err != nil{
		c.JSON(500, gin.H{"error" : "Failed to get filepath"})
		return
	}

	c.JSON(200, gin.H{
		"url" : filepath,
	})
}

func (d *DocumentHandler) Get_document_by_id_handler(c *gin.Context) {
    documentId, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
        return
    }

    createdById, err := uuid.Parse(c.GetString("user_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }

    object, stat, err := d.documentService.Get_document_by_id(c.Request.Context(), documentId, createdById, c.GetString("role"))
    if err != nil {
		 log.Printf("StreamDocument error: %v", err) 
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stream document"})
        return
    }
    defer object.Close()
	 contentType := stat.ContentType
	 log.Printf("streaming file: '%s', content-type: '%s', size: %d", stat.Key, contentType, stat.Size)

    c.Header("Content-Type", stat.ContentType)
    c.Header("Content-Disposition", "inline; filename="+stat.Key)
    c.Header("Content-Length", strconv.FormatInt(stat.Size, 10))

    io.Copy(c.Writer, object)
}

func (d *DocumentHandler) Get_document_by_status_handler(c *gin.Context) {
	statusParam := c.Param("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	status, page, limit, err := d.documentService.Get_document_by_status(statusParam, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  status,
		"page":  page,
		"limit": limit,
	})
}

func (d *DocumentHandler) Get_document_by_type_handler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	docs, page, limit, err := d.documentService.Get_documents_by_type(page, limit)
	if err != nil {
		log.Print("Error fetching public documents: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  docs,
		"page":  page,
		"limit": limit,
	})
}

func (d *DocumentHandler) Get_document_by_group_handler(c *gin.Context) {
	groupId, err := uuid.Parse(c.Param("group"))
	if err != nil {
		log.Print("Error parse groupid uuid: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	group, page, limit, err := d.documentService.Get_document_by_group(groupId, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  group,
		"page":  page,
		"limit": limit,
	})
}

func (d *DocumentHandler) Get_document_by_standard_handler(c *gin.Context) {
	standardId, err := uuid.Parse(c.Param("standard"))
	if err != nil {
		log.Print("error Parse standard uuid: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid standard ID"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	userId, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		log.Print("Error Parse user id: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	role := c.GetString("role")
	standard, page, limit, err := d.documentService.Get_document_by_standard(standardId, userId, page, limit, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": standard, "page": page, "limit": limit})
}

func (d *DocumentHandler) Get_document_by_service_handler(c *gin.Context) {
	serviceId, err := uuid.Parse(c.Param("service"))
	if err != nil {
		log.Print("error parsing documentService uuid: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service ID"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	userId, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		log.Print("Error Parse user id: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	role := c.GetString("role")
	service, page, limit, err := d.documentService.Get_document_by_service(serviceId, userId, page, limit, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": service, "page": page, "limit": limit})
}

func (d *DocumentHandler) Get_document_by_assessment_handler(c *gin.Context) {
	assessmentId, err := uuid.Parse(c.Param("assessment"))
	if err != nil {
		log.Print("error parse assessment uuid: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assessment ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	userId, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		log.Print("Error Parse user id: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	role := c.GetString("role")
	assessment, page, limit, err := d.documentService.Get_document_by_assessment(assessmentId, userId, page, limit, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": assessment, "page": page, "limit": limit})
}

func (d *DocumentHandler) Get_document_by_createdBy_handler(c *gin.Context) {
	userId, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		log.Print("error parsing user_id from token: ", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	docs, page, limit, err := d.documentService.Get_document_by_createdBy(userId, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if docs == nil {
		c.JSON(http.StatusOK, gin.H{
			"data":  []interface{}{},
			"page":  page,
			"limit": limit,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  docs,
		"page":  page,
		"limit": limit,
	})
}

func (d *DocumentHandler) Create_document_handler(c *gin.Context) {
	var req schemas.DocumentRequest
	if err := c.ShouldBind(&req); err != nil {
		log.Print("error parse json: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fileHeader, err := c.FormFile("uploadedFile")
	if err != nil {
		log.Print("Error getting file: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file, _ := fileHeader.Open()
	defer file.Close()

	userId, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := c.GetString("role")

	result, err := d.documentService.Create_document(req, file, fileHeader, userId, role)
	if err != nil {
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	if !result.Status {
		switch result.Message{
		case "filename already exist":
			c.JSON(409, result)
		case "Not Authorized, service/standard/assessment is not under the current group":
			c.JSON(401, result)
		default:
			c.JSON(500, result)
		}
		return
	}

	c.JSON(200, result)
}
 
func (d *DocumentHandler) Update_document_handler(c *gin.Context) {
    var req schemas.UpdateRequest

    if err := c.ShouldBind(&req); err != nil {
        log.Print("Error binding request: ", err)
        c.JSON(400, gin.H{"status": false, "message": err.Error()})
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
            c.JSON(400, gin.H{"status": false, "message": "Failed to read file"})
            return
        }
    } else {
        file, err = fileHeader.Open()
        if err != nil {
            log.Print("Error opening file: ", err)
            c.JSON(500, gin.H{"status": false, "message": "Failed to open file"})
            return
        }
        defer file.Close()
    }

    userId, err := uuid.Parse(c.GetString("user_id"))
    userRole := c.GetString("role")
    if err != nil {
        log.Print("Failed parse user id: ", err)
        c.JSON(400, gin.H{"status": false, "message": "Invalid user"})
        return
    }
    
    result, err := d.documentService.Update_document(req, file, fileHeader, userId, userRole)
    if err != nil {
        log.Print("Failed update document: ", err)
        c.JSON(500, gin.H{"status": false, "message": "Internal server error"})
        return
    }

    if !result.Status {
		switch result.Message{
		case "Cannot Edit Approved Document":
			c.JSON(409, result)
		case "Unauthorized: cannot edit this document":
			c.JSON(401,result)
		default:
			c.JSON(500,result)
		}
        return
    }

    c.JSON(200, result)
}


func (d *DocumentHandler) Delete_document_handler(c *gin.Context) {
	documentIdParam := c.Param("documentId")
	documentId, err := uuid.Parse(documentIdParam)
	if err != nil {
		log.Print("error parse documentid: ", err)
		c.JSON(400, gin.H{"error": "Invalid document ID"})
		return
	}
	userId, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		log.Print("error parsing user id: ", err)
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}
	userRole := c.GetString("role")

	result, err := d.documentService.Delete_document(documentId, userId, userRole)
	if err != nil {
		c.JSON(500, result)
		return
	}

	if !result.Status {
		c.JSON(500, result)
		return
	}

	c.JSON(200, result)
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
		switch result.Message{
		case "Document integrity check failed. The file hash does not match the original document fingerprint":
			c.JSON(400, result)
		default:
			c.JSON(500, result)
		}
		return
	}

	c.JSON(200, result)
}

func (h *DocumentHandler) GetStats(c *gin.Context) {
	userId, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	role := c.GetString("role")

	result, err := h.documentService.GetStats(userId, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Failed to fetch statistics",
		})
		return
	}
	c.JSON(http.StatusOK, result)
}

