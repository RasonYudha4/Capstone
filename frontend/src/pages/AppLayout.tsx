import { Outlet } from 'react-router'
import AppSidebar from '../components/organism/AppSidebar'
import { SidebarProvider } from '@/components/ui/sidebar'
import ChatWidget from '@/components/organism/ChatWidget'

export default function AppLayout() {
    return (
        <div className="flex h-screen overflow-hidden">
            <SidebarProvider>
                <AppSidebar />
                <main className="flex-1 overflow-y-auto bg-gray-50 p-6">
                    <Outlet />
                </main>
            </SidebarProvider>
            <ChatWidget />
        </div>
    )
}