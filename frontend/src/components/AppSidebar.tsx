import { NavLink } from 'react-router'
import { iconMap, sidebarSections } from '../configs/sidebar_config'
import { useAuth } from '../cores/AuthContext'

export default function AppSidebar() {
    const { user } = useAuth()
    const role = user?.role ?? 'staff'

    const filteredSections = sidebarSections.filter(
        (section) => !section.roles || section.roles.includes(role)
    )

    return (
        <aside className="w-64 h-screen bg-white border-r flex flex-col">
            <div className="p-6 border-b">
                <h1 className="font-bold text-lg">MyApp</h1>
            </div>

            <nav className="flex-1 overflow-y-auto p-4 space-y-6">
                {filteredSections.map((section) => (
                    <div key={section.title ?? 'default'}>
                        {section.title && (
                            <p className="text-xs font-bold uppercase tracking-wide text-gray-400 mb-2">
                                {section.title}
                            </p>
                        )}
                        <div className="flex flex-col gap-1">
                            {section.items.map((item) => {
                                const Icon = iconMap[item.icon]
                                return (
                                    <NavLink
                                        key={item.href}
                                        to={item.href}
                                        className={({ isActive }) =>
                                            `flex items-center gap-3 px-3 py-2 rounded-md text-sm transition-colors ${isActive
                                                ? 'bg-primary text-primary-foreground font-medium'
                                                : 'text-gray-600 hover:bg-gray-100'
                                            }`
                                        }
                                    >
                                        <Icon className="w-4 h-4" />
                                        {item.label}
                                    </NavLink>
                                )
                            })}
                        </div>
                    </div>
                ))}
            </nav>

            <div className="p-4 border-t text-sm text-gray-500">
                {user?.name} · <span className="capitalize">{role}</span>
            </div>
        </aside>
    )
}