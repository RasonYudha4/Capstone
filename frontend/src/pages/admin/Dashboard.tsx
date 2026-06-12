import DashboardHeader from "@/components/organism/DashboardHeader"
import { useAuth } from "@/cores/AuthContext"

export default function DashboardPage() {
    const { user } = useAuth()

    return (
        <div className="p-6 max-w-7xl mx-auto">
            <DashboardHeader name={user?.email ?? ''} pendingCount={4} />
        </div>
    )
}