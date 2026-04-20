import NotificationDropdown from './NotificationDropdown'

interface DashboardHeaderProps {
    name: string
    pendingCount: number
    onViewAllNotifications?: () => void
}

export default function DashboardHeader({ name, pendingCount, onViewAllNotifications }: DashboardHeaderProps) {
    return (
        <div className="flex items-start justify-between mb-6">
            <div>
                <h1 className="text-2xl font-bold text-gray-900">Selamat Datang Kembali, {name}</h1>
                <p className="text-sm text-gray-500 mt-1">Terdapat {pendingCount} berkas baru yang perlu diapprove</p>
            </div>
            <div className="mt-1">
                <NotificationDropdown onViewAll={() => { onViewAllNotifications?.(); console.log('view all') }} />
            </div>
        </div>
    )
}