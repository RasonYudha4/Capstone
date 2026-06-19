import type { AgentCommand } from "@/cores/types"

export type CommandHandler = (cmd: AgentCommand) => void

export function createCommandParser(
    onText: (chunk: string) => void,
    onCommand: CommandHandler,
) {
    let buffer = ''

    return function push(chunk: string) {
        buffer += chunk

        while (true) {
            const open = buffer.indexOf('<command>')

            if (open === -1) {
                onText(buffer)
                buffer = ''
                break
            }

            if (open > 0) {
                onText(buffer.slice(0, open))
                buffer = buffer.slice(open)
                continue  
            }

            const close = buffer.indexOf('</command>')

            if (close === -1) {
                break
            }

            const jsonStr = buffer.slice('<command>'.length, close)
            buffer = buffer.slice(close + '</command>'.length)

            try {
                const cmd = JSON.parse(jsonStr) as AgentCommand
                onCommand(cmd)
            } catch (e) {
                console.log('parse failed:', e, 'jsonStr was:', JSON.stringify(jsonStr))
                onText(`<command>${jsonStr}</command>`)
            }
        }
    }
}