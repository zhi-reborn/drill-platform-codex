<template>
  <div class="login-page">
    <!-- 左侧品牌区 (60%) -->
    <div class="login-brand">
      <div class="brand-content">
        <el-icon :size="64" color="#55C3D3"><Monitor /></el-icon>
        <h1>演练流程管理系统</h1>
        <p class="brand-desc">指挥中心大屏 · 流程引擎驱动 · 实时通信同步</p>
      </div>
      <div class="brand-bg-pattern"></div>
    </div>

    <!-- 右侧登录区 (40%) -->
    <div class="login-area">
      <!-- DEV 环境标记 -->
      <el-tag v-if="authMode === 'dev'" class="dev-badge" type="success" effect="dark" size="small">
        DEV 环境
      </el-tag>

      <div class="login-card">
        <h2 class="login-title">
          <template v-if="authMode === 'cas'">企业 SSO 登录</template>
          <template v-else-if="authMode === 'dev'">快捷登录</template>
          <template v-else>账号登录</template>
        </h2>

        <!-- CAS 模式 -->
        <div v-if="authMode === 'cas'" class="cas-mode">
          <el-button type="primary" size="large" class="cas-btn" @click="handleCasLogin" :loading="loading">
            <el-icon :size="20"><Connection /></el-icon>
            使用企业统一认证账号登录
          </el-button>
          <p class="hint">登录后将自动跳转至企业身份提供商进行认证</p>
        </div>

        <!-- DEV 模式 -->
        <div v-else-if="authMode === 'dev'" class="dev-mode">
          <el-select v-model="selectedUser" filterable placeholder="请选择登录用户" size="large" class="user-select">
            <el-option
              v-for="u in devUsers"
              :key="u.id"
              :label="u.real_name || u.username"
              :value="u.id"
            >
              <div class="user-option">
                <span>{{ u.real_name || u.username }} / {{ u.username }}</span>
                <el-tag :size="'small'" :type="roleTagType(u.role) as 'primary' | 'success' | 'warning' | 'info' | 'danger'">{{ roleLabel(u.role) }}</el-tag>
              </div>
            </el-option>
          </el-select>
          <el-button type="primary" size="large" class="login-btn" @click="handleDevLogin" :loading="loading">
            快捷登录
          </el-button>
          <div class="link-btn" @click="authMode = 'local'; selectedUser = null">
            切换手动表单登录
          </div>
        </div>

        <!-- Local 模式 -->
        <div v-else class="local-mode">
          <el-form ref="formRef" :model="form" :rules="rules" @submit.prevent="handleLocalLogin">
            <el-form-item prop="username">
              <el-input
                v-model="form.username"
                placeholder="用户名"
                size="large"
                :prefix-icon="User"
                autocomplete="username"
              />
            </el-form-item>
            <el-form-item prop="password">
              <el-input
                v-model="form.password"
                type="password"
                placeholder="密码"
                size="large"
                :prefix-icon="Lock"
                show-password
                autocomplete="current-password"
                @keyup.enter="handleLocalLogin"
              />
            </el-form-item>
            <el-form-item>
              <el-checkbox v-model="remember">记住我</el-checkbox>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" size="large" class="login-btn" @click="handleLocalLogin" :loading="loading" :disabled="!canSubmit">
                登录
              </el-button>
            </el-form-item>
          </el-form>
          <div class="link-btn" @click="authMode = 'dev'; selectedUser = null">
            切换快捷登录
          </div>
        </div>

        <!-- 错误提示 -->
        <div v-if="error" class="login-error">
          <el-alert type="error" :title="error" show-icon :closable="false" />
        </div>
      </div>

      <p class="copyright">&copy; 2026 基础设施部 技术组件</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { User, Lock, Connection, CircleCheck, Monitor } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import axios from 'axios'
import type { Role } from '@/types'

const router = useRouter()
const authStore = useAuthStore()

const authMode = ref<'cas' | 'dev' | 'local'>((import.meta.env.VITE_AUTH_MODE as 'cas' | 'dev' | 'local') || 'dev')
const loading = ref(false)
const error = ref('')
const remember = ref(false)
const formRef = ref()

interface DevUser {
  id: number
  username: string
  real_name: string
  role: string
  department: string
}

const selectedUser = ref<number | null>(null)
const devUsers = ref<DevUser[]>([])

async function fetchDevUsers() {
  const baseUrl = import.meta.env.VITE_BACKEND_URL || 'http://localhost:8080'
  const urls = [
    `${baseUrl}/api/v1/auth/dev-users`,
    `${baseUrl.replace('localhost', 'host.docker.internal')}/api/v1/auth/dev-users`,
    '/api/v1/auth/dev-users',
  ]
  let lastError = null
  for (const url of urls) {
    try {
      const response = await axios.get(url)
      // 后端返回格式：{ code: 0, message: 'success', data: { items: [...], total, page, page_size } }
      const backendData = response.data.data
      devUsers.value = backendData.items || backendData.list || []
      return
    } catch (e: unknown) {
      lastError = e
      console.warn(`Failed to fetch from ${url}, trying next...`)
    }
  }
  console.error('All fallback URLs exhausted:', lastError)
}

// Local mode
const form = reactive({ username: '', password: '' })
const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }, { min: 3, max: 50, message: '3-50 个字符', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }, { min: 6, message: '至少 6 个字符', trigger: 'blur' }],
}
const canSubmit = computed(() => form.username.length >= 3 && form.password.length >= 6)

function roleLabel(role: string): string {
  const map: Record<string, string> = { admin: '管理员', director: '指挥员', executor: '执行者', viewer: '观察者' }
  return map[role] || role
}

function roleTagType(role: string): string {
  const map: Record<string, string> = { admin: 'danger', director: 'warning', executor: 'success', viewer: 'info' }
  return map[role]
}

const roleDashboards: Record<Role, string> = {
  admin: '/admin',
  director: '/director',
  executor: '/executor',
  viewer: '/viewer',
}

async function handleDevLogin() {
  if (!selectedUser.value) {
    ElMessage.warning('请选择登录用户')
    return
  }
  loading.value = true
  error.value = ''
  try {
    const user = devUsers.value.find(u => u.id === selectedUser.value)!
    const { username } = user
    const baseUrl = import.meta.env.VITE_BACKEND_URL || 'http://localhost:8080'
    const response = await axios.post(`${baseUrl}/api/v1/auth/login`, {
      username,
      password: 'admin123'
    })
    const data = response.data.data
    authStore.token = data.token
    authStore.refreshToken = data.token
    authStore.user = {
      user_id: data.user_id,
      username: data.username,
      name: data.real_name,
      role: data.role as Role,
      department: data.department,
      status: 'active' as const,
    }
    localStorage.setItem('drill_auth', JSON.stringify({
      access_token: data.token,
      refresh_token: data.token,
    }))
    localStorage.setItem('drill_user', JSON.stringify(authStore.user))
    ElMessage.success(`欢迎回来，${user.real_name || user.username}`)
    router.push(roleDashboards[user.role as Role] || '/viewer')
  } catch (e: unknown) {
    if (axios.isAxiosError(e) && e.response?.data?.message) {
      error.value = e.response.data.message === '密码错误'
        ? '该用户密码不是 admin123，请切换到手动表单登录'
        : e.response.data.message
    } else {
      error.value = '登录失败'
    }
  } finally {
    loading.value = false
  }
}

async function handleLocalLogin() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  error.value = ''
  try {
    await authStore.loginWithCredentials(form)
    ElMessage.success('登录成功')
    const user = authStore.user
    router.push(user ? roleDashboards[user.role] : '/viewer')
  } catch (e: unknown) {
    if (axios.isAxiosError(e) && e.response?.data?.message) {
      error.value = e.response.data.message
    } else if (e instanceof Error) {
      error.value = e.message
    } else {
      error.value = '登录失败'
    }
  } finally {
    form.password = ''
    loading.value = false
  }
}

async function handleCasLogin() {
  window.location.href = '/api/v1/auth/cas?redirect=' + encodeURIComponent(window.location.origin + '/cas/callback')
}

// Restore session on mount
onMounted(async () => {
  authStore.restoreSession()
  if (authStore.isAuthenticated && authStore.user) {
    router.push(roleDashboards[authStore.user.role as Role] || '/viewer')
    return
  }
  if (authMode.value === 'dev') {
    try {
      await fetchDevUsers()
    } catch (e) {
      console.error('Failed to load users:', e)
    }
  }
})
</script>

<style lang="scss" scoped>
@use '@/styles/variables' as *;

.login-page {
  background: #020617;
  height: 100vh;
  min-height: 100vh;
  display: flex;
  // 移除 overflow: hidden 以允许下拉框显示
}

.login-brand {
  flex: 0 0 60%;
  background: linear-gradient(135deg, #0D1117 0%, #1A1F2E 50%, #0D1117 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;

  .brand-content {
    text-align: center;
    z-index: 2;
    
    h1 {
      font-size: 48px;
      font-weight: 700;
      color: #d4d8dd;
      margin-top: $spacing-xl;
      margin-bottom: $spacing-sm;
      letter-spacing: 2px;
    }
  }

  .brand-tagline {
    font-size: 20px;
    color: $color-accent;
    margin-bottom: $spacing-base;
    font-weight: 400;
  }

  .brand-desc {
    font-size: 14px;
    color: $text-secondary;
    margin-bottom: $spacing-2xl;
  }

  .brand-features {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: $spacing-sm;
    
    .feature {
      display: flex;
      align-items: center;
      gap: $spacing-sm;
      color: $text-secondary;
      font-size: 14px;
      
      .el-icon {
        color: $color-accent;
      }
    }
  }

  .brand-bg-pattern {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-image: 
      radial-gradient(circle at 20% 30%, rgba(85, 195, 211, 0.08) 0%, transparent 50%),
      radial-gradient(circle at 80% 70%, rgba(85, 195, 211, 0.05) 0%, transparent 50%);
    pointer-events: none;
  }
}

.login-area {
  flex: 0 0 40%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background: #161B22;
  position: relative;
  padding: $spacing-xl;

  .dev-badge {
    position: absolute;
    top: $spacing-xl;
    right: $spacing-xl;
  }

  .login-card {
    width: 100%;
    max-width: 380px;
    padding: $spacing-3xl $spacing-2xl;
    border-radius: $radius-lg;
    background: #0F172A;
    border: 1px solid #1E293B;

    .login-title {
      font-size: 24px;
      font-weight: 600;
      color: #d4d8dd;
      text-align: center;
      margin-bottom: $spacing-2xl;
    }

    .hint {
      font-size: 13px;
      color: $text-tertiary;
      text-align: center;
      margin-top: $spacing-base;
    }
  }

  .cas-btn {
    width: 100%;
    padding: 16px;
    font-size: 16px;
    height: auto;
    background: $color-accent;
    border-color: $color-accent;

    &:hover {
      background: $color-accent-hover;
    }
  }

  .user-select, .login-btn {
    width: 100%;
    margin-bottom: $spacing-base;
  }

  .link-btn {
    text-align: center;
    color: $color-accent;
    cursor: pointer;
    font-size: 13px;
    margin-top: $spacing-base;
    
    &:hover {
      text-decoration: underline;
    }
  }

  .login-error {
    margin-top: $spacing-base;
  }

  .copyright {
    margin-top: $spacing-2xl;
    font-size: 12px;
    color: $text-tertiary;
  }
}

.user-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

:deep(.el-input__wrapper) {
  background-color: rgba(15, 23, 42, 0.9) !important;
  border: 1px solid rgba(203, 213, 225, 0.72) !important;
  box-shadow: 0 0 0 1px rgba(15, 23, 42, 0.2) !important;
}

:deep(.el-input__wrapper.is-focus) {
  border-color: #55c3d3 !important;
  box-shadow: 0 0 0 1px rgba(85, 195, 211, 0.6), 0 0 18px rgba(85, 195, 211, 0.18) !important;
}

:deep(.el-input__inner) {
  color: #f8fafc !important;
  font-weight: 600;
}

:deep(.el-input__inner::placeholder) {
  color: #cbd5e1 !important;
  opacity: 0.88;
}

:deep(.el-input__prefix .el-icon),
:deep(.el-input__suffix .el-icon) {
  color: #cbd5e1 !important;
}

:deep(.el-checkbox__label) {
  color: #dbeafe !important;
  font-weight: 600;
}

:deep(.el-checkbox__input.is-checked + .el-checkbox__label) {
  color: #f8fafc !important;
}

:deep(.el-button.is-disabled.login-btn) {
  color: rgba(15, 23, 42, 0.78) !important;
  background-color: #93c5fd !important;
  border-color: #93c5fd !important;
  opacity: 0.72;
}

:deep(.el-form-item) {
  margin-bottom: 16px;
}

// 修复下拉框显示问题
:deep(.el-select) {
  width: 100%;
}

:deep(.el-select__wrapper) {
  width: 100%;
}

:deep(.el-popper) {
  z-index: 9999 !important;
}

:deep(.el-select-dropdown) {
  z-index: 9999 !important;
  background-color: #1A1F2E !important;
  border: 1px solid #30363D !important;
}

:deep(.el-select-dropdown__item) {
  color: #d4d8dd !important;
  
  &.hover {
    background-color: #0D1117 !important;
  }
  
  &.selected {
    color: #55C3D3 !important;
  }
}
</style>
