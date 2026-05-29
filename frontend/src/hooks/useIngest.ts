import { useMutation } from '@tanstack/react-query'
import { ingestService } from '@/services/ingest_service'
import type { IngestEvidenceBody, IngestKmkResponse, IngestEvidenceResponse } from '@/dtos/ingest_dto'

export const useIngestKmk = () => {
    return useMutation<IngestKmkResponse, Error, { file: File }>({
        mutationFn: ({ file }) => ingestService.ingestKmk(file),
    })
}

export const useIngestEvidence = () => {
    return useMutation<IngestEvidenceResponse, Error, { body: IngestEvidenceBody; file: File }>({
        mutationFn: ({ body, file }) => ingestService.ingestEvidence(body, file),
    })
}