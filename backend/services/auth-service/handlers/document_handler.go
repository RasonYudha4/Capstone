package handlers

import (
	"github.com/gin-gonic/gin"
)

// DocumentHandler provides example endpoints to demonstrate protected routes with RBAC.
// In a real application, these would interact with a document storage service.
type DocumentHandler struct{}

// NewDocumentHandler creates a DocumentHandler instance.
func NewDocumentHandler() *DocumentHandler {
	return &DocumentHandler{}
}

// ListDocuments handles GET /documents.
// Accessible by ALL authenticated roles (staff, admin, master_admin).
func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	user := requireClaims(c)
	if user == nil {
		return
	}

	respondSuccess(c, "Documents retrieved successfully.", gin.H{
		"user": user.Email,
		"role": user.Role,
		"documents": []gin.H{
			{"id": 1, "title": "Project Proposal", "status": "approved"},
			{"id": 2, "title": "Budget Report Q1", "status": "pending"},
			{"id": 3, "title": "Meeting Notes", "status": "draft"},
		},
	})
}

// UploadDocument handles POST /upload.
// Accessible only by "admin" and "master_admin" (enforced by middleware).
func (h *DocumentHandler) UploadDocument(c *gin.Context) {
	user := requireClaims(c)
	if user == nil {
		return
	}

	respondSuccess(c, "Document uploaded successfully.", gin.H{
		"uploaded_by": user.Email,
		"role":        user.Role,
	})
}

// ApproveDocument handles POST /approve.
// Accessible only by "master_admin" (enforced by middleware).
func (h *DocumentHandler) ApproveDocument(c *gin.Context) {
	user := requireClaims(c)
	if user == nil {
		return
	}

	respondSuccess(c, "Document approved successfully.", gin.H{
		"approved_by": user.Email,
		"role":        user.Role,
	})
}
