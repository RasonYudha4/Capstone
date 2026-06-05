import { z } from "zod"

// ─────────────────────────────────────────────
// Request schemas
// ─────────────────────────────────────────────

export const queryRequestSchema = z.object({
    question: z.string().min(1, "Question is required"),
})

// ─────────────────────────────────────────────
// Response schemas
// ─────────────────────────────────────────────

export const queryResponseSchema = z.object({
    answer: z.string(),
})

// Stream returns plain text chunks — no schema needed

// ─────────────────────────────────────────────
// Inferred types
// ─────────────────────────────────────────────

export type QueryRequest  = z.infer<typeof queryRequestSchema>
export type QueryResponse = z.infer<typeof queryResponseSchema>