import { cn } from '@/lib/utils'

interface GradientCardProps {
    children: React.ReactNode
    onClick?: () => void
    className?: string
}

export default function GradientCard({ children, onClick, className }: GradientCardProps) {
    return (
        <div
            onClick={onClick}
            className={cn(
                'rounded-2xl bg-linear-to-r from-[#8571C1] from-30% to-[#AD9DDB]',
                'p-4 transition-[filter] cursor-pointer hover:brightness-90 hover:shadow-md',
                className
            )}
        >
            {children}
        </div>
    )
}