export type Role = 'admin' | 'manager' | 'editor' | 'viewer'

export type UserStatus = 'active' | 'inactive' | 'suspended'

export interface User {
  id: string
  name: string
  email: string
  role: Role
  status: UserStatus
  avatar?: string
  department: string
  joinedAt: string
  lastLogin: string
  phone?: string
}

export interface StatsCard {
  title: string
  value: string | number
  change: number
  changeLabel: string
  icon: string
}

export interface Activity {
  id: string
  user: string
  action: string
  target: string
  time: string
  type: 'create' | 'update' | 'delete' | 'login'
}
