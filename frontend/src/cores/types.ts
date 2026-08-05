import type { UserResponse } from '@/dtos/login_dto'

export type AgentCommand =
  | { type: 'NAVIGATE'; path: string }
  | { type: 'SET_SERVICE_FILTER'; serviceId: string }
  | { type: 'SET_STANDARD_FILTER'; standardId: string }
  | { type: 'SET_ASSESSMENT_FILTER'; assessmentId: string }
  | { type: 'OPEN_DOCUMENT'; documentId: string }

// Re-export UserResponse as User — the Zod schema is the source of truth.
export type User = UserResponse

export type Role = "master-admin" | "admin" | "staff"

export interface AuthContextType {
    user: User | null
    loading: boolean
    login: (user: User) => void
    logout: () => Promise<void>
}

export interface AuthGuardProps {
    allowedRoles: Role[]
}