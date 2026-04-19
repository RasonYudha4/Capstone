import { Outlet } from 'react-router'
import AppSidebar from '../components/AppSidebar'

export default function AppLayout() {
    return (
        <div className="flex h-screen overflow-hidden">
            <AppSidebar />
            <main className="flex-1 overflow-y-auto bg-gray-50 p-6">
                <Outlet />
            </main>
        </div>
    )
}