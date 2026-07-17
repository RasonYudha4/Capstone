"use client"

import { useState, useEffect, useMemo } from "react"
import SectionHeading from "@/components/atoms/SectionHeading"
import StatCard from "@/components/molecules/StatCard"
import RoleBadge from "@/components/atoms/RoleBadge"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Separator } from "@/components/ui/separator"
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogDescription,
} from "@/components/ui/dialog"
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
    UserPlus,
    Mail,
    Send,
    CheckCircle2,
    X,
    Trash2,
    ShieldOff,
    ShieldCheck,
    RefreshCw,
    AlertCircle,
} from "lucide-react"
import { authService } from "@/services/auth_service"
import { groupService } from "@/services/group_service"
import { type UserListItem } from "@/dtos/login_dto"
import { type Group } from "@/dtos/group_dto"
import {
    buildRoleOptions,
    decodeRoleAssignment,
    dedupeGroups,
    getRoleDisplayLabel,
    getUserAssignment,
} from "@/utils/role_assignment"
import { toast } from "sonner"

type RoleFilter = "admin" | "staff" | "invited"

const fieldClass = "rounded-xl border-gray-200 bg-gray-50 focus:ring-[#6B5FAE] focus:border-[#6B5FAE] placeholder:text-gray-400 text-sm"
const labelClass = "text-sm font-bold text-[#6B5FAE]"

const AVATAR_COLORS = [
    { bg: "bg-violet-100", text: "text-violet-600" },
    { bg: "bg-emerald-100", text: "text-emerald-600" },
    { bg: "bg-rose-100", text: "text-rose-500" },
    { bg: "bg-sky-100", text: "text-sky-600" },
    { bg: "bg-amber-100", text: "text-amber-600" },
    { bg: "bg-fuchsia-100", text: "text-fuchsia-600" },
]

function getAvatarColor(email: string) {
    const idx = email.charCodeAt(0) % AVATAR_COLORS.length
    return AVATAR_COLORS[idx]
}

function getInitials(email: string) {
    const parts = email.split("@")[0].split(/[._-]/)
    return parts
        .slice(0, 2)
        .map((p) => p[0]?.toUpperCase() ?? "")
        .join("")
}

const PAGE_SIZE = 10

export default function Admins() {
    // ── data state ──
    const [users, setUsers] = useState<UserListItem[]>([])
    const [groups, setGroups] = useState<Group[]>([])
    const [isLoading, setIsLoading] = useState(true)
    const [loadError, setLoadError] = useState<string | null>(null)

    // ── pending role changes (userId → encoded assignment) ──
    const [pendingAssignments, setPendingAssignments] = useState<Record<string, string>>({})
    const [savingId, setSavingId] = useState<string | null>(null)

    // ── search & filter ──
    const [search, setSearch] = useState("")
    const [roleFilter, setRoleFilter] = useState<RoleFilter[]>([])
    const [page, setPage] = useState(1)

    // ── invite modal ──
    const [inviteOpen, setInviteOpen] = useState(false)
    const [inviteEmail, setInviteEmail] = useState("")
    const [inviteRole, setInviteRole] = useState<string>("")
    const [inviteSuccess, setInviteSuccess] = useState(false)
    const [isSubmitting, setIsSubmitting] = useState(false)

    // ── delete confirm modal ──
    const [deleteTarget, setDeleteTarget] = useState<UserListItem | null>(null)
    const [isDeletingId, setIsDeletingId] = useState<string | null>(null)
    const [isResendingId, setIsResendingId] = useState<string | null>(null)

    // ── suspend confirm modal ──
    const [suspendTarget, setSuspendTarget] = useState<UserListItem | null>(null)
    const [isSuspendingId, setIsSuspendingId] = useState<string | null>(null)

    // ─────────────────────────────────────────────
    // Load users from API
    // ─────────────────────────────────────────────
    const roleOptions = useMemo(() => buildRoleOptions(groups), [groups])

    const fetchUsers = async () => {
        setIsLoading(true)
        setLoadError(null)
        try {
            const [usersRes, groupsRes] = await Promise.all([
                authService.listUsers(),
                groupService.listGroups(),
            ])
            setUsers(usersRes.users)
            setGroups(groupsRes)
            const uniqueGroups = dedupeGroups(groupsRes)
            const initial: Record<string, string> = {}
            usersRes.users.forEach((u) => {
                if (u.role === "admin" || u.role === "staff") {
                    initial[u.user_id] = getUserAssignment(u, uniqueGroups)
                }
            })
            setPendingAssignments(initial)
        } catch (err: any) {
            setLoadError(err.message || "Gagal memuat daftar pengguna.")
        } finally {
            setIsLoading(false)
        }
    }

    useEffect(() => {
        fetchUsers()
    }, [])

    // ─────────────────────────────────────────────
    // Derived stats
    // ─────────────────────────────────────────────
    const activeUsers = users.filter((u) => u.account_status === "active")
    const totalAdmin = activeUsers.filter((u) => u.role === "admin").length
    const totalStaff = activeUsers.filter((u) => u.role === "staff").length

    // ─────────────────────────────────────────────
    // Filtered + paginated list
    // ─────────────────────────────────────────────
    const filteredUsers = useMemo(() => {
        let list = users
        if (search.trim()) {
            const q = search.toLowerCase()
            list = list.filter((u) => u.email.toLowerCase().includes(q))
        }
        if (roleFilter.length > 0) {
            list = list.filter((u) => {
                if (roleFilter.includes("invited")) {
                    if (u.account_status === "invited") return true
                }
                if (roleFilter.includes("admin") && u.role === "admin" && u.account_status !== "invited") return true
                if (roleFilter.includes("staff") && u.role === "staff" && u.account_status !== "invited") return true
                return false
            })
        }
        return list
    }, [users, search, roleFilter])

    const totalPages = Math.max(1, Math.ceil(filteredUsers.length / PAGE_SIZE))
    const pagedUsers = filteredUsers.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE)

    // Reset page when filter/search changes
    useEffect(() => { setPage(1) }, [search, roleFilter])

    // ─────────────────────────────────────────────
    // Save role change
    // ─────────────────────────────────────────────
    const handleSaveRole = async (user: UserListItem) => {
        const pendingValue = pendingAssignments[user.user_id]
        const currentValue = getUserAssignment(user, groups)
        if (!pendingValue || pendingValue === currentValue) {
            toast.info("Role tidak berubah.")
            return
        }

        const assignment = decodeRoleAssignment(pendingValue)
        if (assignment.role === "admin" && !assignment.groupId) {
            toast.error("Pilih kelompok admin terlebih dahulu.")
            return
        }

        setSavingId(user.user_id)
        try {
            await authService.updateUserRole(user.user_id, assignment.role, assignment.groupId)
            const selectedGroup = groups.find((g) => g.group_id === assignment.groupId)
            toast.success(
                `Role ${user.email} berhasil diubah ke '${getRoleDisplayLabel({
                    ...user,
                    role: assignment.role,
                    group_id: assignment.groupId ?? null,
                    group_name: selectedGroup?.group_name ?? null,
                })}'.`
            )
            setUsers((prev) =>
                prev.map((u) =>
                    u.user_id === user.user_id
                        ? {
                            ...u,
                            role: assignment.role,
                            group_id: assignment.role === "staff" ? null : assignment.groupId ?? null,
                            group_name: assignment.role === "staff" ? null : selectedGroup?.group_name ?? null,
                        }
                        : u
                )
            )
        } catch (err: any) {
            toast.error(err.message || "Gagal mengubah role.")
        } finally {
            setSavingId(null)
        }
    }

    // ─────────────────────────────────────────────
    // Suspend / activate
    // ─────────────────────────────────────────────
    const handleToggleStatus = async (user: UserListItem) => {
        if (user.account_status === "active") {
            // Confirm before suspending
            setSuspendTarget(user)
        } else {
            await doUpdateStatus(user, "active")
        }
    }

    const doUpdateStatus = async (user: UserListItem, status: "active" | "suspended") => {
        setIsSuspendingId(user.user_id)
        setSuspendTarget(null)
        try {
            await authService.updateUserStatus(user.user_id, status)
            toast.success(
                status === "suspended"
                    ? `Akun ${user.email} berhasil ditangguhkan.`
                    : `Akun ${user.email} berhasil diaktifkan kembali.`
            )
            setUsers((prev) =>
                prev.map((u) => u.user_id === user.user_id ? { ...u, account_status: status } : u)
            )
        } catch (err: any) {
            toast.error(err.message || "Gagal mengubah status akun.")
        } finally {
            setIsSuspendingId(null)
        }
    }

    // ─────────────────────────────────────────────
    // Delete invited user
    // ─────────────────────────────────────────────
    const handleResendInvitation = async (user: UserListItem) => {
        setIsResendingId(user.user_id)
        try {
            await authService.resendInvitation(user.user_id)
            toast.success(`Undangan dikirim ulang ke ${user.email}.`)
        } catch (err: unknown) {
            const message = err instanceof Error ? err.message : "Gagal mengirim ulang undangan."
            toast.error(message)
        } finally {
            setIsResendingId(null)
        }
    }

    const handleDeleteConfirm = async () => {
        if (!deleteTarget) return
        setIsDeletingId(deleteTarget.user_id)
        setDeleteTarget(null)
        try {
            await authService.deleteUser(deleteTarget.user_id)
            toast.success(`Undangan untuk ${deleteTarget.email} berhasil dihapus.`)
            setUsers((prev) => prev.filter((u) => u.user_id !== deleteTarget.user_id))
        } catch (err: any) {
            toast.error(err.message || "Gagal menghapus pengguna.")
        } finally {
            setIsDeletingId(null)
        }
    }

    // ─────────────────────────────────────────────
    // Invite user
    // ─────────────────────────────────────────────
    const handleInviteSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        if (!inviteEmail || !inviteRole) return
        setIsSubmitting(true)
        try {
            const assignment = decodeRoleAssignment(inviteRole)
            await authService.inviteUser(inviteEmail, assignment.role, assignment.groupId)
            setInviteSuccess(true)
            fetchUsers() // refresh list
        } catch (error: any) {
            toast.error(error.message || "Gagal mengirim undangan. Silakan coba lagi.")
        } finally {
            setIsSubmitting(false)
        }
    }

    const handleInviteReset = () => {
        setInviteEmail("")
        setInviteRole("")
        setInviteSuccess(false)
    }

    const handleInviteClose = () => {
        setInviteOpen(false)
        setTimeout(handleInviteReset, 300)
    }

    const toggleRoleFilter = (f: RoleFilter) => {
        setRoleFilter((prev) =>
            prev.includes(f) ? prev.filter((x) => x !== f) : [...prev, f]
        )
    }

    // ─────────────────────────────────────────────
    // Render
    // ─────────────────────────────────────────────
    return (
        <section>
            {/* ── Header Row ── */}
            <div className="flex items-start justify-between gap-4">
                <div>
                    <SectionHeading icon={UserPenIcon} title="Manajemen Admin" />
                    <p className="text-gray-500 text-sm mt-1">
                        Master Admin dapat mengubah peran pengguna: Staff, Admin Manajemen, Admin Pelayanan, dan kelompok admin lainnya.
                    </p>
                </div>
                <Button
                    id="btn-add-user"
                    onClick={() => setInviteOpen(true)}
                    className="shrink-0 bg-[#6B5FAE] hover:bg-[#5b4f97] text-white rounded-xl px-5 gap-2 shadow-sm shadow-[#6B5FAE]/30 transition-all"
                >
                    <UserPlus className="w-4 h-4" />
                    Tambah Pengguna
                </Button>
            </div>

            {/* ── Invite User Modal ── */}
            <Dialog open={inviteOpen} onOpenChange={handleInviteClose}>
                <DialogContent className="max-w-md rounded-3xl bg-gray-50 p-8 gap-0 [&>button]:hidden">
                    {inviteSuccess ? (
                        <div className="flex flex-col items-center text-center py-4">
                            <button
                                onClick={handleInviteClose}
                                className="absolute top-5 right-5 text-gray-400 hover:text-gray-600 transition-colors"
                            >
                                <X className="w-5 h-5" />
                            </button>
                            <div className="w-16 h-16 rounded-full bg-emerald-100 flex items-center justify-center mb-5">
                                <CheckCircle2 className="w-9 h-9 text-emerald-500" />
                            </div>
                            <h3 className="text-xl font-bold text-gray-900 mb-2">Undangan Berhasil Dikirim!</h3>
                            <p className="text-sm text-gray-500 leading-relaxed mb-1">
                                Tautan aktivasi telah dikirim ke
                            </p>
                            <span className="text-sm font-semibold text-[#6B5FAE] bg-[#6B5FAE]/10 px-3 py-1 rounded-full mb-8">
                                {inviteEmail}
                            </span>
                            <Button
                                onClick={handleInviteReset}
                                variant="outline"
                                className="rounded-xl border-[#6B5FAE] text-[#6B5FAE] hover:bg-[#6B5FAE]/5 px-6"
                            >
                                <UserPlus className="w-4 h-4 mr-2" />
                                Undang Pengguna Lain
                            </Button>
                        </div>
                    ) : (
                        <>
                            <DialogHeader className="mb-4">
                                <div className="flex items-start justify-between">
                                    <div>
                                        <DialogTitle className="text-2xl font-bold text-[#6B5FAE] mb-2">
                                            Undang Pengguna Baru
                                        </DialogTitle>
                                        <DialogDescription className="text-sm text-[#6B5FAE]/80 leading-relaxed">
                                            Masukkan email dan tentukan role karyawan. Sistem akan mengirimkan tautan aktivasi ke email tersebut.
                                        </DialogDescription>
                                    </div>
                                    <button
                                        onClick={handleInviteClose}
                                        className="text-gray-400 hover:text-gray-600 transition-colors mt-1 shrink-0"
                                    >
                                        <X className="w-5 h-5" />
                                    </button>
                                </div>
                                <Separator className="mt-4 bg-[#6B5FAE]/30" />
                            </DialogHeader>

                            <form onSubmit={handleInviteSubmit} className="flex flex-col gap-5">
                                <div>
                                    <label htmlFor="invite-email" className={labelClass}>
                                        <span className="flex items-center gap-1.5 mb-1.5">
                                            <Mail className="w-3.5 h-3.5" />
                                            Email Karyawan
                                        </span>
                                    </label>
                                    <Input
                                        id="invite-email"
                                        type="email"
                                        required
                                        placeholder="contoh@rumahsakit.com"
                                        value={inviteEmail}
                                        onChange={(e) => setInviteEmail(e.target.value)}
                                        className={fieldClass}
                                    />
                                </div>

                                <div>
                                    <label htmlFor="invite-role" className={labelClass}>
                                        <span className="flex items-center gap-1.5 mb-1.5">
                                            <UserPenIcon className="w-3.5 h-3.5" />
                                            Role / Hak Akses
                                        </span>
                                    </label>
                                    <Select
                                        required
                                        value={inviteRole}
                                        onValueChange={setInviteRole}
                                    >
                                        <SelectTrigger id="invite-role" className={fieldClass}>
                                            <SelectValue placeholder="Pilih role pengguna" />
                                        </SelectTrigger>
                                        <SelectContent className="rounded-xl">
                                            {roleOptions.map((option) => (
                                                <SelectItem key={option.value} value={option.value}>
                                                    {option.label}
                                                </SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </div>

                                <div className="flex items-start gap-2.5 bg-amber-50 border border-amber-200 rounded-xl px-4 py-3">
                                    <div className="w-4 h-4 rounded-full bg-amber-400 flex items-center justify-center shrink-0 mt-0.5">
                                        <span className="text-white text-[10px] font-bold">i</span>
                                    </div>
                                    <p className="text-xs text-amber-700 leading-relaxed">
                                        Tautan aktivasi berlaku selama <strong>24 jam</strong>. Karyawan dapat mengatur password mereka sendiri melalui tautan tersebut.
                                    </p>
                                </div>

                                <div className="flex justify-end pt-1">
                                    <Button
                                        id="btn-send-invite"
                                        type="submit"
                                        disabled={isSubmitting || !inviteEmail || !inviteRole}
                                        className="bg-[#6B5FAE] hover:bg-[#5b4f97] text-white rounded-xl px-8 gap-2 font-semibold disabled:opacity-60"
                                    >
                                        {isSubmitting ? (
                                            <span className="flex items-center gap-2">
                                                <svg className="animate-spin w-4 h-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                                                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z" />
                                                </svg>
                                                Mengirim...
                                            </span>
                                        ) : (
                                            <span className="flex items-center gap-2">
                                                <Send className="w-4 h-4" />
                                                Kirim Undangan
                                            </span>
                                        )}
                                    </Button>
                                </div>
                            </form>
                        </>
                    )}
                </DialogContent>
            </Dialog>

            {/* ── Suspend Confirm Modal ── */}
            <Dialog open={!!suspendTarget} onOpenChange={() => setSuspendTarget(null)}>
                <DialogContent className="max-w-sm rounded-2xl p-6">
                    <DialogHeader>
                        <DialogTitle className="text-lg font-bold text-gray-900">Tangguhkan Akun?</DialogTitle>
                        <DialogDescription className="text-sm text-gray-500 mt-1">
                            Akun <strong>{suspendTarget?.email}</strong> akan ditangguhkan dan semua sesi aktif akan dicabut. User tidak dapat login sampai diaktifkan kembali.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="flex gap-3 mt-4 justify-end">
                        <Button variant="outline" className="rounded-xl" onClick={() => setSuspendTarget(null)}>
                            Batal
                        </Button>
                        <Button
                            className="bg-rose-500 hover:bg-rose-600 text-white rounded-xl"
                            onClick={() => suspendTarget && doUpdateStatus(suspendTarget, "suspended")}
                        >
                            <ShieldOff className="w-4 h-4 mr-1.5" />
                            Tangguhkan
                        </Button>
                    </div>
                </DialogContent>
            </Dialog>

            {/* ── Delete Confirm Modal ── */}
            <Dialog open={!!deleteTarget} onOpenChange={() => setDeleteTarget(null)}>
                <DialogContent className="max-w-sm rounded-2xl p-6">
                    <DialogHeader>
                        <DialogTitle className="text-lg font-bold text-gray-900">Hapus Undangan?</DialogTitle>
                        <DialogDescription className="text-sm text-gray-500 mt-1">
                            Undangan yang dikirim ke <strong>{deleteTarget?.email}</strong> akan dihapus permanen. Tindakan ini tidak dapat dibatalkan.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="flex gap-3 mt-4 justify-end">
                        <Button variant="outline" className="rounded-xl" onClick={() => setDeleteTarget(null)}>
                            Batal
                        </Button>
                        <Button
                            className="bg-red-600 hover:bg-red-700 text-white rounded-xl"
                            onClick={handleDeleteConfirm}
                        >
                            <Trash2 className="w-4 h-4 mr-1.5" />
                            Hapus
                        </Button>
                    </div>
                </DialogContent>
            </Dialog>

            {/* ── Stat Cards ── */}
            <div className="mt-8 grid grid-cols-1 md:grid-cols-3 gap-4">
                <StatCard label="Total Pengguna" value={activeUsers.length} icon={User} />
                <StatCard label="Total Admin" icon={Layers}>
                    <div className="flex flex-col gap-1">
                        <span className="text-3xl font-bold text-gray-900">{totalAdmin}</span>
                        <span className="inline-flex items-center w-fit px-2.5 py-0.5 rounded-full text-xs font-semibold bg-amber-50 text-amber-600">
                            Hak akses penuh
                        </span>
                    </div>
                </StatCard>
                <StatCard label="Total Staff" icon={Users}>
                    <div className="flex flex-col gap-1">
                        <span className="text-3xl font-bold text-gray-900">{totalStaff}</span>
                        <span className="inline-flex items-center w-fit px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-600">
                            Akses standar
                        </span>
                    </div>
                </StatCard>
            </div>

            {/* ── User Table ── */}
            <div className="bg-white rounded-2xl border border-gray-100 p-6 min-h-172 mt-8">
                {/* Table header controls */}
                <div className="flex items-center justify-between flex-wrap gap-4">
                    <div className="flex items-center gap-3">
                        <Users className="w-5 h-5 text-[#6B5FAE]" />
                        <p className="font-semibold text-gray-900">Semua Pengguna</p>
                    </div>
                    <div className="flex items-center gap-3">
                        {/* Search */}
                        <div className="relative">
                            <Search className="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" />
                            <Input
                                placeholder="Cari pengguna..."
                                value={search}
                                onChange={(e) => setSearch(e.target.value)}
                                className="pl-9 w-64 bg-gray-50 border-gray-100 rounded-xl"
                            />
                        </div>

                        {/* Refresh */}
                        <Button
                            variant="outline"
                            size="icon"
                            className="rounded-xl border-gray-100 bg-gray-50 text-gray-600"
                            onClick={fetchUsers}
                            title="Refresh data"
                        >
                            <RefreshCw className={`w-4 h-4 ${isLoading ? "animate-spin" : ""}`} />
                        </Button>

                        {/* Filter */}
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button
                                    variant="outline"
                                    className="rounded-xl border-gray-100 bg-gray-50 text-gray-600 font-normal gap-2"
                                >
                                    <SlidersHorizontal className="w-4 h-4" />
                                    Filter
                                    {roleFilter.length > 0 && (
                                        <span className="ml-1 w-5 h-5 rounded-full bg-[#6B5FAE] text-white text-xs flex items-center justify-center">
                                            {roleFilter.length}
                                        </span>
                                    )}
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end" className="w-48">
                                <DropdownMenuLabel>Filter berdasarkan role</DropdownMenuLabel>
                                <DropdownMenuSeparator />
                                {(["admin", "staff", "invited"] as RoleFilter[]).map((f) => (
                                    <DropdownMenuCheckboxItem
                                        key={f}
                                        checked={roleFilter.includes(f)}
                                        onCheckedChange={() => toggleRoleFilter(f)}
                                        onSelect={(e) => e.preventDefault()}
                                        className="capitalize"
                                    >
                                        {f === "invited" ? "Belum Aktif (Invited)" : f.charAt(0).toUpperCase() + f.slice(1)}
                                    </DropdownMenuCheckboxItem>
                                ))}
                                {roleFilter.length > 0 && (
                                    <>
                                        <DropdownMenuSeparator />
                                        <button
                                            onClick={() => setRoleFilter([])}
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

                {/* Table */}
                <div className="mt-6 rounded-xl overflow-hidden border border-gray-100">
                    <table className="w-full text-sm">
                        <thead>
                            <tr className="bg-[#6B5FAE] text-white text-left">
                                <th className="px-6 py-4 font-semibold w-14">No.</th>
                                <th className="px-6 py-4 font-semibold">Email</th>
                                <th className="px-6 py-4 font-semibold">Role Saat Ini</th>
                                <th className="px-6 py-4 font-semibold">Status</th>
                                <th className="px-6 py-4 font-semibold">Ubah Role</th>
                                <th className="px-6 py-4 font-semibold">Aksi</th>
                            </tr>
                        </thead>
                        <tbody>
                            {isLoading ? (
                                <tr>
                                    <td colSpan={6} className="px-6 py-14 text-center">
                                        <div className="flex flex-col items-center gap-3 text-gray-400">
                                            <svg className="animate-spin w-7 h-7 text-[#6B5FAE]" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                                                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z" />
                                            </svg>
                                            <span>Memuat data pengguna...</span>
                                        </div>
                                    </td>
                                </tr>
                            ) : loadError ? (
                                <tr>
                                    <td colSpan={6} className="px-6 py-14 text-center">
                                        <div className="flex flex-col items-center gap-3 text-rose-500">
                                            <AlertCircle className="w-7 h-7" />
                                            <span>{loadError}</span>
                                            <Button
                                                variant="outline"
                                                className="rounded-xl border-rose-200 text-rose-500 hover:bg-rose-50 text-xs"
                                                onClick={fetchUsers}
                                            >
                                                Coba lagi
                                            </Button>
                                        </div>
                                    </td>
                                </tr>
                            ) : pagedUsers.length === 0 ? (
                                <tr>
                                    <td colSpan={6} className="px-6 py-14 text-center text-gray-400">
                                        {search || roleFilter.length > 0
                                            ? "Tidak ada pengguna yang cocok dengan filter ini."
                                            : "Belum ada pengguna yang terdaftar."}
                                    </td>
                                </tr>
                            ) : (
                                pagedUsers.map((user, idx) => {
                                    const avatar = getAvatarColor(user.email)
                                    const initials = getInitials(user.email)
                                    const isInvited = user.account_status === "invited"
                                    const isSuspended = user.account_status === "suspended"
                                    const isSaving = savingId === user.user_id
                                    const isDeleting = isDeletingId === user.user_id
                                    const isResending = isResendingId === user.user_id
                                    const isSuspending = isSuspendingId === user.user_id
                                    const currentAssignment = getUserAssignment(user, groups)
                                    const pendingAssignment = pendingAssignments[user.user_id] ?? currentAssignment
                                    const hasRoleChange = !!pendingAssignment && pendingAssignment !== currentAssignment && !isInvited

                                    return (
                                        <tr
                                            key={user.user_id}
                                            className={`border-t border-gray-100 transition-colors ${isSuspended ? "bg-gray-50 opacity-70" : ""}`}
                                        >
                                            {/* No */}
                                            <td className="px-6 py-4 text-gray-500">
                                                {(page - 1) * PAGE_SIZE + idx + 1}
                                            </td>

                                            {/* Email */}
                                            <td className="px-6 py-4">
                                                <div className="flex items-center gap-3">
                                                    <div
                                                        className={`w-9 h-9 rounded-full flex items-center justify-center text-xs font-semibold shrink-0 ${avatar.bg} ${avatar.text}`}
                                                    >
                                                        {initials}
                                                    </div>
                                                    <div>
                                                        <p className="font-medium text-gray-900">{user.email}</p>
                                                        {isInvited && (
                                                            <span className="inline-flex items-center text-[10px] bg-amber-50 text-amber-600 px-2 py-0.5 rounded-full font-medium mt-0.5">
                                                                Menunggu aktivasi
                                                            </span>
                                                        )}
                                                    </div>
                                                </div>
                                            </td>

                                            {/* Current role */}
                                            <td className="px-6 py-4">
                                                <RoleBadge
                                                    label={getRoleDisplayLabel(user)}
                                                    variant={user.role === "admin" ? "admin" : "staff"}
                                                />
                                            </td>

                                            {/* Status */}
                                            <td className="px-6 py-4">
                                                {isInvited ? (
                                                    <span className="inline-flex items-center text-xs bg-amber-50 text-amber-600 px-2.5 py-1 rounded-full font-medium">
                                                        Invited
                                                    </span>
                                                ) : isSuspended ? (
                                                    <span className="inline-flex items-center text-xs bg-rose-50 text-rose-500 px-2.5 py-1 rounded-full font-medium">
                                                        Suspended
                                                    </span>
                                                ) : (
                                                    <span className="inline-flex items-center text-xs bg-emerald-50 text-emerald-600 px-2.5 py-1 rounded-full font-medium">
                                                        Aktif
                                                    </span>
                                                )}
                                            </td>

                                            {/* Role selector */}
                                            <td className="px-6 py-4">
                                                {isInvited ? (
                                                    <span className="text-xs text-gray-400 italic">—</span>
                                                ) : (
                                                    <Select
                                                        value={pendingAssignment || undefined}
                                                        onValueChange={(v) =>
                                                            setPendingAssignments((prev) => ({
                                                                ...prev,
                                                                [user.user_id]: v,
                                                            }))
                                                        }
                                                        disabled={isSuspended || isSaving}
                                                    >
                                                        <SelectTrigger className="rounded-xl bg-white w-52 text-sm">
                                                            <SelectValue placeholder="Pilih role" />
                                                        </SelectTrigger>
                                                        <SelectContent>
                                                            {roleOptions.map((option) => (
                                                                <SelectItem key={option.value} value={option.value}>
                                                                    {option.label}
                                                                </SelectItem>
                                                            ))}
                                                        </SelectContent>
                                                    </Select>
                                                )}
                                            </td>

                                            {/* Actions */}
                                            <td className="px-6 py-4">
                                                <div className="flex items-center gap-2">
                                                    {isInvited ? (
                                                        <>
                                                            <Button
                                                                size="sm"
                                                                variant="outline"
                                                                className="rounded-xl border-[#6B5FAE]/30 text-[#6B5FAE] hover:bg-[#6B5FAE]/5 gap-1.5"
                                                                disabled={isResending || isDeleting}
                                                                onClick={() => handleResendInvitation(user)}
                                                            >
                                                                {isResending ? (
                                                                    <svg className="animate-spin w-3.5 h-3.5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                                                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                                                                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z" />
                                                                    </svg>
                                                                ) : (
                                                                    <Send className="w-3.5 h-3.5" />
                                                                )}
                                                                Kirim Ulang
                                                            </Button>
                                                            <Button
                                                                size="sm"
                                                                variant="outline"
                                                                className="rounded-xl border-rose-200 text-rose-500 hover:bg-rose-50 gap-1.5"
                                                                disabled={isDeleting || isResending}
                                                                onClick={() => setDeleteTarget(user)}
                                                            >
                                                            {isDeleting ? (
                                                                <svg className="animate-spin w-3.5 h-3.5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                                                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                                                                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z" />
                                                                </svg>
                                                            ) : (
                                                                <Trash2 className="w-3.5 h-3.5" />
                                                            )}
                                                            Hapus
                                                        </Button>
                                                        </>
                                                    ) : (
                                                        <>
                                                            {/* Save role button — only visible when role changed */}
                                                            {hasRoleChange && (
                                                                <Button
                                                                    size="sm"
                                                                    className="bg-[#6B5FAE] hover:bg-[#5b4f97] text-white rounded-xl px-4 gap-1"
                                                                    disabled={isSaving}
                                                                    onClick={() => handleSaveRole(user)}
                                                                >
                                                                    {isSaving ? (
                                                                        <svg className="animate-spin w-3.5 h-3.5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                                                            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                                                                            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z" />
                                                                        </svg>
                                                                    ) : null}
                                                                    Simpan
                                                                </Button>
                                                            )}

                                                            {/* Suspend / Activate */}
                                                            <Button
                                                                size="sm"
                                                                variant="outline"
                                                                className={`rounded-xl gap-1.5 ${isSuspended
                                                                    ? "border-emerald-200 text-emerald-600 hover:bg-emerald-50"
                                                                    : "border-rose-200 text-rose-500 hover:bg-rose-50"
                                                                    }`}
                                                                disabled={isSuspending}
                                                                onClick={() => handleToggleStatus(user)}
                                                            >
                                                                {isSuspending ? (
                                                                    <svg className="animate-spin w-3.5 h-3.5" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                                                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                                                                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8z" />
                                                                    </svg>
                                                                ) : isSuspended ? (
                                                                    <ShieldCheck className="w-3.5 h-3.5" />
                                                                ) : (
                                                                    <ShieldOff className="w-3.5 h-3.5" />
                                                                )}
                                                                {isSuspended ? "Aktifkan" : "Tangguhkan"}
                                                            </Button>
                                                        </>
                                                    )}
                                                </div>
                                            </td>
                                        </tr>
                                    )
                                })
                            )}
                        </tbody>
                    </table>
                </div>

                {/* Pagination */}
                <div className="flex items-center justify-between mt-6 flex-wrap gap-3">
                    <p className="text-sm text-gray-400">
                        Halaman {page} · Menampilkan {pagedUsers.length} dari {filteredUsers.length} pengguna
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