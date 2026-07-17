import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react'
import { authService } from '../services/auth_service'
import type { AuthContextType, User } from './types'

const AuthContext = createContext<AuthContextType | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
    const [user, setUser] = useState<User | null>(null)
    const [loading, setLoading] = useState<boolean>(true)

    const clearSession = useCallback(() => {
        setUser(null)
        localStorage.removeItem('user')
        localStorage.removeItem('accessToken')
        localStorage.removeItem('refreshToken')
    }, [])

    const logout = useCallback(async () => {
        const refreshToken = localStorage.getItem('refreshToken')
        if (refreshToken) {
            try {
                await authService.logout(refreshToken)
            } catch {
                // Still clear local session if revoke fails (e.g. token already expired).
            }
        }
        clearSession()
    }, [clearSession])

    useEffect(() => {
        const stored = localStorage.getItem('user')
        const refreshToken = localStorage.getItem('refreshToken')
        if (stored && refreshToken) {
            authService.me().then((userData) => {
                setUser(userData);
                localStorage.setItem('user', JSON.stringify(userData));
            }).catch(() => {
                clearSession();
            }).finally(() => {
                setLoading(false);
            });
        } else {
            setLoading(false);
        }
    }, [clearSession])

    const login = (userData: User, accessToken: string, refreshToken?: string) => {
        setUser(userData)
        localStorage.setItem('user', JSON.stringify(userData))
        localStorage.setItem('accessToken', accessToken)
        if (refreshToken) {
            localStorage.setItem('refreshToken', refreshToken)
        }
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
