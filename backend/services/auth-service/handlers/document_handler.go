package handlers

import (
	"net/http"

	"auth-service/config"
	"auth-service/models"
	"auth-service/services"

	"github.com/gin-gonic/gin"
)

// DocumentHandler provides example endpoints to demonstrate protected routes with RBAC.
// In a real application, these would interact with a document storage service.
type DocumentHandler struct{}

// NewDocumentHandler creates a DocumentHandler instance.
func NewDocumentHandler() *DocumentHandler {
	return &DocumentHandler{}
}

// getAuthenticatedUser is a helper that extracts the JWT claims from the Gin context.
// Returns nil if the user is not authenticated (should not happen behind JWTAuth middleware).
func getAuthenticatedUser(c *gin.Context) *services.Claims {
	value, exists := c.Get(config.ContextKeyUser)
	if !exists {
		return nil
	}
	claims, ok := value.(*services.Claims)
	if !ok {
		return nil
	}
	return claims
}

// ListDocuments handles GET /documents
// Accessible by ALL authenticated roles (staff, admin, master_admin).
//
// Demonstrates reading user info from the Gin context after JWT middleware.
func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	user := getAuthenticatedUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Authentication required.",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Documents retrieved successfully.",
		Data: gin.H{
			"user":  user.Email,
			"role":  user.Role,
			"documents": []gin.H{
				{"id": 1, "title": "Project Proposal", "status": "approved"},
				{"id": 2, "title": "Budget Report Q1", "status": "pending"},
				{"id": 3, "title": "Meeting Notes", "status": "draft"},
			},
		},
	})
}

// UploadDocument handles POST /upload
// Accessible only by "admin" and "master_admin".
//
// Role enforcement is handled by the RequireRoles middleware in the route definition,
// so by the time this handler runs, we know the user is authorized.
func (h *DocumentHandler) UploadDocument(c *gin.Context) {
	user := getAuthenticatedUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Authentication required.",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Document uploaded successfully.",
		Data: gin.H{
			"uploaded_by": user.Email,
			"role":        user.Role,
		},
	})
}

// ApproveDocument handles POST /approve
// Accessible only by "master_admin".
//
// This is the most restricted endpoint — only the highest-privilege role can approve.
func (h *DocumentHandler) ApproveDocument(c *gin.Context) {
	user := getAuthenticatedUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Authentication required.",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Document approved successfully.",
		Data: gin.H{
			"approved_by": user.Email,
			"role":        user.Role,
		},
	})
}
