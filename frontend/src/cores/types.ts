export type Role = "master-admin" | "admin" | "staff" 

export interface AuthContextType {
    user: User | null
    loading: boolean
    login: (userData: User) => void
    logout: () => void
}

export interface User {
  id: string
  name: string
  role: Role
}

export interface AuthGuardProps {
    allowedRoles: Role[]
}