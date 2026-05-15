import { Home, Users, Folders, History } from 'lucide-react'

export type IconKey = keyof typeof iconMap

export interface NavItem {
    href: string
    label: string
    icon: IconKey
}

export interface NavSection {
    title?: string
    items: NavItem[]
    roles?: string[] 
}

export const iconMap = {
    dashboard: Home,
    folders: Folders,
    users: Users,
    activity: History
} as const

const generalSection: NavSection = {
    title: 'General',
    items: [
        { href: '/dashboard', label: 'Dashboard', icon: 'dashboard' },
    ],
}

const adminSection: NavSection = {
    title: 'Admin Tools',
    // roles: ['admin', 'master-admin'],
    items: [
        { href: '/storage', label: "Storage", icon: 'folders'}
    ],
}

const masterAdminSection: NavSection = {
    title: 'Master Controls',
    // roles: ['master-admin'],
    items: [
        { href: '/admins', label: 'Admins', icon: 'users' },
        { href: '/activity-log', label: 'Activity Log', icon: 'activity'}
    ],
}

export const sidebarSections: NavSection[] = [
    generalSection,
    adminSection,
    masterAdminSection,
]