import { z } from "zod";

// ─────────────────────────────────────────────
// Shared primitives
// ─────────────────────────────────────────────

const uuid = z.string().uuid("Invalid UUID");

export const documentStatus = z.enum(["pending", "approved", "rejected"]);

// ─────────────────────────────────────────────
// Shared responses
// ─────────────────────────────────────────────

// type Response struct
export const responseSchema = z.object({
    status: z.boolean(),
    message: z.string(),
});

// type UploadResponse struct
export const uploadResponseSchema = z.object({
    status: z.boolean(),
    message: z.string(),
    filename: z.string(),
    filesize: z.number().int().nonnegative(),
});

// type DocumentResponse struct
export const documentResponseSchema = z.object({
    document_id:     uuid,
    filename:        z.string(),
    filepath:        z.string(),
    document_type:   z.string(),
    created_by:      z.string(),
    updated_at:      z.string().datetime(),
    assessment:      z.string(),
    status:          documentStatus,
    service_code:    z.string(),
    standard_code:   z.string(),
    assessment_code: z.string(),
});

// ─────────────────────────────────────────────
// Pagination
// pageLimit() clamps: page >= 1, limit 1–100 (default 20)
// ─────────────────────────────────────────────

export const paginationQuerySchema = z.object({
    page: z.coerce.number().int().positive().default(1),
    limit: z.coerce.number().int().positive().max(100).default(10),
});

// ─────────────────────────────────────────────
// Request schemas
// ─────────────────────────────────────────────

// POST /documents/upload — type DocumentRequest struct (multipart/form-data)
export const createDocumentBodySchema = z.object({
    service_id: uuid,
    standard_id: uuid,
    assessment_id: uuid,
    filename: z.string().min(1, "Filename is required"),
    document_type_id: uuid,
    description: z.string().optional(),
});

// PATCH /documents/edit — type UpdateRequest struct (multipart/form-data)
export const updateDocumentBodySchema = z.object({
  document_id: uuid,
  filename: z.string().optional(),     // optional rename
  description: z.string().optional(),  // from catatan
})

// POST /documents/status/update — type ApprovalRequest struct (JSON)
export const approvalRequestSchema = z.object({
    document_id: uuid,
    status: documentStatus,
});

export const groupStatSchema = z.object({
    group_id:       z.string().uuid(),
    group_name:     z.string(),
    total_files:    z.number().int(),
    empty_sections: z.number().int(),
})
 
export const statusStatSchema = z.object({
    approved: z.number().int(),
    pending:  z.number().int(),
    rejected: z.number().int(),
})
 
export const statsResponseSchema = z.object({
    status: z.boolean(),
    total:  z.number().int(),
    groups: z.array(groupStatSchema),
    stats:  statusStatSchema,
})

// ─────────────────────────────────────────────
// Path param schemas
// All filter params are UUIDs — service layer parses with uuid.UUID
// ─────────────────────────────────────────────

export const documentIdParamSchema = z.object({ id: uuid });
export const documentTypeParamSchema = z.object({ type: uuid });
export const documentGroupParamSchema = z.object({ group: uuid });
export const documentServiceParamSchema = z.object({ service: uuid });
export const documentStandardParamSchema = z.object({ standard: uuid });
export const documentAssessmentParamSchema = z.object({ assessment: uuid });
export const documentStatusParamSchema = z.object({ status: documentStatus });
export const deleteDocumentParamSchema = z.object({ documentId: uuid });


// ─────────────────────────────────────────────
// Response schemas
// ─────────────────────────────────────────────

// GET /documents, /type/:type, /groups/:group, /services/:service,
//     /standards/:standard, /assessments/:assessment,
//     /createdBy/my-document, /masterAdmin/status/:status
export const documentListResponseSchema = z.object({
    status: z.boolean(),
    page: z.number().int(),
    limit: z.number().int(),
    data: z.array(documentResponseSchema),
});

// GET /documents/:id
export const fileUrlResponseSchema = z.object({
    Data: z.string().url(),
});

// POST /documents/upload
export { uploadResponseSchema as createDocumentResponseSchema };

// PATCH /documents/edit
// POST  /documents/status/update
// DELETE /documents/delete/:documentId
export { responseSchema as mutationResponseSchema };

// ─────────────────────────────────────────────
// Inferred types
// ─────────────────────────────────────────────

export type DocumentStatus         = z.infer<typeof documentStatus>;
export type DocumentResponse       = z.infer<typeof documentResponseSchema>;
export type DocumentListResponse   = z.infer<typeof documentListResponseSchema>;
export type FileUrlResponse = z.infer<typeof fileUrlResponseSchema>;
export type CreateDocumentBody     = z.infer<typeof createDocumentBodySchema>;
export type UpdateDocumentBody     = z.infer<typeof updateDocumentBodySchema>;
export type ApprovalRequest        = z.infer<typeof approvalRequestSchema>;
export type PaginationQuery        = z.infer<typeof paginationQuerySchema>;
export type ApiResponse            = z.infer<typeof responseSchema>;
export type GroupStat      = z.infer<typeof groupStatSchema>
export type StatusStat     = z.infer<typeof statusStatSchema>
export type StatsResponse  = z.infer<typeof statsResponseSchema>