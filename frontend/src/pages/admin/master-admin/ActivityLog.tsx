import { History, RefreshCw, AlertCircle } from 'lucide-react'
import SectionHeading from '@/components/atoms/SectionHeading'
import ActivityDateGroup from '@/components/organism/ActivityDateGroup'
import { useActivityLog } from '@/hooks/useAudit'

export default function ActivityLog() {
    const { groups, isLoading, isError, error, refetch } = useActivityLog()

    return (
        <section>
            <SectionHeading icon={History} title="Riwayat Aktivitas" />

            <div className="bg-white rounded-2xl border border-gray-100 p-6 min-h-172">
                <div className="overflow-y-auto max-h-160 pr-2">

                    {isError && (
                        <div className="flex items-center gap-3 rounded-xl border border-red-100 bg-red-50 px-4 py-3 mb-4 text-sm text-red-600">
                            <AlertCircle className="size-4 shrink-0" />
                            <span className="flex-1">
                                Gagal memuat aktivitas.{' '}
                                {error instanceof Error ? error.message : 'Terjadi kesalahan.'}
                            </span>
                            <button
                                onClick={() => refetch()}
                                className="flex items-center gap-1.5 rounded-lg bg-red-100 px-3 py-1 text-xs font-medium text-red-700 hover:bg-red-200 transition-colors"
                            >
                                <RefreshCw className="size-3" />
                                Coba lagi
                            </button>
                        </div>
                    )}

                    {isLoading && (
                        <div className="flex items-center justify-center gap-2 py-12">
                            <svg className="animate-spin size-4 text-gray-400" viewBox="0 0 24 24" fill="none">
                                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4l3-3-3-3v4a8 8 0 00-8 8h4z" />
                            </svg>
                            <p className="text-xs text-gray-400">Memuat aktivitas...</p>
                        </div>
                    )}

                    {!isLoading && groups.map(group => (
                        <ActivityDateGroup key={group.date} {...group} />
                    ))}

                    {!isLoading && !isError && groups.length > 0 && (
                        <p className="text-center text-xs text-gray-400 py-4">
                            Semua aktivitas telah dimuat
                        </p>
                    )}

                    {!isLoading && !isError && groups.length === 0 && (
                        <p className="text-center text-xs text-gray-400 py-8">
                            Belum ada aktivitas
                        </p>
                    )}

                </div>
            </div>
        </section>
    )
}