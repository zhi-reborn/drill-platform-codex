<template>
  <div class="messages-page">
    <h1 class="page-title">消息通知</h1>
    <EmptyBox description="暂无新消息" />
  </div>
</template>

<script setup lang="ts">
import EmptyBox from '@/components/common/EmptyBox.vue'
</script>

<style lang="scss" scoped>
@use '@/styles/variables' as *;

.messages-page {
  .page-title {
    font-size: $font-size-2xl;
    font-weight: $font-weight-bold;
    color: $text-primary;
    margin-bottom: $spacing-xl;
  }
}
</style>
