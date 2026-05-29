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
              <div v-for="s in runningSteps" :key="s.id" class="task-row" :class="[{ 'task-timeout': s.status === 'timeout' }]">
                <span class="task-status" :class="'ts-' + s.status" />
                <span class="task-name">{{ s.name }}</span>
                <template v-if="isDirector && !isParentStep(s)">
                  <button class="task-btn task-btn-skip" title="跳过" @click="skipTask(s)">↷</button>
                  <button class="task-btn task-btn-done" title="强制完成" @click="forceCompleteTask(s)">✓</button>
                </template>
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

      <!-- 任务完成弹窗 -->
      <Transition name="modal">
        <div v-if="completionModal.visible" class="completion-modal" @click="completionModal.visible = false">
          <div class="completion-modal-content" @click.stop>
            <div class="completion-icon">✓</div>
            <div class="completion-text">
              <div class="completion-title">任务完成</div>
              <div class="completion-step">{{ completionModal.stepName }}</div>
              <div v-if="completionModal.phaseName" class="completion-phase">{{ completionModal.phaseName }}</div>
            </div>
            <div class="completion-progress">
              <div class="completion-progress-bar"></div>
            </div>
          </div>
        </div>
      </Transition>

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

// 任务完成弹窗
const completionModal = ref({
  visible: false,
  stepName: '',
  phaseName: '',
  timer: null as ReturnType<typeof setTimeout> | null
})

function showCompletionModal(stepName: string, phaseName: string) {
  if (completionModal.value.timer) {
    clearTimeout(completionModal.value.timer)
  }
  completionModal.value.visible = true
  completionModal.value.stepName = stepName
  completionModal.value.phaseName = phaseName
  completionModal.value.timer = setTimeout(() => {
    completionModal.value.visible = false
  }, 3000)
}

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

const parentStepIds = computed(() => {
  const ids = new Set<number>()
  for (const s of steps.value) {
    if (s.parent_step_id) ids.add(s.parent_step_id)
  }
  return ids
})

const isParentStep = (s: StepInstance) => parentStepIds.value.has(s.id)

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
const maxVisibleLogs = ref(8)
const LOG_ROW_H = 28
const logContainerRef = ref<HTMLElement | null>(null)

function updateMaxVisibleLogs() {
  const el = logContainerRef.value
  if (el) {
    const h = el.clientHeight
    maxVisibleLogs.value = Math.max(3, Math.floor(h / LOG_ROW_H))
  }
}

const displayLogs = computed(() => logs.value.slice(0, maxVisibleLogs.value))

// ======== 阶段 Tab ========

// 找到包含 running 步骤的阶段名
function findActivePhaseName(): string | null {
  for (const s of steps.value) {
    if (s.status === 'running') {
      return s.phase || null
    }
  }
  return null
}

// 阶段整体状态：done / running / pending
function getPhaseStatus(phase: TreeNodePhase): string {
  const all = phase.phaseSteps.flatMap(ps => ps.stepNodes)
  if (all.some(s => s.status === 'running' || s.status === 'timeout')) return 'running'
  if (all.every(s => s.status === 'completed' || s.status === 'skipped')) return 'done'
  return 'pending'
}

// 当前选中的阶段索引，默认跟随活动阶段
const selectedPhaseIdx = ref(0)

// 当前阶段的数据（仅展示选中阶段）
const currentPhaseData = computed<TreeNodePhase | null>(() => {
  if (!treeData.value.length) return null
  const idx = selectedPhaseIdx.value
  if (idx < 0 || idx >= treeData.value.length) return treeData.value[0]
  return treeData.value[idx]
})

// 活动阶段自动切换 Tab
watch(() => findActivePhaseName(), (name) => {
  if (!name) return
  const idx = treeData.value.findIndex(p => p.name === name)
  if (idx >= 0) selectedPhaseIdx.value = idx
})

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
let animTime = 0

// 阶段管道布局常量
const PIPE_Y = 30          // 管道节点中心 Y
const PIPE_NODE_W = 130    // 管道节点宽度
const PIPE_NODE_H = 48     // 管道节点高度（增高以提升可读性）
const PIPE_ARROW_LEN = 36  // 箭头长度（节点间距留空）
const TREE_OFFSET_Y = 100  // 子树区域起始 Y（为增高管道节点留空间）

// ======== 子阶段轮式布局常量 ========
const LEFT_COL_RATIO = 0.18   // 左侧列占比
const COL_GAP = 16            // 左右列间距
const LEFT_ROW_H = 48         // 左侧已完成子阶段行高
const NEXT_ROW_H = 48         // 左侧"下一个"行高
const STEP_ROW_H = 36         // 右侧步骤行高（增大以提升可读性）
const PS_ROW_H = 48           // 左侧最近已完成/下一个行高
const STEP_ROW_MAX_W = 500    // 步骤行最大宽度
const PS_HEADER_H = 56        // 右侧当前子阶段标题区高度
const PS_MIN_W = 140          // 子阶段标题最小宽度
const PS_MAX_W = 500          // 子阶段标题最大宽度
const PS_PAD = 28             // 子阶段标题内边距

// 过渡动画状态
let transitionProgress = 1    // 直接显示，不做渐变
let prevCurrentPSIdx = -1     // 上一帧的当前子阶段索引
const TRANSITION_SPEED = 0.03 // 每帧推进量

// 子阶段的状态
function getPhaseStepStatus(ps: TreeNodePhaseStep): string {
  if (ps.stepNodes.some(s => s.status === 'running' || s.status === 'timeout')) return 'running'
  if (ps.stepNodes.every(s => s.status === 'completed' || s.status === 'skipped')) return 'done'
  return 'pending'
}

// 已完成子阶段的耗时
function getPhaseStepElapsed(ps: TreeNodePhaseStep): string | null {
  const starts = ps.stepNodes.map(s => s.start_time).filter(Boolean) as string[]
  const ends = ps.stepNodes.map(s => s.end_time).filter(Boolean) as string[]
  if (!starts.length || !ends.length) return null
  const minStart = Math.min(...starts.map(t => new Date(t).getTime()))
  const maxEnd = Math.max(...ends.map(t => new Date(t).getTime()))
  const diffMs = Math.max(0, maxEnd - minStart)
  const totalSec = Math.floor(diffMs / 1000)
  const h = Math.floor(totalSec / 3600)
  const m = Math.floor((totalSec % 3600) / 60)
  const s = totalSec % 60
  if (h > 0) return `${h}h${m}m`
  if (m > 0) return `${m}m${s}s`
  return `${s}s`
}

// 三区子阶段数据：已完成列表、当前触发的、下一个待触发
const wheelData = computed(() => {
  const phase = currentPhaseData.value
  if (!phase) return { completed: [], current: null, next: null }

  const allPS = phase.phaseSteps
  const completed: TreeNodePhaseStep[] = []
  let current: TreeNodePhaseStep | null = null
  let next: TreeNodePhaseStep | null = null

  // 找到当前触发的子阶段（running/timeout）
  for (const ps of allPS) {
    if (ps.stepNodes.some(s => s.status === 'running' || s.status === 'timeout')) {
      current = ps
      break
    }
  }

  // 已完成 = 所有 done 子阶段
  for (const ps of allPS) {
    if (getPhaseStepStatus(ps) === 'done') completed.push(ps)
  }

  // 下一个 = current 之后第一个 pending
  if (current) {
    const idx = allPS.indexOf(current)
    for (let i = idx + 1; i < allPS.length; i++) {
      if (getPhaseStepStatus(allPS[i]) === 'pending') { next = allPS[i]; break }
    }
  } else if (completed.length > 0 && !current) {
    // 无 running 时（如演练已完成），下一个 = 第一个还没完成的
    const lastDoneIdx = allPS.indexOf(completed[completed.length - 1])
    for (let i = lastDoneIdx + 1; i < allPS.length; i++) {
      if (getPhaseStepStatus(allPS[i]) === 'pending') { next = allPS[i]; break }
    }
  } else if (!completed.length && !current) {
    // 全部 pending，下一个 = 第一个子阶段
    next = allPS[0]
  }

  return { completed, current, next }
})

// 当前子阶段索引（用于过渡动画检测）
const currentPSIdx = computed(() => {
  const wd = wheelData.value
  if (!wd.current) return -1
  const phase = currentPhaseData.value
  if (!phase) return -1
  return phase.phaseSteps.indexOf(wd.current)
})

// 检测子阶段切换 → 触发过渡动画
watch(currentPSIdx, (newIdx) => {
  if (prevCurrentPSIdx >= 0 && newIdx !== prevCurrentPSIdx) {
    transitionProgress = 1  // 切换阶段时直接显示
  }
  prevCurrentPSIdx = newIdx
})

// 用 ctx 测量文字宽度
function calcNodeWidth(ctx: CanvasRenderingContext2D, label: string, extraText?: string): number {
  ctx.font = 'bold 13px "PingFang SC", "Microsoft YaHei", sans-serif'
  let textW = ctx.measureText(label).width
  if (extraText) {
    ctx.font = '11px "PingFang SC", sans-serif'
    textW = Math.max(textW, ctx.measureText(extraText).width)
  }
  return Math.min(PS_MAX_W, Math.max(PS_MIN_W, textW + PS_PAD * 2))
}

function initCanvas() {
  const canvas = flowCanvasRef.value
  if (!canvas) return
  const parent = canvas.parentElement
  if (!parent) return
  const dpr = window.devicePixelRatio || 1
  const rect = parent.getBoundingClientRect()
  const w = rect.width
  const h = Math.max(rect.height, PIPE_Y + PIPE_NODE_H / 2 + 20 + rect.height * 0.8)
  canvas.width = w * dpr
  canvas.height = h * dpr
  canvas.style.width = w + 'px'
  canvas.style.height = h + 'px'
}

// 内容高度现在由画布区域固定（不再动态展开）
function calcContentHeight(): number {
  return 800  // 画布由 flow-area 的 overflow-y:auto 处理
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

// 阶段间的箭头颜色
function arrowColor(fromPhase: TreeNodePhase): string {
  const status = getPhaseStatus(fromPhase)
  if (status === 'done') return '#38BDF8'
  if (status === 'running') return '#67E8F9'
  return '#3A5A7A'
}

function drawFlowTree() {
  const canvas = flowCanvasRef.value
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  const dpr = window.devicePixelRatio || 1
  const canvasW = canvas.width / dpr
  ctx.clearRect(0, 0, canvas.width, canvas.height)
  ctx.save()
  ctx.scale(dpr, dpr)

  const data = treeData.value
  const phase = currentPhaseData.value
  if (!data.length || !phase) { ctx.restore(); return }

  // ========== 顶部阶段管道 ==========

  const n = data.length
  // 均匀铺开：每个节点中心 X = (i + 0.5) * (canvasW / n)
  const spacing = canvasW / n
  const nodeCenters: { x: number; y: number }[] = []

  data.forEach((p, i) => {
    const cx = (i + 0.5) * spacing
    const cy = PIPE_Y
    nodeCenters.push({ x: cx, y: cy })
    const status = getPhaseStatus(p)
    const isSelected = selectedPhaseIdx.value === i

    // 绘制阶段节点
    drawPipelineNode(ctx, cx, cy, p.name, status, isSelected)
  })

  // 阶段间箭头
  for (let i = 0; i < n - 1; i++) {
    const fromX = nodeCenters[i].x + PIPE_NODE_W / 2
    const fromY = nodeCenters[i].y
    const toX = nodeCenters[i + 1].x - PIPE_NODE_W / 2
    const toY = nodeCenters[i + 1].y
    const color = arrowColor(data[i])
    drawArrow(ctx, fromX, fromY, toX, toY, color)
  }

  // ========== 下方：子阶段轮式三区布局 ==========

  const wd = wheelData.value
  const areaTop = TREE_OFFSET_Y
  const areaH = canvas.height / dpr - areaTop - 10
  const leftW = canvasW * LEFT_COL_RATIO
  const rightX = leftW + COL_GAP
  const rightW = canvasW - leftW - COL_GAP - 10

  // 推进过渡动画
  if (transitionProgress < 1) {
    transitionProgress = Math.min(1, transitionProgress + TRANSITION_SPEED)
  }
  const tp = transitionProgress  // 0=刚切换, 1=稳定

  // ===== 左侧列 =====
  drawLeftColumn(ctx, 0, areaTop, leftW, areaH, wd, tp)

  // ===== 右侧详情区 =====
  drawRightArea(ctx, rightX, areaTop, rightW, areaH, wd, tp)

  ctx.restore()
}

// 左侧列：已完成计数 → 最近已完成 → 当前名称 → 待执行计数 → 下一个
function drawLeftColumn(ctx: CanvasRenderingContext2D, x: number, yTop: number, w: number, h: number, wd: { completed: TreeNodePhaseStep[], current: TreeNodePhaseStep | null, next: TreeNodePhaseStep | null }, tp: number) {
  const pad = 10
  const phase = currentPhaseData.value
  const totalPS = phase ? phase.phaseSteps.length : 0
  const doneCount = wd.completed.length
  const remaining = totalPS - doneCount - (wd.current ? 1 : 0)

  // 背景
  ctx.save()
  ctx.fillStyle = 'rgba(5, 15, 30, 0.5)'
  roundRect(ctx, x, yTop, w, h, 6)
  ctx.fill()
  ctx.strokeStyle = 'rgba(0, 150, 255, 0.12)'
  ctx.lineWidth = 1
  roundRect(ctx, x, yTop, w, h, 6)
  ctx.stroke()
  ctx.restore()

  let cursorY = yTop + pad

  // 上方计数：已完成 N
  ctx.save()
  ctx.fillStyle = '#38BDF8'
  ctx.font = 'bold 20px "PingFang SC", sans-serif'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText(`${doneCount}`, x + w / 2, cursorY + 10)
  ctx.fillStyle = '#4A90B0'
  ctx.font = '10px "PingFang SC", sans-serif'
  ctx.fillText('已完成', x + w / 2, cursorY + 26)
  ctx.restore()
  cursorY += 38

  // 最近已完成子阶段（紧跟已完成计数下方）
  if (wd.completed.length > 0) {
    const lastDone = wd.completed[wd.completed.length - 1]
    const elapsed = getPhaseStepElapsed(lastDone)
    drawRecentRow(ctx, x + pad, cursorY + PS_ROW_H / 2, w - pad * 2, lastDone.name, elapsed, 'done')
    cursorY += PS_ROW_H + 6
  }

  // 当前子阶段名称 — 垂直居中（中间主区域）
  if (wd.current) {
    const elapsedStr = getPhaseStepElapsed(wd.current)
    drawCurrentName(ctx, x, yTop + h * 0.45, w, wd.current.name, elapsedStr, tp)
  }

  // 下方区域：待执行计数 + 下一个（固定在底部）
  const bottomStart = yTop + h - 38 - (wd.next ? PS_ROW_H + 4 : 0) - pad

  // 待执行计数
  ctx.save()
  ctx.fillStyle = '#6A8AAA'
  ctx.font = 'bold 20px "PingFang SC", sans-serif'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText(`${remaining}`, x + w / 2, bottomStart + 10)
  ctx.fillStyle = '#5A7A9A'
  ctx.font = '10px "PingFang SC", sans-serif'
  ctx.fillText('待执行', x + w / 2, bottomStart + 26)
  ctx.restore()

  // 下一个待触发子阶段
  if (wd.next) {
    drawRecentRow(ctx, x + pad, bottomStart + 38 + PS_ROW_H / 2, w - pad * 2, wd.next.name, null, 'pending')
  }
}

// 当前子阶段名称（左栏垂直居中，大字发光）
function drawCurrentName(ctx: CanvasRenderingContext2D, x: number, cy: number, w: number, name: string, elapsed: string | null, tp: number) {
  ctx.save()
  const opacity = Math.min(1, tp * 1.5)
  ctx.globalAlpha = opacity

  const pulse = Math.sin(animTime * 0.04) * 0.2 + 0.8
  ctx.shadowColor = 'rgba(103, 232, 249, ' + (0.6 * pulse) + ')'
  ctx.shadowBlur =16 * pulse

  // 名称（大字，居中，可两行）
  ctx.fillStyle = '#67E8F9'
  ctx.font = 'bold 16px "PingFang SC", "Microsoft YaHei", sans-serif'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'

  // 计算单行能放多少字
  ctx.font = 'bold 16px "PingFang SC", "Microsoft YaHei", sans-serif'
  const singleLineW = ctx.measureText(name).width
  const maxW = w - 10

  if (singleLineW > maxW && name.length > 8) {
    // 分两行
    const mid = Math.ceil(name.length * 0.5)
    ctx.fillText(name.slice(0, mid), x + w / 2, cy - 8)
    ctx.fillText(name.slice(mid), x + w / 2, cy + 10)
  } else {
    ctx.fillText(name, x + w / 2, cy - 4)
  }
  ctx.shadowBlur = 0

  // 实时耗时标签：已耗时 x min（自增）
  ctx.font = '11px "PingFang SC", sans-serif'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  if (elapsed) {
    ctx.fillStyle = '#4A90B0'
    ctx.fillText(`已耗时 ${elapsed}`, x + w / 2, cy + (singleLineW > maxW ? 26 : 12))
  } else {
    // 实时计算耗时：从最早 start_time 到当前
    const currentPS = wheelData.value.current
    if (currentPS) {
      const starts = currentPS.stepNodes.map(s => s.start_time).filter(Boolean) as string[]
      if (starts.length) {
        const minStart = Math.min(...starts.map(t => new Date(t).getTime()))
        const nowMs = Date.now()
        const diffMs = Math.max(0, nowMs - minStart)
        const totalSec = Math.floor(diffMs / 1000)
        const m = Math.floor(totalSec / 60)
        const s = totalSec % 60
        ctx.fillStyle = '#4ADE80'
        ctx.fillText(`已耗时 ${m}m${s}s`, x + w / 2, cy + (singleLineW > maxW ? 26 : 12))
      } else {
        ctx.fillStyle = '#4ADE80'
        ctx.fillText('执行中', x + w / 2, cy + (singleLineW > maxW ? 26 : 12))
      }
    }
  }

  ctx.globalAlpha = 1
  ctx.restore()
}

// 最近已完成/下一个行（一样大，两行显示，2行放不下省略）
function drawRecentRow(ctx: CanvasRenderingContext2D, x: number, cy: number, w: number, name: string, elapsed: string | null, kind: 'done' | 'pending') {
  const rowH = PS_ROW_H - 4
  const r = 4
  ctx.save()

  // 背景
  if (kind === 'done') {
    ctx.fillStyle = 'rgba(0, 25, 45, 0.5)'
    ctx.strokeStyle = 'rgba(0, 180, 255, 0.25)'
  } else {
    ctx.fillStyle = 'rgba(10, 22, 40, 0.7)'
    ctx.strokeStyle = 'rgba(0, 120, 200, 0.2)'
  }
  ctx.lineWidth = 1
  roundRect(ctx, x, cy - rowH / 2, w, rowH, r)
  ctx.fill()
  ctx.stroke()

  // 小圆点图标（已完成蓝色，待执行灰色）
  const dotR = 4
  const dotX = x + 10
  ctx.beginPath()
  ctx.arc(dotX, cy - 4, dotR, 0, Math.PI * 2)
  ctx.fillStyle = kind === 'done' ? '#38BDF8' : '#4A5568'
  ctx.fill()
  if (kind === 'done') {
    ctx.shadowColor = '#38BDF8'
    ctx.shadowBlur = 3
    ctx.beginPath()
    ctx.arc(dotX, cy - 4, dotR, 0, Math.PI * 2)
    ctx.fill()
    ctx.shadowBlur = 0
  }

  // 名称（可两行，2行放不下省略）
  ctx.fillStyle = kind === 'done' ? '#8AC0E0' : '#6A8AAA'
  ctx.font = '12px "PingFang SC", sans-serif'
  ctx.textAlign = 'left'
  ctx.textBaseline = 'middle'

  const nameX = x + 20
  const nameW = w - 24
  const maxCharsPerLine = Math.floor(nameW / 12)  // 大约每字12px宽度
  const maxLines = elapsed ? 1 : 2  // 有耗时只能放1行名称

  if (name.length > maxCharsPerLine * maxLines) {
    if (maxLines >= 2 && name.length > maxCharsPerLine) {
      const line1 = name.slice(0, maxCharsPerLine - 1)
      const line2 = name.slice(maxCharsPerLine - 1, maxCharsPerLine * 2 - 2)
      ctx.fillText(line1, nameX, cy - 4)
      ctx.fillText(line2.length > maxCharsPerLine - 1 ? line2.slice(0, maxCharsPerLine - 2) + '..' : line2, nameX, cy + 8)
    } else {
      ctx.fillText(name.slice(0, maxCharsPerLine - 1) + '..', nameX, cy - 4)
    }
  } else if (name.length > maxCharsPerLine) {
    const line1 = name.slice(0, maxCharsPerLine)
    ctx.fillText(line1, nameX, cy - 4)
    ctx.fillText(name.slice(maxCharsPerLine), nameX, cy + 8)
  } else {
    ctx.fillText(name, nameX, cy - 4)
  }

  // 耗时（靠右）
  if (elapsed) {
    ctx.fillStyle = '#4A90B0'
    ctx.font = '10px "PingFang SC", sans-serif'
    ctx.textAlign = 'right'
    ctx.fillText(elapsed, x + w - 4, cy + 8)
  }

  ctx.restore()
}

// 右侧详情区：仅步骤列表（标题已移到左栏）
function drawRightArea(ctx: CanvasRenderingContext2D, x: number, yTop: number, w: number, h: number, wd: { completed: TreeNodePhaseStep[], current: TreeNodePhaseStep | null, next: TreeNodePhaseStep | null }, tp: number) {
  ctx.save()

  // 背景
  ctx.fillStyle = 'rgba(5, 15, 30, 0.3)'
  roundRect(ctx, x, yTop, w, h, 6)
  ctx.fill()

  const currentPS = wd.current
  if (!currentPS) {
    ctx.fillStyle = '#5A7A9A'
    ctx.font = '14px "PingFang SC", sans-serif'
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    const msg = wd.completed.length ? '当前阶段已全部完成' : '等待启动'
    ctx.fillText(msg, x + w / 2, yTop + h / 2)
    ctx.restore()
    return
  }

  // 过渡动画
  const opacity = Math.min(1, tp * 1.5)
  ctx.globalAlpha = opacity

  // 步骤列表（区分顶层步骤和子步骤缩进）
  const pad = 16
  const rowW = Math.min(STEP_ROW_MAX_W, w - pad * 2)
  const SUB_INDENT = 24  // 子步骤缩进量

  // 按序排列步骤：顶层在前，子步骤在后（带缩进）
  const topLevel = currentPS.stepNodes.filter(s => !s.parent_step_id)
  const subSteps = currentPS.stepNodes.filter(s => s.parent_step_id)
  const orderedSteps = [...topLevel, ...subSteps]

  const maxVisible = Math.floor((h - pad * 2) / STEP_ROW_H)
  const steps = orderedSteps.slice(0, maxVisible)

  let rowY = yTop + pad + STEP_ROW_H / 2
  steps.forEach((step) => {
    const isSub = !!step.parent_step_id
    const stepX = isSub ? x + pad + SUB_INDENT : x + pad
    const stepRowW = isSub ? rowW - SUB_INDENT : rowW
    drawStepRow(ctx, stepX, rowY, step, stepRowW)
    rowY += STEP_ROW_H
  })

  // 溢出提示
  if (orderedSteps.length > maxVisible) {
    ctx.fillStyle = '#5A7A9A'
    ctx.font = '10px "PingFang SC", sans-serif'
    ctx.textAlign = 'left'
    ctx.fillText(`+${orderedSteps.length - maxVisible} 更多步骤`, x + pad, rowY + 4)
  }

  ctx.globalAlpha = 1
  ctx.restore()
}

// 管道上的阶段节点
function drawPipelineNode(ctx: CanvasRenderingContext2D, cx: number, cy: number, label: string, status: string, selected: boolean) {
  const w = PIPE_NODE_W
  const h = PIPE_NODE_H
  const r = 8
  ctx.save()

  if (selected) {
    const pulse = Math.sin(animTime * 0.04) * 0.2 + 0.8
    ctx.shadowColor = status === 'running'
      ? 'rgba(103, 232, 249, ' + (0.6 * pulse) + ')'
      : 'rgba(56, 189, 248, ' + (0.4 * pulse) + ')'
    ctx.shadowBlur = 12 * pulse
  } else if (status === 'done') {
    ctx.shadowColor = 'rgba(56, 189, 248, 0.2)'
    ctx.shadowBlur = 4
  } else if (status === 'running') {
    ctx.shadowColor = 'rgba(103, 232, 249, 0.15)'
    ctx.shadowBlur = 6
  } else {
    ctx.shadowColor = 'rgba(56, 189, 248, 0.08)'
    ctx.shadowBlur = 2
  }

  const bg = selected
    ? 'rgba(0, 40, 60, 0.95)'
    : status === 'done' ? 'rgba(0, 25, 45, 0.85)'
    : status === 'running' ? 'rgba(0, 35, 55, 0.9)'
    : 'rgba(10, 22, 40, 0.9)'
  ctx.fillStyle = bg

  const border = selected
    ? (status === 'running' ? '#67E8F9' : '#38BDF8')
    : status === 'done' ? 'rgba(56, 189, 248, 0.5)'
    : status === 'running' ? 'rgba(103, 232, 249, 0.4)'
    : 'rgba(56, 189, 248, 0.35)'
  ctx.strokeStyle = border
  ctx.lineWidth = selected ? 2 : 1

  roundRect(ctx, cx - w / 2, cy - h / 2, w, h, r)
  ctx.fill()
  ctx.stroke()
  ctx.shadowBlur = 0

  ctx.fillStyle = selected
    ? '#67E8F9'
    : status === 'done' ? '#38BDF8'
    : status === 'running' ? '#7DD3FC'
    : '#6A8AAA'
  ctx.font = 'bold 13px "PingFang SC", "Microsoft YaHei", sans-serif'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText(label, cx, cy - 5)

  const statusText = status === 'done' ? '已完成' : status === 'running' ? '执行中' : '待执行'
  ctx.fillStyle = selected
    ? (status === 'running' ? '#4ADE80' : '#38BDF8')
    : status === 'done' ? '#4A90B0'
    : status === 'running' ? '#4ADE80'
    : '#5A7A9A'
  ctx.font = '10px "PingFang SC", sans-serif'
  ctx.fillText(statusText, cx, cy + 10)

  ctx.restore()
}

// 横向箭头（阶段间）
function drawArrow(ctx: CanvasRenderingContext2D, x1: number, y1: number, x2: number, y2: number, color: string) {
  ctx.save()
  ctx.strokeStyle = color
  ctx.lineWidth = 2
  ctx.shadowColor = color
  ctx.shadowBlur = 3

  ctx.beginPath()
  ctx.moveTo(x1, y1)
  ctx.lineTo(x2, y2)
  ctx.stroke()

  const arrowSize = 8
  const angle = Math.atan2(y2 - y1, x2 - x1)
  ctx.fillStyle = color
  ctx.beginPath()
  ctx.moveTo(x2, y2)
  ctx.lineTo(x2 - arrowSize * Math.cos(angle - Math.PI / 6), y2 - arrowSize * Math.sin(angle - Math.PI / 6))
  ctx.lineTo(x2 - arrowSize * Math.cos(angle + Math.PI / 6), y2 - arrowSize * Math.sin(angle + Math.PI / 6))
  ctx.closePath()
  ctx.fill()

  ctx.shadowBlur = 0
  ctx.restore()
}

// 步骤列表行
function drawStepRow(ctx: CanvasRenderingContext2D, x: number, cy: number, step: StepInstance, rowW: number) {
  const isRunning = step.status === 'running'
  const isDone = step.status === 'completed' || step.status === 'skipped'
  const isTimeout = step.status === 'timeout'
  const rowH = STEP_ROW_H - 2
  const r = 3

  ctx.save()

  // 行背景
  if (isRunning) {
    const pulse = Math.sin(animTime * 0.05) * 0.15 + 0.85
    ctx.shadowColor = 'rgba(103, 232, 249, ' + (0.35 * pulse) + ')'
    ctx.shadowBlur = 8 * pulse
    ctx.fillStyle = 'rgba(0, 40, 60, 0.7)'
  } else if (isTimeout) {
    ctx.shadowColor = 'rgba(255, 215, 0, 0.15)'
    ctx.shadowBlur = 3
    ctx.fillStyle = 'rgba(50, 40, 10, 0.4)'
  } else if (isDone) {
    ctx.shadowBlur = 0
    ctx.fillStyle = 'rgba(0, 20, 35, 0.35)'
  } else {
    ctx.shadowBlur = 0
    ctx.fillStyle = 'rgba(8, 18, 30, 0.25)'
  }
  roundRect(ctx, x, cy - rowH / 2, rowW, rowH, r)
  ctx.fill()

  // running/timeout 边框
  if (isRunning) {
    ctx.strokeStyle = '#67E8F9'
    ctx.lineWidth = 1.5
    roundRect(ctx, x, cy - rowH / 2, rowW, rowH, r)
    ctx.stroke()
  } else if (isTimeout) {
    ctx.strokeStyle = 'rgba(255, 215, 0, 0.3)'
    ctx.lineWidth = 1
    roundRect(ctx, x, cy - rowH / 2, rowW, rowH, r)
    ctx.stroke()
  }
  ctx.shadowBlur = 0

  // 状态圆点（放大）
  const dotR = 5
  const dotX = x + 14
  const dotColor = isRunning ? '#67E8F9' : isTimeout ? '#FFD700' : isDone ? '#38BDF8' : '#4A5568'
  ctx.beginPath()
  ctx.arc(dotX, cy, dotR, 0, Math.PI * 2)
  ctx.fillStyle = dotColor
  ctx.fill()
  if (isRunning) {
    ctx.shadowColor = '#67E8F9'
    ctx.shadowBlur = 6
    ctx.beginPath()
    ctx.arc(dotX, cy, dotR, 0, Math.PI * 2)
    ctx.fill()
    ctx.shadowBlur = 0
  }

  // 步骤名称
  const nameX = x + 28
  ctx.fillStyle = isRunning ? '#67E8F9' : isTimeout ? '#FFD700' : isDone ? '#38BDF8' : '#6A8AAA'
  ctx.font = (isRunning ? 'bold ' : '') + '11px "PingFang SC", "Microsoft YaHei", sans-serif'
  ctx.textAlign = 'left'
  ctx.textBaseline = 'middle'
  const name = step.name.length > 22 ? step.name.slice(0, 21) + '..' : step.name
  ctx.fillText(name, nameX, cy)

  // 状态文字（靠右）
  const statusX = x + rowW - 60
  const statusLabel = isRunning ? '执行中' : isTimeout ? '超时' : isDone ? '已完成' : '待执行'
  ctx.fillStyle = isRunning ? '#4ADE80' : isTimeout ? '#FFD700' : isDone ? '#4A90B0' : '#3A4A5A'
  ctx.font = '10px "PingFang SC", sans-serif'
  ctx.fillText(statusLabel, statusX, cy)

  ctx.restore()
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

function redrawCanvas() {
  drawFlowTree()
}

function handleCanvasClick(e: MouseEvent) {
  const canvas = flowCanvasRef.value
  if (!canvas || !treeData.value.length) return
  const rect = canvas.getBoundingClientRect()
  const dpr = window.devicePixelRatio || 1
  const canvasW = canvas.width / dpr
  const x = e.clientX - rect.left
  const y = e.clientY - rect.top

  const data = treeData.value
  const n = data.length
  const spacing = canvasW / n

  // 点击阶段管道节点 → 切换阶段
  for (let i = 0; i < n; i++) {
    const cx = (i + 0.5) * spacing
    const cy = PIPE_Y
    if (Math.abs(x - cx) < PIPE_NODE_W / 2 && Math.abs(y - cy) < PIPE_NODE_H / 2) {
      selectedPhaseIdx.value = i
      nextTick(() => initCanvas())
      return
    }
  }
}

// ======== WebSocket ========

let ws: WebSocket | null = null
let wsReconnectTimer: ReturnType<typeof setTimeout> | null = null
let wsReconnectCount = 0

function connectWS() {
  const id = drillId.value
  if (!id || isNaN(id)) return
  const proto = location.protocol === 'https:' ? 'wss' : 'ws'
  const url = `${proto}://${location.host}/ws/display/${id}?token=${authStore.token}`
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
  const event = msg.event_type || msg.event || msg.type || ''
  const payload = msg.payload || msg.data || msg
  const stepName = payload.step_name || payload.stepName || ''
  const phaseName = payload.phase_name || payload.phaseName || ''

  // 心跳忽略
  if (event === 'ping' || event === 'pong') return

  if (['drill_started', 'drill_paused', 'drill_resumed', 'drill_completed', 'drill_terminated'].includes(event)) {
    scheduleRefresh('drill', 'logs')
    if (event === 'drill_started') addLog('info', '▶', '演练已开始')
    else if (event === 'drill_paused') addLog('warn', '⏸', '演练已暂停')
    else if (event === 'drill_resumed') addLog('info', '▶', '演练已恢复')
    else if (event === 'drill_completed') addLog('info', '✓', '演练已完成')
    else if (event === 'drill_terminated') addLog('error', '⏹', '演练已结束')
  }

  // 步骤事件：增量更新本地数据，不调 API
  if (event.startsWith('step_')) {
    patchLocalStep(payload)
    const phasePrefix = phaseName ? `【${phaseName}】` : ''
    if (event === 'step_started') {
      addLog('info', '●', `${phasePrefix}${stepName} 已开始`)
    } else if (['step_complete', 'step_timeout', 'step_skipped', 'step_issue'].includes(event)) {
      const label = logLabel(event)
      const logType = event === 'step_timeout' ? 'warn' : event === 'step_issue' ? 'error' : 'info'
      addLog(logType, logIcon(event), `${phasePrefix}${stepName} ${label}`)
      if (event === 'step_complete') {
        showCompletionModal(stepName, phaseName)
      }
    }
    return
  }

  if (event === 'timeout_warning') {
    const remaining = payload.remaining_sec || 0
    const phasePrefix = phaseName ? `【${phaseName}】` : ''
    addLog('warn', '⏰', `${phasePrefix}${stepName} 剩余 ${remaining} 秒`)
  }
}

// 增量更新本地步骤数据（不调 API）
function patchLocalStep(payload: any) {
  const stepId = payload.step_id || payload.stepId
  if (!stepId) return

  const newStatus = payload.new_status || payload.newStatus
  if (!newStatus) return

  const idx = steps.value.findIndex((s: StepInstance) => s.id === stepId)
  if (idx === -1) return

  const step = { ...steps.value[idx] }
  step.status = newStatus
  if (payload.start_time) step.start_time = payload.start_time
  if (payload.end_time) step.end_time = payload.end_time
  if (payload.executor) step.assignee_names = payload.executor
  if (payload.comment) step.remark = payload.comment

  const newSteps = [...steps.value]
  newSteps[idx] = step
  steps.value = newSteps
}

function logLabel(event: string): string {
  const map: Record<string, string> = {
    step_started: '已开始',
    step_complete: '已完成',
    step_timeout: '已超时',
    step_skipped: '已跳过',
    step_issue: '异常',
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

let refreshTimer: ReturnType<typeof setTimeout> | null = null
let pendingRefresh: Set<'drill' | 'steps' | 'logs'> = new Set()

function scheduleRefresh(...kinds: ('drill' | 'steps' | 'logs')[]) {
  for (const k of kinds) pendingRefresh.add(k)
  if (refreshTimer) return
  refreshTimer = setTimeout(async () => {
    refreshTimer = null
    const tasks: Promise<void>[] = []
    if (pendingRefresh.has('drill')) tasks.push(fetchDrillData())
    if (pendingRefresh.has('steps')) tasks.push(fetchSteps())
    if (pendingRefresh.has('logs')) tasks.push(fetchLogs())
    pendingRefresh.clear()
    await Promise.all(tasks)
  }, 500)
}

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
    const items = logData.slice(0, 8).map((l: Record<string, unknown>) => {
      const action = (l.Action || l.action || '') as string
      const content = (l.Content || l.content || '') as string
      const msg = content || action
      const logType = action === 'timeout' || action === 'pause' ? 'warn'
        : action === 'issue' || action === 'terminate' ? 'error'
        : 'info'
      return {
        id: (l.ID || l.id) as number,
        time: fmtTime((l.CreatedAt || l.created_at) as string),
        icon: logType === 'error' ? '⚠' : logType === 'warn' ? '⏸' : action === 'start' ? '▶' : '●',
        type: logType,
        msg,
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
  try { await drillApi.start(drillId.value); scheduleRefresh('drill', 'steps') } catch { /* */ }
}
async function handlePause() {
  try { await drillApi.pause(drillId.value); scheduleRefresh('drill') } catch { /* */ }
}
async function handleResume() {
  try { await drillApi.resume(drillId.value); scheduleRefresh('drill') } catch { /* */ }
}
async function handleTerminate() {
  try { await drillApi.terminate(drillId.value); scheduleRefresh('drill') } catch { /* */ }
}

function canActOn(step: StepInstance): boolean {
  if (!canOperateTask.value) return false
  const taskTeam = step.executor_team || ''
  if (!taskTeam) return true
  return userDept.value === taskTeam
}

async function skipTask(step: StepInstance) {
  try {
    await drillApi.skipStep(drillId.value, step.step_template_id || step.id, '指挥跳过')
    scheduleRefresh('steps', 'drill')
  } catch { /* */ }
}

async function forceCompleteTask(step: StepInstance) {
  try {
    await drillApi.forceCompleteStep(drillId.value, step.step_template_id || step.id, '指挥强制完成')
    scheduleRefresh('steps', 'drill')
  } catch { /* */ }
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
    updateMaxVisibleLogs()
  })
  connectWS()
  timerInterval = setInterval(() => {
    updateTimer()
    redrawCanvas()
  }, 1000)
  pollingTimer = setInterval(() => {
    if (!wsConnected.value) scheduleRefresh('drill')
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
  updateMaxVisibleLogs()
}

// 侦听 steps 变化重绘 Canvas（不重设尺寸，避免闪黑）
watch([steps, () => instance.value?.progress_pct], () => {
  nextTick(() => {
    drawFlowTree()
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
  background: #0B1121;
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
.screen-root .flow-area::-webkit-scrollbar-thumb { background: rgba(56, 189, 248, 0.3); border-radius: 3px; }

.bg-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(56, 189, 248, 0.04) 1px, transparent 1px),
    linear-gradient(90deg, rgba(56, 189, 248, 0.04) 1px, transparent 1px);
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
  background: #1E40AF;
}

.bg-glow-br {
  bottom: -150px;
  right: -100px;
  background: #38BDF8;
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
  border-top-color: #38BDF8;
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
  background: rgba(56, 189, 248, 0.15);
  border: 1px solid rgba(56, 189, 248, 0.4);
  color: #38BDF8;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.btn-retry:hover {
  background: rgba(56, 189, 248, 0.25);
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
  background: linear-gradient(90deg, transparent, rgba(56, 189, 248, 0.3) 20%, rgba(103, 232, 249, 0.5) 50%, rgba(56, 189, 248, 0.3) 80%, transparent);
}

.tb-left {
  display: flex;
  align-items: center;
  min-width: 90px;
}

.elapsed-time {
  font-size: 24px;
  font-weight: 700;
  color: #67E8F9;
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
.st-running  { color: #4ADE80; background: rgba(74, 222, 128, 0.1); }
.st-paused   { color: #FFD700; background: rgba(255, 215, 0, 0.1); }
.st-completed { color: #38BDF8; background: rgba(56, 189, 248, 0.1); }
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
  background: linear-gradient(90deg, #0066FF, #67E8F9);
  transition: width 0.8s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 0 8px rgba(56, 189, 248, 0.4);
}

.progress-glow {
  position: absolute;
  top: -4px;
  width: 14px;
  height: 14px;
  background: rgba(103, 232, 249, 0.8);
  border-radius: 50%;
  transform: translateX(-50%);
  box-shadow: 0 0 12px rgba(103, 232, 249, 0.6);
  transition: left 0.8s cubic-bezier(0.4, 0, 0.2, 1);
}

.progress-text {
  font-size: 16px;
  font-weight: 700;
  color: #67E8F9;
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
  border-color: rgba(56, 189, 248, 0.5);
  color: #38BDF8;
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
  background: rgba(74, 222, 128, 0.15);
  border-color: rgba(74, 222, 128, 0.4);
  color: #4ADE80;
}

.btn-ctrl.start:hover {
  background: rgba(74, 222, 128, 0.3);
  box-shadow: 0 0 8px rgba(74, 222, 128, 0.3);
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
  background: rgba(74, 222, 128, 0.15);
  border-color: rgba(74, 222, 128, 0.4);
  color: #4ADE80;
}

.btn-ctrl.resume:hover {
  background: rgba(74, 222, 128, 0.3);
  box-shadow: 0 0 8px rgba(74, 222, 128, 0.3);
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
  color: #38BDF8;
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

.lg-done { background: #38BDF8; box-shadow: 0 0 4px #38BDF8; }
.lg-running { background: #67E8F9; box-shadow: 0 0 8px #67E8F9; animation: pulse-dot 1.5s ease-in-out infinite; }
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
  border-left: 1px solid rgba(56, 189, 248, 0.15);
  backdrop-filter: blur(8px);
}

.rp-block {
  display: flex;
  flex-direction: column;
  border-bottom: 1px solid rgba(56, 189, 248, 0.1);
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
  color: #38BDF8;
  background: rgba(0, 20, 50, 0.5);
}

.rp-ico {
  flex-shrink: 0;
  color: #38BDF8;
}

.rp-badge {
  margin-left: auto;
  background: rgba(0, 180, 255, 0.2);
  color: #67E8F9;
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

.task-row.task-timeout {
  background: rgba(255, 68, 68, 0.08);
  border-left-color: #FF4444;
}

.task-row:hover {
  background: rgba(56, 189, 248, 0.06);
}

.task-btn {
  flex-shrink: 0;
  width: 22px;
  height: 22px;
  border-radius: 3px;
  border: 1px solid rgba(56, 189, 248, 0.3);
  background: rgba(0, 25, 50, 0.5);
  color: #8AC0E0;
  font-size: 11px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
  line-height: 1;
}

.task-btn:hover {
  background: rgba(56, 189, 248, 0.15);
}

.task-btn-skip:hover {
  color: #FFD700;
  border-color: rgba(255, 215, 0, 0.5);
}

.task-btn-done:hover {
  color: #4ADE80;
  border-color: rgba(74, 222, 128, 0.5);
}

.task-status {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
  background: #4A5568;
}

.ts-running { background: #4ADE80; }
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

.log-info .log-msg { color: #38BDF8; }
.log-info .log-icon { color: #38BDF8; }
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
  color: #67E8F9;
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
  background: linear-gradient(90deg, #0066FF, #67E8F9);
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
  border-top: 1px solid rgba(56, 189, 248, 0.15);
  z-index: 1;
  padding: 0 28px;
}

.bb-line {
  position: absolute;
  top: 0;
  left: 28px;
  right: 28px;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(56, 189, 248, 0.3) 20%, rgba(103, 232, 249, 0.3) 50%, rgba(56, 189, 248, 0.3) 80%, transparent);
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

/* 完成弹窗 */
.completion-modal {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
}

.completion-modal-content {
  background: rgba(15, 23, 42, 0.95);
  border: 1px solid rgba(74, 222, 128, 0.3);
  border-radius: 16px;
  padding: 32px 48px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16px;
  box-shadow: 0 0 40px rgba(74, 222, 128, 0.2);
  min-width: 300px;
}

.completion-icon {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: rgba(74, 222, 128, 0.15);
  border: 2px solid #4ADE80;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  color: #4ADE80;
}

.completion-text {
  text-align: center;
}

.completion-title {
  font-size: 20px;
  font-weight: 600;
  color: #F8FAFC;
  margin-bottom: 8px;
}

.completion-step {
  font-size: 16px;
  color: #4ADE80;
}

.completion-phase {
  font-size: 12px;
  color: #64748B;
  margin-top: 4px;
}

.completion-progress {
  width: 100%;
  height: 3px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 2px;
  overflow: hidden;
}

.completion-progress-bar {
  height: 100%;
  background: #4ADE80;
  animation: progress-shrink 3s linear forwards;
}

@keyframes progress-shrink {
  from { width: 100%; }
  to { width: 0%; }
}

/* 弹窗动画 */
.modal-enter-active {
  transition: all 0.3s ease-out;
}

.modal-leave-active {
  transition: all 0.2s ease-in;
}

.modal-enter-from {
  opacity: 0;
  transform: scale(0.9);
}

.modal-leave-to {
  opacity: 0;
  transform: scale(0.95);
}
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