import axioHandler from '@/cores/axios'
import {
    getNotificationsResponseSchema,
    notificationMutationResponseSchema,
    sseEventSchema,
} from '@/dtos/notification_dto'
import type {
    GetNotificationsResponse,
    SSEEvent,
} from '@/dtos/notification_dto'
import type { ApiResponse } from '@/dtos/document_dto'

export const notificationService = {

    getAll: async (): Promise<GetNotificationsResponse> => {
        try {
            const { data } = await axioHandler.get('/notifications')

            const parsed = getNotificationsResponseSchema.safeParse(data)
            if (!parsed.success) {
                console.error('[notificationService] getAll validation failed:', parsed.error.format())
                throw new Error('Gagal memvalidasi data notifikasi dari server.')
            }

            return parsed.data
        } catch (error) {
            console.error('[notificationService] getAll error:', error)
            throw error instanceof Error ? error : new Error('Failed to fetch notifications.')
        }
    },

    markRead: async (id: string): Promise<ApiResponse> => {
        try {
            const { data } = await axioHandler.patch(`/notifications/${id}/read`)

            const parsed = notificationMutationResponseSchema.safeParse(data)
            if (!parsed.success) {
                console.error('[notificationService] markRead validation failed:', parsed.error.format())
                throw new Error('Gagal memvalidasi respons server.')
            }

            return parsed.data as ApiResponse
        } catch (error) {
            console.error('[notificationService] markRead error:', error)
            throw error instanceof Error ? error : new Error('Failed to mark notification as read.')
        }
    },

    markAllRead: async (): Promise<ApiResponse> => {
        try {
            const { data } = await axioHandler.patch('/notifications/read-all')

            const parsed = notificationMutationResponseSchema.safeParse(data)
            if (!parsed.success) {
                console.error('[notificationService] markAllRead validation failed:', parsed.error.format())
                throw new Error('Gagal memvalidasi respons server.')
            }

            return parsed.data as ApiResponse
        } catch (error) {
            console.error('[notificationService] markAllRead error:', error)
            throw error instanceof Error ? error : new Error('Failed to mark all notifications as read.')
        }
    },

    // Opens a persistent SSE connection to /notifications/events.
    // Calls onEvent for each received event, onError on connection failure.
    // Returns a cleanup function — call it to close the connection.
    subscribeToEvents: (
        onEvent: (event: SSEEvent) => void,
        onError?: (error: Error) => void,
    ): (() => void) => {
        const abortController = new AbortController()

        const connect = async () => {
            try {
                const baseUrl = axioHandler.defaults.baseURL ?? ''
                const token = localStorage.getItem('accessToken') ?? ''

                const response = await fetch(`${baseUrl}/notifications/events`, {
                    headers: {
                        Authorization: `Bearer ${token}`,
                        Accept: 'text/event-stream',
                    },
                    signal: abortController.signal,
                })

                if (!response.ok) {
                    onError?.(new Error(`SSE failed: ${response.status}`))
                    return
                }

                const reader = response.body!.getReader()
                const decoder = new TextDecoder()
                let buffer = ''

                while (true) {
                    const { done, value } = await reader.read()
                    if (done) break

                    buffer += decoder.decode(value, { stream: true })
                    const parts = buffer.split('\n\n')
                    buffer = parts.pop() ?? ''

                    for (const part of parts) {
                        const dataLine = part.split('\n').find(l => l.startsWith('data:'))
                        if (!dataLine) continue

                        let raw: unknown
                        try {
                            raw = JSON.parse(dataLine.slice(5).trim())
                        } catch {
                            console.error('[SSE] Failed to parse event JSON:', dataLine)
                            continue
                        }

                        const parsed = sseEventSchema.safeParse(raw)
                        if (!parsed.success) {
                            console.error('[SSE] Event failed validation:', parsed.error.format(), raw)
                            continue
                        }

                        onEvent(parsed.data)
                    }
                }
            } catch (err) {
                if (err instanceof Error && err.name === 'AbortError') return
                onError?.(err instanceof Error ? err : new Error(String(err)))
            }
        }

        connect()
        return () => abortController.abort()
    },
}