import { type LucideIcon, X } from 'lucide-react'

interface PanelHeaderProps {
    icon: LucideIcon
    title: string
    onClose?: () => void
    className?: string
}

export default function PanelHeader({
    icon: Icon,
    title,
    onClose,
    className = 'bg-[#6B5FAE]',
}: PanelHeaderProps) {
    return (
        <div className={`flex items-center justify-between px-4 py-3.5 ${className}`}>
            <div className="flex items-center gap-2.5">
                <div className="w-8 h-8 rounded-full bg-white/20 flex items-center justify-center">
                    <Icon className="w-4 h-4 text-white" />
                </div>
                <span className="text-white font-semibold text-sm">{title}</span>
            </div>
            {onClose && (
                <button
                    onClick={onClose}
                    aria-label="Tutup"
                    className="text-white/70 hover:text-white transition-colors"
                >
                    <X className="w-5 h-5" />
                </button>
            )}
        </div>
    )
}