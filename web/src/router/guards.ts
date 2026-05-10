import type { Router, RouteLocationNormalized, NavigationGuardNext } from 'vue-router'

export function setupGuards(router: Router) {
  router.beforeEach(async (to, from, next) => {
    const requiresAuth = to.meta.requiresAuth !== false
    if (!requiresAuth) return next()

    // Check auth
    const auth = localStorage.getItem('drill_auth')
    if (!auth) return next({ name: 'Login', query: { redirect: to.fullPath } })

    // Check role - lazy import to avoid circular
    const { useAuthStore } = await import('@/stores/auth')
    const authStore = useAuthStore()

    // Ensure user data is loaded
    if (!authStore.user) authStore.restoreSession()

    const requiredRole = to.meta.requiresRole as import('@/types').Role | import('@/types').Role[] | undefined
    if (requiredRole && !authStore.hasRole(requiredRole)) {
      return next({ name: 'NotFound' })
    }

    next()
  })

  router.afterEach((to) => {
    document.title = `${to.meta.title || 'Drill Platform'} - ${import.meta.env.VITE_APP_TITLE || '生产演练平台'}`
  })
}
