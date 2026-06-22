import SectionHeading from '@/components/atoms/SectionHeading'
import ActivityEntry from '@/components/molecules/ActivityEntry'
import { Repeat } from 'lucide-react'
import { useActivityLog } from '@/hooks/useAudit'

export default function ActivityWidget() {
    const { groups, isLoading } = useActivityLog(5)

    return (
        <div className="bg-linear-to-b from-[#8571C1] to-75% to-[#AD9DDB] rounded-2xl p-5 h-full">
            <SectionHeading icon={Repeat} title="Riwayat Aktivitas" iconClassName="text-white" titleClassName="text-white" />
            <div className="mt-2">
                {isLoading ? (
                    <p className="text-sm text-white/60 text-center py-4">Memuat aktivitas...</p>
                ) : groups.length === 0 ? (
                    <p className="text-sm text-white/60 text-center py-4">Belum ada aktivitas.</p>
                ) : (
                    groups.map((group) =>
                        group.items.map((item, i) => {
                            const isLastInGroup = i === group.items.length - 1
                            const isLastGroup = group === groups[groups.length - 1]
                            return (
                                <ActivityEntry
                                    key={item.id}
                                    timestamp={item.timestamp}
                                    actor={item.actor}
                                    action={item.action}
                                    file={item.file}
                                    isLast={isLastInGroup && isLastGroup}
                                />
                            )
                        })
                    )
                )}
            </div>
        </div>
    )
}