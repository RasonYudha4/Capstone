interface Segment {
    value: number
    className: string
}

interface ProgressBarProps {
    segments: Segment[]
    total: number
}

export default function ProgressBar({ segments, total }: ProgressBarProps) {
    return (
        <div className="flex w-full h-3 rounded-full overflow-hidden bg-gray-100">
            {segments.map((seg, i) => (
                <div
                    key={i}
                    className={seg.className}
                    style={{ width: `${(seg.value / total) * 100}%` }}
                />
            ))}
        </div>
    )
}