from __future__ import annotations

import hashlib
from dataclasses import dataclass, field


@dataclass
class Chunk:
    text:        str
    token_count: int
    source:      str
    chunk_index: int
    is_kmk:      bool = False

    bab_code:    str | None = None
    standar_id:  str | None = None
    kelompok:    str | None = None

    ep_id:             str | None = None
    fungsi_pelayanan:  str | None = None
    doc_type:          str | None = None
    nama_berkas:       str | None = None
    deskripsi:         str | None = None
    is_signed:         str | None = None

    vector: list[float] | None = field(default=None, repr=False)

    @property
    def payload(self) -> dict:
        return {
            "text":             self.text,
            "source":           self.source,
            "chunk_index":      self.chunk_index,
            "token_count":      self.token_count,
            "is_kmk":           self.is_kmk,
            "bab_code":         self.bab_code,
            "standar_id":       self.standar_id,
            "kelompok":         self.kelompok,
            "ep_id":            self.ep_id,
            "fungsi_pelayanan": self.fungsi_pelayanan,
            "doc_type":         self.doc_type,
            "nama_berkas":      self.nama_berkas,
            "deskripsi":        self.deskripsi,
            "is_signed":        self.is_signed,
        }

    @property
    def chunk_id(self) -> str:
        return hashlib.sha256(self.text.encode()).hexdigest()