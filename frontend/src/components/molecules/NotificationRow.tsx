export interface Notification {
    id:        string
    actor:     string
    action:    string
    file:      string
    timestamp: Date
    read:      boolean
}

interface NotificationRowProps {
    notification:  Notification
    showSeparator?: boolean
    onMarkRead?:   () => void
}

export default function NotificationRow({
    notification,
    showSeparator,
    onMarkRead,
}: NotificationRowProps) {
    return (
        <>
            <div className="flex items-start gap-3 px-4 py-3">
               

                {/* Content */}
                <div className="flex-1 min-w-0">
                    <p className="text-xs text-gray-800 leading-relaxed">
                        {notification.action}
                    </p>
                    <p className="text-[11px] text-gray-400 mt-0.5">
                        {notification.timestamp.toLocaleString('id-ID', {
                            day:    '2-digit',
                            month:  'short',
                            hour:   '2-digit',
                            minute: '2-digit',
                        })}
                    </p>

                    {/* Mark as read button — only shown on unread */}
                    {!notification.read && onMarkRead && (
                        <button
                            onClick={(e) => {
                                e.stopPropagation()
                                onMarkRead()
                            }}
                            className="mt-1.5 text-[11px] text-[#6B5FAE] bg-[#6B5FAE]/10 hover:bg-[#6B5FAE]/20 px-2 py-0.5 rounded-md transition-colors"
                        >
                            Tandai dibaca
                        </button>
                    )}
                </div>

                {/* Unread dot */}
                {!notification.read && (
                    <span className="w-2 h-2 rounded-full bg-[#6B5FAE] flex-shrink-0 mt-1.5" />
                )}
            </div>

            {showSeparator && <div className="h-px bg-gray-100 mx-4" />}
        </>
    )
}