import { apiRequest } from '../request'
import type { Task, TaskAction } from '@/types'

export const taskApi = {
  getMyTasks: (params?: { status?: string }) => {
    return apiRequest<Task[]>({
      url: '/api/v1/tasks/my',
      method: 'GET',
      params,
    })
  },

  getById: (id: number) => {
    return apiRequest<Task>({
      url: `/api/v1/tasks/${id}`,
      method: 'GET',
    })
  },

  executeAction: (id: number, action: TaskAction) => {
    return apiRequest<Task>({
      url: `/api/v1/tasks/${id}/action`,
      method: 'POST',
      data: action,
    })
  },

  assign: (taskId: number, userId: number) => {
    return apiRequest<Task>({
      url: `/api/v1/tasks/${taskId}/assign`,
      method: 'POST',
      data: { user_id: userId },
    })
  },
}
