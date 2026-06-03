import axioHandler from '@/cores/axios'
import type {
    IngestEvidenceBody,
    IngestKmkResponse,
    IngestEvidenceResponse,
} from '@/dtos/ingest_dto'

export const ingestService = {

    ingestKmk: async (file: File): Promise<IngestKmkResponse> => {
        try {
            const form = new FormData()
            form.append('file', file)
            const { data } = await axioHandler.post('/ingest/kmk', form, {
                headers: { 'Content-Type': 'multipart/form-data' },
            })
            return data
        } catch (error) {
            throw new Error('Failed to ingest KMK document.')
        }
    },

    ingestEvidence: async (body: IngestEvidenceBody, file: File): Promise<IngestEvidenceResponse> => {
        try {
            const form = new FormData()
            form.append('file', file)
            Object.entries(body).forEach(([key, value]) => {
                if (value !== undefined) form.append(key, value)
            })
            const { data } = await axioHandler.post('/ingest/evidence', form, {
                headers: { 'Content-Type': 'multipart/form-data' },
            })
            return data
        } catch (error) {
            throw new Error('Failed to ingest evidence document.')
        }
    },

}