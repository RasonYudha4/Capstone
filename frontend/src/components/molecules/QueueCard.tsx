import { Card, CardContent } from '@/components/ui/card'
import Avatar from '../atoms/Avatar'

interface QueueCardProps {
    category: string
    title: string
    submittedBy: string
    date: string
    onClick?: () => void
}

export default function QueueCard({
    category,
    title,
    submittedBy,
    date,
    onClick,
}: QueueCardProps) {
    return (
        <Card
            onClick={onClick}
            className="rounded-2xl border-0 shadow-none bg-linear-to-r from-[#8571C1] from-30% to-[#AD9DDB] transition-colors cursor-pointer hover:brightness-90 hover:shadow-md"
        >
            <CardContent className="p-4">
                <p className="text-xs text-white font-medium mb-1">{category}</p>
                <h3 className="text-sm font-bold text-white">{title}</h3>
                <div className="flex items-center gap-2 mt-2">
                    <Avatar className="w-5 h-5 text-white" />
                    <span className="text-sm font-medium text-white">{submittedBy}</span>
                    <span className="text-xs text-white">{date}</span>
                </div>
            </CardContent>
        </Card>
    )
}