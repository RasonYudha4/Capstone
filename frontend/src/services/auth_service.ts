import axioHandler from '@/cores/axios'
import {
    apiResponseSchema,
    voidApiResponseSchema,
    loginResponseSchema,
    tokenResponseSchema,
    userResponseSchema,
    refreshResponseSchema,
    listUsersResponseSchema,
    type LoginResponse,
    type TokenResponse,
    type UserResponse,
    type RefreshResponse,
    type ListUsersResponse,
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
    verifyOtp: async (email: string, otp: string, preAuthToken: string): Promise<TokenResponse> => {
        const { data } = await axioHandler.post('/auth/verify-otp', { email, otp, pre_auth_token: preAuthToken })
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

    // POST /auth/refresh
    refresh: async (refreshToken: string): Promise<RefreshResponse> => {
        const { data } = await axioHandler.post('/auth/refresh', {
            refresh_token: refreshToken,
        })
        const parsed = apiResponseSchema(refreshResponseSchema).parse(data)
        if (!parsed.success) {
            throw new Error(parsed.message)
        }
        return refreshResponseSchema.parse(parsed.data)
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
    },

    // POST /auth/invite
    inviteUser: async (email: string, role: string): Promise<void> => {
        const { data } = await axioHandler.post('/auth/invite', { email, role })
        const parsed = voidApiResponseSchema.parse(data)
        if (!parsed.success) {
            throw new Error(parsed.message)
        }
    },

    // POST /auth/complete-invitation
    completeInvitation: async (token: string, password: string): Promise<void> => {
        const { data } = await axioHandler.post('/auth/complete-invitation', { token, password })
        const parsed = voidApiResponseSchema.parse(data)
        if (!parsed.success) {
            throw new Error(parsed.message)
        }
    },

    // GET /auth/users — list all users (master-admin only)
    listUsers: async (): Promise<ListUsersResponse> => {
        const { data } = await axioHandler.get('/auth/users')
        const parsed = apiResponseSchema(listUsersResponseSchema).parse(data)
        if (!parsed.success) {
            throw new Error(parsed.message)
        }
        return listUsersResponseSchema.parse(parsed.data)
    },

    // PUT /auth/users/:id/role — update user role (master-admin only)
    updateUserRole: async (userId: string, role: 'admin' | 'staff'): Promise<void> => {
        const { data } = await axioHandler.put(`/auth/users/${userId}/role`, { role })
        const parsed = voidApiResponseSchema.parse(data)
        if (!parsed.success) {
            throw new Error(parsed.message)
        }
    },

    // PUT /auth/users/:id/status — suspend or activate user (master-admin only)
    updateUserStatus: async (userId: string, status: 'active' | 'suspended'): Promise<void> => {
        const { data } = await axioHandler.put(`/auth/users/${userId}/status`, { status })
        const parsed = voidApiResponseSchema.parse(data)
        if (!parsed.success) {
            throw new Error(parsed.message)
        }
    },

    // DELETE /auth/users/:id — delete invited user (master-admin only)
    deleteUser: async (userId: string): Promise<void> => {
        const { data } = await axioHandler.delete(`/auth/users/${userId}`)
        const parsed = voidApiResponseSchema.parse(data)
        if (!parsed.success) {
            throw new Error(parsed.message)
        }
    },
}