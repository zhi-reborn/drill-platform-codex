<template>
  <div class="empty-box">
    <svg class="empty-icon" viewBox="0 0 200 150" fill="none">
      <rect x="40" y="30" width="120" height="90" rx="8" stroke="#484F58" stroke-width="2" fill="none"/>
      <line x1="70" y1="60" x2="130" y2="60" stroke="#484F58" stroke-width="2" stroke-linecap="round"/>
      <line x1="70" y1="80" x2="110" y2="80" stroke="#484F58" stroke-width="2" stroke-linecap="round"/>
      <circle cx="100" cy="110" r="4" fill="#484F58"/>
    </svg>
    <h3 class="empty-title">{{ title || '暂无数据' }}</h3>
    <p v-if="description" class="empty-desc">{{ description }}</p>
    <el-button v-if="actionText" type="primary" @click="$emit('action')">
      {{ actionText }}
    </el-button>
  </div>
</template>

<script setup lang="ts">
interface Props {
  title?: string
  description?: string
  actionText?: string
}

defineProps<Props>()
defineEmits<{ action: [] }>()
</script>

<style lang="scss" scoped>
@use '@/styles/variables' as *;
@use '@/styles/mixins' as *;

.empty-box {
  @include flex-center;
  flex-direction: column;
  padding: $spacing-4xl $spacing-xl;
  gap: $spacing-base;
  
  .empty-icon {
    width: 200px;
    height: 150px;
    opacity: 0.5;
  }
  
  .empty-title {
    font-size: $font-size-md;
    color: $text-secondary;
    font-weight: $font-weight-medium;
  }
  
  .empty-desc {
    font-size: $font-size-sm;
    color: $text-tertiary;
  }
}
</style>
