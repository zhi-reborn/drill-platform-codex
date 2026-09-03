import { beforeEach, describe, expect, it, vi } from 'vitest'
import { authApi } from './auth'
import { dashboardApi } from './dashboard'

const apiRequest = vi.hoisted(() => vi.fn())
vi.mock('../request', () => ({ apiRequest }))

describe('dashboard and heartbeat API contracts', () => {
  beforeEach(() => apiRequest.mockReset())

  it.each([
    ['getTeam', '/v1/dashboard/team', { team_online_count: null, team_total_count: 8 }],
    ['getMyTemplates', '/v1/dashboard/my-templates', { my_template_count: 0 }],
    ['getStepDuration', '/v1/dashboard/step-duration', { avg_step_duration_seconds: null }],
  ] as const)('%s uses its independent endpoint and preserves empty values', async (method, url, data) => {
    const signal = new AbortController().signal
    apiRequest.mockResolvedValue(data)
    await expect(dashboardApi[method](signal)).resolves.toEqual(data)
    expect(apiRequest).toHaveBeenCalledWith({ url, method: 'GET', signal, silentError: true })
  })

  it('posts a bounded cancellable heartbeat with silent errors', async () => {
    const signal = new AbortController().signal
    await authApi.heartbeat(signal)
    expect(apiRequest).toHaveBeenCalledWith({ url: '/v1/auth/heartbeat', method: 'POST', signal, silentError: true, timeout: 10000 })
  })
})
