<template>
  <div class="screen-root">
    <!-- 加载 -->
    <div v-if="loading" class="overlay-state">
      <div class="loader-ring" />
      <p class="loader-text">正在加载演练数据...</p>
    </div>

    <!-- 错误 -->
    <div v-else-if="error" class="overlay-state error">
      <p>{{ error }}</p>
      <button class="btn-retry" @click="handleRetry">重试</button>
    </div>

    <!-- 主屏 -->
    <template v-else>
      <div class="bg-grid" />
      <div class="bg-glow bg-glow-tl" />
      <div class="bg-glow bg-glow-br" />

      <!-- 顶部信息栏 -->
      <header class="top-bar">
        <div class="tb-line" />
        <div class="tb-left">
          <span class="elapsed-time">{{ displayTime }}</span>
        </div>
        <div class="tb-center">
          <div class="tb-title-row">
            <h1 class="tb-title">{{ instance?.name || '灾备演练' }}</h1>
            <span class="tb-status" :class="'st-' + (instance?.status || 'pending')">{{ statusLabel }}</span>
          </div>
        </div>
        <div class="tb-right">
          <div class="progress-wrap">
            <div class="progress-track">
              <div class="progress-fill" :style="{ width: (instance?.progress_pct || 0) + '%' }" />
              <div class="progress-glow" :style="{ left: (instance?.progress_pct || 0) + '%' }" />
            </div>
            <span class="progress-text">{{ instance?.progress_pct || 0 }}%</span>
          </div>
          <button class="btn-fullscreen" @click="toggleFullscreen" title="全屏模式">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M8 3H5a2 2 0 00-2 2v3m18 0V5a2 2 0 00-2-2h-3m0 18h3a2 2 0 002-2v-3M3 16v3a2 2 0 002 2h3"/>
            </svg>
          </button>
          <template v-if="canControl">
            <button v-if="instance?.status === 'pending'" class="btn-ctrl start" @click="handleStart" title="开始演练">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M8 5v14l11-7z"/></svg>
            </button>
            <button v-if="instance?.status === 'running'" class="btn-ctrl pause" @click="handlePause" title="暂停">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><rect x="6" y="4" width="4" height="16"/><rect x="14" y="4" width="4" height="16"/></svg>
            </button>
            <button v-if="instance?.status === 'paused'" class="btn-ctrl resume" @click="handleResume" title="恢复">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><path d="M8 5v14l11-7z"/></svg>
            </button>
            <button v-if="instance?.status === 'running' || instance?.status === 'paused'" class="btn-ctrl end" @click="handleTerminate" title="结束">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><rect x="6" y="6" width="12" height="12" rx="1"/></svg>
            </button>
</template>
        </div>
      </header>

      <!-- 主体 -->
      <main class="main-body">
        <!-- 流程树区域 -->
        <section class="flow-area">
          <div class="flow-label">演练流程</div>
          <canvas ref="flowCanvasRef" class="flow-canvas" />
          <div class="flow-legend">
            <span class="lg-item"><span class="lg-dot lg-done" />已完成</span>
            <span class="lg-item"><span class="lg-dot lg-running" />执行中</span>
            <span class="lg-item"><span class="lg-dot lg-pending" />待执行</span>
          </div>
        </section>

        <!-- 右侧面板 -->
        <aside class="right-panel">
          <!-- 待办 & 控制 -->
          <div class="rp-block rp-tasks">
            <div class="rp-hd">
              <svg class="rp-ico" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="5" width="18" height="14" rx="2"/><path d="M12 9v4"/><circle cx="12" cy="15" r="1"/></svg>
              <span>待办任务</span>
              <span class="rp-badge">{{ runningSteps.length }}</span>
            </div>
            <div class="rp-body">
              <div v-if="runningSteps.length === 0" class="rp-empty">无进行中的任务</div>
              <div v-for="s in runningSteps.slice(0, 5)" :key="s.id" class="task-row" :class="[{ 'task-timeout': s.status === 'timeout' }, canOperateTask ? 'clickable' : '']" @click="canOperateTask && openTaskDialog(s)">
                <span class="task-status" :class="'ts-' + s.status" />
                <span class="task-name">{{ s.name }}</span>
                <span v-if="s.assignee_names" class="task-user">{{ s.assignee_names }}</span>
              </div>
            </div>
          </div>

          <!-- 任务操作弹框 -->
          <div v-if="selectedTask" class="task-dialog-overlay" @click.self="selectedTask = null">
            <div class="task-dialog">
              <div class="td-hd">
                <span>{{ selectedTask.name }}</span>
                <button class="td-close" @click="selectedTask = null">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                </button>
              </div>
              <div class="td-body">
                <div class="td-info">
                  <div>状态: <span :class="'td-status-' + selectedTask.status">{{ selectedTask.status }}</span></div>
                  <div>执行人: {{ selectedTask.assignee_names || selectedTask.executor_team || '--' }}</div>
                  <div v-if="selectedTask.estimated_duration_minutes">预计: {{ selectedTask.estimated_duration_minutes }}分钟</div>
                </div>
                <div class="td-actions">
                  <button v-if="selectedTask.status === 'running' && canActOn(selectedTask)" class="td-btn done" @click="actOnTask('complete')">✓ 完成</button>
                  <button v-if="selectedTask.status === 'running' && canActOn(selectedTask)" class="td-btn skip" @click="actOnTask('skip')">↷ 跳过</button>
                </div>
              </div>
            </div>
          </div>

          <!-- 实时消息流 -->
          <div class="rp-block rp-logs">
            <div class="rp-hd">
              <svg class="rp-ico" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z"/></svg>
              <span>实时消息</span>
            </div>
            <div ref="logContainerRef" class="rp-body log-body">
              <div v-if="displayLogs.length === 0" class="rp-empty">等待消息推送...</div>
              <div v-for="(log, i) in displayLogs" :key="i" class="log-row" :class="'log-' + log.type">
                <span class="log-time">{{ log.time }}</span>
                <span class="log-icon">{{ log.icon }}</span>
                <span class="log-msg">{{ log.msg }}</span>
              </div>
            </div>
          </div>

          <!-- 计时与预警 -->
          <div class="rp-block rp-timer">
            <div class="rp-hd">
              <svg class="rp-ico" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 6v6l4 2"/></svg>
              <span>计时预警</span>
            </div>
            <div class="timer-body">
              <div class="timer-step-name">{{ currentRunningStep?.name || '无进行中环节' }}</div>
              <div class="timer-countdown" :class="{ warning: stepRemaining <= 60 && stepRemaining > 0, danger: stepRemaining <= 0 && currentRunningStep }">
                {{ stepRemainingStr }}
              </div>
              <div class="timer-label">{{ stepRemainingLabel }}</div>
              <div class="timer-ov">
                <div class="timer-ov-bar">
                  <div class="timer-ov-fill" :style="{ width: (instance?.progress_pct || 0) + '%' }" />
                </div>
                <span class="timer-ov-text">整体进度 {{ instance?.progress_pct || 0 }}%</span>
              </div>
            </div>
          </div>
        </aside>
      </main>

      <!-- 底部公告栏 -->
      <footer class="bottom-bar">
        <div class="bb-line" />
        <div class="bb-item bb-left">
          <svg class="bb-icon" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 16.92v3a2 2 0 01-2.18 2 19.79 19.79 0 01-8.63-3.07 19.5 19.5 0 01-6-6 19.79 19.79 0 01-3.07-8.67A2 2 0 014.11 2h3a2 2 0 012 1.72c.127.96.361 1.903.7 2.81a2 2 0 01-.45 2.11L8.09 9.91a16 16 0 006 6l1.27-1.27a2 2 0 012.11-.45c.907.339 1.85.573 2.81.7A2 2 0 0122 16.92z"/></svg>
          <span class="bb-label">紧急联系人</span>
          <span class="bb-val">{{ emergencyContact }}</span>
        </div>
        <div class="bb-item bb-center">
          <svg class="bb-icon" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
          <span class="bb-label">注意事项</span>
          <span class="bb-val">{{ drillNotes }}</span>
        </div>
        <div class="bb-item bb-right">
          <svg class="bb-icon" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
          <span class="bb-label">演练时间</span>
          <span class="bb-val">{{ scheduleText }}</span>
        </div>
      </footer>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRoute } from 'vue-router'
import { drillApi } from '@/api/modules/drill'
import { useAuthStore } from '@/stores/auth'
import type { DrillInstance, StepInstance, DrillStatus } from '@/types/instance'
import { DRILL_STATUS_LABELS } from '@/types/instance'

const route = useRoute()
const authStore = useAuthStore()
const role = computed(() => authStore.role)
const userDept = computed(() => authStore.user?.department ?? '')
const isDirector = computed(() => role.value === 'director' || role.value === 'admin')
const isExecutor = computed(() => role.value === 'executor')
const isViewer = computed(() => role.value === 'viewer')
const canControl = computed(() => isDirector.value)
const canOperateTask = computed(() => isDirector.value || isExecutor.value)
const drillId = computed(() => Number(route.params.id))

// ======== 数据状态 ========

const loading = ref(true)
const error = ref('')
const instance = ref<DrillInstance | null>(null)
const steps = ref<StepInstance[]>([])
const logs = ref<{ id: number; time: string; icon: string; type: string; msg: string }[]>([])
const wsConnected = ref(false)

// 计时器
const now = ref(Date.now())
const stepRemaining = ref(0)
const selectedTask = ref<StepInstance | null>(null)
let timerInterval: ReturnType<typeof setInterval> | null = null
let pollingTimer: ReturnType<typeof setInterval> | null = null

// ======== 计算属性 ========

const statusLabel = computed(() => {
  if (!instance.value) return '加载中'
  return DRILL_STATUS_LABELS[instance.value.status as DrillStatus] || instance.value.status
})

const elapsed = computed(() => {
  const inst = instance.value as Record<string, unknown> | null
  const started = (inst?.start_time || inst?.started_at) as string | undefined
  if (!started) return '--'
  const start = new Date(started).getTime()
  const diff = Math.max(0, now.value - start)
  const h = Math.floor(diff / 3600000)
  const m = Math.floor((diff % 3600000) / 60000)
  const s = Math.floor((diff % 60000) / 1000)
  if (h > 0) return `${pad(h)}:${pad(m)}:${pad(s)}`
  return `${pad(m)}:${pad(s)}`
})

// 左上角时间：已完成=最终耗时，进行中=自走时钟，待启动=提示
const displayTime = computed(() => {
  const inst = instance.value as Record<string, unknown> | null
  if (!inst) return '--'
  if (inst.status === 'pending') return '待启动'
  if (inst.status === 'completed' || inst.status === 'terminated') {
    const started = (inst.start_time || inst.started_at) as string | undefined
    const ended = (inst.end_time || inst.completed_at) as string | undefined
    if (started && ended) {
      const diff = Math.max(0, new Date(ended).getTime() - new Date(started).getTime())
      const h = Math.floor(diff / 3600000)
      const m = Math.floor((diff % 3600000) / 60000)
      const s = Math.floor((diff % 60000) / 1000)
      if (h > 0) return `${pad(h)}:${pad(m)}:${pad(s)}`
      return `${pad(m)}:${pad(s)}`
    }
    return '--'
  }
  // 进行中/暂停：自走时钟
  return elapsed.value
})

const runningSteps = computed(() => {
  const tasks = steps.value.filter(s => s.status === 'running' || s.status === 'timeout')
  // 执行者只看自己部门的任务
  if (isExecutor.value && userDept.value) {
    return tasks.filter(s => (s.executor_team || '') === userDept.value)
  }
  return tasks
})

const currentRunningStep = computed(() => steps.value.find(s => s.status === 'running'))

const stepRemainingStr = computed(() => {
  if (!currentRunningStep.value) return '--:--'
  if (stepRemaining.value <= 0) return '00:00'
  const m = Math.floor(stepRemaining.value / 60)
  const s = stepRemaining.value % 60
  return `${pad(m)}:${pad(s)}`
})

const stepRemainingLabel = computed(() => {
  if (!currentRunningStep.value) return '等待开始'
  if (stepRemaining.value <= 0) return '已超时'
  return '当前环节剩余'
})

const emergencyContact = computed(() => {
  return '运维指挥中心 (分机: 8888)'
})

const drillNotes = computed(() => {
  return instance.value?.description || '本次为实战演练，请保持通讯畅通，严格按照步骤执行'
})

const scheduleText = computed(() => {
  const inst = instance.value as Record<string, unknown> | null
  // 待启动：显示计划启动时间
  if (inst?.status === 'pending') {
    const planned = (inst?.planned_start || inst?.plannedStart) as string | undefined
    if (planned) {
      const plannedDate = new Date(planned)
      const totalMin = steps.value.reduce((s, x) => s + (x.estimated_duration_minutes || 5), 0)
      if (totalMin > 0) {
        const estEnd = new Date(plannedDate.getTime() + totalMin * 60000)
        return `${fmt(plannedDate)} — ${fmt(estEnd)}（预计 ${totalMin} 分钟）`
      }
      return `${fmt(plannedDate)}`
    }
    // 无计划时间，用预估时长
    const totalMin = steps.value.reduce((s, x) => s + (x.estimated_duration_minutes || 5), 0)
    if (totalMin > 0) return `预计 ${totalMin} 分钟`
    return '--'
  }
  // 已完成/终止：显示实际时间
  if (inst?.status === 'completed' || inst?.status === 'terminated') {
    const startTime = (inst?.start_time || inst?.started_at) as string | undefined
    if (!startTime) return '--'
    const start = new Date(startTime)
    const endTime = (inst?.end_time || inst?.completed_at) as string | undefined
    if (endTime) {
      return `${fmt(start)} — ${fmt(new Date(endTime))}`
    }
  }
  // 进行中：显示开始时间 + 预估耗时
  const startTime = (inst?.start_time || inst?.started_at) as string | undefined
  if (!startTime) return '--'
  const start = new Date(startTime)
  const totalMin = steps.value.reduce((s, x) => s + (x.estimated_duration_minutes || 5), 0)
  const estEnd = new Date(start.getTime() + totalMin * 60000)
  return `${fmt(start)} — ${fmt(estEnd)}（预计 ${totalMin} 分钟）`
})

// 实时日志（最近 8 条）
const displayLogs = computed(() => logs.value.slice(0, 8))
const logContainerRef = ref<HTMLElement | null>(null)

// ======== 树结构 ========

interface TreeNodePhase {
  name: string
  phaseSteps: TreeNodePhaseStep[]
}

interface TreeNodePhaseStep {
  name: string
  stepNodes: StepInstance[]
}

const treeData = computed<TreeNodePhase[]>(() => {
  const phases = new Map<string, Map<string, StepInstance[]>>()
  for (const s of steps.value) {
    const phase = s.phase || '默认阶段'
    const phaseStep = s.phase_step || phase
    let psMap = phases.get(phase)
    if (!psMap) {
      psMap = new Map<string, StepInstance[]>()
      phases.set(phase, psMap)
    }
    let arr = psMap.get(phaseStep)
    if (!arr) {
      arr = []
      psMap.set(phaseStep, arr)
    }
    arr.push(s)
  }
  const result: TreeNodePhase[] = []
  for (const [phaseName, psMap] of phases) {
    const phaseSteps: TreeNodePhaseStep[] = []
    for (const [psName, stepNodes] of psMap) {
      stepNodes.sort((a, b) => a.seq - b.seq)
      phaseSteps.push({ name: psName, stepNodes })
    }
    result.push({ name: phaseName, phaseSteps })
  }
  return result
})

// ======== Canvas 流程树 ========

const flowCanvasRef = ref<HTMLCanvasElement | null>(null)
let animFrameId = 0
let animTime = 0

// 横向布局常量（紧凑）
const PHASE_X = 20
const PHASE_W = 110
const COL_GAP = 30
const PS_X = PHASE_X + PHASE_W + COL_GAP
const PS_W = 100
const STEP_X = PS_X + PS_W + COL_GAP
const STEP_W = 100
const STEP_GAP_Y = 42
const PS_GAP_Y = 46
const expandedPhaseSteps = ref<Set<string>>(new Set())

function initCanvas() {
  const canvas = flowCanvasRef.value
  if (!canvas) return
  const parent = canvas.parentElement
  if (!parent) return
  const dpr = window.devicePixelRatio || 1
  const rect = parent.getBoundingClientRect()
  const w = rect.width
  // 按内容计算所需高度
  const contentH = calcContentHeight() + 40
  const h = Math.max(rect.height, contentH)
  canvas.width = w * dpr
  canvas.height = h * dpr
  canvas.style.width = w + 'px'
  canvas.style.height = h + 'px'
}

// 计算树内容需要的总高度
function calcContentHeight(): number {
  const data = treeData.value
  const aps = findActivePhaseStepName()
  if (!data.length) return 200
  const avgRegionH = 200 / data.length
  let total = 0
  for (const phase of data) {
    const psHeights = phase.phaseSteps.map(ps => {
      const isExpanded = ps.stepNodes.some(s => s.status === 'running') ||
        (aps && ps.name === aps) ||
        expandedPhaseSteps.value.has(ps.name)
      if (!isExpanded || ps.stepNodes.length === 0) return PS_GAP_Y
      const topN = ps.stepNodes.filter(s => !s.parent_step_id).length
      const subN = ps.stepNodes.filter(s => s.parent_step_id).length
      const gap = topN > 0 && subN > 0 ? 14 : 0
      return topN * STEP_GAP_Y + gap + subN * STEP_GAP_Y + PS_GAP_Y
    })
    total += Math.max(psHeights.reduce((a, b) => a + b, 0), avgRegionH)
  }
  return total
}

// 找到活动的 PhaseStep（包含 running 步骤的环节）
function findActivePhaseStepName(): string | null {
  for (const s of steps.value) {
    if (s.status === 'running') {
      return s.phase_step || s.phase || null
    }
  }
  return null
}

function drawFlowTree() {
  const canvas = flowCanvasRef.value
  if (!canvas || !treeData.value.length) return
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  const dpr = window.devicePixelRatio || 1
  ctx.clearRect(0, 0, canvas.width, canvas.height)
  const H = canvas.height / dpr
  ctx.save()
  ctx.scale(dpr, dpr)

  const data = treeData.value
  const activePSName = findActivePhaseStepName()

  // 先算每个 phase 的 ps 总高
  const phaseSizes = data.map(phase => {
    const psHeights = phase.phaseSteps.map(ps => {
      const isExpanded = ps.stepNodes.some(s => s.status === 'running') ||
        (activePSName && ps.name === activePSName) ||
        expandedPhaseSteps.value.has(ps.name)
      if (!isExpanded || ps.stepNodes.length === 0) return PS_GAP_Y
      const topN = ps.stepNodes.filter(s => !s.parent_step_id).length
      const subN = ps.stepNodes.filter(s => s.parent_step_id).length
      const gap = topN > 0 && subN > 0 ? 14 : 0
      return topN * STEP_GAP_Y + gap + subN * STEP_GAP_Y + PS_GAP_Y
    })
    return psHeights.reduce((a, b) => a + b, 0)
  })

  let yCursor = 10
  const phaseNodes: { x: number; y: number; phase: TreeNodePhase }[] = []

  data.forEach((phase, pi) => {
    const regionTop = yCursor
    const regionH = phaseSizes[pi]
    const phaseCenterY = regionTop + regionH / 2
    yCursor += regionH

    drawNode(ctx, PHASE_X + PHASE_W / 2, phaseCenterY, phase.name, 'phase')
    phaseNodes.push({ x: PHASE_X + PHASE_W, y: phaseCenterY, phase })

    const psHeights = phase.phaseSteps.map(ps => {
      const isExpanded = ps.stepNodes.some(s => s.status === 'running') ||
        (activePSName && ps.name === activePSName) ||
        expandedPhaseSteps.value.has(ps.name)
      if (!isExpanded || ps.stepNodes.length === 0) return PS_GAP_Y
      const topN = ps.stepNodes.filter(s => !s.parent_step_id).length
      const subN = ps.stepNodes.filter(s => s.parent_step_id).length
      const gap = topN > 0 && subN > 0 ? 14 : 0
      return topN * STEP_GAP_Y + gap + subN * STEP_GAP_Y + PS_GAP_Y
    })
    const totalPSH = psHeights.reduce((a, b) => a + b, 0)
    let psCursor = regionTop + (regionH - totalPSH) / 2

    // 右侧：PhaseStep 节点（动态间距）
    phase.phaseSteps.forEach((ps, psi) => {
      const psY = psCursor + psHeights[psi] / 2
      psCursor += psHeights[psi]
      const isActive = ps.stepNodes.some(s => s.status === 'running') ||
        (activePSName && ps.name === activePSName)
      const isExpanded = isActive || expandedPhaseSteps.value.has(ps.name)

      // 连线 Phase → PhaseStep
      const connColor = getConnColor(ps.stepNodes)
      drawHCurve(ctx, PHASE_X + PHASE_W, phaseCenterY, PS_X, psY, connColor)

      if (isExpanded && ps.stepNodes.length > 0) {
        drawNode(ctx, PS_X + PS_W / 2, psY, ps.name, 'phase-step')
        const steps = ps.stepNodes
        const topLevel = steps.filter(s => !s.parent_step_id)
        const subSteps = steps.filter(s => s.parent_step_id)
        const indentX = 30

        // 计算总高度并居中
        const topH = topLevel.length > 0 ? topLevel.length * STEP_GAP_Y : 0
        const gapH = (topLevel.length > 0 && subSteps.length > 0) ? 14 : 0
        const subH = subSteps.length > 0 ? subSteps.length * STEP_GAP_Y : 0
        const totalH = topH + gapH + subH
        const startY = psY - totalH / 2

        // 顶层步骤
        topLevel.forEach((step, si) => {
          const sy = startY + si * STEP_GAP_Y + STEP_GAP_Y / 2
          const color = step.status === 'running' ? '#00FFFF'
            : step.status === 'completed' || step.status === 'skipped' ? '#00BFFF'
            : '#1A2A3A'
          drawHCurve(ctx, PS_X + PS_W, psY, STEP_X, sy, color)
          drawStepNode(ctx, STEP_X + STEP_W / 2, sy, step)
        })

        // 子步骤（缩进，留空隙）
        if (subSteps.length > 0) {
          const subStartY = startY + topH + gapH
          subSteps.forEach((step, si) => {
            const sy = subStartY + si * STEP_GAP_Y + STEP_GAP_Y / 2
            const color = step.status === 'running' ? '#00FFFF'
              : step.status === 'completed' || step.status === 'skipped' ? '#00BFFF'
              : '#1A2A3A'
            drawHCurve(ctx, PS_X + PS_W, psY, STEP_X + indentX, sy, color)
            drawStepNode(ctx, STEP_X + indentX + STEP_W / 2, sy, step)
          })
        }
      } else {
        // 非活动分支：折叠显示
        drawCollapsedNode(ctx, PS_X + PS_W / 2, psY, ps.name, ps.stepNodes.length, ps.stepNodes.some(s => s.status === 'running'))
      }
    })
  })

  ctx.restore()
}

// 折叠节点（显示环节名 + 步骤数量）
function drawCollapsedNode(ctx: CanvasRenderingContext2D, x: number, y: number, label: string, stepCount: number, isActive: boolean) {
  const w = PS_W
  const h = 28
  const r = 6
  ctx.save()
  ctx.shadowColor = isActive ? 'rgba(0, 255, 255, 0.4)' : 'rgba(0, 180, 255, 0.2)'
  ctx.shadowBlur = isActive ? 8 : 3
  ctx.fillStyle = isActive ? 'rgba(0, 40, 60, 0.95)' : 'rgba(5, 20, 50, 0.9)'
  ctx.strokeStyle = isActive ? 'rgba(0, 255, 255, 0.5)' : 'rgba(0, 180, 255, 0.3)'
  ctx.lineWidth = 1
  roundRect(ctx, x - w / 2, y - h / 2, w, h, r)
  ctx.fill()
  ctx.stroke()
  ctx.shadowBlur = 0
  ctx.fillStyle = isActive ? '#00FFFF' : '#4A7A9A'
  ctx.font = '11px "PingFang SC", "Microsoft YaHei", sans-serif'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText(label, x, y - 3)
  ctx.fillStyle = '#3A5A6A'
  ctx.font = '9px "PingFang SC", sans-serif'
  ctx.fillText('▸ ' + stepCount + ' steps', x, y + 8)
  ctx.restore()
}

// 横向贝塞尔曲线
function drawHCurve(ctx: CanvasRenderingContext2D, x1: number, y1: number, x2: number, y2: number, color: string) {
  ctx.save()
  ctx.strokeStyle = color
  ctx.lineWidth = 1.5
  ctx.shadowColor = color
  ctx.shadowBlur = 2
  ctx.beginPath()
  ctx.moveTo(x1, y1)
  const cpX = (x1 + x2) / 2
  ctx.bezierCurveTo(cpX, y1, cpX, y2, x2, y2)
  ctx.stroke()
  ctx.shadowBlur = 0
  ctx.restore()
}

function drawNode(ctx: CanvasRenderingContext2D, x: number, y: number, label: string, type: string) {
  const isPhase = type === 'phase'
  const w = isPhase ? 140 : 120
  const h = isPhase ? 36 : 30
  const r = 6
  ctx.save()
  ctx.shadowColor = 'rgba(0, 180, 255, 0.4)'
  ctx.shadowBlur = isPhase ? 12 : 6
  ctx.fillStyle = 'rgba(5, 20, 50, 0.95)'
  ctx.strokeStyle = 'rgba(0, 180, 255, 0.6)'
  ctx.lineWidth = 1
  roundRect(ctx, x - w / 2, y - h / 2, w, h, r)
  ctx.fill()
  ctx.stroke()
  ctx.shadowBlur = 0
  ctx.fillStyle = '#00BFFF'
  ctx.font = `${isPhase ? 13 : 11}px "Courier New", monospace`
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText(label, x, y)
  ctx.restore()
}

function drawStepNode(ctx: CanvasRenderingContext2D, x: number, y: number, step: StepInstance) {
  const w = 110
  const h = 32
  const r = 4
  const isRunning = step.status === 'running'
  const isDone = step.status === 'completed' || step.status === 'skipped'
  const isPending = step.status === 'pending'

  ctx.save()

  if (isRunning) {
    const pulse = Math.sin(animTime * 0.05) * 0.3 + 0.7
    ctx.shadowColor = 'rgba(0, 255, 255, ' + (0.6 * pulse) + ')'
    ctx.shadowBlur = 15 * pulse
    ctx.strokeStyle = '#00FFFF'
    ctx.lineWidth = 2
    ctx.fillStyle = 'rgba(0, 40, 60, 0.95)'
  } else if (isDone) {
    ctx.shadowColor = 'rgba(0, 200, 100, 0.3)'
    ctx.shadowBlur = 4
    ctx.strokeStyle = 'rgba(0, 180, 255, 0.5)'
    ctx.lineWidth = 1
    ctx.fillStyle = 'rgba(0, 25, 40, 0.9)'
  } else {
    ctx.shadowBlur = 0
    ctx.strokeStyle = '#1A2A3A'
    ctx.lineWidth = 1
    ctx.fillStyle = 'rgba(5, 12, 24, 0.9)'
  }

  roundRect(ctx, x - w / 2, y - h / 2, w, h, r)
  ctx.fill()
  ctx.stroke()
  ctx.shadowBlur = 0

  // 名称
  const textColor = isRunning ? '#00FFFF' : isDone ? '#00BFFF' : '#4A5568'
  ctx.fillStyle = textColor
  ctx.font = '10px "PingFang SC", "Microsoft YaHei", sans-serif'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  const name = step.name.length > 8 ? step.name.slice(0, 7) + '..' : step.name
  ctx.fillText(name, x, y - 4)

  // 执行人
  ctx.fillStyle = isPending ? '#2A3A4A' : '#5A7A9A'
  ctx.font = '8px "PingFang SC", sans-serif'
  ctx.fillText(step.assignee_names || step.executor_team || '--', x, y + 8)

  ctx.restore()
}

function getConnColor(nodes: StepInstance[]): string {
  if (nodes.some(s => s.status === 'running')) return '#00FFFF'
  if (nodes.every(s => s.status === 'completed' || s.status === 'skipped')) return '#00BFFF'
  return '#1A2A3A'
}

function roundRect(ctx: CanvasRenderingContext2D, x: number, y: number, w: number, h: number, r: number) {
  ctx.beginPath()
  ctx.moveTo(x + r, y)
  ctx.lineTo(x + w - r, y)
  ctx.quadraticCurveTo(x + w, y, x + w, y + r)
  ctx.lineTo(x + w, y + h - r)
  ctx.quadraticCurveTo(x + w, y + h, x + w - r, y + h)
  ctx.lineTo(x + r, y + h)
  ctx.quadraticCurveTo(x, y + h, x, y + h - r)
  ctx.lineTo(x, y + r)
  ctx.quadraticCurveTo(x, y, x + r, y)
  ctx.closePath()
}

function animateLoop() {
  animTime++
  drawFlowTree()
  animFrameId = requestAnimationFrame(animateLoop)
}

function handleCanvasClick(e: MouseEvent) {
  const canvas = flowCanvasRef.value
  if (!canvas || !treeData.value.length) return
  const rect = canvas.getBoundingClientRect()
  const x = e.clientX - rect.left
  const y = e.clientY - rect.top

  const data = treeData.value
  const aps = findActivePhaseStepName()
  let pyCursor = 10

  for (let pi = 0; pi < data.length; pi++) {
    const phase = data[pi]
    const psHeights = phase.phaseSteps.map(ps => {
      const isExpanded = ps.stepNodes.some(s => s.status === 'running') ||
        (aps && ps.name === aps) ||
        expandedPhaseSteps.value.has(ps.name)
      if (!isExpanded || ps.stepNodes.length === 0) return PS_GAP_Y
      const topN = ps.stepNodes.filter(s => !s.parent_step_id).length
      const subN = ps.stepNodes.filter(s => s.parent_step_id).length
      const gap = topN > 0 && subN > 0 ? 14 : 0
      return topN * STEP_GAP_Y + gap + subN * STEP_GAP_Y + PS_GAP_Y
    })
    const regionH = psHeights.reduce((a, b) => a + b, 0)
    const regionTop = pyCursor
    pyCursor += regionH

    const totalPSH = psHeights.reduce((a, b) => a + b, 0)
    let psCursor = regionTop + (regionH - totalPSH) / 2

    for (let psi = 0; psi < phase.phaseSteps.length; psi++) {
      const psY = psCursor + psHeights[psi] / 2
      psCursor += psHeights[psi]
      if (Math.abs(x - (PS_X + PS_W / 2)) < PS_W / 2 && Math.abs(y - psY) < 15) {
        const ps = phase.phaseSteps[psi]
        if (!ps.stepNodes.some(s => s.status === 'running') && ps.stepNodes.length > 0) {
          const next = new Set(expandedPhaseSteps.value)
          next.has(ps.name) ? next.delete(ps.name) : next.add(ps.name)
          expandedPhaseSteps.value = next
          // 展开后重新计算画布大小
          nextTick(() => initCanvas())
        }
        return
      }
    }
  }
}

// ======== WebSocket ========

let ws: WebSocket | null = null
let wsReconnectTimer: ReturnType<typeof setTimeout> | null = null
let wsReconnectCount = 0

function connectWS() {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  const url = `${proto}://${location.host}/ws/display/${drillId.value}`
  ws = new WebSocket(url)

  ws.onopen = () => {
    wsConnected.value = true
    wsReconnectCount = 0
  }

  ws.onclose = () => {
    wsConnected.value = false
    scheduleReconnect()
  }

  ws.onerror = () => {
    ws?.close()
  }

  ws.onmessage = (ev) => {
    try {
      const msg = JSON.parse(ev.data)
      handleWSMessage(msg)
    } catch { /* ignored */ }
  }
}

function scheduleReconnect() {
  if (wsReconnectTimer) clearTimeout(wsReconnectTimer)
  if (wsReconnectCount >= 10) return
  wsReconnectCount++
  wsReconnectTimer = setTimeout(connectWS, 3000)
}

function handleWSMessage(msg: any) {
  const event = msg.event || msg.type || ''
  const payload = msg.payload || msg.data || msg
  const stepName = payload.step_name || payload.stepName || ''

  if (['drill_started', 'drill_paused', 'drill_resumed', 'drill_completed', 'drill_terminated'].includes(event)) {
    fetchDrillData()
  }

  if (['step_started', 'step_complete', 'step_timeout', 'step_skipped'].includes(event)) {
    fetchSteps()
    addLog(event === 'step_timeout' ? 'warn' : 'info', logIcon(event), `${stepName} ${logLabel(event)}`)
  }

  if (event === 'timeout_warning') {
    const remaining = payload.remaining_sec || 0
    addLog('warn', '⏰', `${stepName} 剩余 ${remaining} 秒`)
  }
}

function logLabel(event: string): string {
  const map: Record<string, string> = {
    step_started: '已开始',
    step_complete: '已完成',
    step_timeout: '已超时',
    step_skipped: '已跳过',
    drill_started: '演练开始',
    drill_paused: '已暂停',
    drill_resumed: '已恢复',
    drill_completed: '已完成',
    drill_terminated: '已终止',
  }
  return map[event] || event
}

function logIcon(event: string): string {
  if (event.includes('timeout')) return '⚠'
  if (event.includes('complete') || event.includes('skipped')) return '✓'
  return '●'
}

function addLog(type: string, icon: string, msg: string) {
  const nowDate = new Date()
  const time = pad(nowDate.getHours()) + ':' + pad(nowDate.getMinutes()) + ':' + pad(nowDate.getSeconds())
  const entry = { id: Date.now(), time, icon, type, msg }
  logs.value.unshift(entry)
  if (logs.value.length > 50) logs.value.length = 50
  nextTick(scrollLogs)
}

function scrollLogs() {
  const el = logContainerRef.value
  if (el) el.scrollTop = 0
}

// ======== 数据加载 ========

async function fetchDrillData() {
  try {
    instance.value = await drillApi.getDetail(drillId.value)
  } catch {
    // 静默失败
  }
}

async function fetchSteps() {
  try {
    const data = await drillApi.getSteps(drillId.value)
    steps.value = (data || []).sort((a: StepInstance, b: StepInstance) => a.seq - b.seq)
  } catch {
    // 静默失败
  }
}

async function fetchLogs() {
  try {
    const data = await drillApi.getLogs(drillId.value)
    const logData = (data || [])
    const items = logData.slice(-8).reverse().map((l: Record<string, unknown>) => {
      const action = (l.Action || l.action || l.Content || '') as string
      return {
        id: (l.ID || l.id) as number,
        time: fmtTime((l.CreatedAt || l.created_at) as string),
        icon: '●',
        type: action.includes('timeout') ? 'warn' : 'info',
        msg: action,
      }
    })
    logs.value = items
  } catch {
    // 静默失败
  }
}

async function loadAllData() {
  loading.value = true
  error.value = ''
  try {
    await Promise.all([fetchDrillData(), fetchSteps(), fetchLogs()])
  } catch {
    error.value = '数据加载失败'
  } finally {
    loading.value = false
  }
}

function handleRetry() {
  error.value = ''
  loadAllData()
}

function toggleFullscreen() {
  const el = document.querySelector('.screen-root') as HTMLElement
  if (!el) return
  if (document.fullscreenElement) {
    document.exitFullscreen()
  } else {
    el.requestFullscreen()
  }
}

async function handleStart() {
  try { await drillApi.start(drillId.value); fetchDrillData(); fetchSteps() } catch { /* */ }
}
async function handlePause() {
  try { await drillApi.pause(drillId.value); fetchDrillData() } catch { /* */ }
}
async function handleResume() {
  try { await drillApi.resume(drillId.value); fetchDrillData() } catch { /* */ }
}
async function handleTerminate() {
  try { await drillApi.terminate(drillId.value); fetchDrillData() } catch { /* */ }
}

function canActOn(step: StepInstance): boolean {
  if (!canOperateTask.value) return false
  const taskTeam = step.executor_team || ''
  if (!taskTeam) return true
  return userDept.value === taskTeam
}

function openTaskDialog(step: StepInstance) {
  selectedTask.value = step
}

async function actOnTask(action: string) {
  if (!selectedTask.value) return
  const step = selectedTask.value
  try {
    if (action === 'complete') {
      await drillApi.completeStep(drillId.value, step.step_template_id || step.id, '大屏快速完成')
    } else if (action === 'skip') {
      await drillApi.skipStep(drillId.value, step.step_template_id || step.id, '大屏跳过')
    }
    selectedTask.value = null
    fetchSteps()
    fetchDrillData()
  } catch {
    // 静默失败
  }
}

// ======== 计时更新 ========

function updateTimer() {
  now.value = Date.now()
  const runningStep = steps.value.find(s => s.status === 'running')
  if (runningStep?.timeout_at) {
    const timeout = new Date(runningStep.timeout_at).getTime()
    stepRemaining.value = Math.max(0, Math.floor((timeout - now.value) / 1000))
  } else {
    stepRemaining.value = 0
  }
}

// ======== 生命周期 ========

onMounted(() => {
  // 阻断父容器滚动条
  const html = document.documentElement
  const body = document.body
  const oldOverflow = { h: html.style.overflow, b: body.style.overflow }
  html.style.overflow = 'hidden'
  body.style.overflow = 'hidden'

  loadAllData().then(() => {
    initCanvas()
    drawFlowTree()
    animFrameId = requestAnimationFrame(animateLoop)
  })
  connectWS()
  timerInterval = setInterval(updateTimer, 1000)
  pollingTimer = setInterval(() => {
    if (!wsConnected.value) fetchDrillData()
  }, 30000)
  window.addEventListener('resize', onResize)

  // Canvas 点击：展开/折叠环节
  const canvasEl = flowCanvasRef.value
  if (canvasEl) canvasEl.addEventListener('click', handleCanvasClick)
})

onUnmounted(() => {
  // 恢复滚动
  document.documentElement.style.overflow = ''
  document.body.style.overflow = ''

  cancelAnimationFrame(animFrameId)
  if (ws) { ws.close(); ws = null }
  if (wsReconnectTimer) clearTimeout(wsReconnectTimer)
  if (timerInterval) clearInterval(timerInterval)
  if (pollingTimer) clearInterval(pollingTimer)
  window.removeEventListener('resize', onResize)
  const canvasEl = flowCanvasRef.value
  if (canvasEl) canvasEl.removeEventListener('click', handleCanvasClick)
})

function onResize() {
  initCanvas()
}

// 侦听 steps 变化重绘 Canvas
watch([steps, () => instance.value?.progress_pct], () => {
  nextTick(() => {
    initCanvas()
  })
}, { deep: false })

// 滚动日志
watch(displayLogs, () => {
  nextTick(scrollLogs)
})

// ======== 工具函数 ========

function pad(n: number): string {
  return n < 10 ? '0' + n : String(n)
}

function fmt(d: Date): string {
  return `${d.getFullYear()}/${pad(d.getMonth() + 1)}/${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function fmtTime(ts: string): string {
  if (!ts) return '--:--:--'
  const d = new Date(ts)
  return `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}
</script>

<style scoped>
/* ===== 全局 ===== */

.screen-root {
  margin: -24px;
  height: calc(100vh - 56px);
  background: #080E1A;
  color: #C0CDE0;
  font-family: 'PingFang SC', 'Microsoft YaHei', 'Helvetica Neue', sans-serif;
  display: grid;
  grid-template-rows: 80px 1fr 56px;
  grid-template-columns: 1fr;
  overflow: hidden;
  user-select: none;
}

/* 隐藏所有滚动条 */
/* 隐藏所有滚动条（除流程区） */
.screen-root .right-panel *::-webkit-scrollbar { display: none; }
.screen-root .right-panel * { scrollbar-width: none; }

/* 流程区滚动条深色风格 */
.screen-root .flow-area::-webkit-scrollbar { width: 5px; }
.screen-root .flow-area::-webkit-scrollbar-track { background: rgba(0, 0, 0, 0.2); }
.screen-root .flow-area::-webkit-scrollbar-thumb { background: rgba(0, 150, 255, 0.3); border-radius: 3px; }

.bg-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(0, 150, 255, 0.04) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 150, 255, 0.04) 1px, transparent 1px);
  background-size: 60px 60px;
  pointer-events: none;
  z-index: 0;
}

.bg-glow {
  position: absolute;
  width: 500px;
  height: 500px;
  border-radius: 50%;
  filter: blur(120px);
  opacity: 0.07;
  pointer-events: none;
  z-index: 0;
}

.bg-glow-tl {
  top: -150px;
  left: -100px;
  background: #0044FF;
}

.bg-glow-br {
  bottom: -150px;
  right: -100px;
  background: #00AAFF;
}

/* ===== 加载与错误 ===== */

.overlay-state {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  z-index: 10;
  background: #060B18;
}

.loader-ring {
  width: 48px;
  height: 48px;
  border: 3px solid rgba(0, 150, 255, 0.2);
  border-top-color: #00BFFF;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.loader-text {
  margin-top: 16px;
  color: rgba(200, 214, 229, 0.7);
  font-size: 14px;
}

.overlay-state.error p {
  color: #FF6B6B;
  margin-bottom: 16px;
}

.btn-retry {
  padding: 8px 24px;
  background: rgba(0, 150, 255, 0.15);
  border: 1px solid rgba(0, 150, 255, 0.4);
  color: #00BFFF;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.btn-retry:hover {
  background: rgba(0, 150, 255, 0.25);
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* ===== 顶部信息栏 ===== */

.top-bar {
  position: relative;
  display: flex;
  align-items: center;
  padding: 0 28px;
  background: linear-gradient(180deg, rgba(0, 20, 50, 0.9) 0%, rgba(6, 11, 24, 0) 100%);
  z-index: 1;
  gap: 40px;
}

.tb-line {
  position: absolute;
  bottom: 0;
  left: 28px;
  right: 28px;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(0, 180, 255, 0.3) 20%, rgba(0, 255, 255, 0.5) 50%, rgba(0, 180, 255, 0.3) 80%, transparent);
}

.tb-left {
  display: flex;
  align-items: center;
  min-width: 90px;
}

.elapsed-time {
  font-size: 24px;
  font-weight: 700;
  color: #00FFFF;
  font-family: 'Courier New', monospace;
  letter-spacing: 2px;
}

.tb-center {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.tb-title-row {
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.tb-title {
  font-size: 22px;
  font-weight: 700;
  color: #E8F0F8;
  margin: 0;
  letter-spacing: 2px;
  text-align: center;
}

.tb-status {
  font-size: 12px;
  font-weight: 500;
  padding: 1px 8px;
  border-radius: 3px;
}

.st-pending  { color: #5A7A9A; background: rgba(90, 122, 154, 0.15); }
.st-running  { color: #00FF88; background: rgba(0, 255, 136, 0.1); }
.st-paused   { color: #FFD700; background: rgba(255, 215, 0, 0.1); }
.st-completed { color: #00BFFF; background: rgba(0, 191, 255, 0.1); }
.st-terminated { color: #FF6666; background: rgba(255, 102, 102, 0.1); }

.progress-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 200px;
}

.progress-track {
  flex: 1;
  height: 6px;
  background: rgba(255, 255, 255, 0.06);
  border-radius: 3px;
  position: relative;
  overflow: visible;
}

.progress-fill {
  height: 100%;
  border-radius: 3px;
  background: linear-gradient(90deg, #0066FF, #00FFFF);
  transition: width 0.8s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 0 8px rgba(0, 180, 255, 0.4);
}

.progress-glow {
  position: absolute;
  top: -4px;
  width: 14px;
  height: 14px;
  background: rgba(0, 255, 255, 0.8);
  border-radius: 50%;
  transform: translateX(-50%);
  box-shadow: 0 0 12px rgba(0, 255, 255, 0.6);
  transition: left 0.8s cubic-bezier(0.4, 0, 0.2, 1);
}

.progress-text {
  font-size: 16px;
  font-weight: 700;
  color: #00FFFF;
  min-width: 40px;
  text-align: right;
}

.tb-right {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  gap: 12px;
}

.btn-fullscreen {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 4px;
  border: 1px solid rgba(0, 150, 255, 0.2);
  background: rgba(0, 30, 60, 0.6);
  color: #5A8AAA;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-fullscreen:hover {
  border-color: rgba(0, 200, 255, 0.5);
  color: #00BFFF;
  background: rgba(0, 40, 80, 0.8);
}

.btn-ctrl {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 4px;
  border: 1px solid rgba(0, 150, 255, 0.2);
  cursor: pointer;
  transition: all 0.2s;
}

.btn-ctrl.start {
  background: rgba(0, 200, 100, 0.15);
  border-color: rgba(0, 200, 100, 0.4);
  color: #00FF88;
}

.btn-ctrl.start:hover {
  background: rgba(0, 200, 100, 0.3);
  box-shadow: 0 0 8px rgba(0, 255, 136, 0.3);
}

.btn-ctrl.pause {
  background: rgba(255, 180, 0, 0.15);
  border-color: rgba(255, 180, 0, 0.4);
  color: #FFD700;
}

.btn-ctrl.pause:hover {
  background: rgba(255, 180, 0, 0.3);
  box-shadow: 0 0 8px rgba(255, 215, 0, 0.3);
}

.btn-ctrl.resume {
  background: rgba(0, 200, 100, 0.15);
  border-color: rgba(0, 200, 100, 0.4);
  color: #00FF88;
}

.btn-ctrl.resume:hover {
  background: rgba(0, 200, 100, 0.3);
  box-shadow: 0 0 8px rgba(0, 255, 136, 0.3);
}

.btn-ctrl.end {
  background: rgba(255, 80, 80, 0.15);
  border-color: rgba(255, 80, 80, 0.4);
  color: #FF6666;
}

.btn-ctrl.end:hover {
  background: rgba(255, 80, 80, 0.3);
  box-shadow: 0 0 8px rgba(255, 102, 102, 0.3);
}

/* ===== 主体 ===== */

.main-body {
  display: grid;
  grid-template-columns: 1fr 280px;
  gap: 0;
  z-index: 1;
  min-height: 0;
  overflow: hidden;
}

/* ===== 流程区域 ===== */

.flow-area {
  position: relative;
  padding: 12px 16px 12px 16px;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow-y: auto;
}

.flow-canvas {
  width: 100%;
  min-height: 100%;
}

.flow-label {
  font-size: 12px;
  color: #00BFFF;
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 2px;
}

.flow-canvas {
  flex: 1;
  width: 100%;
  min-height: 0;
}

.flow-legend {
  display: flex;
  gap: 20px;
  justify-content: center;
  padding: 8px 0 0;
  font-size: 11px;
  color: #5A7A9A;
}

.lg-item { display: flex; align-items: center; gap: 6px; }

.lg-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.lg-done { background: #00BFFF; box-shadow: 0 0 4px #00BFFF; }
.lg-running { background: #00FFFF; box-shadow: 0 0 8px #00FFFF; animation: pulse-dot 1.5s ease-in-out infinite; }
.lg-pending { background: #4A5568; }

@keyframes pulse-dot {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

/* ===== 右侧面板 ===== */

.right-panel {
  display: flex;
  flex-direction: column;
  gap: 0;
  background: rgba(5, 15, 30, 0.7);
  border-left: 1px solid rgba(0, 150, 255, 0.15);
  backdrop-filter: blur(8px);
}

.rp-block {
  display: flex;
  flex-direction: column;
  border-bottom: 1px solid rgba(0, 150, 255, 0.1);
  min-height: 0;
  overflow: hidden;
}

.rp-tasks,
.rp-logs {
  resize: vertical;
}

.rp-hd {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 14px;
  font-size: 12px;
  font-weight: 600;
  color: #00BFFF;
  background: rgba(0, 20, 50, 0.5);
}

.rp-ico {
  flex-shrink: 0;
  color: #00BFFF;
}

.rp-badge {
  margin-left: auto;
  background: rgba(0, 180, 255, 0.2);
  color: #00FFFF;
  padding: 1px 7px;
  border-radius: 10px;
  font-size: 10px;
}

.rp-body {
  flex: 1;
  overflow-y: auto;
  padding: 6px 0;
  scrollbar-width: none;
}

.rp-body::-webkit-scrollbar { display: none; }

.log-body {
  scroll-behavior: smooth;
  overflow-y: auto;
  scrollbar-width: none;
}

.log-body::-webkit-scrollbar { display: none; }

.rp-empty {
  padding: 20px 14px;
  font-size: 11px;
  color: #3A4A5A;
  text-align: center;
}

/* 待办 */
.rp-tasks {
  max-height: 200px;
  flex-shrink: 0;
}

.task-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  font-size: 11px;
  border-left: 2px solid transparent;
  transition: background 0.15s;
}

.task-row.clickable {
  cursor: pointer;
}

.task-row.clickable:hover {
  background: rgba(0, 150, 255, 0.08);
}

.task-row.task-timeout {
  background: rgba(255, 68, 68, 0.08);
  border-left-color: #FF4444;
}

.task-status {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
  background: #4A5568;
}

.ts-running { background: #00FF88; }
.ts-timeout { background: #FFD700; }

.task-name {
  flex: 1;
  color: #C8D6E5;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.task-user {
  color: #5A7A9A;
  font-size: 10px;
  flex-shrink: 0;
}

/* 日志 */
.rp-logs {
  flex: 1;
  min-height: 0;
}

.log-body {
  scroll-behavior: smooth;
  overflow-y: auto;
}

.log-row {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  padding: 5px 14px;
  font-size: 10px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.03);
  line-height: 1.5;
}

.log-time {
  color: #4A5A6A;
  white-space: nowrap;
  min-width: 52px;
}

.log-icon {
  flex-shrink: 0;
  font-size: 9px;
}

.log-msg {
  flex: 1;
  color: #8A9AB0;
  word-break: break-all;
}

.log-warn .log-msg { color: #FFD700; }
.log-warn .log-icon { color: #FFD700; }
.log-error .log-msg { color: #FF6B6B; }
.log-error .log-icon { color: #FF6B6B; }

/* 计时预警 */
.rp-timer {
  max-height: 180px;
  flex-shrink: 0;
}

.timer-body {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 12px 14px;
  gap: 4px;
}

.timer-step-name {
  font-size: 11px;
  color: #5A7A9A;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.timer-countdown {
  font-size: 32px;
  font-weight: 700;
  color: #00FFFF;
  font-family: 'Courier New', monospace;
  letter-spacing: 3px;
}

.timer-countdown.warning {
  color: #FFD700;
  animation: blink 1s ease-in-out infinite;
}

.timer-countdown.danger {
  color: #FF4444;
  animation: blink 0.5s ease-in-out infinite;
}

.timer-label {
  font-size: 10px;
  color: #4A5A6A;
}

.timer-ov {
  width: 100%;
  margin-top: 8px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.timer-ov-bar {
  width: 100%;
  height: 3px;
  background: rgba(255, 255, 255, 0.06);
  border-radius: 2px;
}

.timer-ov-fill {
  height: 100%;
  border-radius: 2px;
  background: linear-gradient(90deg, #0066FF, #00FFFF);
  transition: width 0.6s ease;
}

.timer-ov-text {
  font-size: 10px;
  color: #5A7A9A;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

/* ===== 底部公告栏 ===== */

.bottom-bar {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(0, 15, 35, 0.8);
  border-top: 1px solid rgba(0, 150, 255, 0.15);
  z-index: 1;
  padding: 0 28px;
}

.bb-line {
  position: absolute;
  top: 0;
  left: 28px;
  right: 28px;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(0, 180, 255, 0.3) 20%, rgba(0, 255, 255, 0.3) 50%, rgba(0, 180, 255, 0.3) 80%, transparent);
}

.bb-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.bb-center {
  flex: 1;
  justify-content: center;
  max-width: 50%;
  overflow: hidden;
}

.bb-center .bb-val {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.bb-icon {
  flex-shrink: 0;
  color: #556C88;
}

.bb-label {
  color: #6B8AAA;
  white-space: nowrap;
}

.bb-val {
  color: #7B93AB;
  white-space: nowrap;
}

/* 任务弹框 */
.task-dialog-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
}

.task-dialog {
  background: #0F1A2E;
  border: 1px solid rgba(0, 180, 255, 0.3);
  border-radius: 8px;
  width: 320px;
  box-shadow: 0 0 30px rgba(0, 100, 200, 0.2);
}

.td-hd {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid rgba(0, 150, 255, 0.15);
  font-size: 14px;
  color: #E8F0F8;
  font-weight: 600;
}

.td-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: none;
  background: none;
  color: #5A7A9A;
  cursor: pointer;
  border-radius: 4px;
}

.td-close:hover { color: #FF6666; background: rgba(255, 100, 100, 0.1); }

.td-body {
  padding: 16px;
}

.td-info {
  font-size: 12px;
  color: #7B93AB;
  line-height: 2;
}

.td-status-running { color: #00FF88; }
.td-status-timeout { color: #FFD700; }

.td-actions {
  display: flex;
  gap: 8px;
  margin-top: 14px;
}

.td-btn {
  flex: 1;
  padding: 8px;
  border-radius: 4px;
  border: 1px solid;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.15s;
}

.td-btn.done {
  background: rgba(0, 200, 100, 0.12);
  border-color: rgba(0, 200, 100, 0.4);
  color: #00FF88;
}

.td-btn.done:hover { background: rgba(0, 200, 100, 0.25); }

.td-btn.skip {
  background: rgba(255, 180, 0, 0.12);
  border-color: rgba(255, 180, 0, 0.4);
  color: #FFD700;
}

.td-btn.skip:hover { background: rgba(255, 180, 0, 0.25); }
</style>

<style>
/* 隐藏此页面的面包屑 */
.app-main:has(.screen-root) .app-breadcrumb {
  display: none !important;
}

/* 阻断此页面所有父容器的滚动条 */
.app-main:has(.screen-root),
.app-main:has(.screen-root) .app-content {
  overflow: hidden !important;
}

/* 彻底移除滚动条轨道 */
.app-main:has(.screen-root)::-webkit-scrollbar { display: none !important; }
.app-content:has(.screen-root)::-webkit-scrollbar { display: none !important; }
</style>