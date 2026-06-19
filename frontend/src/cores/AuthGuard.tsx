import { Navigate, Outlet, useLocation } from 'react-router'
import { useAuth } from './AuthContext'
import type { AuthGuardProps } from './types'

export default function AuthGuard({ allowedRoles }: AuthGuardProps) {
    const { user, loading } = useAuth()
    const location = useLocation()

    if (loading) return <div>Loading...</div>

    if (!user) {
        return <Navigate to="/" state={{ from: location }} replace />
    }

    if (!allowedRoles.includes(user.role)) {
        return <Navigate to="/unauthorized" replace />
    }

    return <Outlet />
}