"""
generation/intent_extractor.py — extract structured intent from a user question.

Pulls known standar_ids and bab_codes from the vector store (KMK chunks)
so the LLM chooses from values that actually exist, not from memory.

Returns a dict with:
    query_type : "requirement_lookup" | "evidence_check" | "gap_analysis" | "general"
    standar_id : e.g. "MFK 6" or None
    bab_code   : e.g. "MFK" or None
"""
from __future__ import annotations

import json

from app.repositories.generation.generator import generate
from app.core.store.chromadb import ChromaStore


def extract_intent(question: str) -> dict:
    store = ChromaStore()

    known_standards = store.get_all_unique_values("standar_id", {"is_kmk": True})
    known_bab_codes = store.get_all_unique_values("bab_code",   {"is_kmk": True})

    prompt = f"""
    ATURAN KETAT:
    - standar_id dan bab_code HANYA diisi jika disebutkan EKSPLISIT dalam pertanyaan.
    - Jika tidak ada kata yang cocok persis dengan daftar, kembalikan null.
    - Pertanyaan tentang sistem/file/daftar dokumen = query_type: "general"
    Kamu adalah asisten sistem akreditasi rumah sakit.
    Ekstrak intent dari pertanyaan berikut.

    BAB codes yang tersedia: {known_bab_codes}
    Standar IDs yang tersedia: {known_standards}

    Kembalikan HANYA JSON dengan field berikut (tanpa penjelasan tambahan):
    - query_type: salah satu dari:
        "requirement_lookup" — pengguna ingin tahu isi standar/persyaratan
        "evidence_check"     — pengguna ingin tahu bukti/dokumen yang sudah ada
        "gap_analysis"       — pengguna ingin tahu EP mana yang belum ada buktinya
        "general"            — pertanyaan umum yang tidak masuk kategori di atas
    - standar_id: standar ID spesifik dari daftar di atas, atau null jika tidak disebutkan
    - bab_code: BAB code dari daftar di atas, atau null jika tidak disebutkan

    Pertanyaan: {question}
    """

    raw = generate(prompt)

    try:
        # Strip markdown code fences if LLM wraps response in them
        clean = raw.strip().removeprefix("```json").removeprefix("```").removesuffix("```").strip()
        return json.loads(clean)
    except (json.JSONDecodeError, AttributeError):
        # Safe fallback — treat as general semantic search
        return {"query_type": "general", "standar_id": None, "bab_code": None}