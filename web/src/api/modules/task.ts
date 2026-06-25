import { apiRequest, mutationRequest } from '../request'
import type { MutationResult } from '../request'
import type { StepInstance } from '@/types/instance'

export const taskApi = {
  getMyTasks: () => {
    return apiRequest<StepInstance[]>({
      url: '/v1/tasks/my',
      method: 'GET',
    })
  },

  getById: (id: number) => {
    return apiRequest<StepInstance>({
      url: `/v1/tasks/${id}`,
      method: 'GET',
    })
  },

  start: (id: number): Promise<MutationResult<void>> => {
    return mutationRequest<void>({
      url: `/v1/tasks/${id}/start`,
      method: 'POST',
      actionId: `task:${id}:start`,
    })
  },

  complete: (id: number, remark: string): Promise<MutationResult<void>> => {
    return mutationRequest<void>({
      url: `/v1/tasks/${id}/complete`,
      method: 'POST',
      data: { remark },
      actionId: `task:${id}:complete`,
    })
  },

  reportIssue: (id: number, issueDesc: string): Promise<MutationResult<void>> => {
    return mutationRequest<void>({
      url: `/v1/tasks/${id}/issue`,
      method: 'POST',
      data: { issue_desc: issueDesc },
      actionId: `task:${id}:report-issue`,
    })
  },

  assign: (taskId: number, userId: number) => {
    return apiRequest<StepInstance>({
      url: `/v1/tasks/${taskId}/assign`,
      method: 'POST',
      data: { user_id: userId },
    })
  },
}
