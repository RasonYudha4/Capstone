import ActivityWidget from "@/components/organism/ActivityWidget"
import DashboardHeader from "@/components/organism/DashboardHeader"
import QueueList from "@/components/organism/QueueList"
import StatRow from "@/components/organism/StatRow"
import { useAuth } from "@/cores/AuthContext"
import { useStats } from "@/hooks/useDocument"
import { useIngestKmk } from "@/hooks/useIngest"
import { AlertCircle, CheckCircle, Database, Loader2 } from "lucide-react"

function IngestKmkButton() {
    const { mutate, isPending, isSuccess, isError } = useIngestKmk()

    return (
        <button
            onClick={() => mutate()}
            disabled={isPending}
            title="Ingest KMK Standard"
            className="flex items-center gap-2 px-3 py-2 rounded-xl bg-[#6B5FAE] text-white text-sm font-medium hover:bg-[#5a4f9a] disabled:opacity-60 disabled:cursor-not-allowed transition-colors"
        >
            {isPending ? (
                <Loader2 className="w-4 h-4 animate-spin" />
            ) : isSuccess ? (
                <CheckCircle className="w-4 h-4" />
            ) : isError ? (
                <AlertCircle className="w-4 h-4" />
            ) : (
                <Database className="w-4 h-4" />
            )}
            {isPending ? 'Ingesting...' : 'Ingest KMK'}
        </button>
    )
}

export default function DashboardPage() {
    const { user } = useAuth()
    const { data: stats } = useStats()
    const pendingCount = stats?.stats?.pending ?? 0

    return (
        <div className="p-6 max-w-7xl mx-auto">
            <DashboardHeader
                name={user?.email ?? ''}
                pendingCount={pendingCount}
                actions={<IngestKmkButton />}
            />
            <StatRow />
            <div className="grid grid-cols-1 lg:grid-cols-[1fr_340px] gap-4">
                <QueueList />
                <ActivityWidget />
            </div>
        </div>
    )
}