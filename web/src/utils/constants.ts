// 应用常量定义

// API 配置
export const API_CONFIG = {
  BASE_URL: (import.meta as any).env?.VITE_API_BASE_URL || '/api',
  TIMEOUT: 10000, // 10 秒超时
}

// WebSocket 配置
export const WS_CONFIG = {
  BASE_URL: (import.meta as any).env?.VITE_WS_BASE_URL || 'ws://localhost:8081',
  RECONNECT_INTERVAL: 3000, // 3 秒重连
  MAX_RECONNECT_ATTEMPTS: 5,
  HEARTBEAT_INTERVAL: 30000, // 30 秒心跳
}

// 路由配置
export const ROUTE_CONFIG = {
  LOGIN_PATH: '/login',
  HOME_PATH: '/',
  DASHBOARD_PATH: '/dashboard',
  DRILL_PATH: '/drill',
  TEMPLATE_PATH: '/template',
}

// 用户角色
export const USER_ROLES = {
  ADMIN: 'admin',
  DIRECTOR: 'director',
  EXECUTOR: 'executor',
  VIEWER: 'viewer',
} as const

export type UserRole = (typeof USER_ROLES)[keyof typeof USER_ROLES]

// 演练状态
export const DRILL_STATUS = {
  PENDING: 'pending',
  RUNNING: 'running',
  PAUSED: 'paused',
  COMPLETED: 'completed',
  TERMINATED: 'terminated',
} as const

export type DrillStatus = (typeof DRILL_STATUS)[keyof typeof DRILL_STATUS]

// 步骤状态
export const STEP_STATUS = {
  PENDING: 'pending',
  RUNNING: 'running',
  COMPLETED: 'completed',
  TIMEOUT: 'timeout',
  SKIPPED: 'skipped',
  ISSUE: 'issue',
} as const

export type StepStatus = (typeof STEP_STATUS)[keyof typeof STEP_STATUS]

export const DRILL_STATUS_COLORS: Record<string, string> = {
  [DRILL_STATUS.PENDING]: '#6B7280',
  [DRILL_STATUS.RUNNING]: '#3B82F6',
  [DRILL_STATUS.PAUSED]: '#F59E0B',
  [DRILL_STATUS.COMPLETED]: '#22C55E',
  [DRILL_STATUS.TERMINATED]: '#EF4444',
}

export const STEP_STATUS_COLORS: Record<string, string> = {
  [STEP_STATUS.PENDING]: '#6B7280',
  [STEP_STATUS.RUNNING]: '#3B82F6',
  [STEP_STATUS.COMPLETED]: '#22C55E',
  [STEP_STATUS.TIMEOUT]: '#F59E0B',
  [STEP_STATUS.SKIPPED]: '#6B7280',
  [STEP_STATUS.ISSUE]: '#EF4444',
}

// 分页配置
export const PAGINATION_CONFIG = {
  DEFAULT_PAGE_SIZE: 10,
  PAGE_SIZE_OPTIONS: [10, 20, 50, 100],
}

// 本地存储 key
export const STORAGE_KEYS = {
  TOKEN: 'drill_platform_token',
  USER_INFO: 'drill_platform_user',
  THEME: 'drill_platform_theme',
} as const
