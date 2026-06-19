import { X, FileText, Minus, Plus, Loader2 } from 'lucide-react'
import { Dialog, DialogContent } from '@/components/ui/dialog'
import type { DocumentResponse } from '@/dtos/document_dto'
import { useEffect, useState } from 'react'

interface PublicDocumentModalProps {
    document: DocumentResponse | null
    open: boolean
    isLoading?: boolean
    fileUrl?: string | null
    onClose: () => void
}

export default function PublicDocumentModal({
    document,
    open,
    isLoading = false,
    fileUrl,
    onClose,
}: PublicDocumentModalProps) {
    const [zoom, setZoom] = useState(100)

    useEffect(() => {
        if (open) {
            setZoom(100)
        }
    }, [open])

    const displayName = document?.filename ?? 'Unknown File'

    // We no longer branch on file extension (the backend only returns the
    // object name without an extension), so we always embed the document
    // in an iframe and let the browser/viewer figure out how to render it.
    const embedUrl = fileUrl ? `${fileUrl}#toolbar=0&zoom=${zoom}` : null

    return (
        <Dialog
            open={open}
            onOpenChange={(value) => {
                if (!value) {
                    onClose()
                }
            }}
        >
            <DialogContent
                className="
                    !w-[95vw]
                    !h-[95vh]
                    !max-w-[95vw]
                    !max-h-[95vh]
                    p-0
                    overflow-hidden
                    [&>button]:hidden
    "
            >
                <div className="flex flex-col h-full bg-[#2D2D2D]">

                    {/* Toolbar */}
                    <div className="flex items-center gap-3 px-4 py-3 border-b border-white/10 shrink-0">

                        <FileText className="w-4 h-4 text-white/60 shrink-0" />

                        <span className="text-white text-sm truncate">
                            {displayName}
                        </span>

                        <div className="flex items-center gap-2 ml-auto">

                            <button
                                onClick={() =>
                                    setZoom((z) =>
                                        Math.max(50, z - 10)
                                    )
                                }
                                className="text-white/70 hover:text-white"
                            >
                                <Minus className="w-4 h-4" />
                            </button>

                            <span className="text-white/70 text-xs w-12 text-center">
                                {zoom}%
                            </span>

                            <button
                                onClick={() =>
                                    setZoom((z) =>
                                        Math.min(300, z + 10)
                                    )
                                }
                                className="text-white/70 hover:text-white"
                            >
                                <Plus className="w-4 h-4" />
                            </button>
                        </div>

                        {fileUrl && (
                            <a
                                href={fileUrl}
                                download={displayName}
                                className="text-white/70 hover:text-white"
                                title="Download"
                            >
                                <svg
                                    className="w-4 h-4"
                                    viewBox="0 0 24 24"
                                    fill="none"
                                    stroke="currentColor"
                                    strokeWidth={2}
                                >
                                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                                    <path d="M7 10l5 5 5-5" />
                                    <path d="M12 15V3" />
                                </svg>
                            </a>
                        )}

                        <button
                            onClick={onClose}
                            className="text-white/70 hover:text-white ml-2"
                        >
                            <X className="w-4 h-4" />
                        </button>
                    </div>

                    {/* Content */}
                    <div className="flex-1 bg-[#1E1E1E] overflow-hidden relative">

                        {isLoading && (
                            <div className="absolute inset-0 flex flex-col items-center justify-center gap-3">
                                <Loader2 className="w-8 h-8 animate-spin text-white/40" />
                                <p className="text-white/40 text-sm">
                                    Loading document...
                                </p>
                            </div>
                        )}

                        {!isLoading && !embedUrl && (
                            <div className="absolute inset-0 flex flex-col items-center justify-center gap-3">
                                <FileText className="w-10 h-10 text-white/20" />
                                <p className="text-white/30">
                                    Preview unavailable
                                </p>
                            </div>
                        )}

                        {/* Always render the document in an iframe — we don't
                            have a reliable extension to branch on, since the
                            backend only returns the object name. */}
                        {!isLoading && embedUrl && (
                            <iframe
                                key={embedUrl}
                                src={embedUrl}
                                title={displayName}
                                className="w-full h-full border-0"
                            />
                        )}
                    </div>
                </div>
            </DialogContent>
        </Dialog>
    )
}