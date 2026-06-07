<template>
  <div class="cas-callback-page">
    <el-card class="callback-card">
      <el-icon class="callback-icon" :class="{ error: !!error }" :size="42">
        <CircleCheck v-if="!error" />
        <CircleClose v-else />
      </el-icon>
      <h2>{{ error ? '统一认证登录失败' : '统一认证登录中' }}</h2>
      <p>{{ error || '正在写入登录态，请稍候...' }}</p>
      <el-button v-if="error" type="primary" @click="router.replace('/login')">返回登录</el-button>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { CircleCheck, CircleClose } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import type { Role, TokenResponse } from '@/types'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const error = ref('')

const roleDashboards: Record<Role, string> = {
  admin: '/admin',
  director: '/director',
  executor: '/executor',
  viewer: '/viewer',
}

function queryValue(name: string): string {
  const value = route.query[name]
  return Array.isArray(value) ? value[0] || '' : value || ''
}

onMounted(() => {
  const token = queryValue('token')
  const username = queryValue('username')
  const role = queryValue('role') as Role

  if (!token || !username || !role) {
    error.value = queryValue('error') || 'CAS 回调参数不完整'
    return
  }

  const response: TokenResponse = {
    token,
    user_id: Number(queryValue('user_id') || 0),
    username,
    real_name: queryValue('real_name') || username,
    role,
    department: queryValue('department'),
  }
  authStore.applyTokenResponse(response)
  router.replace(roleDashboards[role] || '/viewer')
})
</script>

<style lang="scss" scoped>
.cas-callback-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #0f172a;
}

.callback-card {
  width: 360px;
  text-align: center;

  h2 {
    margin: 16px 0 8px;
  }

  p {
    margin: 0 0 18px;
    color: #64748b;
  }
}

.callback-icon {
  color: #22c55e;

  &.error {
    color: #ef4444;
  }
}
</style>
