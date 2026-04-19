import { TrendingUp } from 'lucide-react'

interface TrendBadgeProps {
    count: number
    label: string
}

export default function TrendBadge({ count, label }: TrendBadgeProps) {
    return (
        <span className="inline-flex items-center gap-0.5 text-xs text-gray-500">
            <TrendingUp className="w-3 h-3 text-[#6B5FAE]" />
            {count}↑ {label}
        </span>
    )
}