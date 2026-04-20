interface LegendDotProps {
    label: string
    className: string
}

export default function LegendDot({ label, className }: LegendDotProps) {
    return (
        <div className="flex items-center gap-1.5">
            <span className={`w-3 h-3 rounded-sm shrink-0 ${className}`} />
            <span className="text-xs text-gray-500">{label}</span>
        </div>
    )
}