import { apiRequest } from '../request'
import type { User } from '@/types'

export const userApi = {
  getList: (params?: { page?: number; page_size?: number; role?: string }) => {
    return apiRequest<{ items: User[]; total: number }>({
      url: '/v1/users',
      method: 'GET',
      params,
    })
  },

  getById: (id: number) => {
    return apiRequest<User>({
      url: `/v1/users/${id}`,
      method: 'GET',
    })
  },

  create: (data: {
    username: string
    password: string
    name: string
    email: string
    role: string
    phone?: string
    department?: string
  }) => {
    return apiRequest<User>({
      url: '/v1/users',
      method: 'POST',
      data: {
        ...data,
        real_name: data.name,
      },
    })
  },

  update: (id: number, data: Partial<User>) => {
    return apiRequest<User>({
      url: `/v1/users/${id}`,
      method: 'PUT',
      data,
    })
  },

  delete: (id: number) => {
    return apiRequest<void>({
      url: `/v1/users/${id}`,
      method: 'DELETE',
    })
  },

  resetPassword: (id: number, newPassword: string) => {
    return apiRequest<void>({
      url: `/v1/users/${id}/reset-password`,
      method: 'POST',
      data: { password: newPassword },
    })
  },
}
