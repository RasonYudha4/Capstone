import { InboxIcon, DownloadIcon, FileSearch } from 'lucide-react'
import StatCard from '../molecules/StatCard'

export default function StatRow() {
    return (
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
            <StatCard label="Pending" value={12} weeklyCount={2} weeklyLabel="Minggu ini" icon={InboxIcon} />
            <StatCard label="Dokumen Masuk" value={21} weeklyCount={6} weeklyLabel="Minggu ini" icon={DownloadIcon} />
            <StatCard label="Terakhir dilihat" icon={FileSearch}>
                <div className="flex flex-col">
                    <span className="text-sm font-medium text-gray-800">EP 2 PPI Standar 3.pdf</span>
                    <span className="text-sm text-gray-500">2 Mei 2026</span>
                </div>
            </StatCard>
        </div>
    )
}