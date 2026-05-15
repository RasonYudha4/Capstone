import ActivityWidget from "@/components/organism/ActivityWidget"
import DashboardHeader from "@/components/organism/DashboardHeader"
import QueueList from "@/components/organism/QueueList"
import StatRow from "@/components/organism/StatRow"
import { useAuth } from "@/cores/AuthContext"

export default function DashboardPage() {
    const { user } = useAuth()

    return (
        <div className="p-6 max-w-7xl mx-auto">
            <DashboardHeader name={user?.name ?? ''} pendingCount={4} />
        </div>
    )
}