import os
from pathlib import Path
import shutil
import tempfile

from fastapi import APIRouter, File, Form, HTTPException, UploadFile

from app.schemas.response import ApiResponse
from app.services.ingest_service import run_ingest_kmk, run_ingest_evidence

router = APIRouter(prefix="/ingest", tags=["ingest"])


@router.post("/kmk", response_model=ApiResponse)
def ingest_kmk(file: UploadFile = File(...)):
    kmk_dir = "./data/kmk"
    os.makedirs(kmk_dir, exist_ok=True)

    file_path = os.path.join(kmk_dir, file.filename)
    with open(file_path, "wb") as f:
        shutil.copyfileobj(file.file, f)

    result = run_ingest_kmk(file_path)

    if not result.success:
        raise HTTPException(
            status_code=422,
            detail=ApiResponse.error(422, "KMK ingestion failed").model_dump(),
        )

    return ApiResponse.created(
        data={"filename": file.filename, "chunks_upserted": result.chunks_upserted},
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
):
    suffix = Path(file.filename).suffix
    tmp_path = None
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
        }

        result = run_ingest_evidence(tmp_path, form_metadata)

    finally:
        if tmp_path and os.path.exists(tmp_path):
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