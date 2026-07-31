import axioHandler from '@/cores/axios'
import { formOptionsResponseSchema } from '@/dtos/formOption_dto'
import type { FormOptionsResponse } from '@/dtos/formOption_dto'

export const formOptionsService = {
    getAll: async (): Promise<FormOptionsResponse> => {
        try {
            const { data } = await axioHandler.get('/form-option')

            const parsed = formOptionsResponseSchema.safeParse(data)
            if (!parsed.success) {
                console.error('[formOptionsService] Zod validation failed:', parsed.error.format())
                throw new Error('Gagal memvalidasi data form options dari server.')
            }

            return parsed.data
        } catch (error) {
            console.error('[formOptionsService] getAll error:', error)
            throw error instanceof Error ? error : new Error('Failed to fetch form options.')
        }
    },
}