import { useState, useCallback } from 'react'
import { Bell, ArrowRight, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import Badge from '../atoms/Badge'
import NotificationRow, { type Notification } from '../molecules/NotificationRow'
import {
    useNotifications,
    useNotificationEvents,
    useMarkNotificationRead,
    useMarkAllNotificationsRead,
} from '@/hooks/useNotif'
import type { NotificationItem } from '@/dtos/notification_dto'

const INITIAL_VISIBLE = 5

interface NotificationDropdownProps {
    onViewAll: () => void
}

function toNotification(item: NotificationItem): Notification {
    return {
        id:        item.NotificationID,
        actor:     '',
        action:    item.Message,
        file:      '',
        timestamp: new Date(item.CreatedAt),
        read:      item.Read,
    }
}

export default function NotificationDropdown({ onViewAll }: NotificationDropdownProps) {
    const [visibleCount, setVisibleCount] = useState(INITIAL_VISIBLE)

    const { data, isLoading } = useNotifications()
    const markRead            = useMarkNotificationRead()
    const markAllRead         = useMarkAllNotificationsRead()

    useNotificationEvents(
        useCallback((event) => {
            console.log('[SSE]', event.type, event.message)
        }, [])
    )

    const notifications = data?.data ?? []
    const unreadCount   = notifications.filter(n => !n.Read).length
    const visible       = notifications.slice(0, visibleCount)
    const hasMore       = visibleCount < notifications.length

    function handleMarkRead(id: string) {
        markRead.mutate(id)
    }

    function handleMarkAllRead() {
        markAllRead.mutate()
    }

    return (
        <DropdownMenu onOpenChange={() => setVisibleCount(INITIAL_VISIBLE)}>
            <DropdownMenuTrigger asChild>
                <Button
                    variant="outline"
                    size="icon"
                    className="rounded-full border-gray-200 relative"
                    aria-label="Notifikasi"
                >
                    <Bell className="w-4 h-4 text-gray-600" />
                    {unreadCount > 0 && (
                        <span className="absolute -top-1 -right-1 w-4 h-4 rounded-full bg-[#6B5FAE] text-white text-[10px] font-bold flex items-center justify-center">
                            {unreadCount > 9 ? '9+' : unreadCount}
                        </span>
                    )}
                </Button>
            </DropdownMenuTrigger>

            <DropdownMenuContent
                align="end"
                sideOffset={8}
                className="w-80 p-0 rounded-2xl shadow-lg border border-gray-100 overflow-hidden"
            >
                {/* Header */}
                <div className="flex items-center justify-between px-4 py-3">
                    <div className="flex items-center gap-2">
                        <h3 className="font-semibold text-sm text-gray-900">Notifikasi</h3>
                        {unreadCount > 0 && (
                            <Badge className="bg-[#6B5FAE]/10 text-[#6B5FAE]">{unreadCount} baru</Badge>
                        )}
                    </div>
                    <div className="flex items-center gap-1">
                        {unreadCount > 0 && (
                            <Button
                                variant="ghost"
                                size="sm"
                                onClick={handleMarkAllRead}
                                disabled={markAllRead.isPending}
                                className="text-gray-400 hover:text-gray-600 text-xs h-7 px-2 rounded-lg"
                            >
                                {markAllRead.isPending
                                    ? <Loader2 className="w-3 h-3 animate-spin" />
                                    : 'Tandai semua'
                                }
                            </Button>
                        )}
                        <Button
                            variant="ghost"
                            size="sm"
                            onClick={onViewAll}
                            className="text-[#6B5FAE] hover:text-[#6B5FAE] hover:bg-[#6B5FAE]/10 text-xs font-semibold h-7 px-2 rounded-lg"
                        >
                            Lihat semua <ArrowRight className="w-3 h-3 ml-1" />
                        </Button>
                    </div>
                </div>

                <Separator />

                <ScrollArea className="h-[360px]">
                    {isLoading ? (
                        <div className="flex items-center justify-center py-10">
                            <Loader2 className="w-5 h-5 text-gray-300 animate-spin" />
                        </div>
                    ) : notifications.length === 0 ? (
                        <div className="flex flex-col items-center justify-center py-10 gap-2">
                            <Bell className="w-7 h-7 text-gray-200" />
                            <p className="text-xs text-gray-400">Tidak ada notifikasi</p>
                        </div>
                    ) : (
                        <div className="py-1">
                            {visible.map((item, i) => (
                                <div
                                    key={item.NotificationID}
                                    onClick={() => !item.Read && handleMarkRead(item.NotificationID)}
                                    className={`cursor-pointer transition-colors ${
                                        !item.Read
                                            ? 'bg-[#6B5FAE]/[0.03] hover:bg-[#6B5FAE]/[0.07]'
                                            : 'hover:bg-gray-50'
                                    }`}
                                >
                                    <NotificationRow
                                        notification={toNotification(item)}
                                        showSeparator={i < visible.length - 1}
                                        onMarkRead={() => handleMarkRead(item.NotificationID)}
                                    />
                                </div>
                            ))}
                        </div>
                    )}

                    {hasMore && (
                        <>
                            <Separator />
                            <div className="px-4 py-2.5">
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    onClick={() => setVisibleCount(notifications.length)}
                                    className="w-full text-xs text-gray-500 hover:text-[#6B5FAE] hover:bg-[#6B5FAE]/10 rounded-xl h-8"
                                >
                                    Tampilkan {notifications.length - visibleCount} notifikasi lainnya
                                </Button>
                            </div>
                        </>
                    )}
                </ScrollArea>
            </DropdownMenuContent>
        </DropdownMenu>
    )
}