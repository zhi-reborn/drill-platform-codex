<template>
  <ScreenLayout>
    <div class="screen-content">
      <!-- 加载状态 -->
      <div v-if="loading" class="loading-overlay">
        <div class="loading-spinner">
          <div class="spinner"></div>
          <p class="loading-text">正在加载演练数据...</p>
        </div>
      </div>

      <!-- 错误状态 -->
      <div v-else-if="error" class="error-overlay">
        <el-icon class="error-icon"><CircleClose /></el-icon>
        <p class="error-text">{{ error }}</p>
        <el-button type="primary" @click="handleRetry">重试</el-button>
      </div>

      <!-- 主内容 -->
      <template v-else-if="currentDrill">
        <!-- 顶部标题栏 -->
        <div class="screen-header">
          <div class="header-left">
            <h1 class="drill-title">{{ currentDrill.name }}</h1>
            <DrillStatusBadge :status="currentDrill.status" type="drill" />
          </div>
          <div class="header-right">
            <div class="time-display">
              <el-icon><Clock /></el-icon>
              <span>{{ currentTime }}</span>
            </div>
            <el-button class="fullscreen-btn" @click="toggleFullscreen">
              <el-icon><FullScreen /></el-icon>
              全屏
            </el-button>
          </div>
        </div>

        <!-- 核心指标栏 -->
        <div class="metrics-bar">
          <div class="metric-card">
            <div class="metric-icon" :class="getMetricClass(0)">
              <el-icon><Monitor /></el-icon>
            </div>
            <div class="metric-content">
              <span class="metric-label">状态</span>
              <span class="metric-value">{{ getStatusLabel(currentDrill.status) }}</span>
            </div>
          </div>
          <div class="metric-card">
            <div class="metric-icon" :class="getMetricClass(1)">
              <el-icon><Timer /></el-icon>
            </div>
            <div class="metric-content">
              <span class="metric-label">进度</span>
              <span class="metric-value">{{ currentDrill.completed_steps }} / {{ currentDrill.total_steps }}</span>
            </div>
          </div>
          <div class="metric-card">
            <div class="metric-icon" :class="getMetricClass(2)">
              <el-icon><TrendCharts /></el-icon>
            </div>
            <div class="metric-content">
              <span class="metric-label">成功率</span>
              <span class="metric-value">{{ successRate }}%</span>
            </div>
          </div>
          <div class="metric-card">
            <div class="metric-icon" :class="getMetricClass(3)">
              <el-icon><User /></el-icon>
            </div>
            <div class="metric-content">
              <span class="metric-label">创建人</span>
              <span class="metric-value">{{ currentDrill.created_by_name }}</span>
            </div>
          </div>
          <div class="metric-card">
            <div class="metric-icon" :class="getMetricClass(4)">
              <el-icon><Calendar /></el-icon>
            </div>
            <div class="metric-content">
              <span class="metric-label">开始时间</span>
              <span class="metric-value">{{ formatTime(currentDrill.started_at) }}</span>
            </div>
          </div>
        </div>

        <!-- 主体内容区 -->
        <div class="screen-body">
          <!-- 左侧：步骤执行面板 -->
          <div class="left-panel">
            <div class="panel-header">
              <el-icon><List /></el-icon>
              <span>步骤执行</span>
            </div>
            <div class="steps-list">
              <div
                v-for="step in drillSteps"
                :key="step.id"
                class="step-item"
                :class="['step-' + step.status, { 'active': step.status === 'running' }]"
              >
                <div class="step-indicator">
                  <el-icon v-if="step.status === 'completed'"><CircleCheck /></el-icon>
                  <el-icon v-else-if="step.status === 'running'"><Loading /></el-icon>
                  <el-icon v-else-if="step.status === 'issue'"><CircleClose /></el-icon>
                  <el-icon v-else><CircleCheck /></el-icon>
                </div>
                <div class="step-content">
                  <div class="step-name">{{ step.step_name }}</div>
                  <div class="step-meta">
                    <span class="step-assignee" v-if="step.assignee_name">
                      <el-icon><User /></el-icon>
                      {{ step.assignee_name }}
                    </span>
                    <span class="step-duration" v-if="step.started_at">
                      {{ calculateDuration(step) }}
                    </span>
                  </div>
                </div>
                <div class="step-status-badge" :class="'badge-' + step.status">
                  {{ getStepStatusLabel(step.status) }}
                </div>
              </div>
            </div>
          </div>

          <!-- 中间：可视化图表区 -->
          <div class="center-panel">
            <!-- 进度环图 -->
            <div class="chart-card progress-chart">
              <div class="chart-header">
                <el-icon><DataLine /></el-icon>
                <span>执行进度</span>
              </div>
              <div class="chart-body">
                <GaugeChart
                  :data="{
                    name: '进度',
                    value: Math.round(currentDrill.completed_steps / currentDrill.total_steps * 100),
                    max: 100
                  }"
                  height="280px"
                />
              </div>
            </div>

            <!-- 步骤状态分布 -->
            <div class="chart-card status-chart">
              <div class="chart-header">
                <el-icon><PieChart /></el-icon>
                <span>步骤状态分布</span>
              </div>
              <div class="chart-body">
                <PieChartComponent :data="stepStatusData" height="220px" />
              </div>
            </div>
          </div>

          <!-- 右侧：时间线与日志 -->
          <div class="right-panel">
            <div class="panel-header">
              <el-icon><Connection /></el-icon>
              <span>执行时间线</span>
            </div>
            <div class="timeline-container">
              <div class="timeline">
                <div
                  v-for="(step, index) in drillSteps"
                  :key="step.id"
                  class="timeline-item"
                  :class="['timeline-' + step.status, { 'active': step.status === 'running' }]"
                >
                  <div class="timeline-dot">
                    <el-icon v-if="step.status === 'completed'"><CircleCheck /></el-icon>
                    <el-icon v-else-if="step.status === 'running'"><Loading /></el-icon>
                    <el-icon v-else-if="step.status === 'issue'"><CircleClose /></el-icon>
                    <el-icon v-else><CircleCheck /></el-icon>
                  </div>
                  <div class="timeline-content">
                    <div class="timeline-step-name">{{ step.step_name }}</div>
                    <div class="timeline-time">
                      {{ formatTime(step.started_at) }}
                      <span v-if="step.completed_at" class="timeline-end-time">
                        → {{ formatTime(step.completed_at) }}
                      </span>
                    </div>
                    <div class="timeline-duration" v-if="step.started_at && step.completed_at">
                      耗时：{{ calculateDuration(step) }}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 告警列表 -->
            <div class="alerts-section" v-if="alerts.length > 0">
              <div class="panel-header">
                <el-icon><Warning /></el-icon>
                <span>告警事件</span>
              </div>
              <div class="alerts-list">
                <div
                  v-for="alert in alerts"
                  :key="alert.id"
                  class="alert-item"
                  :class="'alert-' + alert.level"
                >
                  <el-icon class="alert-icon"><Warning /></el-icon>
                  <div class="alert-content">
                    <div class="alert-message">{{ alert.message }}</div>
                    <div class="alert-time">{{ formatTime(alert.created_at) }}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 底部进度条 -->
        <div class="bottom-progress">
          <el-progress
            :percentage="Math.round(currentDrill.completed_steps / currentDrill.total_steps * 100)"
            :stroke-width="4"
            :show-text="false"
            :status="currentDrill.status === 'completed' ? 'success' : undefined"
          />
        </div>
      </template>
    </div>
  </ScreenLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute } from 'vue-router'
import {
  Clock,
  FullScreen,
  Monitor,
  Timer,
  TrendCharts,
  User,
  Calendar,
  List,
  DataLine,
  PieChart,
  Warning,
  CircleCheck,
  CircleClose,
  Loading,
  Connection,
} from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import ScreenLayout from '@/components/screen/ScreenLayout.vue'
import GaugeChart from '@/components/charts/GaugeChart.vue'
import PieChartComponent from '@/components/charts/PieChart.vue'
import DrillStatusBadge from '@/components/common/DrillStatusBadge.vue'
import type { DrillInstance, StepInstance } from '@/types'
import instancesData from '@/mock/data/instances.json'
import stepsData from '@/mock/data/steps.json'
import notificationsData from '@/mock/data/notifications.json'

const route = useRoute()
const loading = ref(true)
const error = ref<string | null>(null)
const currentTime = ref(new Date().toLocaleString('zh-CN'))

// 获取演练 ID
const drillId = computed(() => {
  const id = route.params.id
  return typeof id === 'string' ? parseInt(id, 10) : null
})

// 当前演练
const currentDrill = ref<DrillInstance | null>(null)

// 演练步骤
const drillSteps = computed(() => {
  if (!drillId.value) return []
  return (stepsData as StepInstance[])
    .filter(s => s.drill_id === drillId.value)
    .sort((a, b) => a.order_index - b.order_index)
})

// 成功率
const successRate = computed(() => {
  if (drillSteps.value.length === 0) return 0
  const completed = drillSteps.value.filter(s => s.status === 'completed').length
  return Math.round((completed / drillSteps.value.length) * 100)
})

// 步骤状态分布
const stepStatusData = computed(() => {
  const counts: Record<string, number> = {}
  drillSteps.value.forEach(step => {
    counts[step.status] = (counts[step.status] || 0) + 1
  })
  return Object.entries(counts).map(([name, value]) => ({ name, value }))
})

// 告警列表
const alerts = computed(() => {
  const items: Array<{
    id: number
    level: 'info' | 'warning' | 'error'
    message: string
    created_at: string
  }> = []

  // 从步骤中提取异常和超时
  drillSteps.value.forEach((step, i) => {
    if (step.status === 'issue') {
      items.push({
        id: i,
        level: 'error',
        message: `步骤「${step.step_name}」异常${step.error_message ? `: ${step.error_message}` : ''}`,
        created_at: step.started_at || new Date().toISOString(),
      })
    }
    if (step.status === 'timeout') {
      items.push({
        id: i + 1000,
        level: 'warning',
        message: `步骤「${step.step_name}」超时`,
        created_at: step.started_at || new Date().toISOString(),
      })
    }
  })

  // 从通知中补充
  if (currentDrill.value) {
    const relevantNotifs = (notificationsData as Array<Record<string, unknown>>).filter(
      n => n.title === currentDrill.value?.name
    )
    relevantNotifs.forEach((n, i) => {
      const type = n.type as string
      if (type === 'drill_started') {
        items.push({
          id: i + 2000,
          level: 'info',
          message: `演练「${n.title}」已启动`,
          created_at: n.created_at as string,
        })
      }
      if (type === 'drill_paused') {
        items.push({
          id: i + 4000,
          level: 'warning',
          message: `演练「${n.title}」已暂停`,
          created_at: n.created_at as string,
        })
      }
      if (type === 'drill_terminated') {
        items.push({
          id: i + 5000,
          level: 'error',
          message: `演练「${n.title}」已终止`,
          created_at: n.created_at as string,
        })
      }
    })
  }

  items.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
  return items.slice(0, 20)
})

// 时间更新
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

// 加载数据
async function loadData() {
  if (!drillId.value) {
    error.value = '无效的演练 ID'
    loading.value = false
    return
  }

  loading.value = true
  error.value = null

  try {
    // 模拟 API 调用延迟
    await new Promise(resolve => setTimeout(resolve, 300))

    const drill = (instancesData as DrillInstance[]).find(d => d.id === drillId.value)
    if (!drill) {
      error.value = '演练不存在'
      return
    }

    currentDrill.value = drill
  } catch (err) {
    error.value = '加载数据失败'
    console.error('Failed to load drill data:', err)
  } finally {
    loading.value = false
  }
}

function handleRetry() {
  loadData()
}

function getStatusLabel(status: string): string {
  const map: Record<string, string> = {
    running: '进行中',
    paused: '已暂停',
    completed: '已完成',
    terminated: '已终止',
    pending: '待开始',
  }
  return map[status] || status
}

function getStepStatusLabel(status: string): string {
  const map: Record<string, string> = {
    pending: '待开始',
    running: '执行中',
    completed: '已完成',
    timeout: '超时',
    skipped: '已跳过',
    issue: '异常',
  }
  return map[status] || status
}

function calculateDuration(step: StepInstance): string {
  if (!step.started_at) return '-'
  if (!step.completed_at) {
    const start = new Date(step.started_at).getTime()
    const now = Date.now()
    const diff = Math.floor((now - start) / 1000)
    return formatDuration(diff)
  }
  const start = new Date(step.started_at).getTime()
  const end = new Date(step.completed_at).getTime()
  const diff = Math.floor((end - start) / 1000)
  return formatDuration(diff)
}

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins}m ${secs}s`
}

function formatTime(dateStr: string | null | undefined): string {
  if (!dateStr) return '-'
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function getMetricClass(index: number): string {
  const classes = ['metric-primary', 'metric-success', 'metric-info', 'metric-warning', 'metric-info']
  return classes[index] || ''
}

function toggleFullscreen() {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen?.()
  } else {
    document.exitFullscreen?.()
  }
}

// 生命周期
onMounted(() => {
  loadData()
  updateTime()
  timeTimer = window.setInterval(updateTime, 1000)

  // 模拟实时数据刷新
  const refreshTimer = window.setInterval(() => {
    if (currentDrill.value && currentDrill.value.status === 'running') {
      // 实际项目中应从 API 获取最新数据
      console.log('Refreshing drill data...')
    }
  }, 5000)

  onBeforeUnmount(() => {
    clearInterval(timeTimer)
    clearInterval(refreshTimer)
  })
})
</script>

<style lang="scss" scoped>
@use '@/styles/variables' as *;

// 深色主题变量
$bg-primary: #0d1117;
$bg-secondary: #161b22;
$bg-tertiary: #21262d;
$bg-card: #1c2128;
$border-color: #30363d;
$text-primary: #f0f6fc;
$text-secondary: #8b949e;
$text-tertiary: #484f58;
$color-success: #2ea043;
$color-warning: #d29922;
$color-error: #da3633;
$color-info: #58a6ff;
$color-running: #238636;

.screen-content {
  width: 100vw;
  height: 100vh;
  background: $bg-primary;
  color: $text-primary;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

// 加载/错误状态
.loading-overlay,
.error-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: $bg-primary;
  z-index: 100;
}

.loading-spinner {
  text-align: center;

  .spinner {
    width: 48px;
    height: 48px;
    border: 3px solid $bg-tertiary;
    border-top-color: $color-info;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin: 0 auto 16px;
  }

  .loading-text {
    color: $text-secondary;
    font-size: 14px;
  }
}

.error-overlay {
  flex-direction: column;
  gap: 16px;

  .error-icon {
    font-size: 48px;
    color: $color-error;
  }

  .error-text {
    color: $text-secondary;
    font-size: 14px;
  }
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

// 顶部标题栏
.screen-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background: $bg-secondary;
  border-bottom: 1px solid $border-color;

  .header-left {
    display: flex;
    align-items: center;
    gap: 16px;

    .drill-title {
      font-size: 20px;
      font-weight: 600;
      color: $text-primary;
      margin: 0;
    }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 16px;

    .time-display {
      display: flex;
      align-items: center;
      gap: 8px;
      color: $text-secondary;
      font-size: 14px;

      .el-icon {
        font-size: 16px;
      }
    }

    .fullscreen-btn {
      background: $bg-tertiary;
      border-color: $border-color;
      color: $text-primary;

      &:hover {
        background: $bg-card;
        border-color: $color-info;
      }
    }
  }
}

// 指标栏
.metrics-bar {
  display: flex;
  gap: 12px;
  padding: 16px 24px;
  background: $bg-secondary;
  border-bottom: 1px solid $border-color;
}

.metric-card {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: $bg-card;
  border-radius: 8px;
  border: 1px solid $border-color;

  .metric-icon {
    width: 40px;
    height: 40px;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 20px;

    &.metric-primary {
      background: rgba(88, 166, 255, 0.15);
      color: $color-info;
    }

    &.metric-success {
      background: rgba(46, 160, 67, 0.15);
      color: $color-success;
    }

    &.metric-warning {
      background: rgba(210, 153, 34, 0.15);
      color: $color-warning;
    }

    &.metric-info {
      background: rgba(88, 166, 255, 0.1);
      color: $color-info;
    }
  }

  .metric-content {
    display: flex;
    flex-direction: column;

    .metric-label {
      font-size: 12px;
      color: $text-secondary;
      margin-bottom: 4px;
    }

    .metric-value {
      font-size: 18px;
      font-weight: 600;
      color: $text-primary;
    }
  }
}

// 主体内容区
.screen-body {
  flex: 1;
  display: grid;
  grid-template-columns: 320px 1fr 360px;
  gap: 16px;
  padding: 16px;
  overflow: hidden;
}

// 左侧面板
.left-panel,
.right-panel {
  background: $bg-secondary;
  border-radius: 12px;
  border: 1px solid $border-color;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: $bg-tertiary;
  border-bottom: 1px solid $border-color;
  font-size: 14px;
  font-weight: 600;
  color: $text-primary;

  .el-icon {
    font-size: 16px;
  }
}

.steps-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.step-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  margin-bottom: 8px;
  background: $bg-card;
  border-radius: 8px;
  border: 1px solid $border-color;
  transition: all 0.2s;

  &.active {
    border-color: $color-running;
    background: rgba(35, 134, 54, 0.1);
  }

  .step-indicator {
    font-size: 20px;
    flex-shrink: 0;

    .el-icon {
      &.is-loading {
        animation: spin 1s linear infinite;
      }
    }
  }

  .step-content {
    flex: 1;
    min-width: 0;

    .step-name {
      font-size: 14px;
      font-weight: 500;
      color: $text-primary;
      margin-bottom: 4px;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .step-meta {
      display: flex;
      align-items: center;
      gap: 12px;
      font-size: 12px;
      color: $text-secondary;

      .step-assignee,
      .step-duration {
        display: flex;
        align-items: center;
        gap: 4px;

        .el-icon {
          font-size: 12px;
        }
      }
    }
  }

  .step-status-badge {
    padding: 4px 8px;
    border-radius: 4px;
    font-size: 12px;
    font-weight: 500;

    &.badge-pending {
      background: rgba(139, 148, 158, 0.15);
      color: $text-secondary;
    }

    &.badge-running {
      background: rgba(35, 134, 54, 0.2);
      color: $color-running;
    }

    &.badge-completed {
      background: rgba(46, 160, 67, 0.2);
      color: $color-success;
    }

    &.badge-issue,
    &.badge-timeout {
      background: rgba(218, 54, 51, 0.2);
      color: $color-error;
    }

    &.badge-skipped {
      background: rgba(88, 166, 255, 0.15);
      color: $color-info;
    }
  }
}

// 中间图表区
.center-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow-y: auto;
}

.chart-card {
  background: $bg-secondary;
  border-radius: 12px;
  border: 1px solid $border-color;
  display: flex;
  flex-direction: column;

  &.progress-chart {
    flex: 1;
    min-height: 320px;
  }

  &.status-chart {
    flex: 1;
    min-height: 280px;
  }

  .chart-header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 16px;
    background: $bg-tertiary;
    border-bottom: 1px solid $border-color;
    font-size: 14px;
    font-weight: 600;
    color: $text-primary;

    .el-icon {
      font-size: 16px;
    }
  }

  .chart-body {
    flex: 1;
    padding: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
}

// 右侧面板
.right-panel {
  .timeline-container {
    flex: 1;
    overflow-y: auto;
    padding: 16px;
  }

  .alerts-section {
    border-top: 1px solid $border-color;
    max-height: 200px;
    overflow-y: auto;
  }
}

.timeline {
  position: relative;
  padding-left: 24px;

  &::before {
    content: '';
    position: absolute;
    left: 8px;
    top: 0;
    bottom: 0;
    width: 2px;
    background: $border-color;
  }
}

.timeline-item {
  position: relative;
  padding-bottom: 20px;

  &::before {
    content: '';
    position: absolute;
    left: -20px;
    top: 4px;
    width: 18px;
    height: 18px;
    border-radius: 50%;
    background: $bg-card;
    border: 2px solid $border-color;
    z-index: 1;
  }

  &.timeline-completed::before {
    background: $color-success;
    border-color: $color-success;
  }

  &.timeline-running::before {
    background: $color-running;
    border-color: $color-running;
    animation: pulse 1.5s ease-in-out infinite;
  }

  &.timeline-issue::before {
    background: $color-error;
    border-color: $color-error;
  }

  &.active {
    .timeline-content {
      background: rgba(35, 134, 54, 0.1);
    }
  }

  .timeline-dot {
    position: absolute;
    left: -16px;
    top: 4px;
    width: 16px;
    height: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 14px;
    color: $bg-primary;
    z-index: 2;

    .el-icon.is-loading {
      animation: spin 1s linear infinite;
    }
  }

  .timeline-content {
    padding: 10px 12px;
    background: $bg-card;
    border-radius: 8px;
    border: 1px solid $border-color;

    .timeline-step-name {
      font-size: 13px;
      font-weight: 500;
      color: $text-primary;
      margin-bottom: 6px;
    }

    .timeline-time {
      font-size: 12px;
      color: $text-secondary;

      .timeline-end-time {
        color: $text-tertiary;
      }
    }

    .timeline-duration {
      font-size: 11px;
      color: $text-tertiary;
      margin-top: 4px;
    }
  }
}

@keyframes pulse {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba(35, 134, 54, 0.4);
  }
  50% {
    box-shadow: 0 0 0 8px rgba(35, 134, 54, 0);
  }
}

.alerts-list {
  padding: 8px;
}

.alert-item {
  display: flex;
  gap: 12px;
  padding: 10px 12px;
  margin-bottom: 8px;
  background: $bg-card;
  border-radius: 8px;
  border: 1px solid $border-color;

  &.alert-error {
    border-color: rgba($color-error, 0.3);
    background: rgba($color-error, 0.1);
  }

  &.alert-warning {
    border-color: rgba($color-warning, 0.3);
    background: rgba($color-warning, 0.1);
  }

  &.alert-info {
    border-color: rgba($color-info, 0.3);
    background: rgba($color-info, 0.1);
  }

  .alert-icon {
    font-size: 18px;
    flex-shrink: 0;

    &.alert-error {
      color: $color-error;
    }

    &.alert-warning {
      color: $color-warning;
    }

    &.alert-info {
      color: $color-info;
    }
  }

  .alert-content {
    flex: 1;
    min-width: 0;

    .alert-message {
      font-size: 13px;
      color: $text-primary;
      margin-bottom: 4px;
    }

    .alert-time {
      font-size: 11px;
      color: $text-tertiary;
    }
  }
}

// 底部进度条
.bottom-progress {
  padding: 0 24px;
  background: $bg-secondary;
  border-top: 1px solid $border-color;

  :deep(.el-progress) {
    .el-progress__bar {
      background-color: $bg-tertiary;
    }

    .el-progress__indicator {
      color: $text-primary;
    }
  }
}

// 滚动条样式
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: $bg-tertiary;
  border-radius: 4px;
}

::-webkit-scrollbar-thumb {
  background: $border-color;
  border-radius: 4px;

  &:hover {
    background: $text-tertiary;
  }
}
</style>
