import { z } from "zod";

// ─── Request Schemas ──────────────────────────────────────────────────────────

export const loginSchema = z.object({
    email: z.string().email("Please enter a valid email address"),
    password: z.string().min(1, "Password is required"),
});

export const otpSchema = z.object({
    email: z.string().email(),
    otp: z
        .string()
        .length(6, "Please enter the complete 6-digit code")
        .regex(/^\d+$/, "OTP must contain digits only"),
});

// ─── Response Schemas ─────────────────────────────────────────────────────────

export const loginResponseSchema = z.object({
    requires_otp: z.boolean(),
    access_token: z.string().optional(),
    refresh_token: z.string().optional(),
    expires_in: z.string().optional(),
});

export const tokenResponseSchema = z.object({
    access_token: z.string(),
    refresh_token: z.string(),
    expires_in: z.string(),
});

export const userResponseSchema = z.object({
    user_id: z.string(),
    email: z.string(),
    role: z.enum(["staff", "admin", "master-admin"]),
    group_id: z.string().optional(),
});

// For parsing error responses (success=false)
export const voidApiResponseSchema = z.object({
    success: z.boolean(),
    message: z.string(),
});

// ─── Inferred Types ───────────────────────────────────────────────────────────

export type LoginFormValues = z.infer<typeof loginSchema>;
export type OTPFormValues = z.infer<typeof otpSchema>;
export type LoginResponse = z.infer<typeof loginResponseSchema>;
export type TokenResponse = z.infer<typeof tokenResponseSchema>;
export type UserResponse = z.infer<typeof userResponseSchema>;