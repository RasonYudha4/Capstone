import { createContext, useContext, useState, useEffect, type ReactNode } from 'react'
import { authService } from '../services/auth_service'
import type { AuthContextType, User } from './types'

const AuthContext = createContext<AuthContextType | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
    const [user, setUser] = useState<User | null>(null)
    const [loading, setLoading] = useState<boolean>(true)

    useEffect(() => {
        const stored = localStorage.getItem('user')
        const refreshToken = localStorage.getItem('refreshToken')
        if (stored && refreshToken) {
            authService.refresh(refreshToken).then((data) => {
                localStorage.setItem('accessToken', data.access_token);
                // Update refresh token jika backend mengembalikan yang baru
                if (data.refresh_token) {
                    localStorage.setItem('refreshToken', data.refresh_token);
                }
                setUser(JSON.parse(stored) as User);
            }).catch(() => {
                localStorage.removeItem('user');
                localStorage.removeItem('accessToken');
                localStorage.removeItem('refreshToken');
            }).finally(() => {
                setLoading(false);
            });
        } else {
            setLoading(false);
        }
    }, [])

    const login = (userData: User, accessToken: string) => {
        setUser(userData)
        localStorage.setItem('user', JSON.stringify(userData))
        localStorage.setItem('accessToken', accessToken)
    }

    const logout = () => {
        setUser(null)
        localStorage.removeItem('user')
        localStorage.removeItem('accessToken')
        localStorage.removeItem('refreshToken')
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