interface TimelineItemProps {
    children: React.ReactNode
    isLast?: boolean
    dotClassName?: string
    connectorClassName?: string
    className?: string
}

export default function TimelineItem({
    children,
    isLast,
    dotClassName = 'bg-white',
    connectorClassName = 'bg-white/30',
    className
}: TimelineItemProps) {
    return (
        <div className={`flex gap-3 ${className ?? ''}`}>
            <div className="flex flex-col items-center">
                <div className={`w-2 h-2 rounded-full mt-1 shrink-0 ${dotClassName}`} />
                {!isLast && <div className={`w-px flex-1 mt-1 ${connectorClassName}`} />}
            </div>
            <div className="pb-5 flex-1">{children}</div>
        </div>
    )
}