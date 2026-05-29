import { z } from "zod"

// ─────────────────────────────────────────────
// Request schemas
// ─────────────────────────────────────────────

// POST /ingest/kmk — file only, no extra fields
export const ingestKmkBodySchema = z.object({})

// POST /ingest/evidence — multipart/form-data fields
export const ingestEvidenceBodySchema = z.object({
    kelompok:         z.string().min(1, "Kelompok is required"),
    fungsi_pelayanan: z.string().min(1, "Fungsi pelayanan is required"),
    standar_id:       z.string().min(1, "Standar ID is required"),
    ep_id:            z.string().min(1, "EP ID is required"),
    doc_type:         z.string().min(1, "Doc type is required"),
    nama_berkas:      z.string().min(1, "Nama berkas is required"),
    deskripsi:        z.string().optional().default(""),
})

// ─────────────────────────────────────────────
// Response schemas
// ─────────────────────────────────────────────

// POST /ingest/kmk
export const ingestKmkResponseSchema = z.object({
    status:          z.enum(["ok", "failed"]),
    filename:        z.string(),
    chunks_upserted: z.number().int().nonnegative(),
})

// POST /ingest/evidence
export const ingestEvidenceResponseSchema = z.object({
    status:          z.enum(["ok", "failed"]),
    filename:        z.string(),
    chunks_upserted: z.number().int().nonnegative(),
    ep_id:           z.string(),
    standar_id:      z.string(),
})

// ─────────────────────────────────────────────
// Inferred types
// ─────────────────────────────────────────────

export type IngestKmkBody          = z.infer<typeof ingestKmkBodySchema>
export type IngestEvidenceBody     = z.infer<typeof ingestEvidenceBodySchema>
export type IngestKmkResponse      = z.infer<typeof ingestKmkResponseSchema>
export type IngestEvidenceResponse = z.infer<typeof ingestEvidenceResponseSchema>