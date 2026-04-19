"use client";

import { useState } from "react";
import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
    InputOTP,
    InputOTPGroup,
    InputOTPSlot,
} from "@/components/ui/input-otp";
import { otpSchema, type OTPFormValues } from "@/dtos/login_dto";


export default function VerifyOTP() {
    const [isLoading, setIsLoading] = useState(false);
    const [isResending, setIsResending] = useState(false);

    const {
        control,
        handleSubmit,
        watch,
        reset,
        formState: { errors },
    } = useForm<OTPFormValues>({
        resolver: zodResolver(otpSchema),
        defaultValues: { otp: "" },
    });

    const otpValue = watch("otp");

    const onSubmit = async (data: OTPFormValues) => {
        setIsLoading(true);
        try {
            // TODO: plug in OTP verification via auth context
            console.log("OTP submitted:", data.otp);
        } finally {
            setIsLoading(false);
        }
    };

    const onResend = async () => {
        setIsResending(true);
        try {
            // TODO: trigger resend OTP via auth context
            console.log("Resend OTP triggered");
            reset();
        } finally {
            setIsResending(false);
        }
    };

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