import axioHandler from '@/cores/axios'
<<<<<<< HEAD
import type { User } from '@/cores/types'

export const authService = {
    requestOtp: async (email: string): Promise<void> => {
        try {
            await axioHandler.post('/auth/request-otp', { email })
        } catch (error) {
            throw new Error('Failed to send OTP. Please try again.')
        }
    },

    verifyOtp: async (otp: string): Promise<User> => {
        try {
            const { data } = await axioHandler.post('/auth/verify-otp', { otp })
            return data
        } catch (error) {
            throw new Error('Invalid or expired OTP. Please try again.')
        }
    },

    googleSignIn: async (): Promise<User> => {
        try {
            const { data } = await axioHandler.get('/auth/google')
            return data
        } catch (error) {
            throw new Error('Google sign-in failed. Please try again.')
        }
    },

    logout: async (): Promise<void> => {
        try {
            await axioHandler.post('/auth/logout')
        } catch (error) {
            throw new Error('Logout failed. Please try again.')
        }
    },

    me: async (): Promise<User> => {
        try {
            const { data } = await axioHandler.get('/auth/me')
            return data
        } catch (error) {
            throw new Error('Failed to fetch user session.')
        }
=======
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
    // POST /auth/login
    login: async (email: string, password: string): Promise<LoginResponse> => {
        const { data } = await axioHandler.post('/auth/login', { email, password })
        const parsed = apiResponseSchema(loginResponseSchema).parse(data)
        if (!parsed.success) {
            throw new Error(parsed.message)
        }
        return loginResponseSchema.parse(parsed.data)
    },

    // POST /auth/verify-otp
    verifyOtp: async (email: string, otp: string): Promise<TokenResponse> => {
        const { data } = await axioHandler.post('/auth/verify-otp', { email, otp })
        const parsed = apiResponseSchema(tokenResponseSchema).parse(data)
        if (!parsed.success) {
            throw new Error(parsed.message)
        }
        return tokenResponseSchema.parse(parsed.data)
    },

    // POST /auth/resend-otp
    resendOtp: async (email: string): Promise<void> => {
        const { data } = await axioHandler.post('/auth/resend-otp', { email })
        const parsed = voidApiResponseSchema.parse(data)
        if (!parsed.success) {
            throw new Error(parsed.message)
        }
    },

    // POST /auth/logout
    logout: async (refreshToken: string): Promise<void> => {
        const { data } = await axioHandler.post('/auth/logout', { refresh_token: refreshToken })
        const parsed = voidApiResponseSchema.parse(data)
        if (!parsed.success) {
            throw new Error(parsed.message)
        }
    },

    // GET /auth/me
    me: async (): Promise<UserResponse> => {
        const { data } = await axioHandler.get('/auth/me')
        const parsed = apiResponseSchema(userResponseSchema).parse(data)
        if (!parsed.success) {
            throw new Error(parsed.message)
        }
        return userResponseSchema.parse(parsed.data)
>>>>>>> a3f2b8f3f997b51cf2a98b7b1c959f9917687db3
    },
}