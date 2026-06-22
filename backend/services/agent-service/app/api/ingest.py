import os
from pathlib import Path
import shutil
import tempfile

from fastapi import APIRouter, File, Form, HTTPException, UploadFile

from app.schemas.response import ApiResponse
from app.services.ingest_service import run_ingest_kmk, run_ingest_evidence

router = APIRouter(prefix="/ingest", tags=["ingest"])


@router.post("/kmk", response_model=ApiResponse)
def ingest_kmk():
    kmk_dir = "./kmk"
    if not os.path.isdir(kmk_dir):
        raise HTTPException(status_code=404, detail="KMK folder not found.")

    pdf_files = [f for f in os.listdir(kmk_dir) if f.lower().endswith(".pdf")]
    if not pdf_files:
        raise HTTPException(status_code=404, detail="No PDF found in ./kmk folder.")

    file_path = os.path.join(kmk_dir, pdf_files[0])
    result = run_ingest_kmk(file_path)

    if not result.success:
        raise HTTPException(
            status_code=422,
            detail=ApiResponse.error(422, "KMK ingestion failed").model_dump(),
        )

    return ApiResponse.created(
        data={"filename": pdf_files[0], "chunks_upserted": result.chunks_upserted},
        message="KMK document ingested successfully",
    )


@router.post("/evidence", response_model=ApiResponse)
def ingest_evidence(
    file:              UploadFile = File(...),
    kelompok:               str = Form(...),
    fungsi_pelayanan:       str = Form(...),
    standar:                str = Form(...),
    standar_code:           str = Form(...),
    element_penilaian:      str = Form(...),
    element_penilaian_code: str = Form(...),
    doc_type:               str = Form(...),
    nama_berkas:            str = Form(...),
    deskripsi:              str = Form(""),
    service_id:             str = Form(...),
    standard_id:            str = Form(...),
    assessment_id:          str = Form(...),
    document_id:            str = Form(...),
):
    suffix = Path(file.filename).suffix
    try:
        with tempfile.NamedTemporaryFile(delete=False, suffix=suffix) as tmp:
            shutil.copyfileobj(file.file, tmp)
            tmp_path = tmp.name

        form_metadata = {
            "kelompok":                 kelompok,
            "fungsi_pelayanan":         fungsi_pelayanan,
            "standar":                  standar,
            "standar_code":             standar_code,
            "element_penilaian":        element_penilaian,
            "element_penilaian_code":   element_penilaian_code,
            "doc_type":                 doc_type,
            "nama_berkas":              nama_berkas,
            "deskripsi":                deskripsi,
            "service_id":               service_id,
            "standard_id":              standard_id,
            "assessment_id":            assessment_id,
            "document_id":              document_id
        }

        result = run_ingest_evidence(tmp_path, form_metadata)

    finally:
        os.unlink(tmp_path)

    if not result.success:
        raise HTTPException(
            status_code=422,
            detail=ApiResponse.error(422, "Evidence ingestion failed").model_dump(),
        )

    return ApiResponse.created(
        data={
            "filename":          file.filename,
            "chunks_upserted":   result.chunks_upserted,
            "element_penilaian": element_penilaian,
            "standar":           standar,
        },
        message="Evidence document ingested successfully",
    )