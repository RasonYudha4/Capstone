import { z } from "zod";
import { documentStatus, responseSchema } from "./document_dto";

// ─────────────────────────────────────────────
// Shared primitives
// ─────────────────────────────────────────────

const uuid = z.string().uuid("Invalid UUID");

// ─────────────────────────────────────────────
// Notification item
// ─────────────────────────────────────────────

export const notificationItemSchema = z.object({
    NotificationID: z.string().uuid(),
    Message:        z.string(),
    Read:           z.boolean(),
    UserID:         z.string().uuid(),
    CreatedAt:      z.string().datetime({ offset: true }),
    UpdatedAt:      z.string().datetime({ offset: true }),
})
// ─────────────────────────────────────────────
// Path param schemas
// ─────────────────────────────────────────────

// PATCH /notifications/:id/read
export const notificationIdParamSchema = z.object({ id: uuid });

// ─────────────────────────────────────────────
// Response schemas
// ─────────────────────────────────────────────

// GET /notifications
export const getNotificationsResponseSchema = z.object({
    data: z.array(notificationItemSchema),  // remove status — backend doesn't send it
})

// PATCH /notifications/:id/read  → Response
// PATCH /notifications/read-all  → Response
export { responseSchema as notificationMutationResponseSchema };

// GET /notifications/events — SSE stream
// Maps to SSEEvent struct; notification_id is set only after DB insert succeeds
export const sseEventSchema = z.object({
    type: z.enum(["new_document", "document_status_update"]),
    notification_id: uuid.optional(),
    document_id: uuid,
    status: documentStatus,
    message: z.string(),
});

// ─────────────────────────────────────────────
// Inferred types
// ─────────────────────────────────────────────

export type NotificationItem         = z.infer<typeof notificationItemSchema>;
export type GetNotificationsResponse = z.infer<typeof getNotificationsResponseSchema>;
export type SSEEvent                 = z.infer<typeof sseEventSchema>;