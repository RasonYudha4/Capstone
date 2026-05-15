import axioHandler from '@/cores/axios'
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
    },
}