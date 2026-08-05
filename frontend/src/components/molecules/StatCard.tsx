import { type LucideIcon } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import TrendBadge from '../atoms/TrendBadge'
import { cn } from '@/lib/utils'

interface StatCardProps {
    label: string
    value?: number
    weeklyCount?: number
    weeklyLabel?: string
    icon: LucideIcon
    children?: React.ReactNode
    /** Blinking red notification indicator on the icon */
    showDot?: boolean
    onDotClick?: () => void
    /** Make the whole card clickable */
    onClick?: () => void
    clickable?: boolean
}

export default function StatCard({
    label,
    value,
    weeklyCount,
    weeklyLabel,
    icon: Icon,
    children,
    showDot = false,
    onDotClick,
    onClick,
    clickable = false,
}: StatCardProps) {
    const isInteractive = clickable && !!onClick

    return (
        <Card
            className={cn(
                'rounded-2xl border border-gray-100 shadow-none bg-white',
                isInteractive && 'cursor-pointer transition-colors hover:border-[#6B5FAE]/30 hover:bg-[#6B5FAE]/[0.02]',
            )}
            onClick={isInteractive ? onClick : undefined}
            role={isInteractive ? 'button' : undefined}
            tabIndex={isInteractive ? 0 : undefined}
            onKeyDown={
                isInteractive
                    ? (e) => {
                        if (e.key === 'Enter' || e.key === ' ') {
                            e.preventDefault()
                            onClick?.()
                        }
                    }
                    : undefined
            }
        >
            <CardContent className="p-5 flex items-center justify-between gap-4">
                <div className="flex flex-col gap-1 min-w-0">
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
                <div className="relative w-12 h-12 rounded-full bg-[#6B5FAE]/15 flex items-center justify-center shrink-0">
                    <Icon className="w-6 h-6 text-[#6B5FAE]" />
                    {showDot && (
                        <button
                            type="button"
                            onClick={(e) => {
                                e.stopPropagation()
                                onDotClick?.()
                            }}
                            aria-label="Buka dokumen dari notifikasi"
                            className="absolute -top-0.5 -right-0.5 flex h-3 w-3 items-center justify-center rounded-full focus:outline-none focus-visible:ring-2 focus-visible:ring-red-400"
                        >
                            <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-red-400 opacity-75" />
                            <span className="relative block h-2.5 w-2.5 rounded-full bg-red-500 ring-2 ring-white" />
                        </button>
                    )}
                </div>
            </CardContent>
        </Card>
    )
}
