import axios from 'axios'
import type { InternalAxiosRequestConfig, AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import type { ApiResponse } from '@/types'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 15000,
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
    return res.data as AxiosResponse
  },
  (error) => {
    const status = error.response?.status
    const data = error.response?.data
    const message = data?.message
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
}): Promise<T> {
  return request(config) as Promise<T>
}
