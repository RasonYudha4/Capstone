import { createBrowserRouter } from 'react-router'
import { lazy } from 'react'
import AuthGuard from '../cores/AuthGuard'
import wrap from '../helpers/component-wrapper-helper'
import Root from '../pages/Root'
import AppLayout from '../pages/AppLayout'

const Login = lazy(() => import('../pages/auth/Login'))
const Dashboard = lazy(() => import('../pages/Dashboard'))
const Unauthorized = lazy(() => import('../pages/Unauthorized'))
const NotFound = lazy(() => import('../pages/Notfound'))

export const router = createBrowserRouter([
    { path: '/', element: <Root /> },
    { path: '/login', element: wrap(Login) },

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
                            // { path: '/admin/users',    element: wrap(AdminUsers) },
                        ],
                    },

                    // Master-admin only
                    {
                        element: <AuthGuard allowedRoles={['master-admin']} />,
                        children: [
                            // { path: '/master-admin/roles',  element: wrap(MasterRoles) },
                            // { path: '/master-admin/admins', element: wrap(MasterAdmins) },
                        ],
                    },
                ],
            },
        ],
    },

    { path: '/unauthorized', element: wrap(Unauthorized) },
    { path: '*', element: wrap(NotFound) },
])