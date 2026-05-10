<template>
  <div class="metrics-bar">
    <div v-for="metric in metrics" :key="metric.label" class="metric-pill">
      <div class="metric-icon">
        <el-icon :size="20"><component :is="metric.icon" /></el-icon>
      </div>
      <div class="metric-content">
        <span class="metric-label">{{ metric.label }}</span>
        <span class="metric-value">{{ metric.value }}</span>
        <span v-if="metric.trend" class="metric-trend" :class="metric.trend > 0 ? 'up' : 'down'">
          {{ metric.trend > 0 ? '↑' : '↓' }}{{ Math.abs(metric.trend) }}%
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
interface Metric {
  label: string
  value: string | number
  icon: string
  trend?: number
}

defineProps<{ metrics: Metric[] }>()
</script>

<style scoped>
.metrics-bar {
  height: 80px;
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 0 24px;
  background: #161B22;
  border-bottom: 1px solid #30363D;
}

.metric-pill {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: #0D1117;
  border: 1px solid #30363D;
  border-radius: 4px;
}

.metric-icon {
  color: #55C3D3;
  display: flex;
  align-items: center;
}

.metric-content {
  display: flex;
  flex-direction: column;
}

.metric-label {
  font-size: 12px;
  color: #8B949E;
}

.metric-value {
  font-size: 24px;
  font-weight: 700;
  color: #E0E6ED;
  line-height: 1.2;
}

.metric-trend {
  font-size: 11px;
  &.up { color: #2EA043; }
  &.down { color: #DA3633; }
}
</style>
