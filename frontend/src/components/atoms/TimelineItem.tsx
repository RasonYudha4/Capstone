interface TimelineItemProps {
    children: React.ReactNode
    isLast?: boolean
}

export default function TimelineItem({ children, isLast }: TimelineItemProps) {
    return (
        <div className="flex gap-3">
            <div className="flex flex-col items-center">
                <div className="w-2 h-2 rounded-full bg-white mt-1 shrink-0" />
                {!isLast && <div className="w-px flex-1 bg-white/30 mt-1" />}
            </div>
            <div className="pb-5">{children}</div>
        </div>
    )
}