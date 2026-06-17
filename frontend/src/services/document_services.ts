import axioHandler from '@/cores/axios'
import type {
    DocumentListResponse,
    FileUrlResponse,
    CreateDocumentBody,
    UpdateDocumentBody,
    ApprovalRequest,
    ApiResponse,
    PaginationQuery,
    StatsResponse,
} from '@/dtos/document_dto'

export const documentService = {

    getAll: async (query?: PaginationQuery): Promise<DocumentListResponse> => {
        try {
            const { data } = await axioHandler.get('/documents', { params: query })
            return data
        } catch (error) {
            throw new Error('Failed to fetch documents.')
        }
    },

    getById: async (id: string): Promise<{ url: string; contentType: string }> => {
        try {
            const { data, headers } = await axioHandler.get(`/documents/${id}`, {
                responseType: 'blob'
            })
            return {
                url: URL.createObjectURL(data),
                contentType: headers['content-type'] ?? ''
            }
        } catch (error) {
            throw new Error('Failed to fetch document.')
        }
    },

    // ── Public endpoints (no auth) ──────────────────────────────────────────

    getPublicDocuments: async (query?: PaginationQuery): Promise<DocumentListResponse> => {
        try {
            const { data } = await axioHandler.get('/documents/public', { params: query })
            return data
        } catch (error) {
            throw new Error('Failed to fetch public documents.')
        }
    },

    getPublicDocumentById: async (id: string): Promise<{ url: string }> => {
        try {
            const { data } = await axioHandler.get(`/documents/public/${id}`)
            return data
        } catch (error) {
            throw new Error('Failed to fetch public document URL.')
        }
    },

    // ── Existing endpoints ──────────────────────────────────────────────────

    getByType: async (type: string, query?: PaginationQuery): Promise<DocumentListResponse> => {
        try {
            const { data } = await axioHandler.get(`/documents/type/${type}`, { params: query })
            return data
        } catch (error) {
            throw new Error('Failed to fetch documents by type.')
        }
    },

    getByGroup: async (group: string, query?: PaginationQuery): Promise<DocumentListResponse> => {
        try {
            const { data } = await axioHandler.get(`/documents/groups/${group}`, { params: query })
            return data
        } catch (error) {
            throw new Error('Failed to fetch documents by group.')
        }
    },

    getByService: async (service: string, query?: PaginationQuery): Promise<DocumentListResponse> => {
        try {
            const { data } = await axioHandler.get(`/documents/services/${service}`, { params: query })
            return data
        } catch (error) {
            throw new Error('Failed to fetch documents by service.')
        }
    },

    getByStandard: async (standard: string, query?: PaginationQuery): Promise<DocumentListResponse> => {
        try {
            const { data } = await axioHandler.get(`/documents/standards/${standard}`, { params: query })
            return data
        } catch (error) {
            throw new Error('Failed to fetch documents by standard.')
        }
    },

    getByAssessment: async (assessment: string, query?: PaginationQuery): Promise<DocumentListResponse> => {
        try {
            const { data } = await axioHandler.get(`/documents/assessments/${assessment}`, { params: query })
            return data
        } catch (error) {
            throw new Error('Failed to fetch documents by assessment.')
        }
    },

    getMyDocuments: async (query?: PaginationQuery): Promise<DocumentListResponse> => {
        try {
            const { data } = await axioHandler.get('/documents/createdBy/my-document', { params: query })
            return data
        } catch (error) {
            throw new Error('Failed to fetch your documents.')
        }
    },

    getByStatus: async (status: string, query?: PaginationQuery): Promise<DocumentListResponse> => {
        try {
            const { data } = await axioHandler.get(`/documents/masterAdmin/status/${status}`, { params: query })
            return data
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
            console.log(`FormData: key=${key} value=${String(value)} included=${value !== undefined && value !== ''}`)
            if (value !== undefined && value !== '') {
                form.append(key, String(value))
            }
        })

        for (const [key, value] of form.entries()) {
            console.log(`Final FormData: ${key} =`, value)
        }

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
                form.append('file', signedFile)
                const { data } = await axioHandler.post('/documents/status/update', form, {
                    headers: { 'Content-Type': 'multipart/form-data' },
                })
                return data
            }
            const { data } = await axioHandler.post('/documents/status/update', body)
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
            return data
        } catch (error) {
            throw new Error('Failed to fetch statistics.')
        }
    },
}