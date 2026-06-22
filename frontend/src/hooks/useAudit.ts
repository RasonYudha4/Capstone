import { useQuery } from '@tanstack/react-query'
import { auditService } from '@/services/audit_service'
import { mapAuditToActivity, groupActivitiesByDate } from '@/dtos/audit_dto'
import type { ActivityGroup } from '@/dtos/audit_dto'

export const auditKeys = {
    all:   ['audit'] as const,
    lists: () => [...auditKeys.all, 'list'] as const,
}

export const useAudit = () => {
    return useQuery({
        queryKey: auditKeys.lists(),
        queryFn:  () => auditService.getAll(),
        staleTime: 1000 * 60 * 5,
    })
}

export const useActivityLog = (limit?: number) => {
    const query = useAudit()

    const groups: ActivityGroup[] = query.data?.data
        ? groupActivitiesByDate(
            query.data.data
                .map(mapAuditToActivity)
                .sort((a, b) => b.isoTimestamp.localeCompare(a.isoTimestamp))
                .slice(0, limit)
          )
        : []

    return { ...query, groups }
}