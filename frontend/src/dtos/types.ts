import type { UserResponse } from '@/dtos/login_dto'

// Re-export UserResponse as User — the Zod schema is the source of truth.
export type User = UserResponse

export type Role = "master-admin" | "admin" | "staff"

export interface AuthContextType {
    user: User | null
    loading: boolean
    login: (user: User, accessToken: string, refreshToken: string) => void
    logout: () => void
}







export interface AuthGuardProps {
    allowedRoles: Role[]
}