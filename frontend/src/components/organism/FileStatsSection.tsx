import FileStatsBar from '../molecules/FileStatsBar'

const stats = {
    total: 144,
    approved: 90,
    review: 30,
    pending: 14,
    rejected: 10,
}

export default function FileStatsSection() {
    return (
        <section className="bg-white rounded-2xl border border-gray-100 p-6">
            <p className="text-sm font-medium text-gray-500 mb-1">Statistik Berkas</p>
            <p className="text-3xl font-bold text-[#6B5FAE] mb-5">
                {stats.total} Total Berkas
            </p>
            <FileStatsBar {...stats} />
        </section>
    )
}