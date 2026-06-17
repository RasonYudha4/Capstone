import { useState, useMemo, useEffect } from 'react'
import { Upload, ChevronLeft, ChevronRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'

import BreadcrumbNav, { type BreadcrumbSegment } from '../molecules/BreadcrumbNav'
import FileTableRow, { type FileRecord } from '../molecules/FileTableRow'
import FileDetailModal from './FiledetailModal'

import { useFormOptions } from '@/hooks/useFormOption'
import {
  useDocumentsByService,
  useDocumentsByStandard,
  useDocumentsByAssessment,
  documentKeys,
} from '@/hooks/useDocument'
import type { DocumentResponse } from '@/dtos/document_dto'
import { documentService } from '@/services/document_services'
import { useMe } from '@/hooks/useAuth'
interface FileTableSectionProps {
  onUploadClick: () => void
}

const PAGE_SIZE = 10

const formatDate = (d: string) =>
  new Date(d).toLocaleDateString('id-ID', {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  })

// ─── Pagination UI ────────────────────────────────────────────────────────────

interface PaginationProps {
  page: number
  hasMore: boolean
  onChange: (page: number) => void
}

function Pagination({ page, hasMore, onChange }: PaginationProps) {
  if (page === 1 && !hasMore) return null

  const getPages = (): number[] => {
    const start = Math.max(1, page - 2)
    const end   = page
    return Array.from({ length: end - start + 1 }, (_, i) => start + i)
  }

  return (
    <div className="flex items-center gap-1 py-2">
      <button
        onClick={() => onChange(page - 1)}
        disabled={page === 1}
        className="flex items-center justify-center w-8 h-8 rounded-lg text-gray-500 hover:bg-gray-100 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
      >
        <ChevronLeft className="w-4 h-4" />
      </button>

      {getPages().map((p) => (
        <button
          key={p}
          onClick={() => onChange(p)}
          className={`w-8 h-8 rounded-lg text-sm font-medium transition-colors ${
            p === page
              ? 'bg-[#6B5FAE] text-white'
              : 'text-gray-600 hover:bg-gray-100'
          }`}
        >
          {p}
        </button>
      ))}

      <button
        onClick={() => onChange(page + 1)}
        disabled={!hasMore}
        className="flex items-center justify-center w-8 h-8 rounded-lg text-gray-500 hover:bg-gray-100 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
      >
        <ChevronRight className="w-4 h-4" />
      </button>
    </div>
  )
}

// ─── Main Component ───────────────────────────────────────────────────────────

export default function FileTableSection({ onUploadClick }: FileTableSectionProps) {
  const queryClient = useQueryClient()

  const { services, getStandards, getAssessments, isLoading: optionsLoading } = useFormOptions()

  const [selectedServiceId,    setSelectedServiceId]    = useState('')
  const [selectedStandardId,   setSelectedStandardId]   = useState('')
  const [selectedAssessmentId, setSelectedAssessmentId] = useState('')
  const [selectedDocId,        setSelectedDocId]        = useState<string | null>(null)
  const [selectedDocument,     setSelectedDocument]     = useState<DocumentResponse | null>(null)
  const [page,                 setPage]                 = useState(1)
  const { data: me } = useMe() 

  // ─────────────────────────────────────────────
  // Options
  // ─────────────────────────────────────────────

  const serviceOptions = useMemo(() => services.map(s => ({
    label: `${s.code} — ${s.description}`,
    value: s.id,
  })), [services])

  const standardOptions = useMemo(() =>
    getStandards(selectedServiceId).map(s => ({
      label: `${s.code} — ${s.description}`,
      value: s.id,
    })),
  [selectedServiceId, getStandards])

  const assessmentOptions = useMemo(() =>
    getAssessments(selectedServiceId, selectedStandardId).map(a => ({
      label: `${a.code} — ${a.description}`,
      value: a.id,
    })),
  [selectedServiceId, selectedStandardId, getAssessments])

  // ─────────────────────────────────────────────
  // Dynamic Query Selection
  // ─────────────────────────────────────────────

  const paginationQuery = useMemo(() => ({ page, limit: PAGE_SIZE }), [page])

  const serviceQuery    = useDocumentsByService(selectedServiceId, paginationQuery)
  const standardQuery   = useDocumentsByStandard(selectedStandardId, paginationQuery)
  const assessmentQuery = useDocumentsByAssessment(selectedAssessmentId, paginationQuery)

  const { data: docData, isLoading: docsLoading } = useMemo(() => {
    if (selectedAssessmentId) return assessmentQuery
    if (selectedStandardId)   return standardQuery
    if (selectedServiceId)    return serviceQuery
    return { data: undefined, isLoading: false }
  }, [
    selectedServiceId, selectedStandardId, selectedAssessmentId,
    serviceQuery, standardQuery, assessmentQuery,
  ])

  // ─────────────────────────────────────────────
  // Pagination derived state
  // ─────────────────────────────────────────────

  const rawDocs = docData?.data ?? []
  const hasMore = rawDocs.length === PAGE_SIZE

  const handlePageChange = (newPage: number) => {
    if (newPage < 1) return
    if (newPage > page && !hasMore) return
    setPage(newPage)
    document.getElementById('file-table-top')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }

  // ─────────────────────────────────────────────
  // File URL — fetch as blob so auth headers are included
  // ─────────────────────────────────────────────

  const [fileUrl,       setFileUrl]       = useState('')
  const [detailLoading, setDetailLoading] = useState(false)

  useEffect(() => {
    if (!selectedDocId) {
      setFileUrl('')
      return
    }

    let objectUrl = ''
    setDetailLoading(true)

    documentService.getById(selectedDocId)
      .then(({ url }) => {
        objectUrl = url
        setFileUrl(url)
      })
      .catch(() => toast.error('Gagal memuat berkas.'))
      .finally(() => setDetailLoading(false))

    return () => {
      if (objectUrl) URL.revokeObjectURL(objectUrl)
    }
  }, [selectedDocId])

  // ─────────────────────────────────────────────
  // Table data
  // ─────────────────────────────────────────────

  const files: FileRecord[] = useMemo(() =>
    rawDocs.map(d => ({
      id:          d.document_id,
      name:        d.filename,
      type:        d.document_type,
      uploadedBy:  d.created_by,
      lastUpdated: formatDate(d.updated_at),
      status:      d.status,
    })),
  [rawDocs])

  // ─────────────────────────────────────────────
  // Row click
  // ─────────────────────────────────────────────

  const handleRowClick = (file: FileRecord) => {
    const doc = rawDocs.find(d => d.document_id === file.id) ?? null
    setSelectedDocument(doc)
    setSelectedDocId(file.id)
  }

  const selectedFile: FileRecord | null = selectedDocument
    ? {
        id:          selectedDocument.document_id,
        name:        selectedDocument.filename,
        type:        selectedDocument.document_type,
        uploadedBy:  selectedDocument.created_by,
        lastUpdated: formatDate(selectedDocument.updated_at),
        status:      selectedDocument.status,
      }
    : null

  // ─────────────────────────────────────────────
  // Helpers
  // ─────────────────────────────────────────────

  const invalidateList = () => {
    if (selectedAssessmentId)
      queryClient.invalidateQueries({ queryKey: documentKeys.byAssessment(selectedAssessmentId) })
    else if (selectedStandardId)
      queryClient.invalidateQueries({ queryKey: documentKeys.byStandard(selectedStandardId) })
    else if (selectedServiceId)
      queryClient.invalidateQueries({ queryKey: documentKeys.byService(selectedServiceId) })
  }

  const closeModal = () => {
    setSelectedDocId(null)
    setSelectedDocument(null)
    setFileUrl('')
  }

  const resetPage = () => setPage(1)

  // ─────────────────────────────────────────────
  // Action handlers
  // ─────────────────────────────────────────────

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
      if (files.length === 1 && page > 1) setPage(p => p - 1)
    } catch {
      toast.error('Gagal menghapus berkas.')
    }
  }

  // ─────────────────────────────────────────────
  // Breadcrumb
  // ─────────────────────────────────────────────

  const segments: BreadcrumbSegment[] = [
    {
      selected: optionsLoading
        ? 'Memuat...'
        : serviceOptions.find(o => o.value === selectedServiceId)?.label ?? 'Pilih Fungsi Pelayanan',
      options: serviceOptions,
      onChange: (v) => {
        setSelectedServiceId(v)
        setSelectedStandardId('')
        setSelectedAssessmentId('')
        resetPage()
      },
    },
    {
      selected: standardOptions.find(o => o.value === selectedStandardId)?.label ?? 'Pilih Standar',
      options: standardOptions,
      onChange: (v) => {
        setSelectedStandardId(v)
        setSelectedAssessmentId('')
        resetPage()
      },
    },
    {
      selected: assessmentOptions.find(o => o.value === selectedAssessmentId)?.label ?? 'Pilih Elemen Penilaian',
      options: assessmentOptions,
      onChange: (v) => { setSelectedAssessmentId(v); resetPage() },
    },
  ]

  // ─────────────────────────────────────────────
  // Render
  // ─────────────────────────────────────────────

  return (
    <section className="bg-white rounded-2xl border border-gray-100 overflow-hidden">

      <FileDetailModal
        file={selectedFile}
        document={selectedDocument}
        isLoading={detailLoading}
        fileUrl={fileUrl}
        open={!!selectedDocId}
        onOpenChange={(open) => { if (!open) closeModal() }}
        onApprove={handleApprove}
        onReject={handleReject}
        onUpdate={handleUpdate}
        onDelete={handleDelete}
        role={me?.role}
      />

      {/* Toolbar */}
      <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100">
        <BreadcrumbNav segments={segments} />
        <Button
          onClick={onUploadClick}
          className="bg-[#6B5FAE] hover:bg-[#5a4f9a] text-white rounded-xl text-sm gap-2"
        >
          <Upload className="w-4 h-4" />
          Upload new file
        </Button>
      </div>

      {/* Table */}
      <div className="px-6 py-4" id="file-table-top">
        <div className="rounded-2xl overflow-hidden border border-gray-100">
          <div className="overflow-x-auto">
            <Table className="min-w-175">
              <TableHeader>
                <TableRow className="bg-[#6B5FAE]">
                  {['Nama Berkas', 'Tipe Berkas', 'Di Upload oleh', 'Terakhir diperbaharui', 'Status'].map(h => (
                    <TableHead key={h} className="text-white text-center text-sm px-4 py-3">
                      {h}
                    </TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {docsLoading ? (
                  <tr>
                    <td colSpan={5} className="text-center text-sm text-gray-400 py-10">Memuat berkas...</td>
                  </tr>
                ) : !selectedServiceId ? (
                  <tr>
                    <td colSpan={5} className="text-center text-sm text-gray-400 py-10">Pilih Fungsi Pelayanan terlebih dahulu.</td>
                  </tr>
                ) : files.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="text-center text-sm text-gray-400 py-10">Tidak ada berkas.</td>
                  </tr>
                ) : (
                  files.map(file => (
                    <FileTableRow
                      key={file.id}
                      file={file}
                      onClick={() => handleRowClick(file)}
                    />
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </div>

        {/* Pagination */}
        {!docsLoading && files.length > 0 && (
          <div className="flex items-center justify-between mt-3 px-1">
            <p className="text-xs text-gray-400">
              Halaman {page}{!hasMore && ` · ${(page - 1) * PAGE_SIZE + rawDocs.length} berkas total`}
            </p>
            <Pagination
              page={page}
              hasMore={hasMore}
              onChange={handlePageChange}
            />
          </div>
        )}
      </div>
    </section>
  )
}