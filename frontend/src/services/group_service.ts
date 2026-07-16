import axioHandler from '@/cores/axios'
import { apiResponseSchema } from '@/dtos/login_dto'
import { groupListSchema, type Group } from '@/dtos/group_dto'

export const groupService = {
    // GET /groups
    listGroups: async (): Promise<Group[]> => {
        const { data } = await axioHandler.get('/groups')
        const parsed = apiResponseSchema(groupListSchema).parse(data)
        if (!parsed.success) {
            throw new Error(parsed.message)
        }
        return groupListSchema.parse(parsed.data ?? [])
    },
}
