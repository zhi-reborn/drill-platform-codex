# 前端项目初始化 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 完成 Vue3 前端项目的基础骨架，包括目录结构、配置文件、核心依赖、状态管理、路由和工具函数

**Architecture:** 采用标准的 Vue3 + Vite + TypeScript 项目结构，使用 Pinia 进行状态管理，Vue Router 进行路由管理，Axios 封装 HTTP 请求，Element Plus 作为 UI 组件库

**Tech Stack:** Vue 3.4+, Vite 5.x, TypeScript 5.x, Element Plus 2.5+, Pinia 2.x, Vue Router 4.x, Axios 1.x, ECharts 5.4+

**Design System:** 参考 ui-ux-pro-max 输出的 Data-Dense Dashboard 风格
- 主色调：`#0F172A` (slate-900)
- 背景色：`#020617` (slate-950)
- 强调色：`#22C55E` (green-500)
- 字体：Fira Code (标题) + Fira Sans (正文)
- 支持完整的亮色/暗色模式

---

## 任务分解

### Task 1: 创建目录结构和 package.json

**Files:**
- Create: `web/package.json`
- Create: `web/src/api/`
- Create: `web/src/assets/`
- Create: `web/src/components/`
- Create: `web/src/stores/`
- Create: `web/src/router/`
- Create: `web/src/utils/`
- Create: `web/src/views/`
- Create: `web/src/styles/`

- [ ] **Step 1: 创建 web 目录结构**

```bash
mkdir -p web/src/{api,assets,components,stores,router,utils,views,styles}
```

- [ ] **Step 2: 创建 package.json**

```json
{
  "name": "drill-platform-web",
  "version": "1.0.0",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vue-tsc && vite build",
    "preview": "vite preview",
    "lint": "eslint . --ext .vue,.js,.jsx,.cjs,.mjs,.ts,.tsx,.cts,.mts --fix --ignore-path .gitignore"
  },
  "dependencies": {
    "vue": "^3.4.0",
    "vue-router": "^4.2.5",
    "pinia": "^2.1.7",
    "element-plus": "^2.5.0",
    "axios": "^1.6.5",
    "echarts": "^5.4.3"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^5.0.0",
    "vite": "^5.0.10",
    "vue-tsc": "^1.8.25",
    "typescript": "^5.3.3",
    "@types/node": "^20.10.6",
    "@vue/tsconfig": "^0.5.1",
    "sass": "^1.69.7",
    "unplugin-auto-import": "^0.17.3",
    "unplugin-vue-components": "^0.26.0"
  }
}
```

- [ ] **Step 3: 验证目录结构**

```bash
ls -la web/ && ls -la web/src/
```

Expected: 显示所有创建的目录

### Task 2: 创建 TypeScript 配置文件

**Files:**
- Create: `web/tsconfig.json`
- Create: `web/tsconfig.node.json`

- [ ] **Step 1: 创建 tsconfig.json**

```json
{
  "extends": "@vue/tsconfig/tsconfig.dom.json",
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "module": "ESNext",
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "preserve",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "noImplicitAny": false,
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["src/**/*.ts", "src/**/*.tsx", "src/**/*.vue"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
```

- [ ] **Step 2: 创建 tsconfig.node.json**

```json
{
  "compilerOptions": {
    "composite": true,
    "skipLibCheck": true,
    "module": "ESNext",
    "moduleResolution": "bundler",
    "allowSyntheticDefaultImports": true
  },
  "include": ["vite.config.ts", "package.json"]
}
```

### Task 3: 创建 Vite 配置文件

**Files:**
- Create: `web/vite.config.ts`

- [ ] **Step 1: 创建 vite.config.ts**

```typescript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      imports: ['vue', 'vue-router', 'pinia'],
      resolvers: [ElementPlusResolver()],
      dts: 'src/auto-imports.d.ts',
      eslintrc: {
        enabled: true,
      },
    }),
    Components({
      resolvers: [ElementPlusResolver()],
      dts: 'src/components.d.ts',
    }),
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: `@use "@/styles/variables.scss" as *;`,
      },
    },
  },
  server: {
    port: 3000,
    host: true,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://localhost:8081',
        ws: true,
      },
    },
  },
})
```

### Task 4: 创建全局样式和变量

**Files:**
- Create: `web/src/styles/variables.scss`
- Create: `web/src/styles/global.scss`
- Create: `web/src/styles/index.ts`

- [ ] **Step 1: 创建样式变量文件**

```scss
// 设计系统变量 - 基于 Data-Dense Dashboard 风格

// 颜色系统
$colors: (
  'primary': #0F172A,
  'on-primary': #FFFFFF,
  'secondary': #1E293B,
  'accent': #22C55E,
  'background': #020617,
  'foreground': #F8FAFC,
  'muted': #1A1E2F,
  'border': #334155,
  'destructive': #EF4444,
  'ring': #0F172A,
  'success': #22C55E,
  'warning': #F59E0B,
  'error': #EF4444,
  'info': #3B82F6,
);

// 字体系统
$font-family-heading: 'Fira Code', monospace;
$font-family-body: 'Fira Sans', sans-serif;

// 字号系统
$font-sizes: (
  'xs': 0.75rem,    // 12px
  'sm': 0.875rem,   // 14px
  'base': 1rem,     // 16px
  'lg': 1.125rem,   // 18px
  'xl': 1.25rem,    // 20px
  '2xl': 1.5rem,    // 24px
  '3xl': 1.875rem,  // 30px
  '4xl': 2.25rem,   // 36px
);

// 间距系统 (8px 基准)
$spacing: (
  '0': 0,
  '1': 0.25rem,    // 4px
  '2': 0.5rem,     // 8px
  '3': 0.75rem,    // 12px
  '4': 1rem,       // 16px
  '5': 1.25rem,    // 20px
  '6': 1.5rem,     // 24px
  '8': 2rem,       // 32px
  '10': 2.5rem,    // 40px
  '12': 3rem,      // 48px
  '16': 4rem,      // 64px
);

// 圆角系统
$radius: (
  'sm': 0.25rem,   // 4px
  'md': 0.5rem,    // 8px
  'lg': 0.75rem,   // 12px
  'xl': 1rem,      // 16px
  'full': 9999px,
);

// 阴影系统
$shadows: (
  'sm': 0 1px 2px 0 rgb(0 0 0 / 0.05),
  'md': 0 4px 6px -1px rgb(0 0 0 / 0.1),
  'lg': 0 10px 15px -3px rgb(0 0 0 / 0.1),
  'xl': 0 20px 25px -5px rgb(0 0 0 / 0.1),
);

// 断点系统
$breakpoints: (
  'sm': 640px,
  'md': 768px,
  'lg': 1024px,
  'xl': 1280px,
  '2xl': 1536px,
);

// z-index 层级
$z-index: (
  'base': 0,
  'dropdown': 10,
  'sticky': 20,
  'fixed': 40,
  'modal-backdrop': 100,
  'modal': 110,
  'popover': 120,
  'toast': 1000,
);

// 动画时间
$transitions: (
  'fast': 150ms,
  'normal': 300ms,
  'slow': 500ms,
);

// 导出 CSS 变量
:root {
  @each $name, $value in $colors {
    --color-#{$name}: #{$value};
  }
  
  @each $name, $value in $font-sizes {
    --font-#{$name}: #{$value};
  }
  
  @each $name, $value in $spacing {
    --spacing-#{$name}: #{$value};
  }
  
  @each $name, $value in $radius {
    --radius-#{$name}: #{$value};
  }
}
```

- [ ] **Step 2: 创建全局样式文件**

```scss
// 全局样式重置和基础样式

// CSS Reset
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html {
  font-size: 16px;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

body {
  font-family: $font-family-body;
  background-color: var(--color-background);
  color: var(--color-foreground);
  line-height: 1.5;
  min-height: 100vh;
}

// 标题字体
h1, h2, h3, h4, h5, h6 {
  font-family: $font-family-heading;
  font-weight: 600;
  line-height: 1.2;
}

// 链接样式
a {
  color: var(--color-accent);
  text-decoration: none;
  transition: opacity var(--transition-normal, 300ms) ease;
  
  &:hover {
    opacity: 0.8;
  }
}

// 按钮点击反馈
button, [role="button"] {
  cursor: pointer;
  transition: all var(--transition-fast, 150ms) ease;
  
  &:active {
    transform: scale(0.98);
  }
}

// 焦点可见
:focus-visible {
  outline: 2px solid var(--color-ring);
  outline-offset: 2px;
}

// 减少动画
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}

// Element Plus 暗色主题覆盖
html.dark {
  --el-bg-color: var(--color-background);
  --el-bg-color-page: var(--color-background);
  --el-bg-color-overlay: var(--color-muted);
  --el-text-color-primary: var(--color-foreground);
  --el-text-color-regular: var(--color-foreground);
  --el-text-color-secondary: var(--color-foreground);
  --el-border-color: var(--color-border);
  --el-color-primary: var(--color-accent);
}

// 滚动条样式
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: var(--color-muted);
}

::-webkit-scrollbar-thumb {
  background: var(--color-border);
  border-radius: var(--radius-full);
  
  &:hover {
    background: var(--color-secondary);
  }
}
```

- [ ] **Step 3: 创建样式入口文件**

```typescript
// 样式入口文件
import './global.scss'

// 导入 Google Fonts (在 index.html 中也可以通过 link 标签引入)
export const loadFonts = () => {
  const link = document.createElement('link')
  link.rel = 'stylesheet'
  link.href = 'https://fonts.googleapis.com/css2?family=Fira+Code:wght@400;500;600;700&family=Fira+Sans:wght@300;400;500;600;700&display=swap'
  document.head.appendChild(link)
}
```

### Task 5: 创建常量定义和工具函数

**Files:**
- Create: `web/src/utils/constants.ts`
- Create: `web/src/utils/index.ts`

- [ ] **Step 1: 创建常量定义文件**

```typescript
// 应用常量定义

// API 配置
export const API_CONFIG = {
  BASE_URL: import.meta.env.VITE_API_BASE_URL || '/api',
  TIMEOUT: 10000, // 10 秒超时
}

// WebSocket 配置
export const WS_CONFIG = {
  BASE_URL: import.meta.env.VITE_WS_BASE_URL || 'ws://localhost:8081',
  RECONNECT_INTERVAL: 3000, // 3 秒重连
  MAX_RECONNECT_ATTEMPTS: 5,
  HEARTBEAT_INTERVAL: 30000, // 30 秒心跳
}

// 路由配置
export const ROUTE_CONFIG = {
  LOGIN_PATH: '/login',
  HOME_PATH: '/',
  DASHBOARD_PATH: '/dashboard',
  DRILL_PATH: '/drill',
  TEMPLATE_PATH: '/template',
}

// 用户角色
export const USER_ROLES = {
  ADMIN: 'admin',
  DIRECTOR: 'director',
  EXECUTOR: 'executor',
  VIEWER: 'viewer',
} as const

export type UserRole = (typeof USER_ROLES)[keyof typeof USER_ROLES]

// 演练状态
export const DRILL_STATUS = {
  PENDING: 'pending',
  RUNNING: 'running',
  PAUSED: 'paused',
  COMPLETED: 'completed',
  TERMINATED: 'terminated',
} as const

export type DrillStatus = (typeof DRILL_STATUS)[keyof typeof DRILL_STATUS]

// 步骤状态
export const STEP_STATUS = {
  PENDING: 'pending',
  RUNNING: 'running',
  COMPLETED: 'completed',
  TIMEOUT: 'timeout',
  SKIPPED: 'skipped',
  ISSUE: 'issue',
} as const

export type StepStatus = (typeof STEP_STATUS)[keyof typeof STEP_STATUS]

// 状态颜色映射
export const STATUS_COLORS: Record<string, string> = {
  [DRILL_STATUS.PENDING]: '#6B7280',
  [DRILL_STATUS.RUNNING]: '#3B82F6',
  [DRILL_STATUS.PAUSED]: '#F59E0B',
  [DRILL_STATUS.COMPLETED]: '#22C55E',
  [DRILL_STATUS.TERMINATED]: '#EF4444',
  [STEP_STATUS.PENDING]: '#6B7280',
  [STEP_STATUS.RUNNING]: '#3B82F6',
  [STEP_STATUS.COMPLETED]: '#22C55E',
  [STEP_STATUS.TIMEOUT]: '#F59E0B',
  [STEP_STATUS.SKIPPED]: '#6B7280',
  [STEP_STATUS.ISSUE]: '#EF4444',
}

// 分页配置
export const PAGINATION_CONFIG = {
  DEFAULT_PAGE_SIZE: 10,
  PAGE_SIZE_OPTIONS: [10, 20, 50, 100],
}

// 本地存储 key
export const STORAGE_KEYS = {
  TOKEN: 'drill_platform_token',
  USER_INFO: 'drill_platform_user',
  THEME: 'drill_platform_theme',
} as const
```

- [ ] **Step 2: 创建工具函数索引文件**

```typescript
// 工具函数导出
export * from './constants'

// 格式化函数
export const formatDate = (date: Date | string | number, format = 'YYYY-MM-DD HH:mm:ss'): string => {
  const d = new Date(date)
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hours = String(d.getHours()).padStart(2, '0')
  const minutes = String(d.getMinutes()).padStart(2, '0')
  const seconds = String(d.getSeconds()).padStart(2, '0')
  
  return format
    .replace('YYYY', String(year))
    .replace('MM', month)
    .replace('DD', day)
    .replace('HH', hours)
    .replace('mm', minutes)
    .replace('ss', seconds)
}

// 休眠函数
export const sleep = (ms: number): Promise<void> => {
  return new Promise(resolve => setTimeout(resolve, ms))
}

// 防抖函数
export const debounce = <T extends (...args: any[]) => any>(
  func: T,
  wait: number
): ((...args: Parameters<T>) => void) => {
  let timeout: ReturnType<typeof setTimeout> | null = null
  return (...args: Parameters<T>) => {
    if (timeout) clearTimeout(timeout)
    timeout = setTimeout(() => func(...args), wait)
  }
}

// 节流函数
export const throttle = <T extends (...args: any[]) => any>(
  func: T,
  limit: number
): ((...args: Parameters<T>) => void) => {
  let inThrottle: boolean = false
  return (...args: Parameters<T>) => {
    if (!inThrottle) {
      func(...args)
      inThrottle = true
      setTimeout(() => (inThrottle = false), limit)
    }
  }
}

// 深拷贝函数
export const deepClone = <T>(obj: T): T => {
  if (obj === null || typeof obj !== 'object') return obj
  if (Array.isArray(obj)) return obj.map(item => deepClone(item)) as T
  return Object.fromEntries(
    Object.entries(obj).map(([key, value]) => [key, deepClone(value)])
  ) as T
}
```

### Task 6: 创建 Axios 请求封装

**Files:**
- Create: `web/src/api/request.ts`
- Create: `web/src/api/index.ts`

- [ ] **Step 1: 创建 Axios 封装**

```typescript
import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse, AxiosError } from 'axios'
import { API_CONFIG } from '@/utils/constants'
import { useUserStore } from '@/stores/user'

// 响应数据结构
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

// 创建 axios 实例
const createAxiosInstance = (): AxiosInstance => {
  const instance = axios.create({
    baseURL: API_CONFIG.BASE_URL,
    timeout: API_CONFIG.TIMEOUT,
    headers: {
      'Content-Type': 'application/json',
    },
  })

  // 请求拦截器
  instance.interceptors.request.use(
    (config) => {
      const userStore = useUserStore()
      const token = userStore.token
      
      if (token) {
        config.headers.Authorization = `Bearer ${token}`
      }
      
      return config
    },
    (error: AxiosError) => {
      return Promise.reject(error)
    }
  )

  // 响应拦截器
  instance.interceptors.response.use(
    (response: AxiosResponse<ApiResponse>) => {
      const { code, message, data } = response.data
      
      // 业务错误处理
      if (code !== 200) {
        ElMessage.error(message || '请求失败')
        
        // 401 未授权，跳转到登录页
        if (code === 401) {
          const userStore = useUserStore()
          userStore.logout()
          window.location.href = '/login'
        }
        
        return Promise.reject(new Error(message))
      }
      
      return response.data
    },
    (error: AxiosError) => {
      // HTTP 错误处理
      const status = error.response?.status
      const message = error.response?.data || error.message
      
      switch (status) {
        case 400:
          ElMessage.error('请求参数错误')
          break
        case 401:
          ElMessage.error('未授权，请重新登录')
          const userStore = useUserStore()
          userStore.logout()
          window.location.href = '/login'
          break
        case 403:
          ElMessage.error('拒绝访问')
          break
        case 404:
          ElMessage.error('请求资源不存在')
          break
        case 500:
          ElMessage.error('服务器内部错误')
          break
        case 502:
          ElMessage.error('网关错误')
          break
        case 503:
          ElMessage.error('服务不可用')
          break
        case 504:
          ElMessage.error('网关超时')
          break
        default:
          ElMessage.error(message || '网络错误，请稍后重试')
      }
      
      return Promise.reject(error)
    }
  )

  return instance
}

// 导出实例
export const request = createAxiosInstance()

// 封装请求方法
export const api = {
  get<T = any>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return request.get(url, config)
  },
  
  post<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return request.post(url, data, config)
  },
  
  put<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return request.put(url, data, config)
  },
  
  patch<T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return request.patch(url, data, config)
  },
  
  delete<T = any>(url: string, config?: AxiosRequestConfig): Promise<ApiResponse<T>> {
    return request.delete(url, config)
  },
}

export default request
```

- [ ] **Step 2: 创建 API 导出文件**

```typescript
// API 模块导出
export * from './request'
```

### Task 7: 创建 Pinia 状态管理 - User Store

**Files:**
- Create: `web/src/stores/user.ts`

- [ ] **Step 1: 创建用户状态管理**

```typescript
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { STORAGE_KEYS, USER_ROLES, type UserRole } from '@/utils/constants'
import { api, type ApiResponse } from '@/api/request'

// 用户信息接口
export interface UserInfo {
  id: number
  username: string
  nickname: string
  email: string
  role: UserRole
  avatar?: string
  created_at: string
  updated_at: string
}

// 登录请求参数
export interface LoginParams {
  username: string
  password: string
}

// 登录响应数据
export interface LoginResponse {
  token: string
  user: UserInfo
}

export const useUserStore = defineStore('user', () => {
  // 状态
  const token = ref<string>('')
  const userInfo = ref<UserInfo | null>(null)
  
  // 从本地存储恢复状态
  const restoreFromStorage = () => {
    const storedToken = localStorage.getItem(STORAGE_KEYS.TOKEN)
    const storedUser = localStorage.getItem(STORAGE_KEYS.USER_INFO)
    
    if (storedToken) {
      token.value = storedToken
    }
    
    if (storedUser) {
      try {
        userInfo.value = JSON.parse(storedUser)
      } catch (e) {
        console.error('解析用户信息失败:', e)
      }
    }
  }
  
  // 持久化到本地存储
  const persistToStorage = () => {
    if (token.value) {
      localStorage.setItem(STORAGE_KEYS.TOKEN, token.value)
    }
    
    if (userInfo.value) {
      localStorage.setItem(STORAGE_KEYS.USER_INFO, JSON.stringify(userInfo.value))
    }
  }
  
  // 清除本地存储
  const clearStorage = () => {
    localStorage.removeItem(STORAGE_KEYS.TOKEN)
    localStorage.removeItem(STORAGE_KEYS.USER_INFO)
  }
  
  // 计算属性
  const isLoggedIn = computed(() => !!token.value && !!userInfo.value)
  const isAdmin = computed(() => userInfo.value?.role === USER_ROLES.ADMIN)
  const isDirector = computed(() => userInfo.value?.role === USER_ROLES.DIRECTOR)
  const isExecutor = computed(() => userInfo.value?.role === USER_ROLES.EXECUTOR)
  const isViewer = computed(() => userInfo.value?.role === USER_ROLES.VIEWER)
  const username = computed(() => userInfo.value?.username || '')
  const nickname = computed(() => userInfo.value?.nickname || '')
  const role = computed(() => userInfo.value?.role || '')
  
  // 登录
  const login = async (params: LoginParams): Promise<void> => {
    try {
      const response = await api.post<LoginResponse>('/auth/login', params)
      const { token: newToken, user } = response.data
      
      token.value = newToken
      userInfo.value = user
      
      persistToStorage()
      
      ElMessage.success('登录成功')
    } catch (error) {
      throw error
    }
  }
  
  // 登出
  const logout = (): void => {
    token.value = ''
    userInfo.value = null
    clearStorage()
    
    ElMessage.success('已退出登录')
  }
  
  // 获取用户信息
  const getUserInfo = async (): Promise<void> => {
    try {
      const response = await api.get<UserInfo>('/user/info')
      userInfo.value = response.data
      persistToStorage()
    } catch (error) {
      logout()
      throw error
    }
  }
  
  // 更新用户信息
  const updateUserInfo = (info: Partial<UserInfo>): void => {
    if (userInfo.value) {
      userInfo.value = { ...userInfo.value, ...info }
      persistToStorage()
    }
  }
  
  // 初始化
  restoreFromStorage()
  
  return {
    // 状态
    token,
    userInfo,
    // 计算属性
    isLoggedIn,
    isAdmin,
    isDirector,
    isExecutor,
    isViewer,
    username,
    nickname,
    role,
    // 方法
    login,
    logout,
    getUserInfo,
    updateUserInfo,
  }
})
```

### Task 8: 创建 Pinia 状态管理 - WebSocket Store

**Files:**
- Create: `web/src/stores/websocket.ts`

- [ ] **Step 1: 创建 WebSocket 状态管理**

```typescript
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { WS_CONFIG } from '@/utils/constants'
import { useUserStore } from './user'

// WebSocket 消息类型
export interface WSMessage {
  event_type: string
  drill_id?: number
  payload: any
  timestamp: number
}

// WebSocket 事件类型
export type WSEventType = 
  | 'drill_start'
  | 'drill_complete'
  | 'drill_terminate'
  | 'step_start'
  | 'step_complete'
  | 'step_issue'
  | 'task_assigned'
  | 'system_message'

// 事件处理器类型
export type WSEventHandler = (message: WSMessage) => void

export const useWebSocketStore = defineStore('websocket', () => {
  // 状态
  const ws = ref<WebSocket | null>(null)
  const isConnected = ref(false)
  const isConnecting = ref(false)
  const reconnectAttempts = ref(0)
  const lastMessage = ref<WSMessage | null>(null)
  
  // 事件处理器映射
  const eventHandlers = ref<Map<string, Set<WSEventHandler>>>(new Map())
  
  // 心跳定时器
  let heartbeatTimer: ReturnType<typeof setInterval> | null = null
  
  // 计算属性
  const connectionStatus = computed(() => {
    if (isConnecting.value) return 'connecting'
    if (isConnected.value) return 'connected'
    return 'disconnected'
  })
  
  // 注册事件处理器
  const on = (eventType: string, handler: WSEventHandler): void => {
    if (!eventHandlers.value.has(eventType)) {
      eventHandlers.value.set(eventType, new Set())
    }
    eventHandlers.value.get(eventType)!.add(handler)
  }
  
  // 移除事件处理器
  const off = (eventType: string, handler: WSEventHandler): void => {
    const handlers = eventHandlers.value.get(eventType)
    if (handlers) {
      handlers.delete(handler)
    }
  }
  
  // 触发事件
  const emit = (message: WSMessage): void => {
    lastMessage.value = message
    
    // 触发特定事件处理器
    const handlers = eventHandlers.value.get(message.event_type)
    handlers?.forEach(handler => handler(message))
    
    // 触发通用消息处理器
    const allHandlers = eventHandlers.value.get('*')
    allHandlers?.forEach(handler => handler(message))
  }
  
  // 发送消息
  const send = (data: any): void => {
    if (ws.value && ws.value.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify(data))
    } else {
      console.warn('WebSocket 未连接，无法发送消息')
    }
  }
  
  // 连接 WebSocket
  const connect = (drillId?: number): void => {
    if (isConnected.value || isConnecting.value) {
      return
    }
    
    isConnecting.value = true
    
    const userStore = useUserStore()
    const token = userStore.token
    
    // 构建连接 URL
    let url = `${WS_CONFIG.BASE_URL}/ws/tasks`
    if (drillId) {
      url = `${WS_CONFIG.BASE_URL}/ws/display/${drillId}`
    }
    
    // 添加 token 参数
    const separator = url.includes('?') ? '&' : '?'
    url = `${url}${separator}token=${encodeURIComponent(token)}`
    
    try {
      ws.value = new WebSocket(url)
      
      ws.value.onopen = () => {
        isConnected.value = true
        isConnecting.value = false
        reconnectAttempts.value = 0
        console.log('WebSocket 连接成功')
        ElMessage.success('实时通信已连接')
        
        // 启动心跳
        startHeartbeat()
      }
      
      ws.value.onmessage = (event) => {
        try {
          const message: WSMessage = JSON.parse(event.data)
          emit(message)
        } catch (error) {
          console.error('解析 WebSocket 消息失败:', error)
        }
      }
      
      ws.value.onerror = (error) => {
        console.error('WebSocket 错误:', error)
        ElMessage.error('实时通信连接失败')
      }
      
      ws.value.onclose = () => {
        isConnected.value = false
        isConnecting.value = false
        console.log('WebSocket 连接关闭')
        
        // 停止心跳
        stopHeartbeat()
        
        // 尝试重连
        attemptReconnect(drillId)
      }
    } catch (error) {
      console.error('创建 WebSocket 连接失败:', error)
      isConnecting.value = false
    }
  }
  
  // 断开连接
  const disconnect = (): void => {
    stopHeartbeat()
    
    if (ws.value) {
      ws.value.close()
      ws.value = null
    }
    
    isConnected.value = false
    isConnecting.value = false
    reconnectAttempts.value = 0
    eventHandlers.value.clear()
    
    console.log('WebSocket 已断开')
  }
  
  // 尝试重连
  const attemptReconnect = (drillId?: number): void => {
    if (reconnectAttempts.value >= WS_CONFIG.MAX_RECONNECT_ATTEMPTS) {
      console.log('达到最大重连次数，停止重连')
      ElMessage.warning('实时通信断开，请检查网络连接')
      return
    }
    
    reconnectAttempts.value++
    const delay = WS_CONFIG.RECONNECT_INTERVAL * reconnectAttempts.value
    
    console.log(`准备重连 (${reconnectAttempts.value}/${WS_CONFIG.MAX_RECONNECT_ATTEMPTS}), 延迟 ${delay}ms`)
    
    setTimeout(() => {
      connect(drillId)
    }, delay)
  }
  
  // 启动心跳
  const startHeartbeat = (): void => {
    stopHeartbeat()
    
    heartbeatTimer = setInterval(() => {
      if (ws.value && ws.value.readyState === WebSocket.OPEN) {
        send({ event_type: 'heartbeat', timestamp: Date.now() })
      }
    }, WS_CONFIG.HEARTBEAT_INTERVAL)
  }
  
  // 停止心跳
  const stopHeartbeat = (): void => {
    if (heartbeatTimer) {
      clearInterval(heartbeatTimer)
      heartbeatTimer = null
    }
  }
  
  // 订阅演练状态
  const subscribeDrill = (drillId: number): void => {
    connect(drillId)
  }
  
  // 取消订阅
  const unsubscribeDrill = (): void => {
    disconnect()
  }
  
  return {
    // 状态
    ws,
    isConnected,
    isConnecting,
    reconnectAttempts,
    lastMessage,
    // 计算属性
    connectionStatus,
    // 方法
    on,
    off,
    send,
    connect,
    disconnect,
    subscribeDrill,
    unsubscribeDrill,
  }
})
```

### Task 9: 创建 Pinia Store 索引

**Files:**
- Create: `web/src/stores/index.ts`

- [ ] **Step 1: 创建 Store 导出文件**

```typescript
import { createPinia } from 'pinia'
import type { App } from 'vue'

// 创建 pinia 实例
const pinia = createPinia()

// 安装 pinia
export function setupStore(app: App) {
  app.use(pinia)
}

// 导出 store
export * from './user'
export * from './websocket'

export default pinia
```

### Task 10: 创建路由配置

**Files:**
- Create: `web/src/router/index.ts`
- Create: `web/src/router/types.ts`

- [ ] **Step 1: 创建路由类型定义**

```typescript
import type { RouteRecordRaw } from 'vue-router'
import { USER_ROLES } from '@/utils/constants'

// 路由元信息
export interface RouteMeta {
  title?: string
  icon?: string
  requiresAuth?: boolean
  roles?: (typeof USER_ROLES)[keyof typeof USER_ROLES][]
  keepAlive?: boolean
  hidden?: boolean
  breadcrumb?: boolean
}

// 路由记录
export type AppRouteRecord = RouteRecordRaw & {
  meta?: RouteMeta
  children?: AppRouteRecord[]
}
```

- [ ] **Step 2: 创建路由配置**

```typescript
import { createRouter, createWebHistory, type RouteLocationNormalized } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ROUTE_CONFIG } from '@/utils/constants'
import type { AppRouteRecord } from './types'

// 常量路由 (公开)
const constantRoutes: AppRouteRecord[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
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
    component: () => import('@/layouts/default/index.vue'),
    redirect: '/dashboard',
    meta: {
      requiresAuth: true,
    },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/index.vue'),
        meta: {
          title: '指挥中心',
          icon: 'dashboard',
          requiresAuth: true,
          keepAlive: true,
        },
      },
      {
        path: 'drill',
        name: 'Drill',
        component: () => import('@/views/drill/index.vue'),
        meta: {
          title: '演练管理',
          icon: 'drill',
          requiresAuth: true,
          keepAlive: true,
        },
      },
      {
        path: 'drill/:id',
        name: 'DrillDetail',
        component: () => import('@/views/drill/detail.vue'),
        meta: {
          title: '演练详情',
          hidden: true,
          requiresAuth: true,
          breadcrumb: true,
        },
      },
      {
        path: 'template',
        name: 'Template',
        component: () => import('@/views/template/index.vue'),
        meta: {
          title: '模板管理',
          icon: 'template',
          requiresAuth: true,
          keepAlive: true,
        },
      },
      {
        path: 'user',
        name: 'User',
        component: () => import('@/views/user/index.vue'),
        meta: {
          title: '用户管理',
          icon: 'user',
          requiresAuth: true,
          roles: ['admin'],
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
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) {
      return savedPosition
    } else {
      return { top: 0 }
    }
  },
})

// 路由守卫
router.beforeEach(async (to: RouteLocationNormalized, from: RouteLocationNormalized) => {
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
  
  // 检查角色权限
  const allowedRoles = to.meta?.roles
  if (allowedRoles && allowedRoles.length > 0) {
    const userRole = userStore.role
    if (!userRole || !allowedRoles.includes(userRole)) {
      return { path: '/403' }
    }
  }
  
  // 如果已登录且访问登录页，重定向到首页
  if (isLoggedIn && to.path === ROUTE_CONFIG.LOGIN_PATH) {
    return { path: ROUTE_CONFIG.HOME_PATH }
  }
})

export default router
export { constantRoutes, dynamicRoutes }
```

### Task 11: 创建主入口文件

**Files:**
- Create: `web/src/main.ts`

- [ ] **Step 1: 创建 Vue3 入口文件**

```typescript
import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

import App from './App.vue'
import { setupStore } from './stores'
import router from './router'
import { loadFonts } from './styles'

// 创建应用实例
const app = createApp(App)

// 安装 Pinia
setupStore(app)

// 安装 Vue Router
app.use(router)

// 安装 Element Plus
app.use(ElementPlus, {
  locale: zhCn,
  size: 'default',
})

// 注册 Element Plus 图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

// 加载字体
loadFonts()

// 挂载应用
app.mount('#app')

// 导出 app 实例 (用于测试)
export default app
```

### Task 12: 创建根组件

**Files:**
- Create: `web/src/App.vue`

- [ ] **Step 1: 创建根组件**

```vue
<template>
  <router-view />
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useUserStore } from '@/stores/user'

// 初始化用户状态
const userStore = useUserStore()

onMounted(() => {
  // 如果已登录，刷新用户信息
  if (userStore.isLoggedIn) {
    userStore.getUserInfo().catch(console.error)
  }
})
</script>

<style>
#app {
  width: 100%;
  height: 100%;
}
</style>
```

### Task 13: 创建 HTML 入口文件

**Files:**
- Create: `web/index.html`

- [ ] **Step 1: 创建 HTML 文件**

```html
<!DOCTYPE html>
<html lang="zh-CN">
  <head>
    <meta charset="UTF-8" />
    <link rel="icon" type="image/svg+xml" href="/favicon.svg" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta name="description" content="生产演练流程管理系统 - IT 故障演练流程管理平台" />
    <meta name="theme-color" content="#0F172A" />
    <title>生产演练平台</title>
    
    <!-- Google Fonts -->
    <link rel="preconnect" href="https://fonts.googleapis.com" />
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
    <link href="https://fonts.googleapis.com/css2?family=Fira+Code:wght@400;500;600;700&family=Fira+Sans:wght@300;400;500;600;700&display=swap" rel="stylesheet" />
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="/src/main.ts"></script>
  </body>
</html>
```

### Task 14: 创建环境配置文件

**Files:**
- Create: `web/.env`
- Create: `web/.env.production`

- [ ] **Step 1: 创建开发环境配置**

```bash
# 开发环境配置
VITE_API_BASE_URL=/api
VITE_WS_BASE_URL=ws://localhost:8081
```

- [ ] **Step 2: 创建生产环境配置**

```bash
# 生产环境配置
VITE_API_BASE_URL=/api
VITE_WS_BASE_URL=wss://your-domain.com:8081
```

### Task 15: 创建空目录占位文件

**Files:**
- Create: `web/src/components/.gitkeep`
- Create: `web/src/views/.gitkeep`
- Create: `web/src/views/login/.gitkeep`
- Create: `web/src/views/dashboard/.gitkeep`
- Create: `web/src/views/drill/.gitkeep`
- Create: `web/src/views/template/.gitkeep`
- Create: `web/src/views/user/.gitkeep`
- Create: `web/src/views/error/.gitkeep`
- Create: `web/src/layouts/.gitkeep`
- Create: `web/src/layouts/default/.gitkeep`

- [ ] **Step 1: 创建占位文件**

```bash
touch web/src/components/.gitkeep
touch web/src/views/.gitkeep
touch web/src/views/login/.gitkeep
touch web/src/views/dashboard/.gitkeep
touch web/src/views/drill/.gitkeep
touch web/src/views/template/.gitkeep
touch web/src/views/user/.gitkeep
touch web/src/views/error/.gitkeep
touch web/src/layouts/.gitkeep
touch web/src/layouts/default/.gitkeep
```

### Task 16: 安装依赖

**Files:**
- None (npm install)

- [ ] **Step 1: 安装依赖**

```bash
cd web && npm install
```

Expected: 所有依赖安装成功，无错误

- [ ] **Step 2: 验证安装**

```bash
cd web && npm list --depth=0
```

Expected: 显示所有安装的依赖

### Task 17: 验证项目启动

**Files:**
- None (npm run dev)

- [ ] **Step 1: 启动开发服务器**

```bash
cd web && npm run dev
```

Expected: Vite 开发服务器启动，显示访问地址

- [ ] **Step 2: 检查 TypeScript 类型**

```bash
cd web && npm run build -- --dry-run
```

Expected: TypeScript 类型检查通过，无错误

---

## 验证清单

完成所有任务后，验证以下项目：

- [ ] 目录结构完整 (web/src 下所有子目录存在)
- [ ] package.json 包含所有必需依赖
- [ ] TypeScript 配置正确 (无类型错误)
- [ ] Vite 配置正确 (代理、插件正常工作)
- [ ] Axios 封装包含请求/响应拦截器
- [ ] Pinia stores 正常工作 (user/websocket)
- [ ] Vue Router 配置正确 (路由守卫生效)
- [ ] 全局样式和变量定义完整
- [ ] 工具函数和常量定义完整
- [ ] 项目可以成功启动 (npm run dev)
- [ ] 类型检查通过 (npm run build)

---

## 提交信息

```bash
git add web/
git commit -m "feat: 初始化前端项目骨架

- 创建 Vue3 + Vite + TypeScript 项目结构
- 配置 Element Plus + Pinia + Vue Router
- 实现 Axios 请求封装 (请求/响应拦截器)
- 实现用户状态管理 (user store)
- 实现 WebSocket 状态管理 (websocket store)
- 配置路由和路由守卫
- 定义全局样式变量 (基于 Data-Dense Dashboard 设计系统)
- 定义工具函数和常量
- 支持亮色/暗色模式
- 使用 Fira Code + Fira Sans 字体"
```
