import axios from 'axios'
import type { InternalAxiosRequestConfig, AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import type { ApiResponse } from '@/types'
import type { FlowCommand } from '@/types/flowCommand'
import { getKeyForAction, clearKey } from './idempotency'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 60000,
})

request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const auth = localStorage.getItem('drill_auth')
    if (auth) {
      const { access_token } = JSON.parse(auth)
      config.headers.Authorization = `Bearer ${access_token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

request.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    // 解包后端返回的数据
    const res = response.data
    if (res.code !== 0) {
      ElMessage.error(res.message || '请求失败')
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    // 202 Accepted: 异步命令已受理，返回 pending 标记与命令对象
    if (response.status === 202) {
      return { pending: true, command: res.data } as unknown as AxiosResponse
    }
    return res.data as AxiosResponse
  },
  (error) => {
    const status = error.response?.status
    const data = error.response?.data
    const message = data?.message
    if (message) {
      error.message = message
    }
    switch (status) {
      case 401:
        // 不要对登录 API 的 401 执行跳转（已在登录页面）
        if (!error.config?.url?.includes('/auth/login')) {
          localStorage.removeItem('drill_auth')
          window.location.href = '/login'
        }
        break
      case 403:
        ElMessage.error(message || '没有权限执行此操作')
        break
      case 404:
        // 404 时透传后端业务消息（如"演练不存在"），避免上层误判
        if (message) {
          error.message = message
        }
        break
      case 500:
        ElMessage.error(message || '服务器错误')
        break
    }
    return Promise.reject(error)
  }
)

export function apiRequest<T>(config: {
  url: string
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'
  data?: unknown
  params?: Record<string, unknown>
  headers?: Record<string, string>
}): Promise<T> {
  return request(config) as Promise<T>
}

// 异步命令变更结果：200 直接返回 data，202 返回 pending 命令
export interface MutationResult<T = unknown> {
  pending: boolean
  data?: T
  command?: FlowCommand
}

function isPendingResult(result: unknown): result is MutationResult {
  return (
    typeof result === 'object' &&
    result !== null &&
    'pending' in result &&
    (result as MutationResult).pending === true
  )
}

// 变更请求：自动附加 Idempotency-Key，处理 202 异步命令。
// 同一 actionId 在网络失败后重试复用同一把键；收到响应后清除。
export function mutationRequest<T = void>(config: {
  url: string
  method: 'POST' | 'PUT' | 'DELETE' | 'PATCH'
  data?: unknown
  actionId: string
}): Promise<MutationResult<T>> {
  const idempotencyKey = getKeyForAction(config.actionId)
  return apiRequest<MutationResult<T> | T>({
    url: config.url,
    method: config.method,
    data: config.data,
    headers: { 'Idempotency-Key': idempotencyKey },
  })
    .then((result) => {
      if (isPendingResult(result)) {
        ElMessage.success('操作已受理')
        clearKey(config.actionId)
        return result as MutationResult<T>
      }
      clearKey(config.actionId)
      return { pending: false, data: result as T }
    })
    .catch((err) => {
      // 收到 HTTP 错误响应（4xx/5xx）说明操作已确定失败，清除键；
      // 网络失败（无 response）保留键以便重试复用
      if (err.response) {
        clearKey(config.actionId)
      }
      throw err
    })
}
