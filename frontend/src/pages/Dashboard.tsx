import { lazy } from 'react'
import { useAuth } from '../cores/AuthContext'

const MasterAdminDashboard = lazy(() => import('./admin/master-admin/Dashboard'))
const AdminDashboard = lazy(() => import('./admin/Dashboard'))
const StaffDashboard = lazy(() => import('./user/Dashboard'))

export default function Dashboard() {
    const { user } = useAuth()
    const role = user?.role

    if (role === 'master-admin') return <MasterAdminDashboard />
    if (role === 'admin') return <AdminDashboard />
    return <StaffDashboard />
}