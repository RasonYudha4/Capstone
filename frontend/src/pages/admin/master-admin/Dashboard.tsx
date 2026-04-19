import ActivityWidget from "@/components/organism/ActivityWidget"
import DashboardHeader from "@/components/organism/DashboardHeader"
import QueueList from "@/components/organism/QueueList"
import StatRow from "@/components/organism/StatRow"

export default function DashboardPage() {
    const userName = 'Bagas'
    const pendingCount = 4

    return (
        <div className="p-6 max-w-7xl mx-auto">
            <DashboardHeader name={userName} pendingCount={pendingCount} />
            <StatRow />
            <div className="grid grid-cols-1 lg:grid-cols-[1fr_340px] gap-4">
                <QueueList />
                <ActivityWidget />
            </div>
        </div>
    )
}