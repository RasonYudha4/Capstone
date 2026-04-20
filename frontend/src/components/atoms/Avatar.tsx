import { UserCircle } from 'lucide-react'

interface AvatarProps {
    className?: string
}

export default function Avatar({ className = 'w-6 h-6 text-gray-400' }: AvatarProps) {
    return <UserCircle className={className} />
}