import { ChartColumnBigIcon } from 'lucide-react'
import GroupStatCard from '../molecules/GroupStatCard'
import SectionHeading from '../atoms/SectionHeading'
import { useStats } from '@/hooks/useDocument'
 
export default function GroupStatsSection() {
    const { data, isLoading } = useStats()
 
    return (
        <section>
            <SectionHeading icon={ChartColumnBigIcon} title="Statistik Kelompok" />
 
            {isLoading ? (
                <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
                    {Array.from({ length: 4 }).map((_, i) => (
                        <div
                            key={i}
                            className="bg-[#6B5FAE]/20 rounded-2xl h-32 animate-pulse"
                        />
                    ))}
                </div>
            ) : (
                <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
                    {(data?.groups ?? []).map((g) => (
                        <GroupStatCard
                            key={g.group_id}
                            title={g.group_name}
                            fileCount={g.total_files}
                            emptyCount={g.empty_sections}
                        />
                    ))}
                </div>
            )}
        </section>
    )
}