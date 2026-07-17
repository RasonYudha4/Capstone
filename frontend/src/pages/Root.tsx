import { Navigate } from 'react-router'
import { lazy } from 'react'
import { useAuth } from '../cores/AuthContext'

const UserLanding = lazy(() => import('./user/Dashboard'))

export default function RootRoute() {
    const { user } = useAuth()

    if (user?.role === 'admin' || user?.role === 'master-admin') {
        return <Navigate to="/dashboard" replace />
    }

    return <UserLanding />
}