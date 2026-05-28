import type { Role } from './common'

export interface User {
  id?: number
  user_id?: number
  username: string
  real_name?: string
  name?: string
  email?: string
  role: Role
  phone?: string
  department?: string
  status?: number | 'active' | 'disabled' | 'locked'
  last_login_at?: string
  created_at?: string
  updated_at?: string
}

export interface LoginCredentials {
  username: string
  password: string
}

export interface TokenResponse {
  token: string
  user_id: number
  username: string
  real_name: string
  role: string
  department: string
}
