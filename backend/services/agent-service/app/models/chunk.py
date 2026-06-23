from __future__ import annotations

import hashlib
from dataclasses import dataclass, field


@dataclass
class Chunk:
    text:                   str
    token_count:            int
    source:                 str
    chunk_index:            int
    is_kmk:                 bool = False

    standar:                str | None = None
    kelompok:               str | None = None

    element_penilaian:      str | None = None
    fungsi_pelayanan:       str | None = None
    doc_type:               str | None = None
    nama_berkas:            str | None = None
    deskripsi:              str | None = None
    is_signed:              str | None = None

    standar_code:           str | None = None
    element_penilaian_code: str | None = None 

    service_id:             str | None = None
    standard_id:            str | None = None
    assessment_id:          str | None = None
    document_id:            str | None = None

    vector: list[float] | None = field(default=None, repr=False)

    @property
    def payload(self) -> dict:
        return {
            "text":                     self.text,
            "source":                   self.source,
            "chunk_index":              self.chunk_index,
            "token_count":              self.token_count,
            "is_kmk":                   self.is_kmk,
            "standar":                  self.standar,
            "kelompok":                 self.kelompok,
            "element_penilaian":        self.element_penilaian,
            "fungsi_pelayanan":         self.fungsi_pelayanan,
            "doc_type":                 self.doc_type,
            "nama_berkas":              self.nama_berkas,
            "deskripsi":                self.deskripsi,
            "is_signed":                self.is_signed,
            "standar_code":             self.standar_code,
            "element_penilaian_code":   self.element_penilaian_code,
            "service_id":               self.service_id,
            "standard_id":              self.standard_id,
            "assessment_id":            self.assessment_id,
            "document_id":              self.document_id,
        }

    @property
    def chunk_id(self) -> str:
        return hashlib.sha256(self.text.encode()).hexdigest()