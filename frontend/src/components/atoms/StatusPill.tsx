export type FileStatus = 'approved' | 'review' | 'pending' | 'rejected'

const statusConfig: Record<FileStatus, { label: string; className: string }> = {
    approved: { label: 'Sudah di approve', className: 'bg-[#6B5FAE]/20 text-[#6B5FAE]' },
    review: { label: 'Sedang di review', className: 'bg-[#3B2F6E]/20 text-[#3B2F6E]' },
    pending: { label: 'Dalam antrian', className: 'bg-gray-200 text-gray-500' },
    rejected: { label: 'Ditolak', className: 'bg-red-800/20 text-red-800' },
}

interface StatusPillProps {
    status: FileStatus
}

export default function StatusPill({ status }: StatusPillProps) {
    const { label, className } = statusConfig[status]
    return (
        <span className={`inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium ${className}`}>
            {label}
        </span>
    )
}