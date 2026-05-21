<template>
  <div class="screen-root">
    <!-- Loading state -->
    <div v-if="loading" class="overlay-state">
      <div class="loader">
        <div class="loader-ring" />
        <p class="loader-text">正在连接演练数据...</p>
      </div>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="overlay-state error">
      <div class="error-content">
        <el-icon :size="48"><CircleClose /></el-icon>
        <p>{{ error }}</p>
        <el-button type="primary" @click="handleRetry">重试</el-button>
      </div>
    </div>

    <!-- Main content -->
    <div v-else-if="currentDrill" class="screen-content">
      <!-- Header -->
      <header class="screen-header">
        <div class="header-left">
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
          <button class="btn-icon" @click="toggleFullscreen" title="全屏切换">
            <FullScreen :size="16" />
          </button>
        </div>
      </header>

      <!-- KPI Row -->
      <section class="kpi-row">
        <div class="kpi-card">
          <span class="kpi-label">进度</span>
          <span class="kpi-value kpi-value-accent">{{ progressPercent }}%</span>
        </div>
        <div class="kpi-card">
          <span class="kpi-label">已完成</span>
          <span class="kpi-value">{{ completedCount }}<span class="kpi-total">/{{ totalCount }}</span></span>
        </div>
        <div class="kpi-card">
          <span class="kpi-label">异常</span>
          <span class="kpi-value" :class="{ 'kpi-value-error': issueCount > 0 }">{{ issueCount }}</span>
        </div>
        <div class="kpi-card">
          <span class="kpi-label">执行中</span>
          <span class="kpi-value kpi-value-accent">{{ runningCount }}</span>
        </div>
        <div class="kpi-card">
          <span class="kpi-label">预计剩余</span>
          <span class="kpi-value">{{ estimatedRemaining }}</span>
        </div>
      </section>

      <!-- Three-column main -->
      <main class="screen-main">
        <!-- Left: Step timeline -->
        <aside class="panel panel-timeline">
          <div class="panel-header">
            <List :size="12" />
            <span>步骤执行</span>
          </div>
          <div class="panel-body steps-list">
            <div
              v-for="step in drillSteps"
              :key="step.id"
              class="step-item"
              :class="'step-' + step.status"
            >
              <span class="step-dot" :class="'dot-' + step.status" />
              <div class="step-info">
                <span class="step-name">{{ step.name }}</span>
                <span class="step-meta">
                  <span v-if="step.executor_team" class="step-team">{{ step.executor_team }}</span>
                  <span v-if="step.start_time" class="step-duration">{{ calculateDuration(step) }}</span>
                </span>
              </div>
              <span class="step-badge">{{ getStepStatusLabel(step.status) }}</span>
            </div>
          </div>
        </aside>

        <!-- Center: ECharts grid -->
        <section class="panel panel-charts">
          <div class="chart-grid">
            <div class="chart-cell">
              <div ref="bulletChartRef" class="chart-container" />
            </div>
            <div class="chart-cell">
              <div ref="lineChartRef" class="chart-container" />
            </div>
            <div class="chart-cell">
              <div ref="pieChartRef" class="chart-container" />
            </div>
            <div class="chart-cell">
              <div ref="radarChartRef" class="chart-container" />
            </div>
          </div>
        </section>

        <!-- Right: Event logs -->
        <aside class="panel panel-logs">
          <div class="panel-header">
            <Connection :size="12" />
            <span>事件日志</span>
          </div>
          <div class="panel-body logs-list">
            <div
              v-for="(log, idx) in recentLogs"
              :key="log.id || idx"
              class="log-item"
              :class="'log-' + (log.action || '').split('_')[0]"
            >
              <div class="log-border" />
              <div class="log-body">
                <div class="log-action">{{ formatLogAction(log.action) }}</div>
                <div v-if="log.content" class="log-text">{{ log.content }}</div>
                <div class="log-time">{{ formatTime(log.created_at) }}</div>
              </div>
            </div>
            <div v-if="recentLogs.length === 0" class="log-empty">暂无日志</div>
          </div>
        </aside>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import {
  Clock,
  FullScreen,
  CircleClose,
  List,
  Connection,
} from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import type { StepInstance, StepInstanceLog } from '@/types/instance'
import type { DrillInstance } from '@/types/instance'
import { drillApi } from '@/api/modules/drill'

const route = useRoute()
const loading = ref(true)
const error = ref<string | null>(null)
const currentTime = ref(new Date().toLocaleString('zh-CN'))

let ws: WebSocket | null = null
let refreshTimer: number | null = null
let timeTimer: number | null = null

// Chart refs
const bulletChartRef = ref<HTMLElement | null>(null)
const lineChartRef = ref<HTMLElement | null>(null)
const pieChartRef = ref<HTMLElement | null>(null)
const radarChartRef = ref<HTMLElement | null>(null)

let bulletChart: echarts.ECharts | null = null
let lineChart: echarts.ECharts | null = null
let pieChart: echarts.ECharts | null = null
let radarChart: echarts.ECharts | null = null

const drillId = computed(() => {
  const id = route.params.id
  return typeof id === 'string' ? parseInt(id, 10) : null
})

const currentDrill = ref<DrillInstance | null>(null)
const drillSteps = ref<StepInstance[]>([])
const recentLogs = ref<StepInstanceLog[]>([])

// Computed KPIs
const completedCount = computed(() =>
  drillSteps.value.filter(s => s.status === 'completed' || s.status === 'skipped').length
)
const totalCount = computed(() => drillSteps.value.length)
const issueCount = computed(() => drillSteps.value.filter(s => s.status === 'issue').length)
const runningCount = computed(() => drillSteps.value.filter(s => s.status === 'running').length)
const progressPercent = computed(() => {
  if (totalCount.value === 0) return 0
  return Math.round((completedCount.value / totalCount.value) * 100)
})

const estimatedRemaining = computed(() => {
  const remaining = drillSteps.value.filter(s => s.status === 'pending')
  if (remaining.length === 0) return '0分钟'
  // Use estimated_duration_minutes if available, otherwise assume 5 min avg
  let totalMin = 0
  let counted = 0
  remaining.forEach(s => {
    if (s.estimated_duration_minutes && s.estimated_duration_minutes > 0) {
      totalMin += s.estimated_duration_minutes
      counted++
    }
  })
  if (counted === 0) {
    // Calculate from completed steps average
    const completed = drillSteps.value.filter(s => s.status === 'completed' && s.start_time && s.end_time)
    if (completed.length > 0) {
      let sum = 0
      completed.forEach(s => {
        const dur = (new Date(s.end_time!).getTime() - new Date(s.start_time!).getTime()) / 60000
        sum += dur
      })
      totalMin = Math.round(sum / completed.length * remaining.length)
    } else {
      totalMin = remaining.length * 5
    }
  } else {
    totalMin = Math.round(totalMin)
  }
  if (totalMin < 60) return `${totalMin}分钟`
  const h = Math.floor(totalMin / 60)
  const m = totalMin % 60
  return m > 0 ? `${h}小时${m}分钟` : `${h}小时`
})

// Time update
function updateTime() {
  currentTime.value = new Date().toLocaleString('zh-CN', {
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false,
  })
}

// Labels
function getStatusLabel(status: string): string {
  const map: Record<string, string> = {
    running: '执行中', paused: '已暂停', completed: '已完成', terminated: '已终止', pending: '待启动',
  }
  return map[status] || status
}

function getStepStatusLabel(status: string): string {
  const map: Record<string, string> = {
    pending: '待执行', running: '执行中', completed: '已完成', timeout: '超时', skipped: '已跳过', issue: '异常',
  }
  return map[status] || status
}

function formatLogAction(action: string): string {
  const map: Record<string, string> = {
    complete: '步骤完成', issue: '异常上报', start: '步骤启动', timeout: '步骤超时',
    skip: '步骤跳过', force_complete: '强制完成', step_start: '步骤启动', step_complete: '步骤完成',
    step_issue: '异常上报', step_skip: '步骤跳过',
  }
  return map[action] || action
}

function formatTime(dateStr: string | null | undefined): string {
  if (!dateStr) return '--:--'
  const date = new Date(dateStr)
  return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false })
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

// Data loading
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
    drillSteps.value = steps.sort((a, b) => a.seq - b.seq)

    const logs = await drillApi.getLogs(drillId.value)
    recentLogs.value = logs.slice(0, 30)

    loading.value = false
    error.value = null

    connectWebSocket()
    await nextTick()
    initCharts()
  } catch (err: any) {
    error.value = err.message || '加载数据失败'
    console.error('Failed to load drill data:', err)
    loading.value = false
  }
}

function handleRetry() {
  loadData()
}

// WebSocket
function connectWebSocket() {
  if (ws) ws.close()
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

// Fullscreen
function toggleFullscreen() {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen?.()
  } else {
    document.exitFullscreen?.()
  }
}

// ECharts
const CHART_COLORS = {
  accent: '#22C55E',
  error: '#EF4444',
  info: '#3B82F6',
  warning: '#FACC15',
  text: '#F8FAFC',
  textDim: '#64748B',
  grid: '#1E293B',
}

function initCharts() {
  initBulletChart()
  initLineChart()
  initPieChart()
  initRadarChart()

  window.addEventListener('resize', handleResize)
}

function handleResize() {
  bulletChart?.resize()
  lineChart?.resize()
  pieChart?.resize()
  radarChart?.resize()
}

function initBulletChart() {
  if (!bulletChartRef.value) return
  if (bulletChart) bulletChart.dispose()
  bulletChart = echarts.init(bulletChartRef.value)

  const completedSteps = drillSteps.value.filter(s => s.status === 'completed' && s.start_time && s.end_time)
  const data = completedSteps.slice(0, 10).map(s => ({
    name: s.name.length > 12 ? s.name.slice(0, 12) + '...' : s.name,
    value: Math.round((new Date(s.end_time!).getTime() - new Date(s.start_time!).getTime()) / 1000),
  }))

  bulletChart.setOption({
    title: { text: '步骤耗时 TOP 10', left: 'center', top: 8, textStyle: { color: CHART_COLORS.text, fontSize: 13, fontWeight: 600 } },
    grid: { left: 80, right: 30, top: 36, bottom: 20 },
    xAxis: { type: 'value', axisLabel: { color: CHART_COLORS.textDim, formatter: '{s}' }, splitLine: { lineStyle: { color: CHART_COLORS.grid } }, axisLine: { lineStyle: { color: CHART_COLORS.grid } } },
    yAxis: { type: 'category', data: data.map(d => d.name).reverse(), axisLabel: { color: CHART_COLORS.textDim }, axisLine: { show: false }, axisTick: { show: false } },
    series: [{
      type: 'bar', data: data.map(d => d.value).reverse(), barWidth: 16,
      itemStyle: { color: CHART_COLORS.accent, borderRadius: [0, 4, 4, 0] },
      label: { show: true, position: 'right', color: CHART_COLORS.text, fontSize: 11, formatter: '{c}s' },
    }],
  })
}

function initLineChart() {
  if (!lineChartRef.value) return
  if (lineChart) lineChart.dispose()
  lineChart = echarts.init(lineChartRef.value)

  // Build duration trend: cumulative completed steps over time
  const completed = drillSteps.value.filter(s => s.status === 'completed' && s.end_time).sort((a, b) => new Date(a.end_time!).getTime() - new Date(b.end_time!).getTime())
  const labels: string[] = []
  const values: number[] = []
  completed.forEach((s, i) => {
    labels.push(s.name.length > 8 ? s.name.slice(0, 8) + '..' : s.name)
    const dur = s.start_time && s.end_time ? Math.round((new Date(s.end_time).getTime() - new Date(s.start_time).getTime()) / 1000) : 0
    values.push(dur)
  })

  lineChart.setOption({
    title: { text: '耗时趋势', left: 'center', top: 8, textStyle: { color: CHART_COLORS.text, fontSize: 13, fontWeight: 600 } },
    grid: { left: 40, right: 20, top: 36, bottom: 30 },
    xAxis: { type: 'category', data: labels, axisLabel: { color: CHART_COLORS.textDim, rotate: 30 }, axisLine: { lineStyle: { color: CHART_COLORS.grid } }, axisTick: { show: false } },
    yAxis: { type: 'value', axisLabel: { color: CHART_COLORS.textDim, formatter: '{s}s' }, splitLine: { lineStyle: { color: CHART_COLORS.grid } }, axisLine: { lineStyle: { color: CHART_COLORS.grid } } },
    series: [{
      type: 'line', data: values, smooth: true, symbol: 'circle', symbolSize: 6,
      lineStyle: { color: CHART_COLORS.accent, width: 2 },
      itemStyle: { color: CHART_COLORS.accent },
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: 'rgba(34, 197, 94, 0.3)' },
          { offset: 1, color: 'rgba(34, 197, 94, 0.02)' },
        ]),
      },
    }],
  })
}

function initPieChart() {
  if (!pieChartRef.value) return
  if (pieChart) pieChart.dispose()
  pieChart = echarts.init(pieChartRef.value)

  const statusMap: Record<string, { name: string; color: string }> = {
    completed: { name: '已完成', color: CHART_COLORS.accent },
    running: { name: '执行中', color: CHART_COLORS.info },
    issue: { name: '异常', color: CHART_COLORS.error },
    timeout: { name: '超时', color: CHART_COLORS.warning },
    pending: { name: '待执行', color: CHART_COLORS.textDim },
    skipped: { name: '已跳过', color: '#A78BFA' },
  }

  const counts: Record<string, number> = {}
  drillSteps.value.forEach(s => { counts[s.status] = (counts[s.status] || 0) + 1 })
  const data = Object.entries(counts).map(([k, v]) => ({ name: statusMap[k]?.name || k, value: v, itemStyle: { color: statusMap[k]?.color || CHART_COLORS.textDim } }))

  pieChart.setOption({
    title: { text: '状态分布', left: 'center', top: 8, textStyle: { color: CHART_COLORS.text, fontSize: 13, fontWeight: 600 } },
    series: [{
      type: 'pie', radius: ['40%', '65%'], center: ['50%', '58%'],
      roseType: 'area', data,
      label: { color: CHART_COLORS.textDim, fontSize: 11 },
      labelLine: { length: 8, length2: 10 },
    }],
  })
}

function initRadarChart() {
  if (!radarChartRef.value) return
  if (radarChart) radarChart.dispose()
  radarChart = echarts.init(radarChartRef.value)

  // Aggregate workload by executor_team
  const teamCounts: Record<string, number> = {}
  drillSteps.value.forEach(s => {
    if (s.executor_team) {
      teamCounts[s.executor_team] = (teamCounts[s.executor_team] || 0) + 1
    }
  })

  const teams = Object.entries(teamCounts).slice(0, 6)
  if (teams.length === 0) {
    teams.push(['默认团队', drillSteps.value.length])
  }

  radarChart.setOption({
    title: { text: '团队负载', left: 'center', top: 8, textStyle: { color: CHART_COLORS.text, fontSize: 13, fontWeight: 600 } },
    radar: {
      indicator: teams.map(([name]) => ({ name, max: Math.max(...teams.map(([, v]) => v), 1) + 2 })),
      radius: '65%',
      axisName: { color: CHART_COLORS.textDim, fontSize: 11 },
      splitArea: { areaStyle: { color: ['rgba(30,41,59,0.5)', 'rgba(15,23,42,0.5)'] } },
      splitLine: { lineStyle: { color: CHART_COLORS.grid } },
      axisLine: { lineStyle: { color: CHART_COLORS.grid } },
    },
    series: [{
      type: 'radar', data: [{ value: teams.map(([, v]) => v), name: '任务数' }],
      areaStyle: { color: 'rgba(34, 197, 94, 0.2)' },
      lineStyle: { color: CHART_COLORS.accent },
      itemStyle: { color: CHART_COLORS.accent },
    }],
  })
}

function disposeCharts() {
  bulletChart?.dispose(); bulletChart = null
  lineChart?.dispose(); lineChart = null
  pieChart?.dispose(); pieChart = null
  radarChart?.dispose(); radarChart = null
  window.removeEventListener('resize', handleResize)
}

// Lifecycle
onMounted(() => {
  loadData()
  updateTime()
  timeTimer = window.setInterval(updateTime, 1000)
})

onBeforeUnmount(() => {
  if (timeTimer) clearInterval(timeTimer)
  if (refreshTimer) clearInterval(refreshTimer)
  if (ws) { ws.close(); ws = null }
  disposeCharts()
})
</script>

<style lang="scss" scoped>
@use '@/styles/variables' as *;

.screen-root {
  background: #0A0E1A;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  color: $text-primary;
  font-family: $font-family-ui;
  overflow: hidden;
}

// Overlay states
.overlay-state {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  background: #0A0E1A;

  &.error .error-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: $spacing-base;
    color: $text-secondary;

    .el-icon { color: $color-error; }
    p { font-size: $font-size-sm; }
    .el-button {
      background: transparent;
      border: 1px solid $color-accent;
      color: $color-accent;
      &:hover { background: $color-accent-bg; }
    }
  }
}

.loader {
  text-align: center;
  .loader-ring {
    width: 48px;
    height: 48px;
    margin: 0 auto $spacing-base;
    border: 2px solid $border-color-light;
    border-top-color: $color-accent;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }
  .loader-text {
    color: $text-secondary;
    font-size: $font-size-xs;
    letter-spacing: 3px;
    text-transform: uppercase;
  }
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

// Screen content
.screen-content {
  display: flex;
  flex-direction: column;
  height: 100vh;
}

// Header
.screen-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: $spacing-md $spacing-xl;
  height: $header-height;
  background: #0F172A;
  border-bottom: 1px solid $border-color-light;
  flex-shrink: 0;

  .header-left {
    display: flex;
    align-items: center;
    gap: $spacing-md;
  }

  .drill-title {
    font-size: $font-size-lg;
    font-weight: $font-weight-bold;
    margin: 0;
    letter-spacing: 1px;
  }

  .drill-status-tag {
    padding: 2px $spacing-sm;
    border-radius: $radius-sm;
    font-size: $font-size-xs;
    font-weight: $font-weight-semibold;

    &.status-running {
      background: $color-success-bg;
      color: $color-success;
      border: 1px solid $color-accent-border;
    }
    &.status-completed {
      background: $color-success-bg;
      color: $color-success;
      border: 1px solid $color-accent-border;
    }
    &.status-pending {
      background: $color-warning-bg;
      color: $color-warning;
      border: 1px solid rgba(250, 204, 21, 0.2);
    }
    &.status-paused {
      background: $color-warning-bg;
      color: $color-warning;
      border: 1px solid rgba(250, 204, 21, 0.2);
    }
    &.status-terminated {
      background: $color-error-bg;
      color: $color-error;
      border: 1px solid rgba(239, 68, 68, 0.2);
    }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: $spacing-lg;

    .time-display {
      display: flex;
      align-items: center;
      gap: $spacing-xs;
      color: $text-secondary;
      font-size: $font-size-sm;
      font-family: $font-family-mono;
    }

    .btn-icon {
      background: transparent;
      border: 1px solid $border-color-light;
      color: $text-secondary;
      width: 32px;
      height: 32px;
      display: flex;
      align-items: center;
      justify-content: center;
      border-radius: $radius-sm;
      cursor: pointer;
      transition: all 0.2s;

      &:hover {
        border-color: $color-accent;
        color: $color-accent;
      }
    }
  }
}

// KPI row
.kpi-row {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  border-bottom: 1px solid $border-color-light;
  flex-shrink: 0;
}

.kpi-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: $spacing-sm $spacing-base;
  gap: 4px;
  min-height: 56px;
  border-right: 1px solid $border-color-light;

  &:last-child { border-right: none; }

  .kpi-label {
    font-size: $font-size-xs;
    color: $text-tertiary;
    text-transform: uppercase;
    letter-spacing: 1px;
  }

  .kpi-value {
    font-size: $font-size-xl;
    font-weight: $font-weight-bold;
    font-family: $font-family-mono;

    .kpi-total {
      font-weight: $font-weight-regular;
      color: $text-tertiary;
      font-size: $font-size-sm;
    }

    &.kpi-value-accent { color: $color-accent; }
    &.kpi-value-error { color: $color-error; }
  }
}

// Screen main
.screen-main {
  display: flex;
  flex: 1;
  min-height: 0;
}

// Panels
.panel {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-timeline {
  width: 280px;
  background: #0A0F1D;
  border-right: 1px solid $border-color-light;
  flex-shrink: 0;
}

.panel-charts {
  flex: 1;
  background: #0A0E1A;
  min-width: 0;
  padding: $spacing-sm;
}

.panel-logs {
  width: 280px;
  background: #0A0F1D;
  border-left: 1px solid $border-color-light;
  flex-shrink: 0;
}

.panel-header {
  display: flex;
  align-items: center;
  gap: $spacing-xs;
  padding: $spacing-sm $spacing-base;
  background: rgba(34, 197, 94, 0.03);
  border-bottom: 1px solid $border-color-light;
  font-size: $font-size-xs;
  font-weight: $font-weight-semibold;
  color: $text-secondary;
  letter-spacing: 1.5px;
  text-transform: uppercase;
  flex-shrink: 0;

  svg { color: $color-accent; }
}

.panel-body {
  flex: 1;
  overflow-y: auto;
  padding: $spacing-sm;

  &::-webkit-scrollbar { width: 3px; }
  &::-webkit-scrollbar-track { background: transparent; }
  &::-webkit-scrollbar-thumb { background: $border-color; border-radius: 2px; }
}

// Steps list
.steps-list {
  .step-item {
    display: flex;
    align-items: center;
    gap: $spacing-xs;
    padding: $spacing-xs $spacing-sm;
    margin-bottom: 3px;
    background: rgba(0, 0, 0, 0.2);
    border-radius: $radius-sm;
    border: 1px solid transparent;
    transition: all 0.2s;

    &.step-running {
      border-color: $color-accent-border;
      background: $color-accent-bg;
    }
    &.step-issue, &.step-timeout {
      border-color: rgba(239, 68, 68, 0.15);
      background: rgba(239, 68, 68, 0.04);
    }

    .step-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      flex-shrink: 0;
      background: $text-tertiary;

      &.dot-completed { background: $color-accent; }
      &.dot-running {
        background: $color-accent;
        animation: pulse 1.5s ease-in-out infinite;
      }
      &.dot-issue, &.dot-timeout { background: $color-error; }
      &.dot-skipped { background: #A78BFA; }
    }

    .step-info {
      flex: 1;
      min-width: 0;
      display: flex;
      flex-direction: column;
      gap: 2px;
    }

    .step-name {
      font-size: $font-size-xs;
      font-weight: $font-weight-medium;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .step-meta {
      display: flex;
      gap: $spacing-xs;
      font-size: 10px;
      color: $text-tertiary;

      .step-team {
        background: $color-accent-bg;
        padding: 1px 4px;
        border-radius: 2px;
        color: $color-accent;
      }
    }

    .step-badge {
      font-size: 9px;
      font-weight: $font-weight-semibold;
      padding: 1px 4px;
      border-radius: 2px;
      flex-shrink: 0;
      text-transform: uppercase;
    }
  }
}

@keyframes pulse {
  0%, 100% { opacity: 1; box-shadow: 0 0 4px rgba(34, 197, 94, 0.4); }
  50% { opacity: 0.5; box-shadow: 0 0 8px rgba(34, 197, 94, 0.6); }
}

// Chart grid
.chart-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  grid-template-rows: 1fr 1fr;
  gap: $spacing-sm;
  height: 100%;
}

.chart-cell {
  background: #0F172A;
  border: 1px solid $border-color-light;
  border-radius: $radius-md;
  overflow: hidden;
  min-height: 0;
}

.chart-container {
  width: 100%;
  height: 100%;
}

// Logs list
.logs-list {
  .log-item {
    position: relative;
    padding: $spacing-xs $spacing-xs $spacing-xs $spacing-base;
    margin-bottom: $spacing-xs;

    .log-border {
      position: absolute;
      left: 0;
      top: 0;
      bottom: 0;
      width: 3px;
      border-radius: 2px;
    }

    &.log-complete .log-border, &.log-step .log-border { background: $color-accent; }
    &.log-issue .log-border { background: $color-error; }
    &.log-force .log-border { background: $color-info; }
    &.log-skip .log-border { background: #A78BFA; }
    &.log-timeout .log-border { background: $color-warning; }

    .log-body {
      .log-action {
        font-size: $font-size-xs;
        font-weight: $font-weight-medium;
        margin-bottom: 2px;
      }

      .log-text {
        font-size: 11px;
        color: $text-secondary;
        margin-bottom: 2px;
        line-height: 1.3;
      }

      .log-time {
        font-size: 10px;
        font-family: $font-family-mono;
        color: $text-tertiary;
      }
    }
  }

  .log-empty {
    text-align: center;
    color: $text-tertiary;
    font-size: $font-size-xs;
    padding: $spacing-2xl 0;
  }
}
</style>
