import axioHandler from '@/cores/axios'
import { getAuditResponseSchema } from '@/dtos/audit_dto'
import type { GetAuditResponse } from '@/dtos/audit_dto'

export const auditService = {
    getAll: async (): Promise<GetAuditResponse> => {
        try {
            const { data } = await axioHandler.get('/audit')

            const parsed = getAuditResponseSchema.safeParse(data)
            if (!parsed.success) {
                console.error('[auditService] Zod validation failed:', parsed.error.format())
                throw new Error("Data kosong")
            }

            return parsed.data
        } catch (error) {
            console.error('[auditService] getAll error:', error)
            throw error instanceof Error ? error : new Error('Failed to fetch audit logs.')
        }
    },
}