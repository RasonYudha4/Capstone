import { useMutation } from '@tanstack/react-query'
import { ingestService } from '@/services/ingest_service'
import type { IngestKmkResponse } from '@/dtos/ingest_dto'

export const ingestKeys = {
    all: ['ingest'] as const,
}

export const useIngestKmk = () => {
    return useMutation<IngestKmkResponse, Error, void>({
        mutationFn: () => ingestService.ingestKmk(),
    })
}