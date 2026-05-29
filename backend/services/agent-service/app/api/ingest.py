import os
import shutil

from fastapi import APIRouter, File, Form, UploadFile
from app.services.ingest_service import run_ingest_kmk, run_ingest_evidence

router = APIRouter(prefix="/ingest", tags=["ingest"])


@router.post("/kmk")
def ingest_kmk(file: UploadFile = File(...)):
    kmk_dir = "./data/kmk"
    os.makedirs(kmk_dir, exist_ok=True)

    file_path = os.path.join(kmk_dir, file.filename)
    with open(file_path, "wb") as f:
        shutil.copyfileobj(file.file, f)

    result = run_ingest_kmk(file_path)
    return {
        "status":          "ok" if result.success else "failed",
        "filename":        file.filename,
        "chunks_upserted": result.chunks_upserted,
    }


@router.post("/evidence")
def ingest_evidence(
    file:             UploadFile = File(...),
    kelompok:         str = Form(...),
    fungsi_pelayanan: str = Form(...),
    standar_id:       str = Form(...),
    ep_id:            str = Form(...),
    doc_type:         str = Form(...),
    nama_berkas:      str = Form(...),
    deskripsi:        str = Form(""),
):
    upload_dir = "./data/uploads"
    os.makedirs(upload_dir, exist_ok=True)

    file_path = os.path.join(upload_dir, file.filename)
    with open(file_path, "wb") as f:
        shutil.copyfileobj(file.file, f)

    form_metadata = {
        "kelompok":         kelompok,
        "fungsi_pelayanan": fungsi_pelayanan,
        "standar_id":       standar_id,
        "ep_id":            ep_id,
        "doc_type":         doc_type,
        "nama_berkas":      nama_berkas,
        "deskripsi":        deskripsi,
    }

    result = run_ingest_evidence(file_path, form_metadata)
    return {
        "status":          "ok" if result.success else "failed",
        "filename":        file.filename,
        "chunks_upserted": result.chunks_upserted,
        "ep_id":           ep_id,
        "standar_id":      standar_id,
    }