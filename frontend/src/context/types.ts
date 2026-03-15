export type Role = 'admin' | 'master-admin' | 'user'

export interface User {
  id: string
  name: string
  role: Role
}