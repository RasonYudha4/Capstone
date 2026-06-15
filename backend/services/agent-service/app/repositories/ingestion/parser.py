"""
parser.py — Stage 1: extract raw text from source files.

Extraction priority:
  1. pymupdf4llm  — primary: produces clean Markdown, preserves heading/list structure
  2. pdfplumber   — fallback: good layout extraction when pymupdf4llm unavailable
  3. pypdf        — last resort: basic text extraction only

- Per-page text stored separately (enables page_ref metadata)
- Watermark token stripping for SARS KMK documents
- Signature detection on last page (cheap heuristic pass)
- Failed files recorded in result, never silently swallowed
- Returns ParsedDocument dataclasses, not raw dicts
"""
from __future__ import annotations

import re
from pathlib import Path

import pymupdf4llm
import pdfplumber
from pypdf import PdfReader

from app.models import ParsedDocument, SignatureStatus
from app.core.logger import get_logger

log = get_logger("parser")

# ---------------------------------------------------------------------------
# Constants
# ---------------------------------------------------------------------------

_SUPPORTED_EXTENSIONS = {".txt", ".md", ".pdf"}

# Watermark tokens injected by the SARS PDF generator — strip before chunking.
# These are single uppercase tokens that appear on their own lines.
_WATERMARK_RE = re.compile(
    r"(?m)^\s*(?:AT|AN|SE|KE|IA|TE|EN|M|N|R|H)\s*$"
)

# Heuristic: presence of these strings on the last page suggests a signed doc.
_SIGNATURE_HINTS = [
    "mengetahui", "kepala", "direktur", "kaur", "ttd",
    "tanda tangan", "menyetujui", "penandatangan",
]


# ---------------------------------------------------------------------------
# Public interface
# ---------------------------------------------------------------------------

def parse_single(file_path: str, is_kmk: bool = False) -> ParsedDocument | None:
    """Parse a single file. Used by ingest_pipeline for individual uploads."""
    path = Path(file_path)
    if path.suffix not in _SUPPORTED_EXTENSIONS:
        log.warning("unsupported file type: %s", path.suffix)
        return None
    try:
        return _parse_file(path, is_kmk=is_kmk)
    except Exception as exc:
        log.warning("failed to parse %s: %s", path, exc)
        return None

def load_documents(folder: str) -> tuple[list[ParsedDocument], list[str]]:
    """
    Walk *folder* recursively and parse every supported file.

    Returns:
        docs:   successfully parsed documents
        errors: list of "path: reason" strings for failed files
    """
    docs, errors = [], []

    for path in sorted(Path(folder).rglob("*")):
        if path.suffix not in _SUPPORTED_EXTENSIONS:
            continue
        try:
            doc = _parse_file(path)
            if doc:
                docs.append(doc)
            else:
                errors.append(f"{path}: empty after extraction")
        except Exception as exc:
            log.warning("failed to parse %s: %s", path, exc)
            errors.append(f"{path}: {exc}")

    log.info("parsed %d doc(s), %d error(s) from %s", len(docs), len(errors), folder)
    return docs, errors


# ---------------------------------------------------------------------------
# Dispatch
# ---------------------------------------------------------------------------

def _parse_file(path: Path, is_kmk: bool = False) -> ParsedDocument | None:
    if path.suffix == ".pdf":
        return _parse_pdf(path, is_kmk=is_kmk)

    text = path.read_text(encoding="utf-8", errors="ignore").strip()
    if not text:
        return None

    return ParsedDocument(
        source=str(path),
        file_type=path.suffix.lstrip("."),
        raw_text=text,
        pages=[text],
        total_pages=1,
    )


# ---------------------------------------------------------------------------
# PDF extraction — three-tier fallback chain
# ---------------------------------------------------------------------------

def _parse_pdf(path: Path, is_kmk: bool = False) -> ParsedDocument | None:
    """
    Try extractors in priority order.
    Each extractor returns (pages_text, full_text) or raises on hard failure.
    """
    extractors = (
        [_extract_pdfplumber, _extract_pypdf]
        if is_kmk else
        [_extract_pymupdf4llm, _extract_pdfplumber, _extract_pypdf]
    )
    for extractor in extractors:
        try:
            pages_text, full_text = extractor(path)
            if full_text.strip():
                sig_status = _detect_signature_heuristic(
                    pages_text[-1] if pages_text else ""
                )
                return ParsedDocument(
                    source=str(path),
                    file_type="pdf",
                    raw_text=full_text,
                    pages=pages_text,
                    total_pages=len(pages_text),
                    signature_status=sig_status,
                )
        except Exception as exc:
            log.warning(
                "%s failed for %s, trying next extractor: %s",
                extractor.__name__, path.name, exc,
            )

    log.warning("all extractors failed for %s — may be a scanned PDF", path)
    return None


def _extract_pymupdf4llm(path: Path) -> tuple[list[str], str]:
    """
    Primary extractor.
    page_chunks=True returns one dict per page with a 'text' key containing
    Markdown-formatted content — headings, bold, lists are preserved.
    This gives the downstream parser reliable structural anchors without
    relying solely on regex against flat text.
    """
    page_chunks: list[dict] = pymupdf4llm.to_markdown(
        str(path),
        page_chunks=True,   
        show_progress=False,
    )
    pages_text = [_strip_watermarks(chunk["text"]) for chunk in page_chunks]
    full_text  = "\n\n".join(pages_text).strip()
    return pages_text, full_text


def _extract_pdfplumber(path: Path) -> tuple[list[str], str]:
    """
    First fallback. Good for documents where pymupdf4llm produces
    garbled output (e.g. heavy use of non-standard fonts).
    """
    pages_text: list[str] = []
    with pdfplumber.open(str(path)) as pdf:
        for page in pdf.pages:
            raw = page.extract_text() or ""
            pages_text.append(_strip_watermarks(raw))
    full_text = "\n".join(pages_text).strip()
    return pages_text, full_text


def _extract_pypdf(path: Path) -> tuple[list[str], str]:
    """
    Last resort. Loses most layout information but handles edge cases
    that pdfplumber and pymupdf4llm cannot open.
    """
    reader     = PdfReader(str(path))
    pages_text = [_strip_watermarks(p.extract_text() or "") for p in reader.pages]
    full_text  = "\n".join(pages_text).strip()
    return pages_text, full_text


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _strip_watermarks(text: str) -> str:
    return _WATERMARK_RE.sub("", text).strip()


def _detect_signature_heuristic(last_page_text: str) -> SignatureStatus:
    """
    Cheap text-based heuristic for signature detection on the last page.
    For vision-based detection of wet ink signatures (scanned docs)
    use enricher.py instead.
    """
    lower = last_page_text.lower()
    hits  = sum(1 for hint in _SIGNATURE_HINTS if hint in lower)
    if hits >= 3:
        return SignatureStatus.SIGNED
    if hits >= 1:
        return SignatureStatus.PARTIAL
    return SignatureStatus.UNKNOWN