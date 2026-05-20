import { apiRequest } from '../request'
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

  complete: (id: number, remark: string) => {
    return apiRequest({
      url: `/v1/tasks/${id}/complete`,
      method: 'POST',
      data: { remark },
    })
  },

  reportIssue: (id: number, issueDesc: string) => {
    return apiRequest({
      url: `/v1/tasks/${id}/issue`,
      method: 'POST',
      data: { issue_desc: issueDesc },
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
