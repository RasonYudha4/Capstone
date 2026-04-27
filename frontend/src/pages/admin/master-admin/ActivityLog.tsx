'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import SectionHeading from '@/components/atoms/SectionHeading'
import { History } from 'lucide-react'
import ActivityDateGroup from '@/components/organism/ActivityDateGroup'

interface Activity {
    id: string
    timestamp: string       // e.g. "12 April 2026 13:45"
    date: string            // e.g. "12 April 2026" — used for grouping
    timeLabel: string       // e.g. "13:45" — shown on the timeline
    actor: string
    action?: string
    file: string
}

interface ActivityGroup {
    date: string
    items: Activity[]
}

export const mockActivities = [
    { id: '1', date: '27 April 2026', timeLabel: '13:45', timestamp: '27 April 2026 13:45', actor: 'Supriyadi', action: 'mengupload file', file: 'EP 2 PMKP Standar 2.pdf' },
    { id: '2', date: '27 April 2026', timeLabel: '11:20', timestamp: '27 April 2026 11:20', actor: 'Pardi', action: 'mengupload file', file: 'EP 3 PAB Standar 2.pdf' },
    { id: '3', date: '27 April 2026', timeLabel: '09:05', timestamp: '27 April 2026 09:05', actor: 'Priyadi', action: 'menghapus file', file: 'EP 3 PAB Standar 4.pdf' },
    { id: '4', date: '26 April 2026', timeLabel: '16:30', timestamp: '26 April 2026 16:30', actor: 'Sari', action: 'mengupload file', file: 'EP 1 PMKP Standar 1.pdf' },
    { id: '5', date: '26 April 2026', timeLabel: '14:00', timestamp: '26 April 2026 14:00', actor: 'Pardi', action: 'mengupload file', file: 'EP 2 PAB Standar 2.pdf' },
    { id: '6', date: '26 April 2026', timeLabel: '10:15', timestamp: '26 April 2026 10:15', actor: 'Budi', action: 'mengedit file', file: 'EP 5 PAB Standar 1.pdf' },
    { id: '7', date: '25 April 2026', timeLabel: '15:45', timestamp: '25 April 2026 15:45', actor: 'Supriyadi', action: 'mengupload file', file: 'EP 4 PMKP Standar 3.pdf' },
    { id: '8', date: '25 April 2026', timeLabel: '13:00', timestamp: '25 April 2026 13:00', actor: 'Priyadi', action: 'mengunduh file', file: 'EP 2 PAB Standar 5.pdf' },
    { id: '9', date: '25 April 2026', timeLabel: '08:30', timestamp: '25 April 2026 08:30', actor: 'Sari', action: 'menghapus file', file: 'EP 1 PAB Standar 2.pdf' },
    { id: '10', date: '24 April 2026', timeLabel: '17:00', timestamp: '24 April 2026 17:00', actor: 'Budi', action: 'mengupload file', file: 'EP 3 PMKP Standar 1.pdf' },
    { id: '11', date: '24 April 2026', timeLabel: '11:45', timestamp: '24 April 2026 11:45', actor: 'Pardi', action: 'mengedit file', file: 'EP 6 PAB Standar 3.pdf' },
    { id: '12', date: '24 April 2026', timeLabel: '09:20', timestamp: '24 April 2026 09:20', actor: 'Supriyadi', action: 'mengunduh file', file: 'EP 2 PMKP Standar 4.pdf' },
    { id: '13', date: '23 April 2026', timeLabel: '14:10', timestamp: '23 April 2026 14:10', actor: 'Priyadi', action: 'mengupload file', file: 'EP 1 PAB Standar 3.pdf' },
    { id: '14', date: '23 April 2026', timeLabel: '10:00', timestamp: '23 April 2026 10:00', actor: 'Sari', action: 'mengedit file', file: 'EP 4 PAB Standar 2.pdf' },
    { id: '15', date: '22 April 2026', timeLabel: '16:00', timestamp: '22 April 2026 16:00', actor: 'Budi', action: 'mengupload file', file: 'EP 5 PMKP Standar 2.pdf' },
    { id: '16', date: '22 April 2026', timeLabel: '13:30', timestamp: '22 April 2026 13:30', actor: 'Pardi', action: 'menghapus file', file: 'EP 3 PAB Standar 1.pdf' },
    { id: '17', date: '21 April 2026', timeLabel: '11:00', timestamp: '21 April 2026 11:00', actor: 'Supriyadi', action: 'mengupload file', file: 'EP 2 PAB Standar 4.pdf' },
    { id: '18', date: '21 April 2026', timeLabel: '08:45', timestamp: '21 April 2026 08:45', actor: 'Priyadi', action: 'mengunduh file', file: 'EP 6 PMKP Standar 1.pdf' },
    { id: '19', date: '20 April 2026', timeLabel: '15:20', timestamp: '20 April 2026 15:20', actor: 'Sari', action: 'mengedit file', file: 'EP 1 PMKP Standar 3.pdf' },
    { id: '20', date: '20 April 2026', timeLabel: '09:50', timestamp: '20 April 2026 09:50', actor: 'Budi', action: 'mengupload file', file: 'EP 4 PAB Standar 1.pdf' },
    { id: '21', date: '19 April 2026', timeLabel: '14:30', timestamp: '19 April 2026 14:30', actor: 'Pardi', action: 'mengupload file', file: 'EP 2 PMKP Standar 5.pdf' },
    { id: '22', date: '19 April 2026', timeLabel: '10:15', timestamp: '19 April 2026 10:15', actor: 'Supriyadi', action: 'menghapus file', file: 'EP 3 PAB Standar 2.pdf' },
    { id: '23', date: '18 April 2026', timeLabel: '16:45', timestamp: '18 April 2026 16:45', actor: 'Priyadi', action: 'mengupload file', file: 'EP 5 PAB Standar 3.pdf' },
    { id: '24', date: '18 April 2026', timeLabel: '13:00', timestamp: '18 April 2026 13:00', actor: 'Sari', action: 'mengedit file', file: 'EP 1 PAB Standar 4.pdf' },
    { id: '25', date: '17 April 2026', timeLabel: '09:30', timestamp: '17 April 2026 09:30', actor: 'Budi', action: 'mengunduh file', file: 'EP 6 PAB Standar 2.pdf' },
]

const PAGE_SIZE = 20

async function fetchActivities(page: number): Promise<{ data: Activity[]; hasMore: boolean }> {
    await new Promise(resolve => setTimeout(resolve, 400))

    const start = (page - 1) * PAGE_SIZE
    const end = start + PAGE_SIZE
    const data = mockActivities.slice(start, end)

    return {
        data,
        hasMore: end < mockActivities.length,
    }
}

function groupByDate(activities: Activity[]): ActivityGroup[] {
    const map = new Map<string, Activity[]>()
    for (const a of activities) {
        if (!map.has(a.date)) map.set(a.date, [])
        map.get(a.date)!.push(a)
    }
    return Array.from(map.entries()).map(([date, items]) => ({ date, items }))
}

export default function ActivityLog() {
    const [groups, setGroups] = useState<ActivityGroup[]>([])
    const [page, setPage] = useState(1)
    const [hasMore, setHasMore] = useState(true)
    const [loading, setLoading] = useState(false)
    const bottomRef = useRef<HTMLDivElement>(null)

    const load = useCallback(async (p: number) => {
        if (loading) return
        setLoading(true)
        const { data, hasMore } = await fetchActivities(p)
        setGroups(prev => {
            // Merge new items into existing groups
            const allItems = prev.flatMap(g => g.items).concat(data)
            return groupByDate(allItems)
        })
        setHasMore(hasMore)
        setLoading(false)
    }, [loading])

    // Initial load
    useEffect(() => { load(1) }, [])

    // Infinite scroll observer
    useEffect(() => {
        if (!bottomRef.current) return
        const observer = new IntersectionObserver(
            entries => {
                if (entries[0].isIntersecting && hasMore && !loading) {
                    setPage(p => {
                        const next = p + 1
                        load(next)
                        return next
                    })
                }
            },
            { threshold: 0.1 }
        )
        observer.observe(bottomRef.current)
        return () => observer.disconnect()
    }, [hasMore, loading, load])

    return (
        <section>
            <SectionHeading icon={History} title="Riwayat Aktivitas" />
            <div className="bg-white rounded-2xl border border-gray-100 p-6 min-h-172">
                <div className="overflow-y-auto max-h-160 pr-2">

                    {groups.map(group => (
                        <ActivityDateGroup key={group.date} {...group} />
                    ))}

                    <div ref={bottomRef} className="h-4" />

                    {loading && (
                        <p className="text-center text-xs text-gray-400 py-4">Memuat...</p>
                    )}
                    {!hasMore && groups.length > 0 && (
                        <p className="text-center text-xs text-gray-400 py-4">
                            Semua aktivitas telah dimuat
                        </p>
                    )}
                    {!loading && groups.length === 0 && (
                        <p className="text-center text-xs text-gray-400 py-8">
                            Belum ada aktivitas
                        </p>
                    )}

                </div>
            </div>
        </section>
    )
}