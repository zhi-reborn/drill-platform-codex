<template>
  <div class="screen-root">
    <!-- Background layers -->
    <div class="bg-grid" />
    <div class="bg-scan" />
    <div class="bg-vignette" />

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
      <!-- ========== HEADER ========== -->
      <header class="screen-header">
        <div class="header-deco header-deco-left" />
        <div class="header-title-block">
          <h1 class="drill-title">容灾演练指挥中心</h1>
        </div>
        <div class="header-meta">
          <span class="meta-label">系统时间</span>
          <span class="meta-divider">|</span>
          <span class="meta-value">{{ systemTime }}</span>
        </div>
        <button class="btn-icon" @click="toggleFullscreen" title="全屏切换">
          <FullScreen :size="16" />
        </button>
        <div class="header-deco header-deco-right" />
      </header>

      <!-- ========== TOP KPI ROW ========== -->
      <section class="kpi-row">
        <div class="kpi-card">
          <span class="kpi-orb" />
          <div class="kpi-copy">
            <span class="kpi-label-zh">演练状态</span>
          </div>
          <div class="kpi-value-row kpi-status-row">
            <span class="status-dot" :class="'dot-' + (currentDrill.status || '')" />
            <span class="kpi-value-text" :class="'txt-' + (currentDrill.status || '')">{{ drillStatusLabel }}</span>
          </div>
          <span class="kpi-sub">编号: D-{{ String(currentDrill.id).padStart(10, '0') }}</span>
        </div>

        <div class="kpi-card kpi-progress-card">
          <span class="kpi-orb" />
          <div class="kpi-copy">
            <span class="kpi-label-zh">整体进度</span>
          </div>
          <div class="kpi-value-row kpi-progress-row">
            <div class="progress-ring-wrap">
              <svg class="progress-ring-svg" viewBox="0 0 100 100">
                <defs>
                  <linearGradient id="ring-grad" x1="0%" y1="0%" x2="100%" y2="100%">
                    <stop offset="0%" stop-color="#fff" stop-opacity="0.95" />
                    <stop offset="60%" stop-color="#2cf8d8" stop-opacity="1" />
                    <stop offset="100%" stop-color="#00d4aa" stop-opacity="1" />
                  </linearGradient>
                </defs>
                <circle class="progress-ring-bg" cx="50" cy="50" r="46" />
                <circle class="progress-ring-fill"
                  cx="50" cy="50" r="46"
                  :style="{ strokeDashoffset: 289.03 * (1 - progressPercent / 100) }" />
              </svg>
              <span class="progress-ring-text">{{ progressPercent }}<small>%</small></span>
            </div>
            <div class="progress-node-block">
              <div class="node-count-row">
                <span class="kpi-value-num">{{ completedCount }}</span>
                <span class="node-separator">/</span>
                <span class="node-total">{{ totalCount }}</span>
              </div>
              <span class="progress-sub-label">完成节点</span>
            </div>
          </div>
        </div>

        <div class="kpi-card">
          <span class="kpi-orb" />
          <div class="kpi-copy">
            <span class="kpi-label-zh">总耗时</span>
          </div>
          <div class="kpi-value-row mono">
            <span class="kpi-value-num">{{ elapsedParts.h }}</span>
            <span class="kpi-value-sep">:</span>
            <span class="kpi-value-num">{{ elapsedParts.m }}</span>
            <span class="kpi-value-sep">:</span>
            <span class="kpi-value-num">{{ elapsedParts.s }}</span>
          </div>
          <span class="kpi-sub">开始 {{ drillStartTime ? formatHM(drillStartTime) : '--:--' }} 预计剩余 {{ estimatedRemaining }}</span>
        </div>

        <div class="kpi-card">
          <span class="kpi-orb" />
          <div class="kpi-copy">
            <span class="kpi-label-zh">当前阶段/环节</span>
          </div>
          <div class="kpi-value-row">
            <span class="kpi-value-text">阶段{{ currentPhaseIndex + 1 }}</span>
          </div>
          <span class="kpi-sub">{{ currentPhaseName }} · {{ todayStr }}</span>
        </div>
      </section>

      <!-- ========== MAIN GRID ========== -->
      <main class="screen-main">
        <!-- LEFT: Stage overview -->
        <section class="panel panel-stages">
          <div class="panel-header">
            <span class="panel-deco-corner tl" />
            <span class="panel-deco-corner tr" />
            <span class="panel-title-zh">演练阶段总览</span>
            <span class="panel-badge">{{ stages.length }}</span>
          </div>
          <div class="panel-body stages-list">
            <div
              v-for="(stage, idx) in stages"
              :key="idx"
              class="stage-card"
              :class="['stage-' + stage.status]"
            >
              <div class="stage-card-top">
                <div class="stage-name-block">
                  <span class="stage-index">阶段{{ idx + 1 }}</span>
                  <span class="stage-name">{{ stage.name }}</span>
                </div>
                <span class="stage-badge" :class="'badge-' + stage.status">
                  {{ stageBadgeLabel(stage.status) }}
                </span>
              </div>
              <div class="stage-card-mid">
                <span class="stage-time">{{ stage.timeRange }}</span>
                <span class="stage-meta-label">已耗时/限额</span>
              </div>
              <div class="stage-segments">
                <span
                  v-for="(seg, si) in stage.segments"
                  :key="si"
                  class="segment"
                  :class="'seg-' + seg"
                />
              </div>
              <div class="stage-card-bottom">
                <span class="stage-meta">
                  <span class="meta-key">环节</span>
                  <span class="meta-val">{{ stage.completedPhases }} / {{ stage.totalPhases }}</span>
                </span>
                <span class="stage-meta">
                  <span class="meta-key">步骤</span>
                  <span class="meta-val">{{ stage.completedSteps }} / {{ stage.totalSteps }}</span>
                </span>
                <span class="stage-meta">
                  <span class="meta-key">团队</span>
                  <span class="meta-val">{{ stage.team || '运维部' }}</span>
                </span>
              </div>
            </div>
          </div>
        </section>

        <!-- CENTER: Phase ring -->
        <section class="panel panel-center">
          <div class="center-stage">
            <PhaseRing
              :phases="ringPhases"
              :current-index="currentPhaseIndex"
              :progress="progressPercent"
              :center-numerator="completedCount"
              :center-denominator="totalCount"
              :center-hint="`注入应用实例 · 阶段${currentPhaseIndex + 1}`"
              :size="ringSize"
            />
          </div>
        </section>

        <!-- RIGHT: Alerts + logs -->
        <section class="panel panel-right">
          <!-- Active warnings -->
          <div class="sub-panel sub-warn">
            <div class="panel-header">
              <span class="panel-deco-corner tl" />
              <span class="panel-deco-corner tr" />
              <span class="panel-title-zh">执行中步骤</span>
              <span class="panel-realtime">
                <span class="rt-dot" />
                实时
              </span>
            </div>
            <div class="panel-body warn-list">
              <div
                v-for="(alert, ai) in activeAlerts"
                :key="ai"
                class="alert-card"
                :class="'alert-' + alert.level"
              >
                <div class="alert-head">
                  <span class="alert-arrow">▸</span>
                  <span class="alert-title">{{ alert.title }}</span>
                  <span class="alert-team">{{ alert.team }}</span>
                </div>
                <div class="alert-bar">
                  <div class="alert-bar-fill" :style="{ width: alert.progress + '%' }" />
                </div>
                <div class="alert-foot">
                  <span class="alert-meta">团队: {{ alert.team }}</span>
                  <span class="alert-meta">限时: {{ alert.limit }}</span>
                  <span class="alert-meta">已耗时: {{ alert.elapsed }}</span>
                </div>
                <div class="alert-status">
                  <span class="alert-progress-pct">{{ alert.progress }}%</span>
                  <span class="alert-status-tag">{{ alert.statusLabel }}</span>
                </div>
              </div>
              <div v-if="activeAlerts.length === 0" class="empty-tip">暂无活跃告警</div>
            </div>
          </div>

          <!-- Real-time logs -->
          <div class="sub-panel sub-logs">
            <div class="panel-header">
              <span class="panel-deco-corner tl" />
              <span class="panel-deco-corner tr" />
              <span class="panel-title-zh">实时操作日志</span>
            </div>
            <div class="panel-body logs-wrap">
              <div class="log-thead">
                <span class="col-time">时间</span>
                <span class="col-step">节点名称</span>
                <span class="col-desc">描述</span>
              </div>
              <div class="log-tbody">
                <div
                  v-for="(log, idx) in recentLogs"
                  :key="(log.ID ?? idx)"
                  class="log-row"
                  :class="'log-' + logActionClass(log.Action)"
                >
                  <span class="col-time">{{ formatTime(log.CreatedAt) }}</span>
                  <span class="col-step" :title="resolveStepName(log)">{{ resolveStepName(log) }}</span>
                  <span class="col-desc">{{ log.Remark || logActionLabel(log.Action) }}</span>
                </div>
                <div v-if="recentLogs.length === 0" class="empty-tip">暂无日志</div>
              </div>
            </div>
          </div>
        </section>
      </main>

      <!-- Footer decorations -->
      <footer class="screen-footer" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute } from 'vue-router'
import { CircleClose, FullScreen } from '@element-plus/icons-vue'
import type { StepInstance, StepInstanceLog, DrillInstance, StepStatus, DrillStatus } from '@/types/instance'
import { drillApi } from '@/api/modules/drill'
import { useAuthStore } from '@/stores/auth'
import PhaseRing from '@/components/screen/PhaseRing.vue'

const route = useRoute()
const loading = ref(true)
const error = ref<string | null>(null)
const viewportWidth = ref(window.innerWidth)

let ws: WebSocket | null = null
let refreshTimer: number | null = null
let timeTimer: number | null = null
let componentDestroyed = false

// 顶部时间
const systemTime = ref(formatSystemTime(new Date()))
const todayStr = computed(() => {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
})

// 当前路由 drill id
const drillId = computed(() => {
  const id = route.params.id
  return typeof id === 'string' ? parseInt(id, 10) : null
})

// 数据
const currentDrill = ref<DrillInstance | null>(null)
const drillSteps = ref<StepInstance[]>([])
const recentLogs = ref<StepInstanceLog[]>([])

// === 树形步骤辅助 ===
// 构建父子映射，支持任意深度嵌套（阶段→环节→任务→操作步骤）
const childMap = computed(() => {
  const map = new Map<number, StepInstance[]>()
  for (const s of drillSteps.value) {
    const pid = s.parent_step_id
    if (pid) {
      const list = map.get(pid) || []
      list.push(s)
      map.set(pid, list)
    }
  }
  return map
})

// 递归收集某步骤子树中的所有叶子节点
function collectLeaves(stepId: number): StepInstance[] {
  const children = childMap.value.get(stepId)
  if (!children || children.length === 0) {
    // 自身就是叶子
    const self = drillSteps.value.find(s => s.id === stepId)
    return self ? [self] : []
  }
  const leaves: StepInstance[] = []
  for (const c of children) {
    leaves.push(...collectLeaves(c.id))
  }
  return leaves
}

// 叶子步骤：无子节点的步骤（实际执行的操作步骤）
const leafSteps = computed(() => {
  const allSteps = drillSteps.value
  if (allSteps.length === 0) return []
  const hasChild = new Set<number>()
  for (const s of allSteps) {
    if (s.parent_step_id) hasChild.add(s.parent_step_id)
  }
  const leaves = allSteps.filter(s => !hasChild.has(s.id))
  return leaves.length > 0 ? leaves : allSteps
})

// 根步骤（阶段）
const rootSteps = computed(() => {
  const allSteps = drillSteps.value
  const hasParent = new Set<number>()
  for (const s of allSteps) {
    if (s.parent_step_id) hasParent.add(s.id)
  }
  // 没有 parent_step_id 的就是根
  return allSteps.filter(s => !s.parent_step_id).sort((a, b) => a.seq - b.seq)
})

// === KPI 计算 ===
const completedCount = computed(() =>
  leafSteps.value.filter(s => ['completed', 'skipped', 'timeout', 'issue'].includes(s.status)).length
)
const totalCount = computed(() => leafSteps.value.length)
const progressPercent = computed(() => {
  if (totalCount.value === 0) return 0
  return Math.round((completedCount.value / totalCount.value) * 100)
})

// 演练开始时间（兼容 start_time / started_at 两种字段名）
const drillStartTime = computed(() =>
  (currentDrill.value as any)?.start_time || (currentDrill.value as any)?.started_at || null
)

// 总耗时（演练开始至今 / 已完成时刻起算）
const elapsedSeconds = ref(0)
const elapsedParts = computed(() => {
  const t = elapsedSeconds.value
  return {
    h: String(Math.floor(t / 3600)).padStart(2, '0'),
    m: String(Math.floor((t % 3600) / 60)).padStart(2, '0'),
    s: String(t % 60).padStart(2, '0'),
  }
})

// 预计剩余时间（基于运行中/待执行步骤的 timeout 汇总）
const estimatedRemaining = computed(() => {
  // 依赖 elapsedSeconds 每秒刷新
  const _t = elapsedSeconds.value
  const nowMs = Date.now()
  const running = leafSteps.value.filter(s => s.status === 'running')
  const pending = leafSteps.value.filter(s => s.status === 'pending')
  if (running.length === 0 && pending.length === 0) return '--'
  let remainSec = 0
  for (const s of running) {
    if (s.start_time && s.timeout_minutes) {
      const elapsed = (nowMs - new Date(s.start_time).getTime()) / 1000
      remainSec += Math.max(0, s.timeout_minutes * 60 - elapsed)
    } else {
      remainSec += (s.timeout_minutes || 30) * 60
    }
  }
  for (const s of pending) {
    remainSec += (s.timeout_minutes || 30) * 60
  }
  if (remainSec <= 0) return '0m'
  const h = Math.floor(remainSec / 3600)
  const m = Math.ceil((remainSec % 3600) / 60)
  if (h > 0) return `${h}h${m > 0 ? m + 'm' : ''}`
  return `${m}m`
})

// 状态标签
const drillStatusLabel = computed(() => {
  const map: Record<string, string> = {
    running: '执行中', paused: '已暂停', completed: '已完成', terminated: '已终止', pending: '待启动',
  }
  return map[currentDrill.value?.status || ''] || '未知'
})

// === 阶段（stages） ===
// 将步骤均匀分成 4 个阶段；如果演练有 8 步且要演示 6/8，可以从 completedSteps 截取
const STAGE_NAMES = [
  '业务验收与告警',
  '演练复盘与行动',
  '演练启动与人员',
  '基线指标与备份',
]

const stages = computed(() => {
  const allSteps = drillSteps.value
  if (allSteps.length === 0) return []

  const roots = rootSteps.value
  const hasHierarchy = roots.length > 0 && leafSteps.value.length < allSteps.length

  if (hasHierarchy) {
    // 有层级结构：每个根步骤就是一个阶段
    return roots.map((root, idx) => {
      const directChildren = childMap.value.get(root.id) || []
      const leaves = collectLeaves(root.id)
      // 终态判断（completed/skipped/timeout/issue 都视为已结束）
      const isTerminal = (s: StepInstance) => ['completed', 'skipped', 'timeout', 'issue'].includes(s.status)
      // 环节统计：基于每个环节下的叶子步骤完成情况判断
      const completedPhases = directChildren.filter(c => {
        const phaseLeaves = collectLeaves(c.id)
        return phaseLeaves.length > 0 && phaseLeaves.every(l => isTerminal(l))
      }).length
      const totalPhases = directChildren.length
      // 步骤统计：叶子节点（实际操作步骤）
      const finishedLeaves = leaves.filter(s => isTerminal(s)).length
      const totalLeaves = leaves.length
      const running = leaves.some(s => s.status === 'running')
      const hasIssue = leaves.some(s => s.status === 'issue' || s.status === 'timeout')
      const allDone = leaves.every(s => isTerminal(s)) && totalLeaves > 0
      // 部分完成：有已完成的叶子步骤，但不是全部
      const hasProgress = finishedLeaves > 0 && !allDone
      const status = allDone ? 'done' : (running ? 'running' : (hasIssue ? 'issue' : (hasProgress ? 'running' : 'pending')))
      // 时间范围
      const started = leaves.find(s => s.start_time)?.start_time
      const ended = [...leaves].reverse().find(s => s.end_time)?.end_time
      let endStr: string | null = ended ?? null
      if (!endStr) {
        const t = leaves[0]?.timeout_minutes
        if (t) endStr = new Date(Date.now() + t * 60000).toISOString()
      }
      const timeRange = `${formatHM(started)} / ${formatHM(endStr)}`
      // 段落
      const segCount = 18
      const segs: string[] = []
      for (let i = 0; i < segCount; i++) {
        if (i < finishedLeaves) segs.push('done')
        else if (i === finishedLeaves && running) segs.push('active')
        else if (i < totalLeaves) segs.push('todo')
        else segs.push('empty')
      }
      const team = leaves.find(s => s.executor_team)?.executor_team
        || leaves.find(s => s.assignee_names)?.assignee_names
        || '运维部'
      return {
        name: root.name || STAGE_NAMES[idx % STAGE_NAMES.length],
        status,
        timeRange,
        completedPhases,
        totalPhases,
        completedSteps: finishedLeaves,
        totalSteps: totalLeaves,
        segments: segs,
        team,
      }
    })
  }

  // 无层级：均匀分成 4 个阶段
  const total = allSteps.length
  const stageCount = Math.min(4, total)
  const perStage = Math.ceil(total / stageCount)
  const isTerminal = (s: StepInstance) => ['completed', 'skipped', 'timeout', 'issue'].includes(s.status)
  return Array.from({ length: stageCount }).map((_, idx) => {
    const slice = allSteps.slice(idx * perStage, (idx + 1) * perStage)
    const finished = slice.filter(s => isTerminal(s)).length
    const running = slice.some(s => s.status === 'running')
    const hasIssue = slice.some(s => s.status === 'issue' || s.status === 'timeout')
    const allDone = slice.every(s => isTerminal(s)) && slice.length > 0
    const hasProgress = finished > 0 && !allDone
    const status = allDone ? 'done' : (running ? 'running' : (hasIssue ? 'issue' : (hasProgress ? 'running' : 'pending')))
    const started = slice.find(s => s.start_time)?.start_time
    const ended = slice.find(s => s.end_time)?.end_time
    let endStr: string | null = ended ?? null
    if (!endStr) {
      const t = slice[0]?.timeout_minutes
      if (t) endStr = new Date(Date.now() + t * 60000).toISOString()
    }
    const timeRange = `${formatHM(started)} / ${formatHM(endStr)}`
    const segCount = 18
    const segs: string[] = []
    for (let i = 0; i < segCount; i++) {
      if (i < finished) segs.push('done')
      else if (i === finished && running) segs.push('active')
      else if (i < slice.length) segs.push('todo')
      else segs.push('empty')
    }
    const team = slice.find(s => s.executor_team)?.executor_team
      || slice.find(s => s.assignee_names)?.assignee_names
      || '运维部'
    return {
      name: STAGE_NAMES[idx % STAGE_NAMES.length],
      status,
      timeRange,
      completedPhases: 0,
      totalPhases: 0,
      completedSteps: finished,
      totalSteps: slice.length,
      segments: segs,
      team,
    }
  })
})

// 当前阶段 index（高亮 + 中心环）
const currentPhaseIndex = computed(() => {
  const i = stages.value.findIndex(s => s.status === 'running')
  if (i >= 0) return i
  // 如果都在 done，按最后一个 done
  const lastDone = stages.value.map(s => s.status).lastIndexOf('done')
  if (lastDone >= 0) return Math.min(lastDone, stages.value.length - 1)
  return 0
})

const currentPhaseName = computed(() => stages.value[currentPhaseIndex.value]?.name || '演练启动')

const currentPhaseProgress = computed(() => {
  const s = stages.value[currentPhaseIndex.value]
  if (!s) return { num: 0, den: 0 }
  return { num: s.completedSteps, den: s.totalSteps }
})

// 阶段环需要的 4 个相位（主名 + 内圈 6 个标签）
const ringPhases = computed(() => {
  return stages.value.map(s => s.name)
})

const ringSize = computed(() => {
  if (viewportWidth.value < 900) return 330
  if (viewportWidth.value < 1200) return 420
  return 560
})

// === 告警 ===
// 从步骤的"进行中"或异常中推算
// 格式化秒数为 HH:MM:SS
function fmtHMS(totalSec: number): string {
  const h = Math.floor(totalSec / 3600)
  const m = Math.floor((totalSec % 3600) / 60)
  const s = totalSec % 60
  return [h, m, s].map(n => String(n).padStart(2, '0')).join(':')
}

const activeAlerts = computed(() => {
  // 依赖 elapsedSeconds 使 computed 每秒重算，进度条/耗时实时刷新
  const _now = elapsedSeconds.value
  const nowMs = Date.now()

  const alerts: Array<{
    title: string
    team: string
    progress: number
    limit: string
    elapsed: string
    statusLabel: string
    level: 'warn' | 'info' | 'danger'
  }> = []

  // 进行中步骤（只看叶子步骤）
  leafSteps.value
    .filter(s => s.status === 'running')
    .slice(0, 2)
    .forEach(s => {
      const elapsedSec = s.start_time
        ? Math.max(1, Math.round((nowMs - new Date(s.start_time).getTime()) / 1000))
        : 1
      const limitMin = s.timeout_minutes || 30
      const limitSec = limitMin * 60
      const pct = Math.min(99, Math.round((elapsedSec / limitSec) * 100))
      alerts.push({
        title: s.executor_team ? `${s.name} - ${s.executor_team}` : s.name,
        team: s.executor_team || '运维部',
        progress: pct,
        limit: fmtHMS(limitSec),
        elapsed: fmtHMS(elapsedSec),
        statusLabel: pct >= 80 ? '即将超时' : '进行中',
        level: pct >= 80 ? 'danger' : 'warn',
      })
    })

  // 异常步骤（只看叶子步骤）
  leafSteps.value
    .filter(s => s.status === 'issue' || s.status === 'timeout')
    .slice(0, 2)
    .forEach(s => {
      const elapsedSec = s.start_time
        ? Math.max(0, Math.round((nowMs - new Date(s.start_time).getTime()) / 1000))
        : 0
      alerts.push({
        title: s.name,
        team: s.executor_team || '执行组',
        progress: 0,
        limit: fmtHMS((s.timeout_minutes || 30) * 60),
        elapsed: elapsedSec > 0 ? fmtHMS(elapsedSec) : '--',
        statusLabel: s.status === 'timeout' ? '已超时' : '异常',
        level: 'danger',
      })
    })

  if (alerts.length === 0) {
    // 无活跃告警时，展示待执行步骤作为预警
    leafSteps.value
      .filter(s => s.status === 'pending')
      .slice(0, 2)
      .forEach((s, i) => {
        const limitMin = s.timeout_minutes || 30
        alerts.push({
          title: s.name,
          team: s.executor_team || '运维部',
          progress: 0,
          limit: fmtHMS(limitMin * 60),
          elapsed: '待启动',
          statusLabel: '待执行',
          level: 'info',
        })
      })
  }
  return alerts.slice(0, 2)
})

// === 日志（已有数据） ===
function logActionClass(action: string): string {
  if (!action) return 'step'
  if (action.includes('issue') || action.includes('timeout')) return 'danger'
  if (action.includes('skip')) return 'skip'
  if (action.includes('force')) return 'force'
  return 'ok'
}
function logActionLabel(action: string): string {
  const map: Record<string, string> = {
    complete: '完成', step_complete: '完成',
    issue: '异常', step_issue: '异常',
    timeout: '超时', force_complete: '强制完成',
    skip: '跳过', step_skip: '跳过',
    start: '启动', step_start: '启动',
  }
  return map[action] || action
}

// 根据 log 中的 StepInstanceID 在 drillSteps 中查节点名称
function resolveStepName(log: StepInstanceLog): string {
  const id = log?.StepInstanceID
  if (!id) return '-'
  const step = drillSteps.value.find(s => s.id === id)
  if (step?.name) return step.name
  return `步骤 #${id}`
}
function stageBadgeLabel(status: string): string {
  const map: Record<string, string> = {
    done: '已完成', running: '进行中', pending: '待开始', issue: '异常',
  }
  return map[status] || status
}

function formatTime(dateStr: string | null | undefined): string {
  if (!dateStr) return '--:--:--'
  const d = new Date(dateStr)
  return [d.getHours(), d.getMinutes(), d.getSeconds()].map(n => String(n).padStart(2, '0')).join(':')
}
function formatHM(dateStr: string | null | undefined): string {
  if (!dateStr) return '--:--'
  const d = new Date(dateStr)
  return [d.getHours(), d.getMinutes()].map(n => String(n).padStart(2, '0')).join(':')
}
function formatSystemTime(d: Date): string {
  return `${d.getFullYear()}.${String(d.getMonth() + 1).padStart(2, '0')}.${String(d.getDate()).padStart(2, '0')} ${[d.getHours(), d.getMinutes(), d.getSeconds()].map(n => String(n).padStart(2, '0')).join(':')}`
}

// 计时：每秒钟刷新 systemTime / elapsed
function tick() {
  const now = new Date()
  systemTime.value = formatSystemTime(now)
  const started = drillStartTime.value
  if (started) {
    const start = new Date(started).getTime()
    if (!isNaN(start)) {
      const ended = (currentDrill.value as any)?.end_time || (currentDrill.value as any)?.completed_at
      const end = ended
        ? new Date(ended).getTime()
        : now.getTime()
      elapsedSeconds.value = Math.max(0, Math.round((end - start) / 1000))
    }
  }
}

// 数据加载
async function loadData() {
  if (!drillId.value) {
    error.value = '无效的演练 ID'
    loading.value = false
    return
  }
  try {
    const drill = await drillApi.getDetail(drillId.value)
    if (componentDestroyed) return
    currentDrill.value = drill

    const steps = await drillApi.getSteps(drillId.value)
    if (componentDestroyed) return
    drillSteps.value = steps.sort((a, b) => a.seq - b.seq)

    const logs = await drillApi.getLogs(drillId.value)
    if (componentDestroyed) return
    recentLogs.value = logs.slice(0, 30)

    loading.value = false
    error.value = null
    tick()
    // 仅在 WebSocket 未连接时建立连接，避免刷新数据时重连导致循环
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      connectWebSocket()
    }
  } catch (err: any) {
    if (componentDestroyed) return
    error.value = err.message || '加载数据失败'
    console.error('Failed to load drill data:', err)
    loading.value = false
  }
}
function handleRetry() {
  loadData()
}

// WebSocket
const REFRESH_EVENTS = new Set([
  'step_change', 'drill_status',
  'step_started', 'step_complete', 'step_issue', 'step_skipped', 'step_timeout',
  'drill_started', 'drill_paused', 'drill_resumed', 'drill_completed', 'drill_terminated',
  'timeout_warning', 'timeout_alert',
])
const LOG_EVENTS: Record<string, string> = {
  step_started: 'step_start',
  step_complete: 'step_complete',
  step_issue: 'step_issue',
  step_skipped: 'step_skip',
  step_timeout: 'step_timeout',
  drill_started: 'drill_started',
  drill_paused: 'drill_paused',
  drill_resumed: 'drill_resumed',
  drill_completed: 'drill_completed',
  drill_terminated: 'drill_terminated',
}

function connectWebSocket() {
  if (componentDestroyed) return
  if (ws) ws.close()
  const wsProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
  const authStore = useAuthStore()
  const wsUrl = `${wsProtocol}://${window.location.host}/ws/control/${drillId.value}?token=${authStore.token}`
  ws = new WebSocket(wsUrl)
  ws.onmessage = (event) => {
    if (componentDestroyed) return
    try {
      const data = JSON.parse(event.data)
      const eventType = data.type || data.event_type
      if (!eventType) return

      // 步骤/演练状态变化：全量刷新，确保级联状态（父步骤、阶段）正确更新
      if (eventType.startsWith('step_') || eventType.startsWith('drill_') || REFRESH_EVENTS.has(eventType)) {
        loadData()
      }
    } catch (e) { /* ignore */ }
  }
  ws.onerror = () => {
    if (componentDestroyed) return
    startFallbackPolling()
  }
  ws.onclose = () => {
    if (componentDestroyed) return
    if (currentDrill.value?.status === 'running') startFallbackPolling()
  }
}

// 增量应用步骤事件,避免每次都重拉 3 个 API
function applyStepEvent(eventType: string, payload: any) {
  if (!payload) return
  const stepId = Number(payload.step_id ?? payload.id)
  if (!stepId) {
    loadData()
    return
  }
  const idx = drillSteps.value.findIndex(s => s.id === stepId)
  if (idx === -1) {
    // 找不到对应步骤,降级为全量刷新
    loadData()
    return
  }
  const target = drillSteps.value[idx]
  const newStatus = payload.new_status || mapEventToStatus(eventType)
  if (newStatus) target.status = newStatus as StepStatus
  if (payload.start_time) target.start_time = payload.start_time
  if (payload.end_time) target.end_time = payload.end_time
  if (payload.timeout_at) target.timeout_at = payload.timeout_at
  if (payload.remark) target.remark = payload.remark
  if (payload.issue_desc) target.issue_desc = payload.issue_desc
  if (payload.assignee_names) target.assignee_names = payload.assignee_names

  // 推入一条本地日志
  const logAction = LOG_EVENTS[eventType] || eventType
  const newLog: StepInstanceLog = {
    ID: Date.now(),
    StepInstanceID: stepId,
    Action: logAction,
    OperatorID: 0,
    OperatorName: payload.executor || '流程引擎',
    Remark: payload.remark || payload.comment || payload.issue_desc || '',
    CreatedAt: new Date().toISOString(),
  }
  recentLogs.value = [newLog, ...recentLogs.value].slice(0, 30)

  // 重新计算 KPI
  recomputeKpis()
}

function applyDrillEvent(eventType: string, payload: any) {
  if (!payload) return
  if (currentDrill.value) {
    const newStatus = payload.new_status || mapEventToStatus(eventType)
    if (newStatus) currentDrill.value.status = newStatus as DrillStatus
  }
  const newLog: StepInstanceLog = {
    ID: Date.now(),
    StepInstanceID: 0,
    Action: LOG_EVENTS[eventType] || eventType,
    OperatorID: 0,
    OperatorName: payload.operator || '流程引擎',
    Remark: payload.remark || '',
    CreatedAt: new Date().toISOString(),
  }
  recentLogs.value = [newLog, ...recentLogs.value].slice(0, 30)
  recomputeKpis()
}

function mapEventToStatus(eventType: string): string {
  const map: Record<string, string> = {
    step_started: 'running',
    step_complete: 'completed',
    step_issue: 'issue',
    step_skipped: 'skipped',
    step_timeout: 'timeout',
    drill_started: 'running',
    drill_paused: 'paused',
    drill_resumed: 'running',
    drill_completed: 'completed',
    drill_terminated: 'terminated',
  }
  return map[eventType] || ''
}

function recomputeKpis() {
  // 触发 drillSteps / currentDrill 的依赖计算
  // Vue 的响应式系统会自动重算
  tick()
}

function startFallbackPolling() {
  if (componentDestroyed) return
  if (refreshTimer) clearInterval(refreshTimer)
  refreshTimer = window.setInterval(() => {
    if (componentDestroyed) {
      stopFallbackPolling()
      return
    }
    if (currentDrill.value?.status === 'running') loadData()
  }, 5000)
}

function stopFallbackPolling() {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

// 全屏
function toggleFullscreen() {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen?.()
  } else {
    document.exitFullscreen?.()
  }
}

onMounted(() => {
  componentDestroyed = false
  loadData()
  window.addEventListener('resize', handleResize)
  timeTimer = window.setInterval(tick, 1000)
})
onBeforeUnmount(() => {
  componentDestroyed = true
  window.removeEventListener('resize', handleResize)
  if (timeTimer) clearInterval(timeTimer)
  stopFallbackPolling()
  if (ws) { ws.close(); ws = null }
})

function handleResize() {
  viewportWidth.value = window.innerWidth
}
</script>

<style lang="scss" scoped>
@use '@/styles/variables' as *;

// ===== 主题色 =====
$bg-deep: #040a1a;
$bg-mid: #061029;
$bg-card: rgba(8, 24, 56, 0.72);
$line: rgba(0, 212, 255, 0.18);
$line-strong: rgba(0, 212, 255, 0.5);
$neon: #00d4ff;
$neon-dim: #00a0c8;
$neon-soft: rgba(0, 212, 255, 0.15);
$ok: #00ff9c;
$warn: #ffb648;
$danger: #ff4d6a;
$text: #d6e8ff;
$text-dim: #6e8db5;
$text-mute: #4a6b91;

$font-display: 'Orbitron', 'Rajdhani', 'Microsoft YaHei', sans-serif;
$font-mono: 'Share Tech Mono', 'Consolas', monospace;
$font-cn: 'Microsoft YaHei', 'PingFang SC', 'Hiragino Sans GB', sans-serif;

.screen-root {
  position: relative;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  background:
    radial-gradient(circle at 50% 48%, rgba(0, 76, 180, 0.24), transparent 32%),
    radial-gradient(circle at 74% 68%, rgba(255, 122, 0, 0.12), transparent 18%),
    linear-gradient(180deg, #071226 0%, #020713 100%);
  color: $text;
  font-family: $font-cn;
  user-select: none;

  &::before {
    content: '';
    position: absolute;
    inset: 8px 12px 8px;
    border: 1px solid rgba(134, 181, 255, 0.45);
    box-shadow:
      inset 0 0 42px rgba(38, 118, 255, 0.1),
      0 0 28px rgba(38, 118, 255, 0.08);
    pointer-events: none;
    z-index: 1;
  }

  &::after {
    content: '';
    position: absolute;
    inset: 0;
    background-image:
      linear-gradient(30deg, rgba(93, 151, 240, 0.08) 12%, transparent 12.5%, transparent 87%, rgba(93, 151, 240, 0.08) 87.5%, rgba(93, 151, 240, 0.08)),
      linear-gradient(150deg, rgba(93, 151, 240, 0.08) 12%, transparent 12.5%, transparent 87%, rgba(93, 151, 240, 0.08) 87.5%, rgba(93, 151, 240, 0.08)),
      linear-gradient(30deg, rgba(93, 151, 240, 0.08) 12%, transparent 12.5%, transparent 87%, rgba(93, 151, 240, 0.08) 87.5%, rgba(93, 151, 240, 0.08)),
      linear-gradient(150deg, rgba(93, 151, 240, 0.08) 12%, transparent 12.5%, transparent 87%, rgba(93, 151, 240, 0.08) 87.5%, rgba(93, 151, 240, 0.08)),
      linear-gradient(60deg, rgba(40, 99, 180, 0.08) 25%, transparent 25.5%, transparent 75%, rgba(40, 99, 180, 0.08) 75%, rgba(40, 99, 180, 0.08)),
      linear-gradient(60deg, rgba(40, 99, 180, 0.08) 25%, transparent 25.5%, transparent 75%, rgba(40, 99, 180, 0.08) 75%, rgba(40, 99, 180, 0.08));
    background-position: 0 0, 0 0, 18px 32px, 18px 32px, 0 0, 18px 32px;
    background-size: 36px 64px;
    opacity: 0.24;
    mask-image: radial-gradient(circle at center, #000 0%, transparent 82%);
    pointer-events: none;
  }
}

// ===== 背景层 =====
.bg-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(64, 141, 255, 0.08) 1px, transparent 1px),
    linear-gradient(90deg, rgba(64, 141, 255, 0.08) 1px, transparent 1px);
  background-size: 64px 64px;
  transform: perspective(560px) rotateX(58deg) translateY(-120px) scale(1.2);
  transform-origin: 50% 0;
  mask-image: linear-gradient(180deg, transparent, #000 18%, transparent 92%);
  pointer-events: none;
  z-index: 0;
}
.bg-scan {
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(
    0deg,
    transparent 0,
    transparent 2px,
    rgba(0, 212, 255, 0.02) 3px,
    transparent 4px
  );
  pointer-events: none;
  z-index: 0;
  animation: scan 8s linear infinite;
}
@keyframes scan {
  from { background-position-y: 0; }
  to { background-position-y: 100px; }
}
.bg-vignette {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse at top, rgba(0, 212, 255, 0.06), transparent 60%),
    radial-gradient(ellipse at bottom, rgba(0, 80, 160, 0.1), transparent 70%);
  pointer-events: none;
  z-index: 0;
}

// ===== Overlay =====
.overlay-state {
  position: fixed; inset: 0;
  display: flex; align-items: center; justify-content: center;
  z-index: 100; background: $bg-deep;
  &.error .error-content {
    display: flex; flex-direction: column; align-items: center; gap: $spacing-base;
    color: $text-dim;
    .el-icon { color: $danger; }
    p { font-size: $font-size-sm; }
  }
}
.loader {
  text-align: center;
  .loader-ring {
    width: 48px; height: 48px; margin: 0 auto $spacing-base;
    border: 2px solid $line; border-top-color: $neon;
    border-radius: 50%; animation: spin 1s linear infinite;
  }
  .loader-text {
    color: $text-dim; font-size: $font-size-xs;
    letter-spacing: 3px; text-transform: uppercase;
    font-family: $font-display;
  }
}
@keyframes spin { to { transform: rotate(360deg); } }

// ===== 屏幕主容器 =====
.screen-content {
  position: relative;
  z-index: 2;
  display: flex;
  flex-direction: column;
  height: 100vh;
  padding: 12px 20px 4px;
  gap: 8px;
}

// ===== HEADER =====
.screen-header {
  position: relative;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto auto;
  align-items: center;
  column-gap: 16px;
  height: 46px;
  background:
    linear-gradient(90deg, rgba(37, 130, 255, 0.26), rgba(13, 37, 74, 0.1) 44%, rgba(20, 50, 96, 0.35));
  border: 0;
  padding: 0 28px 0 18px;
  box-shadow: inset 0 -1px 0 rgba(111, 178, 255, 0.35);

  &::before, &::after {
    content: '';
    position: absolute;
    top: 50%;
    width: 50px; height: 1px;
    background: linear-gradient(90deg, transparent, $neon);
  }
  &::before { left: 0; }
  &::after { right: 0; transform: rotate(180deg); }

  .header-title-block {
    grid-column: 1;
    text-align: left;
    display: flex;
    align-items: baseline;
    gap: 18px;
    .drill-title {
      font-family: $font-cn;
      font-size: clamp(24px, 2.35vw, 42px);
      font-weight: 900;
      letter-spacing: 4px;
      margin: 0;
      color: #ffffff;
      text-shadow:
        0 0 10px rgba(0, 153, 255, 0.95),
        0 0 24px rgba(0, 153, 255, 0.6);
      white-space: nowrap;
    }
    .drill-title-en {
      display: block;
      margin-top: 0;
      font-family: $font-display;
      font-size: clamp(9px, 1vw, 15px);
      font-weight: 700;
      letter-spacing: 7px;
      color: rgba(194, 214, 255, 0.66);
      white-space: nowrap;
    }
  }
  .header-meta {
    grid-column: 2;
    justify-self: end;
    display: flex; align-items: center; gap: 10px;
    font-family: $font-mono;
    .meta-label { color: $text-dim; font-size: 12px; letter-spacing: 2px; }
    .meta-divider { color: $text-mute; }
    .meta-value {
      color: #ecf6ff;
      font-size: clamp(13px, 1.1vw, 18px);
      font-weight: 700;
      text-shadow: 0 0 8px rgba(95, 171, 255, 0.6);
      letter-spacing: 1px;
    }
  }
  .btn-icon {
    grid-column: 3;
    position: static;
    background: transparent; border: 1px solid $line;
    color: $neon; width: 28px; height: 28px;
    display: flex; align-items: center; justify-content: center;
    cursor: pointer; border-radius: 2px;
    transition: all 0.2s;
    &:hover { border-color: $neon; box-shadow: 0 0 10px $neon-soft; }
  }
  .header-deco {
    position: absolute; top: 0; width: 8px; height: 100%;
    &-left { left: 0; background: linear-gradient(180deg, transparent, $neon, transparent); opacity: 0.7; }
    &-right { right: 0; background: linear-gradient(180deg, transparent, $neon, transparent); opacity: 0.7; }
  }
}

// ===== KPI Row =====
.kpi-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  height: 88px;
  flex-shrink: 0;
  padding: 0;
}
.kpi-card {
  position: relative;
  background:
    linear-gradient(90deg, rgba(14, 50, 112, 0.82), rgba(10, 26, 57, 0.36)),
    linear-gradient(180deg, rgba(53, 138, 255, 0.18), transparent);
  border: 1px solid rgba(105, 165, 255, 0.38);
  padding: 12px 22px 10px 80px;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  grid-template-rows: auto 1fr auto;
  column-gap: 12px;
  align-items: center;
  overflow: hidden;
  transition: all 0.3s;
  clip-path: polygon(0 0, calc(100% - 18px) 0, 100% 18px, 100% 100%, 18px 100%, 0 calc(100% - 18px));

  // 角装饰
  &::before, &::after {
    content: '';
    position: absolute; width: 10px; height: 10px;
    border: 1px solid $neon;
  }
  &::before { top: -1px; left: -1px; border-right: 0; border-bottom: 0; }
  &::after { bottom: -1px; right: -1px; border-left: 0; border-top: 0; }

  .kpi-orb {
    position: absolute;
    left: 24px;
    top: 18px;
    width: 38px;
    height: 38px;
    border-radius: 50%;
    background:
      radial-gradient(circle at 36% 34%, #9ffcff 0 14%, #00d4ff 15% 28%, rgba(20, 255, 189, 0.85) 29% 45%, rgba(0, 74, 165, 0.7) 46% 100%);
    box-shadow:
      0 0 14px rgba(0, 212, 255, 0.85),
      0 0 36px rgba(0, 212, 255, 0.22);

    &::before {
      content: '';
      position: absolute;
      inset: 8px -11px;
      border: 1px solid rgba(0, 212, 255, 0.64);
      border-radius: 50%;
      transform: rotate(-8deg);
    }

    &::after {
      content: '';
      position: absolute;
      left: 13px;
      bottom: -20px;
      width: 18px;
      height: 18px;
      border-radius: 50%;
      border: 2px solid rgba(0, 212, 255, 0.54);
      box-shadow: 0 0 10px rgba(0, 212, 255, 0.55);
    }
  }
  .kpi-copy {
    grid-column: 1;
    grid-row: 1 / 3;
    display: flex;
    align-items: center;
  }
  .kpi-label-zh {
    display: block;
    position: relative;
    font-size: clamp(13px, 1vw, 18px);
    font-weight: 800;
    color: #f0f7ff;
    letter-spacing: 1px;
    padding-bottom: 6px;
    &::after {
      content: '';
      position: absolute;
      left: 0;
      bottom: 0;
      width: 24px;
      height: 2px;
      border-radius: 1px;
      background: linear-gradient(90deg, #2cf8d8, transparent);
    }
  }
  .kpi-label-en {
    font-family: $font-display;
    display: block;
    margin-top: 6px;
    font-size: 10px;
    font-weight: 600;
    color: rgba(196, 214, 255, 0.55);
    letter-spacing: 4px;
    opacity: 1;
  }
  .kpi-value-row {
    grid-column: 2;
    grid-row: 1 / 3;
    justify-self: end;
    margin-top: 0;
    display: flex; align-items: baseline; gap: 4px;
    &.mono { gap: 2px; }
  }
  .kpi-status-row {
    align-items: center;
  }
  .kpi-progress-row {
    align-items: center;
  }
  .kpi-value-num {
    font-family: $font-mono;
    font-size: clamp(24px, 2vw, 38px);
    font-weight: 800;
    color: #2cf8d8;
    text-shadow: 0 0 12px rgba(44, 248, 216, 0.42);
    line-height: 1;
  }
  .kpi-value-unit {
    font-family: $font-mono; font-size: 18px; color: #c8fff5; opacity: 0.85;
  }
  .kpi-value-sep {
    font-family: $font-mono; font-size: clamp(18px, 1.7vw, 28px); color: #2cf8d8; opacity: 0.75;
    transform: translateY(-2px);
  }
  .kpi-value-text {
    font-size: clamp(18px, 1.6vw, 27px); font-weight: 800; color: #ff7a00;
    text-shadow: 0 0 12px rgba(255, 122, 0, 0.5);
  }
  .kpi-sub {
    grid-column: 1 / -1;
    grid-row: 3;
    font-size: 11px; color: rgba(255, 122, 0, 0.78); margin-top: 2px;
    font-family: $font-mono; letter-spacing: 1px;
  }
  .kpi-node-count {
    margin-left: 10px;
    color: #dff6ff;
    font-family: $font-mono;
    font-size: 14px;
  }

  // === Progress card specific: ring + node count layout ===
  &.kpi-progress-card {
    .kpi-value-row.kpi-progress-row {
      align-items: center;
      gap: 12px;
    }
  }
  .progress-ring-wrap {
    position: relative;
    width: 56px;
    height: 56px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .progress-ring-svg {
    width: 100%;
    height: 100%;
    transform: rotate(-90deg);
  }
  .progress-ring-bg {
    fill: none;
    stroke: rgba(255, 255, 255, 0.12);
    stroke-width: 6;
  }
  .progress-ring-fill {
    fill: none;
    stroke: url(#ring-grad);
    stroke-width: 6;
    stroke-linecap: round;
    stroke-dasharray: 289.03;
    transition: stroke-dashoffset 0.8s ease;
  }
  .progress-ring-text {
    position: absolute;
    font-family: $font-mono;
    font-size: 15px;
    font-weight: 800;
    color: #fff;
    line-height: 1;
    small {
      font-size: 10px;
      font-weight: 600;
      opacity: 0.75;
    }
  }
  .progress-node-block {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .node-count-row {
    display: flex;
    align-items: baseline;
    gap: 3px;
    .node-separator {
      font-family: $font-mono;
      font-size: 14px;
      color: rgba(255, 255, 255, 0.45);
    }
    .node-total {
      font-family: $font-mono;
      font-size: 16px;
      color: rgba(255, 255, 255, 0.55);
      font-weight: 600;
    }
  }
  .progress-sub-label {
    font-size: 11px;
    color: rgba(255, 255, 255, 0.5);
    font-family: $font-mono;
    letter-spacing: 1px;
  }
  .status-dot {
    width: 10px; height: 10px; border-radius: 50%;
    background: $ok; box-shadow: 0 0 10px $ok;
    animation: pulse 1.6s ease-in-out infinite;
    &.dot-paused { background: $warn; box-shadow: 0 0 10px $warn; }
    &.dot-completed { background: $neon; box-shadow: 0 0 10px $neon; }
    &.dot-terminated { background: $danger; box-shadow: 0 0 10px $danger; animation: none; }
    &.dot-pending { background: $text-mute; box-shadow: none; animation: none; }
  }
  .txt-running { color: $ok; }
  .txt-paused { color: $warn; }
  .txt-completed { color: $neon; }
  .txt-terminated { color: $danger; }
  .txt-pending { color: $text-dim; }
}
@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.5; transform: scale(0.85); }
}

// ===== MAIN GRID =====
.screen-main {
  display: grid;
  grid-template-columns: minmax(180px, 18vw) minmax(340px, 1fr) minmax(240px, 24vw);
  gap: 20px;
  flex: 1;
  min-height: 0;
  padding: 0;
}

.panel {
  position: relative;
  display: flex; flex-direction: column;
  background: $bg-card;
  border: 0;
  backdrop-filter: blur(2px);
  overflow: hidden;
}
.panel-header {
  position: relative;
  display: flex; align-items: center; gap: 8px;
  height: 34px;
  padding: 0 12px 0 24px;
  background:
    linear-gradient(90deg, rgba(0, 116, 255, 0.62), rgba(8, 29, 67, 0.38) 52%, transparent 100%);
  border-bottom: 0;
  flex-shrink: 0;
  .panel-deco-corner {
    position: absolute; width: 8px; height: 8px; border-color: $neon; border-style: solid; border-width: 0;
    &.tl { top: 0; left: 0; border-top-width: 2px; border-left-width: 2px; }
    &.tr { top: 0; right: 0; border-top-width: 2px; border-right-width: 2px; }
  }
  .panel-title-zh {
    font-size: clamp(14px, 1.2vw, 24px);
    font-weight: 900;
    color: #ffffff;
    letter-spacing: 2px;
    text-shadow: 0 0 10px rgba(64, 170, 255, 0.8);
  }
  .panel-title-en {
    font-family: $font-display;
    font-size: 10px; color: $neon; opacity: 0.7;
    letter-spacing: 2px;
  }
  .panel-badge {
    margin-left: auto;
    background: $neon-soft; border: 1px solid $neon;
    color: $neon; font-family: $font-mono;
    font-size: 11px; padding: 0 6px; height: 18px;
    display: inline-flex; align-items: center;
  }
  .panel-realtime {
    margin-left: auto;
    display: inline-flex; align-items: center; gap: 4px;
    font-size: 11px; color: $ok; font-family: $font-display;
    letter-spacing: 1px;
    .rt-dot {
      width: 6px; height: 6px; border-radius: 50%;
      background: $ok; box-shadow: 0 0 6px $ok;
      animation: pulse 1.4s ease-in-out infinite;
    }
  }
}
.panel-body {
  flex: 1;
  overflow-y: auto;
  padding: 10px 12px;
  min-height: 0;
  &::-webkit-scrollbar { width: 4px; }
  &::-webkit-scrollbar-track { background: transparent; }
  &::-webkit-scrollbar-thumb { background: $line-strong; border-radius: 2px; }
}
.empty-tip {
  text-align: center; color: $text-mute;
  font-size: 12px; padding: 30px 0;
  font-family: $font-mono;
}

// ===== Left stages =====
.panel-stages {
  background: transparent;
  min-height: 0;
  overflow: hidden;
  .stages-list {
    display: flex; flex-direction: column; gap: 14px;
    height: 100%;
  }
  .stage-card {
    position: relative;
    background: linear-gradient(90deg, rgba(11, 21, 38, 0.88), rgba(10, 20, 40, 0.42));
    border: 1px solid rgba(70, 96, 130, 0.28);
    border-left: 3px solid rgba(160, 176, 196, 0.48);
    padding: 8px 12px 9px;
    flex: 1;
    display: flex; flex-direction: column; gap: 4px;
    justify-content: center;
    transition: all 0.2s;

    &.stage-done { border-left-color: $neon; }
    &.stage-running {
      background: linear-gradient(90deg, rgba(74, 35, 6, 0.88), rgba(19, 17, 22, 0.54));
      border-color: #ff9a2f;
      border-left-color: #ff7a00;
      box-shadow: 0 0 16px rgba(255, 122, 0, 0.24), inset 0 0 16px rgba(255, 122, 0, 0.08);
    }
    &.stage-issue { border-left-color: $danger; }
  }
  .stage-card-top {
    display: flex; align-items: center; justify-content: space-between;
    .stage-name-block { display: flex; align-items: baseline; gap: 8px; }
    .stage-index {
      font-family: $font-mono; font-size: 14px;
      color: #ffffff; font-weight: 900;
      text-shadow: 0 0 8px rgba(255, 255, 255, 0.42);
    }
    .stage-name {
      font-size: 13px; color: $text; font-weight: 700;
    }
  }
  .stage-badge {
    font-size: 11px; padding: 2px 8px;
    border-radius: 1px; letter-spacing: 1px;
    &.badge-done { background: rgba(0, 212, 255, 0.15); color: $neon; border: 1px solid $line-strong; }
    &.badge-running { background: rgba(255, 122, 0, 0.14); color: #ffc179; border: 1px solid rgba(255, 122, 0, 0.62); }
    &.badge-pending { background: rgba(110, 141, 181, 0.15); color: $text-dim; border: 1px solid $line; }
    &.badge-issue { background: rgba(255, 77, 106, 0.18); color: $danger; border: 1px solid rgba(255, 77, 106, 0.4); }
  }
  .stage-card-mid {
    display: flex; align-items: center; justify-content: space-between;
    .stage-time { font-family: $font-mono; font-size: 13px; color: $text; font-weight: 600; }
    .stage-meta-label { font-size: 10px; color: $text-mute; letter-spacing: 1px; }
  }
  .stage-segments {
    display: flex; gap: 2px; height: 6px;
    .segment { flex: 1; background: rgba(255, 255, 255, 0.05); }
    .segment.seg-done { background: linear-gradient(180deg, $neon, $neon-dim); box-shadow: 0 0 4px rgba(0, 212, 255, 0.4); }
    .segment.seg-active { background: linear-gradient(180deg, #ff9a2f, #ff7a00); box-shadow: 0 0 6px #ff7a00; animation: pulse 1.2s ease-in-out infinite; }
    .segment.seg-todo { background: rgba(110, 141, 181, 0.25); }
    .segment.seg-empty { background: transparent; border: 1px dashed rgba(110, 141, 181, 0.2); }
  }
  .stage-card-bottom {
    display: flex; justify-content: space-between;
    .stage-meta {
      display: flex; flex-direction: column; gap: 1px;
      .meta-key { font-size: 10px; color: $text-mute; letter-spacing: 1px; }
      .meta-val { font-family: $font-mono; font-size: 12px; color: $text; font-weight: 600; }
    }
  }
}

// ===== Center phase ring =====
.panel-center {
  background:
    radial-gradient(circle at center, rgba(18, 92, 210, 0.22), transparent 49%),
    radial-gradient(circle at center, rgba(4, 18, 49, 0.38), transparent 72%);
  position: relative;
  overflow: visible;
  &::before, &::after {
    content: '';
    position: absolute; width: 14px; height: 14px;
    border: 1px solid $neon;
  }
  &::before { top: -1px; left: -1px; border-right: 0; border-bottom: 0; }
  &::after { bottom: -1px; right: -1px; border-left: 0; border-top: 0; }
}
.center-stage {
  flex: 1;
  display: flex; align-items: center; justify-content: center;
  position: relative;
  &::before, &::after {
    content: '';
    position: absolute; left: 50%;
    transform: translateX(-50%);
    width: 86%; height: 1px;
    background: linear-gradient(90deg, transparent, rgba(87, 152, 255, 0.44), transparent);
  }
  &::before { top: 18%; }
  &::after { bottom: 18%; }
}

// ===== Right column =====
.panel-right {
  display: flex; flex-direction: column;
  gap: 8px;
  background: transparent; border: none; padding: 0;
  min-height: 0;
  .sub-panel {
    position: relative;
    background: transparent; border: 0;
    display: flex; flex-direction: column;
    overflow: hidden;
    min-height: 0;
  }
  .sub-warn { flex: 0 0 auto; }
  .sub-logs { flex: 1; min-height: 0; }
}

// ===== Alerts =====
.warn-list {
  display: flex; flex-direction: column; gap: 8px;
  .alert-card {
    position: relative;
    background: linear-gradient(180deg, rgba(20, 22, 35, 0.86), rgba(16, 26, 47, 0.72));
    border: 1px solid rgba(130, 150, 180, 0.3);
    border-left: 3px solid $warn;
    padding: 10px 12px 12px;
    display: flex; flex-direction: column; gap: 6px;
    &.alert-danger { border-left-color: $danger; }
    &.alert-info { border-left-color: $neon; }
  }
  .alert-head {
    display: flex; align-items: center; gap: 6px;
    .alert-arrow { color: $warn; font-size: 10px; }
    .alert-title {
      flex: 1;
      font-size: 13px; color: #ffffff; font-weight: 800;
      white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
    }
    .alert-team {
      font-family: $font-mono; font-size: 10px;
      color: $neon; background: $neon-soft;
      padding: 1px 5px; letter-spacing: 1px;
    }
  }
  .alert-bar {
    height: 4px; background: rgba(255, 255, 255, 0.06);
    position: relative; overflow: hidden;
    .alert-bar-fill {
      position: absolute; top: 0; left: 0; bottom: 0;
      background: linear-gradient(90deg, $warn, $neon);
      box-shadow: 0 0 6px rgba(255, 182, 72, 0.5);
    }
  }
  .alert-foot {
    display: flex; justify-content: space-between;
    .alert-meta { font-family: $font-mono; font-size: 10px; color: $text-dim; }
  }
  .alert-status {
    display: flex; justify-content: space-between; align-items: center;
    border-top: 1px dashed $line; padding-top: 5px;
    .alert-progress-pct {
      font-family: $font-mono; font-size: 14px; color: $warn; font-weight: 700;
      text-shadow: 0 0 6px rgba(255, 182, 72, 0.3);
    }
    .alert-status-tag {
      font-size: 10px; padding: 1px 6px; color: $ok;
      border: 1px solid rgba(0, 255, 156, 0.3); background: rgba(0, 255, 156, 0.08);
      letter-spacing: 1px;
    }
  }
  .alert-danger .alert-progress-pct { color: $danger; }
  .alert-danger .alert-status-tag { color: $danger; border-color: rgba(255, 77, 106, 0.3); background: rgba(255, 77, 106, 0.08); }
  .alert-info .alert-status-tag { color: $neon; border-color: rgba(0, 212, 255, 0.3); background: rgba(0, 212, 255, 0.08); }
}

// ===== Logs =====
.logs-wrap {
  display: flex; flex-direction: column; padding: 0;
}
.log-thead {
  display: grid; grid-template-columns: 72px 1.8fr 0.6fr;
  padding: 6px 12px;
  background: rgba(0, 60, 130, 0.42);
  border-bottom: 1px solid $line;
  font-family: $font-display; font-size: 10px;
  color: $neon; letter-spacing: 1.5px;
  .col-step { text-align: left; padding-left: 6px; }
}
.log-tbody {
  flex: 1; overflow-y: auto; padding: 4px 0;
  &::-webkit-scrollbar { width: 4px; }
  &::-webkit-scrollbar-track { background: transparent; }
  &::-webkit-scrollbar-thumb { background: $line-strong; border-radius: 2px; }
}
.log-row {
  display: grid; grid-template-columns: 72px 1.8fr 0.6fr;
  align-items: center;
  padding: 7px 12px;
  border-bottom: 1px solid rgba(111, 151, 220, 0.16);
  font-size: 12px;
  .col-time { font-family: $font-mono; color: $text-dim; font-size: 11px; }
  .col-step {
    color: $text;
    white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
    padding-left: 6px;
  }
  .col-desc {
    color: $text;
    white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
  }
  &.log-danger { .col-desc { color: $danger; } }
  &.log-skip { .col-desc { color: #a78bfa; } }
  &.log-force { .col-desc { color: $neon; } }
  &.log-ok { .col-desc { color: $ok; } }
}
// ===== Footer =====
.screen-footer {
  display: flex; justify-content: space-between;
  padding: 0;
  height: 16px;
  flex-shrink: 0;
  font-family: $font-display; font-size: 10px;
  color: $text-mute; letter-spacing: 2px;
}

@media (max-width: 1000px) {
  .screen-content {
    padding: 10px 12px 4px;
    gap: 6px;
  }

  .screen-root::before {
    inset: 6px 8px 6px;
  }

  .screen-header {
    height: 44px;
    padding: 0 12px;

    .header-title-block {
      gap: 8px;

      .drill-title {
        font-size: 22px;
        letter-spacing: 2px;
      }

      .drill-title-en {
        font-size: 8px;
        letter-spacing: 4px;
        max-width: 260px;
        overflow: hidden;
      }
    }

    .header-meta {
      gap: 5px;

      .meta-label,
      .meta-divider {
        display: none;
      }

      .meta-value {
        font-size: 12px;
      }
    }

    .btn-icon {
      width: 26px; height: 26px;
    }
  }

  .kpi-row {
    gap: 6px;
    height: 76px;
    padding: 0;
  }

  .kpi-card {
    padding: 10px 8px 8px 54px;
    column-gap: 4px;

    .kpi-orb {
      left: 14px;
      top: 20px;
      width: 28px;
      height: 28px;

      &::before {
        inset: 5px -7px;
      }

      &::after {
        left: 9px;
        bottom: -13px;
        width: 10px;
        height: 10px;
      }
    }

    .kpi-label-zh {
      font-size: 11px;
      letter-spacing: 0;
      line-height: 1.2;
      padding-bottom: 4px;
      &::after {
        width: 16px;
        height: 1.5px;
      }
    }

    .kpi-label-en {
      font-size: 7px;
      letter-spacing: 2px;
      line-height: 1.2;
    }

    .kpi-value-num {
      font-size: 18px;
    }

    .kpi-value-sep {
      font-size: 14px;
    }

    .kpi-value-unit {
      font-size: 11px;
    }

    .kpi-value-text {
      font-size: 15px;
      white-space: nowrap;
    }

    .kpi-sub {
      font-size: 7px;
      letter-spacing: 0;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .kpi-node-count {
      margin-left: 4px;
      font-size: 8px;
      white-space: nowrap;
    }

    .progress-ring-wrap {
      width: 40px;
      height: 40px;
    }
    .progress-ring-text {
      font-size: 11px;
      small {
        font-size: 8px;
      }
    }
    .progress-sub-label {
      font-size: 8px;
    }
    .node-count-row {
      .kpi-value-num {
        font-size: 16px;
      }
      .node-total {
        font-size: 12px;
      }
      .node-separator {
        font-size: 11px;
      }
    }
  }

  .screen-main {
    grid-template-columns: 190px minmax(260px, 1fr) 190px;
    gap: 10px;
    padding: 0;
  }

  .panel-header {
    height: 34px;
    padding: 0 8px 0 18px;

    .panel-title-zh {
      font-size: 14px;
      letter-spacing: 1px;
    }

    .panel-title-en {
      font-size: 7px;
      letter-spacing: 2px;
    }
  }

  .panel-body {
    padding: 8px;
  }

  .panel-stages {
    .stages-list {
      gap: 8px;
    }

    .stage-card {
      padding: 6px 8px 7px;
      flex: 1;
    }

    .stage-card-top {
      .stage-name-block {
        gap: 5px;
      }

      .stage-index {
        font-size: 12px;
      }

      .stage-name {
        font-size: 10px;
        max-width: 86px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }

    .stage-badge {
      font-size: 9px;
      padding: 1px 5px;
    }

    .stage-card-mid {
      .stage-time {
        font-size: 10px;
      }

      .stage-meta-label {
        font-size: 8px;
      }
    }

    .stage-card-bottom {
      .stage-meta {
        .meta-key {
          font-size: 8px;
        }

        .meta-val {
          font-size: 10px;
        }
      }
    }
  }

  .warn-list {
    .alert-card {
      padding: 8px;
    }

    .alert-head {
      .alert-title {
        font-size: 10px;
      }

      .alert-team {
        font-size: 8px;
      }
    }

    .alert-foot {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 3px 6px;

      .alert-meta {
        font-size: 8px;
      }
    }
  }

  .log-thead,
  .log-row {
    grid-template-columns: 62px 1fr 52px 1fr;
    padding-left: 8px;
    padding-right: 8px;
  }

  .log-row {
    font-size: 10px;

    .col-time {
      font-size: 9px;
    }
  }

  .screen-footer {
    font-size: 8px;
    letter-spacing: 1px;
    padding: 0;
  }
}
</style>
