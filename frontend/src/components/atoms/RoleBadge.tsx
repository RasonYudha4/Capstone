interface RoleBadgeProps {
    role: "Admin" | "Staff"
}

export default function RoleBadge({ role }: RoleBadgeProps) {
    const styles =
        role === "Admin"
            ? "bg-red-50 text-red-500"
            : "bg-emerald-50 text-emerald-600"

    return (
        <span className={`inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold ${styles}`}>
            {role}
        </span>
    )
}