import axioHandler from '@/cores/axios'
import {
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
        return loginResponseSchema.parse(data.data)
    },

    verifyOtp: async (email: string, otp: string): Promise<TokenResponse> => {
        const { data } = await axioHandler.post('/auth/verify-otp', { email, otp })
        return tokenResponseSchema.parse(data.data)
    },

    resendOtp: async (email: string): Promise<void> => {
        await axioHandler.post('/auth/resend-otp', { email })
    },

    me: async (): Promise<UserResponse> => {
        const { data } = await axioHandler.get('/auth/me')
        return userResponseSchema.parse(data.data)
    },

    refresh: async (refreshToken: string): Promise<{ access_token: string }> => {
        const { data } = await axioHandler.post('/auth/refresh', {
            refresh_token: refreshToken,
        })
        return data.data
    },

    logout: async (refreshToken: string): Promise<void> => {
        await axioHandler.post('/auth/logout', { refresh_token: refreshToken })
    },
}