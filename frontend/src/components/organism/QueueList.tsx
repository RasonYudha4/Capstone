import { Target } from 'lucide-react'
import SectionHeading from '../atoms/SectionHeading'
import QueueCard from '../molecules/QueueCard'

const queueItems = [
    { category: 'Peningkatan Mutu dan Keselamatan Pasien', title: 'Elemen Penilaian 2 Standar 2', submittedBy: 'Supriyadi', date: '12 April 2026 13:45' },
    { category: 'Pelayanan dan Asuhan Pasien', title: 'Elemen Penilaian 3 Standar 2', submittedBy: 'Pardi', date: '8 April 2026' },
    { category: 'Pelayanan dan Asuhan Pasien', title: 'Elemen Penilaian 3 Standar 4', submittedBy: 'Priyadi', date: '8 April 2026' },
    { category: 'Pelayanan dan Asuhan Pasien', title: 'Elemen Penilaian 2 Standar 2', submittedBy: 'Pardi', date: '8 April 2026' },
]

export default function QueueList() {
    return (
        <div className="bg-white shadow-xl rounded-2xl border border-gray-100 p-5 h-full">
            <SectionHeading icon={Target} title="Daftar Antrian" />
            <div className="flex flex-col gap-3">
                {queueItems.map((item, i) => <QueueCard key={i} {...item} />)}
            </div>
        </div>
    )
}