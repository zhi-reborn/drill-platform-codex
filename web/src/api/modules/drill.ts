import { apiRequest, mutationRequest } from '../request'
import type { MutationResult } from '../request'
import type { DrillInstance, StepInstance } from '@/types/instance'

export const drillApi = {
  getList: (params?: { page?: number; page_size?: number; status?: string; keyword?: string }) => {
    return apiRequest<{ list: DrillInstance[]; total: number; page: number; page_size: number }>({
      url: '/v1/drills',
      method: 'GET',
      params: params as any,
    })
  },

  getDetail: (id: number) => {
    if (!Number.isFinite(id) || id <= 0) {
      return Promise.reject(new Error('无效的演练 ID'))
    }
    return apiRequest<DrillInstance>({
      url: `/v1/drills/${id}`,
      method: 'GET',
    })
  },

  getSteps: (id: number) => {
    if (!Number.isFinite(id) || id <= 0) {
      return Promise.reject(new Error('无效的演练 ID'))
    }
    return apiRequest<StepInstance[]>({
      url: `/v1/drills/${id}/steps`,
      method: 'GET',
    })
  },

  getLogs: (id: number, limit?: number) => {
    if (!Number.isFinite(id) || id <= 0) {
      return Promise.reject(new Error('无效的演练 ID'))
    }
    return apiRequest<any[]>({
      url: `/v1/drills/${id}/logs`,
      method: 'GET',
      params: limit ? { limit } : undefined,
    })
  },

  create: (data: { template_id: number; name: string; description?: string; planned_start?: string }) => {
    return apiRequest<DrillInstance>({
      url: '/v1/drills',
      method: 'POST',
      data,
    })
  },

  start: (id: number): Promise<MutationResult<void>> => {
    return mutationRequest<void>({
      url: `/v1/drills/${id}/start`,
      method: 'POST',
      actionId: `drill:${id}:start`,
    })
  },

  pause: (id: number): Promise<MutationResult<void>> => {
    return mutationRequest<void>({
      url: `/v1/drills/${id}/pause`,
      method: 'POST',
      actionId: `drill:${id}:pause`,
    })
  },

  resume: (id: number): Promise<MutationResult<void>> => {
    return mutationRequest<void>({
      url: `/v1/drills/${id}/resume`,
      method: 'POST',
      actionId: `drill:${id}:resume`,
    })
  },

  terminate: (id: number): Promise<MutationResult<void>> => {
    return mutationRequest<void>({
      url: `/v1/drills/${id}/terminate`,
      method: 'POST',
      actionId: `drill:${id}:terminate`,
    })
  },

  delete: (id: number): Promise<MutationResult<void>> => {
    return mutationRequest<void>({
      url: `/v1/drills/${id}`,
      method: 'DELETE',
      actionId: `drill:${id}:delete`,
    })
  },

  skipStep: (drillId: number, stepId: number, remark?: string): Promise<MutationResult<void>> => {
    return mutationRequest<void>({
      url: `/v1/drills/${drillId}/steps/skip`,
      method: 'POST',
      data: { step_id: stepId, remark },
      actionId: `drill:${drillId}:step:${stepId}:skip`,
    })
  },

  startStep: (drillId: number, stepId: number, remark?: string): Promise<MutationResult<void>> => {
    return mutationRequest<void>({
      url: `/v1/drills/${drillId}/steps/start`,
      method: 'POST',
      data: { step_id: stepId, remark },
      actionId: `drill:${drillId}:step:${stepId}:start`,
    })
  },

  completeStep: (drillId: number, stepId: number, remark?: string): Promise<MutationResult<void>> => {
    return mutationRequest<void>({
      url: `/v1/drills/${drillId}/steps/complete`,
      method: 'POST',
      data: { step_id: stepId, remark },
      actionId: `drill:${drillId}:step:${stepId}:complete`,
    })
  },

  forceCompleteStep: (drillId: number, stepId: number, remark?: string): Promise<MutationResult<void>> => {
    return mutationRequest<void>({
      url: `/v1/drills/${drillId}/steps/force-complete`,
      method: 'POST',
      data: { step_id: stepId, remark },
      actionId: `drill:${drillId}:step:${stepId}:force-complete`,
    })
  },

  resumeTask: (drillId: number, stepId: number, remark?: string): Promise<MutationResult<void>> => {
    return mutationRequest<void>({
      url: `/v1/drills/${drillId}/steps/resume-task`,
      method: 'POST',
      data: { step_id: stepId, remark },
      actionId: `drill:${drillId}:step:${stepId}:resume-task`,
    })
  },

  assignStep: (drillId: number, stepId: number, userIds: number[]): Promise<MutationResult<void>> => {
    return mutationRequest<void>({
      url: `/v1/drills/${drillId}/steps/assign`,
      method: 'POST',
      data: { step_id: stepId, user_ids: userIds },
      actionId: `drill:${drillId}:step:${stepId}:assign`,
    })
  },

  updateStepInfo: (
    drillId: number,
    stepId: number,
    payload: {
      attributes?: Record<string, string>
      remark?: string
    }
  ): Promise<MutationResult<void>> => {
    return mutationRequest<void>({
      url: `/v1/drills/${drillId}/steps/info`,
      method: 'PUT',
      data: { step_id: stepId, ...payload },
      actionId: `drill:${drillId}:step:${stepId}:update-info`,
    })
  },
}
