import { Navigate } from 'react-router'
import { useAuth } from '../context/AuthContext'
import type { Role } from '../context/types'

const redirectMap: Record<Role, string> = {
    'admin': '/admin',
    'master-admin': '/master-admin',
    'user': '/user',
}

export default function Root() {
    const { user, loading } = useAuth()

    if (loading) return <div>Loading...</div>
    if (!user) return <Navigate to="/login" replace />

    return <Navigate to={redirectMap[user.role]} replace />
}