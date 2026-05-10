import { apiRequest } from '../request'
import type { DrillTemplate } from '@/types'

export const templateApi = {
  getList: (params?: { category?: string }) => {
    return apiRequest<DrillTemplate[]>({
      url: '/api/v1/templates',
      method: 'GET',
      params,
    })
  },

  getById: (id: number) => {
    return apiRequest<DrillTemplate>({
      url: `/api/v1/templates/${id}`,
      method: 'GET',
    })
  },

  create: (data: {
    name: string
    category: string
    description: string
    steps: Array<{
      name: string
      description: string
      step_type: string
      timeout_seconds: number
      order_index: number
    }>
  }) => {
    return apiRequest<DrillTemplate>({
      url: '/api/v1/templates',
      method: 'POST',
      data,
    })
  },

  update: (id: number, data: Partial<DrillTemplate>) => {
    return apiRequest<DrillTemplate>({
      url: `/api/v1/templates/${id}`,
      method: 'PUT',
      data,
    })
  },

  delete: (id: number) => {
    return apiRequest<void>({
      url: `/api/v1/templates/${id}`,
      method: 'DELETE',
    })
  },

  publish: (id: number) => {
    return apiRequest<DrillTemplate>({
      url: `/api/v1/templates/${id}/publish`,
      method: 'POST',
    })
  },
}
