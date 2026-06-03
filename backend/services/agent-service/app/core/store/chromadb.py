"""
store/chromadb.py — ChromaDB backend.

Improvements over retrieval/vectorstore_chromadb.py:
- Accepts list[Chunk] instead of parallel (chunks, embeddings, sources) lists
- Singleton client — one PersistentClient per process, not one per call
- chunk_id used as Chroma document ID — upserts are idempotent on re-runs
- Full Chunk.payload stored as metadata — all fields available for filtering
- query() returns list[SearchResult] typed objects, not raw dicts
- Metadata filter support passed through to Chroma's where clause
- Implements VectorStore ABC — swappable with FaissStore
"""
from __future__ import annotations

import chromadb
from chromadb.config import Settings

from app.core.config import settings
from app.models import Chunk
from app.core.logger import get_logger
from app.core.store.base import SearchResult, VectorStore

log = get_logger("store.chromadb")


class ChromaStore(VectorStore):

    _client: chromadb.PersistentClient | None = None  # module-level singleton

    def __init__(self, path: str = settings.vector_path, collection: str = settings.collection):
        self._collection_name = collection
        self._path = path
        self._col = self._get_collection()

    # ------------------------------------------------------------------
    # Singleton client — one connection per process
    # ------------------------------------------------------------------

    @classmethod
    def _get_client(cls, path: str) -> chromadb.PersistentClient:
        if cls._client is None:
            cls._client = chromadb.PersistentClient(
                path=path,
                settings=Settings(anonymized_telemetry=False),
            )
            log.debug("chroma client initialised at %s", path)
        return cls._client

    def _get_collection(self) -> chromadb.Collection:
        client = self._get_client(self._path)
        return client.get_or_create_collection(
            name=self._collection_name,
            metadata={"hnsw:space": "cosine"},
        )

    # ------------------------------------------------------------------
    # VectorStore interface
    # ------------------------------------------------------------------

    def upsert(self, chunks: list[Chunk]) -> None:
        if not chunks:
            return

        _assert_vectors(chunks)

        ids        = [c.chunk_id for c in chunks]
        documents  = [c.text for c in chunks]
        embeddings = [c.vector for c in chunks]
        metadatas  = [_safe_payload(c.payload) for c in chunks]

        self._col.upsert(
            ids=ids,
            documents=documents,
            embeddings=embeddings,
            metadatas=metadatas,
        )
        log.info("upserted %d chunk(s) to chroma collection '%s'", len(chunks), self._collection_name)

    def query(
        self,
        embedding: list[float],
        top_k:     int = 5,
        filters:   dict | None = None,
    ) -> list[SearchResult]:
        kwargs: dict = dict(
            query_embeddings=[embedding],
            n_results=top_k,
            include=["documents", "metadatas", "distances"],
        )
        if filters:
            kwargs["where"] = _build_chroma_where(filters)

        results = self._col.query(**kwargs)

        docs      = results["documents"][0]
        metas     = results["metadatas"][0]
        distances = results["distances"][0]

        return [
            SearchResult.from_payload(
                payload={**meta, "text": doc},
                score=float(1 - dist),   # cosine distance → similarity
            )
            for doc, meta, dist in zip(docs, metas, distances)
        ]

    def count(self) -> int:
        return self._col.count()

    def delete_collection(self) -> None:
        client = self._get_client(self._path)
        client.delete_collection(self._collection_name)
        self._col = self._get_collection()
        log.warning("collection '%s' dropped and recreated", self._collection_name)

    def delete_by_source(self, source: str) -> None:
        # delete all chunks where metadata source == source
        self._col.delete(where={"source": {"$eq": source}})

    def get_all_unique_values(self, field: str, filters: dict) -> list[str]:
        # used by intent extractor to get known standar_ids and bab_codes
        results = self._col.get(
            where=_build_chroma_where(filters),
            include=["metadatas"]
        )
        return list({m[field] for m in results["metadatas"] if field in m})


# ------------------------------------------------------------------
# Helpers
# ------------------------------------------------------------------

def _safe_payload(payload: dict) -> dict:
    """
    Chroma metadata values must be str | int | float | bool.
    Strip None values — Chroma rejects them.
    """
    return {k: v for k, v in payload.items() if v is not None}


def _build_chroma_where(filters: dict) -> dict:
    """
    Convert a simple {field: value} filter dict to Chroma's where syntax.
    Multiple filters are ANDed together.

    Example:
        {"bab_code": "TKRS", "chunk_type": "ep_unit"}
        → {"$and": [{"bab_code": {"$eq": "TKRS"}}, {"chunk_type": {"$eq": "ep_unit"}}]}
    """
    clauses = [{k: {"$eq": v}} for k, v in filters.items()]
    if len(clauses) == 1:
        return clauses[0]
    return {"$and": clauses}


def _assert_vectors(chunks: list[Chunk]) -> None:
    missing = [c.chunk_index for c in chunks if not c.vector]
    if missing:
        raise ValueError(
            f"chunks missing vectors (indices: {missing[:5]}{'...' if len(missing) > 5 else ''}). "
            "Run embedder.embed_chunks() before upserting."
        )