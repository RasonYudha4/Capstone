import { NavLink, useNavigate } from 'react-router'
import { iconMap, sidebarSections } from '../../configs/sidebar_config'
import { useAuth } from '../../cores/AuthContext'
import { LogOut } from 'lucide-react'
import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarGroup,
    SidebarGroupLabel,
    SidebarHeader,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
} from '@/components/ui/sidebar'

export default function AppSidebar() {
    const { user, logout } = useAuth()
    const navigate = useNavigate()
    const role = user?.role ?? 'staff'

    const filteredSections = sidebarSections.filter(
        (section) => !section.roles || section.roles.includes(role)
    )

    return (
        <Sidebar className="border-r border-gray-100 bg-white shadow-xl">
            {/* Logo */}
            <SidebarHeader className="px-4 py-5">
                <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-xl bg-[#6B5FAE] shrink-0 flex items-center justify-center">
                        <img src='/logo.png' height={35} width={35} />
                    </div>
                    <span className="font-bold text-[#6B5FAE] text-base leading-tight">
                        Document<br />Approval
                    </span>
                </div>
            </SidebarHeader>

            {/* Nav */}
            <SidebarContent className="px-3 py-2">
                {filteredSections.map((section) => (
                    <SidebarGroup key={section.title ?? 'default'}>
                        {section.title && (
                            <SidebarGroupLabel className="text-xs font-semibold uppercase tracking-wide text-gray-400 mb-1 px-2">
                                {section.title}
                            </SidebarGroupLabel>
                        )}
                        <SidebarMenu>
                            {section.items.map((item) => {
                                const Icon = iconMap[item.icon]
                                return (
                                    <SidebarMenuItem key={item.href}>
                                        <NavLink to={item.href}>
                                            {({ isActive }) => (
                                                <SidebarMenuButton
                                                    isActive={isActive}
                                                    className={`
                                                        w-full flex items-center gap-3 px-4 py-2.5 rounded-full text-sm font-medium transition-all
                                                        ${isActive
                                                            ? 'bg-[#6B5FAE] text-white hover:bg-[#6B5FAE] hover:text-white'
                                                            : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
                                                        }
                                                    `}
                                                >
                                                    <Icon className="w-4 h-4 shrink-0" />
                                                    <span>{item.label}</span>
                                                </SidebarMenuButton>
                                            )}
                                        </NavLink>
                                    </SidebarMenuItem>
                                )
                            })}
                        </SidebarMenu>
                    </SidebarGroup>
                ))}
            </SidebarContent>

            {/* Footer / Logout */}
            <SidebarFooter className="px-3 py-4 mt-auto">
                <SidebarMenu>
                    <SidebarMenuItem>
                        <SidebarMenuButton
                            onClick={async () => {
                                await logout()
                                navigate('/login', { replace: true })
                            }}
                            className="w-full flex items-center gap-3 px-4 py-2.5 rounded-full text-sm font-medium text-gray-600 hover:bg-gray-100 hover:text-gray-900 transition-all cursor-pointer"
                        >
                            <LogOut className="w-4 h-4 shrink-0" />
                            <span>Logout</span>
                        </SidebarMenuButton>
                    </SidebarMenuItem>
                </SidebarMenu>
            </SidebarFooter>
        </Sidebar>
    )
}