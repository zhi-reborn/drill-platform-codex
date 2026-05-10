<template>
  <div class="page-placeholder">
    <h1>没有权限</h1>
    <p>页面开发中...</p>
  </div>
</template>

<script setup lang="ts">
defineProps<{ title?: string }>()
</script>

<style scoped>
.page-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 400px;
  gap: 16px;
}
h1 {
  color: #E0E6ED;
  font-size: 24px;
}
p {
  color: #8B949E;
}
</style>
