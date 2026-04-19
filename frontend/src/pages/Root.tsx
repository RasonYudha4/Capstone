import { Navigate } from 'react-router'
import { useAuth } from '../cores/AuthContext'
import type { Role } from '../cores/types'

const redirectMap: Record<Role, string> = {
    "master-admin": '/master-admin',
    "admin": '/admin',
    "staff": '/user',
}

export default function Root() {
    const { user, loading } = useAuth()

    if (loading) return <div>Loading...</div>
    if (!user) return <Navigate to="/login" replace />

    return <Navigate to={redirectMap[user.role]} replace />
}