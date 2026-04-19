import Avatar from '../atoms/Avatar'
import TypingIndicator from '../atoms/TypingIndicator'

export type MessageRole = 'user' | 'assistant'

export interface Message {
    id: string
    role: MessageRole
    content: string
    timestamp: Date
}

interface MessageBubbleProps {
    message: Message
}

export default function MessageBubble({ message }: MessageBubbleProps) {
    const isUser = message.role === 'user'

    return (
        <div className={`flex items-end gap-2 ${isUser ? 'flex-row-reverse' : 'flex-row'}`}>
            {!isUser && <Avatar className="w-8 h-8 text-gray-400 shrink-0 mb-1" />}
            <div
                className={`max-w-[75%] px-4 py-2.5 rounded-2xl text-sm leading-relaxed ${isUser
                    ? 'bg-[#6B5FAE] text-white rounded-br-sm'
                    : 'bg-white text-gray-800 rounded-bl-sm shadow-sm'
                    }`}
            >
                {message.content}
            </div>
        </div>
    )
}

export function TypingBubble() {
    return (
        <div className="flex items-end gap-2">
            <Avatar className="w-8 h-8 text-gray-400 shrink-0 mb-1" />
            <div className="bg-white rounded-2xl rounded-bl-sm shadow-sm">
                <TypingIndicator />
            </div>
        </div>
    )
}