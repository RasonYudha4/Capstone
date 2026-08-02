"""
repositories/generation/gap_inventory_analysis.py — gap analysis and document
inventory reporting against the Chroma store.

Extracted from services/rag_service.py so that store-driven reporting logic
lives in the repository layer alongside the rest of the retrieval/store code.
"""
from __future__ import annotations

from collections import defaultdict

from app.core.store.chromadb import ChromaStore, _build_chroma_where
from app.repositories.generation.generator import generator, generate
from app.repositories.generation.prompt_builder import build_gap_prompt
from app.core.logger import get_logger

log = get_logger("gap_inventory_analysis")


def run_gap_analysis(intent: dict) -> str:
    """Compare KMK (requirement) standar_codes against evidence-covered ones
    and summarize what's missing for the given (optional) fungsi_pelayanan."""
    store = ChromaStore()

    extra = {"fungsi_pelayanan": intent["fungsi_pelayanan"]} if intent.get("fungsi_pelayanan") else {}

    all_ep     = set(store.get_all_unique_values("standar_code", {"is_kmk": True,  **extra}))
    covered_ep = set(store.get_all_unique_values("standar_code", {"is_kmk": False, **extra}))
    missing    = sorted(all_ep - covered_ep)

    prompt = build_gap_prompt(missing, covered_ep, intent)
    return generate(prompt, generator)


def run_inventory(intent: dict) -> str:
    """Summarize stored documents grouped by fungsi_pelayanan → standar,
    optionally scoped by intent filters."""
    store   = ChromaStore()
    filters = {}
    if intent.get("fungsi_pelayanan"):
        filters["fungsi_pelayanan"] = intent["fungsi_pelayanan"]
    if intent.get("standar"):
        filters["standar"] = intent["standar"]

    # NOTE: reaches into ChromaStore's private `_col` handle, same as the
    # original implementation. Worth promoting to a public `store.get(...)`
    # method at some point, but left as-is here to keep this a pure move.
    results = store._col.get(
        where=_build_chroma_where(filters) if filters else None,
        include=["metadatas"],
    )

    # Deduplicate by file name first
    seen  = set()
    files = []
    for meta in results["metadatas"]:
        name = meta.get("nama_berkas") or meta.get("source")
        if name and name not in seen:
            seen.add(name)
            files.append(meta)

    if not files:
        return "Belum ada dokumen yang tersimpan dalam sistem."

    grouped: dict[str, dict[str, int]] = defaultdict(lambda: defaultdict(int))
    for m in files:
        fp  = m.get("fungsi_pelayanan") or "Tidak Diketahui"
        std = m.get("standar") or "-"
        grouped[fp][std] += 1

    lines = []
    total = 0
    for fp, standards in sorted(grouped.items()):
        lines.append(f"\n**{fp}**")
        for std, count in sorted(standards.items()):
            lines.append(f"  - {std}: {count} dokumen")
            total += count

    return f"Dokumen yang tersedia ({total} berkas):\n" + "\n".join(lines)