from fastapi import APIRouter
from app.api import ingest, query

router = APIRouter()
router.include_router(ingest.router)
router.include_router(query.router)