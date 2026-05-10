export type Role = 'admin' | 'director' | 'executor' | 'viewer'

export const ROLE_LABELS: Record<Role, string> = {
  admin: '管理员',
  director: '指挥员',
  executor: '执行者',
  viewer: '观察者',
}

export const ROLE_ORDER: Role[] = ['admin', 'director', 'executor', 'viewer']

export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

export interface PageResult<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

export interface SortConfig {
  field: string
  order: 'asc' | 'desc'
}
