import { useState, useCallback } from 'react'
import { useMutation } from '@tanstack/react-query'
import { queryService } from '@/services/query_service'
import type { QueryRequest, QueryResponse } from '@/dtos/query_dto'

export const useQueryChat = () => {
    return useMutation<QueryResponse, Error, QueryRequest>({
        mutationFn: (body) => queryService.query(body),
    })
}

export const useQueryStream = () => {
    const [answer, setAnswer]     = useState('')
    const [isLoading, setLoading] = useState(false)
    const [error, setError]       = useState<Error | null>(null)

    const submit = useCallback(async (question: string) => {
        setAnswer('')
        setError(null)
        setLoading(true)

        try {
            await queryService.queryStream(
                { question },
                (chunk) => setAnswer((prev) => prev + chunk),
                ()      => setLoading(false),
            )
        } catch (err) {
            setError(err instanceof Error ? err : new Error('Stream failed.'))
            setLoading(false)
        }
    }, [])

    const reset = useCallback(() => {
        setAnswer('')
        setError(null)
        setLoading(false)
    }, [])

    return { answer, isLoading, error, submit, reset }
}