import { type LucideIcon } from 'lucide-react'

interface SectionHeadingProps {
    icon: LucideIcon
    title: string
    iconClassName?: string
    titleClassName?: string
}

export default function SectionHeading({
    icon: Icon,
    title,
    iconClassName = 'text-[#6B5FAE]',
    titleClassName = 'text-gray-900',
}: SectionHeadingProps) {
    return (
        <div className="flex items-center gap-2 mb-4">
            <Icon className={`w-5 h-5 ${iconClassName}`} />
            <h2 className={`font-semibold text-base ${titleClassName}`}>{title}</h2>
        </div>
    )
}