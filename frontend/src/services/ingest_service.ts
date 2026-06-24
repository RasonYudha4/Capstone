import axioHandler from '@/cores/axios'
import { apiResponseSchema } from '@/dtos/login_dto'
import { ingestKmkResponseSchema, type IngestKmkResponse } from '@/dtos/ingest_dto'

export const ingestService = {
    ingestKmk: async (): Promise<IngestKmkResponse> => {
        const { data } = await axioHandler.post('/ingest/kmk')
        const parsed = apiResponseSchema(ingestKmkResponseSchema).parse(data)
        if (!parsed.success) throw new Error(parsed.message)
        return ingestKmkResponseSchema.parse(parsed.data)
    },
}