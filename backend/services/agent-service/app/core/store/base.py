"""
store/base.py — abstract interface for all vector store backends.

never depends on a specific backend. Swap backends by changing one import.
"""
from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass

from app.models import Chunk


@dataclass
class SearchResult:
    """Single result returned from a vector search."""
    text:       str
    source:     str
    score:      float
    is_kmk:     bool = False
    doc_type: str | None = None
    chunk_type: str | None = None
    bab_code:   str | None = None
    standar: str | None = None
    element_penilaian:      str | None = None
    kelompok:   str | None = None
    page_ref:   int | None = None
    is_signed:  str | None = None
    nama_berkas:    str | None = None
    document_id:    str | None = None
    service_id:     str | None = None
    standard_id:    str | None = None
    assessment_id:  str | None = None

    @classmethod
    def from_payload(cls, payload: dict, score: float) -> "SearchResult":
        return cls(
            text       = payload.get("text", ""),
            source     = payload.get("source", ""),
            score      = score,
            is_kmk     = bool(payload.get("is_kmk", False)),
            doc_type   = payload.get("doc_type"),
            chunk_type = payload.get("chunk_type"),
            bab_code   = payload.get("bab_code"),
            standar = payload.get("standar"),
            element_penilaian      = payload.get("element_penilaian"),
            kelompok   = payload.get("kelompok"),
            page_ref   = payload.get("page_ref"),
            is_signed  = payload.get("is_signed"),
            nama_berkas    = payload.get("nama_berkas"),
            document_id    = payload.get("document_id"),
            service_id     = payload.get("service_id"),
            standard_id    = payload.get("standard_id"),
            assessment_id  = payload.get("assessment_id"),
        )


class VectorStore(ABC):
    """
    Abstract base for all vector store backends.
    Both ingestion (upsert) and retrieval (query) use this interface.
    """

    @abstractmethod
    def upsert(self, chunks: list[Chunk]) -> None:
        """
        Write chunks to the store. Must be idempotent —
        upserting the same chunk_id twice updates, never duplicates.
        Chunks must have .vector populated before calling.
        """
        ...

    @abstractmethod
    def query(
        self,
        embedding:  list[float],
        top_k:      int = 5,
        filters:    dict | None = None,
    ) -> list[SearchResult]:
        """
        Search for nearest neighbours.

        Args:
            embedding:  query vector, same dim as indexed vectors
            top_k:      number of results to return
            filters:    optional metadata filters, e.g.
                        {"bab_code": "TKRS", "chunk_type": "ep_unit"}
                        Implementation behaviour varies by backend —
                        ChromaStore supports this natively, FaissStore
                        applies post-hoc filtering.
        """
        ...

    @abstractmethod
    def count(self) -> int:
        """Return total number of indexed chunks."""
        ...

    @abstractmethod
    def delete_collection(self) -> None:
        """Drop and recreate the collection. Used by reindex.py."""
        ...