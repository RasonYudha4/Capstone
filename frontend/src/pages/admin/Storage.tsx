import { useState } from "react"
import FileStatsSection from "@/components/organism/FileStatsSection"
import FileTableSection from "@/components/organism/FileTableSection"
import GroupStatsSection from "@/components/organism/GroupStatsSection"
import UploadModal from "@/components/organism/UploadModal"

export default function Storage() {
    const [uploadOpen, setUploadOpen] = useState(false)

    return (
        <div className="grid grid-cols-1 gap-6 max-w-7xl mx-auto">
            <GroupStatsSection />
            <FileStatsSection />
            <FileTableSection onUploadClick={() => setUploadOpen(true)} />

            <UploadModal
                open={uploadOpen}
                onOpenChange={setUploadOpen}
            />
        </div>
    )
}