import { z } from "zod";

// request schemas

// POST /auth/login
export const loginSchema = z.object({
    email: z
        .string()
        .email("Please enter a valid email address"),
    password: z
        .string()
        .min(8, "Password must be at least 8 characters"),
});

// POST /auth/verify-otp 
export const otpSchema = z.object({
    email: z
        .string()
        .email("Please enter a valid email address"),
    otp: z
        .string()
        .length(6, "Please enter the complete 6-digit code")
        .regex(/^\d+$/, "OTP must contain digits only"),
});

// POST /auth/resend-otp 
export const resendOtpSchema = z.object({
    email: z
        .string()
        .email("Please enter a valid email address"),
});

// POST /auth/refresh
export const refreshSchema = z.object({
    refresh_token: z.string(),
});

// POST /auth/logout
export const logoutSchema = z.object({
    refresh_token: z.string(),
});

// POST /auth/invite
export const inviteSchema = z.object({
    email: z.string().email("Please enter a valid email address"),
    role: z.enum(["admin", "staff"]),
    group_id: z.string().optional(),
});

// POST /auth/complete-invitation
export const completeInvitationSchema = z.object({
    token: z.string(),
    password: z.string().min(8, "Password must be at least 8 characters"),
});

// PUT /auth/users/:id/role
export const updateRoleSchema = z.object({
    role: z.enum(["admin", "staff"]),
    group_id: z.string().optional(),
});

// PUT /auth/users/:id/status
export const updateStatusSchema = z.object({
    status: z.enum(["active", "suspended"]),
});

// response schemas

// generic envelope
export const apiResponseSchema = <T extends z.ZodTypeAny>(dataSchema: T) =>
    z.object({
        success: z.boolean(),
        message: z.string(),
        data: dataSchema.optional(),
    });

// void envelope
export const voidApiResponseSchema = z.object({
    success: z.boolean(),
    message: z.string(),
});

// POST /auth/login
export const loginResponseSchema = z.object({
    requires_otp: z.boolean(),
    access_token: z.string().optional(),
    pre_auth_token: z.string().optional(),
    refresh_token: z.string().optional(),
    expires_in: z.string().optional(),
});

// POST /auth/verify-otp
export const tokenResponseSchema = z.object({
    access_token: z.string(),
    refresh_token: z.string(),
    expires_in: z.string(),
});

// POST /auth/refresh
export const refreshResponseSchema = z.object({
    access_token: z.string(),
    refresh_token: z.string(),
    expires_in: z.string(),
});

// GET /auth/me
export const userResponseSchema = z.object({
    user_id: z.string(),
    email: z.string(),
    role: z.enum(["staff", "admin", "master-admin"]),
    group_id: z.string().optional(),
});

// GET /auth/users — single item
export const userListItemSchema = z.object({
    user_id: z.string(),
    email: z.string(),
    role: z.enum(["staff", "admin", "master-admin"]),
    group_id: z.string().nullable().optional(),
    group_name: z.string().nullable().optional(),
    account_status: z.enum(["invited", "active", "suspended"]),
    verified: z.boolean(),
});

// GET /auth/users — full response
export const listUsersResponseSchema = z.object({
    users: z.array(userListItemSchema),
    total: z.number(),
});

// inferred types

export type LoginFormValues = z.infer<typeof loginSchema>;
export type OTPFormValues = z.infer<typeof otpSchema>;
export type ResendOtpValues = z.infer<typeof resendOtpSchema>;
export type RefreshValues = z.infer<typeof refreshSchema>;
export type LogoutValues = z.infer<typeof logoutSchema>;
export type InviteValues = z.infer<typeof inviteSchema>;
export type CompleteInvitationValues = z.infer<typeof completeInvitationSchema>;
export type UpdateRoleValues = z.infer<typeof updateRoleSchema>;
export type UpdateStatusValues = z.infer<typeof updateStatusSchema>;

export type LoginResponse = z.infer<typeof loginResponseSchema>;
export type TokenResponse = z.infer<typeof tokenResponseSchema>;
export type RefreshResponse = z.infer<typeof refreshResponseSchema>;
export type UserResponse = z.infer<typeof userResponseSchema>;
export type UserListItem = z.infer<typeof userListItemSchema>;
export type ListUsersResponse = z.infer<typeof listUsersResponseSchema>;