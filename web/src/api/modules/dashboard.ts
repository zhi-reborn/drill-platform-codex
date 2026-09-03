import { apiRequest } from '../request'

export interface TeamMetric {
  team_online_count: number | null
  team_total_count: number
}

export interface MyTemplatesMetric {
  my_template_count: number
}

export interface StepDurationMetric {
  avg_step_duration_seconds: number | null
}

export const dashboardApi = {
  getTeam: (signal?: AbortSignal) => apiRequest<TeamMetric>({
    url: '/v1/dashboard/team',
    method: 'GET',
    signal,
    silentError: true,
  }),

  getMyTemplates: (signal?: AbortSignal) => apiRequest<MyTemplatesMetric>({
    url: '/v1/dashboard/my-templates',
    method: 'GET',
    signal,
    silentError: true,
  }),

  getStepDuration: (signal?: AbortSignal) => apiRequest<StepDurationMetric>({
    url: '/v1/dashboard/step-duration',
    method: 'GET',
    signal,
    silentError: true,
  }),
}
