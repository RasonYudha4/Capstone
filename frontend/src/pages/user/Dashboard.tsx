import { useState } from 'react'
import { useNavigate } from 'react-router'
import { LogIn, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
    Table,
    TableBody,
    TableHead,
    TableHeader,
    TableRow,
    TableCell,
} from '@/components/ui/table'
import PublicDocumentModal from '@/components/organism/PublicDocumentModal'
import ChatWidget  from '@/components/organism/ChatWidget'
import { usePublicDocuments, usePublicDocumentUrl } from '@/hooks/useDocument'
import type { DocumentResponse } from '@/dtos/document_dto'

// ── Constants ─────────────────────────────────────────────────────────────────

const DOCUMENT_TYPES = [
    'SPO',
    'Surat Keputusan',
    'Materi',
    'Panduan',
    'Pedoman',
] as const

// ── Document type tab ─────────────────────────────────────────────────────────

function DocumentTypeTab({
    activeType,
    allDocuments,
    onRowClick,
}: {
    activeType: string
    allDocuments: DocumentResponse[]
    onRowClick: (doc: DocumentResponse) => void
}) {
    const files = allDocuments.filter(d => d.document_type === activeType)


    return (
        <div className="flex flex-col gap-6">

            {/* Table */}
            <section className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
                <div className="px-6 py-4 border-b border-gray-100">
                    <p className="text-base font-semibold text-[#6B5FAE]">Daftar Berkas {activeType}</p>
                </div>
                <div className="px-6 py-4">
                    <div className="rounded-2xl overflow-hidden border border-gray-100">
                        <div className="overflow-x-auto">
                            <Table>
                                <TableHeader>
                                    <TableRow className="bg-[#6B5FAE] hover:bg-[#6B5FAE]">
                                        {['Nama Berkas', 'Tipe Berkas', 'Di Upload Oleh', 'Terakhir Diperbarui'].map(h => (
                                            <TableHead key={h} className="text-white text-center font-semibold text-sm px-4 py-3">
                                                {h}
                                            </TableHead>
                                        ))}
                                    </TableRow>
                                </TableHeader>
                                <TableBody>
                                    {files.length === 0 ? (
                                        <tr>
                                            <td colSpan={5} className="text-center text-sm text-gray-400 py-10">
                                                Tidak ada berkas tersedia.
                                            </td>
                                        </tr>
                                    ) : (
                                        files.map(doc => (
                                            <TableRow
                                                key={doc.document_id}
                                                className="cursor-pointer hover:bg-gray-50 transition-colors"
                                                onClick={() => onRowClick(doc)}
                                            >
                                                <TableCell className="text-center text-sm font-medium text-gray-800 px-4 py-3">
                                                    {doc.filename.replace(/\.[^.]+$/, '')}
                                                </TableCell>
                                                <TableCell className="text-center text-sm text-gray-600 px-4 py-3">
                                                    {doc.document_type}
                                                </TableCell>
                                                <TableCell className="text-center text-sm text-gray-600 px-4 py-3">
                                                    {doc.created_by}
                                                </TableCell>
                                                <TableCell className="text-center text-sm text-gray-600 px-4 py-3">
                                                    {new Date(doc.updated_at).toLocaleDateString('id-ID', {
                                                        day: '2-digit', month: 'short', year: 'numeric',
                                                    })}
                                                </TableCell>
                                            </TableRow>
                                        ))
                                    )}
                                </TableBody>
                            </Table>
                        </div>
                    </div>
                </div>
                <div>
                    <ChatWidget></ChatWidget>
                </div>
            </section>
        </div>
    )
}

// ── Main page ─────────────────────────────────────────────────────────────────

export default function UserDashboard() {
    const navigate = useNavigate()
    const [activeType, setActiveType] = useState<string>(DOCUMENT_TYPES[0])
    const [selectedDoc, setSelectedDoc] = useState<DocumentResponse | null>(null)
    const [modalOpen, setModalOpen] = useState(false)

    const { data, isLoading } = usePublicDocuments()
    const allDocuments: DocumentResponse[] = data?.data ?? []

    // fetch file URL only when modal is open and a doc is selected
    const { data: urlData, isLoading: urlLoading } = usePublicDocumentUrl(
        modalOpen && selectedDoc ? selectedDoc.document_id : ''
    )

    const handleRowClick = (doc: DocumentResponse) => {
        setSelectedDoc(doc)
        setModalOpen(true)
    }

    const handleClose = () => {
        setModalOpen(false)
        setSelectedDoc(null)
    }

    return (
        <div className="min-h-screen bg-gray-50 flex flex-col">

            {/* Top nav */}
            <header className="w-full flex items-center justify-between px-8 py-4 bg-white border-b border-gray-100 shadow-sm sticky top-0 z-10">
                <span className="text-lg font-bold text-[#6B5FAE] tracking-tight">Portal</span>
                <Button
                    onClick={() => navigate('/login')}
                    className="bg-[#6B5FAE] hover:bg-[#5a4f9a] text-white rounded-xl text-sm font-semibold gap-2 px-5"
                >
                    <LogIn className="w-4 h-4" />
                    Login
                </Button>
            </header>

            {/* Main content */}
            <main className="flex-1 px-6 py-8 max-w-7xl mx-auto w-full">
                <div className="grid grid-cols-1 gap-6">

                    {/* Header */}
                    <div>
                        <h1 className="text-2xl font-bold text-[#6B5FAE]">Dokumen Publik</h1>
                        <p className="text-sm text-gray-500 mt-1">
                            Dokumen yang tersedia untuk umum — SPO, Surat Keputusan, Materi, Panduan, dan Pedoman.
                        </p>
                    </div>

                    {/* Type tabs */}
                    <div className="flex flex-wrap gap-2">
                        {DOCUMENT_TYPES.map(t => (
                            <button
                                key={t}
                                onClick={() => setActiveType(t)}
                                className={`px-5 py-2 rounded-xl text-sm font-semibold transition-colors ${
                                    activeType === t
                                        ? 'bg-[#6B5FAE] text-white'
                                        : 'bg-white border border-gray-200 text-gray-500 hover:border-[#6B5FAE] hover:text-[#6B5FAE]'
                                }`}
                            >
                                {t}
                            </button>
                        ))}
                    </div>

                    {/* Content */}
                    {isLoading ? (
                        <div className="flex items-center justify-center py-20">
                            <Loader2 className="w-8 h-8 text-[#6B5FAE] animate-spin" />
                        </div>
                    ) : (
                        <DocumentTypeTab
                            activeType={activeType}
                            allDocuments={allDocuments}
                            onRowClick={handleRowClick}
                        />
                    )}
                </div>
            </main>

            {/* Footer */}
            <footer className="text-center py-4 text-xs text-gray-400 border-t border-gray-100">
                © {new Date().getFullYear()} Portal. All rights reserved.
            </footer>

            {/* Detail modal */}
            <PublicDocumentModal
                document={selectedDoc}
                open={modalOpen}
                isLoading={urlLoading}
                fileUrl={urlData?.url ?? null}
                onClose={handleClose}
            />
        </div>
    )
}