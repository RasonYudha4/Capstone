import { z } from "zod"

// ─────────────────────────────────────────────
// Existing schemas (unchanged)
// ─────────────────────────────────────────────

export const assessmentOptionSchema = z.object({
    id:          z.string().uuid(),
    code:        z.string(),
    description: z.string(),
})

export const standardOptionSchema = z.object({
    id:          z.string().uuid(),
    code:        z.string(),
    description: z.string(),
    assessments: z.array(assessmentOptionSchema),
})

export const serviceOptionSchema = z.object({
    id:          z.string().uuid(),
    code:        z.string(),
    description: z.string(),
    standards:   z.array(standardOptionSchema),
})

// ─────────────────────────────────────────────
// New: Document Type
// ─────────────────────────────────────────────

export const documentTypeOptionSchema = z.object({
    id:          z.string().uuid(),
    name:        z.string(),
    description: z.string(),
})

// ─────────────────────────────────────────────
// Updated response schema
// ─────────────────────────────────────────────

export const formOptionsResponseSchema = z.object({
    data: z.object({
        services:       z.array(serviceOptionSchema),
        document_types: z.array(documentTypeOptionSchema),
    }),
})

// ─────────────────────────────────────────────
// Inferred types
// ─────────────────────────────────────────────

export type AssessmentOption    = z.infer<typeof assessmentOptionSchema>
export type StandardOption      = z.infer<typeof standardOptionSchema>
export type ServiceOption       = z.infer<typeof serviceOptionSchema>
export type DocumentTypeOption  = z.infer<typeof documentTypeOptionSchema>
export type FormOptionsResponse = z.infer<typeof formOptionsResponseSchema>