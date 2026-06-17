import { createBrowserRouter } from 'react-router'
import { lazy } from 'react'
import AuthGuard from '../cores/AuthGuard'
import wrap from '../lib/component-wrapper-helper'
import AppLayout from '../pages/AppLayout'

const Login = lazy(() => import('../pages/auth/Login'))
const Verification = lazy(() => import('../pages/auth/Verification'))
const UserLanding = lazy(() => import('../pages/user/Dashboard'))
const Dashboard = lazy(() => import('../pages/Dashboard'))
const Unauthorized = lazy(() => import('../pages/Unauthorized'))
const NotFound = lazy(() => import('../pages/Notfound'))
const Storage = lazy(() => import('../pages/admin/Storage'))
const Admins = lazy(() => import('../pages/admin/master-admin/Admins'))
const Activity = lazy(() => import('../pages/admin/master-admin/ActivityLog'))

export const router = createBrowserRouter([
    // Public landing page — no auth required
    { path: '/', element: wrap(UserLanding) },

    { path: '/login', element: wrap(Login) },
    { path: '/verify', element: wrap(Verification) },

    // All authenticated routes share ONE layout
    {
        element: <AuthGuard allowedRoles={['staff', 'admin', 'master-admin']} />,
        children: [
            {
                element: <AppLayout />,
                children: [
                    // Accessible by all roles
                    { path: '/dashboard', element: wrap(Dashboard) },

                    // Admin + master-admin only
                    {
                        element: <AuthGuard allowedRoles={['admin', 'master-admin']} />,
                        children: [
                            { path: '/storage', element: wrap(Storage) },
                        ],
                    },

                    // Master-admin only
                    {
                        element: <AuthGuard allowedRoles={['master-admin']} />,
                        children: [
                            { path: '/admins', element: wrap(Admins) },
                            { path: '/activity-log', element: wrap(Activity) }
                        ],
                    },
                ],
            },
        ],
    },

    { path: '/unauthorized', element: wrap(Unauthorized) },
    { path: '*', element: wrap(NotFound) },
])