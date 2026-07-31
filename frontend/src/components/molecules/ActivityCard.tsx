import GradientCard from '@/components/atoms/GradientCard'

interface ActivityCardProps {
    actor: string
    action?: string
    file: string | null
}

export default function ActivityCard({ actor, action, file }: ActivityCardProps) {
    return (
        <GradientCard>
            <p className="text-sm font-bold text-white">
                {actor}
                {action && <span className="font-normal"> {action}</span>}
            </p>
            <p className="text-xs text-white/70 mt-1 truncate">"{file}"</p>
        </GradientCard>
    )
}