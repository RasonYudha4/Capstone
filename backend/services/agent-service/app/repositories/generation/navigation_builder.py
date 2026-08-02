"""
repositories/generation/navigation_builder.py — builds UI <command> blocks
from retrieval results / intent, and decides when navigation should fire.

Extracted from services/rag_service.py so that all "turn retrieval results
into UI navigation commands" logic lives in one place in the repository
layer, separate from query orchestration.
"""
from __future__ import annotations

# Shared with rag_service.py, which uses this to detect/promote nav-depth
# intents before deciding whether to call build_navigation_block at all.
NAV_DEPTH_PRIORITY = ["nav_to_document", "nav_to_assessment", "nav_to_standard", "nav_to_service"]


def should_navigate(all_intents: list[str], results: list, min_score: float = 0.85) -> bool:
    """Only navigate if evidence_check is active AND results are genuinely relevant."""
    if "evidence_check" not in all_intents:
        return False
    if not results:
        return False
    return results[0].score >= min_score


def build_navigation_block(intent: dict, results: list, ctx=None) -> str:
    lines = []
    nav_depth = intent.get("nav_depth", "standard")

    if not ctx or ctx.current_path != "/storage":
        lines.append('<command>{"type":"NAVIGATE","path":"/storage"}</command>')

    evidence_results = [r for r in results if not r.is_kmk]
    top = evidence_results[0] if evidence_results else None

    # ── Service ──────────────────────────────────────────────────────────────
    service_id = (top.service_id if top else None)
    if not service_id and intent.get("fungsi_pelayanan") and ctx:
        matched = _match_service(intent, ctx)
        if matched:
            service_id = matched.id

    if service_id:
        lines.append(f'<command>{{"type":"SET_SERVICE_FILTER","serviceId":"{service_id}"}}</command>')

    if nav_depth == "service":
        lines.append("\n📂 Menampilkan layanan yang relevan...")
        return "\n".join(lines)

    # ── Standard ─────────────────────────────────────────────────────────────
    if top and top.standard_id:
        lines.append(f'<command>{{"type":"SET_STANDARD_FILTER","standardId":"{top.standard_id}"}}</command>')

    if nav_depth == "standard":
        label = (top.standar if top else None) or intent.get("standar") or "ini"
        lines.append(f"\n📂 Menampilkan dokumen untuk standar **{label}**...")
        return "\n".join(lines)

    # ── Assessment ───────────────────────────────────────────────────────────
    if top and top.assessment_id:
        lines.append(f'<command>{{"type":"SET_ASSESSMENT_FILTER","assessmentId":"{top.assessment_id}"}}</command>')

    if nav_depth == "assessment":
        lines.append("\n📂 Menampilkan assessment yang relevan...")
        return "\n".join(lines)

    # ── Document ─────────────────────────────────────────────────────────────
    if top and top.document_id:
        lines.append(f'<command>{{"type":"OPEN_DOCUMENT","documentId":"{top.document_id}"}}</command>')
        lines.append(f"\n📄 Membuka **{top.nama_berkas or 'dokumen'}** yang relevan...")
    elif top:
        lines.append(
            "\n⚠️ Dokumen ditemukan dalam indeks tetapi belum memiliki ID — "
            "pastikan metadata `document_id` disimpan saat ingestion."
        )
    else:
        if "evidence_check" in intent.get("all_intents", []):
            lines.append(
                "\n📭 Tidak ditemukan berkas bukti untuk standar ini. "
                "Silakan upload dokumen yang relevan."
            )

    return "\n".join(lines)


def _match_service(intent: dict, ctx) -> object | None:
    if not ctx or not ctx.available_services:
        return None

    search_terms = []
    if intent.get("fungsi_pelayanan"):
        search_terms.append(intent["fungsi_pelayanan"].upper())
    if intent.get("standar"):
        search_terms.append(intent["standar"].split()[0].upper())

    for term in search_terms:
        for svc in ctx.available_services:
            if term in svc.label.upper():
                return svc

    return None