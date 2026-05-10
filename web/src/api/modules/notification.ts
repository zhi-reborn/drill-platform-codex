import { apiRequest } from '../request'
import type { Notification } from '@/types'

export const notificationApi = {
  getList: (params?: { page?: number; page_size?: number; unread_only?: boolean }) => {
    return apiRequest<{ items: Notification[]; total: number }>({
      url: '/v1/notifications',
      method: 'GET',
      params,
    })
  },

  markAsRead: (id: number) => {
    return apiRequest<Notification>({
      url: `/v1/notifications/${id}/read`,
      method: 'POST',
    })
  },

  markAllAsRead: () => {
    return apiRequest<void>({
      url: '/v1/notifications/read-all',
      method: 'POST',
    })
  },

  delete: (id: number) => {
    return apiRequest<void>({
      url: `/v1/notifications/${id}`,
      method: 'DELETE',
    })
  },
}
