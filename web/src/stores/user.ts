import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface UserInfo {
  id: number
  username: string
  realName: string
  role: 'admin' | 'director' | 'executor' | 'viewer'
  department?: string
  phone?: string
}

export const useUserStore = defineStore('user', () => {
  const token = ref<string>('')
  const userInfo = ref<UserInfo | null>(null)

  function setLoginInfo(newToken: string, info: UserInfo) {
    token.value = newToken
    userInfo.value = info
    localStorage.setItem('token', newToken)
    localStorage.setItem('userInfo', JSON.stringify(info))
  }

  function restoreLogin() {
    const savedToken = localStorage.getItem('token')
    const savedUserInfo = localStorage.getItem('userInfo')
    if (savedToken && savedUserInfo) {
      token.value = savedToken
      userInfo.value = JSON.parse(savedUserInfo)
    }
  }

  function logout() {
    token.value = ''
    userInfo.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('userInfo')
  }

  function hasRole(roles: string[]) {
    return userInfo.value && roles.includes(userInfo.value.role)
  }

  const isLoggedIn = computed(() => !!token.value && !!userInfo.value)

  return {
    token,
    userInfo,
    isLoggedIn,
    setLoginInfo,
    restoreLogin,
    logout,
    hasRole
  }
})
