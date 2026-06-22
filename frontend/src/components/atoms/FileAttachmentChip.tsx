import { FileText } from 'lucide-react'

interface FileAttachmentChipProps {
    name: string
    type: string
}

function extLabel(type: string, name: string): string {
    if (type.includes('pdf')) return 'PDF'
    if (type.includes('word') || name.toLowerCase().endsWith('.docx')) return 'DOCX'
    return 'FILE'
}

export default function FileAttachmentChip({ name, type }: FileAttachmentChipProps) {
    return (
        <div className="flex items-center gap-2.5 bg-white/15 rounded-md px-3 py-2 max-w-full">
            <div className="w-7 h-7 rounded-md bg-white/15 flex items-center justify-center shrink-0">
                <FileText className="w-3.5 h-3.5 text-white" />
            </div>
            <div className="flex flex-col min-w-0">
                <span className="text-[10px] text-white/70 leading-tight">{extLabel(type, name)}</span>
                <span className="text-xs text-white leading-tight truncate max-w-44">{name}</span>
            </div>
        </div>
    )
}