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


def run_query_stream(question: str, chat_history: list[dict], app_context=None) -> Generator[str, None, None]:
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

    all_intents = intent.get("all_intents", [intent["query_type"]])

    log.info("active intents: %s", all_intents)

    parts: list[str] = []

    needs_rag = any(i in all_intents for i in ("requirement_lookup", "evidence_check"))

    if needs_rag:
        filters = _build_filters(intent)
        q_vec   = embed_query(resolved_question, embedder)
        results = retrieve(q_vec, top_k=settings.top_k, filters=filters)
        prompt  = build_prompt(resolved_question, results, intent)

        # Collect the full RAG answer first (we need it before appending commands)
        rag_answer = "".join(generate_stream(prompt, generator))
        parts.append(rag_answer)

        # 2. Navigation commands — derived from what RAG actually found
        if "ui_navigation" in all_intents or _should_navigate(all_intents, results):
            nav_block = _build_navigation_block(intent, results, app_context)
            if nav_block:
                parts.append(nav_block)

    # ── Pure navigation (no RAG needed) ───────────────────────────────────────
    elif "ui_navigation" in all_intents:
        nav_block = _build_navigation_block(intent, [], app_context)
        if nav_block:
            parts.append(nav_block)

    # ── Unchanged single-intent paths ─────────────────────────────────────────
    elif intent["query_type"] == "gap_analysis":
        parts.append(_run_gap_analysis(intent))
    elif intent["query_type"] == "inventory":
        parts.append(_run_inventory(intent))
    else:
        q_vec   = embed_query(resolved_question, embedder)
        results = retrieve(q_vec, top_k=settings.top_k, filters=None)
        prompt  = build_prompt(resolved_question, results, intent)
        parts.append("".join(generate_stream(prompt, generator)))

    yield "\n\n".join(parts)


def _should_navigate(all_intents: list[str], results: list) -> bool:
    """Automatically add navigation if evidence was found — user probably wants to open it."""
    return "evidence_check" in all_intents and len(results) > 0


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

def _build_navigation_block(intent: dict, results: list, ctx=None) -> str:
    lines = []

    # Navigate to storage if not already there
    if not ctx or ctx.current_path != "/storage":
        lines.append('<command>{"type":"NAVIGATE","path":"/storage"}</command>')

    # Set filters from intent
    if intent.get("standar") and ctx:
        matched_service = _match_service(intent, ctx)
        if matched_service:
            lines.append(
                f'<command>{{"type":"SET_SERVICE_FILTER",'
                f'"serviceId":"{matched_service.id}","label":"{matched_service.label}"}}</command>'
            )

    # If RAG found actual evidence chunks, open the first matching document
    evidence_results = [r for r in results if not r.is_kmk]
    if evidence_results:
        # metadata on the chunk tells us the real document id
        doc_id       = evidence_results[0].metadata.get("document_id")
        doc_filename = evidence_results[0].metadata.get("nama_berkas", "dokumen")

        if doc_id:
            lines.append(
                f'<command>{{"type":"OPEN_DOCUMENT","documentId":"{doc_id}"}}</command>'
            )
            lines.append(f"\n📄 Membuka **{doc_filename}** yang relevan...")
        else:
            lines.append(
                "\n⚠️ Dokumen ditemukan dalam indeks tetapi belum memiliki ID — "
                "pastikan metadata `document_id` disimpan saat ingestion."
            )
    elif "evidence_check" in intent.get("all_intents", []):
        lines.append(
            "\n📭 Tidak ditemukan berkas bukti untuk standar ini. "
            "Silakan upload dokumen yang relevan."
        )

    log.info("nav block output: %s", "\n".join(lines))

    return "\n".join(lines)

def _match_service(intent: dict, ctx) -> object | None:
    """
    Tries to find the matching service option from AppContext
    based on the bab_code or standar code in the intent.

    Returns the first ServiceOption whose label contains the bab prefix,
    or None if no match is found.
    """
    if not ctx or not ctx.available_services:
        return None

    bab = (intent.get("bab_code") or "").upper()
    standar = (intent.get("standar") or "").upper()

    search_terms = []
    if bab:
        # "BAB TKRS" → "TKRS"
        parts = bab.split()
        search_terms.extend(parts[1:] if len(parts) > 1 else parts)
    if standar:
        # "AP 1.1" → "AP"
        search_terms.append(standar.split()[0])

    for term in search_terms:
        for svc in ctx.available_services:
            if term in svc.label.upper():
                return svc

    return None