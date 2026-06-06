"""
generation/intent_extractor.py — extract structured intent from a user question.
"""
from __future__ import annotations

import re

_REQUIREMENT_KEYWORDS = {
    "standar", "persyaratan", "ketentuan", "regulasi", "kmk", "permenkes",
    "akreditasi", "kriteria", "elemen penilaian", "ep", "bab",
}

_EVIDENCE_KEYWORDS = {
    "bukti", "evidence", "dokumen", "file", "upload", "tersedia",
    "sudah ada", "dimiliki", "lampiran",
}

_GAP_KEYWORDS = {
    "belum", "kurang", "gap", "missing", "tidak ada", "kosong",
    "kekurangan", "apa yang belum",
}


def extract_intent(question: str, gen=None) -> dict:  # gen kept for backward compat
    q = question.lower()
    words = set(re.findall(r'\w+', q))

    if _GAP_KEYWORDS & (words | {q[i:j] for i in range(len(q)) for j in range(i+2, min(i+20, len(q)+1))}):
        query_type = "gap_analysis"
    elif _REQUIREMENT_KEYWORDS & words:
        query_type = "requirement_lookup"
    elif _EVIDENCE_KEYWORDS & words:
        query_type = "evidence_check"
    else:
        query_type = "general"

    # Extract standar_id pattern e.g. "AP 1.1", "MKI.2"
    standar_match = re.search(r'\b([A-Z]{2,5}\.?\s?\d+\.?\d*)\b', question)
    standar_id = standar_match.group(1) if standar_match else None

    # Extract bab_code pattern e.g. "BAB 1", "BAB IV"
    bab_match = re.search(r'\bBAB\s+([IVX]+|\d+)\b', question, re.IGNORECASE)
    bab_code = bab_match.group(0).upper() if bab_match else None

    return {
        "query_type": query_type,
        "standar_id": standar_id,
        "bab_code":   bab_code,
    }