<template>
  <span class="status-badge" :class="[type, status]">
    <span class="status-dot" :class="{ pulse: isRunning }"></span>
    <span class="status-text">{{ label }}</span>
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  status: string
  type?: 'drill' | 'step'
}>()

const drillStatusMap: Record<string, { label: string; color: string }> = {
  pending: { label: '待启动', color: '#6E7681' },
  running: { label: '执行中', color: '#55C3D3' },
  paused: { label: '已暂停', color: '#B8860B' },
  completed: { label: '已完成', color: '#2EA043' },
  terminated: { label: '已终止', color: '#DA3633' },
}

const stepStatusMap: Record<string, { label: string; color: string }> = {
  pending: { label: '待执行', color: '#6E7681' },
  running: { label: '执行中', color: '#55C3D3' },
  completed: { label: '已完成', color: '#2EA043' },
  timeout: { label: '已超时', color: '#B8860B' },
  skipped: { label: '已跳过', color: '#6E7681' },
  issue: { label: '异常', color: '#DA3633' },
}

const statusMap = props.type === 'step' ? stepStatusMap : drillStatusMap
const info = computed(() => statusMap[props.status] ?? { label: props.status, color: '#6E7681' })
const label = computed(() => info.value.label)
const isRunning = computed(() => props.status === 'running')
</script>

<style lang="scss" scoped>
@use '@/styles/variables' as *;

.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: $font-size-sm;
  
  .status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: v-bind('info.color');
    
    &.pulse {
      animation: pulse 2s ease-in-out infinite;
    }
  }
  
  .status-text {
    color: v-bind('info.color');
    font-weight: $font-weight-medium;
  }
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.5; transform: scale(1.2); }
}
</style>
