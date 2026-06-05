# app/api/query.py
from fastapi import APIRouter
from fastapi.responses import StreamingResponse

from app.schemas.query import QueryRequest
from app.schemas.response import ApiResponse
from app.services.rag_service import run_query, run_query_stream

router = APIRouter(prefix="/query", tags=["query"])


@router.post("/", response_model=ApiResponse)
def query(req: QueryRequest):
    answer = run_query(req.question)
    return ApiResponse.ok(
        data={"answer": answer},
        message="Query completed successfully",
    )


@router.post("/stream")
def query_stream(req: QueryRequest):
    return StreamingResponse(
        run_query_stream(req.question),
        media_type="text/plain",
    )