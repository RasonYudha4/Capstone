import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { X, Minus, Plus, FileText } from 'lucide-react'
import {
    Dialog,
    DialogContent,
} from '@/components/ui/dialog'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { ScrollArea } from '@/components/ui/scroll-area'
import StatusPill, { type FileStatus } from '../atoms/StatusPill'
import type { FileRecord } from '../molecules/FileTableRow'
import FileDropzone from '../atoms/FileDropzone'
import { useState } from 'react'
import ConfirmDialog from '../molecules/ConfirmDialog'

const reviewSchema = z.object({
    catatan: z.string().optional(),
})

type ReviewFormValues = z.infer<typeof reviewSchema>

interface FileDetailModalProps {
    file: FileRecord | null
    open: boolean
    onOpenChange: (open: boolean) => void
    onReject?: (file: FileRecord, catatan: string, attachment?: File) => void
    onUpdate?: (file: FileRecord, catatan: string, attachment?: File) => void
}

export default function FileDetailModal({
    file,
    open,
    onOpenChange,
    onReject,
    onUpdate,
}: FileDetailModalProps) {
    const { register, handleSubmit, reset } = useForm<ReviewFormValues>({
        resolver: zodResolver(reviewSchema),
    })

    const [confirm, setConfirm] = useState<{
        open: boolean
        title: string
        description: string
        confirmLabel: string
        variant: 'danger' | 'warning' | 'default'
        onConfirm: () => void
    }>({ open: false, title: '', description: '', confirmLabel: '', variant: 'default', onConfirm: () => { } })

    const [attachedFile, setAttachedFile] = useState<File | null>(null)

    const openConfirm = (config: Omit<typeof confirm, 'open'>) => {
        setConfirm({ open: true, ...config })
    }
    const closeConfirm = () => setConfirm((prev) => ({ ...prev, open: false }))

    const handleClose = () => {
        reset()
        setAttachedFile(null)
        onOpenChange(false)
    }

    const handleReject = handleSubmit((data) => {
        if (!file) return
        openConfirm({
            title: 'Tolak Berkas?',
            description: `Anda akan menolak berkas "${file.name}". Tindakan ini tidak dapat dibatalkan.`,
            confirmLabel: 'Ya, Tolak',
            variant: 'danger',
            onConfirm: () => {
                onReject?.(file, data.catatan ?? '', attachedFile ?? undefined)
                closeConfirm()
                handleClose()
            },
        })
    })

    const handleUpdate = handleSubmit((data) => {
        if (!file) return
        openConfirm({
            title: 'Update Status Berkas?',
            description: `Anda akan memperbarui status berkas "${file.name}". Pastikan data sudah benar.`,
            confirmLabel: 'Ya, Update',
            variant: 'default',
            onConfirm: () => {
                onUpdate?.(file, data.catatan ?? '')
                closeConfirm()
                handleClose()
            },
        })
    })

    if (!file) return null

    return (
        <Dialog open={open} onOpenChange={handleClose}>
            <ConfirmDialog
                open={confirm.open}
                onOpenChange={closeConfirm}
                title={confirm.title}
                description={confirm.description}
                confirmLabel={confirm.confirmLabel}
                variant={confirm.variant}
                onConfirm={confirm.onConfirm}
            />
            <DialogContent className="max-w-[95vw] min-w-340 p-0 rounded-2xl overflow-hidden border-0 shadow-2xl gap-0 [&>button]:hidden">
                <div className="flex h-170">

                    {/* Left — PDF preview */}
                    <div className="flex-1 bg-[#2D2D2D] flex flex-col">
                        {/* PDF toolbar */}
                        <div className="flex items-center gap-3 px-4 py-3 border-b border-white/10">
                            <FileText className="w-4 h-4 text-white/60 shrink-0" />
                            <span className="text-white/80 text-xs font-medium truncate">{file.name}</span>
                            <div className="flex items-center gap-2 ml-auto shrink-0">
                                <button className="text-white/60 hover:text-white transition-colors"><Minus className="w-3.5 h-3.5" /></button>
                                <span className="text-white/60 text-xs">100 %</span>
                                <button className="text-white/60 hover:text-white transition-colors"><Plus className="w-3.5 h-3.5" /></button>
                                <button className="text-white/60 hover:text-white transition-colors ml-1">
                                    <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2}>
                                        <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3" />
                                    </svg>
                                </button>
                            </div>
                        </div>

                        {/* PDF mock page */}
                        <div className="flex-1 overflow-auto flex items-start justify-center p-6">
                            <div className="w-full max-w-sm bg-white rounded shadow-lg p-8 flex flex-col gap-3">
                                <div className="h-3 bg-gray-200 rounded w-3/5" />
                                <div className="h-3 bg-gray-200 rounded w-4/5" />
                                <div className="h-2 bg-gray-100 rounded w-full mt-2" />
                                {Array.from({ length: 12 }).map((_, i) => (
                                    <div key={i} className="h-2 bg-gray-100 rounded" style={{ width: `${75 + Math.random() * 25}%` }} />
                                ))}
                            </div>
                        </div>
                    </div>

                    {/* Right — File detail + review */}
                    <div className="w-70 bg-[#6B5FAE] flex flex-col shrink-0">
                        {/* Header */}
                        <div className="flex items-center justify-between px-5 pt-4 pb-2 shrink-0">
                            <StatusPill status={file.status as FileStatus} />
                            <button onClick={handleClose} className="text-white/70 hover:text-white transition-colors">
                                <X className="w-4 h-4" />
                            </button>
                        </div>

                        {/* Scrollable content */}
                        <ScrollArea className="flex-1 h-0">
                            <div className="px-5 py-3 flex flex-col gap-4">
                                {/* Title & description */}
                                <div>
                                    <h2 className="text-xl font-bold text-white leading-snug">{file.name.replace(/\.[^.]+$/, '')}</h2>
                                    <p className="text-sm text-white/70 mt-2 leading-relaxed">Berkas laporan elemen penilaian 2 standar 2</p>
                                </div>

                                <Separator className="bg-white/20" />

                                {/* Meta */}
                                <div className="flex flex-col gap-3">
                                    <div>
                                        <p className="text-xs text-white/60">Diunggah Oleh :</p>
                                        <p className="text-sm font-bold text-white mt-0.5">{file.uploadedBy}</p>
                                    </div>
                                    <div>
                                        <p className="text-xs text-white/60">Pada Tanggal :</p>
                                        <p className="text-sm font-bold text-white mt-0.5">{file.lastUpdated}</p>
                                    </div>
                                </div>

                                <Separator className="bg-white/20" />

                                {/* Review form */}
                                <div className="flex flex-col gap-2">
                                    <p className="text-xs font-semibold text-white/80">Catatan Review</p>
                                    <Textarea
                                        {...register('catatan')}
                                        placeholder="Tambahkan catatan untuk reviewer..."
                                        rows={4}
                                        className="bg-white/10 border-white/20 text-white placeholder:text-white/40 text-sm rounded-xl resize-none focus:border-white/50 focus:ring-0"
                                    />
                                    <p className="text-xs font-semibold text-white/80 mt-2">Lampiran</p>
                                    <FileDropzone
                                        file={attachedFile}
                                        onFileSelect={(f) => setAttachedFile(f)}
                                    />
                                </div>
                            </div>
                        </ScrollArea>

                        {/* Fixed action buttons */}
                        <div className="flex gap-2 px-5 py-4 shrink-0 border-t border-white/20">
                            <Button
                                type="button"
                                onClick={handleReject}
                                className="flex-1 bg-red-500 hover:bg-red-600 text-white rounded-xl font-semibold text-sm"
                            >
                                Tolak
                            </Button>
                            <Button
                                type="button"
                                onClick={handleUpdate}
                                className="flex-1 bg-white/20 hover:bg-white/30 text-white rounded-xl font-semibold text-sm"
                            >
                                Update
                            </Button>
                        </div>
                    </div>
                </div>
            </DialogContent>
        </Dialog>
    )
}