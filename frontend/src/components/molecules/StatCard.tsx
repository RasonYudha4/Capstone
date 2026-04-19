import { type LucideIcon } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import TrendBadge from '../atoms/TrendBadge'

interface StatCardProps {
    label: string
    value?: number
    weeklyCount?: number
    weeklyLabel?: string
    icon: LucideIcon
    children?: React.ReactNode
}

export default function StatCard({
    label,
    value,
    weeklyCount,
    weeklyLabel,
    icon: Icon,
    children,
}: StatCardProps) {
    return (
        <Card className="rounded-2xl border border-gray-100 shadow-none bg-white">
            <CardContent className="p-5 flex items-center justify-between gap-4">
                <div className="flex flex-col gap-1">
                    <span className="text-sm font-semibold text-gray-700">{label}</span>
                    {value !== undefined ? (
                        <div className="flex items-baseline gap-2">
                            <span className="text-3xl font-bold text-gray-900">{value}</span>
                            {weeklyCount !== undefined && weeklyLabel && (
                                <TrendBadge count={weeklyCount} label={weeklyLabel} />
                            )}
                        </div>
                    ) : (
                        children
                    )}
                </div>
                <div className="w-12 h-12 rounded-full bg-[#6B5FAE]/15 flex items-center justify-center shrink-0">
                    <Icon className="w-6 h-6 text-[#6B5FAE]" />
                </div>
            </CardContent>
        </Card>
    )
}