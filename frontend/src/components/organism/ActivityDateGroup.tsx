'use client'

import { useState } from 'react'
import { cn } from '@/lib/utils'
import ChevronToggle from '@/components/atoms/ChevronToggle'
import ActivityTimelineItem from '@/components/molecules/ActivityTimelineItem'

interface Activity {
    id: string
    timeLabel: string
    actor: string
    action?: string
    file: string
}

interface ActivityDateGroupProps {
    date: string
    items: Activity[]
}

export default function ActivityDateGroup({ date, items }: ActivityDateGroupProps) {
    const [open, setOpen] = useState(true)

    return (
        <div className="mb-6">
            <button
                onClick={() => setOpen(o => !o)}
                className="flex items-center gap-2 w-full mb-3"
            >
                <ChevronToggle open={open} />
                <span className="text-sm font-semibold text-gray-700">{date}</span>
                <div className="flex-1 h-px bg-gray-100" />
                <span className="text-xs text-gray-400 ml-2">{items.length} aktivitas</span>
            </button>

            <div
                className={cn(
                    'grid transition-[grid-template-rows] duration-300 ease-in-out',
                    open ? 'grid-rows-[1fr]' : 'grid-rows-[0fr]'
                )}
            >
                <div className="overflow-hidden">
                    <div className="ml-6">
                        {items.map((item, i) => (
                            <ActivityTimelineItem
                                key={item.id}
                                isLast={i === items.length - 1}
                                {...item}
                            />
                        ))}
                    </div>
                </div>
            </div>
        </div>
    )
}