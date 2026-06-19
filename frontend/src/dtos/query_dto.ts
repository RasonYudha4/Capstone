import { z } from "zod"

export const queryRequestSchema = z.object({
    question: z.string().min(1, "Question is required"),
    session_id: z.string().nullable().optional(),
    app_context: z.object({
        current_path: z.string(),
        service_id: z.string().nullable(),
        standard_id: z.string().nullable(),
        assessment_id: z.string().nullable(),
        available_services: z.array(
            z.object({
                id: z.string(),
                label: z.string(),
            })
        ),
    }).nullable().optional(),
})

export const queryResponseSchema = z.object({
    answer: z.string(),
})

export type QueryRequest  = z.infer<typeof queryRequestSchema>
export type QueryResponse = z.infer<typeof queryResponseSchema>