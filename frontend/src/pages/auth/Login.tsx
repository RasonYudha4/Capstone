"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate } from "react-router";
import { loginSchema, type LoginFormValues } from "@/dtos/login_dto";
import { authService } from "@/services/auth_service";
import { useAuth } from "@/cores/AuthContext";
import { AxiosError } from "axios";
import { voidApiResponseSchema } from "@/dtos/login_dto";

export default function Login() {
    const [isLoading, setIsLoading] = useState(false);
    const [serverError, setServerError] = useState<string | null>(null);
    const navigate = useNavigate();
    const { login } = useAuth();

    const {
        register,
        handleSubmit,
        formState: { errors },
    } = useForm<LoginFormValues>({
        resolver: zodResolver(loginSchema),
        defaultValues: {
            email: "",
            password: "",
        },
    });

    const onSubmit = async (formData: LoginFormValues) => {
        setIsLoading(true);
        setServerError(null);
        try {
            const response = await authService.login(formData.email, formData.password);

            if (response.requires_otp) {
                // Admin / master-admin → navigate to OTP verification
                navigate("/verify", { state: { email: formData.email, preAuthToken: response.pre_auth_token } });
            } else {
                // Staff → tokens returned immediately
                localStorage.setItem("accessToken", response.access_token!);
                localStorage.setItem("refreshToken", response.refresh_token!);

                const user = await authService.me();
                login(user, response.access_token!, response.refresh_token!);
                navigate("/dashboard", { replace: true });
            }
        } catch (error) {
            if (error instanceof AxiosError && error.response?.data) {
                const result = voidApiResponseSchema.safeParse(error.response.data);
                setServerError(result.success ? result.data.message : "An unexpected error occurred.");
            } else if (error instanceof Error) {
                setServerError(error.message);
            } else {
                setServerError("An unexpected error occurred. Please try again.");
            }
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-white px-4">
            <div className="w-full max-w-md bg-[#8571C1] rounded-3xl p-10">

                <div className="flex flex-col items-center text-center mb-8">
                    <div className="w-12 h-12 rounded-full bg-white/15 flex items-center justify-center mb-5">
                        <img src='/logo.png' height={35} width={35} />
                    </div>
                    <h1 className="text-3xl font-semibold text-white mb-2">
                        Welcome back
                    </h1>
                    <p className="text-sm text-white/60 leading-relaxed">
                        Akses progres dan data terkait
                        <br />
                        akreditasi rumah sakit
                    </p>
                </div>

                {serverError && (
                    <div className="mb-4 p-3 rounded-xl bg-red-500/20 border border-red-400/30 text-red-200 text-sm text-center">
                        {serverError}
                    </div>
                )}

                <form onSubmit={handleSubmit(onSubmit)} noValidate>
                    <div className="mb-4">
                        <label className="block text-xs font-medium text-white/75 tracking-wide mb-1.5">
                            Email address
                        </label>
                        <div className="relative">
                            <svg
                                className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-white/45 pointer-events-none"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                                strokeWidth={2}
                            >
                                <rect x="2" y="4" width="20" height="16" rx="2" />
                                <path d="m2 7 10 7 10-7" />
                            </svg>
                            <input
                                id="login-email"
                                type="email"
                                placeholder="you@example.com"
                                autoComplete="email"
                                {...register("email")}
                                className={`w-full pl-10 pr-4 py-3 rounded-xl bg-white/10 border text-white placeholder-white/35 text-sm outline-none transition-colors focus:border-white/50 ${errors.email ? "border-red-400/70" : "border-white/15"
                                    }`}
                            />
                        </div>
                        {errors.email && (
                            <p className="mt-1.5 text-xs text-red-300">
                                {errors.email.message}
                            </p>
                        )}
                    </div>

                    <div className="mb-6">
                        <label className="block text-xs font-medium text-white/75 tracking-wide mb-1.5">
                            Password
                        </label>
                        <div className="relative">
                            <svg
                                className="absolute left-3.5 top-1/2 -translate-y-1/2 w-4 h-4 text-white/45 pointer-events-none"
                                fill="none"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                                strokeWidth={2}
                            >
                                <rect x="3" y="11" width="18" height="11" rx="2" />
                                <path d="M7 11V7a5 5 0 0110 0v4" />
                            </svg>
                            <input
                                id="login-password"
                                type="password"
                                placeholder="••••••••"
                                autoComplete="current-password"
                                {...register("password")}
                                className={`w-full pl-10 pr-4 py-3 rounded-xl bg-white/10 border text-white placeholder-white/35 text-sm outline-none transition-colors focus:border-white/50 ${errors.password ? "border-red-400/70" : "border-white/15"
                                    }`}
                            />
                        </div>
                        {errors.password && (
                            <p className="mt-1.5 text-xs text-red-300">
                                {errors.password.message}
                            </p>
                        )}
                    </div>

                    <button
                        id="login-submit"
                        type="submit"
                        disabled={isLoading}
                        className="w-full py-3 rounded-xl bg-white text-[#6B5FAE] text-sm font-semibold transition-opacity hover:opacity-90 active:scale-[0.98] disabled:opacity-60 disabled:cursor-not-allowed"
                    >
                        {isLoading ? "Signing in..." : "Sign in"}
                    </button>
                </form>

                <div className="flex items-center gap-3 my-6">
                    <div className="flex-1 h-px bg-white/15" />
                    <span className="text-xs text-white/50 font-medium">or</span>
                    <div className="flex-1 h-px bg-white/15" />
                </div>

                <button
                    type="button"
                    onClick={() => navigate("/")}
                    className="w-full py-3 rounded-xl bg-transparent border border-white/25 text-white text-sm font-semibold transition-colors hover:bg-white/10 active:scale-[0.98]"
                >
                    Akses sebagai staff
                </button>
            </div>
        </div>
    );
}