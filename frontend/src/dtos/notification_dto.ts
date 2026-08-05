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
    // Present when the notification is tied to a document (review target).
    // Accept PascalCase (Go current) and snake_case for compatibility.
    DocumentID:         z.string().uuid().nullish(),
    document_id:        z.string().uuid().nullish(),
    Filename:           z.string().optional().default(''),
    DocumentType:       z.string().optional().default(''),
    CreatedBy:          z.string().optional().default(''),
    DocumentStatus:     z.string().optional().default(''),
    DocumentUpdatedAt:  z.string().nullish(),
    ServiceCode:        z.string().optional().default(''),
    StandardCode:       z.string().optional().default(''),
    AssessmentCode:     z.string().optional().default(''),
}).transform((n) => ({
    ...n,
    DocumentID: n.DocumentID ?? n.document_id ?? null,
}))

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
    data: z.array(notificationItemSchema),
})

// PATCH /notifications/:id/read  → Response
// PATCH /notifications/read-all  → Response
export { responseSchema as notificationMutationResponseSchema };

// ─────────────────────────────────────────────
// SSE events — two distinct shapes over the same stream
// ─────────────────────────────────────────────

// Sent once on connect if there are unread notifications missed while offline
export const sseSyncEventSchema = z.object({
    type:  z.literal("sync"),
    count: z.number().int().nonnegative(),
});

// Sent per document lifecycle event; notification_id set only after DB insert succeeds
export const sseNotificationEventSchema = z.object({
    type: z.enum(["new_document", "document_status_update"]),
    notification_id: uuid.optional(),
    document_id: uuid,
    status: documentStatus,
    message: z.string(),
});

// GET /notifications/events — SSE stream (union of both message shapes)
export const sseEventSchema = z.discriminatedUnion("type", [
    sseSyncEventSchema,
    sseNotificationEventSchema.extend({
        // discriminatedUnion needs a single literal per branch on "type";
        // since this branch covers two literal values, split it further:
    }),
]);

// discriminatedUnion requires exactly one literal value per branch,
// so "new_document" / "document_status_update" each need their own schema:
export const sseNewDocumentEventSchema = sseNotificationEventSchema.extend({
    type: z.literal("new_document"),
});
export const sseDocumentStatusUpdateEventSchema = sseNotificationEventSchema.extend({
    type: z.literal("document_status_update"),
});

export const sseStreamEventSchema = z.discriminatedUnion("type", [
    sseSyncEventSchema,
    sseNewDocumentEventSchema,
    sseDocumentStatusUpdateEventSchema,
]);

// ─────────────────────────────────────────────
// Inferred types
// ─────────────────────────────────────────────

export type NotificationItem         = z.infer<typeof notificationItemSchema>;
export type GetNotificationsResponse = z.infer<typeof getNotificationsResponseSchema>;
export type SSESyncEvent             = z.infer<typeof sseSyncEventSchema>;
export type SSENotificationEvent     = z.infer<typeof sseNewDocumentEventSchema | typeof sseDocumentStatusUpdateEventSchema>;
export type SSEEvent                 = z.infer<typeof sseStreamEventSchema>;