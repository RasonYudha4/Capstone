from __future__ import annotations

import re

from app.repositories.generation.generator import generator, generate
from app.core.logger import get_logger

log = get_logger("query_rewriter")

# ---------------------------------------------------------------------------
# Config
# ---------------------------------------------------------------------------

_CYCLES_TO_KEEP = 3         
_TURNS_TO_KEEP  = _CYCLES_TO_KEEP * 2

_DEPENDENCY_PATTERN = re.compile(
    r"\b(tersebut|itu|ini|tadi|tersisa|yang (lain|sebelumnya|kedua|pertama))\b"
    r"|^\s*(dan|atau|bagaimana dengan|gimana dengan)\b",
    re.IGNORECASE,
)

_REWRITE_INSTRUCTION = """Kamu bertugas menulis ulang pertanyaan terbaru pengguna agar berdiri sendiri \
(self-contained), dengan mengganti kata ganti atau rujukan implisit (seperti "itu", "tersebut", \
"ini", "tadi") menggunakan informasi spesifik dari riwayat percakapan di bawah.

ATURAN:
- Jangan menjawab pertanyaannya. Tugasmu HANYA menulis ulang pertanyaan tersebut.
- Jika pertanyaan sudah jelas berdiri sendiri tanpa riwayat, kembalikan apa adanya.
- Jangan menambahkan informasi yang tidak ada di riwayat atau pertanyaan asli.
- Keluarkan HANYA kalimat pertanyaan hasil tulis ulang, tanpa penjelasan tambahan.

RIWAYAT PERCAKAPAN:
{history_block}

PERTANYAAN TERBARU: {question}

PERTANYAAN HASIL TULIS ULANG:"""


def rewrite_query_with_history(
    chat_history: list[dict],
    question:     str,
) -> str:
    """
    Resolves a potentially referentially-dependent question into a
    self-contained one using the last `_CYCLES_TO_KEEP` chat cycles.

    Falls back to the original question when:
      - there's no history to rewrite against, or
      - the question shows no sign of referential dependency (cheap heuristic
        gate, avoids an LLM call on the common case where nothing needs resolving).

    Note: this is intentionally independent of IntentClassifier — rewriting
    needs generation, not embedding-centroid similarity, so it has no
    dependency on the classifier instance.
    """
    if not chat_history:
        return question

    if not _DEPENDENCY_PATTERN.search(question):
        log.info("no dependency signal — skipping rewrite: '%s'", question[:60])
        return question

    recent_turns = chat_history[-_TURNS_TO_KEEP:]
    history_block = _format_history(recent_turns)

    if not history_block:
        return question

    prompt = _REWRITE_INSTRUCTION.format(history_block=history_block, question=question)

    rewritten = generate(prompt, generator).strip()
    rewritten = _sanitize(rewritten, fallback=question)

    log.info("rewrote '%s' → '%s'", question[:60], rewritten[:60])
    return rewritten


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _format_history(turns: list[dict]) -> str:
    lines = []
    for turn in turns:
        role    = turn.get("role", "")
        content = turn.get("content", "")
        if not content:
            continue
        speaker = "Pengguna" if role == "user" else "Asisten"
        lines.append(f"{speaker}: {content}")
    return "\n".join(lines)


def _sanitize(rewritten: str, fallback: str) -> str:
    """
    Guards against degenerate generations: empty output, or output that's
    wildly longer than the original (a sign the model answered the question
    instead of rewriting it, despite instructions).
    """
    if not rewritten:
        return fallback
    if len(rewritten) > 400:
        log.warning("rewrite output unusually long (%d chars) — falling back", len(rewritten))
        return fallback
    return rewritten.strip('"').strip()