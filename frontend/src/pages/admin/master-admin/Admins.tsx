// app/.../Admins.tsx
"use client"

import { useState } from "react"
import SectionHeading from "@/components/atoms/SectionHeading"
import StatCard from "@/components/molecules/StatCard"
import RoleBadge from "@/components/atoms/RoleBadge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select"
import {
    DropdownMenu,
    DropdownMenuCheckboxItem,
    DropdownMenuContent,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
    UserPenIcon,
    Users,
    User,
    Layers,
    Search,
    SlidersHorizontal,
    ChevronLeft,
    ChevronRight,
} from "lucide-react"

type Role = "Admin" | "Staff"

type Department =
    | "Manajemen Rumah Sakit"
    | "Pelayanan Pasien"
    | "Program Nasional"
    | "Keselamatan Pasien"

const DEPARTMENT_OPTIONS: Department[] = [
    "Manajemen Rumah Sakit",
    "Pelayanan Pasien",
    "Program Nasional",
    "Keselamatan Pasien",
]

interface UserRow {
    id: number
    name: string
    initials: string
    email: string
    joinDate: string
    currentRole: Role
    department: Department
    avatarBg: string
    avatarText: string
}

const AVATAR_STYLES = [
    { avatarBg: "bg-violet-100", avatarText: "text-violet-600" },
    { avatarBg: "bg-emerald-100", avatarText: "text-emerald-600" },
    { avatarBg: "bg-rose-100", avatarText: "text-rose-500" },
    { avatarBg: "bg-sky-100", avatarText: "text-sky-600" },
]

const USERS: UserRow[] = [
    { id: 1, name: "Budi Santoso", initials: "BS", email: "budi.santoso@gmail.com", joinDate: "Bergabung 12 Jan 2025", currentRole: "Admin", department: "Manajemen Rumah Sakit", ...AVATAR_STYLES[0] },
    { id: 2, name: "Sari Dewi", initials: "SD", email: "sari.dewi@gmail.com", joinDate: "Bergabung 3 Mar 2025", currentRole: "Staff", department: "Pelayanan Pasien", ...AVATAR_STYLES[1] },
    { id: 3, name: "Rina Kusuma", initials: "RK", email: "rina.kusuma@gmail.com", joinDate: "Bergabung 21 Feb 2025", currentRole: "Admin", department: "Program Nasional", ...AVATAR_STYLES[2] },
    { id: 4, name: "Agus Prabowo", initials: "AP", email: "agus.prabowo@gmail.com", joinDate: "Bergabung 7 Apr 2025", currentRole: "Staff", department: "Keselamatan Pasien", ...AVATAR_STYLES[3] },
    { id: 5, name: "Hendra Wijaya", initials: "HW", email: "hendra.wijaya@gmail.com", joinDate: "Bergabung 30 Jan 2025", currentRole: "Admin", department: "Manajemen Rumah Sakit", ...AVATAR_STYLES[2] },
    { id: 6, name: "Dian Pratama", initials: "DP", email: "dian.pratama@gmail.com", joinDate: "Bergabung 18 May 2025", currentRole: "Staff", department: "Pelayanan Pasien", ...AVATAR_STYLES[2] },
    { id: 7, name: "Maya Nugroho", initials: "MN", email: "maya.nugroho@gmail.com", joinDate: "Bergabung 2 Jun 2025", currentRole: "Staff", department: "Program Nasional", ...AVATAR_STYLES[3] },
]

export default function Admins() {
    const [pendingDepartment, setPendingDepartment] = useState<Record<number, Department>>(
        USERS.reduce((acc, u) => ({ ...acc, [u.id]: u.department }), {} as Record<number, Department>)
    )
    const [departmentFilter, setDepartmentFilter] = useState<Department[]>([])
    const [page, setPage] = useState(1)
    const totalPages = 3

    const handleDepartmentChange = (id: number, department: Department) => {
        setPendingDepartment((prev) => ({ ...prev, [id]: department }))
    }

    const toggleDepartmentFilter = (dept: Department) => {
        setDepartmentFilter((prev) =>
            prev.includes(dept) ? prev.filter((d) => d !== dept) : [...prev, dept]
        )
    }

    const filteredUsers =
        departmentFilter.length === 0
            ? USERS
            : USERS.filter((user) => departmentFilter.includes(user.department))

    return (
        <section>
            <SectionHeading icon={UserPenIcon} title="Manajemen Admin" />
            <p className="text-gray-500 text-sm mt-1">
                Master Admin dapat mengubah peran pengguna (upgrade Staff → Admin atau downgrade Admin → Staff).
            </p>

            <div className="mt-8 grid grid-cols-1 md:grid-cols-3 gap-4">
                <StatCard label="Total Pengguna" value={24} weeklyCount={3} weeklyLabel="minggu ini" icon={User} />

                <StatCard label="Total Admin" icon={Layers}>
                    <div className="flex flex-col gap-1">
                        <span className="text-3xl font-bold text-gray-900">4</span>
                        <span className="inline-flex items-center w-fit px-2.5 py-0.5 rounded-full text-xs font-semibold bg-amber-50 text-amber-600">
                            Hak akses penuh
                        </span>
                    </div>
                </StatCard>

                <StatCard label="Total Staff" icon={Users}>
                    <div className="flex flex-col gap-1">
                        <span className="text-3xl font-bold text-gray-900">20</span>
                        <span className="inline-flex items-center w-fit px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-600">
                            Akses standar
                        </span>
                    </div>
                </StatCard>
            </div>

            <div className="bg-white rounded-2xl border border-gray-100 p-6 min-h-172 mt-8">
                <div className="flex items-center justify-between flex-wrap gap-4">
                    <div className="flex items-center gap-3">
                        <Users className="w-5 h-5 text-[#6B5FAE]" />
                        <p className="font-semibold text-gray-900">Semua Pengguna</p>
                    </div>
                    <div className="flex items-center gap-3">
                        <div className="relative">
                            <Search className="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" />
                            <Input
                                placeholder="Cari pengguna..."
                                className="pl-9 w-64 bg-gray-50 border-gray-100 rounded-xl"
                            />
                        </div>

                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button
                                    variant="outline"
                                    className="rounded-xl border-gray-100 bg-gray-50 text-gray-600 font-normal gap-2"
                                >
                                    <SlidersHorizontal className="w-4 h-4" />
                                    Filter
                                    {departmentFilter.length > 0 && (
                                        <span className="ml-1 w-5 h-5 rounded-full bg-[#6B5FAE] text-white text-xs flex items-center justify-center">
                                            {departmentFilter.length}
                                        </span>
                                    )}
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end" className="w-56">
                                <DropdownMenuLabel>Filter berdasarkan departemen</DropdownMenuLabel>
                                <DropdownMenuSeparator />
                                {DEPARTMENT_OPTIONS.map((dept) => (
                                    <DropdownMenuCheckboxItem
                                        key={dept}
                                        checked={departmentFilter.includes(dept)}
                                        onCheckedChange={() => toggleDepartmentFilter(dept)}
                                        onSelect={(e) => e.preventDefault()}
                                    >
                                        {dept}
                                    </DropdownMenuCheckboxItem>
                                ))}
                                {departmentFilter.length > 0 && (
                                    <>
                                        <DropdownMenuSeparator />
                                        <button
                                            onClick={() => setDepartmentFilter([])}
                                            className="w-full text-left px-2 py-1.5 text-sm text-[#6B5FAE] hover:bg-gray-50 rounded-md"
                                        >
                                            Hapus semua filter
                                        </button>
                                    </>
                                )}
                            </DropdownMenuContent>
                        </DropdownMenu>
                    </div>
                </div>

                <div className="mt-6 rounded-xl overflow-hidden border border-gray-100">
                    <table className="w-full text-sm">
                        <thead>
                            <tr className="bg-[#6B5FAE] text-white text-left">
                                <th className="px-6 py-4 font-semibold w-14">No.</th>
                                <th className="px-6 py-4 font-semibold">Nama Lengkap</th>
                                <th className="px-6 py-4 font-semibold">Email / UPN</th>
                                <th className="px-6 py-4 font-semibold">Role Saat Ini</th>
                                <th className="px-6 py-4 font-semibold">Ubah Role</th>
                                <th className="px-6 py-4 font-semibold">Aksi</th>
                            </tr>
                        </thead>
                        <tbody>
                            {filteredUsers.length === 0 ? (
                                <tr>
                                    <td colSpan={6} className="px-6 py-10 text-center text-gray-400">
                                        Tidak ada pengguna yang cocok dengan filter ini.
                                    </td>
                                </tr>
                            ) : (
                                filteredUsers.map((user) => {
                                    const selectedDept = pendingDepartment[user.id]

                                    return (
                                        <tr key={user.id} className="border-t border-gray-100">
                                            <td className="px-6 py-4 text-gray-500">{user.id}</td>
                                            <td className="px-6 py-4">
                                                <div className="flex items-center gap-3">
                                                    <div
                                                        className={`w-9 h-9 rounded-full flex items-center justify-center text-xs font-semibold ${user.avatarBg} ${user.avatarText}`}
                                                    >
                                                        {user.initials}
                                                    </div>
                                                    <div>
                                                        <p className="font-medium text-gray-900">{user.name}</p>
                                                        <p className="text-xs text-gray-400">{user.joinDate}</p>
                                                    </div>
                                                </div>
                                            </td>
                                            <td className="px-6 py-4 text-gray-600">{user.email}</td>
                                            <td className="px-6 py-4">
                                                <RoleBadge role={user.currentRole} />
                                            </td>
                                            <td className="px-6 py-4">
                                                <div className="flex flex-col gap-1.5 w-40">
                                                    <Select
                                                        value={selectedDept}
                                                        onValueChange={(value: Department) => handleDepartmentChange(user.id, value)}
                                                    >
                                                        <SelectTrigger
                                                            className="rounded-xl bg-white w-44"
                                                            title={selectedDept}
                                                        >
                                                            <SelectValue className="block truncate text-left" />
                                                        </SelectTrigger>
                                                        <SelectContent>
                                                            {DEPARTMENT_OPTIONS.map((dept) => (
                                                                <SelectItem key={dept} value={dept}>
                                                                    {dept}
                                                                </SelectItem>
                                                            ))}
                                                        </SelectContent>
                                                    </Select>
                                                </div>
                                            </td>
                                            <td className="px-6 py-4">
                                                <Button className="bg-[#6B5FAE] hover:bg-[#5b4f97] rounded-xl px-5 text-white">
                                                    Simpan
                                                </Button>
                                            </td>
                                        </tr>
                                    )
                                })
                            )}
                        </tbody>
                    </table>
                </div>

                <div className="flex items-center justify-between mt-6 flex-wrap gap-3">
                    <p className="text-sm text-gray-400">
                        Halaman {page} · Menampilkan {filteredUsers.length} dari 24 pengguna
                    </p>
                    <div className="flex items-center gap-2">
                        <button
                            onClick={() => setPage((p) => Math.max(1, p - 1))}
                            className="w-9 h-9 rounded-xl border border-gray-100 flex items-center justify-center text-gray-400 disabled:opacity-50"
                            disabled={page === 1}
                        >
                            <ChevronLeft className="w-4 h-4" />
                        </button>
                        {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
                            <button
                                key={p}
                                onClick={() => setPage(p)}
                                className={`w-9 h-9 rounded-xl flex items-center justify-center text-sm font-medium ${p === page
                                    ? "bg-[#6B5FAE] text-white"
                                    : "border border-gray-100 text-gray-500"
                                    }`}
                            >
                                {p}
                            </button>
                        ))}
                        <button
                            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                            className="w-9 h-9 rounded-xl border border-gray-100 flex items-center justify-center text-gray-400 disabled:opacity-50"
                            disabled={page === totalPages}
                        >
                            <ChevronRight className="w-4 h-4" />
                        </button>
                    </div>
                </div>
            </div>
        </section>
    )
}