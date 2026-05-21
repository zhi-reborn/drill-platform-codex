import { apiRequest } from '../request'
import type { DrillTemplate } from '@/types'

export const templateApi = {
  getList: (params?: { category?: string; page?: number; page_size?: number }) => {
    return apiRequest<{ list: DrillTemplate[]; total: number }>({
      url: '/v1/templates',
      method: 'GET',
      params,
    })
  },

  getById: (id: number) => {
    return apiRequest<DrillTemplate>({
      url: `/v1/templates/${id}`,
      method: 'GET',
    })
  },

  create: (data: {
    name: string
    category: string
    description?: string
    steps?: Array<{
      name: string
      seq: number
      step_type: string
      timeout_minutes?: number
      guide_content?: string
      is_blocking?: number
      default_assignee_role?: string
    }>
  }) => {
    return apiRequest<DrillTemplate>({
      url: '/v1/templates',
      method: 'POST',
      data,
    })
  },

  update: (id: number, data: {
    name: string
    category: string
    description?: string
  }) => {
    return apiRequest<void>({
      url: `/v1/templates/${id}`,
      method: 'PUT',
      data,
    })
  },

  delete: (id: number) => {
    return apiRequest<void>({
      url: `/v1/templates/${id}`,
      method: 'DELETE',
    })
  },

  clone: (id: number) => {
    return apiRequest<DrillTemplate>({
      url: `/v1/templates/${id}/clone`,
      method: 'POST',
    })
  },

  getCategories: () => {
    return apiRequest<Array<{
      id: number
      value: string
      label: string
      sort_order: number
      tag_type: string
    }>>({
      url: '/v1/template-categories',
      method: 'GET',
    })
  },

  saveCategories: (categories: Array<{
    id?: number
    value: string
    label: string
    tag_type: string
  }>) => {
    return apiRequest<void>({
      url: '/v1/template-categories',
      method: 'POST',
      data: categories,
    })
  },

  toggleStatus: (id: number) => {
    return apiRequest<void>({
      url: `/v1/templates/${id}/toggle-status`,
      method: 'POST',
    })
  },

  updateSteps: (id: number, steps: Array<{
    name: string
    seq: number
    step_type: string
    timeout_minutes?: number
    guide_content?: string
    is_blocking?: number
    default_assignee_role?: string
    executor_team?: string
  }>) => {
    return apiRequest<void>({
      url: `/v1/templates/${id}/steps`,
      method: 'PUT',
      data: { steps },
    })
  },

  updateStep: (id: number, stepId: number, data: {
    name: string
    seq: number
    step_type: string
    timeout_minutes?: number
    guide_content?: string
    is_blocking?: number
    default_assignee_role?: string
    executor_team?: string
    phase?: string
    phase_step?: string
    estimated_duration_minutes?: number
    estimated_start_offset?: number
    task_name?: string
    sub_task?: string
    responsible_department?: string
    responsible_person?: string
    executor?: string
    reviewer?: string
  }) => {
    return apiRequest<void>({
      url: `/v1/templates/${id}/steps/${stepId}`,
      method: 'PUT',
      data,
    })
  },
}
