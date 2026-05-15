import axioHandler from '@/cores/axios'
import type {
    GetNotificationsResponse,
    SSEEvent,
} from '@/schemas/notification.schema'
import type { ApiResponse } from '@/schemas/document.schema'

export const notificationService = {

    getAll: async (): Promise<GetNotificationsResponse> => {
        try {
            const { data } = await axioHandler.get('/notifications')
            return data
        } catch (error) {
            throw new Error('Failed to fetch notifications.')
        }
    },

    markRead: async (id: string): Promise<ApiResponse> => {
        try {
            const { data } = await axioHandler.patch(`/notifications/${id}/read`)
            return data
        } catch (error) {
            throw new Error('Failed to mark notification as read.')
        }
    },

    markAllRead: async (): Promise<ApiResponse> => {
        try {
            const { data } = await axioHandler.patch('/notifications/read-all')
            return data
        } catch (error) {
            throw new Error('Failed to mark all notifications as read.')
        }
    },

    // Opens a persistent SSE connection to /notifications/events.
    // Calls onEvent for each received event, onError on connection failure.
    // Returns a cleanup function — call it to close the connection.
    subscribeToEvents: (
        onEvent: (event: SSEEvent) => void,
        onError?: (error: Event) => void,
    ): (() => void) => {
        const baseUrl = axioHandler.defaults.baseURL ?? ''
        const token = axioHandler.defaults.headers.common['Authorization'] ?? ''

        const url = new URL(`${baseUrl}/notifications/events`)
        if (token) url.searchParams.set('token', String(token).replace('Bearer ', ''))

        const source = new EventSource(url.toString())

        source.onmessage = (e: MessageEvent) => {
            try {
                const parsed: SSEEvent = JSON.parse(e.data)
                onEvent(parsed)
            } catch {
                console.error('Failed to parse SSE event:', e.data)
            }
        }

        source.onerror = (e) => {
            onError?.(e)
        }

        return () => source.close()
    },
}