// hooks/useAgentChat.ts
import { useState, useRef, useCallback } from 'react'
import { useLocation } from 'react-router'
import { queryService } from '@/services/query_service'
import { useFormOptions } from '@/hooks/useFormOption'
import type { Message } from '@/components/molecules/MessageBubble'
import { useFilterStore } from '@/stores/filterStore'
import type { AgentCommand } from '@/cores/types'
import { useAgentCommandExecutor } from './useAgentCommandExecutor'

const INITIAL_MESSAGE: Message = {
    id: 'init',
    role: 'assistant',
    content: 'Halo, ada yang bisa dibantu?',
    timestamp: new Date(),
}

export function useAgentChat() {
    const [messages, setMessages]   = useState<Message[]>([INITIAL_MESSAGE])
    const [isLoading, setIsLoading] = useState(false)
    const [error, setError]         = useState<Error | null>(null)

    const sessionIdRef    = useRef<string | null>(null)
    const streamingIdRef  = useRef<string | null>(null)

    const { pathname }                              = useLocation()
    const { services }                              = useFormOptions()
    const { serviceId, standardId, assessmentId }   = useFilterStore()
    const executeCommand                            = useAgentCommandExecutor()

    // ── Build context snapshot at call time ───────────────────────────────────
    const buildAppContext = useCallback(() => ({
        current_path:       pathname,
        service_id:         serviceId   || null,
        standard_id:        standardId  || null,
        assessment_id:      assessmentId || null,
        available_services: services.map(s => ({
            id:    s.id,
            label: `${s.code} — ${s.description}`,
        })),
    }), [pathname, serviceId, standardId, assessmentId, services])

    // ── Send a message ────────────────────────────────────────────────────────
    const send = useCallback(async (content: string, file?: File) => {
        setError(null)
        setIsLoading(true)

        const assistantId = crypto.randomUUID()
        streamingIdRef.current = assistantId

        setMessages(prev => [
            ...prev,
            {
                id: crypto.randomUUID(),
                role: 'user',
                content,
                timestamp: new Date(),
                attachedFile: file ? { name: file.name, type: file.type } : undefined,
            },
            { id: assistantId,         role: 'assistant', content: '', timestamp: new Date() },
        ])

        try {
            await queryService.queryStream(
                {
                    question:    content,
                    session_id:  sessionIdRef.current,
                    app_context: buildAppContext(),      
                    file,
                },
                // onChunk — append visible text to the streaming bubble
                (chunk) => {
                    setMessages(prev =>
                        prev.map(m =>
                            m.id === streamingIdRef.current
                                ? { ...m, content: m.content + chunk }
                                : m
                        )
                    )
                },
                // onCommand — side-effect: navigate / set filter / open doc
                (cmd: AgentCommand) => {
                    executeCommand(cmd)
                },
                () => setIsLoading(false),
                (sid) => { sessionIdRef.current = sid },
            )
        } catch (err) {
            setError(err instanceof Error ? err : new Error('Stream failed.'))
            setIsLoading(false)
        }
    }, [buildAppContext, executeCommand])

    const reset = useCallback(() => {
        setMessages([INITIAL_MESSAGE])
        setError(null)
        setIsLoading(false)
        sessionIdRef.current   = null
        streamingIdRef.current = null
    }, [])

    return { messages, isLoading, error, send, reset }
}