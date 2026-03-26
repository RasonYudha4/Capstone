import { Navigate } from 'react-router'
import { useAuth } from '../context/AuthContext'
import type { Role } from '../context/types'

const redirectMap: Record<Role, string> = {
    1: '/master-admin',
    2: '/admin',
    3: '/user',
}

export default function Root() {
    const { user, loading } = useAuth()

    if (loading) return <div>Loading...</div>
    if (!user) return <Navigate to="/login" replace />

    return <Navigate to={redirectMap[user.role]} replace />
}