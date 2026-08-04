import axioHandler from '@/cores/axios'
import {
    publicFileUrlResponseSchema,
    type DocumentListResponse,
    type CreateDocumentBody,
    type UpdateDocumentBody,
    type ApprovalRequest,
    type ApiResponse,
    type PaginationQuery,
    type StatsResponse,
} from '@/dtos/document_dto'


// Last-resort fallback for getPublicDocumentById (presigned URL flow).
// getById now fetches the file as a blob and reads Content-Type from
// the response header directly, so this map is not used there.
const EXTENSION_TO_MIME: Record<string, string> = {
    pdf: 'application/pdf',
    jpg: 'image/jpeg',
    jpeg: 'image/jpeg',
    png: 'image/png',
    gif: 'image/gif',
    webp: 'image/webp',
    doc: 'application/msword',
    docx: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
    xls: 'application/vnd.ms-excel',
    xlsx: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
    ppt: 'application/vnd.ms-powerpoint',
    pptx: 'application/vnd.openxmlformats-officedocument.presentationml.presentation',
    txt: 'text/plain',
    csv: 'text/csv',
}

function inferContentTypeFromUrl(url: string): string {
    try {
        const pathname = new URL(url).pathname
        const ext = pathname.split('.').pop()?.toLowerCase() ?? ''
        return EXTENSION_TO_MIME[ext] ?? ''
    } catch {
        return ''
    }
}

/** Unwrap ApiResponse { data: DocumentDataResponse } with a few fallback shapes. */
function unwrapDocumentList(body: unknown, query?: PaginationQuery): DocumentListResponse {
    const root = body as { data?: unknown } | unknown
    const payload =
        root && typeof root === 'object' && root !== null && 'data' in root
            ? (root as { data: unknown }).data
            : root

    if (Array.isArray(payload)) {
        return { data: payload as DocumentListResponse['data'], page: query?.page ?? 1, limit: query?.limit ?? 10 }
    }

    if (payload && typeof payload === 'object') {
        const p = payload as { data?: unknown; page?: number; limit?: number }
        const list = Array.isArray(p.data) ? p.data : []
        return {
            data: list as DocumentListResponse['data'],
            page: p.page ?? query?.page ?? 1,
            limit: p.limit ?? query?.limit ?? 10,
        }
    }

    return { data: [], page: query?.page ?? 1, limit: query?.limit ?? 10 }
}

export const documentService = {

    getAll: async (query?: PaginationQuery): Promise<DocumentListResponse> => {
        try {
            const { data } = await axioHandler.get('/documents', { params: query })
            return data.data
        } catch (error) {
            throw new Error('Failed to fetch documents.')
        }
    },

    getById: async (id: string): Promise<{ url: string; contentType: string }> => {
        try {
            const response = await axioHandler.get(`/documents/${id}`, {
                responseType: 'blob',
            })
            const contentType = response.headers['content-type'] || 'application/octet-stream'
            const blob = new Blob([response.data], { type: contentType })
            const url = URL.createObjectURL(blob)
            return { url, contentType }
        } catch (error) {
            throw new Error('Failed to fetch document.')
        }
    },


    // ── Public endpoints (no auth) ──────────────────────────────────────────

    getPublicDocuments: async (query?: PaginationQuery): Promise<DocumentListResponse> => {
        try {
            const { data } = await axioHandler.get('/documents/public', { params: query })
            return data.data
        } catch (error) {
            throw new Error('Failed to fetch public documents.')
        }
    },

    getPublicDocumentById: async (id: string): Promise<{ url: string; contentType: string }> => {
        try {
            const { data } = await axioHandler.get(`/documents/public/${id}`)
            const parsed = publicFileUrlResponseSchema.parse(data.data)
            const ct = parsed['content-type']
            return {
                url: parsed.presigned_url,
                contentType: ct || inferContentTypeFromUrl(parsed.presigned_url),
            }
        } catch (error) {
            throw new Error('Failed to fetch public document URL.')
        }
    },

    // ── Existing endpoints ──────────────────────────────────────────────────

    getByType: async (type: string, query?: PaginationQuery): Promise<DocumentListResponse> => {
        try {
            const { data } = await axioHandler.get(`/documents/type/${type}`, { params: query })
            return data.data
        } catch (error) {
            throw new Error('Failed to fetch documents by type.')
        }
    },

    getByGroup: async (group: string, query?: PaginationQuery): Promise<DocumentListResponse> => {
        try {
            const { data } = await axioHandler.get(`/documents/groups/${group}`, { params: query })
            return data.data
        } catch (error) {
            throw new Error('Failed to fetch documents by group.')
        }
    },

    getByService: async (service: string, query?: PaginationQuery): Promise<DocumentListResponse> => {
        try {
            const { data } = await axioHandler.get(`/documents/services/${service}`, { params: query })
            return unwrapDocumentList(data, query)
        } catch (error) {
            throw new Error('Failed to fetch documents by service.')
        }
    },

    getByStandard: async (standard: string, query?: PaginationQuery): Promise<DocumentListResponse> => {
        try {
            const { data } = await axioHandler.get(`/documents/standards/${standard}`, { params: query })
            return unwrapDocumentList(data, query)
        } catch (error) {
            throw new Error('Failed to fetch documents by standard.')
        }
    },

    getByAssessment: async (assessment: string, query?: PaginationQuery): Promise<DocumentListResponse> => {
        try {
            const { data } = await axioHandler.get(`/documents/assessments/${assessment}`, { params: query })
            return unwrapDocumentList(data, query)
        } catch (error) {
            throw new Error('Failed to fetch documents by assessment.')
        }
    },

    getMyDocuments: async (query?: PaginationQuery): Promise<DocumentListResponse> => {
        try {
            const { data } = await axioHandler.get('/documents/createdBy/my-document', { params: query })
            return unwrapDocumentList(data, query)
        } catch (error) {
            throw new Error('Failed to fetch your documents.')
        }
    },

    getByStatus: async (status: string, query?: PaginationQuery): Promise<DocumentListResponse> => {
        try {
            const { data } = await axioHandler.get(`/documents/masterAdmin/status/${status}`, { params: query })
            return unwrapDocumentList(data, query)
        } catch (error) {
            throw new Error('Failed to fetch documents by status.')
        }
    },

    upload: async (body: CreateDocumentBody, file: File): Promise<ApiResponse> => {
        const form = new FormData()
        form.append('uploadedFile', file)
        Object.entries(body).forEach(([key, value]) => {
            if (value !== undefined) form.append(key, value)
        })
        const { data } = await axioHandler.post<ApiResponse>('/documents/upload', form, {
            headers: { 'Content-Type': 'multipart/form-data' },
        })
        return data
    },

    update: async (body: UpdateDocumentBody, file?: File): Promise<ApiResponse> => {
        const form = new FormData()

        if (file && file.size > 0) {
            form.append('uploadedFile', file)
        }

        Object.entries(body).forEach(([key, value]) => {
            if (value !== undefined && value !== '') {
                form.append(key, String(value))
            }
        })

        try {
            const { data } = await axioHandler.patch('/documents/edit', form, {
                headers: { 'Content-Type': 'multipart/form-data' },
            })
            return data
        } catch (error) {
            throw new Error('Failed to update document.')
        }
    },
    approve: async (body: ApprovalRequest, signedFile?: File): Promise<ApiResponse> => {
        try {
            if (signedFile && signedFile.size > 0) {
                const form = new FormData()
                form.append('document_id', body.document_id)
                form.append('status', body.status)
                form.append('uploadedFile', signedFile)
                const { data } = await axioHandler.post('/documents/status/update', form, {
                    headers: { 'Content-Type': 'multipart/form-data' },
                })
                return data
            }
            const { data } = await axioHandler.post('/documents/status/update', body, {
                headers: { 'Content-Type': 'multipart/form-data' },
            })
            return data
        } catch (error) {
            throw new Error('Failed to update document status.')
        }
    },

    delete: async (documentId: string): Promise<ApiResponse> => {
        try {
            const { data } = await axioHandler.delete(`/documents/delete/${documentId}`)
            return data
        } catch (error) {
            throw new Error('Failed to delete document.')
        }
    },
}

export const statsService = {
    getStats: async (): Promise<StatsResponse> => {
        try {
            const { data } = await axioHandler.get('/documents/stats')
            const payload = data?.data ?? data
            if (!payload || typeof payload !== 'object') {
                throw new Error('Invalid statistics payload.')
            }
            const stats = payload.stats ?? { approved: 0, pending: 0, rejected: 0 }
            const total =
                typeof payload.total === 'number'
                    ? payload.total
                    : (stats.approved ?? 0) + (stats.pending ?? 0) + (stats.rejected ?? 0)
            return {
                total,
                groups: payload.groups ?? [],
                stats: {
                    approved: stats.approved ?? 0,
                    pending: stats.pending ?? 0,
                    rejected: stats.rejected ?? 0,
                },
            }
        } catch (error) {
            throw new Error('Failed to fetch statistics.')
        }
    },
}