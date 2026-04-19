import { formatDistanceToNow } from 'date-fns'
import { id } from 'date-fns/locale'
import { Separator } from '@/components/ui/separator'

export interface Notification {
    id: string
    actor: string
    action: string
    file: string
    timestamp: Date
    read: boolean
}

interface NotificationRowProps {
    notification: Notification
    showSeparator?: boolean
}

export default function NotificationRow({ notification, showSeparator }: NotificationRowProps) {
    const { actor, action, file, timestamp, read } = notification

    return (
        <>
            <div className={`flex items-start gap-3 px-4 py-3 hover:bg-gray-50 transition-colors cursor-pointer ${!read ? 'bg-[#6B5FAE]/5' : ''}`}>
                <div className="shrink-0 mt-1.5">
                    <span className={`block w-2 h-2 rounded-full ${!read ? 'bg-[#6B5FAE]' : 'bg-transparent'}`} />
                </div>
                <div className="flex-1 min-w-0">
                    <p className="text-sm text-gray-800 leading-snug">
                        <span className="font-semibold">{actor}</span>{' '}
                        {action}{' '}
                        <span className="text-[#6B5FAE] font-medium">"{file}"</span>
                    </p>
                    <p className="text-xs text-gray-400 mt-0.5">
                        {formatDistanceToNow(timestamp, { addSuffix: true, locale: id })}
                    </p>
                </div>
            </div>
            {showSeparator && <Separator className="mx-4 w-auto opacity-50" />}
        </>
    )
}