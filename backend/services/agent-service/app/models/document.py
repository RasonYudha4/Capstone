from __future__ import annotations

from dataclasses import dataclass, field
from enum import Enum


class SignatureStatus(str, Enum):
    SIGNED   = "signed"
    PARTIAL  = "partial"
    UNSIGNED = "unsigned"
    UNKNOWN  = "unknown"


@dataclass
class ParsedDocument:
    source:           str
    file_type:        str
    raw_text:         str
    pages:            list[str] = field(default_factory=list)
    total_pages:      int = 0
    signature_status: SignatureStatus = SignatureStatus.UNKNOWN