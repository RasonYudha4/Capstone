"""
generation/prompt_builder.py — builds prompts for the LLM.

Two prompt types:
- build_prompt()     : standard RAG prompt, context-aware of KMK vs evidence chunks
- build_gap_prompt() : gap analysis prompt, no retrieved chunks needed
"""
from __future__ import annotations

from app.core.store.base import SearchResult


def build_prompt(question: str, results: list[SearchResult], intent: dict) -> str:
    """
    Build a RAG prompt from retrieved chunks.
    Separates KMK requirement chunks from evidence chunks
    so the LLM understands which is the standard and which is the proof.
    """
    if not results:
        return _no_context_prompt(question)

    kmk_chunks      = [r for r in results if r.is_kmk]
    evidence_chunks = [r for r in results if not r.is_kmk]

    sections: list[str] = []

    if kmk_chunks:
        kmk_text = "\n\n---\n\n".join(r.text for r in kmk_chunks)
        sections.append(f"STANDAR AKREDITASI (KMK):\n{kmk_text}")

    if evidence_chunks:
        ev_parts = []
        for r in evidence_chunks:
            label = f"[{r.doc_type or 'dokumen'} — {r.standar_id or ''}]"
            ev_parts.append(f"{label}\n{r.text}")
        ev_text = "\n\n---\n\n".join(ev_parts)
        sections.append(f"BUKTI DOKUMEN:\n{ev_text}")

    context = "\n\n====\n\n".join(sections)

    return f"""Kamu adalah asisten sistem akreditasi rumah sakit.
Gunakan HANYA konteks di bawah untuk menjawab pertanyaan.
Jika jawaban tidak ada dalam konteks, katakan "Informasi tidak ditemukan."

{context}

PERTANYAAN: {question}

JAWABAN:"""


def build_gap_prompt(
    missing: list[str],
    covered: set[str],
    intent: dict,
) -> str:
    """
    Build a prompt for gap analysis — no retrieved chunks, just metadata.
    """
    scope = f" untuk BAB {intent['bab_code']}" if intent.get("bab_code") else ""
    covered_list = "\n".join(f"- {ep}" for ep in sorted(covered)) or "Belum ada"
    missing_list = "\n".join(f"- {ep}" for ep in missing) or "Semua sudah terpenuhi"

    return f"""Kamu adalah asisten sistem akreditasi rumah sakit.
Berikut adalah hasil analisis kelengkapan dokumen bukti{scope}.

SUDAH ADA BUKTI:
{covered_list}

BELUM ADA BUKTI:
{missing_list}

Berikan ringkasan yang jelas tentang status kelengkapan dokumen akreditasi,
sebutkan EP yang masih perlu dilengkapi dan prioritaskan yang belum ada sama sekali.

JAWABAN:"""


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _no_context_prompt(question: str) -> str:
    return f"""Kamu adalah asisten sistem akreditasi rumah sakit.
Tidak ditemukan dokumen yang relevan untuk pertanyaan ini.

PERTANYAAN: {question}

JAWABAN: Informasi tidak ditemukan dalam sistem. Pastikan dokumen terkait sudah diunggah."""