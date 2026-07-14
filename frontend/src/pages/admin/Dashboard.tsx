import { useEffect, useState } from "react"
import { toast } from 'sonner'
import { useQueryClient } from '@tanstack/react-query'
import type { DocumentStatus, DocumentResponse } from "@/dtos/document_dto"
import { useMyDocuments, documentKeys } from "@/hooks/useDocument"
import { documentService } from '@/services/document_services'
import FileDetailModal from '@/components/organism/FiledetailModal'
import { useMe } from '@/hooks/useAuth'
import type { FileRecord } from '@/components/molecules/FileTableRow'

const tabs: { key: DocumentStatus; label: string }[] = [
    { key: "pending", label: "Pending" },
    { key: "rejected", label: "Ditolak" },
    { key: "approved", label: "Disetujui" },
]

const statusStyles: Record<DocumentStatus, string> = {
    pending: "bg-amber-50 text-amber-800",
    rejected: "bg-red-50 text-red-700",
    approved: "bg-green-50 text-green-800",
}

const statusLabel: Record<DocumentStatus, string> = {
    pending: "Pending",
    rejected: "Ditolak",
    approved: "Disetujui",
}

function formatDate(iso: string): string {
    const date = new Date(iso)
    if (isNaN(date.getTime())) return "-"
    return date.toLocaleDateString("id-ID", {
        day: "2-digit",
        month: "short",
        year: "numeric",
    })
}

const LIMIT = 10

export default function DashboardPage() {
    const queryClient = useQueryClient()
    const { data: me } = useMe()

    const [activeTab, setActiveTab] = useState<DocumentStatus>("pending")
    const [page, setPage] = useState(1)

    const { data, isLoading, isError } = useMyDocuments({ page, limit: LIMIT })

    useEffect(() => console.log("Data ", data), []);

    const [selectedDocId, setSelectedDocId] = useState<string | null>(null)
    const [selectedDocument, setSelectedDocument] = useState<DocumentResponse | null>(null)
    const [fileUrl, setFileUrl] = useState('')
    const [contentType, setContentType] = useState('')
    const [detailLoading, setDetailLoading] = useState(false)

    useEffect(() => {
        if (!selectedDocId) {
            setFileUrl('')
            setContentType('')
            return
        }

        let cancelled = false
        setDetailLoading(true)

        documentService.getById(selectedDocId)
            .then(({ url, contentType }) => {
                if (cancelled) return
                setFileUrl(url)
                setContentType(contentType)
            })
            .catch(() => {
                if (!cancelled) toast.error('Gagal memuat berkas.')
            })
            .finally(() => {
                if (!cancelled) setDetailLoading(false)
            })

        return () => {
            cancelled = true
        }
    }, [selectedDocId])

    const handleRowClick = (doc: DocumentResponse) => {
        setSelectedDocument(doc)
        setSelectedDocId(doc.document_id)
    }

    const selectedFile: FileRecord | null = selectedDocument
        ? {
            id: selectedDocument.document_id,
            name: selectedDocument.filename,
            type: selectedDocument.document_type,
            uploadedBy: selectedDocument.created_by,
            lastUpdated: formatDate(selectedDocument.updated_at),
            status: selectedDocument.status,
        }
        : null

    const closeModal = () => {
        setSelectedDocId(null)
        setSelectedDocument(null)
        setFileUrl('')
        setContentType('')
    }

    const invalidateList = () => {
        queryClient.invalidateQueries({ queryKey: documentKeys.mine({ page, limit: LIMIT }) })
        queryClient.invalidateQueries({ queryKey: documentKeys.all })
    }

    const handleApprove = async (file: FileRecord, _catatan: string, attachment?: File) => {
        try {
            await documentService.approve({ document_id: file.id, status: 'approved' }, attachment)
            toast.success('Berkas berhasil disetujui.')
            invalidateList()
            closeModal()
        } catch {
            toast.error('Gagal menyetujui berkas.')
        }
    }

    const handleReject = async (file: FileRecord, _catatan: string, attachment?: File) => {
        try {
            await documentService.approve({ document_id: file.id, status: 'rejected' }, attachment)
            toast.success('Berkas berhasil ditolak.')
            invalidateList()
            closeModal()
        } catch {
            toast.error('Gagal menolak berkas.')
        }
    }

    const handleUpdate = async (file: FileRecord, _catatan: string, attachment?: File, filename?: string) => {
        if (!selectedDocument) return
        try {
            await documentService.update({ document_id: file.id, filename }, attachment)
            toast.success('Berkas berhasil diperbarui.')
            invalidateList()
            closeModal()
        } catch (err) {
            console.error('Update error:', err)
            toast.error('Gagal memperbarui berkas.')
        }
    }

    const handleDelete = async (file: FileRecord) => {
        try {
            await documentService.delete(file.id)
            toast.success('Berkas berhasil dihapus.')
            invalidateList()
            closeModal()
            if (filtered.length === 1 && page > 1) setPage(p => p - 1)
        } catch {
            toast.error('Gagal menghapus berkas.')
        }
    }

    const allDocs = data?.data ?? []
    const filtered = allDocs.filter(d => d.status === activeTab)

    const counts = (["pending", "rejected", "approved"] as DocumentStatus[]).reduce(
        (acc, s) => ({ ...acc, [s]: allDocs.filter(d => d.status === s).length }),
        {} as Record<DocumentStatus, number>
    )

    function handleTabChange(tab: DocumentStatus) {
        setActiveTab(tab)
        setPage(1)
    }

    return (
        <div className="bg-white shadow-xl rounded-2xl border border-gray-100 p-5">
            <FileDetailModal
                file={selectedFile}
                document={selectedDocument}
                isLoading={detailLoading}
                fileUrl={fileUrl}
                contentType={contentType}
                open={!!selectedDocId}
                onOpenChange={(open) => { if (!open) closeModal() }}
                onApprove={handleApprove}
                onReject={handleReject}
                onUpdate={handleUpdate}
                onDelete={handleDelete}
                role={me?.role}
            />

            {/* Header */}
            <div className="flex items-center justify-between mb-4">
                <span className="text-sm font-medium text-gray-800">Dokumen Saya</span>
            </div>

            {/* Tabs */}
            <div className="flex gap-1 mb-4 flex-wrap">
                {tabs.map(t => (
                    <button
                        key={t.key}
                        onClick={() => handleTabChange(t.key)}
                        className={`flex items-center gap-1.5 px-3.5 py-1.5 rounded-full text-sm transition-colors ${activeTab === t.key
                            ? "border border-violet-500 text-violet-700 font-medium"
                            : "text-gray-500 hover:text-gray-700"
                            }`}
                    >
                        {t.label}
                        <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${activeTab === t.key
                            ? "bg-violet-100 text-violet-600"
                            : "bg-gray-100 text-gray-500"
                            }`}>
                            {counts[t.key] ?? 0}
                        </span>
                    </button>
                ))}
            </div>

            {/* Table */}
            <div className="overflow-x-auto">
                <table className="w-full text-sm table-fixed">
                    <colgroup>
                        <col className="w-[22%]" />
                        <col className="w-[14%]" />
                        <col className="w-[14%]" />
                        <col className="w-[14%]" />
                        <col className="w-[14%]" />
                        <col className="w-[14%]" />
                        <col className="w-[8%]" />
                    </colgroup>
                    <thead>
                        <tr className="border-b border-gray-100">
                            <th className="text-left text-xs font-normal text-gray-400 pb-2 pr-4">Nama berkas</th>
                            <th className="text-left text-xs font-normal text-gray-400 pb-2 pr-4">Tipe dokumen</th>
                            <th className="text-left text-xs font-normal text-gray-400 pb-2 pr-4">Layanan</th>
                            <th className="text-left text-xs font-normal text-gray-400 pb-2 pr-4">Standar</th>
                            <th className="text-left text-xs font-normal text-gray-400 pb-2 pr-4">Elemen Penilaian</th>
                            <th className="text-left text-xs font-normal text-gray-400 pb-2 pr-4">Terakhir diperbaharui</th>
                            <th className="text-left text-xs font-normal text-gray-400 pb-2">Status</th>
                        </tr>
                    </thead>
                    <tbody>
                        {isLoading ? (
                            <tr>
                                <td colSpan={7} className="text-center text-gray-400 py-10 text-sm">
                                    Memuat dokumen...
                                </td>
                            </tr>
                        ) : isError ? (
                            <tr>
                                <td colSpan={7} className="text-center text-red-400 py-10 text-sm">
                                    Gagal memuat dokumen. Coba lagi.
                                </td>
                            </tr>
                        ) : filtered.length === 0 ? (
                            <tr>
                                <td colSpan={7} className="text-center text-gray-400 py-10 text-sm">
                                    Tidak ada dokumen.
                                </td>
                            </tr>
                        ) : (
                            filtered.map(doc => (
                                <tr
                                    key={doc.document_id}
                                    onClick={() => handleRowClick(doc)}
                                    className="border-b border-gray-50 hover:bg-gray-50 transition-colors cursor-pointer"
                                >
                                    <td className="py-3 pr-4 max-w-0">
                                        <span className="truncate block text-gray-800" title={doc.filename}>
                                            {doc.filename}
                                        </span>
                                    </td>
                                    <td className="py-3 pr-4 max-w-0">
                                        <span className="truncate block text-gray-600" title={doc.document_type}>
                                            {doc.document_type}
                                        </span>
                                    </td>
                                    <td className="py-3 pr-4 max-w-0">
                                        <span className="truncate block text-gray-600" title={doc.service_code}>
                                            {doc.service_code ?? "-"}
                                        </span>
                                    </td>
                                    <td className="py-3 pr-4 max-w-0">
                                        <span className="truncate block text-gray-600" title={doc.standard_code}>
                                            {doc.standard_code ?? "-"}
                                        </span>
                                    </td>
                                    <td className="py-3 pr-4 max-w-0">
                                        <span className="truncate block text-gray-600" title={doc.assessment_code}>
                                            {doc.assessment_code ?? "-"}
                                        </span>
                                    </td>
                                    <td className="py-3 pr-4 text-gray-500 whitespace-nowrap">
                                        {formatDate(doc.updated_at)}
                                    </td>
                                    <td className="py-3">
                                        <span className={`inline-flex items-center text-xs px-2.5 py-1 rounded-full font-medium ${statusStyles[doc.status as DocumentStatus]}`}>
                                            {statusLabel[doc.status as DocumentStatus] ?? doc.status}
                                        </span>
                                    </td>
                                </tr>
                            ))
                        )}
                    </tbody>
                </table>
            </div>

            {/* Pagination */}
            {!isLoading && !isError && (
                <div className="flex items-center justify-between mt-4 pt-3 border-t border-gray-100">
                    <span className="text-xs text-gray-400">
                        Halaman {page}
                    </span>
                    <div className="flex gap-2">
                        <button
                            onClick={() => setPage(p => Math.max(1, p - 1))}
                            disabled={page === 1}
                            className="text-xs px-3 py-1.5 rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                        >
                            Sebelumnya
                        </button>
                        <button
                            onClick={() => setPage(p => p + 1)}
                            disabled={allDocs.length < LIMIT}
                            className="text-xs px-3 py-1.5 rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                        >
                            Selanjutnya
                        </button>
                    </div>
                </div>
            )}
        </div>
    )
}