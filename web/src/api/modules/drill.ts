import { apiRequest } from '../request'
import type { DrillInstance, StepInstance } from '@/types/instance'

export const drillApi = {
  getList: (params?: { page?: number; page_size?: number; status?: string }) => {
    return apiRequest<{ list: DrillInstance[]; total: number; page: number; page_size: number }>({
      url: '/v1/drills',
      method: 'GET',
      params: params as any,
    })
  },

  getDetail: (id: number) => {
    return apiRequest<DrillInstance>({
      url: `/v1/drills/${id}`,
      method: 'GET',
    })
  },

  getSteps: (id: number) => {
    return apiRequest<StepInstance[]>({
      url: `/v1/drills/${id}/steps`,
      method: 'GET',
    })
  },

  getLogs: (id: number) => {
    return apiRequest<any[]>({
      url: `/v1/drills/${id}/logs`,
      method: 'GET',
    })
  },

  create: (data: { template_id: number; name: string }) => {
    return apiRequest<DrillInstance>({
      url: '/v1/drills',
      method: 'POST',
      data,
    })
  },

  start: (id: number) => {
    return apiRequest<void>({
      url: `/v1/drills/${id}/start`,
      method: 'POST',
    })
  },

  pause: (id: number) => {
    return apiRequest<void>({
      url: `/v1/drills/${id}/pause`,
      method: 'POST',
    })
  },

  resume: (id: number) => {
    return apiRequest<void>({
      url: `/v1/drills/${id}/resume`,
      method: 'POST',
    })
  },

  terminate: (id: number) => {
    return apiRequest<void>({
      url: `/v1/drills/${id}/terminate`,
      method: 'POST',
    })
  },

  delete: (id: number) => {
    return apiRequest<void>({
      url: `/v1/drills/${id}`,
      method: 'DELETE',
    })
  },

  skipStep: (drillId: number, stepId: number, remark?: string) => {
    return apiRequest<void>({
      url: `/v1/drills/${drillId}/steps/skip`,
      method: 'POST',
      data: { step_id: stepId, remark },
    })
  },

  completeStep: (drillId: number, stepId: number, remark?: string) => {
    return apiRequest<void>({
      url: `/v1/drills/${drillId}/steps/complete`,
      method: 'POST',
      data: { step_id: stepId, remark },
    })
  },

  forceCompleteStep: (drillId: number, stepId: number, remark?: string) => {
    return apiRequest<void>({
      url: `/v1/drills/${drillId}/steps/force-complete`,
      method: 'POST',
      data: { step_id: stepId, remark },
    })
  },

  resumeTask: (drillId: number, stepId: number, remark?: string) => {
    return apiRequest<void>({
      url: `/v1/drills/${drillId}/steps/resume-task`,
      method: 'POST',
      data: { step_id: stepId, remark },
    })
  },

  assignStep: (drillId: number, stepId: number, userIds: number[]) => {
    return apiRequest<void>({
      url: `/v1/drills/${drillId}/steps/assign`,
      method: 'POST',
      data: { step_id: stepId, user_ids: userIds },
    })
  },
}
