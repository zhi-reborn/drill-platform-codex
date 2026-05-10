<template>
  <div class="tasks-page">
    <h1 class="page-title">任务中心</h1>
    <EmptyBox description="暂无待处理任务" />
  </div>
</template>

<script setup lang="ts">
import EmptyBox from '@/components/common/EmptyBox.vue'
</script>

<style lang="scss" scoped>
@use '@/styles/variables' as *;

.tasks-page {
  .page-title {
    font-size: $font-size-2xl;
    font-weight: $font-weight-bold;
    color: $text-primary;
    margin-bottom: $spacing-xl;
  }
}
</style>
