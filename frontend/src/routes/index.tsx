import { createBrowserRouter } from 'react-router'
import { lazy } from 'react'
import AuthGuard from '../guards/AuthGuard'
import wrap from '../helpers/component-wrapper-helper'

import AdminLayout from '../layouts/AdminLayout'
import MasterAdminLayout from '../layouts/MasterAdminLayout'
import UserLayout from '../layouts/UserLayout'
import Root from '../pages/Root'

const Login = lazy(() => import('../pages/auth/Login'))
const AdminDashboard = lazy(() => import('../pages/admin/Dashboard'))
const MasterDashboard = lazy(() => import('../pages/master-admin/Dashboard'))
const UserDashboard = lazy(() => import('../pages/user/Dashboard'))
const Unauthorized = lazy(() => import('../pages/Unauthorized'))
const NotFound = lazy(() => import('../pages/Notfound'))

export const router = createBrowserRouter([
    { path: '/', element: <Root /> },
    { path: '/login', element: wrap(Login) },

    {
        element: <AuthGuard allowedRoles={['admin']} />,
        children: [
            {
                element: <AdminLayout />,
                children: [
                    { path: '/admin', element: wrap(AdminDashboard) },
                ]
            }
        ]
    },

    // Master Admin
    {
        element: <AuthGuard allowedRoles={['master-admin']} />,
        children: [
            {
                element: <MasterAdminLayout />,
                children: [
                    { path: '/master-admin', element: wrap(MasterDashboard) },
                ]
            }
        ]
    },

    // User
    {
        element: <AuthGuard allowedRoles={['user']} />,
        children: [
            {
                element: <UserLayout />,
                children: [
                    { path: '/user', element: wrap(UserDashboard) },
                ]
            }
        ]
    },

    { path: '/unauthorized', element: wrap(Unauthorized) },
    { path: '*', element: wrap(NotFound) },
])