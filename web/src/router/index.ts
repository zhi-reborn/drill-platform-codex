import { createRouter, createWebHistory, type RouteLocationNormalized } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ROUTE_CONFIG } from '@/utils/constants'
import type { AppRouteRecord } from './types'

// 常量路由 (公开)
const constantRoutes: AppRouteRecord[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/Login.vue'),
    meta: {
      title: '登录',
      hidden: true,
    },
  },
  {
    path: '/404',
    name: '404',
    component: () => import('@/views/error/404.vue'),
    meta: {
      title: '404',
      hidden: true,
    },
  },
  {
    path: '/500',
    name: '500',
    component: () => import('@/views/error/500.vue'),
    meta: {
      title: '500',
      hidden: true,
    },
  },
]

// 动态路由 (需要权限)
const dynamicRoutes: AppRouteRecord[] = [
  {
    path: '/',
    name: 'Layout',
    component: () => import('@/layouts/default/Index.vue'),
    redirect: '/display',
    meta: {
      requiresAuth: true,
    },
    children: [
      {
        path: 'display',
        name: 'Display',
        component: () => import('@/views/display/Index.vue'),
        meta: {
          title: '指挥中心大屏',
          icon: 'dashboard',
          requiresAuth: true,
          keepAlive: true,
        },
      },
      {
        path: 'console',
        name: 'Console',
        component: () => import('@/views/console/DrillList.vue'),
        meta: {
          title: '演练指挥台',
          icon: 'drill',
          requiresAuth: true,
          keepAlive: true,
        },
      },
      {
        path: 'console/create',
        name: 'DrillCreate',
        component: () => import('@/views/console/DrillCreate.vue'),
        meta: {
          title: '创建演练',
          hidden: true,
          requiresAuth: true,
          breadcrumb: true,
        },
      },
      {
        path: 'console/:id',
        name: 'DrillControl',
        component: () => import('@/views/console/DrillControl.vue'),
        meta: {
          title: '演练控制',
          hidden: true,
          requiresAuth: true,
          breadcrumb: true,
        },
      },
      {
        path: 'workspace',
        name: 'Workspace',
        component: () => import('@/views/workspace/MyTasks.vue'),
        meta: {
          title: '参演工作台',
          icon: 'workspace',
          requiresAuth: true,
          keepAlive: true,
        },
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/404',
    meta: {
      hidden: true,
    },
  },
]

// 创建路由器
const router = createRouter({
  history: createWebHistory(),
  routes: [...constantRoutes, ...dynamicRoutes],
  scrollBehavior(_to, _from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    } else {
      return { top: 0 }
    }
  },
})

// 路由守卫
router.beforeEach(async (to: RouteLocationNormalized, _from: RouteLocationNormalized) => {
  const userStore = useUserStore()
  const { isLoggedIn } = userStore
  
  // 设置页面标题
  document.title = to.meta.title ? `${to.meta.title} - 生产演练平台` : '生产演练平台'
  
  // 检查是否需要登录
  const requiresAuth = to.meta?.requiresAuth
  
  if (requiresAuth && !isLoggedIn) {
    // 未登录，重定向到登录页
    return {
      path: ROUTE_CONFIG.LOGIN_PATH,
      query: { redirect: to.fullPath },
    }
  }
  
  // 如果已登录且访问登录页，重定向到首页
  if (isLoggedIn && to.path === ROUTE_CONFIG.LOGIN_PATH) {
    return { path: '/display' }
  }
})

export default router
export { constantRoutes, dynamicRoutes }
