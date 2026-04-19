import FileStatsSection from "@/components/organism/FileStatsSection";
import FileTableSection from "@/components/organism/FileTableSection";
import GroupStatsSection from "@/components/organism/GroupStatsSection";

export default function Storage() {
    return (
        <div className="grid grid-cols-1 gap-6 max-w-7xl mx-auto">
            <GroupStatsSection />
            <FileStatsSection />
            <FileTableSection />
        </div>
    )
}