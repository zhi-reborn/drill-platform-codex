<template>
  <div class="screen-root cyber-command-screen">
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
      <div class="cyber-bg cyber-bg-grid" />
      <div class="cyber-bg cyber-bg-beams" />
      <div class="cyber-bg cyber-bg-scan" />

      <header class="command-header">
        <div class="header-scanline" />
        <div class="header-title-shell">
          <h1 class="command-title">应急处置指挥中心</h1>
        </div>
        <div class="header-meta">
          <div class="progress-console">
            <span class="drill-name-tag" :title="instance?.name">{{ instance?.name || '演练中' }}</span>
            <span class="console-sep" />
            <span class="system-label">演练进度</span>
            <span class="system-time">{{ liveProgressPct }}%</span>
          </div>
          <button class="btn-fullscreen" @click="toggleFullscreen" title="全屏模式">
            <el-icon><FullScreen /></el-icon>
          </button>
        </div>
        <div v-if="canControl" class="control-strip">
          <button v-if="instance?.status === 'pending'" class="control-btn good" @click="handleStart">开始</button>

        </div>
      </header>

      <main class="command-main">
        <div class="main-radar-sweep" />
        <section class="phase-card-strip">
          <article v-for="(phase, index) in phaseCards" :key="phase.name" class="phase-card" :class="['is-' + phase.status, { active: phase.active }]">
            <div class="phase-card-grid" />
            <div class="phase-accent" />
            <div class="phase-head">
              <h2>阶段{{ index + 1 }} {{ phase.name }}</h2>
              <span class="phase-status">{{ phase.statusText }}</span>
            </div>
            <div class="phase-segments" :aria-label="phase.statusText">
              <span v-for="seg in phase.segmentCount" :key="seg" :class="{ filled: seg <= phase.filledSegments }" />
            </div>
            <div class="phase-stats">
              <span><b>{{ phase.completedPhaseSteps }}</b>/{{ phase.totalPhaseSteps }}<em>环节</em></span>
              <span><b>{{ phase.completedSteps }}</b>/{{ phase.totalSteps }}<em>步骤</em></span>
            </div>
          </article>
        </section>

        <section class="flow-board">
          <div class="flow-board-grid" />
          <div class="flow-row flow-row-top">
            <div v-for="(node, index) in topFlowNodes" :key="node.id" class="flow-node-wrap">
              <div class="flow-node" :class="'is-' + node.status">
                <span class="node-tag">{{ node.status === 'done' ? '✓ ' + node.name : node.name }}</span>
              </div>
              <span v-if="index < topFlowNodes.length - 1" class="flow-arrow right" />
              <span v-else class="flow-arrow turn" />
            </div>
          </div>
          <div class="flow-row flow-row-bottom">
            <div v-for="(node, index) in bottomFlowNodes" :key="node.id" class="flow-node-wrap">
              <span v-if="index > 0" class="flow-arrow left" />
              <div class="flow-node" :class="'is-' + node.status">
                <span class="node-tag">{{ node.status === 'done' ? '✓ ' + node.name : node.name }}</span>
              </div>
            </div>
          </div>
        </section>

        <section class="execution-section">
          <div class="execution-title">
            <h2>执行中步骤</h2>
            <div class="execution-signal">
              <span class="signal-bars"><i /><i /><i /></span>
              <span :class="{ live: wsConnected }">{{ wsConnected ? '实时' : '轮询' }}</span>
            </div>
          </div>
          <div class="execution-carousel">
            <div class="exec-col exec-col-running">
              <div class="exec-col-label"><span class="exec-dot running" />进行中</div>
              <div class="exec-col-cards">
                <article v-for="task in runningCards" :key="task.id" class="execution-card is-running">
                  <div class="card-scan" />
                  <div class="task-card-head">
                    <strong>{{ task.name }}</strong>
                    <span>{{ task.statusText }}</span>
                  </div>
                  <div class="task-progress"><div :style="{ width: task.progress + '%' }" /></div>
                  <p>{{ task.phaseText }}</p>
                </article>
                <div v-if="!runningCards.length" class="exec-empty">暂无进行中步骤</div>
              </div>
            </div>
            <div class="exec-divider" />
            <div class="exec-col exec-col-pending">
              <div class="exec-col-label"><span class="exec-dot pending" />待执行</div>
              <div class="exec-col-cards">
                <article v-for="task in pendingCards" :key="task.id" class="execution-card is-pending">
                  <div class="card-scan" />
                  <div class="task-card-head">
                    <strong>{{ task.name }}</strong>
                    <span>{{ task.statusText }}</span>
                  </div>
                  <div class="task-progress"><div :style="{ width: task.progress + '%' }" /></div>
                  <p>{{ task.phaseText }}</p>
                </article>
                <div v-if="!pendingCards.length" class="exec-empty">暂无待执行步骤</div>
              </div>
            </div>
          </div>
        </section>
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

    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRoute } from 'vue-router'
import { FullScreen } from '@element-plus/icons-vue'
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

// 实时进度：优先从 steps 计算，WebSocket 增量更新时自动同步
const liveProgressPct = computed(() => {
  if (!steps.value.length) return instance.value?.progress_pct ?? 0
  const done = steps.value.filter(s => s.status === 'completed' || s.status === 'skipped').length
  return Math.round((done / steps.value.length) * 100)
})

const currentSystemTime = computed(() => {
  const d = new Date(now.value)
  return `${d.getFullYear()}.${pad(d.getMonth() + 1)}.${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
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
  const tasks = steps.value.filter(s => s.status === 'running')
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
  if (all.some(s => s.status === 'running')) return 'running'
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

const phaseCards = computed(() => {
  const activeName = findActivePhaseName()
  const fallbackActiveIdx = treeData.value.findIndex(p => getPhaseStatus(p) === 'running')
  return treeData.value.map((phase, index) => {
    const allSteps = phase.phaseSteps.flatMap(ps => ps.stepNodes)
    const status = getPhaseStatus(phase)
    const completedSteps = allSteps.filter(s => s.status === 'completed' || s.status === 'skipped').length
    const totalSteps = allSteps.length || 1
    const completedPhaseSteps = phase.phaseSteps.filter(ps => getPhaseStepStatus(ps) === 'done').length
    const totalPhaseSteps = phase.phaseSteps.length || 1
    const segmentCount = 20
    const filledSegments = Math.round((completedSteps / totalSteps) * segmentCount)
    return {
      name: phase.name,
      status,
      statusText: status === 'done' ? '已完成' : status === 'running' ? '进行中' : '待开始',
      active: activeName ? activeName === phase.name : index === fallbackActiveIdx,
      completedSteps,
      totalSteps,
      completedPhaseSteps,
      totalPhaseSteps,
      segmentCount,
      filledSegments,
      timeText: phaseTimeText(allSteps),
    }
  })
})

const flowNodes = computed(() => {
  let source: TreeNodePhaseStep[] = []
  if (currentPhaseData.value?.phaseSteps.length) {
    const phase = currentPhaseData.value
    source = phase.phaseSteps.filter(ps => ps.name !== phase.name)
  } else {
    source = treeData.value.flatMap(p => p.phaseSteps.filter(ps => ps.name !== p.name))
  }
  return source.slice(0, 12).map((ps, index) => ({
    id: `${ps.name}-${index}`,
    name: ps.name,
    index: index + 1,
    status: getPhaseStepStatus(ps) === 'done' ? 'done' : getPhaseStepStatus(ps) === 'running' ? 'running' : 'pending',
  }))
})

const topFlowNodes = computed(() => flowNodes.value.slice(0, Math.ceil(flowNodes.value.length / 2)))
const bottomFlowNodes = computed(() => flowNodes.value.slice(Math.ceil(flowNodes.value.length / 2)).reverse())

// 过滤阶段节点（phase === phase_step）和环节节点（有子步骤的父步骤）
function isLeafStep(s: StepInstance): boolean {
  if (s.phase && s.phase_step && s.phase === s.phase_step) return false
  if (isParentStep(s)) return false
  return true
}

function mapExecCard(step: StepInstance) {
  const progress = step.status === 'completed' || step.status === 'skipped' ? 100 : step.status === 'running' ? 100 : 0
  return {
    id: step.id,
    name: step.name,
    status: step.status,
    statusText: step.status === 'running' ? '进行中' : step.status === 'pending' ? '待执行' : '已完成',
    progress,
    timeText: step.timeout_minutes ? `${pad(Math.floor(step.timeout_minutes / 60))}:${pad(step.timeout_minutes % 60)}:00` : '01:00:00',
    phaseText: `${step.phase || '当前阶段'} — ${step.phase_step || step.executor_team || '执行任务'}`,
    raw: step,
  }
}

const runningCards = computed(() => runningSteps.value.filter(isLeafStep).slice(0, 8).map(mapExecCard))
const pendingCards = computed(() => steps.value.filter(s => s.status === 'pending' && isLeafStep(s)).slice(0, 8).map(mapExecCard))

function phaseTimeText(phaseSteps: StepInstance[]): string {
  const starts = phaseSteps.map(s => s.start_time).filter(Boolean) as string[]
  const ends = phaseSteps.map(s => s.end_time).filter(Boolean) as string[]
  if (!starts.length) return '--:-- / 21:19'
  const start = new Date(Math.min(...starts.map(t => new Date(t).getTime())))
  const end = ends.length ? new Date(Math.max(...ends.map(t => new Date(t).getTime()))) : new Date(now.value)
  return `${pad(start.getHours())}:${pad(start.getMinutes())} / ${pad(end.getHours())}:${pad(end.getMinutes())}`
}

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
  if (ps.stepNodes.some(s => s.status === 'running')) return 'running'
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

  // 找到当前触发的子阶段
  for (const ps of allPS) {
    if (ps.stepNodes.some(s => s.status === 'running')) {
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
        ctx.fillText('进行中', x + w / 2, cy + (singleLineW > maxW ? 26 : 12))
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

  const statusText = status === 'done' ? '已完成' : status === 'running' ? '进行中' : '待执行'
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
  const rowH = STEP_ROW_H - 2
  const r = 3

  ctx.save()

  // 行背景
  if (isRunning) {
    const pulse = Math.sin(animTime * 0.05) * 0.15 + 0.85
    ctx.shadowColor = 'rgba(103, 232, 249, ' + (0.35 * pulse) + ')'
    ctx.shadowBlur = 8 * pulse
    ctx.fillStyle = 'rgba(0, 40, 60, 0.7)'
  } else if (isDone) {
    ctx.shadowBlur = 0
    ctx.fillStyle = 'rgba(0, 20, 35, 0.35)'
  } else {
    ctx.shadowBlur = 0
    ctx.fillStyle = 'rgba(8, 18, 30, 0.25)'
  }
  roundRect(ctx, x, cy - rowH / 2, rowW, rowH, r)
  ctx.fill()

  // running 边框
  if (isRunning) {
    ctx.strokeStyle = '#67E8F9'
    ctx.lineWidth = 1.5
    roundRect(ctx, x, cy - rowH / 2, rowW, rowH, r)
    ctx.stroke()
  }
  ctx.shadowBlur = 0

  // 状态圆点（放大）
  const dotR = 5
  const dotX = x + 14
  const dotColor = isRunning ? '#67E8F9' : isDone ? '#38BDF8' : '#4A5568'
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
  ctx.fillStyle = isRunning ? '#67E8F9' : isDone ? '#38BDF8' : '#6A8AAA'
  ctx.font = (isRunning ? 'bold ' : '') + '11px "PingFang SC", "Microsoft YaHei", sans-serif'
  ctx.textAlign = 'left'
  ctx.textBaseline = 'middle'
  const name = step.name.length > 22 ? step.name.slice(0, 21) + '..' : step.name
  ctx.fillText(name, nameX, cy)

  // 状态文字（靠右）
  const statusX = x + rowW - 60
  const statusLabel = isRunning ? '进行中' : isDone ? '已完成' : '待执行'
  ctx.fillStyle = isRunning ? '#4ADE80' : isDone ? '#4A90B0' : '#3A4A5A'
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
    } else if (['step_complete', 'step_skipped', 'step_issue'].includes(event)) {
      const label = logLabel(event)
      const logType = event === 'step_issue' ? 'error' : 'info'
      addLog(logType, logIcon(event), `${phasePrefix}${stepName} ${label}`)
      if (event === 'step_complete') {
        showCompletionModal(stepName, phaseName)
      }
    }
    return
  }

  if (event === 'timeout_warning') return
}

// 增量更新本地步骤数据（不调 API）
function patchLocalStep(payload: any) {
  const stepId = payload.step_id || payload.stepId
  if (!stepId) return

  const newStatus = payload.new_status || payload.newStatus
  if (!newStatus) return

  const idx = steps.value.findIndex((s: StepInstance) => s.id === stepId)
  if (idx === -1) {
    // 本地没有该步骤，触发全量刷新
    scheduleRefresh('steps')
    return
  }

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
  /* 统一缩放基准：所有 vw 字号基于此，保持比例一致 */
  font-size: clamp(14px, 0.92vw, 17px);
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

/* ===== 参考图样式：大屏2 ===== */

.cyber-command-screen {
  position: relative;
  margin: 0;
  height: 100vh;
  grid-template-rows: clamp(76px, 8vh, 96px) minmax(0, 1fr);
  background:
    radial-gradient(circle at 50% 45%, rgba(12, 70, 132, 0.44), transparent 34%),
    radial-gradient(circle at 78% 70%, rgba(79, 36, 36, 0.22), transparent 28%),
    linear-gradient(180deg, #071a35 0%, #041024 52%, #020916 100%);
  color: #dce9ff;
  font-family: "Microsoft YaHei", "PingFang SC", sans-serif;
  letter-spacing: 0;
  border: 1px solid rgba(39, 165, 230, 0.45);
  box-shadow: inset 0 0 40px rgba(0, 180, 255, 0.16);
}

.cyber-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.cyber-bg-grid {
  background-image:
    linear-gradient(115deg, transparent 0 13%, rgba(32, 119, 194, 0.14) 13.2%, transparent 13.7%),
    linear-gradient(75deg, transparent 0 18%, rgba(32, 119, 194, 0.12) 18.2%, transparent 18.8%),
    linear-gradient(rgba(65, 167, 244, 0.035) 1px, transparent 1px);
  background-size: 310px 100%, 420px 100%, 100% 5px;
  opacity: 0.72;
}

.cyber-bg-beams {
  background:
    linear-gradient(90deg, rgba(0, 217, 255, 0.16), transparent 18%, transparent 82%, rgba(0, 217, 255, 0.14)),
    radial-gradient(ellipse at 50% 42%, rgba(0, 183, 255, 0.14), transparent 42%);
}

.cyber-bg-scan {
  background: repeating-linear-gradient(180deg, transparent 0 3px, rgba(28, 112, 182, 0.06) 3px 5px);
  mix-blend-mode: screen;
}

.command-header {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: clamp(18px, 2vw, 40px);
  padding: 0 clamp(18px, 2.4vw, 44px);
  border-bottom: 1px solid rgba(67, 226, 255, 0.72);
  background:
    linear-gradient(90deg, rgba(12, 89, 151, 0.74), rgba(5, 24, 52, 0.9) 36%, rgba(8, 35, 70, 0.76)),
    repeating-linear-gradient(90deg, rgba(103, 232, 249, 0.08) 0 1px, transparent 1px 54px);
  box-shadow: 0 10px 36px rgba(0, 178, 255, 0.12), inset 0 -1px 0 rgba(255, 255, 255, 0.08);
  overflow: hidden;
}

.command-header::before,
.command-header::after {
  content: "";
  position: absolute;
  pointer-events: none;
}

.command-header::before {
  inset: 8px clamp(10px, 1.4vw, 24px);
  border: 1px solid rgba(103, 232, 249, 0.18);
  clip-path: polygon(0 0, 25% 0, 25% 1px, 75% 1px, 75% 0, 100% 0, 100% 100%, 72% 100%, 72% calc(100% - 1px), 28% calc(100% - 1px), 28% 100%, 0 100%);
}

.command-header::after {
  left: clamp(18px, 2.4vw, 44px);
  right: clamp(18px, 2.4vw, 44px);
  bottom: 0;
  height: 3px;
  background: linear-gradient(90deg, transparent, #29f3ff 18%, #2ff0a0 50%, #29f3ff 82%, transparent);
  opacity: 0.86;
  box-shadow: 0 0 14px rgba(41, 243, 255, 0.56);
}

.header-scanline {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 26%;
  background: linear-gradient(90deg, transparent, rgba(103, 232, 249, 0.18), transparent);
  transform: translateX(-120%);
  animation: header-scan 6.5s linear infinite;
  pointer-events: none;
}

.header-title-shell {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  min-width: 0;
  min-height: 100%;
}

.command-title {
  margin: 0;
  color: #ffffff;
  font-size: clamp(25px, 2.6em, 42px);
  font-weight: 900;
  line-height: 1;
  letter-spacing: 0.08em;
  text-shadow: 0 0 10px rgba(21, 183, 255, 0.82), 0 0 24px rgba(47, 240, 160, 0.2);
  white-space: nowrap;
}

.header-meta {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  gap: clamp(12px, 1.3em, 24px);
  color: #ebf5ff;
  font-family: "Courier New", monospace;
  font-size: clamp(15px, 1.5em, 24px);
  font-weight: 700;
  white-space: nowrap;
}

.progress-console {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: clamp(8px, 0.8vw, 14px);
  min-height: clamp(38px, 3.8vw, 54px);
  padding: 0 12px;
  border: 1px solid rgba(103, 232, 249, 0.26);
  background: rgba(3, 18, 38, 0.48);
  box-shadow: inset 0 0 18px rgba(0, 217, 255, 0.08);
}

.drill-name-tag {
  display: inline-grid;
  place-items: center;
  max-width: clamp(80px, 12vw, 200px);
  color: #f5fbff;
  font-family: "Microsoft YaHei", sans-serif;
  font-size: clamp(15px, 1.25vw, 19px);
  font-weight: 800;
  line-height: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  text-shadow: 0 0 10px rgba(41, 243, 255, 0.55), 0 0 18px rgba(47, 240, 160, 0.24);
}

.console-sep {
  width: 1px;
  height: clamp(16px, 1.6vw, 24px);
  background: linear-gradient(180deg, transparent, rgba(103, 232, 249, 0.5), transparent);
  flex-shrink: 0;
}

.system-label,
.system-time {
  display: inline-grid;
  place-items: center;
  color: #f5fbff;
  line-height: 1;
  text-shadow: 0 0 10px rgba(41, 243, 255, 0.55), 0 0 18px rgba(47, 240, 160, 0.24);
}

.system-label {
  font-family: "Microsoft YaHei", sans-serif;
  font-size: clamp(15px, 1.25vw, 19px);
  font-weight: 800;
}

.system-time {
  font-size: clamp(17px, 1.5vw, 23px);
  font-weight: 900;
  letter-spacing: 0.04em;
  transform: translateY(0.06em);
}

.btn-fullscreen {
  width: clamp(34px, 3vw, 52px);
  height: clamp(34px, 3vw, 52px);
  border: 1px solid rgba(0, 217, 255, 0.65);
  border-radius: 2px;
  background: rgba(0, 47, 82, 0.54);
  color: #03dcff;
  box-shadow: inset 0 0 16px rgba(0, 191, 255, 0.16), 0 0 16px rgba(0, 191, 255, 0.12);
}

.control-strip {
  position: absolute;
  right: clamp(18px, 2.4vw, 44px);
  bottom: -32px;
  display: flex;
  gap: 8px;
}

.control-btn {
  height: 24px;
  padding: 0 12px;
  border-radius: 2px;
  border: 1px solid rgba(0, 217, 255, 0.45);
  background: rgba(4, 23, 49, 0.84);
  color: #bdefff;
  cursor: pointer;
}

.control-btn.good { color: #25f3a2; border-color: rgba(37, 243, 162, 0.45); }
.control-btn.warn { color: #ffd166; border-color: rgba(255, 209, 102, 0.45); }
.control-btn.danger { color: #ff4d7d; border-color: rgba(255, 77, 125, 0.55); }

.command-main {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-rows: clamp(108px, 14vh, 150px) minmax(260px, 1fr) clamp(150px, 20vh, 190px);
  gap: clamp(8px, 1.1vh, 16px);
  padding: clamp(10px, 1.2vh, 18px) clamp(18px, 2vw, 36px) clamp(8px, 1vh, 16px);
  overflow: hidden;
}

.main-radar-sweep {
  position: absolute;
  left: 50%;
  top: 48%;
  width: min(68vw, 820px);
  aspect-ratio: 1;
  border-radius: 50%;
  transform: translate(-50%, -50%);
  background:
    conic-gradient(from 0deg, transparent 0 78%, rgba(47, 240, 160, 0.1), rgba(41, 243, 255, 0.18), transparent 88% 100%),
    radial-gradient(circle, transparent 0 58%, rgba(41, 243, 255, 0.1) 58.4% 58.8%, transparent 59.2% 100%);
  opacity: 0.46;
  animation: radar-sweep 16s linear infinite;
  pointer-events: none;
  will-change: transform;
}

.phase-card-strip {
  position: relative;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: clamp(6px, 0.75vw, 12px);
  padding: clamp(5px, 0.65vw, 10px);
  border: 1px solid rgba(48, 188, 235, 0.28);
  border-radius: clamp(16px, 1.5vw, 26px);
  background: linear-gradient(180deg, rgba(5, 25, 54, 0.72), rgba(2, 12, 30, 0.5));
  box-shadow: inset 0 0 24px rgba(0, 205, 255, 0.06), 0 0 24px rgba(0, 166, 255, 0.06);
  overflow: hidden;
}

.phase-card-strip::before {
  content: "";
  position: absolute;
  left: 8%;
  right: 8%;
  top: 50%;
  height: 2px;
  background: linear-gradient(90deg, rgba(20, 226, 255, 0.1), rgba(20, 226, 255, 0.55), rgba(29, 255, 154, 0.24));
  box-shadow: 0 0 14px rgba(20, 226, 255, 0.22);
  transform: translateY(-50%);
  pointer-events: none;
}

.phase-card {
  position: relative;
  z-index: 1;
  min-width: 0;
  padding: clamp(9px, 0.95vw, 16px) clamp(10px, 1.15vw, 20px);
  border: 1px solid rgba(72, 124, 177, 0.28);
  border-radius: clamp(10px, 1vw, 16px);
  background: linear-gradient(180deg, rgba(6, 30, 64, 0.82), rgba(4, 14, 31, 0.56));
  box-shadow: inset 0 -3px 0 rgba(65, 120, 170, 0.22);
  overflow: hidden;
}

.phase-card-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(103, 232, 249, 0.05) 1px, transparent 1px),
    linear-gradient(90deg, rgba(103, 232, 249, 0.04) 1px, transparent 1px);
  background-size: 24px 24px;
  opacity: 0.5;
  pointer-events: none;
}

.phase-card > *:not(.phase-card-grid) {
  position: relative;
  z-index: 1;
}

.phase-card:first-child {
  border-left: 1px solid rgba(72, 124, 177, 0.28);
}

.phase-card::after {
  content: "";
  position: absolute;
  top: 12px;
  right: -6px;
  bottom: 12px;
  width: 1px;
  background: linear-gradient(180deg, transparent, rgba(48, 188, 235, 0.34), transparent);
  z-index: 2;
}

.phase-card:last-child::after { display: none; }
.phase-card.is-done {
  border-color: rgba(47, 240, 160, 0.42);
  background: linear-gradient(180deg, rgba(12, 74, 58, 0.86), rgba(5, 30, 32, 0.62));
  box-shadow: inset 0 -4px 0 #2ff0a0, inset 0 0 24px rgba(47, 240, 160, 0.14);
}
.phase-card.is-running {
  z-index: 3;
  border-color: rgba(255, 176, 64, 0.62);
  background: linear-gradient(180deg, rgba(99, 58, 10, 0.92), rgba(48, 31, 17, 0.72));
  box-shadow: inset 0 -4px 0 #ffb13d, inset 0 0 24px rgba(255, 177, 61, 0.18), 0 0 20px rgba(255, 154, 47, 0.24);
}
.phase-card.is-pending {
  opacity: 0.88;
  border-color: rgba(112, 145, 176, 0.28);
  background: linear-gradient(180deg, rgba(30, 54, 82, 0.7), rgba(12, 24, 42, 0.58));
  box-shadow: inset 0 -4px 0 rgba(112, 145, 176, 0.42), inset 0 0 18px rgba(112, 145, 176, 0.08);
}

.phase-card.is-running .phase-accent {
  animation: accent-flow 2.2s ease-in-out infinite;
}
.phase-card.active {
  transform: translateY(-2px);
  border-color: #ffb13d;
}

.phase-accent {
  position: absolute;
  left: clamp(12px, 1vw, 18px);
  right: clamp(12px, 1vw, 18px);
  bottom: 0;
  height: 3px;
  background: linear-gradient(90deg, transparent, #7d9fbd, transparent);
  opacity: 0.42;
}

.is-done .phase-accent { background: linear-gradient(90deg, transparent, #2ff0a0, transparent); opacity: 0.72; }
.is-running .phase-accent { background: linear-gradient(90deg, transparent, #ffb13d, transparent); opacity: 0.86; }
.is-pending .phase-accent { background: linear-gradient(90deg, transparent, #7d9fbd, transparent); opacity: 0.34; }

.phase-head {
  display: flex;
  justify-content: space-between;
  gap: clamp(6px, 0.7vw, 10px);
  align-items: center;
  min-width: 0;
}

.phase-head h2 {
  min-width: 0;
  margin: 0;
  color: #f5fbff;
  font-size: clamp(15px, 1.45vw, 22px);
  font-weight: 800;
  line-height: 1.15;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.phase-status {
  flex: 0 0 auto;
  padding: 4px clamp(7px, 0.8vw, 11px);
  border: 1px solid currentColor;
  color: #2ff0a0;
  background: rgba(47, 240, 160, 0.12);
  font-size: clamp(12px, 1.05vw, 15px);
  font-weight: 700;
  line-height: 1;
}

.is-done .phase-status { color: #2ff0a0; background: rgba(47, 240, 160, 0.13); box-shadow: 0 0 12px rgba(47, 240, 160, 0.18); }
.is-running .phase-status { color: #ffb13d; background: rgba(255, 177, 61, 0.14); box-shadow: 0 0 14px rgba(255, 154, 47, 0.24); }
.is-pending .phase-status { color: #8aa6bf; background: rgba(112, 145, 176, 0.12); }

.phase-segments {
  display: grid;
  grid-template-columns: repeat(20, 1fr);
  gap: 2px;
  margin: clamp(8px, 0.9vh, 14px) 0 clamp(6px, 0.7vh, 11px);
}

.phase-segments span {
  height: clamp(5px, 0.75vh, 9px);
  background: rgba(83, 120, 158, 0.34);
  box-shadow: inset 0 0 6px rgba(8, 30, 62, 0.55);
}

.phase-segments span.filled {
  background: linear-gradient(180deg, #8aa6bf, #506d88);
  box-shadow: 0 0 8px rgba(138, 166, 191, 0.24);
}

.is-done .phase-segments span.filled {
  background: linear-gradient(180deg, #45ffc0, #0bbd78);
  box-shadow: 0 0 9px rgba(47, 240, 160, 0.36);
}

.is-running .phase-segments span.filled {
  background: linear-gradient(180deg, #ffd46a, #ff9a2f);
  box-shadow: 0 0 10px rgba(255, 177, 61, 0.42);
}

.phase-stats {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: clamp(6px, 0.75vw, 12px);
  color: #d8efff;
  font-family: "Courier New", monospace;
  font-size: clamp(13px, 1.25vw, 20px);
  font-weight: 700;
  text-shadow: 0 0 10px rgba(28, 222, 255, 0.34);
}

.phase-stats span {
  min-width: 0;
  padding: 2px clamp(5px, 0.55vw, 8px);
  border-radius: 8px;
  background: linear-gradient(180deg, rgba(8, 214, 255, 0.1), rgba(5, 45, 90, 0.14));
  box-shadow: inset 0 0 12px rgba(0, 206, 255, 0.1), 0 0 14px rgba(0, 216, 255, 0.06);
  white-space: nowrap;
}

.phase-stats b {
  color: #ffffff;
  font-size: 1.24em;
  text-shadow: 0 0 8px rgba(255, 255, 255, 0.64), 0 0 18px rgba(16, 224, 255, 0.52);
}

.is-done .phase-stats b { text-shadow: 0 0 8px rgba(255, 255, 255, 0.6), 0 0 18px rgba(47, 240, 160, 0.48); }
.is-running .phase-stats b { text-shadow: 0 0 8px rgba(255, 255, 255, 0.6), 0 0 18px rgba(255, 177, 61, 0.52); }
.is-pending .phase-stats b { color: #d8e7f3; text-shadow: 0 0 12px rgba(138, 166, 191, 0.32); }

.phase-stats em {
  margin-left: 8px;
  color: #ffffff;
  font-style: normal;
  font-family: "Microsoft YaHei", sans-serif;
  font-size: 0.88em;
  font-weight: 700;
  text-shadow: 0 0 8px rgba(255, 255, 255, 0.64), 0 0 18px rgba(16, 224, 255, 0.52);
}

.flow-board {
  --flow-row-gap: clamp(26px, 4vh, 52px);
  --arrow-gap: clamp(40px, 4vw, 68px);
  --arrow-margin: 8px;
  --tag-half-h: calc(clamp(14px, 1.4vh, 22px) + clamp(15px, 1.3vw, 24px) * 0.6 + 2px);
  position: relative;
  display: grid;
  grid-template-rows: 1fr 1fr;
  align-content: center;
  row-gap: var(--flow-row-gap);
  padding: clamp(18px, 2.4vh, 42px) clamp(32px, 4.2vw, 84px);
  min-height: 0;
  border: 1px solid rgba(255, 176, 64, 0.42);
  border-radius: clamp(18px, 1.6vw, 28px);
  background:
    linear-gradient(180deg, rgba(99, 58, 10, 0.18), rgba(48, 31, 17, 0.16)),
    radial-gradient(circle at 50% 50%, rgba(255, 177, 61, 0.08), transparent 48%);
  box-shadow: inset 0 0 28px rgba(255, 177, 61, 0.08), 0 0 24px rgba(255, 154, 47, 0.12);
  overflow: hidden;
}

.flow-board-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(255, 177, 61, 0.045) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 177, 61, 0.035) 1px, transparent 1px);
  background-size: 42px 42px;
  mask-image: radial-gradient(circle at center, black 0 62%, transparent 90%);
  pointer-events: none;
}

.flow-row {
  display: grid;
  grid-auto-flow: column;
  grid-auto-columns: 1fr;
  align-items: stretch;
  column-gap: var(--arrow-gap);
}

.flow-node-wrap {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 0;
}

.flow-node {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 0;
}

.node-tag {
  position: relative;
  z-index: 2;
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: center;
  padding: clamp(14px, 1.4vh, 22px) clamp(14px, 1.2vw, 24px);
  border-radius: clamp(10px, 1vw, 16px);
  border: 2px solid rgba(0, 210, 255, 0.64);
  color: #12e4ff;
  font-size: clamp(15px, 1.5em, 24px);
  font-weight: 700;
  background: rgba(4, 31, 55, 0.76);
  box-shadow: 0 0 28px rgba(0, 209, 255, 0.18), inset 0 0 18px rgba(0, 209, 255, 0.12);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-shadow: 0 0 10px rgba(0, 211, 255, 0.24);
}

.flow-node.is-pending {
  color: #4e779d;
}

.flow-node.is-pending .node-tag {
  border-color: rgba(78, 119, 157, 0.48);
  color: #5b83a8;
  box-shadow: none;
}

.flow-node.is-done .node-tag {
  border-color: rgba(47, 240, 160, 0.72);
  color: #a8ffd9;
  background: linear-gradient(180deg, rgba(12, 74, 58, 0.88), rgba(5, 30, 32, 0.68));
  box-shadow: 0 0 18px rgba(47, 240, 160, 0.22), inset 0 0 20px rgba(47, 240, 160, 0.12), inset 0 -4px 0 #2ff0a0;
  text-shadow: 0 0 10px rgba(47, 240, 160, 0.5);
}

.flow-node-wrap:has(.flow-node.is-done) .flow-arrow {
  background: linear-gradient(90deg, #2ff0a0, #1bc98a);
  box-shadow: 0 0 10px rgba(47, 240, 160, 0.5);
}

.flow-node-wrap:has(.flow-node.is-done) .flow-arrow.right::after { border-left-color: #1bc98a; }
.flow-node-wrap:has(.flow-node.is-done) .flow-arrow.left::after { border-right-color: #1bc98a; }
.flow-node-wrap:has(.flow-node.is-done) .flow-arrow.turn::after { border-top-color: #1bc98a; }

.flow-node.is-running .node-tag {
  border-color: rgba(255, 177, 61, 0.92);
  color: #ffe1a3;
  background: linear-gradient(180deg, rgba(99, 58, 10, 0.92), rgba(48, 31, 17, 0.72));
  box-shadow: 0 0 20px rgba(255, 154, 47, 0.26), inset 0 0 24px rgba(255, 177, 61, 0.16), inset 0 -4px 0 #ffb13d;
  text-shadow: 0 0 10px rgba(255, 177, 61, 0.55);
  animation: node-pulse 2s ease-in-out infinite;
}

.flow-node-wrap:has(.flow-node.is-running) .flow-arrow {
  background: linear-gradient(90deg, #ffd46a, #ff9a2f);
  box-shadow: 0 0 12px rgba(255, 177, 61, 0.62);
}

.flow-node-wrap:has(.flow-node.is-running) .flow-arrow.right::after { border-left-color: #ff9a2f; }
.flow-node-wrap:has(.flow-node.is-running) .flow-arrow.left::after { border-right-color: #ff9a2f; }
.flow-node-wrap:has(.flow-node.is-running) .flow-arrow.turn::after { border-top-color: #ff9a2f; }

@keyframes node-pulse {
  0%, 100% { box-shadow: 0 0 20px rgba(255, 154, 47, 0.26), inset 0 0 24px rgba(255, 177, 61, 0.16), inset 0 -4px 0 #ffb13d; }
  50% { box-shadow: 0 0 30px rgba(255, 154, 47, 0.42), inset 0 0 28px rgba(255, 177, 61, 0.24), inset 0 -4px 0 #ffb13d; }
}

.flow-arrow {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  height: 3px;
  background: linear-gradient(90deg, #24e9ff, #1dff9a);
  box-shadow: 0 0 10px rgba(0, 245, 194, 0.5);
  z-index: 1;
}

.flow-arrow::after {
  content: "";
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  border-top: 7px solid transparent;
  border-bottom: 7px solid transparent;
}

.flow-arrow.right {
  left: calc(100% + var(--arrow-margin));
  width: calc(var(--arrow-gap) - var(--arrow-margin) * 2);
}

.flow-arrow.right::after {
  right: -1px;
  border-left: 12px solid #1dff9a;
}

.flow-arrow.left {
  right: calc(100% + var(--arrow-margin));
  width: calc(var(--arrow-gap) - var(--arrow-margin) * 2);
  background: linear-gradient(90deg, #1dff9a, #24e9ff);
}

.flow-arrow.left::after {
  left: -1px;
  border-right: 12px solid #1dff9a;
}

.flow-arrow.turn {
  position: absolute;
  left: 50%;
  top: calc(50% + var(--tag-half-h) + 8px);
  width: 5px;
  height: calc(100% - var(--tag-half-h) * 2 + var(--flow-row-gap) - 24px);
  background: linear-gradient(180deg, #24e9ff, #1dff9a);
  box-shadow: 0 0 14px rgba(0, 245, 194, 0.65), 0 0 4px rgba(36, 233, 255, 0.9);
  border-radius: 3px;
  transform: translateX(-50%);
  z-index: 1;
  animation: turn-flow 2.4s ease-in-out infinite;
}

@keyframes turn-flow {
  0%, 100% { box-shadow: 0 0 14px rgba(0, 245, 194, 0.65), 0 0 4px rgba(36, 233, 255, 0.9); }
  50% { box-shadow: 0 0 22px rgba(0, 245, 194, 0.9), 0 0 8px rgba(36, 233, 255, 1); }
}

.flow-arrow.turn::after {
  left: 50%;
  top: auto;
  bottom: -2px;
  transform: translateX(-50%);
  border-left: 10px solid transparent;
  border-right: 10px solid transparent;
  border-top: 14px solid #1dff9a;
  border-bottom: none;
  filter: drop-shadow(0 0 6px rgba(0, 245, 194, 0.8));
}

.execution-section {
  position: relative;
  display: grid;
  grid-template-rows: clamp(34px, 4vh, 44px) minmax(0, 1fr);
  min-height: 0;
  border: 1px solid rgba(57, 220, 255, 0.18);
  background: linear-gradient(180deg, rgba(3, 19, 39, 0.18), rgba(3, 12, 28, 0.08));
  overflow: hidden;
}

.execution-section::before {
  content: "";
  position: absolute;
  inset: 0;
  background: linear-gradient(115deg, transparent 0 42%, rgba(103, 232, 249, 0.08) 49%, transparent 56% 100%);
  transform: translateX(-80%);
  animation: panel-sweep 8s ease-in-out infinite;
  pointer-events: none;
}

.execution-title {
  min-height: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 clamp(18px, 1.8vw, 34px);
  border-left: 4px solid #03dfff;
  border-right: 4px solid #03dfff;
  background: linear-gradient(90deg, rgba(9, 91, 167, 0.9), rgba(7, 50, 96, 0.72), rgba(4, 22, 45, 0.18));
  box-shadow: 0 0 24px rgba(0, 193, 255, 0.16);
}

.execution-title h2 {
  margin: 0;
  color: #ffffff;
  font-size: clamp(17px, 1.6em, 24px);
  font-weight: 800;
}

.execution-signal {
  display: flex;
  align-items: center;
  gap: 10px;
}

.execution-title span {
  color: #627f9f;
  font-size: clamp(13px, 1.1em, 18px);
}

.signal-bars {
  display: inline-grid;
  grid-template-columns: repeat(3, 4px);
  align-items: end;
  gap: 3px;
  height: 18px;
}

.signal-bars i {
  display: block;
  width: 4px;
  height: 7px;
  background: #2ff0a0;
  box-shadow: 0 0 8px rgba(47, 240, 160, 0.42);
  animation: signal-rise 1.4s ease-in-out infinite;
}

.signal-bars i:nth-child(2) { height: 12px; animation-delay: 0.16s; }
.signal-bars i:nth-child(3) { height: 17px; animation-delay: 0.32s; }

.execution-title span.live {
  color: #21f69e;
  text-shadow: 0 0 10px rgba(33, 246, 158, 0.52);
}

.execution-signal > span:last-child::before {
  content: "";
  display: inline-block;
  width: 9px;
  height: 9px;
  margin-right: 8px;
  border-radius: 50%;
  background: currentColor;
  box-shadow: 0 0 10px currentColor;
}

.execution-carousel {
  display: flex;
  align-items: stretch;
  justify-content: center;
  gap: 0;
  overflow: hidden;
  min-height: 0;
  padding: clamp(6px, 0.8vh, 12px) clamp(12px, 1.4vw, 24px) 0;
}

.exec-col {
  display: flex;
  flex-direction: column;
  flex: 0 1 auto;
  max-width: 50%;
  min-width: 0;
  min-height: 0;
}

.exec-col-label {
  display: flex;
  align-items: center;
  gap: 6px;
  padding-bottom: 6px;
  color: #8aa2bf;
  font-size: clamp(12px, 1em, 15px);
  font-weight: 700;
  white-space: nowrap;
}

.exec-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.exec-dot.running {
  background: #ffb13d;
  box-shadow: 0 0 8px rgba(255, 177, 61, 0.7);
}

.exec-dot.pending {
  background: #4e779d;
  box-shadow: 0 0 6px rgba(78, 119, 157, 0.5);
}

.exec-col-cards {
  display: grid;
  grid-auto-flow: column;
  grid-auto-columns: clamp(160px, 18vw, 240px);
  gap: clamp(6px, 0.8vw, 12px);
  overflow-x: auto;
  min-height: 0;
  padding-bottom: 6px;
  scrollbar-width: thin;
  scrollbar-color: rgba(0, 219, 255, 0.35) transparent;
}

.exec-divider {
  width: 1px;
  margin: 0 clamp(8px, 1vw, 16px);
  background: linear-gradient(180deg, transparent, rgba(103, 232, 249, 0.28), transparent);
  flex-shrink: 0;
}

.exec-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  min-height: 60px;
  color: #4a6788;
  font-size: 14px;
  font-weight: 600;
}

.execution-card {
  position: relative;
  min-width: 0;
  height: 100%;
  box-sizing: border-box;
  padding: clamp(10px, 1vw, 16px);
  border: 1px solid rgba(0, 190, 255, 0.28);
  border-radius: 3px;
  background:
    linear-gradient(180deg, rgba(2, 12, 29, 0.9), rgba(4, 24, 39, 0.78)),
    repeating-linear-gradient(90deg, rgba(103, 232, 249, 0.035) 0 1px, transparent 1px 30px);
  box-shadow: inset 0 0 18px rgba(0, 185, 255, 0.08);
  overflow: hidden;
}

.card-scan {
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, transparent, rgba(47, 240, 160, 0.1), transparent);
  transform: translateX(-120%);
  animation: card-scan 4.8s ease-in-out infinite;
  pointer-events: none;
}

.execution-card > *:not(.card-scan) {
  position: relative;
  z-index: 1;
}

.task-card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.task-card-head strong {
  color: #e8f3ff;
  font-size: clamp(15px, 1.3em, 22px);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.task-card-head span {
  padding: 3px 10px;
  border: 1px solid rgba(0, 207, 255, 0.44);
  border-radius: 2px;
  color: #0bdfff;
  background: rgba(0, 207, 255, 0.08);
  white-space: nowrap;
}

.task-progress {
  height: 11px;
  margin: 14px 0 11px;
  border-radius: 12px;
  background: rgba(50, 102, 132, 0.5);
  overflow: hidden;
}

.task-progress div {
  height: 100%;
  background: linear-gradient(90deg, #146f90, #12d7f5);
  box-shadow: 0 0 10px rgba(18, 215, 245, 0.42);
}

.execution-card p {
  margin: 14px 0 0;
  color: #8aa2bf;
  font-weight: 700;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.execution-card.is-running {
  border-color: rgba(255, 177, 61, 0.5);
  box-shadow: inset 0 0 18px rgba(255, 154, 47, 0.1), 0 0 14px rgba(255, 154, 47, 0.08);
}

.execution-card.is-running .task-card-head span {
  border-color: rgba(255, 177, 61, 0.6);
  color: #ffd9a0;
  background: rgba(255, 177, 61, 0.1);
}

.execution-card.is-running .task-progress div {
  background: linear-gradient(90deg, #b8740e, #ffb13d);
  box-shadow: 0 0 10px rgba(255, 177, 61, 0.42);
}

.execution-card.is-pending {
  border-color: rgba(78, 119, 157, 0.4);
  opacity: 0.82;
}

.execution-card.is-pending .task-card-head span {
  border-color: rgba(78, 119, 157, 0.5);
  color: #6b8db0;
  background: rgba(78, 119, 157, 0.08);
}

.execution-card.is-pending .task-progress div {
  background: linear-gradient(90deg, #2a4a66, #4e779d);
  box-shadow: none;
}

@keyframes header-scan {
  0% { transform: translateX(-120%); opacity: 0; }
  12% { opacity: 1; }
  58% { opacity: 0.55; }
  100% { transform: translateX(420%); opacity: 0; }
}

@keyframes radar-sweep {
  from { transform: translate(-50%, -50%) rotate(0deg); }
  to { transform: translate(-50%, -50%) rotate(360deg); }
}

@keyframes accent-flow {
  0%, 100% { opacity: 0.42; transform: scaleX(0.72); }
  50% { opacity: 0.9; transform: scaleX(1); }
}

@keyframes panel-sweep {
  0%, 32% { transform: translateX(-90%); opacity: 0; }
  42% { opacity: 0.8; }
  72%, 100% { transform: translateX(90%); opacity: 0; }
}

@keyframes signal-rise {
  0%, 100% { opacity: 0.35; transform: scaleY(0.64); }
  50% { opacity: 1; transform: scaleY(1); }
}

@keyframes card-scan {
  0%, 38% { transform: translateX(-120%); opacity: 0; }
  48% { opacity: 0.8; }
  72%, 100% { transform: translateX(120%); opacity: 0; }
}

@media (prefers-reduced-motion: reduce) {
  .header-scanline,
  .main-radar-sweep,
  .phase-card.is-running .phase-accent,
  .flow-node.is-running .node-tag,
  .flow-arrow.turn,
  .execution-section::before,
  .signal-bars i,
  .card-scan {
    animation: none !important;
  }
}

@media (max-width: 1180px) {
  .command-header { grid-template-columns: minmax(0, 1fr) auto; gap: 10px; padding-block: 6px; }
  .header-meta { justify-content: flex-end; min-width: 0; }
  .progress-console { flex: 0 0 auto; min-height: 32px; padding-inline: 8px; gap: 6px; }
  .drill-name-tag { max-width: 90px; font-size: 13px; }
  .phase-card-strip { gap: 6px; padding: 5px; }
  .phase-card {
    padding: 8px 9px 7px;
    border-radius: 10px;
  }
  .phase-card::after { display: none; }
  .phase-head { gap: 5px; }
  .phase-head h2 {
    font-size: clamp(14px, 1.55vw, 16px);
    letter-spacing: -0.02em;
  }
  .phase-status {
    padding: 3px 6px;
    font-size: 11px;
  }
  .phase-segments {
    gap: 2px;
    margin: 6px 0 5px;
  }
  .phase-segments span { height: 5px; }
  .phase-stats {
    gap: 4px;
    font-size: clamp(13px, 1.6vw, 16px);
  }
  .phase-stats span {
    flex: 1 1 0;
    padding: 1px 4px;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .phase-stats em {
    margin-left: 4px;
    font-size: 0.78em;
  }
  .flow-board { padding-inline: 24px; }
}

@media (max-width: 1060px) {
  .phase-card-strip {
    gap: 5px;
    padding-inline: 4px;
  }
  .phase-card { padding-inline: 8px; }
  .phase-head h2 { font-size: 14px; }
  .phase-status {
    padding-inline: 5px;
    font-size: 10px;
  }
  .phase-stats {
    font-size: 14px;
    line-height: 1;
  }
  .phase-stats b { font-size: 1.12em; }
  .phase-stats em {
    margin-left: 3px;
    font-size: 0.72em;
  }
  .exec-col-cards {
    grid-auto-columns: clamp(130px, 22vw, 180px);
    gap: 5px;
  }
  .exec-divider { margin-inline: 5px; }
  .execution-card { padding: 8px; }
  .task-card-head strong { font-size: 14px; }
  .task-card-head span { padding: 2px 6px; font-size: 11px; }
  .task-progress { height: 8px; margin: 8px 0 6px; }
  .execution-card p { margin-top: 8px; font-size: 12px; }
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

/* 大屏2 进入角色路由时也要独占视口；旧 /screen/:id 不经过该布局，不受影响。 */
.app-layout:has(.screen-root) .app-header,
.app-layout:has(.screen-root) .app-sidebar {
  display: none !important;
}

.app-layout:has(.screen-root) .app-main {
  margin-left: 0 !important;
  padding-top: 0 !important;
  min-height: 100vh !important;
}

.app-layout:has(.screen-root) .app-content {
  padding: 0 !important;
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
