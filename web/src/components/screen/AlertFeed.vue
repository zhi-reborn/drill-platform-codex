<template>
  <div class="alert-feed">
    <h3 class="panel-title">实时事件</h3>
    <div class="feed-list" ref="feedRef">
      <div v-for="alert in alerts" :key="alert.id" class="alert-item" :class="alert.level">
        <span class="alert-time">{{ formatTime(alert.created_at) }}</span>
        <span class="alert-icon">
          <el-icon :size="14"><component :is="alert.icon" /></el-icon>
        </span>
        <span class="alert-message">{{ alert.message }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface Alert {
  id: number
  level: 'info' | 'warning' | 'error'
  message: string
  icon: string
  created_at: string
}

defineProps<{ alerts: Alert[] }>()
const feedRef = ref<HTMLElement>()

function formatTime(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}
</script>

<style scoped>
.alert-feed {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #161B22;
  padding: 16px;
  overflow: hidden;
}

.feed-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.alert-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #0D1117;
  border-radius: 4px;
  font-size: 12px;
  border-left: 3px solid;

  &.info { border-color: #55C3D3; }
  &.warning { border-color: #B8860B; }
  &.error { border-color: #DA3633; }

  .alert-time { color: #8B949E; flex-shrink: 0; }
  .alert-icon { color: inherit; flex-shrink: 0; }
  .alert-message { color: #E0E6ED; }
}
</style>
