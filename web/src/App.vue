<template>
  <el-config-provider :locale="zhCn">
    <router-view />
  </el-config-provider>
</template>

<script setup lang="ts">
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import { useAuthStore } from '@/stores/auth'
import { usePresenceHeartbeat } from '@/composables/usePresenceHeartbeat'

const authStore = useAuthStore()
usePresenceHeartbeat(() => authStore.token)
</script>
