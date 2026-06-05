"""
ingest_pipeline.py — orchestrates all ingestion stages.

Two public entry points:
- run_ingest_kmk(kmk_path)
    Called once at system setup via POST /ingest/kmk.
    No form metadata — chunk_document() uses KMK path automatically.

- run_ingest_evidence(file_path, form_metadata)
    Called on every evidence upload via POST /ingest.
    form_metadata comes from the upload form fields.
    Deletes stale chunks for the same source before upserting.

Embedder lifecycle
------------------
OVModelForFeatureExtraction is expensive to load (~seconds).
_embedder initialises it exactly once per process and caches the
instance at module level.  Both pipeline functions call _embedder
transparently — callers never manage the model lifetime themselves.
"""
from __future__ import annotations

import time
from dataclasses import dataclass, field
from pathlib import Path

from app.core.config import settings
from app.repositories.ingestion.chunker import chunk_document
from app.repositories.ingestion.embedder import EmbedderModel, EmbeddingError, embed_chunks
from app.repositories.ingestion.enricher import enrich_document
from app.repositories.ingestion.parser import parse_single
from app.models import Chunk, ParsedDocument, IngestResult
from app.core.logger import get_logger, timer
from app.core.store.chromadb import ChromaStore

log = get_logger("ingest_pipeline")


# ---------------------------------------------------------------------------
# Module-level embedder singleton
# ---------------------------------------------------------------------------

_embedder = EmbedderModel.from_pretrained(
    settings.embed_model_path,
    device=settings.embed_device,
)

# ---------------------------------------------------------------------------
# KMK ingestion — called once at system setup
# ---------------------------------------------------------------------------

def run_ingest_kmk(kmk_path: str) -> IngestResult:
    """
    Ingest the KMK standard document.
    No form metadata — chunk_document() detects KMK path via form_metadata=None.
    Safe to re-run: upsert is idempotent on chunk_id.
    """
    t_start = time.perf_counter()
    result  = IngestResult()

    # ── Parse ────────────────────────────────────────────────────────────────
    with timer(log, "parsing KMK"):
        doc = parse_single(kmk_path)

    if not doc:
        log.error("failed to parse KMK at %s", kmk_path)
        result.docs_failed = 1
        result.elapsed_s   = time.perf_counter() - t_start
        return result

    result.docs_found  = 1
    result.docs_parsed = 1

    # ── Chunk ────────────────────────────────────────────────────────────────
    with timer(log, "chunking KMK"):
        chunks = chunk_document(doc)   # form_metadata=None → KMK path

    result.chunks_total = len(chunks)
    log.info("KMK → %d standar chunks", len(chunks))

    # ── Embed ────────────────────────────────────────────────────────────────
    with timer(log, "embedding KMK chunks"):
        try:
            embed_chunks(
                chunks,
                embedder=_embedder,
                batch_size=settings.embed_batch_size,
            )
        except EmbeddingError as exc:
            log.error("embedding failed: %s", exc)
            result.docs_failed = 1
            result.elapsed_s   = time.perf_counter() - t_start
            return result

    # ── Upsert ───────────────────────────────────────────────────────────────
    with timer(log, "upserting KMK chunks"):
        try:
            store = ChromaStore()
            store.upsert(chunks)
            result.chunks_upserted = len(chunks)
        except Exception as exc:
            log.error("upsert failed: %s", exc)
            result.docs_failed = 1
            result.elapsed_s   = time.perf_counter() - t_start
            return result

    result.elapsed_s = time.perf_counter() - t_start
    result.log_summary(log)
    return result


# ---------------------------------------------------------------------------
# Evidence ingestion — called on every upload form submission
# ---------------------------------------------------------------------------

def run_ingest_evidence(
    file_path: str,
    form_metadata: dict,
) -> IngestResult:
    """
    Ingest a single evidence document uploaded via the form.

    form_metadata expected keys:
        kelompok, fungsi_pelayanan, standar_id, ep_id,
        doc_type, nama_berkas, deskripsi

    Stale chunk cleanup: deletes any existing chunks for this source
    before upserting so re-uploads don't leave orphan chunks.
    """
    t_start = time.perf_counter()
    result  = IngestResult()

    # ── Parse ────────────────────────────────────────────────────────────────
    with timer(log, "parsing evidence document"):
        doc = parse_single(file_path)

    if not doc:
        log.error("failed to parse %s", file_path)
        result.docs_failed = 1
        result.elapsed_s   = time.perf_counter() - t_start
        return result

    result.docs_found  = 1
    result.docs_parsed = 1

    # ── Enrich (signature detection) ─────────────────────────────────────────
    enrich_document(doc)

    # ── Chunk ────────────────────────────────────────────────────────────────
    with timer(log, "chunking evidence document"):
        chunks = chunk_document(doc, form_metadata=form_metadata)

    result.chunks_total = len(chunks)
    log.info(
        "%s → %d chunks (ep=%s sig=%s)",
        Path(file_path).name,
        len(chunks),
        form_metadata.get("ep_id"),
        doc.signature_status.value,
    )

    # ── Embed ────────────────────────────────────────────────────────────────
    with timer(log, "embedding evidence chunks"):
        try:
            embed_chunks(
                chunks,
                embedder=_embedder,
                batch_size=settings.embed_batch_size,
            )
        except EmbeddingError as exc:
            log.error("embedding failed: %s", exc)
            result.docs_failed = 1
            result.elapsed_s   = time.perf_counter() - t_start
            return result

    # ── Stale chunk cleanup + Upsert ─────────────────────────────────────────
    with timer(log, "upserting evidence chunks"):
        try:
            store = ChromaStore()
            store.delete_by_source(file_path)   # remove old chunks for this file
            store.upsert(chunks)
            result.chunks_upserted = len(chunks)
        except Exception as exc:
            log.error("upsert failed: %s", exc)
            result.docs_failed = 1
            result.elapsed_s   = time.perf_counter() - t_start
            return result

    result.elapsed_s = time.perf_counter() - t_start
    result.log_summary(log)
    return result