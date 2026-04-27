import { ChevronRight } from 'lucide-react'
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

export interface BreadcrumbOption {
    label: string
    value: string
}

export interface BreadcrumbSegment {
    selected: string
    options: BreadcrumbOption[]
    onChange: (value: string) => void
}

interface BreadcrumbNavProps {
    segments: BreadcrumbSegment[]
}

export default function BreadcrumbNav({ segments }: BreadcrumbNavProps) {
    return (
        <div className="flex items-center gap-2 flex-wrap">
            {segments.map((segment, i) => (
                <div key={i} className="flex items-center gap-2">
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <button className="text-base font-semibold text-gray-900 hover:text-[#6B5FAE] transition-colors flex items-center gap-1 outline-none">
                                {segment.selected}
                                <svg className="w-3.5 h-3.5 text-gray-400 mt-0.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.5}>
                                    <path d="m6 9 6 6 6-6" />
                                </svg>
                            </button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="start" className="rounded-xl border border-gray-100 shadow-lg min-w-45">
                            {segment.options.map((opt) => (
                                <DropdownMenuItem
                                    key={opt.value}
                                    onClick={() => segment.onChange(opt.value)}
                                    className={`text-sm cursor-pointer rounded-lg ${opt.label === segment.selected
                                            ? 'text-[#6B5FAE] font-semibold bg-[#6B5FAE]/8'
                                            : 'text-gray-700'
                                        }`}
                                >
                                    {opt.label}
                                </DropdownMenuItem>
                            ))}
                        </DropdownMenuContent>
                    </DropdownMenu>

                    {i < segments.length - 1 && (
                        <ChevronRight className="w-4 h-4 text-gray-400" />
                    )}
                </div>
            ))}
        </div>
    )
}