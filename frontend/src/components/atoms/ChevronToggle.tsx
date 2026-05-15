import { ChevronDown } from 'lucide-react'
import { cn } from '@/lib/utils'

interface ChevronToggleProps {
    open: boolean
    className?: string
}

export default function ChevronToggle({ open, className }: ChevronToggleProps) {
    return (
        <ChevronDown
            className={cn(
                'w-4 h-4 text-gray-400 transition-transform duration-200 shrink-0',
                !open && '-rotate-90',
                className
            )}
        />
    )
}