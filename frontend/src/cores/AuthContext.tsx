import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react'
import { authService } from '../services/auth_service'
import type { AuthContextType, User } from './types'

const AuthContext = createContext<AuthContextType | null>(null)

function clearLegacyTokenStorage() {
    localStorage.removeItem('user')
    localStorage.removeItem('accessToken')
    localStorage.removeItem('refreshToken')
}

export function AuthProvider({ children }: { children: ReactNode }) {
    const [user, setUser] = useState<User | null>(null)
    const [loading, setLoading] = useState<boolean>(true)

    const clearSession = useCallback(() => {
        setUser(null)
        clearLegacyTokenStorage()
    }, [])

    const logout = useCallback(async () => {
        try {
            await authService.logout()
        } catch {
            // Still clear local session if revoke fails (e.g. token already expired).
        }
        clearSession()
    }, [clearSession])

    useEffect(() => {
        // Drop any tokens previously stored in localStorage (XSS surface).
        clearLegacyTokenStorage()

        // Session lives in HttpOnly cookies — bootstrap by calling /auth/me.
        authService.me()
            .then((userData) => {
                setUser(userData)
            })
            .catch(() => {
                setUser(null)
            })
            .finally(() => {
                setLoading(false)
            })
    }, [])

    const login = (userData: User) => {
        setUser(userData)
    }

    return (
        <AuthContext.Provider value={{ user, login, logout, loading }}>
            {children}
        </AuthContext.Provider>
    )
}

export function useAuth(): AuthContextType {
    const context = useContext(AuthContext)
    if (!context) throw new Error('useAuth must be used within AuthProvider')
    return context
}
