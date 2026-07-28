import type { Group } from '@/dtos/group_dto'
import type { UserListItem } from '@/dtos/login_dto'

export type RoleAssignment = {
    role: 'admin' | 'staff'
    groupId?: string
}

/** Encoded value for Select: "staff" or "admin:<group_id>" */
export function encodeRoleAssignment({ role, groupId }: RoleAssignment): string {
    if (role === 'staff') return 'staff'
    return groupId ? `admin:${groupId}` : 'admin'
}

export function decodeRoleAssignment(value: string): RoleAssignment {
    if (value === 'staff') return { role: 'staff' }
    if (value.startsWith('admin:')) {
        return { role: 'admin', groupId: value.slice(6) }
    }
    return { role: 'admin' }
}

export function dedupeGroups(groups: Group[]): Group[] {
    const seen = new Set<string>()
    const unique: Group[] = []
    for (const group of groups) {
        const key = group.group_name.trim().toLowerCase()
        if (seen.has(key)) continue
        seen.add(key)
        unique.push(group)
    }
    return unique
}

export function getUserAssignment(user: UserListItem, groups: Group[] = []): string {
    if (user.role === 'staff') return 'staff'
    if (user.role !== 'admin') return ''

    const deduped = dedupeGroups(groups)
    const match =
        deduped.find((g) => g.group_id === user.group_id) ??
        deduped.find(
            (g) =>
                user.group_name &&
                g.group_name.trim().toLowerCase() === user.group_name.trim().toLowerCase(),
        )

    if (match) return encodeRoleAssignment({ role: 'admin', groupId: match.group_id })
    if (user.group_id) return encodeRoleAssignment({ role: 'admin', groupId: user.group_id })
    return ''
}

export function assignmentsEqual(a: string, b: string): boolean {
    return a === b
}

export function groupToShortLabel(groupName: string): string {
    const lower = groupName.toLowerCase()
    if (lower.includes('manajemen')) return 'Admin Manajemen'
    if (lower.includes('pelayanan')) return 'Admin Pelayanan'
    if (lower.includes('keselamatan')) return 'Admin Keselamatan'
    if (lower.includes('nasional')) return 'Admin Program Nasional'
    return groupName
}

export function getRoleDisplayLabel(user: UserListItem): string {
    if (user.role === 'staff') return 'Staff'
    if (user.group_name) return groupToShortLabel(user.group_name)
    return 'Admin'
}

export function buildRoleOptions(groups: Group[]): Array<{ value: string; label: string }> {
    const options: Array<{ value: string; label: string }> = [
        { value: 'staff', label: 'Staff' },
    ]
    for (const group of dedupeGroups(groups)) {
        options.push({
            value: encodeRoleAssignment({ role: 'admin', groupId: group.group_id }),
            label: groupToShortLabel(group.group_name),
        })
    }
    return options
}
