# app/api/query.py
import os
from pathlib import Path
import tempfile

from fastapi import APIRouter, File, Form, UploadFile
from fastapi.responses import StreamingResponse

from app.schemas.query import AppContext, QueryRequest
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
def query_stream(
    question: str = Form(...),
    session_id: str | None = Form(None),
    app_context: str | None = Form(None),  
    file: UploadFile | None = File(None),
):
    session_id = session_id or conversation_store.create_session()
    history = conversation_store.get_history(session_id)
    conversation_store.append_turn(session_id, role="user", content=question)

    parsed_ctx = AppContext.model_validate_json(app_context) if app_context else None

    tmp_path = None
    if file is not None:
        suffix = Path(file.filename).suffix
        with tempfile.NamedTemporaryFile(delete=False, suffix=suffix) as tmp:
            tmp.write(file.file.read())
            tmp_path = tmp.name

    def _stream_and_record():
        chunks: list[str] = []
        try:
            stream = run_query_stream(
                question, history, parsed_ctx,
                file_path=tmp_path,
                content_type=file.content_type if file else None,
            )
            for chunk in stream:
                chunks.append(chunk)
                yield chunk
        finally:
            if tmp_path:
                os.unlink(tmp_path)
            conversation_store.append_turn(session_id, role="assistant", content="".join(chunks))

    response = StreamingResponse(_stream_and_record(), media_type="text/plain")
    response.headers["X-Session-Id"] = session_id
    return response