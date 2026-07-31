package services

import (
	"bytes"
	"capstone/app/schemas"
	"context"
	"errors"
	"mime/multipart"
	"net/textproto"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─────────────────────────────────────────────
// Mock: DocumentRepo
// ─────────────────────────────────────────────

type MockDocumentRepo struct {
	mock.Mock
}

func (m *MockDocumentRepo) GetDocuments(limit, offset int) ([]schemas.DocumentResponse, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]schemas.DocumentResponse), args.Error(1)
}

func (m *MockDocumentRepo) Get_document_by_id(documentId, createdById uuid.UUID, role string) (string, interface{}, error) {
	args := m.Called(documentId, createdById, role)
	return args.String(0), args.Get(1), args.Error(2)
}

func (m *MockDocumentRepo) Get_document_by_status(status string, limit, offset int) ([]schemas.DocumentResponse, error) {
	args := m.Called(status, limit, offset)
	return args.Get(0).([]schemas.DocumentResponse), args.Error(1)
}

func (m *MockDocumentRepo) Get_documents_by_type(limit, offset int) ([]schemas.DocumentResponse, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]schemas.DocumentResponse), args.Error(1)
}

func (m *MockDocumentRepo) Get_document_by_group(groupId uuid.UUID, limit, offset int) ([]schemas.DocumentResponse, error) {
	args := m.Called(groupId, limit, offset)
	return args.Get(0).([]schemas.DocumentResponse), args.Error(1)
}

func (m *MockDocumentRepo) Get_document_by_standard(standardId, createdById uuid.UUID, limit, offset int, role string) ([]schemas.DocumentResponse, error) {
	args := m.Called(standardId, createdById, limit, offset, role)
	return args.Get(0).([]schemas.DocumentResponse), args.Error(1)
}

func (m *MockDocumentRepo) Get_document_by_service(serviceId, createdById uuid.UUID, limit, offset int, role string) ([]schemas.DocumentResponse, error) {
	args := m.Called(serviceId, createdById, limit, offset, role)
	return args.Get(0).([]schemas.DocumentResponse), args.Error(1)
}

func (m *MockDocumentRepo) Get_document_by_assessment(assessmentId, createdById uuid.UUID, limit, offset int, role string) ([]schemas.DocumentResponse, error) {
	args := m.Called(assessmentId, createdById, limit, offset, role)
	return args.Get(0).([]schemas.DocumentResponse), args.Error(1)
}

func (m *MockDocumentRepo) Get_document_by_createdBy(createdById uuid.UUID, limit, offset int) ([]schemas.DocumentResponse, error) {
	args := m.Called(createdById, limit, offset)
	return args.Get(0).([]schemas.DocumentResponse), args.Error(1)
}

func (m *MockDocumentRepo) Check_document_name(fileName string, header *multipart.FileHeader) (bool, error) {
	args := m.Called(fileName, header)
	return args.Bool(0), args.Error(1)
}

func (m *MockDocumentRepo) Is_public_document(documentTypeId uuid.UUID) (bool, error) {
	args := m.Called(documentTypeId)
	return args.Bool(0), args.Error(1)
}

func (m *MockDocumentRepo) Create_document(
	assessmentId, documentTypeId, createdById, standardId, serviceId uuid.UUID,
	fileName, filepath, fileHash, objectId, role string,
	isPublic bool,
) (string, bool, bool, error) {
	args := m.Called(assessmentId, documentTypeId, createdById, standardId, serviceId, fileName, filepath, fileHash, objectId, role, isPublic)
	return args.String(0), args.Bool(1), args.Bool(2), args.Error(3)
}

func (m *MockDocumentRepo) Get_admin_email() (string, uuid.UUID, error) {
	args := m.Called()
	return args.String(0), args.Get(1).(uuid.UUID), args.Error(2)
}

func (m *MockDocumentRepo) Document_is_approved(documentId uuid.UUID) (bool, error) {
	args := m.Called(documentId)
	return args.Bool(0), args.Error(1)
}

func (m *MockDocumentRepo) Check_document_owner(documentId, createdById uuid.UUID) (bool, error) {
	args := m.Called(documentId, createdById)
	return args.Bool(0), args.Error(1)
}

func (m *MockDocumentRepo) Get_document_filePath(documentId uuid.UUID) (string, error) {
	args := m.Called(documentId)
	return args.String(0), args.Error(1)
}

func (m *MockDocumentRepo) Get_object_id(documentId uuid.UUID) (string, error) {
	args := m.Called(documentId)
	return args.String(0), args.Error(1)
}

func (m *MockDocumentRepo) Update_document(documentId, createdById uuid.UUID, fileName, filePath, description, hash string) (int64, error) {
	args := m.Called(documentId, createdById, fileName, filePath, description, hash)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDocumentRepo) Delete_document(documentId, createdById uuid.UUID, userRole string) (string, bool, error) {
	args := m.Called(documentId, createdById, userRole)
	return args.String(0), args.Bool(1), args.Error(2)
}

func (m *MockDocumentRepo) Check_document_hash(documentId uuid.UUID) (string, error) {
	args := m.Called(documentId)
	return args.String(0), args.Error(1)
}

func (m *MockDocumentRepo) Approval_document(documentId uuid.UUID, status, signedDocPath string, file multipart.File) (int64, error) {
	args := m.Called(documentId, status, signedDocPath, file)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockDocumentRepo) Get_document_owner_email(documentId uuid.UUID) (string, uuid.UUID, error) {
	args := m.Called(documentId)
	return args.String(0), args.Get(1).(uuid.UUID), args.Error(2)
}

func (m *MockDocumentRepo) Get_stats(createdById uuid.UUID, role string) (interface{}, schemas.StatusStat, error) {
	args := m.Called(createdById, role)
	return args.Get(0), args.Get(1).(schemas.StatusStat), args.Error(2)
}

// ─────────────────────────────────────────────
// Mock: AuditRepo
// ─────────────────────────────────────────────

type MockAuditRepo struct {
	mock.Mock
}

func (m *MockAuditRepo) SaveAudit(action, desc string, userId, documentId uuid.UUID, source string, createdAt, updatedAt time.Time) (string, error) {
	args := m.Called(action, desc, userId, documentId, source, createdAt, updatedAt)
	return args.String(0), args.Error(1)
}

// ─────────────────────────────────────────────
// Mock: StorageRepo
// ─────────────────────────────────────────────

type MockStorageRepo struct {
	mock.Mock
}

func (m *MockStorageRepo) GetMinioObject(ctx context.Context, filePath string) (*minio.Object, *minio.ObjectInfo, error) {
	args := m.Called(ctx, filePath)
	obj, _ := args.Get(0).(*minio.Object)
	info, _ := args.Get(1).(*minio.ObjectInfo)
	return obj, info, args.Error(2)
}

func (m *MockStorageRepo) Upload_document(reader *bytes.Reader, objectId string, isPublic bool, size int64, contentType string) (string, error) {
	args := m.Called(reader, objectId, isPublic, size, contentType)
	return args.String(0), args.Error(1)
}

func (m *MockStorageRepo) Delete_document(filePath string) error {
	args := m.Called(filePath)
	return args.Error(0)
}

func (m *MockStorageRepo) Update_document(file multipart.File, header *multipart.FileHeader, oldPath, fileName, objectId string) (string, string, string, error) {
	args := m.Called(file, header, oldPath, fileName, objectId)
	return args.String(0), args.String(1), args.String(2), args.Error(3)
}

func (m *MockStorageRepo) Update_documentFile(file multipart.File, header *multipart.FileHeader, oldPath, objectId string) (string, string, string, error) {
	args := m.Called(file, header, oldPath, objectId)
	return args.String(0), args.String(1), args.String(2), args.Error(3)
}

func (m *MockStorageRepo) Update_documentName(oldPath, fileName string) (string, string, error) {
	args := m.Called(oldPath, fileName)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockStorageRepo) GenerateObjectHMAC(objectId, filePath string) (string, error) {
	args := m.Called(objectId, filePath)
	return args.String(0), args.Error(1)
}

// ─────────────────────────────────────────────
// Mock: NotificationService
// ─────────────────────────────────────────────

type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) NotifyDeptHead(email, filename string) {
	m.Called(email, filename)
}

func (m *MockNotificationService) NotifyOwner(email, filename, msg string) {
	m.Called(email, filename, msg)
}

// ─────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────

func makeFileHeader(filename, contentType string, size int64) *multipart.FileHeader {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Type", contentType)
	return &multipart.FileHeader{
		Filename: filename,
		Header:   h,
		Size:     size,
	}
}

// ─────────────────────────────────────────────
// pageLimit tests (pure function — no mocks needed)
// ─────────────────────────────────────────────

// pageLimit is unexported; test via Get_Document behaviour or expose it.
// Here we test its effect through Get_Document.

// ─────────────────────────────────────────────
// Get_Document
// ─────────────────────────────────────────────

func TestGet_Document_Success(t *testing.T) {
	repoMock := new(MockDocumentRepo)
	auditMock := new(MockAuditRepo)
	storageMock := new(MockStorageRepo)

	expected := []schemas.DocumentResponse{{Filename: "test.pdf"}}
	// page=1, limit=5 → offset=0
	repoMock.On("GetDocuments", 5, 0).Return(expected, nil)

	svc := NewDocumentService(repoMock, auditMock, storageMock, nil)
	docs, err := svc.Get_Document(1, 5)

	assert.NoError(t, err)
	assert.Equal(t, expected, docs)
	repoMock.AssertExpectations(t)
}

func TestGet_Document_RepoError(t *testing.T) {
	repoMock := new(MockDocumentRepo)
	auditMock := new(MockAuditRepo)
	storageMock := new(MockStorageRepo)

	repoMock.On("GetDocuments", 10, 0).Return([]schemas.DocumentResponse{}, errors.New("db error"))

	svc := NewDocumentService(repoMock, auditMock, storageMock, nil)
	docs, err := svc.Get_Document(1, 0) // limit=0 → clamped to 10

	assert.Error(t, err)
	assert.Nil(t, docs)
	repoMock.AssertExpectations(t)
}

func TestGet_Document_PageLimitClamping(t *testing.T) {
	repoMock := new(MockDocumentRepo)
	auditMock := new(MockAuditRepo)
	storageMock := new(MockStorageRepo)

	// page=-5 → 1, limit=999 → 10, offset=0
	repoMock.On("GetDocuments", 10, 0).Return([]schemas.DocumentResponse{}, nil)

	svc := NewDocumentService(repoMock, auditMock, storageMock, nil)
	_, err := svc.Get_Document(-5, 999)

	assert.NoError(t, err)
	repoMock.AssertExpectations(t)
}

// ─────────────────────────────────────────────
// Get_document_by_status
// ─────────────────────────────────────────────

func TestGet_document_by_status_Success(t *testing.T) {
	repoMock := new(MockDocumentRepo)
	auditMock := new(MockAuditRepo)
	storageMock := new(MockStorageRepo)

	expected := []schemas.DocumentResponse{{Filename: "doc.pdf"}}
	repoMock.On("Get_document_by_status", "pending", 10, 0).Return(expected, nil)

	svc := NewDocumentService(repoMock, auditMock, storageMock, nil)
	docs, page, limit, err := svc.Get_document_by_status("pending", 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, expected, docs)
	assert.Equal(t, 1, page)
	assert.Equal(t, 10, limit)
	repoMock.AssertExpectations(t)
}

func TestGet_document_by_status_RepoError(t *testing.T) {
	repoMock := new(MockDocumentRepo)
	auditMock := new(MockAuditRepo)
	storageMock := new(MockStorageRepo)

	repoMock.On("Get_document_by_status", "approved", 10, 0).Return([]schemas.DocumentResponse{}, errors.New("db error"))

	svc := NewDocumentService(repoMock, auditMock, storageMock, nil)
	docs, _, _, err := svc.Get_document_by_status("approved", 1, 10)

	assert.Error(t, err)
	assert.Empty(t, docs)
	repoMock.AssertExpectations(t)
}

// ─────────────────────────────────────────────
// Delete_document
// ─────────────────────────────────────────────

func TestDelete_document_Success(t *testing.T) {
	repoMock := new(MockDocumentRepo)
	auditMock := new(MockAuditRepo)
	storageMock := new(MockStorageRepo)

	docId := uuid.New()
	userId := uuid.New()

	repoMock.On("Delete_document", docId, userId, "user").Return("path/to/file.pdf", true, nil)
	storageMock.On("Delete_document", "path/to/file.pdf").Return(nil)
	auditMock.On("SaveAudit", "delete", mock.AnythingOfType("string"), userId, docId, "client", mock.Anything, mock.Anything).Return("audit-id", nil)

	svc := NewDocumentService(repoMock, auditMock, storageMock, nil)
	resp, err := svc.Delete_document(docId, userId, "user")

	assert.NoError(t, err)
	assert.True(t, resp.Status)
	repoMock.AssertExpectations(t)
	storageMock.AssertExpectations(t)
	auditMock.AssertExpectations(t)
}

func TestDelete_document_NotFound(t *testing.T) {
	repoMock := new(MockDocumentRepo)
	auditMock := new(MockAuditRepo)
	storageMock := new(MockStorageRepo)

	docId := uuid.New()
	userId := uuid.New()

	repoMock.On("Delete_document", docId, userId, "user").Return("", false, nil)

	svc := NewDocumentService(repoMock, auditMock, storageMock, nil)
	resp, err := svc.Delete_document(docId, userId, "user")

	assert.NoError(t, err)
	assert.False(t, resp.Status)
	assert.Equal(t, "Document Not found", resp.Message)
	storageMock.AssertNotCalled(t, "Delete_document", mock.Anything)
}

func TestDelete_document_RepoError(t *testing.T) {
	repoMock := new(MockDocumentRepo)
	auditMock := new(MockAuditRepo)
	storageMock := new(MockStorageRepo)

	docId := uuid.New()
	userId := uuid.New()

	repoMock.On("Delete_document", docId, userId, "user").Return("", false, errors.New("db error"))

	svc := NewDocumentService(repoMock, auditMock, storageMock, nil)
	resp, err := svc.Delete_document(docId, userId, "user")

	assert.Error(t, err)
	assert.False(t, resp.Status)
}

// ─────────────────────────────────────────────
// Update_document
// ─────────────────────────────────────────────

func TestUpdate_document_ApprovedDocument_CannotEdit(t *testing.T) {
	repoMock := new(MockDocumentRepo)
	auditMock := new(MockAuditRepo)
	storageMock := new(MockStorageRepo)

	docId := uuid.New()
	userId := uuid.New()

	repoMock.On("Document_is_approved", docId).Return(true, nil)

	svc := NewDocumentService(repoMock, auditMock, storageMock, nil)
	req := schemas.UpdateRequest{DocumentId: docId.String(), FileName: "new.pdf"}
	resp, err := svc.Update_document(req, nil, nil, userId, "user")

	assert.NoError(t, err)
	assert.False(t, resp.Status)
	assert.Equal(t, "Cannot Edit Approved Document", resp.Message)
	repoMock.AssertNotCalled(t, "Update_document", mock.Anything)
}

func TestUpdate_document_Unauthorized(t *testing.T) {
	repoMock := new(MockDocumentRepo)
	auditMock := new(MockAuditRepo)
	storageMock := new(MockStorageRepo)

	docId := uuid.New()
	userId := uuid.New()

	repoMock.On("Document_is_approved", docId).Return(false, nil)
	repoMock.On("Check_document_owner", docId, userId).Return(false, nil)

	svc := NewDocumentService(repoMock, auditMock, storageMock, nil)
	req := schemas.UpdateRequest{DocumentId: docId.String(), FileName: "new.pdf"}
	resp, err := svc.Update_document(req, nil, nil, userId, "user") // not master-admin

	assert.NoError(t, err)
	assert.False(t, resp.Status)
	assert.Equal(t, "Unauthorized: cannot edit this document", resp.Message)
}

func TestUpdate_document_MasterAdmin_SkipsOwnerCheck(t *testing.T) {
	repoMock := new(MockDocumentRepo)
	auditMock := new(MockAuditRepo)
	storageMock := new(MockStorageRepo)

	docId := uuid.New()
	userId := uuid.New()

	repoMock.On("Document_is_approved", docId).Return(false, nil)
	repoMock.On("Get_document_filePath", docId).Return("old/path.pdf", nil)
	repoMock.On("Get_object_id", docId).Return("obj-id", nil)
	repoMock.On("Update_document", docId, userId, "", "", "Some description", "").Return(int64(1), nil)
	auditMock.On("SaveAudit", "edit", mock.AnythingOfType("string"), userId, docId, "client", mock.Anything, mock.Anything).Return("audit-id", nil)

	svc := NewDocumentService(repoMock, auditMock, storageMock, nil)
	req := schemas.UpdateRequest{DocumentId: docId.String(), Description: "Some description"}
	resp, err := svc.Update_document(req, nil, nil, userId, "master-admin")

	assert.NoError(t, err)
	assert.True(t, resp.Status)
	repoMock.AssertNotCalled(t, "Check_document_owner", mock.Anything, mock.Anything)
}

func TestUpdate_document_DocumentNotFound(t *testing.T) {
	repoMock := new(MockDocumentRepo)
	auditMock := new(MockAuditRepo)
	storageMock := new(MockStorageRepo)

	docId := uuid.New()
	userId := uuid.New()

	repoMock.On("Document_is_approved", docId).Return(false, nil)
	repoMock.On("Get_document_filePath", docId).Return("old/path.pdf", nil)
	repoMock.On("Get_object_id", docId).Return("obj-id", nil)
	repoMock.On("Update_document", docId, userId, "", "", "", "").Return(int64(0), nil)

	svc := NewDocumentService(repoMock, auditMock, storageMock, nil)
	req := schemas.UpdateRequest{DocumentId: docId.String()}
	resp, err := svc.Update_document(req, nil, nil, userId, "master-admin")

	assert.NoError(t, err)
	assert.False(t, resp.Status)
	assert.Equal(t, "Document not Found", resp.Message)
}

// ─────────────────────────────────────────────
// Approval_document
// ─────────────────────────────────────────────

func TestApproval_document_IntegrityCheckFail(t *testing.T) {
	repoMock := new(MockDocumentRepo)
	auditMock := new(MockAuditRepo)
	storageMock := new(MockStorageRepo)

	docId := uuid.New()
	userId := uuid.New()

	repoMock.On("Check_document_hash", docId).Return("stored-hash", nil)
	repoMock.On("Get_object_id", docId).Return("obj-id", nil)
	repoMock.On("Get_document_filePath", docId).Return("path/file.pdf", nil)
	storageMock.On("GenerateObjectHMAC", "obj-id", "path/file.pdf").Return("different-hash", nil)

	svc := NewDocumentService(repoMock, auditMock, storageMock, nil)
	resp, err := svc.Approval_document(docId, userId, "approved", nil, nil)

	assert.NoError(t, err)
	assert.False(t, resp.Status)
	assert.Contains(t, resp.Message, "integrity check failed")
}

func TestApproval_document_AlreadyApproved(t *testing.T) {
	repoMock := new(MockDocumentRepo)
	auditMock := new(MockAuditRepo)
	storageMock := new(MockStorageRepo)

	docId := uuid.New()
	userId := uuid.New()
	sameHash := "abc123"

	repoMock.On("Check_document_hash", docId).Return(sameHash, nil)
	repoMock.On("Get_object_id", docId).Return("obj-id", nil)
	repoMock.On("Get_document_filePath", docId).Return("path/file.pdf", nil)
	storageMock.On("GenerateObjectHMAC", "obj-id", "path/file.pdf").Return(sameHash, nil)
	repoMock.On("Document_is_approved", docId).Return(true, nil)

	svc := NewDocumentService(repoMock, auditMock, storageMock, nil)
	resp, err := svc.Approval_document(docId, userId, "approved", nil, nil)

	assert.NoError(t, err)
	assert.False(t, resp.Status)
	assert.Equal(t, "Cannot change status for approved document", resp.Message)
}

func TestApproval_document_Reject_Success(t *testing.T) {
	repoMock := new(MockDocumentRepo)
	auditMock := new(MockAuditRepo)
	storageMock := new(MockStorageRepo)
	notifMock := new(MockNotificationService)

	docId := uuid.New()
	userId := uuid.New()
	ownerId := uuid.New()
	sameHash := "abc123"

	repoMock.On("Check_document_hash", docId).Return(sameHash, nil)
	repoMock.On("Get_object_id", docId).Return("obj-id", nil)
	repoMock.On("Get_document_filePath", docId).Return("path/file.pdf", nil)
	storageMock.On("GenerateObjectHMAC", "obj-id", "path/file.pdf").Return(sameHash, nil)
	repoMock.On("Document_is_approved", docId).Return(false, nil)
	repoMock.On("Approval_document", docId, "rejected", "", nil).Return(int64(1), nil)
	auditMock.On("SaveAudit", "update", mock.Anything, userId, docId, "client", mock.Anything, mock.Anything).Return("doc.pdf", nil)
	repoMock.On("Get_document_owner_email", docId).Return("owner@mail.com", ownerId, nil)
	notifMock.On("NotifyOwner", "owner@mail.com", "doc.pdf", mock.AnythingOfType("string")).Return()

	svc := NewDocumentService(repoMock, auditMock, storageMock, notifMock)
	resp, err := svc.Approval_document(docId, userId, "rejected", nil, nil)

	// Give goroutines time to run
	time.Sleep(50 * time.Millisecond)

	assert.NoError(t, err)
	assert.True(t, resp.Status)
	assert.Equal(t, "Success updating document status", resp.Message)
}

// ─────────────────────────────────────────────
// GetStats
// ─────────────────────────────────────────────

func TestGetStats_Success(t *testing.T) {
	repoMock := new(MockDocumentRepo)
	auditMock := new(MockAuditRepo)
	storageMock := new(MockStorageRepo)

	userId := uuid.New()
	stat := schemas.StatsSummary{Approved: 5, Pending: 3, Rejected: 2}
	repoMock.On("Get_stats", userId, "user").Return(nil, stat, nil)

	svc := NewDocumentService(repoMock, auditMock, storageMock, nil)
	result, err := svc.GetStats(userId, "user")

	assert.NoError(t, err)
	assert.True(t, result.Status)
	assert.Equal(t, 10, result.Total) // 5+3+2
	assert.Equal(t, stat, result.Stats)
}

func TestGetStats_RepoError(t *testing.T) {
	repoMock := new(MockDocumentRepo)
	auditMock := new(MockAuditRepo)
	storageMock := new(MockStorageRepo)

	userId := uuid.New()
	repoMock.On("Get_stats", userId, "user").Return(nil, schemas.StatsSummary{}, errors.New("db error"))

	svc := NewDocumentService(repoMock, auditMock, storageMock, nil)
	result, err := svc.GetStats(userId, "user")

	assert.Error(t, err)
	assert.False(t, result.Status)
}

// ─────────────────────────────────────────────
// Get_public_document_by_id
// ─────────────────────────────────────────────

func TestGet_public_document_by_id_Success(t *testing.T) {
	repoMock := new(MockDocumentRepo)
	auditMock := new(MockAuditRepo)
	storageMock := new(MockStorageRepo)

	docId := uuid.New()
	repoMock.On("Get_document_by_id", docId, uuid.Nil, "public").Return("public/doc.pdf", nil, nil)

	svc := NewDocumentService(repoMock, auditMock, storageMock, nil)
	path, err := svc.Get_public_document_by_id(docId, "public")

	assert.NoError(t, err)
	assert.Equal(t, "public/doc.pdf", path)
}

func TestGet_public_document_by_id_NotFound(t *testing.T) {
	repoMock := new(MockDocumentRepo)
	auditMock := new(MockAuditRepo)
	storageMock := new(MockStorageRepo)

	docId := uuid.New()
	repoMock.On("Get_document_by_id", docId, uuid.Nil, "public").Return("", nil, errors.New("not found"))

	svc := NewDocumentService(repoMock, auditMock, storageMock, nil)
	path, err := svc.Get_public_document_by_id(docId, "public")

	assert.Error(t, err)
	assert.Empty(t, path)
}
