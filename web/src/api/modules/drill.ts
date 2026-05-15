import { apiRequest } from '../request'
import type { DrillInstance, StepInstance, StepLog } from '@/types'

export const drillApi = {
  getList: (params?: { page?: number; page_size?: number; status?: string }) => {
    return apiRequest<{ list: DrillInstance[]; total: number }>({
      url: '/v1/drills',
      method: 'GET',
      params,
    })
  },

  getById: (id: number) => {
    return apiRequest<DrillInstance>({
      url: `/v1/drills/${id}`,
      method: 'GET',
    })
  },

  create: (data: {
    template_id: number
    name: string
    description: string
  }) => {
    return apiRequest<DrillInstance>({
      url: '/v1/drills',
      method: 'POST',
      data,
    })
  },

  start: (id: number) => {
    return apiRequest<DrillInstance>({
      url: `/v1/drills/${id}/start`,
      method: 'POST',
    })
  },

  pause: (id: number) => {
    return apiRequest<DrillInstance>({
      url: `/v1/drills/${id}/pause`,
      method: 'POST',
    })
  },

  resume: (id: number) => {
    return apiRequest<DrillInstance>({
      url: `/v1/drills/${id}/resume`,
      method: 'POST',
    })
  },

  terminate: (id: number) => {
    return apiRequest<DrillInstance>({
      url: `/v1/drills/${id}/terminate`,
      method: 'POST',
    })
  },

  getSteps: (drillId: number) => {
    return apiRequest<StepInstance[]>({
      url: `/v1/drills/${drillId}/steps`,
      method: 'GET',
    })
  },

  getStepLogs: (stepId: number) => {
    return apiRequest<StepLog[]>({
      url: `/v1/steps/${stepId}/logs`,
      method: 'GET',
    })
  },

  getLogs: (drillId: number) => {
    return apiRequest<any[]>({
      url: `/v1/drills/${drillId}/logs`,
      method: 'GET',
    })
  },
}
