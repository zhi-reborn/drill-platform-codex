import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User, Role, LoginCredentials, TokenResponse } from '@/types'
import { authApi } from '@/api/modules/auth'
import usersData from '@/mock/data/users.json'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string>('')
  const refreshToken = ref<string>('')
  const user = ref<User | null>(null)
  const isAuthenticated = computed(() => !!token.value && !!user.value)

  const role = computed(() => user.value?.role ?? '' as Role)
  const userName = computed(() => user.value?.name ?? '未登录')
  const userInitial = computed(() => user.value?.name?.charAt(0) ?? '?')
  const roleName = computed(() => {
    const labels: Record<Role, string> = {
      admin: '管理员',
      director: '指挥员',
      executor: '执行者',
      viewer: '观察者',
    }
    return labels[role.value] ?? '未知'
  })
  const roleType = computed(() => {
    const types: Record<Role, string> = {
      admin: 'danger',
      director: 'warning',
      executor: 'success',
      viewer: 'info',
    }
    return types[role.value] ?? 'info'
  })

  function hasRole(requiredRole: Role | Role[]): boolean {
    if (!user.value) return false
    if (Array.isArray(requiredRole)) return requiredRole.includes(user.value.role)
    const hierarchy: Role[] = ['admin', 'director', 'executor', 'viewer']
    const userIndex = hierarchy.indexOf(user.value.role)
    const requiredIndex = hierarchy.indexOf(requiredRole)
    return userIndex >= requiredIndex
  }

  function hasPermission(permission: string): boolean {
    const perms: Record<Role, string[]> = {
      admin: ['*'],
      director: ['drill:create', 'drill:read', 'drill:update', 'drill:pause', 'drill:resume', 'drill:terminate'],
      executor: ['drill:read', 'task:read', 'task:execute'],
      viewer: ['drill:read'],
    }
    const userPerms = perms[user.value?.role ?? 'viewer'] ?? []
    return userPerms.includes('*') || userPerms.includes(permission)
  }

  async function loginWithCredentials(credentials: LoginCredentials): Promise<void> {
    const response = await authApi.login(credentials)
    token.value = response.access_token
    refreshToken.value = response.refresh_token
    await fetchCurrentUser()
    localStorage.setItem('drill_auth', JSON.stringify({
      access_token: response.access_token,
      refresh_token: response.refresh_token,
      userId: user.value?.id,
    }))
  }

  async function loginWithUser(userObj: User): Promise<void> {
    const mockToken = `mock_token_${userObj.id}_${Date.now()}`
    token.value = mockToken
    refreshToken.value = mockToken
    localStorage.setItem('drill_auth', JSON.stringify({
      access_token: mockToken,
      refresh_token: mockToken,
      userId: userObj.id,
    }))
    user.value = userObj
    localStorage.setItem('drill_user', JSON.stringify(userObj))
  }

  async function fetchCurrentUser(): Promise<void> {
    try {
      const currentUser = await authApi.getCurrentUser()
      user.value = currentUser
    } catch {
      user.value = null
    }
  }

  async function casLogin(ticket: string): Promise<void> {
    // CAS login - to be implemented when backend supports it
    // For now, this is a placeholder for future CAS integration
    console.log('CAS login requested with ticket:', ticket)
  }

  function logout(): void {
    token.value = ''
    refreshToken.value = ''
    user.value = null
    localStorage.removeItem('drill_auth')
    localStorage.removeItem('drill_user')
    window.location.href = '/login'
  }

  function restoreSession(): void {
    const auth = localStorage.getItem('drill_auth')
    const userData = localStorage.getItem('drill_user')
    if (auth) {
      const { access_token, refresh_token, userId } = JSON.parse(auth)
      token.value = access_token
      refreshToken.value = refresh_token
    }
    if (userData) {
      user.value = JSON.parse(userData)
    }
  }

  return {
    token, refreshToken, user, isAuthenticated,
    role, userName, userInitial, roleName, roleType,
    hasRole, hasPermission,
    loginWithCredentials, loginWithUser, casLogin,
    fetchCurrentUser, logout, restoreSession,
  }
})
