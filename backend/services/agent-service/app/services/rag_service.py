"""
pipeline/rag_service.py — orchestrates retrieval and generation at query time.
"""
from __future__ import annotations

import time
from typing import Generator

from app.core.config import settings
from app.repositories.generation.generator import generator, generate, generate_stream
from app.repositories.generation.intent_extractor import IntentClassifier, extract_intent
from app.repositories.generation.prompt_builder import build_prompt, build_gap_prompt
from app.repositories.generation.query_rewriter import rewrite_query_with_history
from app.repositories.ingestion.embedder import embedder, embed_query
from app.core.logger import get_logger
from app.repositories.retrieval.retriever import retrieve
from app.core.store.chromadb import ChromaStore, _build_chroma_where

log = get_logger("rag_pipeline")
_classifier = IntentClassifier(embedder)

def run_query(question: str) -> str:
    log.info("=== query start: '%s' ===", question[:80])
    t_start = time.perf_counter()

    intent = extract_intent(question, _classifier)
    log.info(
        "intent: query_type=%s standar=%s bab_code=%s",
        intent["query_type"],
        intent.get("standar"),
        intent.get("bab_code"),
    )

    # ── Routed handlers (no retrieval needed) ────────────────────────────────
    if intent["query_type"] == "gap_analysis":
        answer = _run_gap_analysis(intent)
        log.info("=== gap analysis done in %.1fms ===", (time.perf_counter() - t_start) * 1000)
        return answer

    if intent["query_type"] == "inventory":
        answer = _run_inventory(intent)
        log.info("=== inventory done in %.1fms ===", (time.perf_counter() - t_start) * 1000)
        return answer

    # ── RAG path ─────────────────────────────────────────────────────────────
    filters = _build_filters(intent)
    q_vec   = embed_query(question, embedder)
    results = retrieve(q_vec, top_k=settings.top_k, filters=filters)
    prompt  = build_prompt(question, results, intent)

    log.info("prompt built — %d chunks, %d chars", len(results), len(prompt))

    answer = generate(prompt, generator)
    log.info("=== query done in %.1fms ===", (time.perf_counter() - t_start) * 1000)
    return answer


def run_query_stream(question: str, chat_history: list[dict]) -> Generator[str, None, None]:
    log.info("=== stream query start: '%s' ===", question[:80])
    t_start = time.perf_counter()

    resolved_question = rewrite_query_with_history(chat_history, question)
    if resolved_question != question:
        log.info("query rewritten: '%s' → '%s'", question[:60], resolved_question[:60])

    intent = extract_intent(resolved_question, _classifier)
    log.info(
        "intent: query_type=%s standar=%s bab_code=%s",
        intent["query_type"],
        intent.get("standar"),
        intent.get("bab_code"),
    )

    # ── Routed handlers — yield as single chunk to keep interface consistent ──
    if intent["query_type"] == "gap_analysis":
        answer = _run_gap_analysis(intent)
        log.info("=== stream gap analysis done in %.1fms ===", (time.perf_counter() - t_start) * 1000)
        yield answer
        return

    if intent["query_type"] == "inventory":
        answer = _run_inventory(intent)
        log.info("=== stream inventory done in %.1fms ===", (time.perf_counter() - t_start) * 1000)
        yield answer
        return

    # ── RAG streaming path ────────────────────────────────────────────────────
    filters = _build_filters(intent)

    t_embed_start = time.perf_counter()
    q_vec = embed_query(resolved_question, embedder)
    log.info("query embed=%.1fms", (time.perf_counter() - t_embed_start) * 1000)

    t_retrieve_start = time.perf_counter()
    results = retrieve(q_vec, top_k=settings.top_k, filters=filters)
    log.info("retrieval=%.1fms chunks=%d", (time.perf_counter() - t_retrieve_start) * 1000, len(results))

    prompt = build_prompt(resolved_question, results, intent)
    log.info(
        "pipeline overhead=%.1fms | prompt_chars=%d",
        (time.perf_counter() - t_start) * 1000,
        len(prompt),
    )

    yield from generate_stream(prompt, generator)


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _build_filters(intent: dict) -> dict | None:
    filters: dict = {}

    if intent["query_type"] == "requirement_lookup":
        filters["is_kmk"] = True
    elif intent["query_type"] == "evidence_check":
        filters["is_kmk"] = False

    if intent.get("standar"):
        filters["standar_code"] = intent["standar"]
    elif intent.get("element_penilaian"):
        filters["element_penilaian_code"] = intent["element_penilaian"]
    elif intent.get("bab_code"):
        filters["bab_code"] = intent["bab_code"]

    return filters or None

def _run_gap_analysis(intent: dict) -> str:
    store = ChromaStore()

    extra = {"bab_code": intent["bab_code"]} if intent.get("bab_code") else {}

    all_ep     = set(store.get_all_unique_values("standar_code", {"is_kmk": True,  **extra}))
    covered_ep = set(store.get_all_unique_values("standar_code", {"is_kmk": False, **extra}))
    missing    = sorted(all_ep - covered_ep)

    prompt = build_gap_prompt(missing, covered_ep, intent)
    return generate(prompt, generator)

def _run_inventory(intent: dict) -> str:
    store   = ChromaStore()
    filters = {k: v for k, v in intent.get("filters", {}).items() if v}

    results = store._col.get(
        where=_build_chroma_where(filters) if filters else None,
        include=["metadatas"],
    )

    seen  = set()
    files = []
    for meta in results["metadatas"]:
        name = meta.get("nama_berkas") or meta.get("source")
        if name and name not in seen:
            seen.add(name)
            files.append(meta)

    if not files:
        return "Belum ada dokumen yang tersimpan dalam sistem."

    lines = [
        f"- {m.get('nama_berkas', m.get('source', '?'))} "
        f"| {m.get('kelompok', '-')} / {m.get('fungsi_pelayanan', '-')} "
        f"| standar: {m.get('standar', '-')}"
        for m in files
    ]
    return f"Dokumen yang tersedia ({len(files)} berkas):\n" + "\n".join(lines)