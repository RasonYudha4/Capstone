import { useEffect, useState } from 'react'
import { InboxIcon, DownloadIcon, FileSearch } from 'lucide-react'
import StatCard from '../molecules/StatCard'
import { statsService } from '@/services/document_services'
import { auditService } from '@/services/audit_service'
import type { StatsResponse } from '@/dtos/document_dto'

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
    const [stats, setStats]           = useState<StatsResponse | null>(null)
    const [lastOpened, setLastOpened] = useState<LastOpened | null>(null)

    useEffect(() => {
        // fetch stats
        statsService.getStats().then(setStats).catch(console.error)

        // fetch audit and find last "open" for current user
        auditService.getAll().then(res => {
            // get current user email from localStorage/session
            // adjust the key to wherever your auth stores the email
            const currentUser =
                localStorage.getItem('userEmail') ??
                sessionStorage.getItem('userEmail') ??
                null

            const lastOpen = res.data
                .filter(a =>
                    a.action.toLowerCase() === 'open' &&
                    (currentUser ? a.username === currentUser : true)
                )
                // sort descending by created_at — take the most recent
                .sort((a, b) => b.created_at.localeCompare(a.created_at))
                .at(0)

            if (lastOpen) {
                setLastOpened({
                    name: lastOpen.document_name,
                    date: lastOpen.created_at,
                })
            }
        }).catch(console.error)
    }, [])

    const pendingCount  = stats?.stats.pending ?? 0
    const totalCount    = stats?.total         ?? 0
    const weeklyPending = stats ? getWeeklyCount(pendingCount, 'snapshot_pending') : 0
    const weeklyTotal   = stats ? getWeeklyCount(totalCount,   'snapshot_total')   : 0

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
                        <span className="text-sm font-medium text-gray-800 truncate max-w-[160px]">
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