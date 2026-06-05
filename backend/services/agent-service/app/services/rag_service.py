"""
pipeline/rag_pipeline.py — orchestrates retrieval and generation at query time.

Features:
- embed_query() now takes (text, embedder: EmbedderModel) — no more URL/model args.
  The embedder is a module-level singleton loaded once at startup.
- generate() / generate_stream() are now local OV calls, not HTTP requests.
  Their signatures are unchanged so this file barely moves.
"""
from __future__ import annotations

import time
from typing import Generator

from app.core.config import settings
from app.repositories.generation.generator import generate, generate_stream
from app.repositories.generation.intent_extractor import extract_intent
from app.repositories.generation.prompt_builder import build_prompt, build_gap_prompt
from app.repositories.ingestion.embedder import EmbedderModel, embed_query
from app.core.logger import get_logger
from app.repositories.retrieval.retriever import retrieve
from app.core.store.chromadb import ChromaStore

log = get_logger("rag_pipeline")

# ── Embedder singleton — loaded once when the pipeline module is first imported
# This mirrors the generator's lru_cache pattern: pay the cold-start cost once,
# not on every query.
_embedder = EmbedderModel.from_pretrained(
    settings.embed_model_path,
    device=settings.embed_device,
)


def run_query(question: str) -> str:
    log.info("=== query start: '%s' ===", question[:80])
    t_start = time.perf_counter()

    # ── Extract intent ───────────────────────────────────────────────────────
    intent = extract_intent(question)
    log.info(
        "intent: query_type=%s standar_id=%s bab_code=%s",
        intent["query_type"],
        intent.get("standar_id"),
        intent.get("bab_code"),
    )

    # ── Gap analysis — pure metadata, no vector search ───────────────────────
    if intent["query_type"] == "gap_analysis":
        answer = _run_gap_analysis(intent)
        log.info(
            "=== gap analysis done in %.1fms ===",
            (time.perf_counter() - t_start) * 1000,
        )
        return answer

    # ── Build filters from intent ────────────────────────────────────────────
    filters = _build_filters(intent)

    # ── Embed → Retrieve → Generate ──────────────────────────────────────────
    # embed_query now takes (text, embedder) — no URL or model string needed.
    q_vec   = embed_query(question, _embedder)
    results = retrieve(q_vec, top_k=settings.top_k, filters=filters)
    prompt  = build_prompt(question, results, intent)

    log.info("prompt built — %d chunks, %d chars", len(results), len(prompt))

    answer = generate(prompt)

    log.info("=== query done in %.1fms ===", (time.perf_counter() - t_start) * 1000)
    return answer


def run_query_stream(question: str) -> Generator[str, None, None]:
    log.info("=== stream query start: '%s' ===", question[:80])

    intent  = extract_intent(question)
    filters = _build_filters(intent)

    q_vec   = embed_query(question, _embedder)
    results = retrieve(q_vec, top_k=settings.top_k, filters=filters)
    prompt  = build_prompt(question, results, intent)

    log.info("prompt built — %d chunks, %d chars", len(results), len(prompt))

    yield from generate_stream(prompt)


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _build_filters(intent: dict) -> dict | None:
    filters: dict = {}

    if intent["query_type"] == "requirement_lookup":
        filters["is_kmk"] = True
    elif intent["query_type"] == "evidence_check":
        filters["is_kmk"] = False

    if intent.get("standar_id"):
        filters["standar_id"] = intent["standar_id"]
    elif intent.get("bab_code"):
        filters["bab_code"] = intent["bab_code"]

    return filters or None


def _run_gap_analysis(intent: dict) -> str:
    store = ChromaStore()

    extra = {"bab_code": intent["bab_code"]} if intent.get("bab_code") else {}

    all_ep     = set(store.get_all_unique_values("standar_id", {"is_kmk": True,  **extra}))
    covered_ep = set(store.get_all_unique_values("standar_id", {"is_kmk": False, **extra}))
    missing    = sorted(all_ep - covered_ep)

    prompt = build_gap_prompt(missing, covered_ep, intent)
    return generate(prompt)