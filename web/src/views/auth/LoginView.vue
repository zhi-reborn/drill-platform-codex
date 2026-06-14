<template>
  <div class="login-page">
    <!-- 全局静态背景层（已优化为低 GPU 占用） -->
    <div class="bg-stack">
      <div class="bg-grid"></div>
      <div class="bg-scan"></div>
      <div class="bg-glow bg-glow-1"></div>
      <div class="bg-glow bg-glow-2"></div>
      <div class="bg-particles">
        <span v-for="n in 10" :key="n" class="particle" :style="particleStyle(n)"></span>
      </div>
    </div>

    <!-- 左侧品牌区 (60%) -->
    <div class="login-brand">
      <!-- 顶部状态条 -->
      <div class="brand-statusbar">
        <span class="status-dot"></span>
        <span class="status-text">SYSTEM ONLINE</span>
        <span class="status-divider"></span>
        <span class="status-text">{{ currentTime }}</span>
        <span class="status-divider"></span>
        <span class="status-text">v2.6 · STABLE</span>
      </div>

      <div class="brand-content">
        <!-- 雷达脉冲容器 -->
        <div class="brand-emblem">
          <div class="radar-pulse"></div>
          <div class="radar-pulse delay-1"></div>
          <div class="radar-pulse delay-2"></div>
          <div class="emblem-core">
            <el-icon :size="56"><Monitor /></el-icon>
          </div>
          <div class="emblem-ring"></div>
          <div class="emblem-arc"></div>
        </div>

        <h1 class="brand-title">
          <span class="title-cn">应急处置流程管理系统</span>
          <span class="title-en">EMERGENCY RESPONSE COMMAND CENTER</span>
        </h1>

        <!-- 装饰分隔线 -->
        <div class="brand-divider">
          <span class="divider-line"></span>
          <span class="divider-mark">◆</span>
          <span class="divider-line"></span>
        </div>

        <p class="brand-desc">
          <span class="desc-item"><i></i>指挥中心大屏</span>
          <span class="desc-item"><i></i>流程引擎驱动</span>
          <span class="desc-item"><i></i>实时通信同步</span>
        </p>
      </div>

      <!-- 角落装饰 -->
      <div class="corner corner-tl"></div>
      <div class="corner corner-tr"></div>
      <div class="corner corner-bl"></div>
      <div class="corner corner-br"></div>
    </div>

    <!-- 右侧登录区 (40%) -->
    <div class="login-area">
      <!-- DEV 环境标记 -->
      <el-tag v-if="authMode === 'dev'" class="dev-badge" type="success" effect="dark" size="small">
        DEV 环境
      </el-tag>

      <div class="login-card">
        <!-- 卡片装饰边框 -->
        <span class="card-edge edge-tl"></span>
        <span class="card-edge edge-tr"></span>
        <span class="card-edge edge-bl"></span>
        <span class="card-edge edge-br"></span>

        <h2 class="login-title">
          <span class="title-prefix">//</span>
          <template v-if="authMode === 'cas'">企业 SSO 登录</template>
          <template v-else-if="authMode === 'dev'">快捷登录</template>
          <template v-else>账号登录</template>
          <span class="title-cursor"></span>
        </h2>

        <!-- CAS 模式 -->
        <div v-if="authMode === 'cas'" class="cas-mode">
          <el-button type="primary" size="large" class="cas-btn" @click="handleCasLogin" :loading="loading">
            <el-icon :size="20"><Connection /></el-icon>
            使用企业统一认证账号登录
          </el-button>
          <p class="hint">登录后将自动跳转至企业身份提供商进行认证</p>
          <div class="link-btn" @click="authMode = 'local'">
            使用账号密码登录
          </div>
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
                autocomplete="off"
                @keyup.enter="handleLocalLogin"
                @input="(val: string) => { form.password = val }"
              />
            </el-form-item>
            <el-form-item>
              <el-checkbox v-model="remember">记住我</el-checkbox>
            </el-form-item>
            <el-form-item>
              <el-button type="primary" size="large" class="login-btn" @click="handleLocalLogin" :loading="loading" :disabled="!canSubmit">
                <span class="btn-text">登 录</span>
                <span class="btn-arrow">→</span>
              </el-button>
            </el-form-item>
          </el-form>
          <div v-if="initialAuthMode === 'cas'" class="link-btn" @click="authMode = 'cas'">
            使用企业统一认证登录
          </div>
        </div>

        <!-- 错误提示 -->
        <div v-if="error" class="login-error">
          <el-alert type="error" :title="error" show-icon :closable="false" />
        </div>
      </div>

      <p class="copyright">
        <span class="copyright-bracket">[</span>
        © 2026 基础设施部 · 技术组件
        <span class="copyright-bracket">]</span>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRouter } from 'vue-router'
import { User, Lock, Connection, Monitor } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import axios from 'axios'
import type { Role } from '@/types'

const router = useRouter()
const authStore = useAuthStore()

const initialAuthMode = (import.meta.env.VITE_AUTH_MODE as 'cas' | 'dev' | 'local') || 'dev'
const authMode = ref<'cas' | 'dev' | 'local'>(initialAuthMode)
const loading = ref(false)
const error = ref('')
const remember = ref(false)
const formRef = ref()

// 实时时钟
const currentTime = ref('')
let clockTimer: number | null = null
function updateClock() {
  const d = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  currentTime.value = `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

// 数据指标 - 已移除 sigilCode 装饰元素

// 粒子样式生成（基于索引的伪随机分布）
function particleStyle(n: number) {
  const seed = (n * 9301 + 49297) % 233280
  const rand = (offset: number) => ((seed + offset * 7919) % 1000) / 1000
  const left = rand(1) * 100
  const top = rand(2) * 100
  const delay = rand(3) * 8
  const duration = 6 + rand(4) * 8
  const size = 1 + rand(5) * 2
  const opacity = 0.2 + rand(6) * 0.5
  return {
    left: `${left}%`,
    top: `${top}%`,
    width: `${size}px`,
    height: `${size}px`,
    opacity: String(opacity),
    animationDelay: `${delay}s`,
    animationDuration: `${duration}s`,
  }
}

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
const form = ref({ username: '', password: '' })
const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }, { min: 3, max: 50, message: '3-50 个字符', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }, { min: 6, message: '至少 6 个字符', trigger: 'blur' }],
}
const canSubmit = computed(() => form.value.username.length >= 3 && form.value.password.length >= 6)

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
  // 展开为普通对象，避免 Vue Proxy 在某些浏览器下序列化异常
  const credentials = { username: form.value.username, password: form.value.password }
  if (!credentials.username.trim() || !credentials.password.trim()) {
    error.value = '请输入用户名和密码'
    return
  }
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  error.value = ''
  try {
    await authStore.loginWithCredentials(credentials)
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
    form.value.password = ''
    loading.value = false
  }
}

async function handleCasLogin() {
  window.location.href = '/api/v1/auth/cas?redirect=' + encodeURIComponent(window.location.origin + '/cas/callback')
}

// Restore session on mount
onMounted(async () => {
  updateClock()
  clockTimer = window.setInterval(updateClock, 1000)

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

onBeforeUnmount(() => {
  if (clockTimer !== null) {
    window.clearInterval(clockTimer)
    clockTimer = null
  }
})
</script>

<style lang="scss" scoped>
@use '@/styles/variables' as *;

// ===== 设计变量 =====
$cyber-bg: #050912;
$cyber-bg-deep: #020610;
$cyber-panel: #0a1220;
$cyber-cyan: #55C3D3;
$cyber-cyan-bright: #7DDDED;
$cyber-blue: #2D7EF7;
$cyber-line: rgba(85, 195, 211, 0.18);
$cyber-line-dim: rgba(85, 195, 211, 0.08);
$cyber-text: #d4e2ee;
$cyber-text-dim: #6b7d92;

.login-page {
  position: relative;
  background: $cyber-bg;
  height: 100vh;
  min-height: 100vh;
  display: flex;
  font-family: $font-family-ui;
  overflow: hidden;
}

// ===== 全局动态背景 =====
.bg-stack {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 0;
}

.bg-grid {
  position: absolute;
  inset: -2px;
  background-image:
    linear-gradient($cyber-line-dim 1px, transparent 1px),
    linear-gradient(90deg, $cyber-line-dim 1px, transparent 1px);
  background-size: 48px 48px;
  // 静态遮罩 - 不使用 mask-image filter（部分低端设备渲染慢）
  // 使用 opacity 控制视觉强度
  opacity: 0.55;
}

.bg-scan {
  position: absolute;
  left: 0; right: 0;
  height: 160px;
  background: linear-gradient(
    180deg,
    transparent 0%,
    rgba(85, 195, 211, 0.06) 45%,
    rgba(85, 195, 211, 0.12) 50%,
    rgba(85, 195, 211, 0.06) 55%,
    transparent 100%
  );
  // 用 transform 而非 top 触发 GPU 合成层（更高效）
  will-change: transform;
  animation: scan-line 10s linear infinite;
  // 移除 filter:blur 以节省 GPU/CPU
}

@keyframes scan-line {
  0% { transform: translate3d(0, -100%, 0); }
  100% { transform: translate3d(0, 100vh, 0); }
}

// 静态径向光斑（不使用 filter:blur(120px)，改用 radial-gradient 实现柔光，零 GPU 滤镜成本）
.bg-glow {
  position: absolute;
  pointer-events: none;
}

.bg-glow-1 {
  width: 720px;
  height: 720px;
  top: -25%;
  left: 5%;
  background: radial-gradient(
    circle at center,
    rgba(85, 195, 211, 0.18) 0%,
    rgba(85, 195, 211, 0.08) 30%,
    transparent 60%
  );
}

.bg-glow-2 {
  width: 640px;
  height: 640px;
  bottom: -20%;
  right: -8%;
  background: radial-gradient(
    circle at center,
    rgba(45, 126, 247, 0.16) 0%,
    rgba(45, 126, 247, 0.06) 30%,
    transparent 60%
  );
}

.bg-particles {
  position: absolute;
  inset: 0;
}

.particle {
  position: absolute;
  background: $cyber-cyan;
  border-radius: 50%;
  // 仅保留小阴影，避免 currentColor box-shadow 在多粒子时累积成本
  box-shadow: 0 0 4px rgba(85, 195, 211, 0.6);
  will-change: transform, opacity;
  animation: particle-drift linear infinite;
}

@keyframes particle-drift {
  0% { transform: translate3d(0, 0, 0); opacity: 0; }
  10% { opacity: 1; }
  90% { opacity: 0.5; }
  100% { transform: translate3d(20px, -110vh, 0); opacity: 0; }
}

// ===== 左侧品牌区 =====
.login-brand {
  position: relative;
  flex: 0 0 60%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  z-index: 1;
  padding: $spacing-3xl;
}

.brand-statusbar {
  position: absolute;
  top: 28px;
  left: 32px;
  display: flex;
  align-items: center;
  gap: 12px;
  font-family: $font-family-mono;
  font-size: 11px;
  letter-spacing: 1.5px;
  color: $cyber-text-dim;

  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #4ade80;
    box-shadow: 0 0 8px #4ade80;
    animation: status-blink 2s ease-in-out infinite;
  }

  .status-divider {
    width: 1px;
    height: 12px;
    background: rgba(212, 226, 238, 0.2);
  }

  .status-text {
    text-transform: uppercase;
  }
}

@keyframes status-blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

.brand-content {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  z-index: 2;
  animation: content-fade-in 1.2s ease-out;
}

@keyframes content-fade-in {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

.brand-emblem {
  position: relative;
  width: 180px;
  height: 180px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: $spacing-2xl;

  .radar-pulse {
    position: absolute;
    inset: 0;
    border-radius: 50%;
    border: 1px solid $cyber-cyan;
    opacity: 0;
    animation: radar 3.6s cubic-bezier(0.22, 0.61, 0.36, 1) infinite;

    &.delay-1 { animation-delay: 1.2s; }
    &.delay-2 { animation-delay: 2.4s; }
  }

  .emblem-core {
    position: relative;
    width: 100px;
    height: 100px;
    border-radius: 50%;
    background: radial-gradient(circle at 30% 30%, rgba(85, 195, 211, 0.3), rgba(10, 18, 32, 0.9));
    border: 1px solid rgba(85, 195, 211, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    color: $cyber-cyan-bright;
    box-shadow:
      0 0 30px rgba(85, 195, 211, 0.4),
      inset 0 0 20px rgba(85, 195, 211, 0.15);
    z-index: 2;
  }

  .emblem-ring {
    position: absolute;
    inset: 10px;
    border: 1px dashed rgba(85, 195, 211, 0.4);
    border-radius: 50%;
    animation: ring-rotate 20s linear infinite;
  }

  .emblem-arc {
    position: absolute;
    inset: -4px;
    border-radius: 50%;
    border: 2px solid transparent;
    border-top-color: $cyber-cyan;
    border-right-color: rgba(85, 195, 211, 0.4);
    animation: ring-rotate 4s linear infinite reverse;
  }
}

@keyframes radar {
  0% { transform: scale(0.5); opacity: 0.8; }
  100% { transform: scale(1.4); opacity: 0; }
}

@keyframes ring-rotate {
  to { transform: rotate(360deg); }
}

.brand-title {
  margin: 0 0 $spacing-base;
  display: flex;
  flex-direction: column;
  gap: $spacing-sm;

  .title-cn {
    font-size: 44px;
    font-weight: 700;
    color: $cyber-text;
    letter-spacing: 6px;
    line-height: 1.1;
    background: linear-gradient(135deg, #ffffff 0%, $cyber-cyan-bright 50%, $cyber-cyan 100%);
    -webkit-background-clip: text;
    background-clip: text;
    -webkit-text-fill-color: transparent;
    text-shadow: 0 0 30px rgba(85, 195, 211, 0.3);
  }

  .title-en {
    font-family: $font-family-mono;
    font-size: 12px;
    font-weight: 400;
    letter-spacing: 4px;
    color: $cyber-text-dim;
    text-transform: uppercase;
  }
}

.brand-divider {
  display: flex;
  align-items: center;
  gap: $spacing-md;
  margin: $spacing-xl 0 $spacing-lg;
  width: 320px;
  max-width: 60%;

  .divider-line {
    flex: 1;
    height: 1px;
    background: linear-gradient(90deg, transparent, rgba(85, 195, 211, 0.5), transparent);
  }

  .divider-mark {
    font-size: 10px;
    color: $cyber-cyan;
    text-shadow: 0 0 8px rgba(85, 195, 211, 0.6);
    animation: mark-pulse 2.4s ease-in-out infinite;
  }
}

@keyframes mark-pulse {
  0%, 100% { transform: rotate(0deg) scale(1); opacity: 0.8; }
  50% { transform: rotate(180deg) scale(1.2); opacity: 1; }
}

.brand-desc {
  display: flex;
  gap: $spacing-xl;
  margin: 0;
  font-size: 13px;
  color: $cyber-text;
  letter-spacing: 1px;

  .desc-item {
    display: inline-flex;
    align-items: center;
    gap: 8px;

    i {
      width: 4px;
      height: 4px;
      background: $cyber-cyan;
      border-radius: 50%;
      box-shadow: 0 0 6px $cyber-cyan;
    }
  }
}

.brand-sigil { display: none; }

// 角落装饰
.corner {
  position: absolute;
  width: 60px;
  height: 60px;
  border: 1px solid $cyber-cyan;
  opacity: 0.6;
}

.corner-tl { top: 24px; left: 24px; border-right: none; border-bottom: none; }
.corner-tr { top: 24px; right: 24px; border-left: none; border-bottom: none; }
.corner-bl { bottom: 24px; left: 24px; border-right: none; border-top: none; }
.corner-br { bottom: 24px; right: 24px; border-left: none; border-top: none; }

// ===== 右侧登录区 =====
.login-area {
  position: relative;
  flex: 0 0 40%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: $spacing-xl;
  z-index: 1;

  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    bottom: 0;
    width: 1px;
    background: linear-gradient(
      to bottom,
      transparent 0%,
      $cyber-line 30%,
      $cyber-cyan 50%,
      $cyber-line 70%,
      transparent 100%
    );
  }

  .dev-badge {
    position: absolute;
    top: $spacing-xl;
    right: $spacing-xl;
    z-index: 5;
  }

  .login-card {
    position: relative;
    width: 100%;
    max-width: 400px;
    padding: $spacing-3xl $spacing-2xl;
    // 不使用 backdrop-filter（GPU 高成本），改用不透明纯色背景模拟玻璃质感
    background: linear-gradient(135deg, #0a1322 0%, #0e1a2e 100%);
    border: 1px solid rgba(85, 195, 211, 0.25);
    box-shadow:
      0 8px 32px rgba(0, 0, 0, 0.5),
      inset 0 1px 0 rgba(255, 255, 255, 0.04);
    animation: card-fade-in 0.8s ease-out 0.1s backwards;

    &::after {
      content: '';
      position: absolute;
      top: 0;
      left: 10%;
      right: 10%;
      height: 1px;
      background: linear-gradient(90deg, transparent, $cyber-cyan, transparent);
      opacity: 0.6;
    }
  }

  .card-edge {
    position: absolute;
    width: 16px;
    height: 16px;
    border: 2px solid $cyber-cyan;
    pointer-events: none;
  }

  .edge-tl { top: -1px; left: -1px; border-right: none; border-bottom: none; }
  .edge-tr { top: -1px; right: -1px; border-left: none; border-bottom: none; }
  .edge-bl { bottom: -1px; left: -1px; border-right: none; border-top: none; }
  .edge-br { bottom: -1px; right: -1px; border-left: none; border-top: none; }

  .login-title {
    font-size: 22px;
    font-weight: 600;
    color: $cyber-text;
    text-align: left;
    margin-bottom: $spacing-2xl;
    letter-spacing: 2px;
    display: flex;
    align-items: center;
    gap: 8px;

    .title-prefix {
      color: $cyber-cyan;
      font-family: $font-family-mono;
      font-weight: 700;
    }

    .title-cursor {
      display: inline-block;
      width: 8px;
      height: 18px;
      background: $cyber-cyan;
      margin-left: 4px;
      animation: cursor-blink 1s steps(2) infinite;
    }
  }
}

@keyframes card-fade-in {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes cursor-blink {
  50% { opacity: 0; }
}

.cas-btn {
  width: 100%;
  padding: 16px;
  font-size: 16px;
  height: auto;
  background: $cyber-cyan;
  border-color: $cyber-cyan;
  letter-spacing: 2px;

  &:hover {
    background: $cyber-cyan-bright;
    border-color: $cyber-cyan-bright;
  }
}

.user-select {
  width: 100%;
  margin-bottom: $spacing-base;
}

.login-btn {
  width: 100%;
  margin-bottom: $spacing-base;
  position: relative;
  overflow: hidden;
  background: linear-gradient(90deg, $cyber-cyan 0%, $cyber-blue 100%) !important;
  border: none !important;
  letter-spacing: 4px;
  font-weight: 600;
  height: 46px;
  transition: all 0.3s ease;

  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: -100%;
    width: 100%;
    height: 100%;
    background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.25), transparent);
    transition: left 0.6s ease;
  }

  &:hover:not(.is-disabled)::before {
    left: 100%;
  }

  &:hover:not(.is-disabled) {
    box-shadow: 0 0 25px rgba(85, 195, 211, 0.6);
    transform: translateY(-1px);
  }

  .btn-arrow {
    margin-left: 8px;
    display: inline-block;
    transition: transform 0.3s ease;
  }

  &:hover:not(.is-disabled) .btn-arrow {
    transform: translateX(4px);
  }
}

.hint {
  font-size: 13px;
  color: $cyber-text-dim;
  text-align: center;
  margin-top: $spacing-base;
}

.link-btn {
  text-align: center;
  color: $cyber-cyan;
  cursor: pointer;
  font-size: 13px;
  margin-top: $spacing-base;
  letter-spacing: 1px;
  transition: color 0.2s;

  &:hover {
    color: $cyber-cyan-bright;
    text-shadow: 0 0 8px rgba(85, 195, 211, 0.5);
  }
}

.login-error {
  margin-top: $spacing-base;
}

.copyright {
  margin-top: $spacing-2xl;
  font-size: 11px;
  color: $cyber-text-dim;
  font-family: $font-family-mono;
  letter-spacing: 2px;

  .copyright-bracket {
    color: $cyber-cyan;
    margin: 0 6px;
  }
}

.user-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

// ===== Element Plus 深色覆盖 =====
:deep(.el-input__wrapper) {
  background-color: rgba(5, 12, 22, 0.6) !important;
  border: 1px solid rgba(85, 195, 211, 0.25) !important;
  box-shadow: none !important;
  transition: all 0.3s ease;

  &:hover {
    border-color: rgba(85, 195, 211, 0.5) !important;
  }
}

:deep(.el-input__wrapper.is-focus) {
  border-color: $cyber-cyan !important;
  box-shadow:
    0 0 0 1px rgba(85, 195, 211, 0.4),
    0 0 20px rgba(85, 195, 211, 0.2) !important;
  background-color: rgba(5, 12, 22, 0.85) !important;
}

:deep(.el-input__inner) {
  color: $cyber-text !important;
  font-weight: 500;
  letter-spacing: 1px;
}

:deep(.el-input__inner::placeholder) {
  color: $cyber-text-dim !important;
  opacity: 0.7;
  letter-spacing: 0.5px;
}

:deep(.el-input__prefix .el-icon),
:deep(.el-input__suffix .el-icon) {
  color: $cyber-cyan !important;
}

:deep(.el-checkbox__label) {
  color: $cyber-text !important;
  font-size: 13px;
  letter-spacing: 0.5px;
}

:deep(.el-checkbox__inner) {
  background-color: rgba(5, 12, 22, 0.6) !important;
  border-color: rgba(85, 195, 211, 0.4) !important;
}

:deep(.el-checkbox__input.is-checked .el-checkbox__inner) {
  background-color: $cyber-cyan !important;
  border-color: $cyber-cyan !important;
}

:deep(.el-checkbox__input.is-checked + .el-checkbox__label) {
  color: $cyber-cyan-bright !important;
}

:deep(.el-button.is-disabled.login-btn) {
  background: linear-gradient(90deg, rgba(85, 195, 211, 0.3) 0%, rgba(45, 126, 247, 0.3) 100%) !important;
  color: rgba(212, 226, 238, 0.5) !important;
  opacity: 0.6;
}

:deep(.el-form-item) {
  margin-bottom: 18px;
}

// 选择框
:deep(.el-select) {
  width: 100%;
}

:deep(.el-select__wrapper) {
  width: 100%;
  background-color: rgba(5, 12, 22, 0.6) !important;
  border: 1px solid rgba(85, 195, 211, 0.25) !important;
  box-shadow: none !important;

  &:hover {
    border-color: rgba(85, 195, 211, 0.5) !important;
  }
}

:deep(.el-select__wrapper.is-focused) {
  border-color: $cyber-cyan !important;
  box-shadow: 0 0 0 1px rgba(85, 195, 211, 0.3) !important;
}

:deep(.el-select__placeholder) {
  color: $cyber-text-dim !important;
}

:deep(.el-select__placeholder.is-transparent) {
  color: $cyber-text-dim !important;
}

:deep(.el-popper) {
  z-index: 9999 !important;
}

:deep(.el-select-dropdown) {
  z-index: 9999 !important;
  background-color: $cyber-panel !important;
  border: 1px solid rgba(85, 195, 211, 0.3) !important;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.6) !important;
}

:deep(.el-select-dropdown__item) {
  color: $cyber-text !important;

  &.hover {
    background-color: rgba(85, 195, 211, 0.1) !important;
  }

  &.selected {
    color: $cyber-cyan-bright !important;
    background-color: rgba(85, 195, 211, 0.15) !important;
  }
}

// 响应式：小屏隐藏品牌区
@media (max-width: 960px) {
  .login-brand { display: none; }
  .login-area { flex: 1; }
}

// 性能降级：当用户系统开启"减少动画"或低性能设备时，关闭循环动画
@media (prefers-reduced-motion: reduce) {
  .bg-scan,
  .particle,
  .radar-pulse,
  .emblem-ring,
  .emblem-arc,
  .status-dot,
  .divider-mark,
  .title-cursor {
    animation: none !important;
  }
  .bg-scan { display: none; }
  .bg-particles { display: none; }
}
</style>
