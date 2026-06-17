"""
Server-owned, in-memory conversation history.

Design intent: the client never constructs or sends `chat_history` content.
It only ever holds a `session_id` (opaque to the client) and sends that back
on each request. All role-stamping happens here, at write-time, based on
which code path wrote the turn — never inferred from anything the client claims.

This is intentionally a minimal in-memory store, not a queue/db-backed one.
Acceptable starting point for a single-process deployment; swap the dict for
Redis or a table later if you need multi-instance or durability across restarts.
"""
from __future__ import annotations

import time
import uuid
from collections import deque
from threading import Lock

from app.core.logger import get_logger

log = get_logger("conversation_store")

_CYCLES_TO_KEEP = 3                  # must match query_rewriter._CYCLES_TO_KEEP
_TURNS_TO_KEEP  = _CYCLES_TO_KEEP * 2
_SESSION_TTL_S  = 60 * 60 * 4        # 4h idle expiry — adjust to real usage patterns


class _Session:
    __slots__ = ("turns", "last_seen")

    def __init__(self) -> None:
        self.turns: deque[dict] = deque(maxlen=_TURNS_TO_KEEP)
        self.last_seen: float = time.time()


class ConversationStore:
    """Thread-safe, process-local session store."""

    def __init__(self) -> None:
        self._sessions: dict[str, _Session] = {}
        self._lock = Lock()

    def create_session(self) -> str:
        session_id = str(uuid.uuid4())
        with self._lock:
            self._sessions[session_id] = _Session()
        log.info("session created: %s", session_id)
        return session_id

    def append_turn(self, session_id: str, role: str, content: str) -> None:
        """
        role is supplied by the calling code path, never by client input.
        Callers in this codebase: the API layer stamps 'user' for the
        incoming question, the service layer stamps 'assistant' for
        whatever generate()/generate_stream() produced.
        """
        if role not in ("user", "assistant"):
            raise ValueError(f"invalid role: {role!r}")

        with self._lock:
            session = self._sessions.get(session_id)
            if session is None:
                session = _Session()
                self._sessions[session_id] = session
            session.turns.append({"role": role, "content": content})
            session.last_seen = time.time()

    def get_history(self, session_id: str) -> list[dict]:
        with self._lock:
            session = self._sessions.get(session_id)
            if session is None:
                return []
            session.last_seen = time.time()
            return list(session.turns)

    def sweep_expired(self) -> int:
        """Call periodically (e.g. a background task) to bound memory growth."""
        cutoff = time.time() - _SESSION_TTL_S
        with self._lock:
            expired = [sid for sid, s in self._sessions.items() if s.last_seen < cutoff]
            for sid in expired:
                del self._sessions[sid]
        if expired:
            log.info("swept %d expired session(s)", len(expired))
        return len(expired)


conversation_store = ConversationStore()