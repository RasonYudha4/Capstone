// ChatPanel.tsx — now a "dumb" controlled component
import { useEffect, useRef } from 'react'
import { FileSearch } from 'lucide-react'
import { ScrollArea } from '@/components/ui/scroll-area'
import PanelHeader from '../atoms/PanelHeader'
import InputBar from '../molecules/InputBar'
import MessageBubble, { type Message } from '../molecules/MessageBubble'

interface ChatPanelProps {
    onClose: () => void
    messages: Message[]
    isLoading: boolean
    error: Error | null
    onSend: (content: string, file?: File) => void
}

export default function ChatPanel({ onClose, messages, isLoading, error, onSend }: ChatPanelProps) {
    const bottomRef = useRef<HTMLDivElement>(null)

    useEffect(() => {
        bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
    }, [messages])

    return (
        <div className="flex flex-col w-85 h-120 bg-gray-50 rounded-2xl shadow-2xl overflow-hidden border border-gray-100">
            <PanelHeader
                icon={FileSearch}
                title="System Assistant"
                onClose={onClose}
                className="bg-[#6B5FAE] rounded-t-2xl"
            />
            <ScrollArea className="h-90">
                <div className="flex flex-col gap-4 p-4">
                    {messages.map((msg) => (
                        <MessageBubble key={msg.id} message={msg} />
                    ))}
                    {error && (
                        <p className="text-xs text-red-400 text-center">{error.message}</p>
                    )}
                    <div ref={bottomRef} />
                </div>
            </ScrollArea>
            <InputBar onSend={onSend} disabled={isLoading} showAttachment />
        </div>
    )
}