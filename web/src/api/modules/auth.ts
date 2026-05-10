import { apiRequest } from '../request'
import type { LoginCredentials, TokenResponse, User } from '@/types'

export const authApi = {
  login: (credentials: LoginCredentials) => {
    return apiRequest<TokenResponse>({
      url: '/api/v1/auth/login',
      method: 'POST',
      data: credentials,
    })
  },

  logout: () => {
    return apiRequest<void>({
      url: '/api/v1/auth/logout',
      method: 'POST',
    })
  },

  getCurrentUser: () => {
    return apiRequest<User>({
      url: '/api/v1/auth/current',
      method: 'GET',
    })
  },

  refreshToken: () => {
    return apiRequest<TokenResponse>({
      url: '/api/v1/auth/refresh',
      method: 'POST',
    })
  },
}
