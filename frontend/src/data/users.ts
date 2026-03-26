import type { User } from '../context/types'

interface UserWithCredentials extends User {
    email: string
    password: string
}

export const dummyUsers: UserWithCredentials[] = [
    {
        id: '1',
        name: 'Alice Reyes',
        email: 'admin@example.com',
        password: 'admin123',
        role: 2,
    },
    {
        id: '2',
        name: 'Bruno Santos',
        email: 'master@example.com',
        password: 'master123',
        role: 1,
    },
    {
        id: '3',
        name: 'Clara Mendoza',
        email: 'user@example.com',
        password: 'user123',
        role: 3,
    },
]

export function findUser(email: string, password: string): User | null {
    const match = dummyUsers.find(
        (u) => u.email === email && u.password === password
    )
    if (!match) return null

    // strip credentials before returning — context should never hold them
    const { email: _, password: __, ...user } = match
    return user
}