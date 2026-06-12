"""
chunker.py — Stage 2: split ParsedDocuments into Chunks.

Two paths:
- chunk_kmk()      : KMK document only, one chunk per Standar section
                     (standar statement + maksud + all EP letters)
- chunk_evidence() : evidence documents from upload form,
                     generic sentence-boundary splitting + form metadata attached

Dispatch:
- form_metadata present → chunk_evidence()
- form_metadata absent  → chunk_kmk()
"""
from __future__ import annotations

import re

from app.models import Chunk, ParsedDocument
from app.core.logger import get_logger

log = get_logger("chunker")

# ---------------------------------------------------------------------------
# KMK structure patterns — only used by chunk_kmk()
# ---------------------------------------------------------------------------

_STANDAR_RE  = re.compile(r"Standar\s+([A-Z]+(?:\.[A-Z]+)?\s+[\d]+(?:\.[\d]+)?)")
_MAKSUD_RE   = re.compile(r"Maksud\s+dan\s+Tujuan\s+([A-Z]+\s+[\d.]+)")
_EP_BLOCK_RE = re.compile(r"Elemen\s+Penilaian\s+([A-Z]+\s+[\d.]+)")
_EP_ITEM_RE  = re.compile(r"^([a-z]\))\s+(.+?)(?=^[a-z]\)|$)", re.MULTILINE | re.DOTALL)

_KELOMPOK_MAP = {
    "TKRS": "manajemen_rs", "KPS": "manajemen_rs", "MFK": "manajemen_rs",
    "PMKP": "manajemen_rs", "MRMIK": "manajemen_rs", "PPI": "manajemen_rs",
    "PPK": "manajemen_rs",
    "AKP": "pelayanan_pasien", "HPK": "pelayanan_pasien", "PP": "pelayanan_pasien",
    "PAP": "pelayanan_pasien", "PAB": "pelayanan_pasien", "PKPO": "pelayanan_pasien",
    "KE": "pelayanan_pasien",
    "SKP": "skp",
    "PROGNAS": "prognas",
}


# ---------------------------------------------------------------------------
# Public interface
# ---------------------------------------------------------------------------

def chunk_document(
    doc: ParsedDocument,
    form_metadata: dict | None = None,
    max_tokens: int = 512,
    overlap_tokens: int = 64,
    min_tokens: int = 40,
) -> list[Chunk]:
    """
    Main entry point.

    Args:
        doc:           parsed document from parser.py
        form_metadata: populated from upload form for evidence documents.
                       None means this is the KMK document.
        max_tokens:    max tokens per chunk (generic path only)
        overlap_tokens: overlap between consecutive chunks (generic path only)
        min_tokens:    minimum tokens — chunks below this are dropped
    """
    if form_metadata:
        log.debug("using evidence chunker for %s", doc.source)
        chunks = chunk_evidence(doc, form_metadata, max_tokens, overlap_tokens, min_tokens)
    else:
        log.debug("using KMK chunker for %s", doc.source)
        chunks = chunk_kmk(doc, min_tokens)

    log.info(
        "%s → %d chunks (signature=%s)",
        doc.source,
        len(chunks),
        doc.signature_status.value if form_metadata else "n/a",
    )
    
    return chunks


# ---------------------------------------------------------------------------
# KMK chunker — one chunk per Standar section
# ---------------------------------------------------------------------------

def chunk_kmk(
    doc: ParsedDocument,
    min_tokens: int = 40,
) -> list[Chunk]:
    """
    One chunk per Standar section. Each chunk contains:
    - Standar statement
    - Maksud dan Tujuan
    - All Elemen Penilaian letters

    This gives the LLM everything it needs to understand a requirement
    in a single self-contained chunk.
    """
    chunks: list[Chunk] = []
    sections = _split_into_standar_sections(doc.raw_text)

    for section in sections:
        bab_code  = _extract_bab_code(section["standar"])
        kelompok  = _KELOMPOK_MAP.get(bab_code)

        # Build one complete self-contained chunk
        ep_lines = "\n".join(
            f"{letter} {body.strip()}"
            for letter, body in section["ep_items"]
        )
        text = (
            f"Standar {section['standar']}: {section['standar_text']}\n\n"
            f"Maksud dan Tujuan:\n{section['maksud_text']}\n\n"
            f"Elemen Penilaian:\n{ep_lines}"
        ).strip()

        tok = _count_tokens(text)
        if tok < min_tokens:
            log.debug("dropping small KMK chunk for %s (%d tokens)", section["standar"], tok)
            continue

        chunks.append(Chunk(
            text        = text,
            token_count = tok,
            source      = doc.source,
            chunk_index = len(chunks),
            is_kmk      = True,
            bab_code    = bab_code,
            standar  = section["standar"],
            kelompok    = kelompok,
        ))

    return chunks


# ---------------------------------------------------------------------------
# Evidence chunker — generic splitting + form metadata
# ---------------------------------------------------------------------------

def chunk_evidence(
    doc: ParsedDocument,
    form_metadata: dict,
    max_tokens: int = 512,
    overlap_tokens: int = 64,
    min_tokens: int = 40,
) -> list[Chunk]:
    """
    Generic sentence-boundary chunking for evidence documents.
    All meaningful metadata comes from the upload form, not content detection.
    """
    raw_chunks = _chunk_generic(doc, max_tokens, overlap_tokens, min_tokens)

    # Attach form metadata to every chunk
    enriched = []
    for chunk in raw_chunks:
        chunk.is_kmk           = False
        chunk.bab_code         = form_metadata.get("bab_code")
        chunk.standar       = form_metadata.get("standar")
        chunk.kelompok         = form_metadata.get("kelompok")
        chunk.element_penilaian            = form_metadata.get("element_penilaian")
        chunk.fungsi_pelayanan = form_metadata.get("fungsi_pelayanan")
        chunk.doc_type         = form_metadata.get("doc_type")
        chunk.nama_berkas      = form_metadata.get("nama_berkas")
        chunk.deskripsi        = form_metadata.get("deskripsi", "")
        chunk.is_signed        = doc.signature_status.value
        chunk.standar_code           = form_metadata.get("standar_code")
        chunk.element_penilaian_code = form_metadata.get("element_penilaian_code")
        enriched.append(chunk)

    return enriched


# ---------------------------------------------------------------------------
# Generic chunker — sentence-boundary aware, token-based sliding window
# ---------------------------------------------------------------------------

def _chunk_generic(
    doc: ParsedDocument,
    max_tokens: int,
    overlap_tokens: int,
    min_tokens: int,
) -> list[Chunk]:
    sentences = _split_sentences(doc.raw_text)
    chunks: list[Chunk] = []
    buffer: list[str] = []
    buffer_tokens = 0

    for sentence in sentences:
        s_tokens = _count_tokens(sentence)

        if s_tokens > max_tokens:
            # Single sentence exceeds limit — hard split it
            if buffer:
                chunks.extend(_flush(buffer, buffer_tokens, doc, len(chunks), min_tokens))
                buffer, buffer_tokens = [], 0
            chunks.extend(_hard_split(sentence, doc, len(chunks), max_tokens, min_tokens))
            continue

        if buffer_tokens + s_tokens > max_tokens and buffer:
            chunks.extend(_flush(buffer, buffer_tokens, doc, len(chunks), min_tokens))
            overlap_buf, overlap_tok = _tail_overlap(buffer, overlap_tokens)
            buffer, buffer_tokens = overlap_buf, overlap_tok

        buffer.append(sentence)
        buffer_tokens += s_tokens

    if buffer:
        chunks.extend(_flush(buffer, buffer_tokens, doc, len(chunks), min_tokens))

    return chunks


def _flush(
    buffer: list[str],
    tok: int,
    doc: ParsedDocument,
    idx: int,
    min_tokens: int,
) -> list[Chunk]:
    text = " ".join(buffer).strip()
    if _count_tokens(text) < min_tokens:
        return []
    return [Chunk(
        text        = text,
        token_count = tok,
        source      = doc.source,
        chunk_index = idx,
    )]


def _hard_split(
    text: str,
    doc: ParsedDocument,
    start_idx: int,
    max_tokens: int,
    min_tokens: int,
) -> list[Chunk]:
    """Last resort: split an oversized single sentence by word windows."""
    words  = text.split()
    chunks = []
    i      = 0
    while i < len(words):
        window     = words[i : i + max_tokens]
        chunk_text = " ".join(window)
        tok        = _count_tokens(chunk_text)
        if tok >= min_tokens:
            chunks.append(Chunk(
                text        = chunk_text,
                token_count = tok,
                source      = doc.source,
                chunk_index = start_idx + len(chunks),
            ))
        i += max_tokens
    return chunks


def _tail_overlap(buffer: list[str], overlap_tokens: int) -> tuple[list[str], int]:
    result, tok = [], 0
    for sentence in reversed(buffer):
        s_tok = _count_tokens(sentence)
        if tok + s_tok > overlap_tokens:
            break
        result.insert(0, sentence)
        tok += s_tok
    return result, tok


# ---------------------------------------------------------------------------
# KMK section parsing helpers
# ---------------------------------------------------------------------------

def _split_into_standar_sections(text: str) -> list[dict]:
    """
    Walk the KMK text and collect each Standar block.
    Returns list of dicts: {standar, standar_text, maksud_text, ep_items}
    """
    sections   = []
    boundaries = [(m.start(), m.group(1).strip()) for m in _STANDAR_RE.finditer(text)]

    for i, (start, standar) in enumerate(boundaries):
        end   = boundaries[i + 1][0] if i + 1 < len(boundaries) else len(text)
        block = text[start:end]

        standar_text = _extract_between(block, _STANDAR_RE, _MAKSUD_RE) or ""
        maksud_text  = _extract_between(block, _MAKSUD_RE,  _EP_BLOCK_RE) or ""
        ep_items     = _EP_ITEM_RE.findall(block)   # list of (letter, body) tuples

        sections.append({
            "standar":   standar,
            "standar_text": standar_text.strip(),
            "maksud_text":  maksud_text.strip(),
            "ep_items":     ep_items,
        })

    return sections


def _extract_between(text: str, start_re: re.Pattern, end_re: re.Pattern) -> str:
    m_start = start_re.search(text)
    if not m_start:
        return ""
    body_start = m_start.end()
    m_end      = end_re.search(text, body_start)
    body_end   = m_end.start() if m_end else len(text)
    return text[body_start:body_end].strip()


def _extract_bab_code(standar: str) -> str:
    return standar.split()[0] if standar else "UNKNOWN"


# ---------------------------------------------------------------------------
# Shared helpers
# ---------------------------------------------------------------------------

_SENTENCE_RE = re.compile(r"(?<=[.!?;])\s+")


def _split_sentences(text: str) -> list[str]:
    parts = _SENTENCE_RE.split(text)
    return [p.strip() for p in parts if p.strip()]


def _count_tokens(text: str) -> int:
    """
    Fast approximation: 1 token ≈ 4 chars for Latin/Indonesian text.
    Replace with tiktoken or tokenizer.encode() for exact counts.
    """
    return max(1, len(text) // 4)