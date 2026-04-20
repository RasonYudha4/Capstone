import { useState } from 'react'
import { Bell, ArrowRight } from 'lucide-react'
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

const mockNotifications: Notification[] = [
    { id: '1', actor: 'Supriyadi', action: 'mengupload file', file: 'EP 2 PMKP Standar 2.pdf', timestamp: new Date('2026-04-12T13:45:00'), read: false },
    { id: '2', actor: 'Pardi', action: 'mengupload file', file: 'EP 3 PAB Standar 2.pdf', timestamp: new Date('2026-04-08T14:00:00'), read: false },
    { id: '3', actor: 'Priyadi', action: 'mengupload file', file: 'EP 3 PAB Standar 4.pdf', timestamp: new Date('2026-04-08T14:00:00'), read: true },
    { id: '4', actor: 'Pardi', action: 'mengupload file', file: 'EP 2 PAB Standar 2.pdf', timestamp: new Date('2026-04-08T14:00:00'), read: true },
    { id: '5', actor: 'Supriyadi', action: 'mengupload file', file: 'EP 1 PMKP Standar 1.pdf', timestamp: new Date('2026-04-07T09:30:00'), read: true },
    { id: '6', actor: 'Wulandari', action: 'mengupload file', file: 'EP 4 MKI Standar 3.pdf', timestamp: new Date('2026-04-06T11:00:00'), read: true },
    { id: '7', actor: 'Hartono', action: 'mengupload file', file: 'EP 1 PAB Standar 1.pdf', timestamp: new Date('2026-04-05T15:20:00'), read: true },
    { id: '8', actor: 'Pardi', action: 'mengupload file', file: 'EP 5 SKP Standar 2.pdf', timestamp: new Date('2026-04-04T10:10:00'), read: true },
    { id: '9', actor: 'Priyadi', action: 'mengupload file', file: 'EP 2 HPK Standar 1.pdf', timestamp: new Date('2026-04-03T08:45:00'), read: true },
    { id: '10', actor: 'Wulandari', action: 'mengupload file', file: 'EP 3 MKI Standar 2.pdf', timestamp: new Date('2026-04-02T16:30:00'), read: true },
]

const INITIAL_VISIBLE = 5

interface NotificationDropdownProps {
    onViewAll: () => void
}

export default function NotificationDropdown({ onViewAll }: NotificationDropdownProps) {
    const [visibleCount, setVisibleCount] = useState(INITIAL_VISIBLE)
    const unreadCount = mockNotifications.filter((n) => !n.read).length
    const visible = mockNotifications.slice(0, visibleCount)
    const hasMore = visibleCount < mockNotifications.length

    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <Button variant="outline" size="icon" className="rounded-full border-gray-200 relative" aria-label="Notifikasi">
                    <Bell className="w-4 h-4 text-gray-600" />
                    {unreadCount > 0 && (
                        <span className="absolute -top-1 -right-1 w-4 h-4 rounded-full bg-[#6B5FAE] text-white text-[10px] font-bold flex items-center justify-center">
                            {unreadCount}
                        </span>
                    )}
                </Button>
            </DropdownMenuTrigger>

            <DropdownMenuContent align="end" sideOffset={8} className="w-80 p-0 rounded-2xl shadow-lg border border-gray-100 overflow-hidden">
                {/* Header */}
                <div className="flex items-center justify-between px-4 py-3">
                    <div className="flex items-center gap-2">
                        <h3 className="font-semibold text-sm text-gray-900">Notifikasi</h3>
                        {unreadCount > 0 && (
                            <Badge className="bg-[#6B5FAE]/10 text-[#6B5FAE]">{unreadCount} baru</Badge>
                        )}
                    </div>
                    <Button
                        variant="ghost"
                        size="sm"
                        onClick={onViewAll}
                        className="text-[#6B5FAE] hover:text-[#6B5FAE] hover:bg-[#6B5FAE]/10 text-xs font-semibold h-7 px-2 rounded-lg"
                    >
                        Lihat semua <ArrowRight className="w-3 h-3 ml-1" />
                    </Button>
                </div>

                <Separator />

                <ScrollArea className="h-90">
                    <div className="py-1">
                        {visible.map((n, i) => (
                            <NotificationRow
                                key={n.id}
                                notification={n}
                                showSeparator={i < visible.length - 1}
                            />
                        ))}
                    </div>

                    {hasMore && (
                        <>
                            <Separator />
                            <div className="px-4 py-2.5">
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    onClick={() => setVisibleCount(mockNotifications.length)}
                                    className="w-full text-xs text-gray-500 hover:text-[#6B5FAE] hover:bg-[#6B5FAE]/10 rounded-xl h-8"
                                >
                                    Tampilkan {mockNotifications.length - visibleCount} notifikasi lainnya
                                </Button>
                            </div>
                        </>
                    )}
                </ScrollArea>
            </DropdownMenuContent>
        </DropdownMenu>
    )
}