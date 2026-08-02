"""
services/rag_service.py — orchestrates retrieval and generation at query time.
"""
from __future__ import annotations

from typing import Generator

from app.core.config import settings
from app.repositories.generation.generator import generator, generate_stream
from app.repositories.generation.intent_extractor import classifier, extract_intent
from app.repositories.generation.prompt_builder import build_prompt
from app.repositories.generation.query_rewriter import rewrite_query_with_history
from app.repositories.ingestion.embedder import embedder, embed_query
from app.repositories.ingestion.parser import parse_for_upload
from app.core.logger import get_logger
from app.repositories.retrieval.retriever import retrieve
from app.repositories.generation.gap_inventory_analysis import run_gap_analysis, run_inventory
from app.repositories.generation.navigation_builder import (
    NAV_DEPTH_PRIORITY,
    build_navigation_block,
    should_navigate,
)

log = get_logger("rag_pipeline")


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
        doc_context = parse_for_upload(file_path, content_type)
        if doc_context is None:
            log.warning("file extraction failed or returned empty: %s", file_path)
            yield "Dokumen ini tidak dapat dibaca (kemungkinan hasil scan tanpa teks)."
            return
        log.info("file extracted: %d chars from %s", len(doc_context), file_path)

    resolved_question = rewrite_query_with_history(chat_history, question)
    if resolved_question != question:
        log.info("query rewritten: '%s' → '%s'", question[:60], resolved_question[:60])

    intent = extract_intent(resolved_question, classifier)

    any_nav_depth = any(i in intent.get("all_intents", []) for i in NAV_DEPTH_PRIORITY)

    if (
        intent["query_type"] == "ui_navigation"
        and not intent.get("standar")
        and not intent.get("fungsi_pelayanan")
        and not any_nav_depth
    ):
        intent["query_type"] = "general"
        intent["all_intents"] = [
            i for i in intent.get("all_intents", []) if i != "ui_navigation"
        ]

    if (
        intent["query_type"] not in ("ui_navigation", "gap_analysis", "inventory")
        and any_nav_depth
    ):
        log.info(
            "promoting query_type=%s → ui_navigation (nav_depth intents active: %s)",
            intent["query_type"], intent["all_intents"],
        )
        intent["query_type"] = "ui_navigation"

    log.info(
        "intent: query_type=%s standar=%s fungsi_pelayanan=%s",
        intent["query_type"], intent.get("standar"), intent.get("fungsi_pelayanan"),
    )

    all_intents = intent.get("all_intents", [intent["query_type"]])
    log.info("active intents: %s", all_intents)

    needs_rag = any(i in all_intents for i in ("requirement_lookup", "evidence_check")) or doc_context is not None

    if intent["query_type"] == "gap_analysis":
        yield run_gap_analysis(intent)
        return

    if intent["query_type"] == "inventory":
        yield run_inventory(intent)
        return

    if not needs_rag and any(i in all_intents for i in ("nav_to_document", "nav_to_assessment", "nav_to_standard")):
        needs_rag = True

    if needs_rag:
        filters = _build_filters(intent)
        if doc_context is not None and not filters:
            filters = {"is_kmk": True}

        q_vec   = embed_query(resolved_question, embedder)
        results = retrieve(q_vec, top_k=settings.top_k, filters=filters)
        prompt  = build_prompt(resolved_question, results, intent, doc_context=doc_context)

        yield from generate_stream(prompt, generator)

        should_nav = (
            "ui_navigation" in all_intents
            or any(i in all_intents for i in NAV_DEPTH_PRIORITY)
            or should_navigate(all_intents, results)
        )
        if should_nav:
            nav_block = build_navigation_block(intent, results, app_context)
            if nav_block:
                yield "\n\n" + nav_block

    elif (
        "ui_navigation" in all_intents
        or any(i in all_intents for i in NAV_DEPTH_PRIORITY)
    ):
        nav_block = build_navigation_block(intent, [], app_context)
        if nav_block:
            yield nav_block

    else:
        yield from generate_stream(resolved_question, generator)


# ---------------------------------------------------------------------------
# Helpers (orchestration-specific, kept local to this service)
# ---------------------------------------------------------------------------

def _build_filters(intent: dict) -> dict | None:
    filters: dict = {}
    all_intents = intent.get("all_intents", [])

    if intent["query_type"] == "requirement_lookup":
        filters["is_kmk"] = True
    elif intent["query_type"] == "evidence_check":
        filters["is_kmk"] = False
    elif "requirement_lookup" in all_intents and "evidence_check" not in all_intents:
        filters["is_kmk"] = True
    elif "evidence_check" in all_intents and "requirement_lookup" not in all_intents:
        filters["is_kmk"] = False
    elif any(i in all_intents for i in ("nav_to_document", "nav_to_assessment", "nav_to_standard")):
        filters["is_kmk"] = False

    if intent.get("standar"):
        filters["standar_code"] = intent["standar"]
    elif intent.get("fungsi_pelayanan"):
        filters["fungsi_pelayanan"] = intent["fungsi_pelayanan"]

    return filters or None