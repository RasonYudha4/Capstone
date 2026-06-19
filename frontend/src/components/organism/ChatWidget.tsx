import { useState } from 'react'
import { MessageCircleQuestion, X } from 'lucide-react'
import ChatPanel from './ChatPanel'
import { useAgentChat } from '@/hooks/useAgentChat'

export default function ChatWidget() {
    const [isOpen, setIsOpen] = useState(false)
    const { messages, isLoading, error, send } = useAgentChat()

    return (
        <div className="fixed bottom-6 right-6 z-50 flex flex-col items-end gap-3">
            {isOpen && (
                <ChatPanel
                    onClose={() => setIsOpen(false)}
                    messages={messages}
                    isLoading={isLoading}
                    error={error}
                    onSend={send}
                />
            )}

            <div className="relative">
                <button
                    onClick={() => setIsOpen((p) => !p)}
                    aria-label={isOpen ? 'Tutup chat' : 'Buka chat'}
                    className="w-14 h-14 rounded-full bg-[#6B5FAE] text-white shadow-lg flex items-center justify-center hover:bg-[#5a4f9a] active:scale-95 transition-all duration-200"
                >
                    {isOpen ? <X className="w-5 h-5" /> : <MessageCircleQuestion className="w-5 h-5" />}
                </button>

                {!isOpen && (
                    <div className="absolute bottom-16 right-0 bg-[#6B5FAE] text-white text-sm font-medium px-4 py-2.5 rounded-2xl rounded-br-sm whitespace-nowrap shadow-md pointer-events-none">
                        Ada yang ingin ditanyakan?
                    </div>
                )}
            </div>
        </div>
    )
}