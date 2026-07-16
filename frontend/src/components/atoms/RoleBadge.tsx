interface RoleBadgeProps {
    label: string
    variant?: "admin" | "staff"
}

export default function RoleBadge({ label, variant = "staff" }: RoleBadgeProps) {
    const styles =
        variant === "admin"
            ? "bg-red-50 text-red-500"
            : "bg-emerald-50 text-emerald-600"

    return (
        <span className={`inline-flex items-center px-3 py-1 rounded-full text-xs font-semibold ${styles}`}>
            {label}
        </span>
    )
}
