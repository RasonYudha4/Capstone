import { Users } from 'lucide-react'

interface GroupStatCardProps {
    title: string
    fileCount: number
    emptyCount: number
}

export default function GroupStatCard({ title, fileCount, emptyCount }: GroupStatCardProps) {
    return (
        <div className="bg-[#6B5FAE] rounded-2xl p-4 flex flex-col gap-3">
            {/* Icon + title */}
            <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-lg bg-white/20 flex items-center justify-center shrink-0">
                    <Users className="w-4 h-4 text-white" />
                </div>
                <p className="text-sm font-semibold text-white leading-snug">{title}</p>
            </div>

            {/* Stats */}
            <div className="flex justify-between gap-6">
                <div className=' flex flex-col items-center justify-center'>
                    <p className="text-2xl font-bold text-white">{fileCount}</p>
                    <p className="text-xs text-white/70">File terupload</p>
                </div>
                <div className=' flex flex-col items-center justify-center'>
                    <p className="text-2xl font-bold text-white">{emptyCount}</p>
                    <p className="text-xs text-white/70">Bagian kosong</p>
                </div>
            </div>
        </div>
    )
}