"""
generation/intent_extractor.py
"""
from __future__ import annotations

import re
import numpy as np

from app.repositories.ingestion.embedder import embedder, EmbedderModel, embed_query
from app.core.logger import get_logger

log = get_logger("intent_extractor")

# ---------------------------------------------------------------------------
# Label sentences — richer = better signal
# ---------------------------------------------------------------------------

_PRIMARY_INTENTS = {
    "requirement_lookup", "evidence_check",
    "gap_analysis", "inventory",
    "general", "ui_navigation",
}

_NAV_DEPTH_PRIORITY = ["nav_to_document", "nav_to_assessment", "nav_to_standard", "nav_to_service"]

_INTENT_LABELS: dict[str, list[str]] = {
    "requirement_lookup": [
        "apa bunyi standar SKP 2 secara lengkap",
        "jelaskan isi elemen penilaian EP 2 untuk standar PPK 3",
        "apa yang diatur dalam KMK tentang pelayanan pasien",
        "bacakan ketentuan standar TKRS 3.2",
        "apa persyaratan regulasi yang tercantum dalam standar MKI",
        "bagaimana bunyi standar PMKP 4.1 menurut KARS",
        "apa definisi yang digunakan dalam standar PAP untuk identifikasi pasien",
        "tuliskan teks lengkap elemen penilaian 2 untuk standar PP 3",
        "apa ketentuan normatif yang berlaku untuk bab KPS",
        "sebutkan isi KMK yang mengatur asesmen pasien",
    ],

    "evidence_check": [
        "apakah sudah ada bukti untuk standar AP 1.1",
        "dokumen apa yang sudah kami upload untuk elemen penilaian ini",
        "cek apakah sertifikat pelatihan staf sudah masuk sistem untuk bab SKP",
        "sudahkah ada berkas untuk EP 3 standar TKRS",
        "apakah laporan audit sudah diunggah untuk kelompok ini",
        "berkas apa saja yang sudah tersedia untuk elemen penilaian 2",
        "apakah bukti observasi sudah ada di sistem untuk standar Prognas 1.1",
        "sudah ada lampiran apa untuk bab PAP standar 2",
        "apakah panduan pelayanan sudah tersimpan untuk standar TKRS 3.2",
        "cek dokumen yang sudah kami miliki untuk elemen penilaian ini",
        "apakah bukti untuk elemen penilaian ini sudah ada di sistem",
        "dokumen yang sudah diunggah untuk EP ini apa saja",
        "apa saja bukti yang sudah tersedia untuk standar ini",
    ],

    "gap_analysis": [
        "standar mana saja yang belum memiliki bukti sama sekali",
        "elemen penilaian apa yang masih kosong dokumennya",
        "apa yang masih kurang untuk memenuhi standar PAP 1.1",
        "bab mana yang paling banyak kekurangan dokumen",
        "dokumen apa yang belum kami kumpulkan untuk elemen ini",
        "tunjukkan gap antara persyaratan dan bukti yang ada",
        "standar apa yang belum terpenuhi di kelompok TKRS",
        "elemen penilaian mana yang tidak punya satupun bukti",
        "apa yang masih harus kami lengkapi sebelum akreditasi",
        "berapa banyak standar yang belum ada evidence-nya",
        "kita masih kurang dokumen apa untuk standar ini",
        "apa saja yang masih belum lengkap untuk elemen penilaian ini",
        "dokumen apa yang belum kami unggah untuk standar ini",
    ],

    "inventory": [
        "tampilkan semua dokumen yang sudah tersimpan di sistem",
        "berikan daftar lengkap seluruh berkas yang diupload",
        "rekap semua file yang ada di database kami",
        "list semua dokumen akreditasi yang tersimpan",
        "berapa total berkas yang sudah ada di sistem",
        "tampilkan keseluruhan dokumen yang sudah diinput tim",
        "apa saja dokumen yang sudah ada tanpa filter apapun",
        "tampilkan inventaris lengkap semua file akreditasi",
        "sebutkan seluruh berkas yang sudah diunggah ke sistem",
        "berikan gambaran umum semua dokumen yang tersimpan",
    ],

    "general": [
        "halo selamat pagi",
        "terima kasih atas bantuannya",
        "apa itu akreditasi rumah sakit",
        "siapa yang menerbitkan standar KARS",
        "kapan jadwal visitasi akreditasi kami",
        "apa perbedaan antara standar dan elemen penilaian",
        "bagaimana cara menggunakan fitur upload dokumen",
        "tolong bantu saya memahami sistem ini",
        "apa fungsi dari aplikasi manajemen akreditasi ini",
        "siapa yang bertanggung jawab mengelola dokumen akreditasi",
    ],

    "ui_navigation": [
        "buka halaman penyimpanan dokumen",
        "navigasi ke bagian berkas bab TKRS",
        "arahkan saya ke halaman upload",
        "filter tampilan berdasarkan standar AP",
        "pergi ke halaman dokumen untuk kelompok ini",
        "bawa saya ke bagian berkas elemen penilaian ini",
        "buka filter standar dan pilih MKI",
        "aktifkan filter untuk fungsi pelayanan ini",
        "arahkan ke halaman storage dan filter berdasarkan elemen penilaian",
        "tampilkan halaman dengan berkas yang sudah difilter berdasarkan bab ini",
    ],

    # nav depth — vocabulary must stay distinct from ui_navigation and each other
    "nav_to_service": [
        "filter halaman berdasarkan kelompok layanan",
        "saring tampilan menurut fungsi pelayanan",
        "pilih kelompok manajemen rumah sakit di filter",
        "set filter layanan ke TKRS",
        "terapkan filter fungsi pelayanan",
        "atur filter kelompok layanan saja",
        "cukup filter sampai level layanan",
        "saya hanya perlu filter layanannya",
        "pilih layanan MFK di dropdown",
        "terapkan filter kelompok pelayanan pasien",
        "set layanan ke KPS saja",
        "filter berdasarkan kelompok layanan cukup",
        "pilihin layanan ini di dropdown filter",
        "set filter menuju kelompok layanan saja",
        "berhenti di filter layanan tidak perlu lanjut",
        "pilih fungsi layanan yang relevan di filter",
    ],
    "nav_to_standard": [
        "filter halaman sampai level standar",
        "terapkan filter standar akreditasi",
        "set filter standar ke TKRS 1",
        "pilih standar MFK 11 di halaman",
        "saring dokumen berdasarkan nomor standar",
        "atur filter standar saja tidak perlu sampai dokumen",
        "cukup sampai filter standar",
        "terapkan filter standar dan berhenti di situ",
        "pilih standar yang relevan di dropdown",
        "buka PAB 2",
        "apa saja yang terdapat dalam PPI 1.1",
        "set filter ke standar AP 1.1",
        "saring halaman berdasarkan kode standar",
        "filter standar cukup tidak perlu buka filenya",
    ],
    "nav_to_assessment": [
        "buka elemen penilaian 5",
        "filter halaman sampai level elemen penilaian",
        "terapkan filter assessment untuk EP ini",
        "set filter penilaian ke MFK 11.e",
        "pilih elemen penilaian di halaman storage",
        "saring berdasarkan kode elemen penilaian",
        "atur filter sampai level assessment saja",
        "cukup sampai filter elemen penilaian",
        "terapkan filter EP dan berhenti di situ",
        "pilih assessment yang relevan di dropdown",
        "set filter ke elemen penilaian SKP 1.a",
        "saring halaman berdasarkan EP tertentu",
        "filter assessment cukup tidak perlu buka dokumennya",
    ],
    "nav_to_document": [
        "langsung buka filenya",
        "buka berkasnya sekarang",
        "pergi langsung ke dokumennya",
        "akses langsung file laporan pelatihan",
        "buka langsung SPO yang dimaksud",
        "tampilkan isi dokumennya sekarang",
        "saya mau langsung ke filenya",
        "buka dokumennya bukan filternya",
        "akses berkas SK Direktur langsung",
        "langsung ke filenya jangan filter dulu",
        "buka laporan yang dimaksud sekarang",
        "akses langsung dokumen buktinya",
    ],
}

# ---------------------------------------------------------------------------
# Lexical tie-breakers
#
# Embeddings are weak at exactly two things our label set depends on:
#   1. completion polarity ("sudah ada" vs "belum ada") — evidence_check and
#      gap_analysis share nearly all their vocabulary and differ mainly here.
#   2. closed-vocabulary depth markers ("layanan" / "elemen penilaian" /
#      "standar") — these are crisp discriminators, no reason to leave them
#      to cosine similarity alone.
# Both are applied as narrow, gated nudges — never as a replacement for the
# centroid scores — so they can only break ties between genuine contenders.
# ---------------------------------------------------------------------------

_LACK_MARKERS = re.compile(r'\b(belum|masih\s+kurang|tidak\s+ada)\b', re.IGNORECASE)
# NOTE: bare "kurang" deliberately excluded — it's too generic (e.g. "kurang
# jelas" = "not clear enough", a quality complaint, not an evidence-gap
# statement) and none of the gap_analysis training sentences use it without
# a "masih" prefix, which is already covered by "masih\s+kurang" above.

_HAVE_MARKERS = re.compile(r'\b(sudah|udah|telah)\b', re.IGNORECASE)
_POLARITY_BONUS = 0.03  # ~2x the observed correct/incorrect margin overlap
_POLARITY_CONTENTION_WINDOW = 0.02  # how close to the top score counts as "in contention"

# "X sudah Y belum?" is the Indonesian yes/no tag-question construction
# ("has it been uploaded or not?") — the trailing "belum" is a question tag,
# not a stated absence, so it must NOT trip _LACK_MARKERS. Checked before
# _LACK_MARKERS whenever both words are present. Matches "udah" as well as
# "sudah" (colloquial contraction, very common in real queries) and doesn't
# require "belum" to be the last word — real phrasing often has trailing
# clauses ("... belum ya", "... belum di sistem").
_TAG_QUESTION_RE = re.compile(r'\b(sudah|udah)\b.*\bbelum\b', re.IGNORECASE)

# Shared "bulk/listing action" signal — used to distinguish "show/open ALL
# X" from a single-item lookup, in two separate decisions:
#   1. _apply_polarity_bonus — skip the evidence_check/gap_analysis nudge,
#      since sudah/udah here is incidental phrasing ("...yang sudah
#      tersimpan"), not an evidence-completeness signal.
#   2. _infer_nav_depth's nav_to_document anchor — "buka semua berkas..." is
#      a listing/filter action, not opening one specific file.
# Kept as a single regex so the two decisions can't silently diverge on
# which words count as "bulk" (they previously used two different lists).
_BULK_ACTION_RE = re.compile(r'\b(semua|seluruh|total|keseluruhan|database)\b', re.IGNORECASE)

# Greeting/thanks openers are only treated as a hard "general" short-circuit
# when the ENTIRE message is essentially just the greeting — not merely
# when it starts with one. "halo, tolong buka EP 5 untuk standar TKRS" is a
# real request wearing a greeting, not a greeting.
_GREETING_PREFIX_RE = re.compile(
    r'^\s*(halo+|hai+|hi|hello|selamat\s+(pagi|siang|sore|malam)|apa\s+kabar|terima\s?kasih|makasih)\b',
    re.IGNORECASE,
)
_GREETING_TAIL_WORD_LIMIT = 3  # "halo Claude apa kabar" (3 trailing words) is
                                # still just a greeting; longer remainders are
                                # treated as a real request instead.


def _is_greeting_only(question: str) -> bool:
    m = _GREETING_PREFIX_RE.match(question)
    if not m:
        return False
    remainder = question[m.end():].strip(" ,.!?\n\t")
    return len(remainder.split()) <= _GREETING_TAIL_WORD_LIMIT


# Explicit navigation imperatives ("arahkan", "navigasi", ...) are almost
# exclusively used in our domain to mean "take me to X" — none of the other
# label sets use this vocabulary — so treat them as a near-unambiguous
# ui_navigation signal rather than leaving it to raw cosine similarity.
_NAV_VERB_RE = re.compile(r'\b(arahkan|navigasi|bawa\s+saya|pergi\s+ke|pindah\s+ke|buka\s)\b', re.IGNORECASE)
_NAV_VERB_BONUS = 0.02

_DEPTH_ANCHORS: list[tuple[str, re.Pattern]] = [
    ("nav_to_document", re.compile(
        r'\b(langsung|buka|akses)\b.*\b(file|berkas|dokumen)\w*\b'
        r'|\b(file|berkas|dokumen)\w*\b.*\b(langsung|buka|akses)\b',
        re.IGNORECASE,
    )),
    ("nav_to_assessment", re.compile(r'\belemen\s+penilaian\b|\bEP\s?\d', re.IGNORECASE)),
    ("nav_to_standard", re.compile(r'\bstandar\b', re.IGNORECASE)),
    ("nav_to_service", re.compile(r'\b(layanan|kelompok)\b', re.IGNORECASE)),
]
# NOTE: order matters and must match the "deepest wins" philosophy already
# encoded in _NAV_DEPTH_PRIORITY (document > assessment > standard > service)
# — when a question anchors on more than one depth, resolve to the more
# specific one. standard and service were previously transposed here, which
# made "pilih layanan MFK terus filter ke standar MFK 11" wrongly resolve to
# service instead of standard.


class IntentClassifier:
    """
    Embeds label sentences once at init, then classifies queries
    by picking the label whose centroid is closest to the query vector.
    """

    def __init__(self, embedder: EmbedderModel) -> None:
        self._embedder = embedder
        self._centroids = self._build_centroids()
        log.info("intent classifier ready — %d classes", len(self._centroids))

    def _build_centroids(self) -> dict[str, np.ndarray]:
        centroids = {}
        for label, sentences in _INTENT_LABELS.items():
            vecs = np.array(self._embedder.infer(sentences))
            centroids[label] = vecs.mean(axis=0)
            # re-normalise the centroid
            norm = np.linalg.norm(centroids[label])
            centroids[label] /= max(norm, 1e-9)
        return centroids

    def _all_scores(self, question: str) -> dict[str, float]:
        q_vec = np.array(embed_query(question, self._embedder))
        return {label: float(np.dot(q_vec, centroid)) for label, centroid in self._centroids.items()}

    @staticmethod
    def _apply_polarity_bonus(scores: dict[str, float], question: str) -> dict[str, float]:
        """Only nudges when evidence_check or gap_analysis is genuinely close
        to the top score — proximity-based, not strict top-2 rank, since a
        third label (e.g. requirement_lookup) can otherwise crowd one of the
        pair out of literal top-2 while it's still a real contender."""
        if "evidence_check" not in scores or "gap_analysis" not in scores:
            return scores
        if _BULK_ACTION_RE.search(question):
            # "semua/total/seluruh/..." signals a listing action, not a
            # check against one specific EP/standard — sudah/belum here is
            # incidental phrasing, not the evidence-completeness signal
            # this bonus is meant to disambiguate.
            return scores

        best = max(scores.values())
        in_contention = (
            best - scores["evidence_check"] <= _POLARITY_CONTENTION_WINDOW
            or best - scores["gap_analysis"] <= _POLARITY_CONTENTION_WINDOW
        )
        if not in_contention:
            return scores

        if _TAG_QUESTION_RE.search(question):
            scores["evidence_check"] += _POLARITY_BONUS
        elif _LACK_MARKERS.search(question):
            scores["gap_analysis"] += _POLARITY_BONUS
        elif _HAVE_MARKERS.search(question):
            scores["evidence_check"] += _POLARITY_BONUS
        return scores

    def _adjusted_scores(self, question: str) -> dict[str, float]:
        """Primary-intent scores after all tie-break heuristics are applied —
        the same computation classify() uses internally to pick a label.
        Exposed as its own method so callers (and eval/diagnostic tooling)
        can inspect exactly what drove a decision, instead of only raw
        cosine similarity which can silently diverge once heuristics like
        these are added."""
        all_scores = self._all_scores(question)
        scores = {label: score for label, score in all_scores.items() if label in _PRIMARY_INTENTS}

        # Fold nav-depth centroid strength into ui_navigation. classify() only
        # sees _PRIMARY_INTENTS, so without this, a query with strong domain
        # vocabulary ("elemen penilaian", "standar TKRS") can lose to
        # requirement_lookup even when a nav_to_* label scored highest overall.
        nav_depth_best = max(
            (all_scores[l] for l in _NAV_DEPTH_PRIORITY if l in all_scores),
            default=None,
        )
        if nav_depth_best is not None and "ui_navigation" in scores:
            scores["ui_navigation"] = max(scores["ui_navigation"], nav_depth_best)

        if _NAV_VERB_RE.search(question) and "ui_navigation" in scores:
            scores["ui_navigation"] += _NAV_VERB_BONUS

        return self._apply_polarity_bonus(scores, question)

    def classify(self, question: str) -> str:
        if _is_greeting_only(question):
            log.info("intent scores: skipped (greeting-only short-circuit) → general")
            return "general"

        scores = self._adjusted_scores(question)
        best = max(scores, key=scores.__getitem__)
        log.info("intent scores: %s → %s", {k: f"{v:.3f}" for k, v in scores.items()}, best)
        return best

    def classify_multi(self, question: str, threshold: float = 0.02) -> list[str]:
        if _is_greeting_only(question):
            log.info("multi-intent scores: skipped (greeting-only short-circuit) → active=['general']")
            return ["general"]

        # Deliberately uses raw _all_scores (primary + nav_to_* labels
        # together), not _adjusted_scores — classify_multi needs the
        # individual nav_to_* label scores intact so _infer_nav_depth can
        # rank them; _adjusted_scores folds them into ui_navigation instead.
        scores = self._apply_polarity_bonus(self._all_scores(question), question)

        best_score = max(scores.values())

        active = [
            label for label, score in scores.items()
            if best_score - score <= threshold
        ]
        active.sort(key=lambda l: scores[l], reverse=True)

        log.info(
            "multi-intent scores: %s → active=%s",
            {k: f"{v:.3f}" for k, v in scores.items()},
            active,
        )
        return active if active else [max(scores, key=scores.__getitem__)]


classifier = IntentClassifier(embedder)


def extract_intent(question: str, classifier: IntentClassifier) -> dict:
    if _is_greeting_only(question):
        log.info("intent: greeting-only short-circuit → general")
        return {
            "query_type": "general",
            "all_intents": ["general"],
            "standar": None,
            "fungsi_pelayanan": None,
            "element_penilaian": None,
            "nav_depth": "standard",
            "filters": {},
        }

    query_type       = classifier.classify(question)
    all_intents      = classifier.classify_multi(question)
    standar_match    = re.search(r'\b([A-Z]{2,5}\.?\s?\d+(?:\.[\d]+)?)\b', question, re.IGNORECASE)
    fungsi_match     = re.search(r'\b(TKRS|KPS|MFK|PMKP|MRMIK|PPI|PPK|AKP|HPK|PP|PAP|PAB|PKPO|KE|SKP|PROGNAS)\b', question, re.IGNORECASE)
    ep_match         = re.search(r'\belemen\s+penilaian\s+(\d+|[a-z])\b', question, re.IGNORECASE)

    return {
        "query_type":      query_type,
        "all_intents":     all_intents,
        "standar":         standar_match.group(1).upper() if standar_match else None,
        "fungsi_pelayanan": fungsi_match.group(1).upper() if fungsi_match else None,
        "element_penilaian": ep_match.group(1) if ep_match else None,
        "nav_depth":       _infer_nav_depth(all_intents, question),
        "filters":         {},
    }


def _infer_nav_depth(all_intents: list[str], question: str = "") -> str:
    """
    Depth ordering combines two signals:
      1. Lexical anchors (closed vocabulary: 'layanan', 'elemen penilaian',
         'standar', 'langsung ke file') — crisp discriminators the embedder
         is weak on, so these take priority when present.
      2. Fallback to whichever nav_to_* label classify_multi ranked highest
         (original behavior), used when no anchor matches or question isn't
         supplied by the caller.

    `question` is optional and defaults to "" for backward compatibility with
    existing callers that only pass `all_intents` — but anchor-based
    resolution only runs when a question string is actually provided.
    """
    active_nav = set(all_intents) & set(_NAV_DEPTH_PRIORITY)

    if question:
        for label, pattern in _DEPTH_ANCHORS:
            if label not in active_nav:
                continue
            if label == "nav_to_document" and _BULK_ACTION_RE.search(question):
                continue  # "buka semua berkas..." is a listing action, not opening one file
            if pattern.search(question):
                return label.replace("nav_to_", "")

    # all_intents is already sorted by descending score in classify_multi() —
    # trust that order instead of re-imposing a separate fixed priority.
    for label in all_intents:
        if label in active_nav:
            return label.replace("nav_to_", "")

    return "standard"