import { z } from "zod";

export const loginSchema = z.object({
    email: z
        .email("Please enter a valid email address"),
});

export const otpSchema = z.object({
    otp: z
        .string()
        .length(6, "Please enter the complete 6-digit code")
        .regex(/^\d+$/, "OTP must contain digits only"),
});

export type LoginFormValues = z.infer<typeof loginSchema>;
export type OTPFormValues = z.infer<typeof otpSchema>;