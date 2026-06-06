import { useEffect, useRef, useState } from 'react'
import { FileSearch } from 'lucide-react'
import { ScrollArea } from '@/components/ui/scroll-area'
import { queryService } from '@/services/query_service'
import PanelHeader from '../atoms/PanelHeader'
import InputBar from '../molecules/InputBar'
import MessageBubble, { type Message } from '../molecules/MessageBubble'

const INITIAL_MESSAGE: Message = {
    id: 'init',
    role: 'assistant',
    content: 'Halo, ada yang bisa dibantu?',
    timestamp: new Date(),
}

interface ChatPanelProps {
    onClose: () => void
}

export default function ChatPanel({ onClose }: ChatPanelProps) {
    const [messages, setMessages] = useState<Message[]>([INITIAL_MESSAGE])
    const [isLoading, setIsLoading] = useState(false)
    const [error, setError] = useState<Error | null>(null)
    const bottomRef = useRef<HTMLDivElement>(null)
    const streamingIdRef = useRef<string | null>(null)

    useEffect(() => {
        bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
    }, [messages])

    const handleSend = async (content: string) => {
        setError(null)
        setIsLoading(true)

        const assistantId = crypto.randomUUID()
        streamingIdRef.current = assistantId

        setMessages((prev) => [
            ...prev,
            { id: crypto.randomUUID(), role: 'user', content, timestamp: new Date() },
            { id: assistantId, role: 'assistant', content: '', timestamp: new Date() },
        ])

        try {
            await queryService.queryStream(
                { question: content },
                (chunk) => {
                    setMessages((prev) =>
                        prev.map((m) =>
                            m.id === streamingIdRef.current
                                ? { ...m, content: m.content + chunk }
                                : m
                        )
                    )
                },
            )
        } catch (err) {
            setError(err instanceof Error ? err : new Error('Stream failed.'))
        } finally {
            setIsLoading(false)
        }
    }

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
            <InputBar onSend={handleSend} disabled={isLoading} showAttachment />
        </div>
    )
}