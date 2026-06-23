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
from app.repositories.ingestion.upload_extractor import extract_text_from_upload
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
        "intent: query_type=%s standar=%s fungsi_pelayanan=%s",
        intent["query_type"],
        intent.get("standar"),
        intent.get("fungsi_pelayanan"),
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


def run_query_stream(
    question: str,
    chat_history: list[dict],
    app_context=None,
    file_path: str | None = None,
    content_type: str | None = None,
) -> Generator[str, None, None]:
    log.info("=== stream query start: '%s' ===", question[:80])

    doc_context = None
    if file_path:
        doc_context = extract_text_from_upload(file_path, content_type)
        if doc_context is None:
            log.warning("file extraction failed or returned empty: %s", file_path)
            yield "Dokumen ini tidak dapat dibaca (kemungkinan hasil scan tanpa teks)."
            return
        log.info("file extracted: %d chars from %s", len(doc_context), file_path)

    resolved_question = rewrite_query_with_history(chat_history, question)
    if resolved_question != question:
        log.info("query rewritten: '%s' → '%s'", question[:60], resolved_question[:60])

    intent = extract_intent(resolved_question, _classifier)

    if intent["query_type"] == "ui_navigation" and not intent.get("standar") and not intent.get("fungsi_pelayanan"):
        intent["query_type"] = "general"
        if "ui_navigation" in intent.get("all_intents", []):
            intent["all_intents"].remove("ui_navigation")

    log.info(
        "intent: query_type=%s standar=%s fungsi_pelayanan=%s",
        intent["query_type"], intent.get("standar"), intent.get("fungsi_pelayanan"),
    )

    all_intents = intent.get("all_intents", [intent["query_type"]])
    log.info("active intents: %s", all_intents)

    needs_rag = any(i in all_intents for i in ("requirement_lookup", "evidence_check")) or doc_context is not None

    if needs_rag:
        filters = _build_filters(intent)
        if doc_context is not None and not filters:
            filters = {"is_kmk": True}  

        q_vec   = embed_query(resolved_question, embedder)
        results = retrieve(q_vec, top_k=settings.top_k, filters=filters)
        prompt  = build_prompt(resolved_question, results, intent, doc_context=doc_context)

        yield from generate_stream(prompt, generator)

        if "ui_navigation" in all_intents or _should_navigate(all_intents, results):
            nav_block = _build_navigation_block(intent, results, app_context)
            if nav_block:
                yield "\n\n" + nav_block

    elif "ui_navigation" in all_intents:
        nav_block = _build_navigation_block(intent, [], app_context)
        if nav_block:
            yield nav_block

    elif intent["query_type"] == "gap_analysis":
        yield _run_gap_analysis(intent)

    elif intent["query_type"] == "inventory":
        yield _run_inventory(intent)

    else:
        yield from generate_stream(resolved_question, generator)


def _should_navigate(all_intents: list[str], results: list, min_score: float = 0.85) -> bool:
    """Only navigate if evidence_check is active AND results are genuinely relevant."""
    if "evidence_check" not in all_intents:
        return False
    if not results:
        return False
    return results[0].score >= min_score


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
    elif intent.get("fungsi_pelayanan"):
        filters["fungsi_pelayanan"] = intent["fungsi_pelayanan"]

    return filters or None

def _run_gap_analysis(intent: dict) -> str:
    store = ChromaStore()

    extra = {"fungsi_pelayanan": intent["fungsi_pelayanan"]} if intent.get("fungsi_pelayanan") else {}

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

_NAV_DEPTH_PRIORITY = ["nav_to_document", "nav_to_assessment", "nav_to_standard", "nav_to_service"]

def _infer_nav_depth(all_intents: list[str]) -> str:
    for depth_intent in _NAV_DEPTH_PRIORITY:
        if depth_intent in all_intents:
            return depth_intent.replace("nav_to_", "")
    return "standard"

def _build_navigation_block(intent: dict, results: list, ctx=None) -> str:
    lines = []
    nav_depth = intent.get("nav_depth", "standard")

    if not ctx or ctx.current_path != "/storage":
        lines.append('<command>{"type":"NAVIGATE","path":"/storage"}</command>')

    evidence_results = [r for r in results if not r.is_kmk]
    top = evidence_results[0] if evidence_results else None

    # ── Service ──────────────────────────────────────────────────────────────
    service_id = (top.service_id if top else None)
    if not service_id and intent.get("fungsi_pelayanan") and ctx:  # ← was bab_code
        matched = _match_service(intent, ctx)
        if matched:
            service_id = matched.id

    if service_id:
        lines.append(f'<command>{{"type":"SET_SERVICE_FILTER","serviceId":"{service_id}"}}</command>')

    if nav_depth == "service":
        lines.append("\n📂 Menampilkan layanan yang relevan...")
        return "\n".join(lines)

    # ── Standard ─────────────────────────────────────────────────────────────
    if top and top.standard_id:
        lines.append(f'<command>{{"type":"SET_STANDARD_FILTER","standardId":"{top.standard_id}"}}</command>')

    if nav_depth == "standard":
        label = (top.standar_code if top else None) or intent.get("standar") or "ini"
        lines.append(f"\n📂 Menampilkan dokumen untuk standar **{label}**...")
        return "\n".join(lines)

    # ── Assessment ───────────────────────────────────────────────────────────
    if top and top.assessment_id:
        lines.append(f'<command>{{"type":"SET_ASSESSMENT_FILTER","assessmentId":"{top.assessment_id}"}}</command>')

    if nav_depth == "assessment":
        lines.append("\n📂 Menampilkan assessment yang relevan...")
        return "\n".join(lines)

    # ── Document ─────────────────────────────────────────────────────────────
    if top and top.document_id:
        lines.append(f'<command>{{"type":"OPEN_DOCUMENT","documentId":"{top.document_id}"}}</command>')
        lines.append(f"\n📄 Membuka **{top.nama_berkas or 'dokumen'}** yang relevan...")
    elif top:
        lines.append(
            "\n⚠️ Dokumen ditemukan dalam indeks tetapi belum memiliki ID — "
            "pastikan metadata `document_id` disimpan saat ingestion."
        )
    else:
        if "evidence_check" in intent.get("all_intents", []):
            lines.append(
                "\n📭 Tidak ditemukan berkas bukti untuk standar ini. "
                "Silakan upload dokumen yang relevan."
            )

    return "\n".join(lines)

def _match_service(intent: dict, ctx) -> object | None:
    """
    Tries to find the matching service option from AppContext
    based on the fungsi_pelayanan or standar code in the intent.

    Returns the first ServiceOption whose label contains the bab prefix,
    or None if no match is found.
    """
    if not ctx or not ctx.available_services:
        return None

    fungsi = (intent.get("fungsi_pelayanan") or "").upper()
    standar = (intent.get("standar") or "").upper()

    search_terms = []
    if fungsi:
        # "BAB TKRS" → "TKRS"
        parts = fungsi.split()
        search_terms.extend(parts[1:] if len(parts) > 1 else parts)
    if standar:
        # "AP 1.1" → "AP"
        search_terms.append(standar.split()[0])

    for term in search_terms:
        for svc in ctx.available_services:
            if term in svc.label.upper():
                return svc

    return None