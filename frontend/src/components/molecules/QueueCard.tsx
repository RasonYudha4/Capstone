import Avatar from '../atoms/Avatar'
import GradientCard from '../atoms/GradientCard'

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
        <GradientCard onClick={onClick}>
            <p className="text-xs text-white font-medium mb-1">{category}</p>
            <h3 className="text-sm font-bold text-white">{title}</h3>
            <div className="flex items-center gap-2 mt-2">
                <Avatar className="w-5 h-5 text-white" />
                <span className="text-sm font-medium text-white">{submittedBy}</span>
                <span className="text-xs text-white">{date}</span>
            </div>
        </GradientCard>
    )
}