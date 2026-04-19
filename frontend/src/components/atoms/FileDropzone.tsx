import { useRef, useState } from 'react'
import { Upload } from 'lucide-react'

interface FileDropzoneProps {
    onFileSelect: (file: File) => void
    file?: File | null
}

export default function FileDropzone({ onFileSelect, file }: FileDropzoneProps) {
    const [isDragging, setIsDragging] = useState(false)
    const inputRef = useRef<HTMLInputElement>(null)

    const handleDrop = (e: React.DragEvent) => {
        e.preventDefault()
        setIsDragging(false)
        const dropped = e.dataTransfer.files[0]
        if (dropped) onFileSelect(dropped)
    }

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const selected = e.target.files?.[0]
        if (selected) onFileSelect(selected)
    }

    return (
        <div
            onClick={() => inputRef.current?.click()}
            onDragOver={(e) => { e.preventDefault(); setIsDragging(true) }}
            onDragLeave={() => setIsDragging(false)}
            onDrop={handleDrop}
            className={`w-full h-full min-h-50 rounded-xl border-2 border-dashed flex flex-col items-center justify-center gap-3 cursor-pointer transition-colors ${isDragging
                    ? 'border-white bg-white/10'
                    : 'border-white/30 bg-transparent hover:border-white/60 hover:bg-white/10'
                }`}
        >
            <input ref={inputRef} type="file" className="hidden" onChange={handleChange} />
            <Upload className="w-10 h-10 text-gray-400" />
            {file ? (
                <p className="text-sm text-[#6B5FAE] font-medium text-center px-4">{file.name}</p>
            ) : (
                <p className="text-sm text-gray-400 text-center px-4">Pilih file yang akan di upload</p>
            )}
        </div>
    )
}