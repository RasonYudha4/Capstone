import FileStatsBar from '../molecules/FileStatsBar'
import { useStats } from '@/hooks/useDocument'

export default function FileStatsSection() {
    const { data, isLoading } = useStats()

    if (isLoading) {
        return (
            <section className="bg-white rounded-2xl border border-gray-100 p-6 animate-pulse">
                <div className="h-4 bg-gray-200 rounded w-32 mb-2" />
                <div className="h-8 bg-gray-200 rounded w-48 mb-5" />
                <div className="h-4 bg-gray-200 rounded w-full" />
            </section>
        )
    }

    const total    = data?.total          ?? 0
    const approved = data?.stats.approved ?? 0
    const pending  = data?.stats.pending  ?? 0
    const rejected = data?.stats.rejected ?? 0

    return (
        <section className="bg-white rounded-2xl border border-gray-100 p-6">
            <p className="text-sm font-medium text-gray-500 mb-1">Statistik Berkas</p>
            <p className="text-3xl font-bold text-[#6B5FAE] mb-5">
                {total} Total Berkas
            </p>
            <FileStatsBar
                total={total}
                approved={approved}
                pending={pending}
                rejected={rejected}
            />
        </section>
    )
}