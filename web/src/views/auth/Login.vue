<template>
  <div class="login-page">
    <div class="login-container">
      <!-- Logo 和标题 -->
      <div class="login-header">
        <div class="logo">
          <svg viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg">
            <rect width="48" height="48" rx="12" fill="url(#logo-gradient)" />
            <path d="M14 24L22 32L34 16" stroke="white" stroke-width="3" stroke-linecap="round" stroke-linejoin="round" />
            <defs>
              <linearGradient id="logo-gradient" x1="0" y1="0" x2="48" y2="48">
                <stop stop-color="#3B82F6" />
                <stop offset="1" stop-color="#8B5CF6" />
              </linearGradient>
            </defs>
          </svg>
        </div>
        <h1 class="login-title">生产演练流程管理系统</h1>
        <p class="login-subtitle">Drill Platform</p>
      </div>

      <!-- 登录表单 -->
      <el-card class="login-card">
        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-position="top"
          size="large"
        >
          <el-form-item label="用户名" prop="username">
            <el-input
              v-model="form.username"
              placeholder="请输入用户名"
              prefix-icon="User"
              @keyup.enter="handleLogin"
            />
          </el-form-item>

          <el-form-item label="密码" prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="请输入密码"
              prefix-icon="Lock"
              show-password
              @keyup.enter="handleLogin"
            />
          </el-form-item>

          <el-form-item>
            <el-button
              type="primary"
              class="login-button"
              :loading="loading"
              @click="handleLogin"
            >
              登录
            </el-button>
          </el-form-item>
        </el-form>

        <!-- 提示信息 -->
        <div class="login-tips">
          <p>默认账户：</p>
          <p><code>admin</code> / <code>admin123</code></p>
        </div>
      </el-card>

      <!-- 版权信息 -->
      <div class="login-footer">
        <p>&copy; 2024 Drill Platform. All rights reserved.</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { post } from '@/api/request'

const router = useRouter()
const route = useRoute()

const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  username: '',
  password: ''
})

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少 6 位', trigger: 'blur' }
  ]
}

// 登录
const handleLogin = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (!valid) return

    loading.value = true
    try {
      // 调用登录 API
      const data = await post('/auth/login', {
        username: form.username,
        password: form.password
      })

      // 保存 token
      if (data?.token) {
        localStorage.setItem('token', data.token)
      }

      ElMessage.success('登录成功')

      // 跳转
      const redirect = route.query.redirect as string
      if (redirect) {
        router.push(redirect)
      } else {
        router.push('/display')
      }
    } catch (error: any) {
      // 错误已在 axios 拦截器中处理
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: linear-gradient(135deg, #020617 0%, #0F172A 50%, #1E293B 100%);
  padding: 20px;
}

.login-container {
  width: 100%;
  max-width: 420px;
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.logo {
  display: inline-flex;
  margin-bottom: 20px;
}

.logo svg {
  width: 64px;
  height: 64px;
}

.login-title {
  font-size: 24px;
  font-weight: 700;
  margin: 0;
  color: var(--color-foreground, #F8FAFC);
  font-family: 'Fira Sans', sans-serif;
}

.login-subtitle {
  font-size: 14px;
  color: var(--color-muted-foreground, #94A3B8);
  margin-top: 8px;
  letter-spacing: 2px;
}

.login-card {
  background-color: var(--color-muted, #1A1E2F);
  border: 1px solid var(--color-border, #334155);
}

:deep(.el-form-item__label) {
  color: var(--color-foreground, #F8FAFC);
  font-weight: 500;
}

:deep(.el-input__wrapper) {
  background-color: var(--color-secondary, #1E293B);
  border: 1px solid var(--color-border, #334155);
}

:deep(.el-input__inner) {
  color: var(--color-foreground, #F8FAFC);
}

.login-button {
  width: 100%;
  height: 44px;
  font-size: 16px;
  font-weight: 600;
}

.login-tips {
  margin-top: 16px;
  padding: 12px;
  background-color: var(--color-secondary, #1E293B);
  border-radius: 6px;
  font-size: 13px;
  color: var(--color-muted-foreground, #94A3B8);
  text-align: center;
}

.login-tips code {
  display: inline-block;
  margin-top: 4px;
  padding: 4px 8px;
  background-color: var(--color-background, #020617);
  border-radius: 4px;
  font-family: 'Fira Code', monospace;
  color: var(--color-accent, #22C55E);
}

.login-footer {
  margin-top: 24px;
  text-align: center;
}

.login-footer p {
  font-size: 12px;
  color: var(--color-muted-foreground, #94A3B8);
}
</style>
