import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

export interface RouteMeta {
  title: string
  icon?: string
  requiresRole?: string | string[]
  hidden?: boolean
  requiresAuth?: boolean
  layout?: 'default' | 'blank'
  keepAlive?: boolean
}

const routes: RouteRecordRaw[] = [
  // 首页 - 重定向到登录页
  {
    path: '/',
    name: 'Home',
    redirect: '/login',
  },

  // 登录
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/LoginView.vue'),
    meta: { title: '登录', requiresAuth: false },
  },

  // Admin 路由
  {
    path: '/admin',
    component: () => import('@/components/layout/AppLayout.vue'),
    meta: { title: '管理', requiresRole: 'admin' },
    children: [
      {
        path: '',
        name: 'AdminDashboard',
        component: () => import('@/views/admin/DashboardView.vue'),
        meta: { title: '系统概览', icon: 'DataAnalysis' },
      },
      {
        path: 'users',
        name: 'AdminUsers',
        component: () => import('@/views/admin/UsersView.vue'),
        meta: { title: '用户管理', icon: 'User' },
      },
      {
        path: 'drills',
        name: 'AdminDrills',
        component: () => import('@/views/admin/DrillListView.vue'),
        meta: { title: '全部演练', icon: 'Monitor' },
      },
    ],
  },

  // Director 路由
  {
    path: '/director',
    component: () => import('@/components/layout/AppLayout.vue'),
    meta: { title: '指挥', requiresRole: ['admin', 'director'] },
    children: [
      {
        path: '',
        name: 'DirectorDashboard',
        component: () => import('@/views/director/DashboardView.vue'),
        meta: { title: '指挥概览', icon: 'DataAnalysis' },
      },
      {
        path: 'templates',
        name: 'DirectorTemplates',
        component: () => import('@/views/director/TemplatesView.vue'),
        meta: { title: '模板管理', icon: 'Document' },
      },
      {
        path: 'create',
        name: 'DirectorCreate',
        component: () => import('@/views/director/CreateDrillView.vue'),
        meta: { title: '创建演练', icon: 'Plus' },
      },
      {
        path: 'monitor/:id(\\d+)',
        name: 'DirectorMonitor',
        component: () => import('@/views/director/MonitorView.vue'),
        meta: { title: '实时监控', icon: 'VideoCamera' },
      },
      {
        path: 'messages',
        name: 'DirectorMessages',
        component: () => import('@/views/MessagesView.vue'),
        meta: { title: '消息中心', icon: 'Bell', hidden: true },
      },
    ],
  },

  // Executor 路由
  {
    path: '/executor',
    component: () => import('@/components/layout/AppLayout.vue'),
    meta: { title: '执行', requiresRole: 'executor' },
    children: [
      {
        path: '',
        name: 'ExecutorTasks',
        component: () => import('@/views/executor/TasksView.vue'),
        meta: { title: '我的任务', icon: 'Tickets' },
      },
      {
        path: 'tasks/:id(\\d+)',
        name: 'ExecutorTaskDetail',
        component: () => import('@/views/executor/TaskDetailView.vue'),
        meta: { title: '任务详情', icon: 'EditPen', hidden: true },
      },
      {
        path: 'messages',
        name: 'ExecutorMessages',
        component: () => import('@/views/MessagesView.vue'),
        meta: { title: '消息中心', icon: 'Bell', hidden: true },
      },
    ],
  },

  // Viewer 路由
  {
    path: '/viewer',
    component: () => import('@/components/layout/AppLayout.vue'),
    meta: { title: '观察', requiresRole: 'viewer' },
    children: [
      {
        path: '',
        name: 'ViewerDashboard',
        component: () => import('@/views/viewer/DashboardView.vue'),
        meta: { title: '演练概览', icon: 'View' },
      },
      {
        path: 'drills/:id(\\d+)',
        name: 'ViewerDrillDetail',
        component: () => import('@/views/viewer/DrillDetailView.vue'),
        meta: { title: '演练详情', icon: 'Document', hidden: true },
      },
    ],
  },

  // 大屏路由 (无 layout) - 仅支持单演练模式
  {
    path: '/screen/:id(\\d+)',
    name: 'Screen',
    component: () => import('@/views/ScreenView.vue'),
    meta: { title: '监控大屏', requiresAuth: false, layout: 'blank' },
  },
  // 重定向旧版 /screen/all 到首页
  {
    path: '/screen/all',
    redirect: '/viewer',
  },

  // 全局路由
  {
    path: '/403',
    name: 'Forbidden',
    component: () => import('@/views/ForbiddenView.vue'),
    meta: { title: '没有权限', requiresAuth: false },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/NotFoundView.vue'),
    meta: { title: '页面不存在', requiresAuth: false },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 }),
})

export default router
