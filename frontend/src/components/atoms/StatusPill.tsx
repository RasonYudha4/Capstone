export type FileStatus = 'pending' | 'approved' | 'rejected'

type StatusConfig = {
  label: string
  className: string
}

// Central config
const statusConfig: Record<FileStatus, StatusConfig> = {
  pending: {
    label: 'Pending',
    className: 'bg-gray-100 text-yellow-700',
  },
  approved: {
    label: 'Approved',
    className: 'bg-gray-100 text-green-700',
  },
  rejected: {
    label: 'Rejected',
    className: 'bg-gray-100 text-red-700',
  },
}

interface StatusPillProps {
  status?: string | null
  className?: string
}

export default function StatusPill({ status, className = '' }: StatusPillProps) {
  // ─────────────────────────────────────────────
  // Normalize incoming status
  // ─────────────────────────────────────────────
  const normalizedStatus = typeof status === 'string'
    ? status.trim().toLowerCase()
    : ''

  const config = statusConfig[normalizedStatus as FileStatus]

  // ─────────────────────────────────────────────
  // Fallback (prevents crash)
  // ─────────────────────────────────────────────
  if (!config) {
    return (
      <span
        className={`px-3 py-1  text-xs font-medium bg-gray-200 text-gray-600 ${className}`}
      >
        Unknown
      </span>
    )
  }

  // ─────────────────────────────────────────────
  // Normal render
  // ─────────────────────────────────────────────
  return (
    <span
      className={`px-3 py-1  text-xs font-medium ${config.className} ${className}`}
    >
      {config.label}
    </span>
  )
}