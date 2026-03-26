export type Role = 1 | 2 | 3

export interface User {
  id: string
  name: string
  role: Role
}