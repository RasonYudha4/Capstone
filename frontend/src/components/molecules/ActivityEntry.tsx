import TimelineItem from '../atoms/TimelineItem'

interface ActivityEntryProps {
    timestamp: string
    actor: string
    action?: string
    file: string | null
    isLast?: boolean
}

export default function ActivityEntry({
    timestamp,
    actor,
    action,
    file,
    isLast,
}: ActivityEntryProps) {
    return (
        <TimelineItem isLast={isLast}>
            <p className="text-xs text-white/70">{timestamp}</p>
            <p className="text-sm text-white mt-0.5">
                <span className="font-bold">{actor}</span>
                {action && <span> {action}</span>}
            </p>
            {file ? (
                <p className="text-xs text-white/70 mt-0.5">"{file}"</p>
            ) : null}
        </TimelineItem>
    )
}