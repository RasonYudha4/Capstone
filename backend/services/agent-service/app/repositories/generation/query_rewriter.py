from __future__ import annotations

import re

from app.repositories.generation.generator import generator, generate
from app.core.logger import get_logger
from app.core.config import settings

log = get_logger("query_rewriter")

# ---------------------------------------------------------------------------
# Config
# ---------------------------------------------------------------------------

_DEPENDENCY_PATTERN = re.compile(
    r"\b(tersebut|itu|ini|tadi|tersisa|yang (lain|sebelumnya|kedua|pertama))\b"
    r"|^\s*(dan|atau|bagaimana dengan|gimana dengan|kalau|terus|lalu)\b",
    re.IGNORECASE,
)

# Elliptical follow-ups ("Berapa lama?", "Sejak kapan?") often carry no
# pronoun at all, so the regex above misses them. A short question with no
# explicit dependency marker is still very likely to be context-bound —
# this is a second, wider-recall signal that catches what the regex can't.
_SHORT_QUESTION_WORD_THRESHOLD = 6

# Assistant turns (full RAG answers) are the main source of noise in the
# rewrite prompt relative to short user questions — truncate to a gist so
# they still convey "what was being discussed" without dominating the
# prompt or dragging in off-topic detail.
_MAX_ASSISTANT_TURN_CHARS = 240

_REWRITE_INSTRUCTION = """Kamu bertugas menulis ulang pertanyaan terbaru pengguna agar berdiri sendiri \
(self-contained), dengan mengganti kata ganti atau rujukan implisit (seperti "itu", "tersebut", \
"ini", "tadi") menggunakan informasi spesifik dari riwayat percakapan di bawah.

ATURAN:
- Jangan menjawab pertanyaannya. Tugasmu HANYA menulis ulang pertanyaan tersebut.
- Jika pertanyaan sudah jelas berdiri sendiri tanpa riwayat, kembalikan APA ADANYA tanpa perubahan.
- Jangan menambahkan informasi yang tidak ada di riwayat atau pertanyaan asli.
- Jika riwayat tidak relevan dengan pertanyaan terbaru, ABAIKAN riwayat dan kembalikan pertanyaan apa adanya.
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

    Dependency detection uses two signals (either triggers a rewrite call):
      1. explicit referential/connective words (pronouns, "bagaimana dengan", etc.)
      2. short questions with no dependency markers — these are usually elliptical
         follow-ups ("Berapa lama?", "Sejak kapan?") that carry no pronoun but are
         still context-bound. Long, self-contained questions rarely need history.

    Note: this is intentionally independent of IntentClassifier — rewriting
    needs generation, not embedding-centroid similarity, so it has no
    dependency on the classifier instance.
    """
    if not chat_history:
        return question

    if not _looks_dependent(question):
        log.info("no dependency signal — skipping rewrite: '%s'", question[:60])
        return question

    history_block = _format_history(chat_history)
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

def _looks_dependent(question: str) -> bool:
    if _DEPENDENCY_PATTERN.search(question):
        return True
    word_count = len(question.split())
    return word_count <= _SHORT_QUESTION_WORD_THRESHOLD


def _format_history(turns: list[dict]) -> str:
    """
    Formats all turns currently held in the session buffer (bounded upstream
    by settings.rewrite_turns_to_keep — see conversation_store._Session).
    Assistant turns are truncated to a gist so a long RAG answer doesn't
    dominate the prompt or dilute the signal from the actual user turns.
    """
    lines = []
    for turn in turns:
        role    = turn.get("role", "")
        content = turn.get("content", "")
        if not content:
            continue
        speaker = "Pengguna" if role == "user" else "Asisten"
        if role == "assistant":
            content = _truncate(content, _MAX_ASSISTANT_TURN_CHARS)
        lines.append(f"{speaker}: {content}")
    return "\n".join(lines)


def _truncate(text: str, max_chars: int) -> str:
    if len(text) <= max_chars:
        return text
    return text[:max_chars].rstrip() + "…"


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