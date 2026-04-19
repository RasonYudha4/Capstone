import { useEffect, useRef, useState } from 'react'
import { FileSearch } from 'lucide-react'
import { ScrollArea } from '@/components/ui/scroll-area'
import PanelHeader from '../atoms/PanelHeader'
import InputBar from '../atoms/InputBar'
import MessageBubble, { type Message, TypingBubble } from '../molecules/MessageBubble'

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
    const [isTyping, setIsTyping] = useState(false)
    const bottomRef = useRef<HTMLDivElement>(null)

    useEffect(() => {
        bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
    }, [messages, isTyping])

    const handleSend = async (content: string) => {
        const userMsg: Message = {
            id: crypto.randomUUID(),
            role: 'user',
            content,
            timestamp: new Date(),
        }
        setMessages((prev) => [...prev, userMsg])
        setIsTyping(true)

        // TODO: replace with actual assistant API call
        await new Promise((res) => setTimeout(res, 1200))
        setMessages((prev) => [
            ...prev,
            {
                id: crypto.randomUUID(),
                role: 'assistant',
                content: 'Terima kasih atas pertanyaannya. Tim kami akan segera membantu Anda.',
                timestamp: new Date(),
            },
        ])
        setIsTyping(false)
    }

    return (
        <div className="flex flex-col w-85 h-120 bg-gray-50 rounded-2xl shadow-2xl overflow-hidden border border-gray-100">
            <PanelHeader
                icon={FileSearch}
                title="System Assistant"
                onClose={onClose}
                className="bg-[#6B5FAE] rounded-t-2xl"
            />
            <ScrollArea className="flex-1">
                <div className="flex flex-col gap-4 p-4">
                    {messages.map((msg) => (
                        <MessageBubble key={msg.id} message={msg} />
                    ))}
                    {isTyping && <TypingBubble />}
                    <div ref={bottomRef} />
                </div>
            </ScrollArea>
            <InputBar onSend={handleSend} disabled={isTyping} showAttachment />
        </div>
    )
}