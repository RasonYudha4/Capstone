import axioHandler from '@/cores/axios'
import {
    apiResponseSchema,
    voidApiResponseSchema,
    loginResponseSchema,
    tokenResponseSchema,
    userResponseSchema,
    type LoginResponse,
    type TokenResponse,
    type UserResponse,
} from '@/dtos/login_dto'

export const authService = {
    login: async (email: string, password: string): Promise<LoginResponse> => {
        const { data } = await axioHandler.post('/auth/login', { email, password })
        const parsed = apiResponseSchema(loginResponseSchema).parse(data)
        if (!parsed.success) throw new Error(parsed.message)
        return loginResponseSchema.parse(parsed.data)
    },

    verifyOtp: async (email: string, otp: string): Promise<TokenResponse> => {
        const { data } = await axioHandler.post('/auth/verify-otp', { email, otp })
        const parsed = apiResponseSchema(tokenResponseSchema).parse(data)
        if (!parsed.success) throw new Error(parsed.message)
        return tokenResponseSchema.parse(parsed.data)
    },

    resendOtp: async (email: string): Promise<void> => {
        const { data } = await axioHandler.post('/auth/resend-otp', { email })
        const parsed = voidApiResponseSchema.parse(data)
        if (!parsed.success) throw new Error(parsed.message)
    },

    logout: async (refreshToken: string): Promise<void> => {
        const { data } = await axioHandler.post('/auth/logout', { refresh_token: refreshToken })
        const parsed = voidApiResponseSchema.parse(data)
        if (!parsed.success) throw new Error(parsed.message)
    },

    me: async (): Promise<UserResponse> => {
        const { data } = await axioHandler.get('/auth/me')
        const parsed = apiResponseSchema(userResponseSchema).parse(data)
        if (!parsed.success) throw new Error(parsed.message)
        return userResponseSchema.parse(parsed.data)
    },
}