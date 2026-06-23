"""
retrieval/retriever.py — query-time retrieval logic.
"""
from __future__ import annotations

from app.core.config import settings
from app.core.logger import get_logger, timer
from app.core.store.base import SearchResult
from app.core.store.chromadb import ChromaStore

log = get_logger("retriever")

def retrieve(
    question_embedding: list[float],
    top_k:   int         = settings.top_k,
    filters: dict | None = None,
) -> list[SearchResult]:
    store = ChromaStore()

    with timer(log, f"vector search top_k={top_k}"):
        results = store.query(question_embedding, top_k=top_k, filters=filters)

    log.info("retrieved %d chunk(s)", len(results))

    for i, r in enumerate(results):
        log.info(
            "  [%d] score=%.4f is_kmk=%-5s bab=%-6s standar=%s doc_type=%-15s source=%s",
            i + 1,
            r.score,
            r.is_kmk,
            r.fungsi_pelayanan   or "n/a",
            r.standar or "n/a",
            r.doc_type   or "n/a",
            r.source,
        )

    return results