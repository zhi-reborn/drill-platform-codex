import { apiRequest } from '../request'
import type { DashboardData, DrillInstance } from '@/types'

export const displayApi = {
  getDashboard: () => {
    return apiRequest<DashboardData>({
      url: '/api/v1/display/dashboard',
      method: 'GET',
    })
  },

  getActiveDrills: () => {
    return apiRequest<DrillInstance[]>({
      url: '/api/v1/display/active-drills',
      method: 'GET',
    })
  },

  getDrillDetail: (id: number) => {
    return apiRequest<DrillInstance>({
      url: `/api/v1/display/drills/${id}`,
      method: 'GET',
    })
  },
}
