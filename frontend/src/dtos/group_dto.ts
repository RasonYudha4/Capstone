import { z } from "zod";

export const groupSchema = z.object({
    group_id: z.string(),
    group_name: z.string(),
});

export const groupListSchema = z.array(groupSchema);

export type Group = z.infer<typeof groupSchema>;
