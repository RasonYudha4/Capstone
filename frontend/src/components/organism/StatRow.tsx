import { InboxIcon, DownloadIcon, FileSearch } from 'lucide-react'
import StatCard from '../molecules/StatCard'
import { useStats } from '@/hooks/useDocument'
import { useAudit } from '@/hooks/useAudit'
import { useMe } from '@/hooks/useAuth'

interface LastOpened {
    name: string
    date: string
}

function formatDate(iso: string): string {
    return new Date(iso).toLocaleDateString('id-ID', {
        day: 'numeric',
        month: 'long',
        year: 'numeric',
    })
}

function getWeeklyCount(current: number, key: string): number {
    const stored = localStorage.getItem(key)
    const now = Date.now()
    if (stored) {
        const { value, timestamp } = JSON.parse(stored) as { value: number; timestamp: number }
        if (now - timestamp < 7 * 24 * 60 * 60 * 1000) {
            return Math.max(0, current - value)
        }
    }
    localStorage.setItem(key, JSON.stringify({ value: current, timestamp: now }))
    return 0
}

export default function StatRow() {
    const { data: stats } = useStats()
    const { data: audit } = useAudit()
    const { data: me } = useMe()

    const lastOpened: LastOpened | null = (() => {
        if (!audit?.data || !me?.email) return null

        const lastOpen = audit.data
            .filter(a =>
                a.action.toLowerCase() === 'open' &&
                a.username === me.email &&
                a.document_name != null
            )
            .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
            .at(0)

        if (!lastOpen) return null
        return { name: lastOpen.document_name!, date: lastOpen.created_at }
    })()

    const pendingCount = stats?.stats.pending ?? 0
    const totalCount = stats?.total ?? 0
    const weeklyPending = stats ? getWeeklyCount(pendingCount, 'snapshot_pending') : 0
    const weeklyTotal = stats ? getWeeklyCount(totalCount, 'snapshot_total') : 0

    return (
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
            <StatCard
                label="Pending"
                value={pendingCount}
                weeklyCount={weeklyPending}
                weeklyLabel="Minggu ini"
                icon={InboxIcon}
            />
            <StatCard
                label="Dokumen Masuk"
                value={totalCount}
                weeklyCount={weeklyTotal}
                weeklyLabel="Minggu ini"
                icon={DownloadIcon}
            />
            <StatCard label="Terakhir Dilihat" icon={FileSearch}>
                {lastOpened ? (
                    <div className="flex flex-col">
                        <span className="text-sm font-medium text-gray-800 truncate max-w-40">
                            {lastOpened.name}
                        </span>
                        <span className="text-sm text-gray-500">
                            {formatDate(lastOpened.date)}
                        </span>
                    </div>
                ) : (
                    <span className="text-sm text-gray-400">Belum ada</span>
                )}
            </StatCard>
        </div>
    )
}