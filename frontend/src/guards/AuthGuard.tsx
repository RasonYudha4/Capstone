import { Navigate, Outlet, useLocation } from 'react-router'
import { useAuth } from '../context/AuthContext'
import type { Role } from '../context/types'

interface AuthGuardProps {
    allowedRoles: Role[]
}

export default function AuthGuard({ allowedRoles }: AuthGuardProps) {
    const { user, loading } = useAuth()
    const location = useLocation()

    if (loading) return <div>Loading...</div>

    if (!user) {
        return <Navigate to="/login" state={{ from: location }} replace />
    }

    // Root redirect based on role
    if (!allowedRoles) {
        const redirectMap: Record<Role, string> = {
            'admin': '/admin',
            'master-admin': '/master-admin',
            'user': '/user',
        }
        return <Navigate to={redirectMap[user.role]} replace />
    }

    if (!allowedRoles.includes(user.role)) {
        return <Navigate to="/unauthorized" replace />
    }

    return <Outlet />
}