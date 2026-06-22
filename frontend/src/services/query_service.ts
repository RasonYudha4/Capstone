import axioHandler from '@/cores/axios'
import type { AgentCommand } from '@/cores/types'
import type {
    QueryRequest,
    QueryResponse,
} from '@/dtos/query_dto'
import { createCommandParser, type CommandHandler } from '@/lib/command-stream-parser'

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
        onCommand?: (cmd: AgentCommand) => void,
        onDone?: () => void,
        onSessionId?: (sessionId: string) => void,
    ): Promise<void> => {
        try {
            const formData = new FormData()
            formData.append('question', body.question)
            if (body.session_id) formData.append('session_id', body.session_id)
            if (body.app_context) formData.append('app_context', JSON.stringify(body.app_context))
            if (body.file) formData.append('file', body.file)

            const response = await fetch(
                `${axioHandler.defaults.baseURL}/query/stream`,
                {
                    method: 'POST',
                    body: formData,   // no Content-Type header — browser sets the multipart boundary itself
                }
            )

            if (!response.ok || !response.body) {
                throw new Error('Stream request failed.')
            }

            const sessionId = response.headers.get('X-Session-Id')
            if (sessionId) onSessionId?.(sessionId)

            const reader = response.body.getReader()
            const decoder = new TextDecoder()
            const commandHandler: CommandHandler = onCommand ? (cmd) => onCommand(cmd) : () => {}
            const parser = createCommandParser(onChunk, commandHandler)

            while (true) {
                const { done, value } = await reader.read()
                if (done) break
                parser(decoder.decode(value, { stream: true }))
            }

            onDone?.()
        } catch (error) {
            throw new Error('Failed to stream answer.')
        }
    },
}