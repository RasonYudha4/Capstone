import { ChartColumnBigIcon } from 'lucide-react'
import GroupStatCard from '../molecules/GroupStatCard'
import SectionHeading from '../atoms/SectionHeading'

const groups = [
    { title: 'Manajemen Rumah Sakit', fileCount: 32, emptyCount: 4 },
    { title: 'Pelayanan Berfokus Pasien', fileCount: 55, emptyCount: 8 },
    { title: 'Sasaran Keselamatan Pasien', fileCount: 14, emptyCount: 8 },
    { title: 'Program Nasional', fileCount: 18, emptyCount: 3 },
]

export default function GroupStatsSection() {
    return (
        <section>
            <SectionHeading icon={ChartColumnBigIcon} title="Statistik Kelompok" />
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
                {groups.map((g) => <GroupStatCard key={g.title} {...g} />)}
            </div>
        </section>
    )
}