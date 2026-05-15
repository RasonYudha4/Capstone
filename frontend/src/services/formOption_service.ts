import axioHandler from '@/cores/axios'
import type { FormOptionsResponse } from '@/dtos/formOption_dto'

export const formOptionsService = {
    getAll: async (): Promise<FormOptionsResponse> => {
        try {
            const { data } = await axioHandler.get('/form-option')
            return data
        } catch (error) {
            throw new Error('Failed to fetch form options.')
        }
    },
}