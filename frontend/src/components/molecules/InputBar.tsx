import { useState, useRef, type KeyboardEvent, type ChangeEvent } from 'react'
import { Send, Paperclip, X } from 'lucide-react'

interface InputBarProps {
    placeholder?: string
    onSend: (value: string, file?: File) => void | Promise<void>
    disabled?: boolean
    showAttachment?: boolean
}

const ACCEPTED_TYPES = '.pdf,.docx'

export default function InputBar({
    placeholder = 'Ketik sesuatu...',
    onSend,
    disabled,
    showAttachment = true,
}: InputBarProps) {
    const [value, setValue] = useState('')
    const [file, setFile] = useState<File | undefined>(undefined)
    const inputRef = useRef<HTMLInputElement>(null)
    const fileRef = useRef<HTMLInputElement>(null)

    const handleSend = () => {
        const trimmed = value.trim()
        if (!trimmed || disabled) return
        onSend(trimmed, file)
        setValue('')
        setFile(undefined)
        inputRef.current?.focus()
    }

    const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault()
            handleSend()
        }
    }

    const handleFileChange = (e: ChangeEvent<HTMLInputElement>) => {
        const selected = e.target.files?.[0]
        if (selected) setFile(selected)
        e.target.value = '' // allow re-selecting the same file later
    }

    return (
        <div className="flex flex-col gap-2 px-3 py-3 bg-white border-t border-gray-100">
            {file && (
                <div className="flex items-center gap-2 bg-gray-100 rounded-md px-2.5 py-1.5 text-xs text-gray-600 w-fit max-w-full border border-gray-200">
                    <Paperclip className="w-3.5 h-3.5 text-gray-400 shrink-0" />
                    <span className="truncate max-w-44">{file.name}</span>
                    <button
                        type="button"
                        onClick={() => setFile(undefined)}
                        aria-label="Hapus lampiran"
                        className="text-gray-400 hover:text-red-400 shrink-0 p-0.5"
                    >
                        <X className="w-3 h-3" />
                    </button>
                </div>
            )}

            <div className="flex items-center gap-2">
                {showAttachment && (
                    <>
                        <input
                            ref={fileRef}
                            type="file"
                            accept={ACCEPTED_TYPES}
                            onChange={handleFileChange}
                            className="hidden"
                        />
                        <button
                            type="button"
                            aria-label="Lampirkan file"
                            onClick={() => fileRef.current?.click()}
                            disabled={disabled}
                            className="w-9 h-9 rounded-full border border-gray-200 flex items-center justify-center text-gray-400 hover:text-[#6B5FAE] hover:border-[#6B5FAE] transition-colors shrink-0 disabled:opacity-50"
                        >
                            <Paperclip className="w-4 h-4" />
                        </button>
                    </>
                )}
                <input
                    ref={inputRef}
                    type="text"
                    value={value}
                    onChange={(e) => setValue(e.target.value)}
                    onKeyDown={handleKeyDown}
                    placeholder={placeholder}
                    disabled={disabled}
                    className="flex-1 bg-gray-100 rounded-full px-4 py-2 text-sm text-gray-800 placeholder-gray-400 outline-none focus:ring-2 focus:ring-[#6B5FAE]/30 disabled:opacity-50 transition-all"
                />
                <button
                    type="button"
                    onClick={handleSend}
                    disabled={(!value.trim() && !file) || disabled}
                    aria-label="Kirim"
                    className="text-[#6B5FAE] disabled:text-gray-300 transition-colors shrink-0"
                >
                    <Send className="w-5 h-5" />
                </button>
            </div>
        </div>
    )
}