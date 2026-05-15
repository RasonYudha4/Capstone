import { useEffect, useRef } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { notificationService } from '@/services/notif_services'
import type { SSEEvent } from '@/dtos/notification_dto'

const NOTIF_KEY = ['notifications'] as const

// ─────────────────────────────────────────────
// GET /notifications
// ─────────────────────────────────────────────
export function useNotifications() {
    return useQuery({
        queryKey: NOTIF_KEY,
        queryFn:  notificationService.getAll,
    })
}

// ─────────────────────────────────────────────
// PATCH /notifications/:id/read
// ─────────────────────────────────────────────
export function useMarkNotificationRead() {
    const qc = useQueryClient()

    return useMutation({
        mutationFn: (id: string) => notificationService.markRead(id),
        // Optimistic update: flip the single item to Read=true immediately
        onMutate: async (id: string) => {
            await qc.cancelQueries({ queryKey: NOTIF_KEY })
            const prev = qc.getQueryData(NOTIF_KEY)

            qc.setQueryData(NOTIF_KEY, (old: any) => {
                if (!old?.data) return old
                return {
                    ...old,
                    data: old.data.map((n: any) =>
                        n.NotificationID === id ? { ...n, Read: true } : n
                    ),
                }
            })

            return { prev }
        },
        onError: (_err, _id, ctx) => {
            // Roll back on error
            if (ctx?.prev) qc.setQueryData(NOTIF_KEY, ctx.prev)
        },
        onSettled: () => {
            qc.invalidateQueries({ queryKey: NOTIF_KEY })
        },
    })
}

// ─────────────────────────────────────────────
// PATCH /notifications/read-all
// ─────────────────────────────────────────────
export function useMarkAllNotificationsRead() {
    const qc = useQueryClient()

    return useMutation({
        mutationFn: notificationService.markAllRead,
        onMutate: async () => {
            await qc.cancelQueries({ queryKey: NOTIF_KEY })
            const prev = qc.getQueryData(NOTIF_KEY)

            qc.setQueryData(NOTIF_KEY, (old: any) => {
                if (!old?.data) return old
                return {
                    ...old,
                    data: old.data.map((n: any) => ({ ...n, Read: true })),
                }
            })

            return { prev }
        },
        onError: (_err, _vars, ctx) => {
            if (ctx?.prev) qc.setQueryData(NOTIF_KEY, ctx.prev)
        },
        onSettled: () => {
            qc.invalidateQueries({ queryKey: NOTIF_KEY })
        },
    })
}

// ─────────────────────────────────────────────
// GET /notifications/events  (SSE)
// ─────────────────────────────────────────────
// Opens a persistent SSE connection and calls onEvent for each message.
// Also invalidates the notifications query so the list auto-refreshes
// whenever a new notification arrives.
export function useNotificationEvents(
    onEvent: (event: SSEEvent) => void,
) {
    const qc       = useQueryClient()
    // Keep a stable ref so the effect doesn't re-run when onEvent changes
    const onEventRef = useRef(onEvent)
    useEffect(() => { onEventRef.current = onEvent }, [onEvent])

    useEffect(() => {
        const cleanup = notificationService.subscribeToEvents(
            (event) => {
                onEventRef.current(event)
                // Refetch the notification list to include the new item
                qc.invalidateQueries({ queryKey: NOTIF_KEY })
            },
            (err) => {
                console.error('[SSE] connection error', err)
            },
        )

        return cleanup
    }, [qc]) // qc is stable; SSE reconnects only on mount/unmount
}