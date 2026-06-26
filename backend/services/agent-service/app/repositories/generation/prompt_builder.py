"""
generation/prompt_builder.py
"""
from __future__ import annotations

from app.core.store.base import SearchResult

# ---------------------------------------------------------------------------
# Response style instructions keyed by query_type
# ---------------------------------------------------------------------------

_STYLE: dict[str, str] = {
    "requirement_lookup": (
        "Jawab secara ringkas dan terstruktur. "
        "Gunakan poin-poin hanya jika ada lebih dari satu persyaratan. "
        "Maksimal 150 kata kecuali detail teknis memang diperlukan."
    ),
    "evidence_check": (
        "Jawab langsung: sebutkan dokumen yang ada, tipe, dan statusnya. "
        "Jangan tambahkan penjelasan panjang. Maksimal 100 kata."
    ),
    "gap_analysis": (
        "Fokus pada apa yang kurang. "
        "Sebutkan EP yang belum terpenuhi dalam format daftar singkat. "
        "Maksimal 150 kata."
    ),
    "inventory": (
        "Tampilkan dalam format daftar. "
        "Satu baris berdasarkan fungsi pelayanan dan standar."
        "Berikan jumlah dokumentnya"
    ),
    "general": (
        "Jawab dengan singkat dan jelas. "
        "Maksimal 80 kata. Hindari penjelasan yang tidak diminta."
    ),
}

_DEFAULT_STYLE = (
    "Jawab secara ringkas. Maksimal 120 kata. "
    "Berikan detail hanya jika pertanyaan memang meminta penjelasan mendalam."
)


def build_prompt(question: str, results: list[SearchResult], intent: dict, doc_context: str | None = None) -> str:
    query_type = intent.get("query_type", "general")
    style      = _STYLE.get(query_type, _DEFAULT_STYLE)

    if not results and not doc_context:
        return _no_context_prompt(question, intent)

    sections: list[str] = []

    if doc_context:
        sections.append(f"DOKUMEN YANG DIUNGGAH USER:\n{doc_context}")

    kmk_chunks      = [r for r in results if r.is_kmk]
    evidence_chunks = [r for r in results if not r.is_kmk]

    if kmk_chunks:
        kmk_text = "\n\n---\n\n".join(r.text for r in kmk_chunks)
        sections.append(f"STANDAR AKREDITASI (KMK):\n{kmk_text}")

    if evidence_chunks:
        ev_parts = []
        for r in evidence_chunks:
            label = f"[{r.doc_type or 'dokumen'} — {r.standar or ''}]"
            ev_parts.append(f"{label}\n{r.text}")
        ev_text = "\n\n---\n\n".join(ev_parts)
        sections.append(f"BUKTI DOKUMEN:\n{ev_text}")

    context = "\n\n====\n\n".join(sections)

    return f"""Kamu adalah asisten sistem akreditasi rumah sakit.
        Gunakan HANYA konteks di bawah untuk menjawab pertanyaan.
        Jika dokumen yang diunggah user disertakan, bandingkan isinya dengan STANDAR AKREDITASI (KMK)
        untuk membantu menentukan standar/EP yang relevan.
        Jika jawaban tidak ada dalam konteks, katakan "Informasi tidak ditemukan."

        INSTRUKSI GAYA JAWABAN: {style}

        {context}

        PERTANYAAN: {question}

        JAWABAN:"""


def build_gap_prompt(
    missing: list[str],
    covered: set[str],
    intent:  dict,
) -> str:
    scope        = f" untuk BAB {intent['fungsi_pelayanan']}" if intent.get("fungsi_pelayanan") else ""
    covered_list = "\n".join(f"- {ep}" for ep in sorted(covered)) or "Belum ada"
    missing_list = "\n".join(f"- {ep}" for ep in missing)         or "Semua sudah terpenuhi"

    return f"""Kamu adalah asisten sistem akreditasi rumah sakit.
        Berikut hasil analisis kelengkapan dokumen bukti{scope}.

        SUDAH ADA BUKTI:
        {covered_list}

        BELUM ADA BUKTI:
        {missing_list}

        INSTRUKSI: Buat ringkasan singkat status kelengkapan. Prioritaskan EP yang belum ada.
        Gunakan format daftar. Maksimal 150 kata.

        JAWABAN:"""


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _no_context_prompt(question: str, intent: dict) -> str:
    style = _STYLE.get(intent.get("query_type", "general"), _DEFAULT_STYLE)
    return f"""Kamu adalah asisten sistem akreditasi rumah sakit.
        Tidak ditemukan dokumen yang relevan untuk pertanyaan ini.

        INSTRUKSI GAYA JAWABAN: {style}

        PERTANYAAN: {question}

        JAWABAN: Informasi tidak ditemukan dalam sistem. Pastikan dokumen terkait sudah diunggah."""