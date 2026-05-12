import type { Router, RouteLocationNormalized, NavigationGuardNext } from 'vue-router'
import type { User, Role } from '@/types'

export function setupGuards(router: Router) {
  router.beforeEach(async (to, from, next) => {
    const requiresAuth = to.meta.requiresAuth !== false
    if (!requiresAuth) return next()

    const auth = localStorage.getItem('drill_auth')
    if (!auth) return next({ name: 'Login', query: { redirect: to.fullPath } })

    const { useAuthStore } = await import('@/stores/auth')
    const authStore = useAuthStore()

    if (!authStore.user) {
      const userData = localStorage.getItem('drill_user')
      if (userData) {
        authStore.user = JSON.parse(userData) as User
      }
    }

    if (!authStore.user) {
      return next({ name: 'Login' })
    }

    const requiredRole = to.meta.requiresRole as Role | Role[] | undefined
    if (requiredRole && !authStore.hasRole(requiredRole)) {
      return next({ name: 'NotFound' })
    }

    next()
  })

  router.afterEach((to) => {
    document.title = `${to.meta.title || 'Drill Platform'} - ${import.meta.env.VITE_APP_TITLE || '生产演练平台'}`
  })
}
