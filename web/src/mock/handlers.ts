import type { AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import axios from 'axios'

import usersData from './data/users.json'
import templatesData from './data/templates.json'
import instancesData from './data/instances.json'
import stepsData from './data/steps.json'
import tasksData from './data/tasks.json'
import notificationsData from './data/notifications.json'
import dashboardData from './data/dashboard.json'

interface MockSession {
  currentUserId?: number
}

const mockSession: MockSession = {}

function getUserIdFromToken(token: string): number | undefined {
  // mock token 格式：mock_token_{userId}_{timestamp}
  const match = token.match(/^mock_token_(\d+)_/)
  if (match) {
    return Number(match[1])
  }
  return undefined
}

function getCurrentUserIdFromHeaders(config: InternalAxiosRequestConfig): number | undefined {
  const authHeader = config.headers?.Authorization as string | undefined
  if (authHeader && authHeader.startsWith('Bearer ')) {
    const token = authHeader.replace('Bearer ', '')
    return getUserIdFromToken(token)
  }
  return mockSession.currentUserId
}

function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

function createResponse<T>(data: T): AxiosResponse<{ code: number; message: string; data: T }> {
  return {
    data: {
      code: 0,
      message: 'success',
      data,
    },
    status: 200,
    statusText: 'OK',
    headers: {},
    config: {} as InternalAxiosRequestConfig,
  }
}

function createErrorResponse(message: string, status = 400): AxiosResponse {
  return {
    data: {
      code: status,
      message,
      data: null,
    },
    status,
    statusText: status === 401 ? 'Unauthorized' : status === 403 ? 'Forbidden' : 'Bad Request',
    headers: {},
    config: {} as InternalAxiosRequestConfig,
  }
}

function findUserByUsername(username: string) {
  return usersData.find((u) => u.username === username)
}

function findUserById(id: number) {
  return usersData.find((u) => u.id === id)
}

export function setupMock(): void {
  console.log('[Mock API] Setting up mock adapter...')
  const mockAdapter = async (config: InternalAxiosRequestConfig): Promise<AxiosResponse> => {
    await delay(200 + Math.random() * 300)

    const url = config.url || ''
    const baseURL = config.baseURL || ''
    const fullUrl = baseURL + url
    const method = (config.method || 'GET').toUpperCase()

    console.log('[Mock API] Intercepted:', { fullUrl, url, baseURL, method, startsWithV1: url.startsWith('/v1/') })

    if (!url.startsWith('/v1/')) {
      console.log('[Mock API] Passing through (not /v1/):', url)
      // 不是 /v1/ 的请求，让原始 axios 处理（会发 HTTP 请求）
      return axios.request(config)
    }

    console.log('[Mock API] Handling mock request:', url)
    const mockResponse = await handleMockRequest(url, method, config)
    return mockResponse
  }

  axios.defaults.adapter = mockAdapter
  console.log('[Mock API] Mock adapter installed')
}

async function handleMockRequest(
  url: string,
  method: string,
  config: InternalAxiosRequestConfig
): Promise<AxiosResponse> {
  console.log('[Mock API] Request:', { url, method, userId: getCurrentUserIdFromHeaders(config) })
  const params = config.params as Record<string, unknown> | undefined
  const data = config.data as Record<string, unknown> | undefined

  if (/^\/v1\/auth\/login$/.test(url) && method === 'POST') {
    const { username, password } = data as { username: string; password: string }
    const user = findUserByUsername(username)

    if (!user) {
      return createErrorResponse('用户不存在', 401)
    }

    if (!password || password.length < 1) {
      return createErrorResponse('密码错误', 401)
    }

    mockSession.currentUserId = user.id

    return createResponse({
      access_token: `mock_token_${user.id}_${Date.now()}`,
      refresh_token: `mock_refresh_${user.id}`,
      expires_in: 7200,
      token_type: 'Bearer',
    })
  }

  if (/^\/v1\/auth\/logout$/.test(url) && method === 'POST') {
    mockSession.currentUserId = undefined
    return createResponse(null)
  }

  if (/^\/v1\/auth\/current$/.test(url) && method === 'GET') {
    const currentUserId = getCurrentUserIdFromHeaders(config)
    if (!currentUserId) {
      return createErrorResponse('未登录', 401)
    }
    const user = findUserById(currentUserId)
    if (!user) {
      return createErrorResponse('用户不存在', 404)
    }
    return createResponse(user)
  }

  if (/^\/v1\/auth\/refresh$/.test(url) && method === 'POST') {
    const currentUserId = getCurrentUserIdFromHeaders(config)
    if (!currentUserId) {
      return createErrorResponse('未登录', 401)
    }
    const user = findUserById(currentUserId)
    return createResponse({
      access_token: `mock_token_${user?.id}_${Date.now()}`,
      refresh_token: `mock_refresh_${user?.id}`,
      expires_in: 7200,
      token_type: 'Bearer',
    })
  }

  if (/^\/v1\/drills$/.test(url) && method === 'GET') {
    let filtered = [...instancesData]

    if (params?.status) {
      filtered = filtered.filter((i) => i.status === params.status)
    }

    const page = Number(params?.page) || 1
    const pageSize = Number(params?.page_size) || 10
    const start = (page - 1) * pageSize
    const end = start + pageSize
    const paginated = filtered.slice(start, end)

    return createResponse({
      items: paginated,
      total: filtered.length,
    })
  }

  if (/^\/v1\/drills\/\d+$/.test(url) && method === 'GET') {
    const id = Number(url.split('/').pop())
    const drill = instancesData.find((i) => i.id === id)
    if (!drill) {
      return createErrorResponse('演练实例不存在', 404)
    }
    return createResponse(drill)
  }

  if (/^\/v1\/drills$/.test(url) && method === 'POST') {
    const newDrill = {
      id: Math.max(...instancesData.map((i) => i.id)) + 1,
      template_id: Number(data?.template_id),
      template_name: templatesData.find((t) => t.id === data?.template_id)?.name || 'Unknown',
      name: data?.name || 'New Drill',
      description: data?.description || '',
      status: 'pending',
      created_by: mockSession.currentUserId || 1,
      created_by_name: findUserById(mockSession.currentUserId || 1)?.name || 'Unknown',
      current_step_index: 0,
      total_steps: 0,
      completed_steps: 0,
      created_at: new Date().toISOString(),
    }
    return createResponse(newDrill)
  }

  if (/^\/v1\/drills\/\d+\/start$/.test(url) && method === 'POST') {
    const id = Number(url.split('/')[3])
    const drill = instancesData.find((i) => i.id === id)
    if (!drill) {
      return createErrorResponse('演练实例不存在', 404)
    }
    return createResponse({ ...drill, status: 'running', started_at: new Date().toISOString() })
  }

  if (/^\/v1\/drills\/\d+\/pause$/.test(url) && method === 'POST') {
    const id = Number(url.split('/')[3])
    const drill = instancesData.find((i) => i.id === id)
    if (!drill) {
      return createErrorResponse('演练实例不存在', 404)
    }
    return createResponse({ ...drill, status: 'paused', paused_at: new Date().toISOString() })
  }

  if (/^\/v1\/drills\/\d+\/resume$/.test(url) && method === 'POST') {
    const id = Number(url.split('/')[3])
    const drill = instancesData.find((i) => i.id === id)
    if (!drill) {
      return createErrorResponse('演练实例不存在', 404)
    }
    return createResponse({ ...drill, status: 'running' })
  }

  if (/^\/v1\/drills\/\d+\/terminate$/.test(url) && method === 'POST') {
    const id = Number(url.split('/')[3])
    const drill = instancesData.find((i) => i.id === id)
    if (!drill) {
      return createErrorResponse('演练实例不存在', 404)
    }
    return createResponse({ ...drill, status: 'terminated' })
  }

  if (/^\/v1\/drills\/\d+\/steps$/.test(url) && method === 'GET') {
    const drillId = Number(url.split('/')[3])
    const steps = stepsData.filter((s) => s.drill_id === drillId)
    return createResponse(steps)
  }

  if (/^\/v1\/steps\/\d+\/logs$/.test(url) && method === 'GET') {
    const stepId = Number(url.split('/')[3])
    const logs = stepsData
      .filter((s) => s.id === stepId)
      .map((s) => ({
        id: s.id,
        step_instance_id: s.id,
        action: s.status === 'completed' ? 'complete' : s.status === 'issue' ? 'issue' : 'skip',
        operator_id: s.assignee_id || 1,
        operator_name: s.assignee_name || 'System',
        comment: s.error_message || s.result_json || '',
        created_at: s.completed_at || s.started_at || new Date().toISOString(),
      }))
    return createResponse(logs)
  }

  if (/^\/v1\/templates$/.test(url) && method === 'GET') {
    let filtered = [...templatesData]
    if (params?.category) {
      filtered = filtered.filter((t) => t.category === params.category)
    }
    return createResponse(filtered)
  }

  if (/^\/v1\/templates\/\d+$/.test(url) && method === 'GET') {
    const id = Number(url.split('/').pop())
    const template = templatesData.find((t) => t.id === id)
    if (!template) {
      return createErrorResponse('模板不存在', 404)
    }
    return createResponse(template)
  }

  if (/^\/v1\/templates$/.test(url) && method === 'POST') {
    const newTemplate = {
      id: Math.max(...templatesData.map((t) => t.id)) + 1,
      name: data?.name || 'New Template',
      category: data?.category || 'disaster_recovery',
      description: data?.description || '',
      version: '1.0.0',
      created_by: mockSession.currentUserId || 1,
      created_by_name: findUserById(mockSession.currentUserId || 1)?.name || 'Unknown',
      status: 'draft',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      steps: data?.steps || [],
    }
    return createResponse(newTemplate)
  }

  if (/^\/v1\/templates\/\d+\/publish$/.test(url) && method === 'POST') {
    const id = Number(url.split('/')[3])
    const template = templatesData.find((t) => t.id === id)
    if (!template) {
      return createErrorResponse('模板不存在', 404)
    }
    return createResponse({ ...template, status: 'published', updated_at: new Date().toISOString() })
  }

  if (/^\/v1\/tasks\/my$/.test(url) && method === 'GET') {
    const currentUserId = getCurrentUserIdFromHeaders(config)
    if (!currentUserId) {
      return createErrorResponse('未登录', 401)
    }
    let filtered = tasksData.filter((t) => t.assigned_to === currentUserId)
    if (params?.status) {
      filtered = filtered.filter((t) => t.status === params.status)
    }
    return createResponse(filtered)
  }

  if (/^\/v1\/tasks\/\d+$/.test(url) && method === 'GET') {
    const id = Number(url.split('/').pop())
    const task = tasksData.find((t) => t.id === id)
    if (!task) {
      return createErrorResponse('任务不存在', 404)
    }
    return createResponse(task)
  }

  if (/^\/v1\/tasks\/\d+\/action$/.test(url) && method === 'POST') {
    const id = Number(url.split('/')[3])
    const task = tasksData.find((t) => t.id === id)
    if (!task) {
      return createErrorResponse('任务不存在', 404)
    }
    const action = data?.action as string
    const taskWithResult = task as typeof task & { result?: string }
    const updatedTask = {
      ...task,
      status:
        action === 'complete'
          ? 'completed'
          : action === 'issue'
            ? 'issued'
            : action === 'skip'
              ? 'skipped'
              : task.status,
      result: (data?.result as string) || taskWithResult.result,
      error_message: (data?.error_message as string) || task.error_message,
      updated_at: new Date().toISOString(),
    }
    return createResponse(updatedTask)
  }

  if (/^\/v1\/users$/.test(url) && method === 'GET') {
    let filtered = [...usersData]
    if (params?.role) {
      filtered = filtered.filter((u) => u.role === params.role)
    }
    const page = Number(params?.page) || 1
    const pageSize = Number(params?.page_size) || 10
    const start = (page - 1) * pageSize
    const end = start + pageSize
    const paginated = filtered.slice(start, end)
    return createResponse({
      items: paginated,
      total: filtered.length,
    })
  }

  if (/^\/v1\/users\/\d+$/.test(url) && method === 'GET') {
    const id = Number(url.split('/').pop())
    const user = usersData.find((u) => u.id === id)
    if (!user) {
      return createErrorResponse('用户不存在', 404)
    }
    return createResponse(user)
  }

  if (/^\/v1\/notifications$/.test(url) && method === 'GET') {
    const currentUserId = getCurrentUserIdFromHeaders(config)
    console.log('[Mock API] /v1/notifications GET:', {
      currentUserId,
      authHeader: config.headers?.Authorization,
    })
    if (!currentUserId) {
      return createErrorResponse('未登录', 401)
    }
    let filtered = notificationsData.filter((n) => n.user_id === currentUserId)
    console.log('[Mock API] Filtered notifications for user', currentUserId, ':', filtered.length, 'items')
    if (params?.unread_only) {
      filtered = filtered.filter((n) => !n.is_read)
    }
    const page = Number(params?.page) || 1
    const pageSize = Number(params?.page_size) || 10
    const start = (page - 1) * pageSize
    const end = start + pageSize
    const paginated = filtered.slice(start, end)
    return createResponse({
      items: paginated,
      total: filtered.length,
    })
  }

  if (/^\/v1\/notifications\/\d+\/read$/.test(url) && method === 'POST') {
    const id = Number(url.split('/')[3])
    const notification = notificationsData.find((n) => n.id === id)
    if (!notification) {
      return createErrorResponse('通知不存在', 404)
    }
    return createResponse({ ...notification, is_read: true })
  }

  if (/^\/v1\/notifications\/read-all$/.test(url) && method === 'POST') {
    const currentUserId = getCurrentUserIdFromHeaders(config)
    if (!currentUserId) {
      return createErrorResponse('未登录', 401)
    }
    return createResponse(null)
  }

  if (/^\/v1\/display\/dashboard$/.test(url) && method === 'GET') {
    return createResponse(dashboardData)
  }

  if (/^\/v1\/display\/active-drills$/.test(url) && method === 'GET') {
    const activeDrills = instancesData.filter((i) => i.status === 'running' || i.status === 'paused')
    return createResponse(activeDrills)
  }

  if (/^\/v1\/display\/drills\/\d+$/.test(url) && method === 'GET') {
    const id = Number(url.split('/').pop())
    const drill = instancesData.find((i) => i.id === id)
    if (!drill) {
      return createErrorResponse('演练实例不存在', 404)
    }
    return createResponse(drill)
  }

  if (/^\/v1\/config\/system$/.test(url) && method === 'GET') {
    return createResponse({
      site_name: '生产演练平台',
      version: '1.0.0',
      maintenance_mode: false,
    })
  }

  return createErrorResponse('API 未实现', 404)
}
