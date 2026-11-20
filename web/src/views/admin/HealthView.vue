<template>
  <div class="page-container">
    <div class="page-header">
      <h2>节点健康</h2>
      <div class="header-actions">
        <span class="auto-refresh-label">自动刷新</span>
        <el-switch v-model="autoRefresh" />
        <el-button
          type="primary"
          size="small"
          :loading="loading"
          @click="fetchHealth"
        >
          刷新
        </el-button>
      </div>
    </div>

    <div v-loading="loading && !health" class="page-content health-content">
      <template v-if="health">
        <!-- 集群总览横幅 -->
        <div class="status-hero" :class="health.overall">
          <div class="hero-orb">
            <span class="orb-core" />
          </div>
          <div class="hero-main">
            <div class="hero-title">
              {{ health.overall === 'healthy' ? '系统运行正常' : '系统状态降级' }}
            </div>
            <div class="hero-sub">
              <span class="hero-count">{{ onlineCount }}/{{ health.nodes.length }} 个节点在线</span>
              <span v-if="health.warning" class="hero-warning">{{ health.warning }}</span>
            </div>
          </div>
          <div class="hero-chips">
            <span
              v-for="n in health.nodes"
              :key="n.instance_id"
              class="cluster-chip"
              :class="n.state"
            >
              <span class="chip-dot" />{{ n.instance_id }}
            </span>
          </div>
        </div>

        <!-- 节点卡片 -->
        <el-row :gutter="20" class="node-row">
          <el-col
            v-for="(node, i) in health.nodes"
            :key="node.instance_id"
            :xs="24"
            :md="12"
            :xl="8"
          >
            <div class="node-card" :class="cardClass(node)" :style="{ '--i': String(i) }">
              <div class="card-strip" />

              <div class="node-header">
                <div class="node-title">
                  <span class="node-name">{{ node.instance_id }}</span>
                  <span class="role-chip">{{ roleLabel(node.role) }}</span>
                  <span class="state-chip" :class="node.state === 'online' ? 'online' : 'offline'">
                    <span class="live-dot" />
                    {{ node.state === 'online' ? '在线' : '离线' }}
                  </span>
                  <span v-if="hasFail(node)" class="state-chip warn">组件异常</span>
                </div>
                <div class="node-sub">
                  <span>Worker · {{ workerLabel(node.worker_status) }}</span>
                  <span>心跳 <span class="mono">{{ formatTime(node.heartbeat_at) }}</span></span>
                </div>
              </div>

              <!-- 组件状态 -->
              <div class="components">
                <div
                  v-for="name in componentNames(node)"
                  :key="name"
                  class="component-item"
                  :class="{ fail: compFailed(node, name) }"
                >
                  <span class="comp-label">{{ compLabel(name) }}</span>
                  <span
                    v-if="node.components[name]"
                    class="comp-status"
                    :class="node.components[name].status === 'ok' ? 'ok' : 'fail'"
                  >
                    <span class="comp-dot" />
                    {{ node.components[name].status === 'ok' ? '正常' : '异常' }}
                    <span
                      v-if="node.components[name].status === 'ok' && node.components[name].latency_ms > 0"
                      class="comp-latency"
                    >{{ node.components[name].latency_ms }}ms</span>
                  </span>
                  <span v-else class="comp-status none"><span class="comp-dot" />未启用</span>
                </div>
              </div>

              <!-- 运行指标 -->
              <div class="metrics">
                <div class="metric">
                  <span class="metric-value mono">{{ formatUptime(node.uptime_seconds) }}</span>
                  <span class="metric-label">运行时长</span>
                </div>
                <div class="metric">
                  <span class="metric-value mono">{{ node.goroutines }}</span>
                  <span class="metric-label">协程数</span>
                </div>
                <div class="metric">
                  <span class="metric-value mono">{{ formatBytes(node.heap_alloc_bytes) }}</span>
                  <span class="metric-label">堆内存</span>
                </div>
                <div class="metric">
                  <span class="metric-value mono">{{ formatBytes(node.sys_bytes) }}</span>
                  <span class="metric-label">进程内存</span>
                </div>
                <div class="metric">
                  <span class="metric-value mono">{{ wsTotal(node) }}</span>
                  <span class="metric-label">WS 连接</span>
                </div>
                <div class="metric">
                  <span class="metric-value mono">
                    {{ node.db_pool ? `${node.db_pool.in_use}/${node.db_pool.open}` : '—' }}
                  </span>
                  <span class="metric-label">DB 池 使用/总</span>
                </div>
              </div>
            </div>
          </el-col>
        </el-row>

        <div class="updated-at">
          <span class="live-flag" :class="{ on: autoRefresh }">
            <span class="flag-dot" />{{ autoRefresh ? 'LIVE' : '已暂停' }}
          </span>
          数据更新于 <span class="mono">{{ formatTime(health.updated_at) }}</span> · 每 10 秒自动刷新
        </div>
      </template>

      <el-empty v-if="!loading && !health" description="暂无健康数据" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { healthApi } from '@/api'
import type { SystemHealth, NodeHealth } from '@/api/modules/health'

const REFRESH_INTERVAL = 10 * 1000

const health = ref<SystemHealth | null>(null)
const loading = ref(false)
const autoRefresh = ref(true)
let timer: ReturnType<typeof setInterval> | null = null

const onlineCount = computed(
  () => health.value?.nodes.filter(n => n.state === 'online').length ?? 0
)

async function fetchHealth() {
  loading.value = true
  try {
    health.value = await healthApi.getSystemHealth()
  } finally {
    loading.value = false
  }
}

function startTimer() {
  stopTimer()
  timer = setInterval(fetchHealth, REFRESH_INTERVAL)
}

function stopTimer() {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
}

watch(autoRefresh, enabled => (enabled ? startTimer() : stopTimer()))

onMounted(() => {
  fetchHealth()
  if (autoRefresh.value) startTimer()
})

onUnmounted(stopTimer)

// 组件展示顺序：核心组件在前，其余按 key 追加
const BASE_COMPONENTS = ['db', 'redis', 'publisher', 'subscriber']

function componentNames(node: NodeHealth): string[] {
  const names = BASE_COMPONENTS.filter(n => n in node.components)
  for (const key of Object.keys(node.components)) {
    if (!names.includes(key)) names.push(key)
  }
  return names.length > 0 ? names : BASE_COMPONENTS
}

const COMP_LABELS: Record<string, string> = {
  db: '数据库',
  redis: 'Redis',
  publisher: '事件发布',
  subscriber: '事件订阅',
}

function compLabel(name: string): string {
  return COMP_LABELS[name] ?? name
}

function compFailed(node: NodeHealth, name: string): boolean {
  const c = node.components[name]
  return !!c && c.status !== 'ok'
}

function hasFail(node: NodeHealth): boolean {
  return componentNames(node).some(n => compFailed(node, n))
}

function cardClass(node: NodeHealth) {
  return {
    offline: node.state === 'offline',
    degraded: node.state === 'online' && hasFail(node),
  }
}

function roleLabel(role: string): string {
  if (role === 'api') return 'API 节点'
  if (role === 'worker') return 'Worker 节点'
  return 'API + Worker'
}

const WORKER_LABELS: Record<string, string> = {
  standby: '待机',
  recovering: '恢复中',
  'leader-ready': '主节点就绪',
  stopping: '停止中',
}

function workerLabel(status: string): string {
  if (!status) return '未启用'
  return WORKER_LABELS[status] ?? status
}

function wsTotal(node: NodeHealth): number | string {
  if (!node.ws_connections) return '—'
  return Object.values(node.ws_connections).reduce((sum, n) => sum + n, 0)
}

function formatBytes(bytes: number): string {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  let value = bytes
  let unit = 0
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024
    unit++
  }
  return `${value.toFixed(unit === 0 ? 0 : 1)} ${units[unit]}`
}

function formatUptime(seconds: number): string {
  if (!seconds) return '—'
  const d = Math.floor(seconds / 86400)
  const h = Math.floor((seconds % 86400) / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  if (d > 0) return `${d}天${h}小时`
  if (h > 0) return `${h}小时${m}分`
  return `${m}分${seconds % 60}秒`
}

function formatTime(unix: number): string {
  if (!unix) return '—'
  return new Date(unix * 1000).toLocaleTimeString('zh-CN', { hour12: false })
}
</script>

<style scoped lang="scss">
@use '@/styles/variables' as *;

.mono {
  font-family: $font-family-mono;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;

  .header-actions {
    display: flex;
    align-items: center;
    gap: $spacing-sm;

    .auto-refresh-label {
      font-size: $font-size-sm;
      color: $text-tertiary;
    }
  }
}

.health-content {
  min-height: 240px;
}

/* ---------- 集群总览横幅 ---------- */
.status-hero {
  display: flex;
  align-items: center;
  gap: $spacing-xl;
  padding: $spacing-xl $spacing-2xl;
  margin-bottom: $spacing-xl;
  border-radius: $radius-lg;
  border: 1px solid rgba($color-success, 0.22);
  background:
    radial-gradient(circle at 1px 1px, rgba($text-primary, 0.04) 1px, transparent 0) 0 0 / 18px 18px,
    linear-gradient(135deg, rgba($color-success, 0.09), rgba($color-accent, 0.05) 55%, transparent 92%),
    $bg-secondary;
  box-shadow: $shadow-sm;
  animation: heroIn 0.45s ease both;

  &.degraded {
    border-color: rgba($color-warning, 0.35);
    background:
      radial-gradient(circle at 1px 1px, rgba($text-primary, 0.04) 1px, transparent 0) 0 0 / 18px 18px,
      linear-gradient(135deg, rgba($color-warning, 0.13), rgba($color-warning, 0.04) 55%, transparent 92%),
      $bg-secondary;
  }
}

.hero-orb {
  position: relative;
  flex-shrink: 0;
  width: 48px;
  height: 48px;

  &::before,
  &::after {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: 50%;
    border: 2px solid rgba($color-success, 0.4);
    animation: ripple 2.8s cubic-bezier(0.2, 0.6, 0.4, 1) infinite;
  }

  &::after {
    animation-delay: 1.4s;
  }

  .orb-core {
    position: absolute;
    inset: 13px;
    border-radius: 50%;
    background: $color-success;
    box-shadow:
      0 0 14px rgba($color-success, 0.55),
      inset 0 1px 3px rgba(255, 255, 255, 0.4);
  }
}

.status-hero.degraded .hero-orb {
  &::before,
  &::after {
    border-color: rgba($color-warning, 0.4);
    animation-duration: 3.8s;
  }

  .orb-core {
    background: $color-warning;
    box-shadow:
      0 0 14px rgba($color-warning, 0.55),
      inset 0 1px 3px rgba(255, 255, 255, 0.4);
  }
}

.hero-main {
  flex: 1;
  min-width: 0;

  .hero-title {
    font-size: $font-size-xl;
    font-weight: $font-weight-semibold;
    color: $text-primary;
    line-height: 1.3;
  }

  .hero-sub {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: $spacing-xs $spacing-md;
    margin-top: $spacing-xs;
    font-size: $font-size-sm;
    color: $text-secondary;
  }

  .hero-count {
    font-weight: $font-weight-medium;
  }

  .hero-warning {
    color: $color-warning;
  }
}

.hero-chips {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: $spacing-sm;
  max-width: 40%;

  .cluster-chip {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 4px 12px;
    border-radius: 999px;
    font-family: $font-family-mono;
    font-size: $font-size-xs;
    color: $text-secondary;
    background: $bg-secondary;
    border: 1px solid rgba($color-success, 0.35);
    box-shadow: $shadow-sm;

    .chip-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%;
      background: $color-success;
      box-shadow: 0 0 5px rgba($color-success, 0.7);
    }

    &.offline {
      border-color: rgba($color-error, 0.35);

      .chip-dot {
        background: $color-error;
        box-shadow: 0 0 5px rgba($color-error, 0.7);
      }
    }
  }
}

/* ---------- 节点卡片 ---------- */
.node-card {
  position: relative;
  margin-bottom: $spacing-lg;
  background: $bg-secondary;
  border: 1px solid $border-color;
  border-radius: $radius-lg;
  box-shadow: $shadow-sm;
  overflow: hidden;
  transition: transform 0.25s ease, box-shadow 0.25s ease, border-color 0.25s ease;
  animation: cardIn 0.45s ease both;
  animation-delay: calc(var(--i, 0) * 70ms);

  &:hover {
    transform: translateY(-3px);
    box-shadow: $shadow-lg;
    border-color: rgba($color-accent, 0.35);
  }

  .card-strip {
    height: 3px;
    background: linear-gradient(90deg, rgba($color-success, 0.9), rgba($color-success, 0.25) 55%, transparent);
  }

  &.degraded .card-strip {
    background: linear-gradient(90deg, rgba($color-warning, 0.95), rgba($color-warning, 0.25) 55%, transparent);
  }

  &.offline {
    border-color: rgba($color-error, 0.38);

    .card-strip {
      background: linear-gradient(90deg, rgba($color-error, 0.9), rgba($color-error, 0.25) 55%, transparent);
    }

    // 离线节点的快照数据已过期，弱化展示
    .components,
    .metrics {
      opacity: 0.55;
    }
  }
}

.node-header {
  padding: $spacing-lg $spacing-lg 0;

  .node-title {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: $spacing-sm;
  }

  .node-name {
    font-family: $font-family-mono;
    font-size: $font-size-md;
    font-weight: $font-weight-semibold;
    color: $text-primary;
    letter-spacing: 0.3px;
  }

  .role-chip {
    padding: 2px 8px;
    border-radius: 4px;
    font-size: $font-size-xs;
    color: $text-secondary;
    background: $bg-deep;
    border: 1px solid $border-color;
  }

  .state-chip {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 2px 9px;
    border-radius: 999px;
    font-size: $font-size-xs;
    font-weight: $font-weight-medium;

    &.online {
      color: $color-success;
      background: $color-success-bg;
      border: 1px solid rgba($color-success, 0.3);
    }

    &.offline {
      color: #fff;
      background: $color-error;
    }

    &.warn {
      color: $color-warning;
      background: $color-warning-bg;
      border: 1px solid rgba($color-warning, 0.35);
    }

    .live-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%;
      background: currentColor;
      animation: breathe 2.2s ease-in-out infinite;
    }
  }

  .node-sub {
    display: flex;
    flex-wrap: wrap;
    gap: $spacing-xs $spacing-lg;
    margin-top: $spacing-sm;
    font-size: $font-size-xs;
    color: $text-tertiary;

    .mono {
      color: $text-secondary;
    }
  }
}

/* ---------- 组件状态 ---------- */
.components {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: $spacing-sm;
  padding: $spacing-lg;

  .component-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    border-radius: $radius-base;
    background: $bg-tertiary;
    border: 1px solid transparent;
    font-size: $font-size-sm;
    transition: border-color 0.2s ease, background-color 0.2s ease;

    &.fail {
      border-color: rgba($color-error, 0.35);
      background: $color-error-bg;
    }
  }

  .comp-label {
    color: $text-secondary;
  }

  .comp-status {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    font-weight: $font-weight-medium;

    &.ok {
      color: $color-success;
    }

    &.fail {
      color: $color-error;

      .comp-dot {
        animation: breathe 1.4s ease-in-out infinite;
      }
    }

    &.none {
      color: $text-disabled;
      font-weight: $font-weight-regular;
    }

    .comp-latency {
      font-family: $font-family-mono;
      font-size: $font-size-xs;
      font-weight: $font-weight-regular;
      color: $text-tertiary;
    }
  }

  .comp-dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: currentColor;
  }
}

/* ---------- 运行指标 ---------- */
.metrics {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: $spacing-md $spacing-sm;
  margin: 0 $spacing-lg $spacing-lg;
  padding-top: $spacing-base;
  border-top: 1px dashed $border-color;

  .metric {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;

    .metric-value {
      font-size: $font-size-md;
      font-weight: $font-weight-semibold;
      color: $text-primary;
    }

    .metric-label {
      font-size: $font-size-xs;
      color: $text-tertiary;
    }
  }
}

/* ---------- 底部状态行 ---------- */
.updated-at {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: $spacing-sm;
  margin-top: $spacing-sm;
  font-size: $font-size-xs;
  color: $text-tertiary;

  .live-flag {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    padding: 1px 9px;
    border-radius: 999px;
    font-size: 10px;
    font-weight: $font-weight-semibold;
    letter-spacing: 1px;

    &.on {
      color: $color-success;
      background: $color-success-bg;

      .flag-dot {
        animation: breathe 1.6s ease-in-out infinite;
      }
    }

    &:not(.on) {
      color: $text-disabled;
      background: $bg-deep;
    }

    .flag-dot {
      width: 5px;
      height: 5px;
      border-radius: 50%;
      background: currentColor;
    }
  }
}

/* ---------- 动画 ---------- */
@keyframes ripple {
  0% {
    transform: scale(0.5);
    opacity: 0.8;
  }

  100% {
    transform: scale(1.35);
    opacity: 0;
  }
}

@keyframes breathe {
  0%,
  100% {
    opacity: 1;
  }

  50% {
    opacity: 0.35;
  }
}

@keyframes cardIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }

  to {
    opacity: 1;
    transform: none;
  }
}

@keyframes heroIn {
  from {
    opacity: 0;
    transform: translateY(-6px);
  }

  to {
    opacity: 1;
    transform: none;
  }
}

@media (max-width: 768px) {
  .status-hero {
    flex-direction: column;
    align-items: flex-start;
    gap: $spacing-base;
  }

  .hero-chips {
    max-width: 100%;
    justify-content: flex-start;
  }
}

@media (prefers-reduced-motion: reduce) {
  .status-hero,
  .node-card,
  .hero-orb,
  .live-dot,
  .comp-dot,
  .flag-dot {
    animation: none !important;
  }
}
</style>
