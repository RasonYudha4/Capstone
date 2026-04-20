import { useState, useRef, type KeyboardEvent } from 'react'
import { Send, Paperclip } from 'lucide-react'

interface InputBarProps {
    placeholder?: string
    onSend: (value: string) => void | Promise<void>
    disabled?: boolean
    showAttachment?: boolean
}

export default function InputBar({
    placeholder = 'Ketik sesuatu...',
    onSend,
    disabled,
    showAttachment = true,
}: InputBarProps) {
    const [value, setValue] = useState('')
    const ref = useRef<HTMLInputElement>(null)

    const handleSend = () => {
        const trimmed = value.trim()
        if (!trimmed || disabled) return
        onSend(trimmed)
        setValue('')
        ref.current?.focus()
    }

    const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault()
            handleSend()
        }
    }

    return (
        <div className="flex items-center gap-2 px-3 py-3 bg-white border-t border-gray-100">
            {showAttachment && (
                <button
                    type="button"
                    aria-label="Lampirkan file"
                    className="w-9 h-9 rounded-full border border-gray-200 flex items-center justify-center text-gray-400 hover:text-[#6B5FAE] hover:border-[#6B5FAE] transition-colors shrink-0"
                >
                    <Paperclip className="w-4 h-4" />
                </button>
            )}
            <input
                ref={ref}
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
                disabled={!value.trim() || disabled}
                aria-label="Kirim"
                className="text-[#6B5FAE] disabled:text-gray-300 transition-colors shrink-0"
            >
                <Send className="w-5 h-5" />
            </button>
        </div>
    )
}