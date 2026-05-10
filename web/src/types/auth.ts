import type { Role } from './common'

export interface User {
  id: number
  username: string
  name: string
  email: string
  role: Role
  phone?: string
  department?: string
  status: 'active' | 'disabled' | 'locked'
  last_login_at?: string
  created_at: string
  updated_at: string
}

export interface LoginCredentials {
  username: string
  password: string
}

export interface TokenResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  token_type: 'Bearer'
}
