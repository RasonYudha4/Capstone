package services

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"capstone/app/schemas"
)

func (s *DocumentService) TriggerIngestEvidence(file io.Reader, fileName string, req schemas.DocumentRequest) {
	agentURL := os.Getenv("AGENT_SERVICE_URL")
	if agentURL == "" {
		log.Println("[ingest] AGENT_SERVICE_URL not set, skipping evidence ingest")
		return
	}

	kelompok, err := s.repo.GetGroupNameByServiceId(req.ServicesId)
	if err != nil {
		log.Printf("[ingest] could not resolve kelompok for service %s: %v", req.ServicesId, err)
		kelompok = ""
	}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	fw, err := mw.CreateFormFile("file", fileName)
	if err != nil {
		log.Printf("[ingest] multipart file field error: %v", err)
		return
	}
	if _, err = io.Copy(fw, file); err != nil {
		log.Printf("[ingest] file copy error: %v", err)
		return
	}

	fields := map[string]string{
		"kelompok":         kelompok,
		"fungsi_pelayanan": req.ServicesId,
		"standar_id":       req.StandardId,
		"ep_id":            req.AssessmentId,
		"doc_type":         req.DocumentTypeId,
		"nama_berkas":      req.FileName,
		"deskripsi":        req.Description,
	}
	for k, v := range fields {
		if err := mw.WriteField(k, v); err != nil {
			log.Printf("[ingest] write field %s error: %v", k, err)
			return
		}
	}
	mw.Close()

	resp, err := http.Post(
		fmt.Sprintf("%s/ingest/evidence", agentURL),
		mw.FormDataContentType(),
		&buf,
	)
	if err != nil {
		log.Printf("[ingest] POST /ingest/evidence failed: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[ingest] unexpected status %d: %s", resp.StatusCode, body)
		return
	}

	log.Printf("[ingest] evidence ingested OK for ep=%s standar=%s", req.AssessmentId, req.StandardId)
}