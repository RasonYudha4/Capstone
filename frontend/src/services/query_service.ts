import axioHandler from '@/cores/axios'
import type {
    QueryRequest,
    QueryResponse,
} from '@/dtos/query_dto'

export const queryService = {

    query: async (body: QueryRequest): Promise<QueryResponse> => {
        try {
            const { data } = await axioHandler.post('/query/', body)
            return data
        } catch (error) {
            throw new Error('Failed to fetch answer.')
        }
    },

    queryStream: async (
        body: QueryRequest,
        onChunk: (chunk: string) => void,
        onDone?: () => void,
        onSessionId?: (sessionId: string) => void,
    ): Promise<void> => {
        try {
            const response = await fetch(
                `${axioHandler.defaults.baseURL}/query/stream`,
                {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(body),
                }
            )

            if (!response.ok || !response.body) {
                throw new Error('Stream request failed.')
            }

            const sessionId = response.headers.get('X-Session-Id')
            if (sessionId) {
                onSessionId?.(sessionId)
            }

            const reader = response.body.getReader()
            const decoder = new TextDecoder()

            while (true) {
                const { done, value } = await reader.read()
                if (done) break
                onChunk(decoder.decode(value, { stream: true }))
            }

            onDone?.()
        } catch (error) {
            throw new Error('Failed to stream answer.')
        }
    },
}