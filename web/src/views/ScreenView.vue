<template>
  <div class="screen-wrapper">
    <!-- 科技背景层 -->
    <div class="particle-grid"></div>
    <div class="scan-line"></div>
    <div class="ambient-glow"></div>

    <!-- 加载状态 -->
    <div v-if="loading" class="overlay-state">
      <div class="loader">
        <div class="loader-ring"></div>
        <p class="loader-text">正在连接演练数据...</p>
      </div>
    </div>

    <!-- 错误状态 -->
    <div v-else-if="error" class="overlay-state error">
      <div class="error-content">
        <el-icon :size="48"><CircleClose /></el-icon>
        <p>{{ error }}</p>
        <el-button type="primary" @click="handleRetry">重试</el-button>
      </div>
    </div>

    <!-- 主内容 -->
    <div v-else-if="currentDrill" class="screen-content">
      <!-- 顶部标题栏 -->
      <header class="screen-header cyber-border-bottom">
        <div class="header-left">
          <div class="title-glow"></div>
          <h1 class="drill-title">{{ currentDrill.name }}</h1>
          <span class="drill-status-tag" :class="'status-' + (currentDrill.status || '')">
            {{ getStatusLabel(currentDrill.status || '') }}
          </span>
        </div>
        <div class="header-right">
          <div class="time-display">
            <Clock :size="14" />
            <span class="time-value">{{ currentTime }}</span>
          </div>
          <button class="cyber-btn" @click="toggleFullscreen" title="全屏切换">
            <FullScreen :size="16" />
          </button>
        </div>
      </header>

      <!-- 核心指标栏 -->
      <section class="metrics-bar">
        <!-- 核心指标组 -->
        <div class="metric-group primary-metrics">
          <div class="metric-card core-metric running">
            <div class="metric-corners"></div>
            <div class="metric-icon-box">
              <Timer :size="14" />
            </div>
            <div class="metric-body">
              <span class="metric-label">执行中</span>
              <span class="metric-value" :data-value="drillSteps.filter(s => s.status === 'running').length">
                {{ drillSteps.filter(s => s.status === 'running').length }}
              </span>
            </div>
          </div>
          <div class="metric-card core-metric completed">
            <div class="metric-corners"></div>
            <div class="metric-icon-box">
              <CircleCheck :size="14" />
            </div>
            <div class="metric-body">
              <span class="metric-label">已完成</span>
              <span class="metric-value">
                {{ drillSteps.filter(s => s.status === 'completed').length }}<span class="metric-total">/{{ drillSteps.length }}</span>
              </span>
            </div>
          </div>
          <div class="metric-card core-metric rate">
            <div class="metric-corners"></div>
            <div class="metric-icon-box">
              <DataLine :size="14" />
            </div>
            <div class="metric-body">
              <span class="metric-label">成功率</span>
              <span class="metric-value" :class="successRate === 100 ? 'success' : successRate > 0 ? 'warning' : ''">{{ successRate }}%</span>
            </div>
          </div>
        </div>

        <!-- 分隔线 -->
        <div class="metric-divider"></div>

        <!-- 次要指标组 -->
        <div class="metric-group secondary-metrics">
          <div class="metric-card mini-metric">
            <div class="metric-icon-box">
              <User :size="12" />
            </div>
            <div class="metric-body">
              <span class="metric-label">创建人</span>
              <span class="metric-value">{{ currentDrill.created_by_name || '-' }}</span>
            </div>
          </div>
          <div class="metric-card mini-metric">
            <div class="metric-icon-box">
              <Calendar :size="12" />
            </div>
            <div class="metric-body">
              <span class="metric-label">开始</span>
              <span class="metric-value">{{ currentDrill.start_time ? formatTime(currentDrill.start_time) : '-' }}</span>
            </div>
          </div>
        </div>
      </section>

      <!-- 主体内容区 -->
      <main class="screen-body">
        <!-- 左侧：步骤执行面板 -->
        <aside class="left-panel cyber-panel">
          <div class="panel-header">
            <List :size="12" />
            <span>步骤执行面板</span>
          </div>
          <div class="steps-list">
            <div
              v-for="step in drillSteps"
              :key="step.id"
              class="step-item"
              :class="['step-' + step.status, { 'active-step': step.status === 'running' }]"
            >
              <div class="step-indicator">
                <component :is="getStepStatusIcon(step.status)" :size="14" />
              </div>
              <div class="step-body">
                <div class="step-name">{{ step.name }}</div>
                <div class="step-meta">
                  <span v-if="step.executor_team" class="step-team">{{ step.executor_team }}</span>
                  <span v-if="step.start_time" class="step-duration">{{ calculateDuration(step) }}</span>
                </div>
              </div>
              <span class="step-badge" :class="'badge-' + step.status">
                {{ getStepStatusLabel(step.status) }}
              </span>
            </div>
          </div>
        </aside>

        <!-- 中间：进度仪表盘 -->
        <section class="center-panel">
          <div class="gauge-container cyber-card">
            <div class="panel-header">
              <DataLine :size="12" />
              <span>执行进度</span>
            </div>
            <div class="gauge-chart">
              <svg viewBox="0 0 200 200" class="gauge-svg">
                <defs>
                  <linearGradient id="gaugeGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                    <stop offset="0%" stop-color="#00d4ff" />
                    <stop offset="100%" stop-color="#00e88a" />
                  </linearGradient>
                  <filter id="glow">
                    <feGaussianBlur stdDeviation="2.5" result="coloredBlur"/>
                    <feMerge>
                      <feMergeNode in="coloredBlur"/>
                      <feMergeNode in="SourceGraphic"/>
                    </feMerge>
                  </filter>
                </defs>
                <circle cx="100" cy="100" r="80" fill="none" stroke="rgba(255,255,255,0.04)" stroke-width="8" />
                <circle
                  cx="100" cy="100" r="80"
                  fill="none"
                  stroke="url(#gaugeGradient)"
                  stroke-width="8"
                  stroke-linecap="round"
                  :stroke-dasharray="503"
                  :stroke-dashoffset="503 - (503 * progressPercent / 100)"
                  transform="rotate(-90 100 100)"
                  filter="url(#glow)"
                  class="gauge-progress"
                />
                <text x="100" y="95" text-anchor="middle" fill="#f8fafc" font-size="32" font-weight="700" font-family="'SF Mono', 'Fira Code', monospace">
                  {{ progressPercent }}%
                </text>
                <text x="100" y="115" text-anchor="middle" fill="rgba(148, 163, 184, 0.5)" font-size="9" font-family="sans-serif" letter-spacing="2">
                  PROGRESS
                </text>
              </svg>
            </div>
          </div>

          <div class="status-bars cyber-card">
            <div class="panel-header">
              <PieChart :size="14" />
              <span>状态分布</span>
            </div>
            <div class="bars-container">
              <div v-for="(item, idx) in stepStatusData" :key="item.name" class="status-bar-row">
                <span class="bar-label">{{ item.name }}</span>
                <div class="bar-track">
                  <div
                    class="bar-fill"
                    :class="'bar-' + getBarColorClass(item.name)"
                    :style="{ width: Math.max((item.value / drillSteps.length) * 100, 3) + '%' }"
                  />
                </div>
                <span class="bar-value">{{ item.value }}</span>
              </div>
            </div>
          </div>
        </section>

        <!-- 右侧：时间线与日志 -->
        <aside class="right-panel cyber-panel">
          <div class="panel-header">
            <Connection :size="14" />
            <span>执行日志</span>
          </div>
          <div class="logs-list">
            <div
              v-for="(log, idx) in drillLogs"
              :key="log.ID || idx"
              class="log-item"
              :class="'log-' + (log.action || '')"
            >
              <div class="log-dot" />
              <div class="log-content">
                <div class="log-action">{{ formatLogAction(log.action) }}</div>
                <div class="log-detail" v-if="log.content">{{ log.content }}</div>
                <div class="log-time">{{ formatTime(log.CreatedAt) }}</div>
              </div>
            </div>
          </div>
        </aside>
      </main>

      <!-- 底部进度条 -->
      <footer class="bottom-bar cyber-border-top">
        <div class="bottom-progress-track">
          <div class="bottom-progress-fill" :style="{ width: progressPercent + '%' }" />
        </div>
      </footer>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute } from 'vue-router'
import {
  Clock,
  FullScreen,
  User,
  Calendar,
  List,
  DataLine,
  PieChart,
  CircleCheck,
  CircleClose,
  Loading,
  Connection,
  Timer,
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import type { StepInstance } from '@/types'
import { drillApi } from '@/api/modules/drill'

const route = useRoute()
const loading = ref(true)
const error = ref<string | null>(null)
const currentTime = ref(new Date().toLocaleString('zh-CN'))
let ws: WebSocket | null = null
let refreshTimer: number | null = null

const drillId = computed(() => {
  const id = route.params.id
  return typeof id === 'string' ? parseInt(id, 10) : null
})

const currentDrill = ref<any>(null)
const drillSteps = ref<StepInstance[]>([])

const progressPercent = computed(() => {
  if (drillSteps.value.length === 0) return 0
  const completed = drillSteps.value.filter(s => s.status === 'completed').length
  return Math.round((completed / drillSteps.value.length) * 100)
})

function formatLogAction(action: string): string {
  const map: Record<string, string> = {
    complete: '步骤完成',
    issue: '异常上报',
    start: '步骤启动',
    timeout: '步骤超时',
    skip: '步骤跳过',
    force_complete: '强制完成',
  }
  return map[action] || action
}

function getBarColorClass(name: string): string {
  const map: Record<string, string> = {
    '已完成': 'completed',
    '执行中': 'running',
    '待执行': 'pending',
    '超时': 'timeout',
    '异常': 'issue',
    '已跳过': 'skipped',
  }
  return map[name] || 'pending'
}

const stepStatusData = computed(() => {
  const counts: Record<string, number> = {}
  drillSteps.value.forEach(step => {
    counts[step.status] = (counts[step.status] || 0) + 1
  })
  const statusMap: Record<string, string> = {
    pending: '待执行',
    running: '执行中',
    completed: '已完成',
    timeout: '超时',
    issue: '异常',
    skipped: '已跳过',
  }
  return Object.entries(counts).map(([key, value]) => ({
    name: statusMap[key] || key,
    value,
  }))
})

const drillLogs = ref<any[]>([])

let timeTimer: number

function updateTime() {
  currentTime.value = new Date().toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  })
}

async function loadData() {
  if (!drillId.value) {
    error.value = '无效的演练 ID'
    loading.value = false
    return
  }

  try {
    const drill = await drillApi.getDetail(drillId.value)
    currentDrill.value = drill

    const steps = await drillApi.getSteps(drillId.value)
    drillSteps.value = steps.sort((a: StepInstance, b: StepInstance) => a.seq - b.seq)

    const logs = await drillApi.getLogs(drillId.value)
    drillLogs.value = logs.slice(0, 30)

    loading.value = false
    error.value = null

    connectWebSocket()
  } catch (err: any) {
    error.value = err.message || '加载数据失败'
    console.error('Failed to load drill data:', err)
    loading.value = false
  }
}

function getStatusLabel(status: string): string {
  const map: Record<string, string> = {
    running: '执行中',
    paused: '已暂停',
    completed: '已完成',
    terminated: '已终止',
    pending: '待启动',
  }
  return map[status] || status
}

function getStepStatusLabel(status: string): string {
  const map: Record<string, string> = {
    pending: '待执行',
    running: '执行中',
    completed: '已完成',
    timeout: '超时',
    skipped: '已跳过',
    issue: '异常',
  }
  return map[status] || status
}

function getStepStatusIcon(status: string) {
  switch (status) {
    case 'completed': return CircleCheck
    case 'running': return Loading
    case 'issue': return CircleClose
    case 'timeout': return CircleClose
    default: return CircleCheck
  }
}

function calculateDuration(step: StepInstance): string {
  if (!step.start_time) return '-'
  const start = new Date(step.start_time).getTime()
  const end = step.end_time ? new Date(step.end_time).getTime() : Date.now()
  const diff = Math.floor((end - start) / 1000)
  const mins = Math.floor(diff / 60)
  const secs = diff % 60
  return mins > 0 ? `${mins}分${secs}秒` : `${secs}秒`
}

function formatTime(dateStr: string | null | undefined): string {
  if (!dateStr) return '--:--'
  const date = new Date(dateStr)
  return date.toLocaleTimeString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  })
}

function connectWebSocket() {
  if (ws) {
    ws.close()
  }
  const wsProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
  const wsUrl = `${wsProtocol}://${window.location.host}/ws/control/${drillId.value}`
  
  ws = new WebSocket(wsUrl)
  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      if (data.type === 'step_change' || data.type === 'drill_status') {
        loadData()
      }
    } catch (e) {
      console.warn('WebSocket message parse error:', e)
    }
  }
  ws.onerror = () => {
    console.warn('WebSocket error, falling back to polling')
    startFallbackPolling()
  }
  ws.onclose = () => {
    if (currentDrill.value?.status === 'running') {
      startFallbackPolling()
    }
  }
}

function startFallbackPolling() {
  if (refreshTimer) clearInterval(refreshTimer)
  refreshTimer = window.setInterval(() => {
    if (currentDrill.value?.status === 'running') {
      loadData()
    }
  }, 3000)
}

function toggleFullscreen() {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen?.()
  } else {
    document.exitFullscreen?.()
  }
}

function handleRetry() {
  loadData()
}

onMounted(() => {
  loadData()
  updateTime()
  timeTimer = window.setInterval(updateTime, 1000)
})

onBeforeUnmount(() => {
  clearInterval(timeTimer)
  if (refreshTimer) clearInterval(refreshTimer)
  if (ws) {
    ws.close()
    ws = null
  }
})
</script>

<style lang="scss" scoped>
// ===== 核心变量 =====
$bg-deep: #020617;
$bg-surface: rgba(8, 18, 35, 0.7);
$bg-glass: rgba(15, 25, 45, 0.5);
$bg-card: rgba(10, 20, 40, 0.8);

$border-subtle: rgba(0, 180, 255, 0.1);
$border-active: rgba(0, 212, 255, 0.35);

$text-primary: #f8fafc;
$text-secondary: rgba(148, 163, 184, 0.8);
$text-muted: rgba(100, 120, 150, 0.6);

$cyan: #00d4ff;
$cyan-dim: rgba(0, 212, 255, 0.15);
$green: #00e88a;
$green-dim: rgba(0, 232, 138, 0.15);
$red: #ff3b5c;
$red-dim: rgba(255, 59, 92, 0.15);
$yellow: #ffaa00;
$yellow-dim: rgba(255, 170, 0, 0.15);
$purple: #a855f7;
$purple-dim: rgba(168, 85, 247, 0.15);

.screen-wrapper {
  width: 100vw;
  height: 100vh;
  background: $bg-deep;
  color: $text-primary;
  overflow: hidden;
  position: relative;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
}

// ===== 背景层 =====
.particle-grid {
  position: absolute;
  inset: 0;
  background-image:
    radial-gradient(circle at 1px 1px, rgba(0, 180, 255, 0.08) 1px, transparent 0);
  background-size: 40px 40px;
  z-index: 0;
  animation: gridFloat 60s linear infinite;
}

@keyframes gridFloat {
  0% { transform: translateY(0); }
  100% { transform: translateY(40px); }
}

.scan-line {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, transparent, rgba(0, 212, 255, 0.15), transparent);
  animation: scanMove 8s ease-in-out infinite;
  z-index: 1;
  pointer-events: none;
}

@keyframes scanMove {
  0%, 100% { top: 0; opacity: 0; }
  5%, 95% { opacity: 1; }
  100% { top: 100vh; opacity: 0; }
}

.ambient-glow {
  position: absolute;
  top: -30%;
  left: 50%;
  transform: translateX(-50%);
  width: 80vw;
  height: 60vh;
  background: radial-gradient(ellipse, rgba(0, 100, 200, 0.06) 0%, transparent 70%);
  z-index: 0;
  pointer-events: none;
}

// ===== 通用状态 =====
.overlay-state {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  background: $bg-deep;

  &.error .error-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    color: $text-secondary;

    .el-icon { color: $red; }

    p { font-size: 14px; }

    .el-button {
      background: transparent;
      border: 1px solid $cyan;
      color: $cyan;
      &:hover { background: $cyan-dim; }
    }
  }
}

.loader {
  text-align: center;

  .loader-ring {
    width: 48px;
    height: 48px;
    margin: 0 auto 16px;
    border: 2px solid $border-subtle;
    border-top-color: $cyan;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  .loader-text {
    color: $text-secondary;
    font-size: 13px;
    letter-spacing: 3px;
    text-transform: uppercase;
  }
}

@keyframes spin { to { transform: rotate(360deg); } }

// ===== 主内容 =====
.screen-content {
  position: relative;
  z-index: 2;
  display: flex;
  flex-direction: column;
  height: 100vh;
  animation: fadeIn 0.6s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

// ===== Header =====
.screen-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background: rgba(5, 12, 25, 0.8);
  border-bottom: 1px solid $border-subtle;
  position: relative;

  &::after {
    content: '';
    position: absolute;
    bottom: -1px;
    left: 10%;
    right: 10%;
    height: 1px;
    background: linear-gradient(90deg, transparent, $border-active, transparent);
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 10px;

    .title-glow {
      width: 3px;
      height: 16px;
      background: linear-gradient(180deg, $cyan, transparent);
      border-radius: 2px;
      flex-shrink: 0;
    }

    .drill-title {
      font-size: 15px;
      font-weight: 700;
      letter-spacing: 1.5px;
      margin: 0;
      color: $text-primary;
    }

    .drill-status-tag {
      padding: 3px 10px;
      border-radius: 3px;
      font-size: 11px;
      font-weight: 600;
      letter-spacing: 1px;
      text-transform: uppercase;

      &.status-running {
        background: $green-dim;
        color: $green;
        border: 1px solid rgba(0, 232, 138, 0.3);
        animation: pulseGlow 3s ease-in-out infinite;
      }
      &.status-pending {
        background: $yellow-dim;
        color: $yellow;
        border: 1px solid rgba(255, 170, 0, 0.3);
      }
      &.status-completed {
        background: $green-dim;
        color: $green;
        border: 1px solid rgba(0, 232, 138, 0.4);
      }
      &.status-paused {
        background: $yellow-dim;
        color: $yellow;
        border: 1px solid rgba(255, 170, 0, 0.3);
      }
      &.status-terminated {
        background: $red-dim;
        color: $red;
        border: 1px solid rgba(255, 59, 92, 0.3);
      }
    }
  }

  @keyframes pulseGlow {
    0%, 100% { box-shadow: 0 0 4px rgba(0, 232, 138, 0.2); }
    50% { box-shadow: 0 0 12px rgba(0, 232, 138, 0.3); }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 16px;

    .time-display {
      display: flex;
      align-items: center;
      gap: 6px;
      color: $text-secondary;
      font-size: 13px;
      font-family: 'SF Mono', 'Fira Code', monospace;
    }

    .cyber-btn {
      background: transparent;
      border: 1px solid $border-subtle;
      color: $text-secondary;
      width: 32px;
      height: 32px;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: 4px;
      cursor: pointer;
      transition: all 0.2s;

      &:hover {
        border-color: $cyan;
        color: $cyan;
        box-shadow: 0 0 8px rgba(0, 212, 255, 0.15);
      }
    }
  }
}

// ===== Metrics =====
.metrics-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  background: rgba(5, 12, 25, 0.5);
  border-bottom: 1px solid $border-subtle;
}

.metric-group {
  display: flex;
  gap: 10px;

  &.primary-metrics { flex: 1; }
  &.secondary-metrics { gap: 8px; }
}

.metric-divider {
  width: 1px;
  height: 32px;
  background: linear-gradient(180deg, transparent, $border-active, transparent);
}

.metric-card {
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  background: $bg-card;
  border-radius: 6px;
  border: 1px solid $border-subtle;
  transition: all 0.3s;

  &:hover {
    border-color: $border-active;
  }

  .metric-corners {
    position: absolute;
    inset: 0;
    pointer-events: none;
    border-radius: 6px;

    &::before, &::after {
      content: '';
      position: absolute;
      width: 8px;
      height: 8px;
      border-color: $cyan;
      opacity: 0;
      transition: opacity 0.3s;
    }

    &::before {
      top: -1px;
      left: -1px;
      border-top: 2px solid;
      border-left: 2px solid;
    }

    &::after {
      bottom: -1px;
      right: -1px;
      border-bottom: 2px solid;
      border-right: 2px solid;
    }
  }

  &:hover .metric-corners::before,
  &:hover .metric-corners::after {
    opacity: 0.6;
  }

  &.core-metric {
    flex: 1;
    padding: 6px 8px;

    .metric-value {
      font-size: 16px;
    }

    &.running .metric-icon-box { background: $cyan-dim; color: $cyan; }
    &.completed .metric-icon-box { background: $green-dim; color: $green; }
    &.rate .metric-icon-box { background: $purple-dim; color: $purple; }
  }

  &.mini-metric {
    .metric-icon-box {
      width: 24px;
      height: 24px;
      border-radius: 4px;
      background: rgba(0, 180, 255, 0.06);
      color: $text-secondary;
    }

    .metric-label { font-size: 10px; }
    .metric-value { font-size: 12px; }
  }

  .metric-icon-box {
    width: 28px;
    height: 28px;
    border-radius: 5px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    background: rgba(0, 180, 255, 0.08);
    color: $cyan;
  }

  .metric-body {
    display: flex;
    flex-direction: column;
    min-width: 0;

    .metric-label {
      font-size: 10px;
      color: $text-muted;
      text-transform: uppercase;
      letter-spacing: 1px;
      line-height: 1;
      margin-bottom: 4px;
    }

    .metric-value {
      font-size: 14px;
      font-weight: 700;
      font-family: 'SF Mono', 'Fira Code', monospace;
      color: $text-primary;
      line-height: 1;

      .metric-total {
        font-weight: 400;
        color: $text-secondary;
        font-size: 0.75em;
      }

      &.success { color: $green; }
      &.warning { color: $yellow; }
    }
  }
}

// ===== Body =====
.screen-body {
  flex: 1;
  display: grid;
  grid-template-columns: 260px 1fr 280px;
  gap: 10px;
  padding: 10px;
  overflow: hidden;
}

.cyber-panel {
  background: $bg-surface;
  border-radius: 8px;
  border: 1px solid $border-subtle;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  backdrop-filter: blur(12px);
}

.panel-header {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  background: rgba(0, 180, 255, 0.03);
  border-bottom: 1px solid $border-subtle;
  font-size: 10px;
  font-weight: 600;
  color: $text-secondary;
  letter-spacing: 2px;
  text-transform: uppercase;
  flex-shrink: 0;

  svg { color: $cyan; }
}

// Steps
.steps-list {
  flex: 1;
  overflow-y: auto;
  padding: 6px;

  &::-webkit-scrollbar { width: 4px; }
  &::-webkit-scrollbar-track { background: transparent; }
  &::-webkit-scrollbar-thumb { background: $border-subtle; border-radius: 2px; }
}

.step-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 8px;
  margin-bottom: 3px;
  background: rgba(0, 0, 0, 0.25);
  border-radius: 4px;
  border: 1px solid transparent;
  transition: all 0.25s;

  &.active-step {
    border-color: rgba(0, 232, 138, 0.25);
    background: rgba(0, 232, 138, 0.04);

    .step-indicator svg {
      animation: spin 1s linear infinite;
      color: $green;
    }
  }

  .step-indicator {
    flex-shrink: 0;
    color: $text-muted;

    .step-completed & { color: $green; }
    .step-issue &, .step-timeout & { color: $red; }
  }

  .step-body {
    flex: 1;
    min-width: 0;

    .step-name {
      font-size: 12px;
      font-weight: 500;
      color: $text-primary;
      margin-bottom: 3px;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .step-meta {
      display: flex;
      gap: 8px;
      font-size: 10px;
      color: $text-muted;

      .step-team {
        background: $cyan-dim;
        padding: 1px 5px;
        border-radius: 2px;
        color: $cyan;
      }
    }
  }

  .step-badge {
    padding: 2px 6px;
    border-radius: 3px;
    font-size: 9px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    flex-shrink: 0;

    &.badge-completed { background: $green-dim; color: $green; }
    &.badge-running { background: $cyan-dim; color: $cyan; }
    &.badge-pending { background: rgba(255, 255, 255, 0.03); color: $text-muted; }
    &.badge-issue, &.badge-timeout { background: $red-dim; color: $red; }
    &.badge-skipped { background: $purple-dim; color: $purple; }
  }
}

// Center
.center-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
  overflow: hidden;
}

.cyber-card {
  background: $bg-surface;
  border-radius: 8px;
  border: 1px solid $border-subtle;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  backdrop-filter: blur(12px);

  &.gauge-container { flex: 1.3; }
  &.status-bars { flex: 0.7; }
}

.gauge-chart {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 8px;
}

.gauge-svg {
  width: 100%;
  max-width: 200px;
  height: auto;
}

.gauge-progress {
  transition: stroke-dashoffset 0.8s cubic-bezier(0.4, 0, 0.2, 1);
}

.bars-container {
  flex: 1;
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.status-bar-row {
  display: flex;
  align-items: center;
  gap: 8px;

  .bar-label {
    font-size: 11px;
    color: $text-secondary;
    width: 42px;
    flex-shrink: 0;
  }

  .bar-track {
    flex: 1;
    height: 6px;
    background: rgba(255, 255, 255, 0.04);
    border-radius: 3px;
    overflow: hidden;
  }

  .bar-fill {
    height: 100%;
    border-radius: 3px;
    transition: width 0.6s cubic-bezier(0.4, 0, 0.2, 1);

    &.bar-completed { background: linear-gradient(90deg, $green, rgba(0, 232, 138, 0.5)); box-shadow: 0 0 6px rgba(0, 232, 138, 0.3); }
    &.bar-running { background: linear-gradient(90deg, $cyan, rgba(0, 212, 255, 0.5)); box-shadow: 0 0 6px rgba(0, 212, 255, 0.3); }
    &.bar-pending { background: rgba(255, 255, 255, 0.1); }
    &.bar-issue, &.bar-timeout { background: linear-gradient(90deg, $red, rgba(255, 59, 92, 0.5)); box-shadow: 0 0 6px rgba(255, 59, 92, 0.3); }
    &.bar-skipped { background: linear-gradient(90deg, $purple, rgba(168, 85, 247, 0.5)); }
  }

  .bar-value {
    font-size: 12px;
    font-weight: 700;
    font-family: 'SF Mono', monospace;
    color: $text-primary;
    width: 16px;
    text-align: right;
    flex-shrink: 0;
  }
}

// Right: Logs
.logs-list {
  flex: 1;
  overflow-y: auto;
  padding: 10px;

  &::-webkit-scrollbar { width: 4px; }
  &::-webkit-scrollbar-track { background: transparent; }
  &::-webkit-scrollbar-thumb { background: $border-subtle; border-radius: 2px; }
}

.log-item {
  position: relative;
  padding: 0 0 10px 16px;

  &:last-child { padding-bottom: 0; }

  &::before {
    content: '';
    position: absolute;
    left: 4px;
    top: 8px;
    bottom: 0;
    width: 1px;
    background: $border-subtle;
  }

  &:last-child::before { display: none; }

  .log-dot {
    position: absolute;
    left: 1px;
    top: 4px;
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: $text-muted;
    box-shadow: 0 0 4px rgba(0, 212, 255, 0.2);
  }

  &.log-complete .log-dot { background: $green; box-shadow: 0 0 6px rgba(0, 232, 138, 0.4); }
  &.log-issue .log-dot { background: $red; box-shadow: 0 0 6px rgba(255, 59, 92, 0.4); animation: pulseDot 1.5s ease-in-out infinite; }
  &.log-start .log-dot { background: $cyan; }

  @keyframes pulseDot {
    0%, 100% { box-shadow: 0 0 4px rgba(255, 59, 92, 0.3); }
    50% { box-shadow: 0 0 10px rgba(255, 59, 92, 0.5); }
  }

  .log-content {
    .log-action {
      font-size: 12px;
      font-weight: 500;
      color: $text-primary;
      margin-bottom: 2px;
    }

    .log-detail {
      font-size: 11px;
      color: $text-secondary;
      margin-bottom: 2px;
    }

    .log-time {
      font-size: 10px;
      font-family: 'SF Mono', monospace;
      color: $text-muted;
    }
  }
}

// Bottom bar
.bottom-bar {
  position: relative;
  padding: 0;
  background: rgba(5, 12, 25, 0.8);
  border-top: 1px solid $border-subtle;
  flex-shrink: 0;

  &::before {
    content: '';
    position: absolute;
    top: -1px;
    left: 15%;
    right: 15%;
    height: 1px;
    background: linear-gradient(90deg, transparent, $border-active, transparent);
  }
}

.bottom-progress-track {
  height: 2px;
  background: rgba(255, 255, 255, 0.03);
  overflow: hidden;
}

.bottom-progress-fill {
  height: 100%;
  background: linear-gradient(90deg, $cyan, $green);
  box-shadow: 0 0 8px rgba(0, 212, 255, 0.3);
  transition: width 0.6s cubic-bezier(0.4, 0, 0.2, 1);
  animation: progressGlow 3s ease-in-out infinite;
}

@keyframes progressGlow {
  0%, 100% { box-shadow: 0 0 6px rgba(0, 212, 255, 0.2); }
  50% { box-shadow: 0 0 14px rgba(0, 212, 255, 0.4); }
}
</style>
