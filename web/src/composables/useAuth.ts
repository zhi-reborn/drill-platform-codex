import { useAuthStore } from '@/stores/auth'

export function useAuth() {
  const store = useAuthStore()
  return {
    user: () => store.user,
    isAuthenticated: () => store.isAuthenticated,
    role: () => store.role,
    hasRole: (r: Parameters<typeof store.hasRole>[0]) => store.hasRole(r),
    hasPermission: (p: string) => store.hasPermission(p),
    login: store.loginWithCredentials,
    loginAsUser: store.loginWithUser,
    logout: store.logout,
  }
}
