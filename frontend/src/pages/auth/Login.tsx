"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { loginSchema, type LoginFormValues } from "@/dtos/login_dto";

export default function Login() {
    const [isEmailLoading, setIsEmailLoading] = useState(false);
    const [isGoogleLoading, setIsGoogleLoading] = useState(false);

    const {
        register,
        handleSubmit,
        formState: { errors },
    } = useForm<LoginFormValues>({
        resolver: zodResolver(loginSchema),
        defaultValues: {
            email: "",
        },
    });

    const onEmailSubmit = async (data: LoginFormValues) => {
        setIsEmailLoading(true);
        try {
            // TODO: plug in auth context / API call
            console.log("Email login:", data);
        } finally {
            setIsEmailLoading(false);
        }
    };

    const onGoogleSignIn = async () => {
        setIsGoogleLoading(true);
        try {
            // TODO: plug in Google OAuth via auth context
            console.log("Google sign-in triggered");
        } finally {
            setIsGoogleLoading(false);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-white px-4">
            <div className="w-full max-w-md bg-[#8571C1] rounded-3xl p-10">

                <div className="flex flex-col items-center text-center mb-8">
                    <div className="w-12 h-12 rounded-full bg-white/15 flex items-center justify-center mb-5">
                        <svg
                            className="w-6 h-6 text-white"
                            fill="currentColor"
                            viewBox="0 0 24 24"
                        >
                            <path d="M12 2C8.13 2 5 5.13 5 9c0 2.38 1.19 4.47 3 5.74V17c0 .55.45 1 1 1h6c.55 0 1-.45 1-1v-2.26C17.81 13.47 19 11.38 19 9c0-3.87-3.13-7-7-7z" />
                            <path d="M9 21c0 .55.45 1 1 1h4c.55 0 1-.45 1-1v-1H9v1z" />
                        </svg>
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

                <form onSubmit={handleSubmit(onEmailSubmit)} noValidate>
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

                    <button
                        type="submit"
                        disabled={isEmailLoading}
                        className="w-full py-3 rounded-xl bg-white text-[#6B5FAE] text-sm font-semibold transition-opacity hover:opacity-90 active:scale-[0.98] disabled:opacity-60 disabled:cursor-not-allowed"
                    >
                        {isEmailLoading ? "Sending..." : "Continue with email"}
                    </button>
                </form>

                <div className="flex items-center gap-3 my-6">
                    <div className="flex-1 h-px bg-white/20" />
                    <span className="text-xs text-white/40">or</span>
                    <div className="flex-1 h-px bg-white/20" />
                </div>

                <button
                    type="button"
                    onClick={onGoogleSignIn}
                    disabled={isGoogleLoading}
                    className="w-full py-3 rounded-xl border border-white/20 bg-white/8 text-white text-sm font-medium flex items-center justify-center gap-2.5 transition-colors hover:bg-white/15 active:scale-[0.98] disabled:opacity-60 disabled:cursor-not-allowed"
                >
                    {isGoogleLoading ? (
                        <span>Redirecting...</span>
                    ) : (
                        <>
                            <svg width="16" height="16" viewBox="0 0 24 24">
                                <path fill="#EA4335" d="M23.745 12.27c0-.79-.07-1.54-.19-2.27h-11.3v4.51h6.47c-.29 1.48-1.14 2.73-2.4 3.58v3h3.86c2.26-2.09 3.56-5.17 3.56-8.82z" />
                                <path fill="#4285F4" d="M12.255 24c3.24 0 5.95-1.08 7.93-2.91l-3.86-3c-1.08.72-2.45 1.16-4.07 1.16-3.13 0-5.78-2.11-6.73-4.96h-3.98v3.09C3.515 21.3 7.615 24 12.255 24z" />
                                <path fill="#FBBC05" d="M5.525 14.29c-.25-.72-.38-1.49-.38-2.29s.14-1.57.38-2.29V6.62h-3.98a11.86 11.86 0 000 10.76l3.98-3.09z" />
                                <path fill="#34A853" d="M12.255 4.75c1.77 0 3.35.61 4.6 1.8l3.42-3.42C18.205 1.19 15.495 0 12.255 0c-4.64 0-8.74 2.7-10.71 6.62l3.98 3.09c.95-2.85 3.6-4.96 6.73-4.96z" />
                            </svg>
                            Sign in with Google
                        </>
                    )}
                </button>
            </div>
        </div>
    );
}