import { z } from 'zod'

export const ingestKmkResponseSchema = z.object({
    filename: z.string(),
    chunks_upserted: z.number(),
})

export type IngestKmkResponse = z.infer<typeof ingestKmkResponseSchema>