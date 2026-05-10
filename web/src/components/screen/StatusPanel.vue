<template>
  <div class="status-panel">
    <h3 class="panel-title">活跃演练</h3>
    <div class="status-list">
      <div v-for="drill in drills" :key="drill.id" class="status-card" :class="drill.status">
        <div class="card-header">
          <span class="drill-name">{{ drill.name }}</span>
          <span class="status-dot" :class="drill.status"></span>
        </div>
        <div class="card-progress">
          <div class="progress-bar">
            <div
              class="progress-fill"
              :style="{ width: `${(drill.completed_steps / drill.total_steps) * 100}%` }"
            ></div>
          </div>
          <span class="progress-text">{{ drill.completed_steps }}/{{ drill.total_steps }}</span>
        </div>
        <div class="card-footer">
          <span class="assignee" v-if="drill.created_by_name">
            <el-icon><User /></el-icon>
            {{ drill.created_by_name }}
          </span>
          <span class="template">{{ drill.template_name }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { DrillInstance } from '@/types'

defineProps<{ drills: DrillInstance[] }>()
</script>

<style scoped>
.status-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #1A1F2E;
  border-right: 1px solid #30363D;
  padding: 12px;
  overflow: hidden;
}

.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;

  &.running {
    background: #55C3D3;
    box-shadow: 0 0 6px #55C3D3;
    animation: pulse 2s ease-in-out infinite;
  }
  &.paused { background: #B8860B; }
  &.completed { background: #2EA043; }
  &.terminated { background: #DA3633; }
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.4; transform: scale(1.3); }
}

.panel-title {
  font-size: 14px;
  font-weight: 600;
  color: #8B949E;
  margin-bottom: 12px;
}

.status-list {
  flex: 1;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.status-card {
  padding: 12px;
  border-radius: 4px;
  border: 1px solid #30363D;
  background: #0D1117;
  
  &.running { border-left: 3px solid #55C3D3; }
  &.paused { border-left: 3px solid #B8860B; }
  &.completed { border-left: 3px solid #2EA043; }
  &.terminated { border-left: 3px solid #DA3633; }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  
  .drill-name {
    font-size: 13px;
    font-weight: 600;
    color: #E0E6ED;
  }
}

.card-progress {
  margin-bottom: 8px;
  
  .progress-bar {
    height: 4px;
    background: #21262D;
    border-radius: 2px;
    overflow: hidden;
    margin-bottom: 4px;
  }
  
  .progress-fill {
    height: 100%;
    background: #55C3D3;
    transition: width 0.3s ease;
  }
  
  .progress-text {
    font-size: 11px;
    color: #8B949E;
  }
}

.card-footer {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #6E7681;
  
  .assignee {
    display: flex;
    align-items: center;
    gap: 4px;
  }
}
</style>
