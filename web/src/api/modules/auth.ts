import { apiRequest } from '../request'
import type { LoginCredentials, TokenResponse, User } from '@/types'

export const authApi = {
  login: (credentials: LoginCredentials) => {
    return apiRequest<TokenResponse>({
      url: '/v1/auth/login',
      method: 'POST',
      data: credentials,
    })
  },

  logout: () => {
    return apiRequest<void>({
      url: '/v1/auth/logout',
      method: 'POST',
    })
  },

  getCurrentUser: () => {
    return apiRequest<User>({
      url: '/v1/auth/me',
      method: 'GET',
    })
  },

  refreshToken: () => {
    return apiRequest<TokenResponse>({
      url: '/v1/auth/refresh',
      method: 'POST',
    })
  },

  devUsers: () => {
    return apiRequest<Array<{
      id: number
      username: string
      real_name: string
      role: string
      department: string
      status: number
    }>>({
      url: '/v1/auth/dev-users',
      method: 'GET',
    })
  },
}
