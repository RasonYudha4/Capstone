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

_INTENT_LABELS: dict[str, list[str]] = {
    "requirement_lookup": [
        "apa isi standar akreditasi ini",
        "jelaskan persyaratan elemen penilaian",
        "apa ketentuan regulasi KMK untuk bab ini",
        "sebutkan kriteria akreditasi rumah sakit",
        "apa yang dimaksud dengan standar AP 1.1",
        "bagaimana persyaratan untuk memenuhi elemen penilaian ini",
        "jelaskan isi permenkes terkait akreditasi",
        "apa saja kriteria yang harus dipenuhi untuk bab TKRS",
        "sebutkan regulasi yang berlaku untuk standar ini",
        "apa ketentuan KMK mengenai pelayanan pasien",
        "bagaimana bunyi standar MKI 2.1",
        "apa yang diatur dalam elemen penilaian EP 3",
    ],
    "evidence_check": [
        "dokumen bukti apa yang sudah ada",
        "apakah file ini sudah diupload",
        "bukti apa yang sudah dimiliki untuk standar ini",
        "lampiran apa yang tersedia",
        "apakah dokumen pelatihan sudah ada di sistem",
        "sudah ada berkas untuk elemen penilaian ini",
        "cek apakah bukti akreditasi sudah tersimpan",
        "dokumen apa yang sudah kami kumpulkan untuk bab ini",
        "apakah laporan sudah diupload ke sistem",
        "tampilkan bukti yang tersedia untuk standar AP",
        "file apa saja yang sudah ada untuk kelompok ini",
        "apakah sertifikat pelatihan sudah masuk ke sistem",
    ],
    "gap_analysis": [
        "apa yang belum ada dalam dokumen kita",
        "standar mana yang belum terpenuhi",
        "dokumen apa yang masih kurang",
        "gap apa yang perlu dilengkapi",
        "elemen penilaian mana yang belum ada buktinya",
        "apa saja kekurangan dokumen untuk akreditasi",
        "standar apa yang belum memiliki evidence",
        "bab mana yang masih kosong dokumennya",
        "apa yang missing dari kelengkapan berkas kita",
        "berkas apa yang belum kami lengkapi untuk bab TKRS",
        "adakah elemen penilaian yang tidak memiliki bukti",
        "tunjukkan standar yang belum terpenuhi buktinya",
    ],
    "inventory": [
        "data apa saja yang kita miliki",
        "tampilkan semua dokumen yang tersimpan",
        "berikan daftar seluruh file yang ada",
        "list semua berkas dalam sistem",
        "file apa saja yang sudah tersimpan di sistem",
        "berikan rekap semua dokumen yang sudah diinput",
        "tampilkan seluruh berkas yang telah diupload",
        "apa saja dokumen yang sudah ada di database",
        "sebutkan semua file yang dimiliki sistem",
        "berikan daftar lengkap dokumen akreditasi yang tersimpan",
        "dokumen apa saja yang sudah masuk ke sistem",
        "tampilkan inventaris berkas yang tersedia",
    ],
    "general": [
        "apa itu akreditasi rumah sakit",
        "bagaimana cara kerja sistem ini",
        "tolong bantu saya",
        "terima kasih",
        "halo selamat pagi",
        "apa fungsi dari aplikasi ini",
        "bagaimana cara menggunakan fitur upload",
        "siapa yang bertanggung jawab atas akreditasi",
        "apa perbedaan antara standar dan elemen penilaian",
        "kapan jadwal akreditasi berikutnya",
        "apa itu KARS",
        "jelaskan proses akreditasi secara umum",
    ],
    "ui_navigation": [
        "tunjukkan dokumen untuk standar ini",
        "buka halaman storage",
        "navigasi ke berkas untuk bab TKRS",
        "tampilkan file untuk fungsi pelayanan ini",
        "bawa saya ke dokumen yang relevan",
        "pergi ke halaman penyimpanan",
        "filter berkas berdasarkan standar AP",
        "arahkan saya ke berkas yang ada",
        "buka filter untuk elemen penilaian ini",
        "tunjukkan cara menemukan dokumen tersebut",
        "bagaimana cara mengakses berkas itu",
        "dimana saya bisa lihat dokumen untuk bab ini",
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
        }
        best = max(scores, key=scores.__getitem__)
        log.info(
            "intent scores: %s → %s",
            {k: f"{v:.3f}" for k, v in scores.items()},
            best,
        )
        return best
    
    def classify_multi(self, question: str, threshold: float = 0.70) -> list[str]:
        q_vec = np.array(embed_query(question, self._embedder))
        scores = {
            label: float(np.dot(q_vec, centroid))
            for label, centroid in self._centroids.items()
        }
        # Return all labels that pass the threshold, sorted by score descending
        active = [
            label for label, score in scores.items()
            if score >= threshold
        ]
        active.sort(key=lambda l: scores[l], reverse=True)

        log.info(
            "multi-intent scores: %s → active=%s",
            {k: f"{v:.3f}" for k, v in scores.items()},
            active,
        )
        return active if active else [max(scores, key=scores.__getitem__)]


def extract_intent(question: str, classifier: IntentClassifier) -> dict:
    query_type = classifier.classify(question)

    all_intents = classifier.classify_multi(question, threshold=0.70)
    standar_match = re.search(r'\b([A-Z]{2,5}\.?\s?\d+\.?\d*)\b', question)
    bab_match     = re.search(r'\bBAB\s+([A-Z]+|\d+)\b', question, re.IGNORECASE)

    return {
        "query_type": query_type,
        "all_intents": all_intents, 
        "standar":    standar_match.group(1) if standar_match else None,
        "bab_code":   bab_match.group(0).upper() if bab_match else None,
        "filters":    {},
    }