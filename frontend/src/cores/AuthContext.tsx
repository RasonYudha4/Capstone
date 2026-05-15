import { createContext, useContext, useState, useEffect, type ReactNode } from 'react'
import type { AuthContextType, User } from './types'

const AuthContext = createContext<AuthContextType | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
    const [user, setUser] = useState<User | null>(null)
    const [loading, setLoading] = useState<boolean>(true)

    useEffect(() => {
        const stored = localStorage.getItem('user')
        const token = localStorage.getItem('accessToken')
        if (stored && token) {
            setUser(JSON.parse(stored) as User)
        }
        setLoading(false)
    }, [])

    const login = (userData: User, accessToken: string, refreshToken: string) => {
        setUser(userData)
        localStorage.setItem('user', JSON.stringify(userData))
        localStorage.setItem('accessToken', accessToken)
        localStorage.setItem('refreshToken', refreshToken)
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