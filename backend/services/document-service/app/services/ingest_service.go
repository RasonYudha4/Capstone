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

	kelompok, err := s.repo.Get_group_name_by_serviceid(req.ServicesId)
	if err != nil {
		log.Printf("[ingest] could not resolve kelompok for service %s: %v", req.ServicesId, err)
		kelompok = ""
	}

	fungsiPelayanan, err := s.repo.Get_service_name_byid(req.ServicesId)
	if err != nil {
		log.Printf("[ingest] could not resolve fungsi_pelayanan for service %s: %v", req.ServicesId, err)
		fungsiPelayanan = ""
	}

	standar, standarCode, err := s.repo.Get_standard_name_byid(req.StandardId)
	if err != nil {
		log.Printf("[ingest] could not resolve standar for standard %s: %v", req.StandardId, err)
		standar, standarCode = "", ""
	}

	elementPenilaian, elementPenilaianCode, err := s.repo.Get_assessment_name_byid(req.AssessmentId)
	if err != nil {
		log.Printf("[ingest] could not resolve element_penilaian for assessment %s: %v", req.AssessmentId, err)
		elementPenilaian, elementPenilaianCode = "", ""
	}

	docType, err := s.repo.Get_document_type_byid(req.DocumentTypeId)
	if err != nil {
		log.Printf("[ingest] could not resolve doc_type for document_type %s: %v", req.DocumentTypeId, err)
		docType = ""
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
		"kelompok":          		kelompok,
		"fungsi_pelayanan":  		fungsiPelayanan,
		"standar":           		standar,
		"standar_code":      		standarCode,
		"element_penilaian": 		elementPenilaian,
		"element_penilaian_code": 	elementPenilaianCode,
		"doc_type":          		docType,
		"nama_berkas":       		req.FileName,
		"deskripsi":         		req.Description,
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
		log.Printf("[ingest] status %d: %s", resp.StatusCode, body)
		return
	}

	log.Printf("[ingest] evidence ingested OK — kelompok=%s fungsi=%s standar=%s ep=%s",
		kelompok, fungsiPelayanan, standar, elementPenilaian)
}