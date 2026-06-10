// web/src/config/menu.ts
import type { Role } from '@/types'

export interface MenuItem {
  path: string
  title: string
  icon: string
  roles?: Role[]
  children?: MenuItem[]
}

export const menuConfig: MenuItem[] = [
  {
    path: '/dashboard',
    title: '工作台',
    icon: 'DataAnalysis',
    roles: ['admin', 'director', 'executor', 'viewer'],
    children: [
      { path: '/admin', title: '系统概览', icon: 'DataAnalysis', roles: ['admin'] },
      { path: '/admin/users', title: '用户管理', icon: 'User', roles: ['admin'] },
      { path: '/director', title: '指挥概览', icon: 'DataAnalysis', roles: ['admin', 'director'] },
      { path: '/executor', title: '任务中心', icon: 'Tickets', roles: ['executor'] },
      { path: '/viewer', title: '演练总览', icon: 'View', roles: ['viewer'] },
    ],
  },
  {
    path: '/drills',
    title: '演练管理',
    icon: 'Monitor',
    roles: ['admin', 'director'],
    children: [
      { path: '/director/templates', title: '模板库', icon: 'Document', roles: ['admin', 'director'] },
      { path: '/director/create', title: '创建演练', icon: 'Plus', roles: ['admin', 'director'] },
      { path: '/director/drills', title: '演练列表', icon: 'Monitor', roles: ['admin', 'director'] },
    ],
  },
  {
    path: '/messages',
    title: '消息中心',
    icon: 'Bell',
    roles: ['admin', 'director', 'executor', 'viewer'],
  },
]

export function getVisibleMenus(role: Role): MenuItem[] {
  return menuConfig.filter(item => {
    if (!item.roles) return true
    return item.roles.includes(role)
  }).map(item => {
    if (!item.children) return item
    return {
      ...item,
      children: item.children.filter(child => {
        if (!child.roles) return true
        return child.roles.includes(role)
      }),
    }
  }).filter(item => {
    if (!item.children || item.children.length > 0) return true
    return false
  })
}
