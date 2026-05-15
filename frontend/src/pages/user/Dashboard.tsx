import { useState } from 'react'
import {
    Table,
    TableBody,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table'
import FileTableRow, { type FileRecord } from '@/components/molecules/FileTableRow'
import FileStatsBar from '@/components/molecules/FileStatsBar'
import { useDocumentsByType } from '@/hooks/useDocument'

// The 3 allowed document type IDs from your DB
// Replace these with the actual UUIDs from your document_types table
const ALLOWED_TYPES = [
    { label: 'Pedoman',                     id: 'UUID_PEDOMAN_HERE' },
    { label: 'Panduan',                     id: 'UUID_PANDUAN_HERE' },
    { label: 'Standar Prosedur Operasional', id: 'UUID_SPO_HERE' },
]

function DocumentTypeTab({
    typeId,
    label,
}: {
    typeId: string
    label: string
}) {
    const { data, isLoading } = useDocumentsByType(typeId)

    const files: FileRecord[] = (data?.data ?? []).map(d => ({
        id:          d.document_id,
        name:        d.filename,
        type:        d.document_type,
        uploadedBy:  d.created_by,
        lastUpdated: new Date(d.updated_at).toLocaleDateString('id-ID', {
            day: '2-digit', month: 'short', year: 'numeric',
        }),
        status: d.status,
    }))

    const total    = files.length
    const approved = files.filter(f => f.status === 'approved').length
    const pending  = files.filter(f => f.status === 'pending').length
    const rejected = files.filter(f => f.status === 'rejected').length

    return (
        <div className="flex flex-col gap-6">
            {/* Stats */}
            <section className="bg-white rounded-2xl border border-gray-100 p-6">
                <p className="text-sm font-medium text-gray-500 mb-1">Statistik Berkas — {label}</p>
                <p className="text-3xl font-bold text-[#6B5FAE] mb-5">
                    {total} Total Berkas
                </p>
                <FileStatsBar
                    total={total}
                    approved={approved}
                    pending={pending}
                    rejected={rejected}
                />
            </section>

            {/* Table */}
            <section className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
                <div className="px-6 py-4 border-b border-gray-100">
                    <p className="text-base font-semibold text-[#6B5FAE]">Daftar Berkas {label}</p>
                </div>
                <div className="px-6 py-4">
                    <div className="rounded-2xl overflow-hidden border border-gray-100">
                        <div className="overflow-x-auto">
                            <Table className="min-w-175">
                                <TableHeader>
                                    <TableRow className="bg-[#6B5FAE] hover:bg-[#6B5FAE]">
                                        {['Nama Berkas', 'Tipe Berkas', 'Di Upload oleh', 'Terakhir diperbaharui', 'Status'].map((h) => (
                                            <TableHead key={h} className="text-white text-center font-semibold text-sm px-4 py-3">
                                                {h}
                                            </TableHead>
                                        ))}
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {isLoading ? (
                                        <tr>
                                            <td colSpan={5} className="text-center text-sm text-gray-400 py-10">
                                                Memuat berkas...
                                            </td>
                                        </tr>
                                    ) : files.length === 0 ? (
                                        <tr>
                                            <td colSpan={5} className="text-center text-sm text-gray-400 py-10">
                                                Tidak ada berkas tersedia.
                                            </td>
                                        </tr>
                                    ) : (
                                        files.map(file => (
                                            <FileTableRow key={file.id} file={file} />
                                        ))
                                    )}
                                </TableBody>
                            </Table>
                        </div>
                    </div>
                </div>
            </section>
        </div>
    )
}

export default function UserDashboard() {
    const [activeType, setActiveType] = useState(ALLOWED_TYPES[0].id)
    const activeLabel = ALLOWED_TYPES.find(t => t.id === activeType)?.label ?? ''

    return (
        <div className="grid grid-cols-1 gap-6 max-w-7xl mx-auto">

            {/* Header */}
            <div>
                <h1 className="text-2xl font-bold text-[#6B5FAE]">Dokumen Publik</h1>
                <p className="text-sm text-gray-500 mt-1">
                    Dokumen yang tersedia untuk umum — Pedoman, Panduan, dan SPO.
                </p>
            </div>

            {/* Type Tabs */}
            <div className="flex gap-2">
                {ALLOWED_TYPES.map(t => (
                    <button
                        key={t.id}
                        onClick={() => setActiveType(t.id)}
                        className={`px-5 py-2 rounded-xl text-sm font-semibold transition-colors ${
                            activeType === t.id
                                ? 'bg-[#6B5FAE] text-white'
                                : 'bg-white border border-gray-200 text-gray-500 hover:border-[#6B5FAE] hover:text-[#6B5FAE]'
                        }`}
                    >
                        {t.label}
                    </button>
                ))}
            </div>

            {/* Content */}
            <DocumentTypeTab typeId={activeType} label={activeLabel} />
        </div>
    )
}