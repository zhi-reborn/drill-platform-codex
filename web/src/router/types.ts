import type { RouteRecordRaw } from 'vue-router'
import { USER_ROLES } from '@/utils/constants'

// 路由元信息
export interface RouteMeta {
  title?: string
  icon?: string
  requiresAuth?: boolean
  roles?: (typeof USER_ROLES)[keyof typeof USER_ROLES][]
  keepAlive?: boolean
  hidden?: boolean
  breadcrumb?: boolean
}

// 路由记录
export type AppRouteRecord = RouteRecordRaw & {
  meta?: RouteMeta
  children?: AppRouteRecord[]
}
