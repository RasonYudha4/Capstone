"use client"

import { useState, useEffect } from "react"
import { useSearchParams, useNavigate } from "react-router"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from "@/components/ui/card"
import { ShieldCheck, Eye, EyeOff, Lock, CheckCircle2, ChevronRight, Loader2 } from "lucide-react"
import { authService } from "@/services/auth_service"
import { toast } from "sonner"

export default function SetupPassword() {
    const [searchParams] = useSearchParams()
    const navigate = useNavigate()
    const token = searchParams.get("token")

    const [password, setPassword] = useState("")
    const [confirmPassword, setConfirmPassword] = useState("")
    const [showPassword, setShowPassword] = useState(false)
    const [isSubmitting, setIsSubmitting] = useState(false)
    const [isSuccess, setIsSuccess] = useState(false)

    useEffect(() => {
        if (!token) {
            toast.error("Token tidak valid atau tidak ditemukan.")
            navigate("/login")
        }
    }, [token, navigate])

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()

        if (password !== confirmPassword) {
            toast.error("Password tidak cocok.")
            return
        }

        if (password.length < 8) {
            toast.error("Password minimal 8 karakter.")
            return
        }

        const hasUpperCase = /[A-Z]/.test(password)
        const hasLowerCase = /[a-z]/.test(password)
        const hasNumber = /[0-9]/.test(password)

        if (!hasUpperCase || !hasLowerCase || !hasNumber) {
            toast.error("Password harus mengandung minimal satu huruf besar, satu huruf kecil, dan satu angka.")
            return
        }

        setIsSubmitting(true)
        try {
            if (!token) throw new Error("Token missing")
            await authService.completeInvitation(token, password)
            setIsSuccess(true)
        } catch (error: any) {
            toast.error(error.message || "Gagal mengatur password. Mungkin token sudah kadaluarsa.")
        } finally {
            setIsSubmitting(false)
        }
    }

    if (isSuccess) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-[#F9F9F9] p-4">
                <Card className="w-full max-w-md border-none shadow-2xl rounded-3xl overflow-hidden animate-in fade-in zoom-in duration-500">
                    <div className="bg-[#6B5FAE] p-8 flex justify-center">
                        <div className="w-20 h-20 bg-white/20 rounded-full flex items-center justify-center">
                            <CheckCircle2 className="w-10 h-10 text-white" />
                        </div>
                    </div>
                    <CardContent className="p-8 text-center">
                        <h2 className="text-2xl font-bold text-gray-900 mb-3">Password Berhasil Diatur!</h2>
                        <p className="text-gray-500 mb-8 leading-relaxed">
                            Akun Anda telah aktif. Sekarang Anda dapat masuk ke dalam sistem menggunakan email dan password baru Anda.
                        </p>
                        <Button
                            onClick={() => navigate("/login")}
                            className="w-full bg-[#6B5FAE] hover:bg-[#5b4f97] text-white rounded-xl h-12 font-semibold transition-all group"
                        >
                            Masuk Ke Dashboard
                            <ChevronRight className="ml-2 w-4 h-4 group-hover:translate-x-1 transition-transform" />
                        </Button>
                    </CardContent>
                </Card>
            </div>
        )
    }

    return (
        <div className="min-h-screen flex items-center justify-center bg-[#F9F9F9] p-4">
            <Card className="w-full max-w-md border-none shadow-2xl rounded-3xl overflow-hidden">
                <CardHeader className="bg-[#6B5FAE] text-white p-8 space-y-2">
                    <div className="flex items-center gap-3 mb-2">
                        <div className="p-2 bg-white/20 rounded-lg">
                            <ShieldCheck className="w-5 h-5 text-white" />
                        </div>
                        <span className="text-sm font-medium tracking-wider uppercase opacity-80">Aktivasi Akun</span>
                    </div>
                    <CardTitle className="text-2xl font-bold">Atur Password Anda</CardTitle>
                    <CardDescription className="text-white/80">
                        Silakan buat password baru untuk menyelesaikan proses pendaftaran akun Anda.
                    </CardDescription>
                </CardHeader>

                <CardContent className="p-8 pb-0">
                    <form id="setup-form" onSubmit={handleSubmit} className="space-y-6">
                        <div className="space-y-2">
                            <label className="text-sm font-semibold text-[#6B5FAE] flex items-center gap-2">
                                <Lock className="w-3.5 h-3.5" />
                                Password Baru
                            </label>
                            <div className="relative">
                                <Input
                                    type={showPassword ? "text" : "password"}
                                    placeholder="Min. 8 karakter"
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    className="h-12 rounded-xl border-gray-200 bg-gray-50 focus:ring-[#6B5FAE] focus:border-[#6B5FAE] pr-10"
                                    required
                                />
                                <button
                                    type="button"
                                    onClick={() => setShowPassword(!showPassword)}
                                    className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                                >
                                    {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                                </button>
                            </div>
                        </div>

                        <div className="space-y-2">
                            <label className="text-sm font-semibold text-[#6B5FAE] flex items-center gap-2">
                                <Lock className="w-3.5 h-3.5" />
                                Konfirmasi Password
                            </label>
                            <Input
                                type={showPassword ? "text" : "password"}
                                placeholder="Ulangi password baru"
                                value={confirmPassword}
                                onChange={(e) => setConfirmPassword(e.target.value)}
                                className="h-12 rounded-xl border-gray-200 bg-gray-50 focus:ring-[#6B5FAE] focus:border-[#6B5FAE]"
                                required
                            />
                        </div>

                        <div className="bg-blue-50 border border-blue-100 rounded-xl p-4 space-y-2">
                            <p className="text-xs font-bold text-blue-700">Syarat Password:</p>
                            <ul className="text-[11px] text-blue-600/80 space-y-1 list-disc pl-4 leading-tight">
                                <li>Minimal 8 karakter</li>
                                <li>Kombinasi huruf besar & kecil</li>
                                <li>Harus mengandung angka</li>
                            </ul>
                        </div>
                    </form>
                </CardContent>

                <CardFooter className="p-8">
                    <Button
                        form="setup-form"
                        type="submit"
                        disabled={isSubmitting || !password || !confirmPassword}
                        className="w-full bg-[#6B5FAE] hover:bg-[#5b4f97] text-white rounded-xl h-12 font-semibold shadow-lg shadow-[#6B5FAE]/20 transition-all disabled:opacity-70"
                    >
                        {isSubmitting ? (
                            <>
                                <Loader2 className="mr-2 h-4 h-4 animate-spin" />
                                Menyimpan...
                            </>
                        ) : (
                            "Aktifkan Akun Saya"
                        )}
                    </Button>
                </CardFooter>
            </Card>
        </div>
    )
}
