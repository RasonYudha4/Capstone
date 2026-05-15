import ActivityCard from '@/components/molecules/ActivityCard'
import TimelineItem from '../atoms/TimelineItem'

interface ActivityTimelineItemProps {
    timeLabel: string
    actor: string
    action?: string
    file: string
    isLast: boolean
}

export default function ActivityTimelineItem({
    timeLabel,
    actor,
    action,
    file,
    isLast,
}: ActivityTimelineItemProps) {
    return (
        <div className="flex gap-4">
            <div className="w-14 shrink-0 text-right">
                <span className="text-xs text-gray-400 leading-none mt-2 block">
                    {timeLabel}
                </span>
            </div>

            <TimelineItem
                isLast={isLast}
                dotClassName="bg-[#8571C1]"
                connectorClassName="bg-gray-200"
                className="flex-1"
            >
                <ActivityCard actor={actor} action={action} file={file} />
            </TimelineItem>
        </div>
    )
}