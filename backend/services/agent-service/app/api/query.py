# app/api/query.py
from fastapi import APIRouter
from fastapi.responses import StreamingResponse

from app.schemas.query import QueryRequest
from app.schemas.response import ApiResponse
from app.services.rag_service import run_query, run_query_stream
from app.core.conversation_store import conversation_store

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
    session_id = req.session_id or conversation_store.create_session()
 
    history = conversation_store.get_history(session_id)
    conversation_store.append_turn(session_id, role="user", content=req.question)
 
    def _stream_and_record():
        chunks: list[str] = []
        for chunk in run_query_stream(req.question, history):
            chunks.append(chunk)
            yield chunk
        conversation_store.append_turn(
            session_id, role="assistant", content="".join(chunks)
        )
 
    response = StreamingResponse(_stream_and_record(), media_type="text/plain")
    # Client reads this once on first reply and echoes it back on subsequent
    # requests in the same conversation. Opaque to the client — not constructed
    # or interpreted by it, just stored and replayed.
    response.headers["X-Session-Id"] = session_id
    return response