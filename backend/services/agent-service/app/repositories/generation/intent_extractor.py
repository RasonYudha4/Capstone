# generation/intent_extractor.py

from __future__ import annotations

import re
import numpy as np

from app.repositories.ingestion.embedder import EmbedderModel, embed_query
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
        "tampilkan semua dokumen MFK",           
        "buka semua berkas untuk layanan ini",   
        "lihat semua file kelompok ini",         
        "tampilkan seluruh dokumen layanan MFK",
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
        "apa saja yang terdapat dalam PPI 1.1"
        "set filter ke standar AP 1.1",
        "saring halaman berdasarkan kode standar",
        "filter standar cukup tidak perlu buka filenya",
    ],
    "nav_to_assessment": [
        "buka elemen penilaian 5"
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

    def classify(self, question: str) -> str:
        q_vec = np.array(embed_query(question, self._embedder))
        scores = {
            label: float(np.dot(q_vec, centroid))
            for label, centroid in self._centroids.items()
            if label in _PRIMARY_INTENTS
        }
        best = max(scores, key=scores.__getitem__)
        log.info("intent scores: %s → %s", {k: f"{v:.3f}" for k, v in scores.items()}, best)
        return best
    
    def classify_multi(self, question: str, threshold: float = 0.02) -> list[str]:
        q_vec = np.array(embed_query(question, self._embedder))
        scores = {
            label: float(np.dot(q_vec, centroid))
            for label, centroid in self._centroids.items()
        }

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


def extract_intent(question: str, classifier: IntentClassifier) -> dict:
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
        "nav_depth":       _infer_nav_depth(all_intents),
        "filters":         {},
    }

def _infer_nav_depth(all_intents: list[str]) -> str:
    for depth_intent in _NAV_DEPTH_PRIORITY:
        if depth_intent in all_intents:
            return depth_intent.replace("nav_to_", "")
    return "standard"