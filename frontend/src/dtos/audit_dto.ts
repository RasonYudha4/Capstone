import { z } from "zod";

const uuid = z.string().uuid("Invalid UUID");

export const auditResponseSchema = z.object({
    audit_id:      uuid,
    action:        z.string(),
    description:   z.string(),
    username:      z.string(),
    document_name: z.string().nullable(),   
    source:        z.string().optional(),
    created_at:    z.string().datetime({ offset: true }),
    updated_at:    z.string().datetime({ offset: true }),
});

// ✅ API returns { data: [...] } — no status field
export const getAuditResponseSchema = z.object({
    data: z.array(auditResponseSchema),
});

export type AuditResponse    = z.infer<typeof auditResponseSchema>;
export type GetAuditResponse = z.infer<typeof getAuditResponseSchema>;

// ─── UI models ────────────────────────────────────────────────────────────────

export interface Activity {
    id:           string
    timestamp:    string
    date:         string
    isoDate:      string
    isoTimestamp: string   // full ISO string for precise sorting
    timeLabel:    string
    actor:        string
    action:       string
    file:         string   // ✅ normalized fallback, never null in the UI model
    rawAction:    string
}

export interface ActivityGroup {
    date:  string
    items: Activity[]
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

const ACTION_LABEL_MAP: Record<string, string> = {
    insert:   'mengupload file',
    edit:     'mengedit file',
    update:   'mengubah status file',
    delete:   'menghapus file',
    open:     'membuka file',
    download: 'mengunduh file',
    login:    'login',
    login_fail: 'gagal login',
    lockout:  'akun terkunci',
    error:    'error',
}

const FILE_ACTIONS = new Set(['insert', 'edit', 'update', 'delete', 'open', 'download'])

/** Auth reuses action_type "delete" for logout — detect via description. */
function isLogoutAudit(action: string, description?: string | null): boolean {
    if (action.toLowerCase() !== 'delete') return false
    const d = (description ?? '').toLowerCase()
    return d.includes('logout') || d.includes('logged out')
}

function isTokenRefreshAudit(action: string, description?: string | null): boolean {
    if (action.toLowerCase() !== 'update') return false
    return (description ?? '').toLowerCase().includes('token refreshed')
}

export function mapActionLabel(action: string, description?: string | null): string {
    if (isLogoutAudit(action, description)) return 'logout'
    if (isTokenRefreshAudit(action, description)) return 'memperbarui sesi'
    return ACTION_LABEL_MAP[action.toLowerCase()] ?? action
}

export function parseTimestamp(iso: string): { date: string; timeLabel: string; timestamp: string } {
    const d = new Date(iso)
    const date = d.toLocaleDateString('id-ID', { day: '2-digit', month: 'long', year: 'numeric' })
    const timeLabel = d.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', hour12: false })
    return { date, timeLabel, timestamp: `${date} ${timeLabel}` }
}

export function mapAuditToActivity(dto: AuditResponse): Activity {
    const { date, timeLabel, timestamp } = parseTimestamp(dto.created_at)
    const action = mapActionLabel(dto.action, dto.description)
    const isFileOp =
        FILE_ACTIONS.has(dto.action.toLowerCase()) &&
        !isLogoutAudit(dto.action, dto.description) &&
        !isTokenRefreshAudit(dto.action, dto.description)

    return {
        id:        dto.audit_id,
        timestamp,
        date,
        isoDate:      new Date(dto.created_at).toISOString().slice(0, 10),
        isoTimestamp: dto.created_at,
        timeLabel,
        actor:     dto.username,
        action,
        // Non–file audits (logout, token refresh, …) have no document_name —
        // avoid the misleading fallback "-"
        file:      isFileOp ? (dto.document_name ?? '—') : '',
        rawAction: dto.action,
    }
}

export function groupActivitiesByDate(activities: Activity[]): ActivityGroup[] {
    const map = new Map<string, Activity[]>()
    for (const a of activities) {
        if (!map.has(a.date)) map.set(a.date, [])
        map.get(a.date)!.push(a)
    }
    return Array.from(map.entries())
        .map(([date, items]) => ({
            date,
            items: items.sort((a, b) => b.isoTimestamp.localeCompare(a.isoTimestamp)),
        }))
        .sort((a, b) => b.items[0].isoDate.localeCompare(a.items[0].isoDate))
}