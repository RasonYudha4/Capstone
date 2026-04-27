import { useState } from 'react'
import { Upload } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
    Table,
    TableBody,
    TableHead,
    TableHeader,
    TableRow,
} from '@/components/ui/table'
import BreadcrumbNav, { type BreadcrumbSegment } from '../molecules/BreadcrumbNav'
import FileTableRow, { type FileRecord } from '../molecules/FileTableRow'
import UploadModal from './UploadModal'
import FileDetailModal from './FiledetailModal'

// TODO: replace with actual data from API
const mockFiles: FileRecord[] = [
    { id: '1', name: 'Kebijakan PMKP.pdf', type: 'PDF', uploadedBy: 'Supriyadi', lastUpdated: '12 Apr 2026', status: 'approved' },
    { id: '2', name: 'SOP Pelayanan.docx', type: 'DOCX', uploadedBy: 'Pardi', lastUpdated: '8 Apr 2026', status: 'review' },
    { id: '3', name: 'Laporan Audit Q1.xlsx', type: 'XLSX', uploadedBy: 'Priyadi', lastUpdated: '8 Apr 2026', status: 'pending' },
    { id: '4', name: 'Panduan Keselamatan.pdf', type: 'PDF', uploadedBy: 'Wulandari', lastUpdated: '5 Apr 2026', status: 'rejected' },
    { id: '5', name: 'Notulen Rapat.docx', type: 'DOCX', uploadedBy: 'Hartono', lastUpdated: '2 Apr 2026', status: 'approved' },
]

const kelompokOptions = [
    { label: 'Tata Kelola Rumah Sakit', value: 'tkrs' },
    { label: 'Manajemen Rumah Sakit', value: 'mrs' },
    { label: 'Pelayanan Berfokus Pasien', value: 'pbp' },
    { label: 'Sasaran Keselamatan Pasien', value: 'skp' },
    { label: 'Program Nasional', value: 'pn' },
]

const standarOptions = [
    { label: 'Standar 1', value: 's1' },
    { label: 'Standar 2', value: 's2' },
    { label: 'Standar 3', value: 's3' },
    { label: 'Standar 4', value: 's4' },
]

const elemenOptions = [
    { label: 'Elemen Penilaian 1', value: 'ep1' },
    { label: 'Elemen Penilaian 2', value: 'ep2' },
    { label: 'Elemen Penilaian 3', value: 'ep3' },
    { label: 'Elemen Penilaian 4', value: 'ep4' },
]

export default function FileTableSection() {
    const [kelompok, setKelompok] = useState('Tata Kelola Rumah Sakit')
    const [standar, setStandar] = useState('Standar 1')
    const [elemen, setElemen] = useState('Elemen Penilaian 1')
    const [uploadOpen, setUploadOpen] = useState(false)
    const [selectedFile, setSelectedFile] = useState<FileRecord | null>(null)

    const segments: BreadcrumbSegment[] = [
        { selected: kelompok, options: kelompokOptions, onChange: (v) => setKelompok(kelompokOptions.find(o => o.value === v)?.label ?? v) },
        { selected: standar, options: standarOptions, onChange: (v) => setStandar(standarOptions.find(o => o.value === v)?.label ?? v) },
        { selected: elemen, options: elemenOptions, onChange: (v) => setElemen(elemenOptions.find(o => o.value === v)?.label ?? v) },
    ]

    return (
        <section className="bg-white rounded-2xl border border-gray-100 overflow-hidden">
            <UploadModal open={uploadOpen} onOpenChange={setUploadOpen} />
            <FileDetailModal
                file={selectedFile}
                open={!!selectedFile}
                onOpenChange={(open) => { if (!open) setSelectedFile(null) }}
                onReject={(file, catatan) => console.log('Rejected', file.id, catatan)}
                onUpdate={(file, catatan) => console.log('Updated', file.id, catatan)}
            />

            {/* Toolbar */}
            <div className="flex items-center justify-between px-6 py-4 border-b border-gray-100">
                <BreadcrumbNav segments={segments} />
                <Button
                    onClick={() => setUploadOpen(true)}
                    className="bg-[#6B5FAE] hover:bg-[#5a4f9a] text-white rounded-xl text-sm gap-2 shrink-0 ml-4"
                >
                    <Upload className="w-4 h-4" />
                    Upload new file
                </Button>
            </div>

            {/* Table */}
            <div className=' px-6'>
                <div className="rounded-2xl overflow-hidden border border-gray-100">
                    <div className="overflow-x-auto">
                        <Table className='min-w-175'>
                            <TableHeader>
                                <TableRow className="bg-[#6B5FAE] hover:bg-[#6B5FAE]">
                                    {['Nama Berkas', 'Tipe Berkas', 'Di Upload oleh', 'Terakhir diperbaharui', 'Status'].map((h) => (
                                        <TableHead key={h} className="text-white text-center font-semibold text-sm px-4 py-3">
                                            {h}
                                        </TableHead>
                                    ))}
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {mockFiles.map((file) => (
                                    <FileTableRow key={file.id} file={file} onClick={() => setSelectedFile(file)} />
                                ))}
                            </TableBody>
                        </Table>
                    </div>
                </div>
            </div>
        </section>
    )
}