import ProgressBar from '../atoms/ProgressBar'
import LegendDot from '../atoms/LegendDot'

interface FileStatsBarProps {
    total: number
    approved: number
    pending: number
    rejected: number
}

export default function FileStatsBar({ total, approved, pending, rejected }: FileStatsBarProps) {
    const segments = [
        { value: approved, className: 'bg-[#6B5FAE]' },
        { value: pending,  className: 'bg-gray-200' },
        { value: rejected, className: 'bg-red-800' },
    ]

    const legends = [
        { label: 'Sudah di approve', className: 'bg-[#6B5FAE]' },
        { label: 'Dalam antrian',   className: 'bg-gray-200 border border-gray-300' },
        { label: 'Ditolak',         className: 'bg-red-800' },
    ]

    return (
        <div className="flex flex-col gap-3">
            <ProgressBar segments={segments} total={total} />
            <div className="flex items-center gap-4 flex-wrap">
                {legends.map((l) => <LegendDot key={l.label} {...l} />)}
            </div>
        </div>
    )
}