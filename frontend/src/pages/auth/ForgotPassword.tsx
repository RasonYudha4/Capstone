"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate } from "react-router";
import { forgotPasswordSchema, type ForgotPasswordFormValues } from "@/dtos/login_dto";
import { authService } from "@/services/auth_service";
import { AxiosError } from "axios";
import { voidApiResponseSchema } from "@/dtos/login_dto";
import { CheckCircle2, ChevronLeft } from "lucide-react";

export default function ForgotPassword() {
    const [isLoading, setIsLoading] = useState(false);
    const [serverError, setServerError] = useState<string | null>(null);
    const [isSuccess, setIsSuccess] = useState(false);
    const navigate = useNavigate();

    const {
        register,
        handleSubmit,
        formState: { errors },
        getValues,
    } = useForm<ForgotPasswordFormValues>({
        resolver: zodResolver(forgotPasswordSchema),
        defaultValues: { email: "" },
    });

    const onSubmit = async (formData: ForgotPasswordFormValues) => {
        setIsLoading(true);
        setServerError(null);
        try {
            await authService.forgotPassword(formData.email);
            setIsSuccess(true);
        } catch (error) {
            if (error instanceof AxiosError) {
                if (!error.response) {
                    setServerError("Tidak dapat terhubung ke server. Pastikan backend berjalan dan coba lagi.");
                } else if (error.response.status >= 500) {
                    setServerError("Server sedang bermasalah. Coba restart backend (docker compose up -d) lalu coba lagi.");
                } else if (error.response.data) {
                    const result = voidApiResponseSchema.safeParse(error.response.data);
                    setServerError(result.success ? result.data.message : "Terjadi kesalahan. Silakan coba lagi.");
                } else {
                    setServerError("Terjadi kesalahan. Silakan coba lagi.");
                }
            } else if (error instanceof Error) {
                setServerError(error.message);
            } else {
                setServerError("Terjadi kesalahan. Silakan coba lagi.");
            }
        } finally {
            setIsLoading(false);
        }
    };

    if (isSuccess) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-white px-4">
                <div className="w-full max-w-md bg-[#8571C1] rounded-3xl p-10 text-center">
                    <div className="w-14 h-14 rounded-full bg-white/15 flex items-center justify-center mx-auto mb-5">
                        <CheckCircle2 className="w-7 h-7 text-white" />
                    </div>
                    <h1 className="text-2xl font-semibold text-white mb-3">Cek Email Anda</h1>
                    <p className="text-sm text-white/70 leading-relaxed mb-2">
                        Jika akun dengan email <strong className="text-white">{getValues("email")}</strong> terdaftar,
                        link reset password telah dikirim.
                    </p>
                    <p className="text-xs text-white/50 mb-8">Link berlaku selama 1 jam.</p>
                    <button
                        type="button"
                        onClick={() => navigate("/login")}
                        className="w-full py-3 rounded-xl bg-white text-[#6B5FAE] text-sm font-semibold transition-opacity hover:opacity-90"
                    >
                        Kembali ke Login
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen flex items-center justify-center bg-white px-4">
            <div className="w-full max-w-md bg-[#8571C1] rounded-3xl p-10">
                <button
                    type="button"
                    onClick={() => navigate("/login")}
                    className="flex items-center gap-1.5 text-white/60 hover:text-white text-xs mb-6 transition-colors"
                >
                    <ChevronLeft className="w-4 h-4" />
                    Kembali ke login
                </button>

                <div className="flex flex-col items-center text-center mb-8">
                    <h1 className="text-2xl font-semibold text-white mb-2">Lupa Password?</h1>
                    <p className="text-sm text-white/60 leading-relaxed">
                        Masukkan email akun Anda. Kami akan mengirim link untuk reset password.
                    </p>
                </div>

                {serverError && (
                    <div className="mb-4 p-3 rounded-xl bg-red-500/20 border border-red-400/30 text-red-200 text-sm text-center">
                        {serverError}
                    </div>
                )}

                <form onSubmit={handleSubmit(onSubmit)} noValidate>
                    <div className="mb-6">
                        <label className="block text-xs font-medium text-white/75 tracking-wide mb-1.5">
                            Email address
                        </label>
                        <input
                            id="forgot-email"
                            type="email"
                            placeholder="you@example.com"
                            autoComplete="email"
                            {...register("email")}
                            className={`w-full px-4 py-3 rounded-xl bg-white/10 border text-white placeholder-white/35 text-sm outline-none transition-colors focus:border-white/50 ${errors.email ? "border-red-400/70" : "border-white/15"}`}
                        />
                        {errors.email && (
                            <p className="mt-1.5 text-xs text-red-300">{errors.email.message}</p>
                        )}
                    </div>

                    <button
                        id="forgot-submit"
                        type="submit"
                        disabled={isLoading}
                        className="w-full py-3 rounded-xl bg-white text-[#6B5FAE] text-sm font-semibold transition-opacity hover:opacity-90 disabled:opacity-60 disabled:cursor-not-allowed"
                    >
                        {isLoading ? "Mengirim..." : "Kirim Link Reset"}
                    </button>
                </form>
            </div>
        </div>
    );
}
