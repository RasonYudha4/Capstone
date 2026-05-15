import { z } from "zod";

const uuid = z.string().uuid("Invalid UUID");

export const auditResponseSchema = z.object({
    audit_id:      uuid,
    action:        z.string(),
    description:   z.string(),
    username:      z.string(),
    document_name: z.string(),
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
    id:        string
    timestamp: string
    date:      string
    isoDate:      string
    isoTimestamp: string   // full ISO string for precise sorting
    timeLabel:    string
    actor:     string
    action:    string
    file:      string
    rawAction: string
}

export interface ActivityGroup {
    date:  string
    items: Activity[]
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

const ACTION_LABEL_MAP: Record<string, string> = {
    insert:   'mengupload file',
    edit:     'mengedit file',
    update:   'mengubah Status file',
    delete:   'menghapus file',
    download: 'mengunduh file',
}

export function mapActionLabel(action: string): string {
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
    return {
        id:        dto.audit_id,
        timestamp,
        date,
        isoDate:      new Date(dto.created_at).toISOString().slice(0, 10),
        isoTimestamp: dto.created_at,
        timeLabel,
        actor:     dto.username,
        action:    mapActionLabel(dto.action),
        file:      dto.document_name,
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