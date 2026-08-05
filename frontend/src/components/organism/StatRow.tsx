import { useState } from 'react'
import { InboxIcon, DownloadIcon, FileSearch } from 'lucide-react'
import { toast } from 'sonner'
import { useQueryClient } from '@tanstack/react-query'
import StatCard from '../molecules/StatCard'
import FileDetailModal from './FiledetailModal'
import { useStats, documentKeys } from '@/hooks/useDocument'
import { useAudit } from '@/hooks/useAudit'
import { useMe } from '@/hooks/useAuth'
import { useNotifications, useMarkNotificationRead } from '@/hooks/useNotif'
import { documentService } from '@/services/document_services'
import type { NotificationItem } from '@/dtos/notification_dto'
import type { DocumentResponse, DocumentStatus } from '@/dtos/document_dto'
import type { FileRecord } from '../molecules/FileTableRow'

interface LastOpened {
    documentId: string
    name: string
    date: string
}

function formatDate(iso: string): string {
    return new Date(iso).toLocaleDateString('id-ID', {
        day: 'numeric',
        month: 'long',
        year: 'numeric',
    })
}

function getWeeklyCount(current: number, key: string): number {
    const stored = localStorage.getItem(key)
    const now = Date.now()
    if (stored) {
        const { value, timestamp } = JSON.parse(stored) as { value: number; timestamp: number }
        if (now - timestamp < 7 * 24 * 60 * 60 * 1000) {
            return Math.max(0, current - value)
        }
    }
    localStorage.setItem(key, JSON.stringify({ value: current, timestamp: now }))
    return 0
}

function toDocumentResponse(item: NotificationItem): DocumentResponse | null {
    if (!item.DocumentID) return null

    const status = (['pending', 'approved', 'rejected'] as const).includes(
        item.DocumentStatus as DocumentStatus,
    )
        ? (item.DocumentStatus as DocumentStatus)
        : 'pending'

    return {
        document_id: item.DocumentID,
        filename: item.Filename || 'Dokumen',
        document_type: item.DocumentType || '',
        created_by: item.CreatedBy || '',
        updated_at: item.DocumentUpdatedAt || item.UpdatedAt,
        status,
        service_code: item.ServiceCode || '',
        standard_code: item.StandardCode || '',
        assessment_code: item.AssessmentCode || '',
    }
}

function toFileRecord(doc: DocumentResponse): FileRecord {
    return {
        id: doc.document_id,
        name: doc.filename,
        type: doc.document_type ?? '',
        uploadedBy: doc.created_by,
        lastUpdated: doc.updated_at,
        status: doc.status,
    }
}

export default function StatRow() {
    const { data: stats } = useStats()
    const { data: audit } = useAudit()
    const { data: me } = useMe()
    const { data: notifData } = useNotifications()
    const markRead = useMarkNotificationRead()
    const queryClient = useQueryClient()

    const [selected, setSelected] = useState<DocumentResponse | null>(null)
    const [fileUrl, setFileUrl] = useState<string | undefined>()
    const [contentType, setContentType] = useState('')
    const [modalOpen, setModalOpen] = useState(false)
    const [modalLoading, setModalLoading] = useState(false)

    const lastOpened: LastOpened | null = (() => {
        if (!audit?.data || !me?.email) return null

        const lastOpen = audit.data
            .filter(a =>
                a.action.toLowerCase() === 'open' &&
                a.username === me.email &&
                a.document_name != null &&
                !!a.document_id
            )
            .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
            .at(0)

        if (!lastOpen?.document_id) return null
        return {
            documentId: lastOpen.document_id,
            name: lastOpen.document_name!,
            date: lastOpen.created_at,
        }
    })()

    const notifications = notifData?.data ?? []
    const unreadNotifications = notifications.filter(n => !n.Read && n.DocumentID)
    const hasUnread = unreadNotifications.length > 0

    const pendingCount = stats?.stats.pending ?? 0
    const totalCount = stats?.total ?? 0
    const weeklyPending = stats ? getWeeklyCount(pendingCount, 'snapshot_pending') : 0
    const weeklyTotal = stats ? getWeeklyCount(totalCount, 'snapshot_total') : 0

    async function openDocument(document: DocumentResponse) {
        setSelected(document)
        setModalOpen(true)
        setModalLoading(true)

        try {
            const { url, contentType: ct } = await documentService.getById(document.document_id)
            setFileUrl(prev => {
                if (prev) URL.revokeObjectURL(prev)
                return url
            })
            setContentType(ct)
        } catch {
            setFileUrl(prev => {
                if (prev) URL.revokeObjectURL(prev)
                return undefined
            })
            setContentType('')
            toast.error('Gagal memuat berkas dokumen.')
        } finally {
            setModalLoading(false)
        }
    }

    async function openUnreadDocument() {
        const latest = unreadNotifications
            .slice()
            .sort((a, b) => new Date(b.CreatedAt).getTime() - new Date(a.CreatedAt).getTime())[0]

        if (!latest) {
            toast.error('Tidak ada notifikasi dokumen yang belum dibaca.')
            return
        }

        const document = toDocumentResponse(latest)
        if (!document) {
            toast.error('Dokumen untuk notifikasi ini tidak tersedia.')
            return
        }

        if (!latest.Read) markRead.mutate(latest.NotificationID)
        await openDocument(document)
    }

    async function openLastViewedDocument() {
        if (!lastOpened) {
            toast.error('Belum ada dokumen yang dilihat.')
            return
        }

        await openDocument({
            document_id: lastOpened.documentId,
            filename: lastOpened.name,
            document_type: '',
            created_by: me?.email ?? '',
            updated_at: lastOpened.date,
            status: 'pending',
            service_code: '',
            standard_code: '',
            assessment_code: '',
        })
    }

    function handleModalClose(open: boolean) {
        if (!open) {
            setModalOpen(false)
            setSelected(null)
            setFileUrl(prev => {
                if (prev) URL.revokeObjectURL(prev)
                return undefined
            })
            setContentType('')
        }
    }

    async function afterAction() {
        setModalOpen(false)
        setSelected(null)
        setFileUrl(prev => {
            if (prev) URL.revokeObjectURL(prev)
            return undefined
        })
        setContentType('')
        queryClient.invalidateQueries({ queryKey: documentKeys.all })
        queryClient.invalidateQueries({ queryKey: documentKeys.stats })
    }

    return (
        <>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
                <StatCard
                    label="Pending"
                    value={pendingCount}
                    weeklyCount={weeklyPending}
                    weeklyLabel="Minggu ini"
                    icon={InboxIcon}
                />
                <StatCard
                    label="Dokumen Masuk"
                    value={totalCount}
                    weeklyCount={weeklyTotal}
                    weeklyLabel="Minggu ini"
                    icon={DownloadIcon}
                    showDot={hasUnread}
                    onDotClick={() => void openUnreadDocument()}
                />
                <StatCard
                    label="Terakhir Dilihat"
                    icon={FileSearch}
                    clickable={!!lastOpened}
                    onClick={() => void openLastViewedDocument()}
                >
                    {lastOpened ? (
                        <div className="flex flex-col">
                            <span className="text-sm font-medium text-gray-800 truncate max-w-40">
                                {lastOpened.name}
                            </span>
                            <span className="text-sm text-gray-500">
                                {formatDate(lastOpened.date)}
                            </span>
                        </div>
                    ) : (
                        <span className="text-sm text-gray-400">Belum ada</span>
                    )}
                </StatCard>
            </div>

            {selected && (
                <FileDetailModal
                    open={modalOpen}
                    onOpenChange={handleModalClose}
                    file={toFileRecord(selected)}
                    document={selected}
                    isLoading={modalLoading}
                    fileUrl={fileUrl}
                    contentType={contentType}
                    role={me?.role}
                    onApprove={async (file, _catatan, attachment) => {
                        try {
                            await documentService.approve(
                                { document_id: file.id, status: 'approved' },
                                attachment,
                            )
                            toast.success('Berkas berhasil disetujui.')
                            await afterAction()
                        } catch {
                            toast.error('Gagal menyetujui berkas.')
                        }
                    }}
                    onReject={async (file, _catatan, attachment) => {
                        try {
                            await documentService.approve(
                                { document_id: file.id, status: 'rejected' },
                                attachment,
                            )
                            toast.success('Berkas berhasil ditolak.')
                            await afterAction()
                        } catch {
                            toast.error('Gagal menolak berkas.')
                        }
                    }}
                    onUpdate={async (file, _catatan, attachment, filename) => {
                        try {
                            await documentService.update(
                                { document_id: file.id, filename },
                                attachment,
                            )
                            toast.success('Berkas berhasil diperbarui.')
                            await afterAction()
                        } catch {
                            toast.error('Gagal memperbarui berkas.')
                        }
                    }}
                    onDelete={async (file) => {
                        try {
                            await documentService.delete(file.id)
                            toast.success('Berkas berhasil dihapus.')
                            await afterAction()
                        } catch {
                            toast.error('Gagal menghapus berkas.')
                        }
                    }}
                />
            )}
        </>
    )
}
