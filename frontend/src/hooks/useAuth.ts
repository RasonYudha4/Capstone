import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { authService } from '@/services/auth_service'
import type {
    LoginFormValues,
    OTPFormValues,
    LoginResponse,
    TokenResponse,
    UserResponse,
} from '@/dtos/login_dto'

// ─── Query Keys ───────────────────────────────────────────────────────────────

export const authKeys = {
    all: ['auth'] as const,
    me: () => [...authKeys.all, 'me'] as const,
}

// ─── Queries ──────────────────────────────────────────────────────────────────

export const useMe = (enabled = true) => {
    return useQuery<UserResponse>({
        queryKey: authKeys.me(),
        queryFn: () => authService.me(),
        enabled,
        staleTime: 1000 * 60 * 5,
    })
}

// ─── Mutations ────────────────────────────────────────────────────────────────

export const useLogin = () => {
    return useMutation<LoginResponse, Error, LoginFormValues>({
        mutationFn: ({ email, password }) => authService.login(email, password),
    })
}

export const useVerifyOtp = () => {
    const queryClient = useQueryClient()

    return useMutation<TokenResponse, Error, OTPFormValues & { preAuthToken: string }>({
        mutationFn: ({ email, otp, preAuthToken }) => authService.verifyOtp(email, otp, preAuthToken),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: authKeys.me() })
        },
    })
}

export const useResendOtp = () => {
    return useMutation<void, Error, string>({
        mutationFn: (email) => authService.resendOtp(email),
    })
}

export const useLogout = () => {
    const queryClient = useQueryClient()

    return useMutation<void, Error, string>({
        mutationFn: (refreshToken) => authService.logout(refreshToken),
        onSuccess: () => {
            queryClient.removeQueries({ queryKey: authKeys.all })
        },
    })
}