"use client";

import { useState, useEffect } from "react";
import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useNavigate, useLocation } from "react-router";
import {
    InputOTP,
    InputOTPGroup,
    InputOTPSlot,
} from "@/components/ui/input-otp";
import { otpSchema, type OTPFormValues } from "@/dtos/login_dto";
import { authService } from "@/services/auth_service";
import { useAuth } from "@/cores/AuthContext";
import { AxiosError } from "axios";
import { voidApiResponseSchema } from "@/dtos/login_dto";


export default function VerifyOTP() {
    const [isLoading, setIsLoading] = useState(false);
    const [isResending, setIsResending] = useState(false);
    const [serverError, setServerError] = useState<string | null>(null);
    const [resendSuccess, setResendSuccess] = useState<string | null>(null);

    const navigate = useNavigate();
    const location = useLocation();
    const { login } = useAuth();

    // Email passed from Login page via route state
    const email = (location.state as { email?: string })?.email;

    // Guard: redirect to login if no email in state
    useEffect(() => {
        if (!email) {
            navigate("/login", { replace: true });
        }
    }, [email, navigate]);

    const {
        control,
        handleSubmit,
        watch,
        reset,
        formState: { errors },
    } = useForm<OTPFormValues>({
        resolver: zodResolver(otpSchema),
        defaultValues: { email: email ?? "", otp: "" },
    });

    const otpValue = watch("otp");

    const onSubmit = async (data: OTPFormValues) => {
        setIsLoading(true);
        setServerError(null);
        try {
            const tokens = await authService.verifyOtp(data.email, data.otp);

            // Store tokens temporarily so the /auth/me call has a valid Bearer token
            localStorage.setItem("accessToken", tokens.access_token);
            localStorage.setItem("refreshToken", tokens.refresh_token);

            const user = await authService.me();
            login(user, tokens.access_token, tokens.refresh_token);
            navigate("/dashboard", { replace: true });
        } catch (error) {
            if (error instanceof AxiosError && error.response?.data) {
                const result = voidApiResponseSchema.safeParse(error.response.data);
                setServerError(result.success ? result.data.message : "Verification failed.");
            } else if (error instanceof Error) {
                setServerError(error.message);
            } else {
                setServerError("Verification failed. Please try again.");
            }
        } finally {
            setIsLoading(false);
        }
    };

    const onResend = async () => {
        if (!email) return;
        setIsResending(true);
        setServerError(null);
        setResendSuccess(null);
        try {
            await authService.resendOtp(email);
            reset({ email, otp: "" });
            setResendSuccess("A new OTP has been sent to your email.");
        } catch (error) {
            if (error instanceof AxiosError && error.response?.data) {
                const result = voidApiResponseSchema.safeParse(error.response.data);
                setServerError(result.success ? result.data.message : "Failed to resend OTP.");
            } else if (error instanceof Error) {
                setServerError(error.message);
            } else {
                setServerError("Failed to resend OTP. Please try again.");
            }
        } finally {
            setIsResending(false);
        }
    };

    if (!email) return null;

    return (
        <div className="min-h-screen flex items-center justify-center bg-white px-4">
            <div className="w-full max-w-md bg-[#6B5FAE] rounded-3xl p-10">

                {/* Header */}
                <div className="flex flex-col items-center text-center mb-8">
                    <div className="w-12 h-12 rounded-full bg-white/15 flex items-center justify-center mb-5">
                        <svg
                            className="w-6 h-6 text-white"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                            strokeWidth={2}
                        >
                            <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z"
                            />
                        </svg>
                    </div>
                    <h1 className="text-3xl font-semibold text-white mb-2">
                        Verification Code
                    </h1>
                    <p className="text-sm text-white/60 leading-relaxed">
                        Silahkan periksa email dan masukkan kode
                        <br />
                        verifikasi 6-digit yang diterima.
                    </p>
                </div>

                {/* Error Message */}
                {serverError && (
                    <div className="mb-4 p-3 rounded-xl bg-red-500/20 border border-red-400/30 text-red-200 text-sm text-center">
                        {serverError}
                    </div>
                )}

                {/* Success Message */}
                {resendSuccess && (
                    <div className="mb-4 p-3 rounded-xl bg-green-500/20 border border-green-400/30 text-green-200 text-sm text-center">
                        {resendSuccess}
                    </div>
                )}

                {/* OTP Form */}
                <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
                    <div className="flex flex-col items-center">
                        <Controller
                            control={control}
                            name="otp"
                            render={({ field }) => (
                                <InputOTP
                                    maxLength={6}
                                    value={field.value}
                                    onChange={field.onChange}
                                    disabled={isLoading}
                                >
                                    <InputOTPGroup className="gap-3">
                                        {[0, 1, 2, 3, 4, 5].map((index) => (
                                            <InputOTPSlot
                                                key={index}
                                                index={index}
                                                className="
                          w-11 h-14 rounded-xl text-lg font-semibold
                          bg-white/15 border-white/25 text-white
                          caret-white
                          focus:border-white focus:ring-0 focus:bg-white/20
                          data-[active=true]:border-white data-[active=true]:bg-white/20
                          transition-all
                        "
                                            />
                                        ))}
                                    </InputOTPGroup>
                                </InputOTP>
                            )}
                        />
                        {errors.otp && (
                            <p className="text-red-300 text-xs mt-3">{errors.otp.message}</p>
                        )}
                    </div>

                    <button
                        id="verify-submit"
                        type="submit"
                        disabled={isLoading || otpValue.length < 6}
                        className="w-full py-3 rounded-xl bg-white text-[#6B5FAE] text-sm font-semibold transition-all hover:opacity-90 active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        {isLoading ? "Verifying..." : "Continue"}
                    </button>
                </form>

                {/* Resend */}
                <p className="text-center text-sm text-white/50 mt-6">
                    Tidak mendapatkan kode?{" "}
                    <button
                        id="resend-otp"
                        type="button"
                        onClick={onResend}
                        disabled={isResending}
                        className="text-white font-semibold hover:text-white/80 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        {isResending ? "Mengirim..." : "Kirim Ulang"}
                    </button>
                </p>
            </div>
        </div>
    );
}