from fastapi import APIRouter
from fastapi.responses import StreamingResponse

from app.schemas.query import QueryRequest          # was app.models.query
from app.services.rag_service import run_query, run_query_stream 

router = APIRouter(prefix="/query", tags=["query"])

@router.post("/")
def query(req: QueryRequest):
    answer = run_query(req.question)
    return {"answer": answer}

@router.post("/stream")
def query_stream(req: QueryRequest):
    return StreamingResponse(
        run_query_stream(req.question),
        media_type="text/plain"
    )