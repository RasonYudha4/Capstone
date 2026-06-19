import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { X, Minus, Plus, FileText, Loader2, CheckCircle, XCircle, RefreshCw, Trash2 } from 'lucide-react'
import {
    Dialog,
    DialogContent,
} from '@/components/ui/dialog'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import { ScrollArea } from '@/components/ui/scroll-area'
import StatusPill, { type FileStatus } from '../atoms/StatusPill'
import type { FileRecord } from '../molecules/FileTableRow'
import FileDropzone from '../molecules/FileDropzone'
import { useState, useEffect } from 'react'
import ConfirmDialog from '../molecules/ConfirmDialog'
import type { DocumentResponse } from '@/dtos/document_dto'


const reviewSchema = z.object({
    catatan: z.string().optional(),
    filename: z.string().optional(),
})

type ReviewFormValues = z.infer<typeof reviewSchema>

interface FileDetailModalProps {
    file: FileRecord | null
    document?: DocumentResponse | null
    open: boolean
    isLoading?: boolean
    fileUrl?: string
    role?: 'staff' | 'admin' | 'master-admin'
    onOpenChange: (open: boolean) => void
    onApprove?: (file: FileRecord, catatan: string, attachment?: File) => void
    onReject?: (file: FileRecord, catatan: string, attachment?: File) => void
    onUpdate?: (file: FileRecord, catatan: string, attachment?: File, filename?: string) => Promise<void> | void
    onDelete?: (file: FileRecord) => void
}

// ── helpers ────────────────────────────────────────────────────────────────

function formatDate(iso?: string | null): string {
    if (!iso) return '—'
    return new Date(iso).toLocaleString('id-ID', {
        day: '2-digit',
        month: 'long',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
    })
}

function MetaRow({ label, value }: { label: string; value?: string | null }) {
    return (
        <div>
            <p className="text-[11px] font-medium text-white/50 uppercase tracking-wider">{label}</p>
            <p className="text-sm font-semibold text-white mt-0.5 break-words">{value ?? '—'}</p>
        </div>
    )
}

// ── component ──────────────────────────────────────────────────────────────

export default function FileDetailModal({
    file,
    document,
    open,
    isLoading = false,
    fileUrl,
    role,
    onOpenChange,
    onApprove,
    onReject,
    onUpdate,
    onDelete,
}: FileDetailModalProps) {
    // ── role-based permission ──
    const canReview = role === 'master-admin'

    // ── display values ──
    const displayName      = document?.filename     ?? file?.name       ?? '—'
    const displayCreatedBy = document?.created_by   ?? file?.uploadedBy ?? '—'
    const displayUpdatedAt = formatDate(document?.updated_at)
    const displayStatus    = (document?.status ?? file?.status) as FileStatus | undefined

    // strip extension for display and form default
    const strippedName = displayName.replace(/\.[^.]+$/, '')

    const { register, handleSubmit, reset } = useForm<ReviewFormValues>({
        resolver: zodResolver(reviewSchema),
        defaultValues: {
            filename: strippedName,
            catatan: '',
        },
    })

    const [zoom, setZoom] = useState(100)
    const [attachedFile, setAttachedFile] = useState<File | null>(null)
    const [isUpdating, setIsUpdating] = useState(false)

    const [confirm, setConfirm] = useState<{
        open: boolean
        title: string
        description: string
        confirmLabel: string
        variant: 'danger' | 'warning' | 'default'
        onConfirm: () => void
    }>({ open: false, title: '', description: '', confirmLabel: '', variant: 'default', onConfirm: () => {} })

    // ── sync form when document/file changes ──
    useEffect(() => {
        const name = (document?.filename ?? file?.name ?? '').replace(/\.[^.]+$/, '')
        console.log('[FileDetailModal] syncing form, name:', name)
        reset({
            filename: name,
            catatan: '',
        })
    }, [document?.filename, file?.name, reset])

    const openConfirm = (config: Omit<typeof confirm, 'open'>) =>
        setConfirm({ open: true, ...config })
    const closeConfirm = () => setConfirm((prev) => ({ ...prev, open: false }))

    const handleClose = () => {
        reset({ filename: strippedName, catatan: '' })
        setAttachedFile(null)
        setZoom(100)
        onOpenChange(false)
    }

    // ── action handlers ──

    const handleApprove = handleSubmit((data) => {
        if (!file) return
        openConfirm({
            title: 'Setujui Berkas?',
            description: `Anda akan menyetujui berkas "${file.name}".`,
            confirmLabel: 'Ya, Setujui',
            variant: 'default',
            onConfirm: () => {
                closeConfirm()
                onApprove?.(file, data.catatan ?? '', attachedFile ?? undefined)
            },
        })
    })

    const handleReject = handleSubmit((data) => {
        if (!file) return
        openConfirm({
            title: 'Tolak Berkas?',
            description: `Anda akan menolak berkas "${file.name}". Tindakan ini tidak dapat dibatalkan.`,
            confirmLabel: 'Ya, Tolak',
            variant: 'danger',
            onConfirm: () => {
                closeConfirm()
                onReject?.(file, data.catatan ?? '', attachedFile ?? undefined)
            },
        })
    })

    const handleUpdate = handleSubmit((data) => {
        if (!file) return

        const newFilename = data.filename?.trim() || undefined

        console.log('[FileDetailModal] handleUpdate called')
        console.log('  file:', file)
        console.log('  data.filename:', data.filename)
        console.log('  newFilename (sent to backend):', newFilename)
        console.log('  attachedFile:', attachedFile?.name ?? 'none')
        console.log('  catatan:', data.catatan)

        openConfirm({
            title: 'Update Status Berkas?',
            description: `Anda akan memperbarui berkas "${file.name}". Pastikan data sudah benar.`,
            confirmLabel: 'Ya, Update',
            variant: 'warning',
            onConfirm: async () => {
                closeConfirm()
                setIsUpdating(true)
                try {
                    await onUpdate?.(file, data.catatan ?? '', attachedFile ?? undefined, newFilename)
                    console.log('[FileDetailModal] update success, reloading page...')
                    onOpenChange(false)
                    window.location.reload()
                } catch (err) {
                    console.error('[FileDetailModal] update failed:', err)
                } finally {
                    setIsUpdating(false)
                }
            },
        })
    })

    const handleDelete = () => {
        if (!file) return
        openConfirm({
            title: 'Hapus Berkas?',
            description: `Anda akan menghapus berkas "${file.name}" secara permanen. Tindakan ini tidak dapat dibatalkan.`,
            confirmLabel: 'Ya, Hapus',
            variant: 'danger',
            onConfirm: () => {
                closeConfirm()
                onDelete?.(file)
            },
        })
    }

    // ── preview url ──
    // We no longer branch on file extension (the backend only returns the
    // object name without an extension), so we always embed the document
    // in an iframe and let the browser/viewer figure out how to render it.
    const embedUrl = fileUrl ? `${fileUrl}#toolbar=0&zoom=${zoom}` : null

    const disabled = isLoading || isUpdating || !file

    return (
        <>
            <ConfirmDialog
                open={confirm.open}
                onOpenChange={closeConfirm}
                title={confirm.title}
                description={confirm.description}
                confirmLabel={confirm.confirmLabel}
                variant={confirm.variant}
                onConfirm={confirm.onConfirm}
            />

            <Dialog open={open} onOpenChange={handleClose}>
                <DialogContent className="max-w-[95vw] min-w-340 p-0 rounded-2xl overflow-hidden border-0 shadow-2xl gap-0 [&>button]:hidden">
                    <div className="flex h-170">

                        {/* ── Left — file preview ─────────────────────────────── */}
                        <div className="flex-1 bg-[#2D2D2D] flex flex-col min-w-0">

                            {/* Toolbar */}
                            <div className="flex items-center gap-3 px-4 py-3 border-b border-white/10 shrink-0">
                                <FileText className="w-4 h-4 text-white/60 shrink-0" />
                                <span className="text-white/80 text-xs font-medium truncate">{displayName}</span>

                                <div className="flex items-center gap-2 ml-auto shrink-0">
                                    <button
                                        onClick={() => setZoom((z) => Math.max(50, z - 10))}
                                        className="text-white/60 hover:text-white transition-colors"
                                    >
                                        <Minus className="w-3.5 h-3.5" />
                                    </button>
                                    <span className="text-white/60 text-xs w-10 text-center">{zoom}%</span>
                                    <button
                                        onClick={() => setZoom((z) => Math.min(200, z + 10))}
                                        className="text-white/60 hover:text-white transition-colors"
                                    >
                                        <Plus className="w-3.5 h-3.5" />
                                    </button>
                                </div>

                                {fileUrl && (
                                    <a
                                        href={fileUrl}
                                        download={displayName}
                                        className="text-white/60 hover:text-white transition-colors shrink-0"
                                        title="Download"
                                    >
                                        <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
                                            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3" />
                                        </svg>
                                    </a>
                                )}
                            </div>

                            {/* Preview area */}
                            <div className="flex-1 overflow-hidden relative">
                                {(isLoading || isUpdating) && (
                                    <div className="absolute inset-0 flex flex-col items-center justify-center gap-3 bg-[#2D2D2D]">
                                        <Loader2 className="w-8 h-8 text-white/40 animate-spin" />
                                        <p className="text-white/40 text-xs">
                                            {isUpdating ? 'Menyimpan perubahan...' : 'Memuat berkas...'}
                                        </p>
                                    </div>
                                )}

                                {!isLoading && !isUpdating && !embedUrl && (
                                    <div className="absolute inset-0 flex flex-col items-center justify-center gap-2">
                                        <FileText className="w-10 h-10 text-white/20" />
                                        <p className="text-white/30 text-xs">Pratinjau tidak tersedia</p>
                                    </div>
                                )}

                                {/* Always render the document in an iframe — we don't
                                    have a reliable extension to branch on, since the
                                    backend only returns the object name. */}
                                {!isLoading && !isUpdating && embedUrl && (
                                    <iframe
                                        key={embedUrl}
                                        src={embedUrl}
                                        className="w-full h-full border-0"
                                        title={displayName}
                                    />
                                )}
                            </div>
                        </div>

                        {/* ── Right — detail + review ──────────────────────────── */}
                        <div className="w-72 bg-[#6B5FAE] flex flex-col shrink-0">

                            {/* Header */}
                            <div className="flex items-center justify-between px-5 pt-4 pb-2 shrink-0">
                                <StatusPill status={displayStatus} />
                                <button onClick={handleClose} className="text-white/70 hover:text-white transition-colors">
                                    <X className="w-4 h-4" />
                                </button>
                            </div>

                            {/* Scrollable content */}
                            <ScrollArea className="flex-1 h-0">
                                {isLoading ? (
                                    <div className="flex flex-col gap-3 px-5 py-6 animate-pulse">
                                        <div className="h-5 bg-white/20 rounded w-4/5" />
                                        <div className="h-4 bg-white/10 rounded w-3/5" />
                                        <div className="h-px bg-white/20 my-1" />
                                        <div className="h-3 bg-white/10 rounded w-2/5" />
                                        <div className="h-4 bg-white/20 rounded w-3/5" />
                                        <div className="h-3 bg-white/10 rounded w-2/5 mt-2" />
                                        <div className="h-4 bg-white/20 rounded w-4/5" />
                                    </div>
                                ) : (
                                    <div className="px-5 py-3 flex flex-col gap-4">

                                        {/* File name + description */}
                                        <div>
                                            <h2 className="text-lg font-bold text-white leading-snug break-words">
                                                {strippedName}
                                            </h2>
                                            {document?.document_type && (
                                                <p className="text-xs text-white/60 mt-1 leading-relaxed">
                                                    {document.document_type}
                                                </p>
                                            )}
                                        </div>

                                        <Separator className="bg-white/20" />

                                        {/* Metadata */}
                                        <div className="flex flex-col gap-3">
                                            <MetaRow label="Diunggah Oleh" value={displayCreatedBy} />
                                            <MetaRow label="Terakhir Diperbarui" value={displayUpdatedAt} />
                                        </div>

                                        <Separator className="bg-white/20" />

                                        {/* Review form */}
                                        <div className="flex flex-col gap-2">
                                            <p className="text-xs font-semibold text-white/80">Nama Berkas</p>
                                            <Input
                                                {...register('filename')}
                                                placeholder={strippedName}
                                                className="bg-white/10 border-white/20 text-white placeholder:text-white/40 text-sm rounded-xl focus:border-white/50 focus:ring-0"
                                            />
                                            <p className="text-xs font-semibold text-white/80 mt-1">Catatan Review</p>
                                            <Textarea
                                                {...register('catatan')}
                                                placeholder="Tambahkan catatan untuk reviewer..."
                                                rows={3}
                                                className="bg-white/10 border-white/20 text-white placeholder:text-white/40 text-sm rounded-xl resize-none focus:border-white/50 focus:ring-0"
                                            />
                                            <p className="text-xs font-semibold text-white/80 mt-1">Lampiran</p>
                                            <FileDropzone
                                                file={attachedFile}
                                                onFileSelect={(f) => setAttachedFile(f)}
                                            />
                                        </div>
                                    </div>
                                )}
                            </ScrollArea>

                           {/* Action buttons */}
<div className="flex flex-col gap-2 px-5 py-4 shrink-0 border-t border-white/20">

    {/* Approve + Reject — master-admin only, hidden when approved */}
    {canReview && displayStatus !== 'approved' && (
        <div className="flex gap-2">
            <Button
                type="button"
                onClick={handleApprove}
                disabled={disabled}
                className="flex-1 bg-emerald-500 hover:bg-emerald-600 text-white rounded-xl font-semibold text-xs h-9 disabled:opacity-40 gap-1.5"
            >
                <CheckCircle className="w-3.5 h-3.5" />
                Setujui
            </Button>
            <Button
                type="button"
                onClick={handleReject}
                disabled={disabled}
                className="flex-1 bg-red-500 hover:bg-red-600 text-white rounded-xl font-semibold text-xs h-9 disabled:opacity-40 gap-1.5"
            >
                <XCircle className="w-3.5 h-3.5" />
                Tolak
            </Button>
        </div>
    )}

    {/* Update + Delete — Update hidden when approved */}
    <div className="flex gap-2">
        {displayStatus !== 'approved' && (
            <Button
                type="button"
                onClick={handleUpdate}
                disabled={disabled}
                className="flex-1 bg-white/20 hover:bg-white/30 text-white rounded-xl font-semibold text-xs h-9 disabled:opacity-40 gap-1.5"
            >
                {isUpdating
                    ? <Loader2 className="w-3.5 h-3.5 animate-spin" />
                    : <RefreshCw className="w-3.5 h-3.5" />
                }
                {isUpdating ? 'Menyimpan...' : 'Update'}
            </Button>
        )}
        <Button
            type="button"
            onClick={handleDelete}
            disabled={disabled}
            className="flex-1 bg-white/10 hover:bg-red-500/40 text-white/70 hover:text-white rounded-xl font-semibold text-xs h-9 disabled:opacity-40 gap-1.5 transition-colors"
        >
            <Trash2 className="w-3.5 h-3.5" />
            Hapus
        </Button>
    </div>
                            </div>
                        </div>
                    </div>
                </DialogContent>
            </Dialog>
        </>
    )
}