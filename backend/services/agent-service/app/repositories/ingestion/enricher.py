"""
enricher.py — Stage 1.5: enrich ParsedDocuments with signature status.

Runs after parser.py, before chunker.py — evidence documents only.
Uses heuristic text-based detection only — no vision model needed
since evidence documents are digital PDFs with extractable text.

Signature detection skips trailing photo/documentation pages by filtering
for pages with substantial text content before checking for signature hints.
"""
from __future__ import annotations

from pathlib import Path

from models import ParsedDocument, SignatureStatus
from backend.services.rag_test.app.core.logger import get_logger

log = get_logger("enricher")

_LAST_N_PAGES = 2
_MIN_PAGE_CHARS = 100

_SIGNATURE_HINTS = [
    "mengetahui", "kepala", "direktur", "kaur",
    "ttd", "tanda tangan", "menyetujui", "penandatangan",
    "jabatan", "nama", "nip", "tandatangan",
    "ditetapkan", "disahkan", "disetujui",
]


# ---------------------------------------------------------------------------
# Public interface
# ---------------------------------------------------------------------------

def enrich_document(doc: ParsedDocument) -> ParsedDocument:
    """
    Detect signature status and attach to doc.
    Called only for evidence documents — KMK does not need this.
    """
    doc.signature_status = _detect_signature(doc)
    log.info(
        "%s → signature_status=%s",
        Path(doc.source).name,
        doc.signature_status.value,
    )
    return doc


# ---------------------------------------------------------------------------
# Signature detection
# ---------------------------------------------------------------------------

def _detect_signature(doc: ParsedDocument) -> SignatureStatus:
    """
    Walk pages in reverse, skipping trailing image-only pages (< 100 chars).
    Check the last N meaningful text pages for signature hints.
    """
    text_pages = [
        page for page in reversed(doc.pages)
        if len(page.strip()) > _MIN_PAGE_CHARS
    ]

    if not text_pages:
        log.warning(
            "no text pages found in %s — defaulting to UNKNOWN",
            Path(doc.source).name,
        )
        return SignatureStatus.UNKNOWN

    # Take last 2 meaningful text pages
    last_text = " ".join(text_pages[:_LAST_N_PAGES]).lower()
    status = _heuristic_signature(last_text)

    log.debug(
        "%s — checked %d text page(s), hints found → %s",
        Path(doc.source).name,
        min(len(text_pages), _LAST_N_PAGES),
        status.value,
    )
    return status


def _heuristic_signature(text: str) -> SignatureStatus:
    hits = sum(1 for hint in _SIGNATURE_HINTS if hint in text)
    log.debug("signature hints matched: %d / %d", hits, len(_SIGNATURE_HINTS))
    if hits >= 3:
        return SignatureStatus.SIGNED
    if hits >= 1:
        return SignatureStatus.PARTIAL
    return SignatureStatus.UNKNOWN