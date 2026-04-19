import StatusPill, { type FileStatus } from '../atoms/StatusPill'

export interface FileRecord {
    id: string
    name: string
    type: string
    uploadedBy: string
    lastUpdated: string
    status: FileStatus
}

interface FileTableRowProps {
    file: FileRecord
}

export default function FileTableRow({ file }: FileTableRowProps) {
    return (
        <tr className="border-b border-gray-100 hover:bg-gray-50 transition-colors">
            <td className="px-4 py-3 text-sm text-gray-800 font-medium">{file.name}</td>
            <td className="px-4 py-3 text-sm text-gray-500">{file.type}</td>
            <td className="px-4 py-3 text-sm text-gray-500">{file.uploadedBy}</td>
            <td className="px-4 py-3 text-sm text-gray-500">{file.lastUpdated}</td>
            <td className="px-4 py-3">
                <StatusPill status={file.status} />
            </td>
        </tr>
    )
}