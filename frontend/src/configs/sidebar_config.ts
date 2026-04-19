import { LayoutDashboard, Users, Settings, ShieldCheck, UserCog } from 'lucide-react'

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
    dashboard: LayoutDashboard,
    users: Users,
    settings: Settings,
    shield: ShieldCheck,
    userCog: UserCog,
} as const

const generalSection: NavSection = {
    title: 'General',
    items: [
        { href: '/dashboard', label: 'Dashboard', icon: 'dashboard' },
    ],
}

const adminSection: NavSection = {
    title: 'Admin Tools',
    roles: ['admin', 'master-admin'],
    items: [
        // { href: '/admin/users', label: 'Users', icon: 'users' },
        // { href: '/admin/settings', label: 'Settings', icon: 'settings' },
    ],
}

const masterAdminSection: NavSection = {
    title: 'Master Controls',
    roles: ['master-admin'],
    items: [
        // { href: '/master-admin/roles', label: 'Roles', icon: 'shield' },
        // { href: '/master-admin/admins', label: 'Admins', icon: 'userCog' },
    ],
}

export const sidebarSections: NavSection[] = [
    generalSection,
    adminSection,
    masterAdminSection,
]