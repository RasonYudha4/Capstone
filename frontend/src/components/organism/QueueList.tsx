import { useEffect, useState } from 'react'
import { FileText, User, Tag, Clock, Fingerprint, Eye } from 'lucide-react'
import { documentService } from '@/services/document_services'
import type { DocumentResponse } from '@/dtos/document_dto'
import FileDetailModal from '@/components/organism/FiledetailModal'
import type { FileRecord } from '@/components/molecules/FileTableRow'

// ── helpers ────────────────────────────────────────────────────────────────

function StatusBadge({ status }: { status: string }) {
    const map: Record<string, string> = {
        approved: 'bg-green-100 text-green-800',
        rejected: 'bg-red-100 text-red-800',
        pending:  'bg-amber-100 text-amber-800',
    }
    return (
        <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${map[status] ?? map.pending}`}>
            {status || 'pending'}
        </span>
    )
}

function formatDate(iso: string) {
    if (!iso || iso.startsWith('0001')) return '—'
    return new Date(iso).toLocaleDateString('id-ID', {
        day: 'numeric',
        month: 'short',
        year: 'numeric',
    })
}

function shortId(id: string) {
    return id ? id.slice(0, 8) + '…' : '—'
}

function toFileRecord(doc: DocumentResponse): FileRecord {
    return {
        id:         doc.document_id,
        name:       doc.filename,
        type:       doc.document_type ?? '',
        size:       '',
        uploadedBy: doc.created_by,
        date:       doc.updated_at,
        status:     (doc.status as FileRecord['status']) ?? 'pending',
    }
}

// ── DocCard ────────────────────────────────────────────────────────────────

interface DocCardProps {
    doc: DocumentResponse
    onView: (doc: DocumentResponse) => void
}

function DocCard({ doc, onView }: DocCardProps) {
    return (
        <div className="bg-white border border-gray-100 rounded-2xl p-4 flex flex-col gap-3 hover:border-[#6B5FAE]/30 hover:shadow-sm transition-all">
            <div className="flex items-start justify-between gap-3">
                <div className="flex items-center gap-3 min-w-0">
                    <div className="w-9 h-9 rounded-xl bg-[#6B5FAE]/10 flex items-center justify-center shrink-0">
                        <FileText className="w-5 h-5 text-[#6B5FAE]" />
                    </div>
                    <div className="min-w-0">
                        <p className="text-sm font-medium text-gray-900 truncate">{doc.filename}</p>
                    </div>
                </div>
                <StatusBadge status={doc.status} />
            </div>

            <div className="flex flex-wrap items-center gap-x-4 gap-y-1">
                <span className="flex items-center gap-1 text-xs text-gray-400">
                    <User className="w-3 h-3" />
                    {doc.created_by || '—'}
                </span>
                <span className="flex items-center gap-1 text-xs text-gray-400">
                    <Tag className="w-3 h-3" />
                    {doc.document_type || 'No type'}
                </span>
                <span className="flex items-center gap-1 text-xs text-gray-400">
                    <Clock className="w-3 h-3" />
                    {formatDate(doc.updated_at)}
                </span>
                <span className="flex items-center gap-1 text-xs text-gray-400" title={doc.assessment}>
                    <Fingerprint className="w-3 h-3" />
                    {shortId(doc.assessment ?? '')}
                </span>
            </div>

            <div className="flex items-center border-t border-gray-100 pt-3">
                <button
                    onClick={() => onView(doc)}
                    className="flex items-center gap-1.5 text-xs font-medium px-3 py-1.5 rounded-sm bg-[#6B5FAE]/10 border border-[#6B5FAE]/20 text-[#6B5FAE] hover:bg-[#6B5FAE]/20 transition-colors cursor-pointer active:bg-purple-100"
                >
                    <Eye className="w-3.5 h-3.5" />
                    Review
                </button>
            </div>
        </div>
    )
}

// ── QueueList ──────────────────────────────────────────────────────────────

export default function QueueList() {
    const [docs, setDocs]             = useState<DocumentResponse[]>([])
    const [loading, setLoading]       = useState(true)

    const [selected, setSelected]         = useState<DocumentResponse | null>(null)
    const [fileUrl, setFileUrl]           = useState<string | undefined>()
    const [modalOpen, setModalOpen]       = useState(false)
    const [modalLoading, setModalLoading] = useState(false)

    // ── fetch / refetch ──
    async function fetchQueue() {
        setLoading(true)
        try {
            const res = await documentService.getByStatus('pending')
            setDocs(res.data ?? [])
        } catch {
            setDocs([])
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => { fetchQueue() }, [])

    // ── cleanup blob URL on unmount ──
    useEffect(() => {
        return () => {
            if (fileUrl) URL.revokeObjectURL(fileUrl)
        }
    }, [fileUrl])

    // ── open modal + fetch blob ──
    async function handleView(doc: DocumentResponse) {
        setSelected(doc)
        setModalOpen(true)
        setModalLoading(true)
        try {
            const { url } = await documentService.getById(doc.document_id)
            setFileUrl(url)
        } catch {
            setFileUrl(undefined)
        } finally {
            setModalLoading(false)
        }
    }

    // ── close modal ──
    function handleModalClose(open: boolean) {
        if (!open) {
            if (fileUrl) URL.revokeObjectURL(fileUrl)
            setModalOpen(false)
            setSelected(null)
            setFileUrl(undefined)
        }
    }

    // ── after approve / reject / update / delete ──
    async function afterAction() {
        if (fileUrl) URL.revokeObjectURL(fileUrl)
        setModalOpen(false)
        setSelected(null)
        setFileUrl(undefined)
        await fetchQueue()
    }

    if (loading) {
        return (
            <div className="bg-white shadow-xl rounded-2xl border border-gray-100 p-5 h-full flex items-center justify-center">
                <span className="text-sm text-gray-400">Memuat antrian…</span>
            </div>
        )
    }

    return (
        <>
            <div className="bg-white shadow-xl rounded-2xl border border-gray-100 p-5 h-full">
                <div className="flex items-center justify-between mb-4">
                    <h2 className="text-sm font-semibold text-gray-700">Daftar Antrian</h2>
                    <span className="text-xs bg-[#6B5FAE]/10 text-[#6B5FAE] font-medium px-2.5 py-0.5 rounded-full">
                        {docs.length} dokumen
                    </span>
                </div>

                {docs.length === 0 ? (
                    <div className="flex flex-col items-center justify-center py-10 text-gray-400 gap-2">
                        <FileText className="w-8 h-8" />
                        <span className="text-sm">Tidak Ada Dokumen Yang Membutuhkan Persetujuan</span>
                    </div>
                ) : (
                    <div className="flex flex-col gap-3">
                        {docs.map(doc => (
                            <DocCard key={doc.document_id} doc={doc} onView={handleView} />
                        ))}
                    </div>
                )}
            </div>

            {selected && (
                <FileDetailModal
                    open={modalOpen}
                    onOpenChange={handleModalClose}
                    file={toFileRecord(selected)}
                    document={selected}
                    isLoading={modalLoading}
                    fileUrl={fileUrl}
                    role="master-admin"
                    onApprove={async (file, _catatan, attachment) => {
                        await documentService.approve(
                            { document_id: file.id, status: 'approved' },
                            attachment
                        )
                        await afterAction()
                    }}
                    onReject={async (file, _catatan, attachment) => {
                        await documentService.approve(
                            { document_id: file.id, status: 'rejected' },
                            attachment
                        )
                        await afterAction()
                    }}
                    onUpdate={async (file, catatan, attachment, filename) => {
                        await documentService.update(
                            { document_id: file.id, filename, description: catatan },
                            attachment
                        )
                        await afterAction()
                    }}
                    onDelete={async (file) => {
                        await documentService.delete(file.id)
                        await afterAction()
                    }}
                />
            )}
        </>
    )
}