import { createContext, useContext, useState, useEffect, type ReactNode } from 'react'
import { setAccessToken, getAccessToken } from './tokenStore'
import { authService } from '../services/auth_service'
import type { AuthContextType, User } from './types'

const AuthContext = createContext<AuthContextType | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
    const [user, setUser] = useState<User | null>(null)
    const [loading, setLoading] = useState<boolean>(true)

    useEffect(() => {
        const stored = localStorage.getItem('user')
        if (stored) {
            authService.refresh().then((data) => {
                setAccessToken(data.access_token);
                setUser(JSON.parse(stored) as User);
            }).catch(() => {
                localStorage.removeItem('user');
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
        setAccessToken(accessToken)
    }

    const logout = () => {
        setUser(null)
        localStorage.removeItem('user')
        setAccessToken(null)
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