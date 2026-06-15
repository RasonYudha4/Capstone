import { Streamdown } from 'streamdown'
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
    const isEmpty = !isUser && message.content === ''

    return (
        <div className={`flex items-end gap-2 ${isUser ? 'flex-row-reverse' : 'flex-row'}`}>
            {!isUser && <Avatar className="w-8 h-8 text-gray-400 shrink-0 mb-1" />}
            <div
                className={`max-w-[75%] px-4 py-2.5 rounded-2xl text-sm leading-relaxed ${isUser
                    ? 'bg-[#6B5FAE] text-white rounded-br-sm'
                    : 'bg-white text-gray-800 rounded-bl-sm shadow-sm'
                    }`}
            >
                {isUser ? (
                    message.content
                ) : isEmpty ? (
                    <TypingIndicator />
                ) : (
                    <Streamdown className="[&_p]:mb-2 [&_p:last-child]:mb-0 [&_strong]:font-semibold [&_strong]:text-gray-900 [&_ul]:mt-1 [&_ul]:mb-2 [&_ul]:pl-4 [&_ul]:space-y-1 [&_ol]:mt-1 [&_ol]:mb-2 [&_ol]:pl-4 [&_ol]:list-decimal [&_li]:text-gray-800 [&_code]:bg-gray-100 [&_code]:text-[#6B5FAE] [&_code]:px-1.5 [&_code]:py-0.5 [&_code]:rounded [&_code]:text-xs [&_code]:font-mono">
                        {message.content}
                    </Streamdown>
                )}
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