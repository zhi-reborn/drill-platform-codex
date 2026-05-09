<template>
  <router-view />
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useUserStore } from '@/stores/user'

// 初始化用户状态
const userStore = useUserStore()

onMounted(() => {
  // 恢复登录状态
  userStore.restoreLogin()
})
</script>

<style>
#app {
  width: 100%;
  height: 100%;
}
</style>
