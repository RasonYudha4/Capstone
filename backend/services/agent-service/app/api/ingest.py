import os
import shutil

from fastapi import APIRouter, File, UploadFile
from app.services.ingest_service import run_ingest

router = APIRouter(prefix="/ingest", tags=["ingest"])

@router.post("/")
def ingest(file: UploadFile = File(...)):
    upload_dir = "./docs"
    os.makedirs(upload_dir, exist_ok=True)

    file_path = os.path.join(upload_dir, file.filename)
    with open(file_path, "wb") as f:
        shutil.copyfileobj(file.file, f)

    run_ingest(upload_dir)
    return {"status": "ok", "filename": file.filename}