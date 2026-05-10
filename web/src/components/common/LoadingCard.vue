<template>
  <div class="loading-card" :style="{ height, width }">
    <div class="skeleton-title" />
    <div class="skeleton-text" v-for="i in rows" :key="i" :style="{ width: randomWidth() }" />
  </div>
</template>

<script setup lang="ts">
interface Props {
  rows?: number
  height?: string
  width?: string
}

withDefaults(defineProps<Props>(), {
  rows: 3
})

function randomWidth(): string {
  const min = 60
  const max = 100
  return `${min + Math.floor(Math.random() * (max - min))}%`
}
</script>

<style lang="scss" scoped>
@use '@/styles/variables' as *;

.loading-card {
  padding: $spacing-base;
  background: $bg-secondary;
  border-radius: $radius-base;
  
  .skeleton-title {
    height: 20px;
    width: 60%;
    background: linear-gradient(90deg, $bg-tertiary 25%, $border-color 50%, $bg-tertiary 75%);
    background-size: 200% 100%;
    animation: shimmer 1.5s infinite;
    border-radius: $radius-sm;
    margin-bottom: $spacing-md;
  }
  
  .skeleton-text {
    height: 14px;
    margin-bottom: $spacing-sm;
    background: linear-gradient(90deg, $bg-tertiary 25%, $border-color 50%, $bg-tertiary 75%);
    background-size: 200% 100%;
    animation: shimmer 1.5s infinite;
    border-radius: $radius-sm;
  }
}

@keyframes shimmer {
  0% { background-position: -200% 0; }
  100% { background-position: 200% 0; }
}
</style>
