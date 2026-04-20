import SectionHeading from '@/components/atoms/SectionHeading'
import ActivityEntry from '@/components/molecules/ActivityEntry'
import { Repeat } from 'lucide-react'

const activities = [
    { timestamp: '12 April 2026 13:45', actor: 'Supriyadi', action: '', file: 'EP 2 PMKP Standar 2.pdf' },
    { timestamp: '8 April 2026 14:00', actor: 'Pardi', action: 'mengupload file', file: 'EP 3 PAB Standar 2.pdf' },
    { timestamp: '8 April 2026 14:00', actor: 'Priyadi', action: 'mengupload file', file: 'EP 3 PAB Standar 4.pdf' },
    { timestamp: '8 April 2026 14:00', actor: 'Pardi', action: 'mengupload file', file: 'EP 2 PAB Standar 2.pdf' },
]

export default function ActivityWidget() {
    return (
        <div className=" bg-linear-to-b from-[#8571C1] to-75% to-[#AD9DDB] rounded-2xl p-5 h-full">
            <SectionHeading icon={Repeat} title="Riwayat Aktivitas" iconClassName="text-white" titleClassName="text-white" />
            <div className="mt-2">
                {activities.map((item, i) => (
                    <ActivityEntry key={i} {...item} isLast={i === activities.length - 1} />
                ))}
            </div>
        </div>
    )
}