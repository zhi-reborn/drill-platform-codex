import { apiRequest } from '../request'

export const configApi = {
  getSystemConfig: () => {
    return apiRequest<Record<string, unknown>>({
      url: '/v1/config/system',
      method: 'GET',
    })
  },

  updateSystemConfig: (data: Record<string, unknown>) => {
    return apiRequest<Record<string, unknown>>({
      url: '/v1/config/system',
      method: 'PUT',
      data,
    })
  },
}
