package api

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"

	"capstone/app/schemas"
	"capstone/app/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DocumentHandler struct {
	documentService     *services.DocumentService
	notificationService *services.NotificationService
}

func NewDocumentHandler(document *services.DocumentService, notification *services.NotificationService) *DocumentHandler {
	return &DocumentHandler{
		documentService:     document,
		notificationService: notification,
	}
}

func RespondError(c *gin.Context, httpStatus int, message string) {
	c.JSON(httpStatus, schemas.ApiResponse{
		Success:    false,
		Message:    message,
		StatusCode: httpStatus,
		Data:       nil,
	})
}

func RespondSuccess(c *gin.Context, httpStatus int, message string, data any) {
	c.JSON(httpStatus, schemas.ApiResponse{
		Success:    true,
		Message:    message,
		StatusCode: httpStatus,
		Data:       data,
	})
}

func (d *DocumentHandler) Get_public_document_by_id_handler(c *gin.Context) {
	documentId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		RespondError(c, 400, "Bad Request, Error parse document id")
		return
	}

	status,url, contentType, err := d.documentService.Get_public_document_by_id(documentId, "public")
	if err != nil {
		log.Print("error getting public document url: ", err)
		RespondError(c, 500, "Failed Getting Url")
		return
	}

	if !status.Status{
		switch status.Message{
		case "Document integrity check failed. The file hash does not match the original document fingerprint":
			RespondError(c, 409, status.Message)
			return
		default:
			RespondError(c,500,"Internal Server Error")
			return
		}
	}

	RespondSuccess(c, 200, "Success getting Url", schemas.PresignedUrlResponse{
		PresignedUrl: url,
		ContentType:  contentType,
	})
}

func (d *DocumentHandler) Get_document_by_id_handler(c *gin.Context) {
	documentId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		RespondError(c, http.StatusBadRequest, "Bad Request, Error parse document id")
		return
	}

	createdById, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		RespondError(c, http.StatusBadRequest, "Bad Request, Error parse user id")
		return
	}

	status,object, stat, err := d.documentService.Get_document_by_id(c.Request.Context(), documentId, createdById, c.GetString("role"))
	if err != nil {
		log.Printf("StreamDocument error: %v", err)
		RespondError(c, http.StatusInternalServerError, "Failed to stream document")
		return
	}
	defer object.Close()

	
	if !status.Status{
		switch status.Message{
		case "Document integrity check failed. The file hash does not match the original document fingerprint":
			RespondError(c, 409, status.Message)
			return
		default:
			RespondError(c,500,"Internal Server Error")
			return
		}
	}

	contentType := stat.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	log.Printf("streaming file: content-type: '%s', size: %d", contentType, stat.Size)

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, stat.Key))
	c.Header("Content-Length", strconv.FormatInt(stat.Size, 10))
	c.Status(http.StatusOK)

	if _, err := io.Copy(c.Writer, object); err != nil {
		log.Printf("error streaming document to client: %v", err)
	}
}

func (d *DocumentHandler) Get_document_by_status_handler(c *gin.Context) {
	statusParam := c.Param("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	data, page, limit, err := d.documentService.Get_document_by_status(statusParam, page, limit)
	if err != nil {
		RespondError(c, 500, "Internal Server Error")
		return
	}
	RespondSuccess(c, 200, "Success", schemas.DocumentDataResponse{
		Data:  data,
		Page:  page,
		Limit: limit,
	})
}

func (d *DocumentHandler) Get_document_by_type_handler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	data, page, limit, err := d.documentService.Get_documents_by_type(page, limit)
	if err != nil {
		RespondError(c, 500, "Internal Server Error")
		return
	}

	RespondSuccess(c, 200, "Success", schemas.DocumentDataResponse{
		Data:  data,
		Page:  page,
		Limit: limit,
	})
}

func (d *DocumentHandler) Get_document_by_group_handler(c *gin.Context) {
	groupId, err := uuid.Parse(c.Param("group"))
	if err != nil {
		RespondError(c, 400, "Bad Request, Error parse Group Id")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	data, page, limit, err := d.documentService.Get_document_by_group(groupId, page, limit)
	if err != nil {
		RespondError(c, 500, "Internal Server Error")
		return
	}
	RespondSuccess(c, 200, "Success", schemas.DocumentDataResponse{
		Data:  data,
		Page:  page,
		Limit: limit,
	})
}

func (d *DocumentHandler) Get_document_by_standard_handler(c *gin.Context) {
	standardId, err := uuid.Parse(c.Param("standard"))
	if err != nil {
		RespondError(c, 400, "Bad Request, Error parse Standard Id")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	userId, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		RespondError(c, 400, "Bad Request, Error parse User Id")
		return
	}

	role := c.GetString("role")
	data, page, limit, err := d.documentService.Get_document_by_standard(standardId, userId, page, limit, role)
	if err != nil {
		RespondError(c, 500, "Internal Server Error")
		return
	}
	RespondSuccess(c, 200, "Success", schemas.DocumentDataResponse{
		Data:  data,
		Page:  page,
		Limit: limit,
	})
}

func (d *DocumentHandler) Get_document_by_service_handler(c *gin.Context) {
	serviceId, err := uuid.Parse(c.Param("service"))
	if err != nil {
		RespondError(c, 400, "Bad Request, Error parse service id")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	userId, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		RespondError(c, 400, "Bad Request, Error parse User Id")
		return
	}

	role := c.GetString("role")
	data, page, limit, err := d.documentService.Get_document_by_service(serviceId, userId, page, limit, role)
	if err != nil {
		RespondError(c, 500, "Internal Server Error")
		return
	}
	RespondSuccess(c, 200, "Success", schemas.DocumentDataResponse{
		Data:  data,
		Page:  page,
		Limit: limit,
	})
}

func (d *DocumentHandler) Get_document_by_assessment_handler(c *gin.Context) {
	assessmentId, err := uuid.Parse(c.Param("assessment"))
	if err != nil {
		RespondError(c, 400, "Bad Request, Error Parse assessment id")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	userId, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		RespondError(c, 400, "Bad Request, error parse user id")
		return
	}

	role := c.GetString("role")
	data, page, limit, err := d.documentService.Get_document_by_assessment(assessmentId, userId, page, limit, role)
	if err != nil {
		RespondError(c, 500, "Internal Server Error")
		return
	}
	RespondSuccess(c, 200, "Success", schemas.DocumentDataResponse{
		Data:  data,
		Page:  page,
		Limit: limit,
	})
}

func (d *DocumentHandler) Get_document_by_createdBy_handler(c *gin.Context) {
	userId, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		RespondError(c, 400, "Bad Request, Error parse user id")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	data, page, limit, err := d.documentService.Get_document_by_createdBy(userId, page, limit)
	if err != nil {
		RespondError(c, 500, "Internal Server Error")
		return
	}

	RespondSuccess(c, 200, "Success", schemas.DocumentDataResponse{
		Data:  data,
		Page:  page,
		Limit: limit,
	})
}

func (d *DocumentHandler) Create_document_handler(c *gin.Context) {
	var req schemas.DocumentRequest
	if err := c.ShouldBind(&req); err != nil {
		RespondError(c, 400, "Bad Request, Error parse multiform/form")
		return
	}

	fileHeader, err := c.FormFile("uploadedFile")
	if err != nil {
		RespondError(c, 400, "Bad Request, Error Getting File")
		return
	}

	fileExtension := filepath.Ext(fileHeader.Filename)
	allowedFile := map[string]bool{
		".pdf":  true,
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}

	allowedMimeTypes := map[string]bool{
		"application/pdf": true,
		"image/jpeg":      true,
		"image/jpg":       true,
		"image/png":       true,
	}

	if !allowedFile[fileExtension] {
		RespondError(c, 400, "Bad Request, File Not Allowed")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		RespondError(c, 500, "Internal Server Error, Error Opening File")
		return
	}
	defer file.Close()

	userId, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		RespondError(c, 400, "Bad Request, Error parse user id")
		return
	}

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		RespondError(c, 500, "Internal Server Error, Error Reading File")
		return
	}
	detectedType := http.DetectContentType(buf[:n])

	if !allowedMimeTypes[detectedType] {
		RespondError(c, 400, "Bad Request, File Content Doesn't Match Allowed Types")
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		RespondError(c, 500, "Internal Server Error, Error Seeking File")
		return
	}

	role := c.GetString("role")

	result, err := d.documentService.Create_document(req, file, fileHeader.Filename, fileHeader.Size, fileHeader.Header.Get("Content-Type"), userId, role)
	if err != nil {
		RespondError(c, 500, "Internal Server Error")
		return
	}

	if !result.Status {
		switch result.Message {
		case "filename already exist":
			RespondError(c, 409, "Conflict, Filename already exist")
		case "Not Authorized, service/standard/assessment is not under the current group":
			RespondError(c, 401, "Unauthorized, Access denied")
		default:
			RespondError(c, 500, "Internal Server Error")
		}
		return
	}

	RespondSuccess(c, 200, "Successfully uploaded file", nil)
}

func (d *DocumentHandler) Update_document_handler(c *gin.Context) {
	var req schemas.UpdateRequest
	if err := c.ShouldBind(&req); err != nil {
		RespondError(c, 400, "Bad Request, Error parse multiform/form")
		return
	}

	allowedFile := map[string]bool{
		".pdf":  true,
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}

	allowedMimeTypes := map[string]bool{
		"application/pdf": true,
		"image/jpeg":      true,
		"image/jpg":       true,
		"image/png":       true,
	}

	var (
		file        multipart.File
		fileSize    int64
		contentType string
	)

	fileHeader, err := c.FormFile("uploadedFile")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			file = nil
			fileHeader = nil
		} else {
			RespondError(c, 400, "Bad Request, Error getting file")
			return
		}
	} else {
		fileExtension := filepath.Ext(fileHeader.Filename)
		if !allowedFile[fileExtension] {
			RespondError(c, 400, "Bad Request, File Not Allowed")
			return
		}

		file, err = fileHeader.Open()
		if err != nil {
			RespondError(c, 500, "Internal Server Error, Error Opening File")
			return
		}
		defer file.Close()

		buf := make([]byte, 512)
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			RespondError(c, 500, "Internal Server Error, Error Reading File")
			return
		}
		detectedType := http.DetectContentType(buf[:n])

		if !allowedMimeTypes[detectedType] {
			RespondError(c, 400, "Bad Request, File Content Doesn't Match Allowed Types")
			return
		}

		if _, err := file.Seek(0, io.SeekStart); err != nil {
			RespondError(c, 500, "Internal Server Error, Error Seeking File")
			return
		}

		fileSize = fileHeader.Size
		contentType = detectedType
	}

	userId, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		RespondError(c, 400, "Bad Request, Error parse user id")
		return
	}
	userRole := c.GetString("role")

	result, err := d.documentService.Update_document(req, file, fileSize, contentType, userId, userRole)
	if err != nil {
		RespondError(c, 500, "Internal Server Error")
		return
	}

	if !result.Status {
		switch result.Message {
		case "Cannot Edit Approved Document":
			RespondError(c, 409, "Conflict, Cannot edit Approved document")
		case "Unauthorized: cannot edit this document":
			RespondError(c, 401, "Unauthorized, Access Denied")
		default:
			RespondError(c, 500, "Internal Server Error")
		}
		return
	}

	RespondSuccess(c, 200, "Success Updating File", nil)
}

func (d *DocumentHandler) Delete_document_handler(c *gin.Context) {
	documentId, err := uuid.Parse(c.Param("documentId"))
	if err != nil {
		RespondError(c, 400, "Bad Request, Error parse Document Id")
		return
	}

	userId, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		RespondError(c, 400, "Bad Request, Error parse User id")
		return
	}
	userRole := c.GetString("role")

	result, err := d.documentService.Delete_document(documentId, userId, userRole)
	if err != nil {
		RespondError(c, 500, "Internal Server Error")
		return
	}

	if !result.Status {
		RespondError(c, 500, result.Message)
		return
	}

	RespondSuccess(c, 200, "Success Delete File", nil)
}

func (d *DocumentHandler) Approval_document_handler(c *gin.Context) {
	var req schemas.ApprovalRequest
	if err := c.ShouldBind(&req); err != nil {
		RespondError(c, 400, "Bad Request, Error parse Multiform/form")
		return
	}

	documentId, err := uuid.Parse(req.DocumentId)
	if err != nil {
		RespondError(c, 400, "Bad Request, Error Parse document id")
		return
	}

	userId, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		RespondError(c, 400, "Bad Request, Error parse User id")
		return
	}

	var (
		file        multipart.File
		fileSize    int64
		contentType string
	)

	if fileHeader, err := c.FormFile("uploadedFile"); err == nil {

		fileExtension := filepath.Ext(fileHeader.Filename)
		allowedFile := map[string]bool{
			".pdf":  true,
			".jpg":  true,
			".jpeg": true,
			".png":  true,
		}

		allowedMimeTypes := map[string]bool{
			"application/pdf": true,
			"image/jpeg":      true,
			"image/png":       true,
		}

		if !allowedFile[fileExtension] {
			RespondError(c, 400, "Bad Request, File Not Allowed")
			return
		}

		f, err := fileHeader.Open()
		if err != nil {
			RespondError(c, 400, "Bad Request, Error Getting File")
			return
		}
		defer f.Close()

		buf := make([]byte, 512)
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			RespondError(c, 500, "Internal Server Error, Error Reading File")
			return
		}
		detectedType := http.DetectContentType(buf[:n])

		if !allowedMimeTypes[detectedType] {
			RespondError(c, 400, "Bad Request, File Content Doesn't Match Allowed Types")
			return
		}

		if _, err := f.Seek(0, io.SeekStart); err != nil {
			RespondError(c, 500, "Internal Server Error, Error Seeking File")
			return
		}

		file = f
		fileSize = fileHeader.Size
		contentType = detectedType
	}

	result, err := d.documentService.Approval_document(documentId, userId, req.Status, file, fileSize, contentType)
	if err != nil {
		RespondError(c, 500, "Internal Server Error")
		return
	}

	if !result.Status {
		switch result.Message {
		case "Document integrity check failed. The file hash does not match the original document fingerprint":
			RespondError(c, 400, "Bad Request, File hash does not match the original document fingerprint")
		default:
			RespondError(c, 500, "Internal Server Error")
		}
		return
	}

	RespondSuccess(c, 200, "Success Document approval", nil)
}

func (h *DocumentHandler) GetStats(c *gin.Context) {
	userId, err := uuid.Parse(c.GetString("user_id"))
	if err != nil {
		RespondError(c, 400, "Bad Request, Error parse user id")
		return
	}

	role := c.GetString("role")
	result, err := h.documentService.GetStats(userId, role)
	if err != nil {
		RespondError(c, 500, "Internal Server Error")
		return
	}

	RespondSuccess(c, 200, "Getting Current Stats", result)
}
