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
          <span class="title-rail is-left" aria-hidden="true" />
          <h1 class="command-title" data-text="应急指挥中心">应急指挥中心</h1>
          <span class="title-rail is-right" aria-hidden="true" />
        </div>
        <div class="header-meta">
          <button class="btn-fullscreen" @click="toggleFullscreen" title="全屏模式">
            <el-icon><FullScreen /></el-icon>
          </button>
        </div>
        <div v-if="canControl" class="control-strip">
          <button v-if="instance?.status === 'pending'" class="control-btn good" @click="handleStart">开始</button>

        </div>
      </header>

      <main class="command-main">
        <div class="main-rect-sweep" />
        <section ref="phaseFlowRef" class="phase-flow-chamber" :class="'phase-state-' + selectedPhaseStatus" aria-label="阶段与环节流程">
          <svg class="phase-chamber-surface" aria-hidden="true">
            <defs>
              <linearGradient id="phaseSurface" x1="0" y1="0" x2="1" y2="1">
                <stop offset="0" stop-color="#0a304b" />
                <stop offset="0.6" stop-color="#04192e" />
                <stop offset="1" stop-color="#07283d" />
              </linearGradient>
              <linearGradient id="phaseOutline" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0" stop-color="var(--phase-link-color)" />
                <stop offset="0.32" stop-color="#41bdd6" />
                <stop offset="1" stop-color="#246880" />
              </linearGradient>
            </defs>
            <path :d="chamberPath" fill="url(#phaseSurface)" stroke="url(#phaseOutline)" stroke-width="1.5" />
          </svg>
          <section class="phase-card-strip" aria-label="选择预览阶段" @scroll.passive="updateChamberOutline">
            <template v-for="(phase, index) in phaseCards" :key="phase.name">
              <button
                type="button"
                class="phase-card"
                :class="['is-' + phase.status, { active: index === selectedPhaseIdx }]"
                :aria-pressed="index === selectedPhaseIdx"
                aria-controls="selected-phase-flow"
                :title="`查看阶段${index + 1} ${phase.name}`"
                @click="selectPhase(index)"
              >
                <span class="phase-card-grid" aria-hidden="true" />
                <span class="phase-accent" aria-hidden="true" />
                <span class="phase-head">
                  <span class="phase-name"><span class="phase-number">阶段{{ index + 1 }}</span> {{ phase.name }}</span>
                  <span class="phase-status">{{ phase.statusText }}</span>
                </span>
                <span class="phase-segments" aria-hidden="true">
                  <span v-for="seg in phase.segmentCount" :key="seg" :class="{ filled: seg <= phase.filledSegments }" />
                </span>
                <span class="phase-stats">
                  <span><b>{{ phase.completedPhaseSteps }}</b>/{{ phase.totalPhaseSteps }}<em>环节</em></span>
                  <span><b>{{ phase.completedSteps }}</b>/{{ phase.totalSteps }}<em>步骤</em></span>
                </span>
              </button>
              <span v-if="index < phaseCards.length - 1" class="phase-sequence-arrow" :class="'is-' + phase.status" aria-hidden="true">
                <i class="seq-rail" />
                <i class="seq-flow" />
                <i class="seq-comet" />
                <i class="seq-head" />
              </span>
            </template>
          </section>

          <section id="selected-phase-flow" class="flow-board" :class="{ 'all-done': selectedPhaseStatus === 'done', 'is-preview': selectedPhaseStatus !== 'running' }">
            <div class="flow-board-grid" />
            <header class="flow-board-heading" aria-live="polite">
              <div class="flow-board-label">
                <span class="label-pulse" aria-hidden="true" />
                <span>{{ selectedPhaseStatus === 'running' ? '当前环节' : '阶段预览' }}</span>
              </div>
              <div class="board-signal" :class="{ live: wsConnected }">
                <span class="signal-bars"><i /><i /><i /></span>
                <span>{{ wsConnected ? '实时联动' : '轮询同步' }}</span>
              </div>
            </header>
            <div v-if="!flowNodes.length" class="flow-empty">该阶段暂无环节</div>
            <div v-show="flowNodes.length" ref="flowViewportRef" class="flow-viewport">
              <div ref="flowTrackRef" class="flow-track" :style="trackTransform">
                <!-- 虚拟开始节点：首个环节聚焦时左侧仍有延伸，保持居中选人效果 -->
                <div class="flow-node-wrap is-virtual" :style="focusStyle(-1)">
                  <div class="flow-node is-virtual-start">
                    <span class="virtual-badge">
                      <i class="virtual-glyph virtual-glyph-start" />
                    </span>
                    <span class="virtual-name">开始</span>
                  </div>
                </div>
                <span class="flow-arrow is-virtual" :style="virtualArrowStyle('start')" aria-hidden="true">
                  <i class="arrow-port" />
                </span>
                <template v-for="(node, index) in flowNodes" :key="node.id">
                  <div class="flow-node-wrap" :style="focusStyle(index)">
                    <div class="flow-node" :class="'is-' + node.status">
                      <span v-if="node.status === 'running'" class="node-ripple" aria-hidden="true" />
                      <span class="node-tag">
                        <span class="node-label">{{ node.status === 'done' ? '✓ ' + node.name : node.name }}</span>
                      </span>
                      <span v-if="node.status === 'running'" class="node-motion" aria-hidden="true">
                        <i class="node-gear" />
                        <i class="node-live-dot" />
                      </span>
                      <ul v-if="node.steps.length" class="node-steps">
                        <li
                          v-for="step in node.steps.slice(0, NODE_STEP_LIMIT)"
                          :key="step.id"
                          class="node-step"
                          :class="'is-' + step.status"
                        >
                          <i class="step-ico" aria-hidden="true" />
                          <span class="step-name" :title="step.name">{{ step.name }}</span>
                          <i v-if="step.status === 'done'" class="step-check" aria-hidden="true">
                            <svg viewBox="0 0 12 12"><path d="M2.4 6.4 L5 9 L9.6 3.4" /></svg>
                          </i>
                        </li>
                        <li v-if="node.steps.length > NODE_STEP_LIMIT" class="node-step is-more">
                          <i class="step-ico" aria-hidden="true" />
                          <span class="step-name">另有 {{ node.steps.length - NODE_STEP_LIMIT }} 个步骤…</span>
                        </li>
                      </ul>
                    </div>
                  </div>
                  <span
                    class="flow-arrow"
                    :class="'is-' + node.status"
                    :style="arrowStyle(index)"
                    aria-hidden="true"
                  >
                    <i class="arrow-port" />
                  </span>
                </template>
                <!-- 虚拟结束节点：末尾环节聚焦时右侧仍有延伸 -->
                <div class="flow-node-wrap is-virtual" :style="focusStyle(flowNodes.length)">
                  <div class="flow-node is-virtual-end">
                    <span class="virtual-badge">
                      <i class="virtual-glyph virtual-glyph-end" />
                    </span>
                    <span class="virtual-name">结束</span>
                  </div>
                </div>
              </div>
            </div>

            <!-- 左下角：演练概览（名称 + 进度） -->
            <aside class="flow-brief" aria-label="演练概览">
              <div class="brief-title">
                <span class="brief-sigil" aria-hidden="true" />
                <span>演练概览</span>
              </div>
              <div class="brief-name" :title="instance?.name">{{ instance?.name || '未命名演练' }}</div>
              <div class="brief-progress">
                <div class="brief-ring" :style="{ '--p': liveProgressPct }">
                  <span class="brief-ring-val">{{ liveProgressPct }}<em>%</em></span>
                </div>
                <div class="brief-meta">
                  <span class="brief-status" :class="'is-' + (instance?.status || 'pending')">{{ statusLabel }}</span>
                  <span class="brief-count">步骤 <b>{{ completedStepCount }}</b> / {{ totalStepCount }}</span>
                  <span class="brief-clock">{{ displayTime }}</span>
                </div>
              </div>
              <div class="brief-bar">
                <div class="brief-bar-fill" :style="{ width: liveProgressPct + '%' }" />
              </div>
            </aside>

            <!-- 右下角：执行日志（最新在下，自动贴底滚动） -->
            <aside class="flow-log-panel" aria-label="执行日志">
              <header class="log-head">
                <span class="log-title"><span class="log-dot" aria-hidden="true" />执行日志</span>
                <span class="log-tail">最新 ↓</span>
              </header>
              <div ref="logContainerRef" class="log-body" aria-live="polite">
                <p v-if="!orderedLogs.length" class="log-empty">暂无执行日志</p>
                <div v-for="log in orderedLogs" :key="log.id" class="log-row" :class="'is-' + log.type">
                  <span class="log-time">{{ log.time }}</span>
                  <span class="log-icon" aria-hidden="true">{{ log.icon }}</span>
                  <span class="log-msg" :title="log.msg">{{ log.msg }}</span>
                </div>
              </div>
            </aside>
          </section>

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
import { ref, computed, onMounted, onUnmounted, nextTick, watch, type CSSProperties } from 'vue'
import { useRoute } from 'vue-router'
import { FullScreen } from '@element-plus/icons-vue'
import { drillApi } from '@/api/modules/drill'
import { useAuthStore } from '@/stores/auth'
import type { DrillInstance, StepInstance, DrillStatus } from '@/types/instance'
import { DRILL_STATUS_LABELS } from '@/types/instance'
import { getFlowFocusIndex, getPhaseChamberPath, getPhaseFlowNodes, getPhaseStripScrollLeft, useScreenPhaseSelection } from './screenPhaseFlow'

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

// 实时进度：基于叶子步骤计算全局完成率（与 ScreenView 统一口径）
const liveProgressPct = computed(() => {
  if (!steps.value.length) return instance.value?.progress_pct ?? 0
  const leafSteps = steps.value.filter(isLeafStep)
  const list = leafSteps.length > 0 ? leafSteps : steps.value
  const done = list.filter(s => ['completed', 'skipped', 'timeout', 'issue'].includes(s.status)).length
  const total = list.length
  return total === 0 ? 0 : Math.round((done / total) * 100)
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

// 实时日志（按容器高度自适应条数）
const maxVisibleLogs = ref(8)
const LOG_ROW_H = 26
const logContainerRef = ref<HTMLElement | null>(null)

function updateMaxVisibleLogs() {
  const el = logContainerRef.value
  if (el) {
    const h = el.clientHeight
    maxVisibleLogs.value = Math.max(3, Math.floor(h / LOG_ROW_H))
  }
}

const displayLogs = computed(() => logs.value.slice(0, maxVisibleLogs.value))

// logs[0] 为最新，展示时倒序排列，使最新日志位于列表底部
const orderedLogs = computed(() => displayLogs.value.slice().reverse())

// 演练概览：叶子步骤口径的完成统计
const leafStepsAll = computed(() => {
  const leaf = steps.value.filter(isLeafStep)
  return leaf.length > 0 ? leaf : steps.value
})
const totalStepCount = computed(() => leafStepsAll.value.length)
const completedStepCount = computed(
  () => leafStepsAll.value.filter(s => ['completed', 'skipped', 'timeout', 'issue'].includes(s.status)).length,
)

// ======== 阶段 Tab ========

// 阶段整体状态：done / running / pending（仅看叶子步骤，避免父步骤状态滞后）
function getPhaseStatus(phase: TreeNodePhase): string {
  const all = phase.phaseSteps.flatMap(ps => ps.stepNodes)
  const leafSteps = all.filter(isLeafStep)
  const stepsToCheck = leafSteps.length > 0 ? leafSteps : all
  if (stepsToCheck.some(s => s.status === 'running')) return 'running'
  if (stepsToCheck.every(s => s.status === 'completed' || s.status === 'skipped')) return 'done'
  return 'pending'
}

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

const phaseSummaries = computed(() => treeData.value.map(phase => ({ name: phase.name, status: getPhaseStatus(phase) })))
const { selectedPhaseIdx } = useScreenPhaseSelection(phaseSummaries, drillId)
const currentPhaseData = computed<TreeNodePhase | null>(() => treeData.value[selectedPhaseIdx.value] ?? null)
const selectedPhaseStatus = computed(() => phaseSummaries.value[selectedPhaseIdx.value]?.status ?? 'pending')

function selectPhase(index: number) {
  selectedPhaseIdx.value = index
}

const phaseCards = computed(() => {
  return treeData.value.map(phase => {
    const allSteps = phase.phaseSteps.flatMap(ps => ps.stepNodes)
    const leafSteps = allSteps.filter(isLeafStep)
    const status = getPhaseStatus(phase)
    const completedSteps = leafSteps.filter(s => s.status === 'completed' || s.status === 'skipped').length
    const totalSteps = leafSteps.length
    const completedPhaseSteps = phase.phaseSteps.filter(ps => getPhaseStepStatus(ps) === 'done').length
    const totalPhaseSteps = phase.phaseSteps.length || 1
    const segmentCount = 20
    const filledSegments = Math.round((completedSteps / (totalSteps || 1)) * segmentCount)
    return {
      name: phase.name,
      status,
      statusText: status === 'done' ? '已完成' : status === 'running' ? '进行中' : '待开始',
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

// 环节节点下最多展示的任务步骤数
const NODE_STEP_LIMIT = 3

function normalizeStepStatus(status: string): string {
  if (status === 'completed') return 'done'
  if (status === 'running') return 'running'
  if (status === 'skipped') return 'skipped'
  if (status === 'timeout' || status === 'issue') return 'issue'
  return 'pending'
}

const flowNodes = computed(() => getPhaseFlowNodes(currentPhaseData.value, getPhaseStepStatus, link => {
  const leafSteps = link.stepNodes.filter(isLeafStep)
  const list = leafSteps.length > 0 ? leafSteps : link.stepNodes
  return list.map(s => ({ id: String(s.id), name: s.name, status: normalizeStepStatus(s.status) }))
}))

// 环节轮播：单行横向排列，进行中节点聚焦居中（游戏选人式）
const flowViewportRef = ref<HTMLElement | null>(null)
const flowTrackRef = ref<HTMLElement | null>(null)
const focusShift = ref(0)
let flowViewportObserver: ResizeObserver | null = null

const focusedNodeIndex = computed(() => getFlowFocusIndex(flowNodes.value))

// 每个节点相对进行中节点的聚焦样式：中间放大、两侧渐小渐暗
function focusScale(index: number): number {
  const r = focusedNodeIndex.value
  const d = r === -1 ? 0 : index - r
  const abs = Math.abs(d)
  // 进行中节点放大 1.3，其余按距离递减
  return abs === 0 ? 1.3 : Math.max(0.55, 1 - (abs - 1) * 0.16)
}

function focusOpacity(index: number): number {
  const r = focusedNodeIndex.value
  const d = r === -1 ? 0 : index - r
  const abs = Math.abs(d)
  return abs === 0 ? 1 : Math.max(0.32, 1 - abs * 0.24)
}

function focusStyle(index: number): CSSProperties {
  const r = focusedNodeIndex.value
  const d = r === -1 ? 0 : index - r
  return {
    transform: `scale(${focusScale(index).toFixed(3)})`,
    opacity: focusOpacity(index).toFixed(3),
    zIndex: String(30 - Math.abs(d)),
  }
}

// 衔接箭头：通过负边距吃掉卡片缩放留出的空白，使任意相邻卡片的视觉间距恒等于 --arrow-gap
const flowWrapWidth = ref(180)

function arrowStyle(index: number): CSSProperties {
  const w = flowWrapWidth.value
  // 卡片缩放后，wrap 两侧留出的空白 = w * (1 - scale) / 2（scale > 1 时为负，即卡片外溢）
  const extendL = (w * (1 - focusScale(index))) / 2
  const extendR = (w * (1 - focusScale(index + 1))) / 2
  return {
    marginLeft: `${(-extendL).toFixed(1)}px`,
    marginRight: `${(-extendR).toFixed(1)}px`,
    opacity: Math.min(focusOpacity(index), focusOpacity(index + 1)).toFixed(3),
  }
}

const trackTransform = computed<CSSProperties>(() => ({
  transform: `translateX(${focusShift.value}px)`,
}))

// 虚拟起止节点：徽章直径占 wrap 宽度比例（与 CSS 中 .virtual-badge 的 width 对应）
const VIRTUAL_BADGE_RATIO = 0.4

// 虚拟节点与相邻卡片的衔接箭头：徽章远小于 wrap，需按徽章实际视觉边缘收拢负边距
function virtualArrowStyle(side: 'start' | 'end'): CSSProperties {
  const w = flowWrapWidth.value
  const virtualIndex = side === 'start' ? -1 : flowNodes.value.length
  const nodeIndex = side === 'start' ? 0 : flowNodes.value.length - 1
  const badgeInset = (w * (1 - VIRTUAL_BADGE_RATIO * focusScale(virtualIndex))) / 2
  const cardExtend = (w * (1 - focusScale(nodeIndex))) / 2
  return {
    marginLeft: `${(-(side === 'start' ? badgeInset : cardExtend)).toFixed(1)}px`,
    marginRight: `${(-(side === 'start' ? cardExtend : badgeInset)).toFixed(1)}px`,
    opacity: Math.min(focusOpacity(virtualIndex), focusOpacity(nodeIndex)).toFixed(3),
  }
}

// 平移轨道，使进行中节点对准视口中线
function recomputeFocusShift() {
  const viewport = flowViewportRef.value
  const track = flowTrackRef.value
  if (!viewport || !track) return
  const r = focusedNodeIndex.value
  const items = track.querySelectorAll<HTMLElement>('.flow-node-wrap')
  if (items.length && flowWrapWidth.value !== items[0].offsetWidth) {
    flowWrapWidth.value = items[0].offsetWidth
    nextTick(recomputeFocusShift)
    return
  }
  // items[0] 是虚拟开始节点，真实节点索引需 +1
  if (r < 0 || !items.length || r + 1 >= items.length) {
    focusShift.value = 0
    return
  }
  const target = items[r + 1]
  const targetCenter = target.offsetLeft + target.offsetWidth / 2
  focusShift.value = Math.round(viewport.clientWidth / 2 - targetCenter)
}

watch(flowNodes, () => nextTick(recomputeFocusShift))

// ======== 所选阶段与环节内容的连接 ========
const phaseFlowRef = ref<HTMLElement | null>(null)
const chamberPath = ref('')
let chamberResizeObserver: ResizeObserver | null = null

function updateChamberOutline() {
  const mainEl = phaseFlowRef.value
  if (!mainEl) return
  const card = mainEl.querySelector<HTMLElement>('.phase-card.active')
  const board = mainEl.querySelector<HTMLElement>('.flow-board')
  if (!board) return
  const m = mainEl.getBoundingClientRect()
  const cr = card?.getBoundingClientRect()
  chamberPath.value = getPhaseChamberPath(m.width, m.height, board.getBoundingClientRect().top - m.top,
    cr ? { left: cr.left - m.left, right: cr.right - m.left, top: cr.top - m.top } : undefined)
}

watch(phaseCards, () => {
  nextTick(updatePhaseLayout)
})

function updatePhaseLayout() {
  const strip = phaseFlowRef.value?.querySelector<HTMLElement>('.phase-card-strip')
  const card = strip?.querySelector<HTMLElement>('.phase-card.active')
  if (strip && card) {
    const gutter = parseFloat(getComputedStyle(strip).paddingLeft)
    strip.scrollLeft = getPhaseStripScrollLeft(strip.scrollLeft, strip.clientWidth, card.offsetLeft - gutter, card.offsetWidth + gutter * 2)
  }
  updateChamberOutline()
}

watch(selectedPhaseIdx, () => nextTick(updatePhaseLayout))

// 模板在加载结束后才创建，观察实际挂载的元素。
watch([phaseFlowRef, flowViewportRef], ([chamber, viewport]) => {
  chamberResizeObserver?.disconnect()
  flowViewportObserver?.disconnect()
  if (chamber) {
    chamberResizeObserver = new ResizeObserver(updatePhaseLayout)
    chamberResizeObserver.observe(chamber)
  }
  if (viewport) {
    flowViewportObserver = new ResizeObserver(recomputeFocusShift)
    flowViewportObserver.observe(viewport)
  }
  nextTick(() => { updatePhaseLayout(); recomputeFocusShift() })
}, { flush: 'post' })

// 过滤阶段节点（phase === phase_step）和环节节点（有子步骤的父步骤）
function isLeafStep(s: StepInstance): boolean {
  if (s.phase && s.phase_step && s.phase === s.phase_step) return false
  if (isParentStep(s)) return false
  return true
}

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

// 子阶段的状态（仅看叶子步骤，避免父步骤状态滞后导致环节无法标记完成）
function getPhaseStepStatus(ps: TreeNodePhaseStep): string {
  const leafSteps = ps.stepNodes.filter(isLeafStep)
  const stepsToCheck = leafSteps.length > 0 ? leafSteps : ps.stepNodes
  if (stepsToCheck.some(s => s.status === 'running')) return 'running'
  if (stepsToCheck.every(s => s.status === 'completed' || s.status === 'skipped')) return 'done'
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
  const socket = new WebSocket(url)
  ws = socket

  socket.onopen = () => {
    if (socket !== ws) return
    wsConnected.value = true
    wsReconnectCount = 0
  }

  socket.onclose = () => {
    if (socket !== ws) return
    wsConnected.value = false
    scheduleReconnect()
  }

  socket.onerror = () => {
    socket.close()
  }

  socket.onmessage = (ev) => {
    if (socket !== ws) return
    try {
      // 后端 WritePump 会把消息打包成 JSON 数组批量发送，需逐条处理
      const parsed = JSON.parse(ev.data)
      const messages = Array.isArray(parsed) ? parsed : [parsed]
      messages.forEach(msg => handleWSMessage(msg))
    } catch { /* ignored */ }
  }
}

function disconnectWS() {
  const socket = ws
  ws = null
  socket?.close()
  wsConnected.value = false
  if (wsReconnectTimer) clearTimeout(wsReconnectTimer)
  wsReconnectTimer = null
  wsReconnectCount = 0
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
    scheduleRefresh('drill', 'steps', 'logs')
    if (event === 'drill_started') addLog('info', '▶', '演练已开始')
    else if (event === 'drill_paused') addLog('warn', '⏸', '演练已暂停')
    else if (event === 'drill_resumed') addLog('info', '▶', '演练已恢复')
    else if (event === 'drill_completed') addLog('info', '✓', '演练已完成')
    else if (event === 'drill_terminated') addLog('error', '⏹', '演练已结束')
  }

  // 步骤事件：增量更新本地数据，不调 API
  if (event.startsWith('step_')) {
    patchLocalStep(event, payload)
    scheduleRefresh('steps', 'drill', 'logs')
    const phasePrefix = phaseName ? `【${phaseName}】` : ''
    if (event === 'step_started') {
      addLog('info', '●', `${phasePrefix}${stepName} 已开始`)
    } else if (['step_complete', 'step_completed', 'step_skipped', 'step_issue', 'step_timeout'].includes(event)) {
      const label = logLabel(event)
      const logType = event === 'step_issue' ? 'error' : 'info'
      addLog(logType, logIcon(event), `${phasePrefix}${stepName} ${label}`)
      if (event === 'step_complete' || event === 'step_completed') {
        showCompletionModal(stepName, phaseName)
      }
    }
    return
  }

  if (event === 'timeout_warning') return
}

// 增量更新本地步骤数据，全量刷新会随后校准父级和阶段状态
function patchLocalStep(event: string, payload: any) {
  const stepId = Number(payload.step_id || payload.stepId || payload.id || payload.step_instance_id)
  if (!stepId) return

  const newStatus = payload.new_status || payload.newStatus || mapStepEventToStatus(event)
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

function mapStepEventToStatus(event: string): string {
  const map: Record<string, string> = {
    step_started: 'running',
    step_complete: 'completed',
    step_completed: 'completed',
    step_skipped: 'skipped',
    step_issue: 'issue',
    step_timeout: 'timeout',
  }
  return map[event] || ''
}

function logLabel(event: string): string {
  const map: Record<string, string> = {
    step_started: '已开始',
    step_complete: '已完成',
    step_completed: '已完成',
    step_skipped: '已跳过',
    step_issue: '异常',
    step_timeout: '已超时',
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

// 最新日志在底部，故始终贴底滚动
function scrollLogs() {
  const el = logContainerRef.value
  if (el) el.scrollTop = el.scrollHeight
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
  const requestId = drillId.value
  try {
    const data = await drillApi.getDetail(drillId.value)
    if (requestId !== drillId.value) return
    instance.value = data
  } catch {
    // 静默失败
  }
}

async function fetchSteps() {
  const requestId = drillId.value
  try {
    const data = await drillApi.getSteps(drillId.value)
    if (requestId !== drillId.value) return
    steps.value = (data || []).sort((a: StepInstance, b: StepInstance) => a.seq - b.seq)
  } catch {
    // 静默失败
  }
}

async function fetchLogs() {
  const requestId = drillId.value
  try {
    const data = await drillApi.getLogs(drillId.value)
    if (requestId !== drillId.value) return
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
  const requestId = drillId.value
  loading.value = true
  error.value = ''
  try {
    await Promise.all([fetchDrillData(), fetchSteps(), fetchLogs()])
  } catch {
    error.value = '数据加载失败'
  } finally {
    if (requestId === drillId.value) loading.value = false
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

watch(drillId, () => {
  disconnectWS()
  if (refreshTimer) clearTimeout(refreshTimer)
  refreshTimer = null
  pendingRefresh.clear()
  steps.value = []
  instance.value = null
  logs.value = []
  loadAllData()
  connectWS()
})

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
    updatePhaseLayout()
    nextTick(recomputeFocusShift)
  })
  connectWS()
  timerInterval = setInterval(() => {
    updateTimer()
    redrawCanvas()
  }, 1000)
  pollingTimer = setInterval(() => {
    if (!wsConnected.value) scheduleRefresh('drill', 'steps', 'logs')
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

  disconnectWS()
  if (refreshTimer) clearTimeout(refreshTimer)
  if (timerInterval) clearInterval(timerInterval)
  if (pollingTimer) clearInterval(pollingTimer)
  window.removeEventListener('resize', onResize)
  chamberResizeObserver?.disconnect()
  flowViewportObserver?.disconnect()
  const canvasEl = flowCanvasRef.value
  if (canvasEl) canvasEl.removeEventListener('click', handleCanvasClick)
})

function onResize() {
  initCanvas()
  updateMaxVisibleLogs()
  updatePhaseLayout()
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
  background-image: linear-gradient(rgba(65, 167, 244, 0.035) 1px, transparent 1px);
  background-size: 100% 5px;
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
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
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
  grid-column: 2;
  justify-self: center;
  display: flex;
  align-items: center;
  gap: clamp(14px, 1.6vw, 26px);
  min-width: 0;
  min-height: 100%;
}

/* 标题两侧对称能量导轨：渐隐光线 + 内端菱形焊点 */
.title-rail {
  position: relative;
  flex-shrink: 0;
  width: clamp(36px, 5vw, 110px);
  height: 2px;
  background: linear-gradient(90deg, transparent, rgba(41, 243, 255, 0.55) 78%, #29f3ff);
  box-shadow: 0 0 8px rgba(41, 243, 255, 0.45);
}

.title-rail.is-right {
  background: linear-gradient(90deg, #29f3ff, rgba(41, 243, 255, 0.55) 22%, transparent);
}

.title-rail::after {
  content: "";
  position: absolute;
  top: 50%;
  width: 7px;
  height: 7px;
  background: #2ff0a0;
  transform: translateY(-50%) rotate(45deg);
  box-shadow: 0 0 10px rgba(47, 240, 160, 0.75);
}

.title-rail.is-left::after { left: -2px; }
.title-rail.is-right::after { right: -2px; }

/* 沿导轨向标题汇聚的能量光点，两侧对称形成仪式感 */
.title-rail::before {
  content: "";
  position: absolute;
  top: 50%;
  width: 11px;
  height: 6px;
  border-radius: 50%;
  background: radial-gradient(circle, #ffffff 0 28%, #7ff0ff 58%, transparent 76%);
  filter: drop-shadow(0 0 6px rgba(41, 243, 255, 0.9));
  transform: translateY(-50%);
  opacity: 0;
  pointer-events: none;
}

.title-rail.is-left::before {
  left: -4px;
  right: auto;
  animation: rail-run-left 3.6s ease-in-out infinite;
}

.title-rail.is-right::before {
  right: -4px;
  left: auto;
  animation: rail-run-right 3.6s ease-in-out infinite;
  animation-delay: 0.25s;
}

.command-title {
  position: relative;
  margin: 0;
  background: linear-gradient(180deg, #ffffff 6%, #e8fbff 38%, #9bebfc 70%, #46cdea);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  font-size: clamp(25px, 2.6em, 42px);
  font-weight: 900;
  line-height: 1;
  letter-spacing: 0.2em;
  padding-left: 0.2em; /* 平衡尾部字距，保持光学居中 */
  filter: drop-shadow(0 0 10px rgba(21, 183, 255, 0.8)) drop-shadow(0 0 26px rgba(47, 240, 160, 0.24));
  white-space: nowrap;
}

/* 字面流光：与底层文字同形的高光带，周期性扫过标题 */
.command-title::after {
  content: attr(data-text);
  position: absolute;
  inset: 0;
  background: linear-gradient(100deg, transparent 36%, rgba(255, 255, 255, 0.95) 47%, rgba(47, 240, 160, 0.65) 53%, transparent 64%);
  background-size: 260% 100%;
  background-repeat: no-repeat;
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  animation: title-shine 5.4s ease-in-out infinite;
  pointer-events: none;
}

@keyframes title-shine {
  0% { background-position: 190% 0; }
  58%, 100% { background-position: -70% 0; }
}

@keyframes rail-run-left {
  0% { left: -4px; opacity: 0; }
  16%, 84% { opacity: 1; }
  100% { left: calc(100% - 7px); opacity: 0; }
}

@keyframes rail-run-right {
  0% { right: -4px; opacity: 0; }
  16%, 84% { opacity: 1; }
  100% { right: calc(100% - 7px); opacity: 0; }
}

.header-meta {
  position: relative;
  z-index: 1;
  grid-column: 3;
  justify-self: end;
  display: flex;
  align-items: center;
  gap: clamp(14px, 1.5em, 28px);
  color: #ebf5ff;
  font-family: "Courier New", monospace;
  font-size: clamp(15px, 1.5em, 24px);
  font-weight: 700;
  white-space: nowrap;
}

.btn-fullscreen {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: clamp(34px, 3vw, 52px);
  height: clamp(34px, 3vw, 52px);
  border: 1px solid rgba(0, 217, 255, 0.65);
  border-radius: 6px;
  background:
    linear-gradient(135deg, rgba(0, 60, 100, 0.7), rgba(0, 30, 60, 0.5)),
    rgba(0, 47, 82, 0.54);
  color: #03dcff;
  box-shadow:
    inset 0 0 16px rgba(0, 191, 255, 0.16),
    0 0 16px rgba(0, 191, 255, 0.12),
    inset 0 1px 0 rgba(103, 232, 249, 0.15);
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
}

.btn-fullscreen::before {
  content: "";
  position: absolute;
  inset: 3px;
  border: 1px solid rgba(103, 232, 249, 0.12);
  border-radius: 3px;
  pointer-events: none;
  transition: border-color 0.3s ease;
}

.btn-fullscreen::after {
  content: "";
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: conic-gradient(from 0deg, transparent 0%, rgba(0, 217, 255, 0.08) 10%, transparent 20%);
  animation: btn-rotate 6s linear infinite;
  pointer-events: none;
}

.btn-fullscreen :deep(.el-icon) {
  position: relative;
  z-index: 1;
  font-size: clamp(16px, 1.4vw, 22px);
  filter: drop-shadow(0 0 4px rgba(3, 220, 255, 0.5));
  transition: all 0.3s ease;
}

.btn-fullscreen:hover {
  border-color: rgba(103, 232, 249, 0.9);
  background:
    linear-gradient(135deg, rgba(0, 80, 130, 0.85), rgba(0, 50, 90, 0.7)),
    rgba(0, 60, 100, 0.7);
  color: #67e8f9;
  box-shadow:
    inset 0 0 20px rgba(0, 191, 255, 0.25),
    0 0 24px rgba(0, 191, 255, 0.3),
    0 0 48px rgba(0, 191, 255, 0.1);
  transform: translateY(-1px);
}

.btn-fullscreen:hover::before {
  border-color: rgba(103, 232, 249, 0.3);
}

.btn-fullscreen:hover :deep(.el-icon) {
  filter: drop-shadow(0 0 8px rgba(103, 232, 249, 0.8));
  transform: scale(1.08);
}

.btn-fullscreen:active {
  transform: translateY(0) scale(0.95);
  box-shadow:
    inset 0 0 24px rgba(0, 191, 255, 0.3),
    0 0 8px rgba(0, 191, 255, 0.15);
  transition-duration: 0.1s;
}

@keyframes btn-rotate {
  to { transform: rotate(360deg); }
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
  grid-template-rows: minmax(0, 1fr);
  gap: clamp(8px, 1.1vh, 16px);
  padding: clamp(10px, 1.2vh, 18px) clamp(18px, 2vw, 36px) clamp(8px, 1vh, 16px);
  overflow: hidden;
}

.main-rect-sweep {
  position: absolute;
  left: clamp(18px, 2vw, 36px);
  right: clamp(18px, 2vw, 36px);
  top: calc(clamp(10px, 1.2vh, 18px) + clamp(108px, 14vh, 150px) + clamp(8px, 1.1vh, 16px));
  bottom: clamp(8px, 1vh, 16px);
  border-radius: 8px;
  background:
    linear-gradient(90deg, transparent 0%, rgba(41, 243, 255, 0.04) 42%, rgba(47, 240, 160, 0.16) 48%, rgba(255, 213, 106, 0.22) 50%, rgba(47, 240, 160, 0.16) 52%, rgba(41, 243, 255, 0.04) 58%, transparent 100%),
    linear-gradient(180deg, transparent 0%, rgba(41, 243, 255, 0.08) 48%, rgba(255, 213, 106, 0.1) 50%, rgba(41, 243, 255, 0.08) 52%, transparent 100%);
  border-left: 1px solid rgba(255, 213, 106, 0.22);
  border-right: 1px solid rgba(47, 240, 160, 0.18);
  box-shadow: inset 0 0 26px rgba(41, 243, 255, 0.08), 0 0 26px rgba(47, 240, 160, 0.08);
  opacity: 0.62;
  animation: rect-sweep 7.2s linear infinite;
  pointer-events: none;
  transform: translateX(-112%);
  will-change: transform, opacity;
}

.main-rect-sweep::before,
.main-rect-sweep::after {
  content: "";
  position: absolute;
  inset: 0;
  border-radius: inherit;
  pointer-events: none;
}

.main-rect-sweep::before {
  background:
    repeating-linear-gradient(180deg, rgba(255, 255, 255, 0.08) 0 1px, transparent 1px 12px),
    repeating-linear-gradient(90deg, rgba(47, 240, 160, 0.1) 0 1px, transparent 1px 72px);
  mix-blend-mode: screen;
  opacity: 0.46;
}

.main-rect-sweep::after {
  width: 10px;
  left: 50%;
  right: auto;
  background: linear-gradient(180deg, transparent, rgba(255, 226, 160, 0.8), transparent);
  box-shadow: 0 0 18px rgba(255, 213, 106, 0.62), 0 0 36px rgba(47, 240, 160, 0.32);
}

.phase-flow-chamber {
  --phase-height: clamp(132px, 15.5vh, 170px);
  --phase-link-color: #52dfff;
  --phase-tab-tint: rgba(35, 167, 200, 0.16);
  position: relative;
  isolation: isolate;
  display: grid;
  grid-template-rows: var(--phase-height) minmax(0, 1fr);
  min-width: 0;
  min-height: 0;
}

.phase-flow-chamber.phase-state-running { --phase-link-color: #ffcc76; --phase-tab-tint: rgba(225, 153, 51, 0.2); }
.phase-flow-chamber.phase-state-done { --phase-link-color: #64e7b0; --phase-tab-tint: rgba(52, 190, 137, 0.16); }

.phase-chamber-surface {
  position: absolute;
  z-index: -1;
  inset: 0;
  width: 100%;
  height: 100%;
  overflow: visible;
  filter: drop-shadow(0 4px 12px rgba(0, 8, 20, 0.28));
  pointer-events: none;
}

.phase-card-strip {
  --phase-gap: clamp(32px, 3.6vw, 64px);
  --phase-card-w: calc((100% - var(--phase-gap) * 3) / 4);
  position: relative;
  display: flex;
  flex-wrap: nowrap;
  align-items: flex-start;
  padding: 10px clamp(24px, 3vw, 48px) 0;
  z-index: 3;
  min-width: 0;
  min-height: 0;
  overflow-x: auto;
  overflow-y: hidden;
  scrollbar-width: none;
}

.phase-card-strip::-webkit-scrollbar {
  display: none;
}

/* 阶段之间的方向导管：底线 + 行进虚线 + 流动光点 + 双层箭头，全部指向下一阶段 */
.phase-sequence-arrow {
  --seq-c1: #35617c;                              /* 上游端（暗） */
  --seq-c2: #5f93b3;                              /* 下游端（亮），明暗对比表达方向 */
  --seq-glow: rgba(103, 232, 249, 0.28);
  --seq-dur: 2.6s;
  --seq-rail-l: 6%;
  --seq-rail-r: 24%;
  position: relative;
  flex: 0 0 var(--phase-gap);
  align-self: center;
  height: calc(100% - 18px);
  overflow: hidden;
}

/* 导管底线 */
.seq-rail {
  position: absolute;
  top: 50%;
  left: var(--seq-rail-l);
  right: var(--seq-rail-r);
  height: 2px;
  transform: translateY(-50%);
  border-radius: 2px;
  background: linear-gradient(90deg, transparent, var(--seq-c1) 16%, var(--seq-c2));
  box-shadow: 0 0 6px var(--seq-glow);
  opacity: 0.9;
}

/* 向下一阶段行进的虚线 */
.seq-flow {
  position: absolute;
  top: 50%;
  left: var(--seq-rail-l);
  right: var(--seq-rail-r);
  height: 2px;
  transform: translateY(-50%);
  background-image: repeating-linear-gradient(90deg, var(--seq-c2) 0 5px, transparent 5px 13px);
  background-size: 13px 100%;
  mask-image: linear-gradient(90deg, transparent, #000 26%, #000 82%, transparent);
  animation: seq-march var(--seq-dur) linear infinite;
}

/* 沿导管飞驰的光点 */
.seq-comet {
  position: absolute;
  top: 50%;
  left: var(--seq-rail-l);
  width: 11px;
  height: 11px;
  margin-top: -5.5px;
  border-radius: 50%;
  background: radial-gradient(circle, #ffffff 0 20%, var(--seq-c2) 52%, transparent 72%);
  filter: drop-shadow(0 0 7px var(--seq-glow));
  opacity: 0;
  animation: seq-comet var(--seq-dur) cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

/* 双层箭头：前实后虚，形成推进感 */
.seq-head {
  position: absolute;
  top: 50%;
  right: 5%;
  width: 14px;
  height: 12px;
  transform: translateY(-50%);
}

.seq-head::before,
.seq-head::after {
  content: "";
  position: absolute;
  top: 50%;
  right: 0;
  width: 9px;
  height: 9px;
  margin-top: -5px;
  border-top: 2px solid var(--seq-c2);
  border-right: 2px solid var(--seq-c2);
  border-radius: 1px;
  transform: rotate(45deg);
  filter: drop-shadow(0 0 5px var(--seq-glow));
}

.seq-head::before {
  right: 7px;
  opacity: 0.4;
  animation: seq-chevron var(--seq-dur) ease-in-out infinite;
}

.seq-head::after {
  animation: seq-chevron var(--seq-dur) ease-in-out infinite;
  animation-delay: calc(var(--seq-dur) / -2);
}

@keyframes seq-march {
  to { background-position: 13px 0; }
}

@keyframes seq-comet {
  0% { opacity: 0; transform: translateX(0) scale(0.55); }
  14% { opacity: 1; transform: translateX(0) scale(1); }
  78% { opacity: 1; }
  100% { opacity: 0; transform: translateX(calc(var(--phase-gap) * 0.6)) scale(0.7); }
}

@keyframes seq-chevron {
  0%, 100% { opacity: 0.32; transform: translateX(0) rotate(45deg); }
  50% { opacity: 1; transform: translateX(3px) rotate(45deg); }
}

/* 已完成：绿色能量稳态输送 */
.phase-sequence-arrow.is-done {
  --seq-c1: #2c8f77;
  --seq-c2: #57e6b6;
  --seq-glow: rgba(47, 240, 160, 0.5);
  --seq-dur: 2.4s;
}

/* 进行中：金色高活跃，节奏最快 */
.phase-sequence-arrow.is-running {
  --seq-c1: #b8802f;
  --seq-c2: #ffd07a;
  --seq-glow: rgba(255, 177, 61, 0.62);
  --seq-dur: 1.5s;
}

.phase-sequence-arrow.is-running .seq-rail { box-shadow: 0 0 10px var(--seq-glow); }

/* 待开始：冷蓝低速，光点更暗 */
.phase-sequence-arrow.is-pending {
  --seq-dur: 3.8s;
}

.phase-sequence-arrow.is-pending .seq-comet { filter: drop-shadow(0 0 4px var(--seq-glow)); opacity: 0.5; }

.phase-card {
  position: relative;
  z-index: 1;
  flex: 1 0 var(--phase-card-w);
  display: grid;
  grid-template-rows: auto minmax(0, 1fr) auto;
  align-content: stretch;
  min-width: 210px;
  min-height: 0;
  height: calc(100% - 18px);
  padding: clamp(7px, 0.72vw, 13px) clamp(10px, 1.15vw, 20px);
  border: 1px solid rgba(72, 124, 177, 0.28);
  border-radius: 10px;
  background: linear-gradient(180deg, #102e49, #081d34);
  box-shadow: inset 0 -3px 0 rgba(65, 120, 170, 0.22);
  overflow: hidden;
  font: inherit;
  text-align: left;
  cursor: pointer;
  transition: border-color 180ms ease, box-shadow 180ms ease;
}

.phase-card:hover { border-color: rgba(105, 219, 248, 0.8); }
.phase-card:focus-visible { outline: 2px solid #ddfbff; outline-offset: 3px; }

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

.phase-card > *:not(.phase-card-grid):not(.phase-accent) {
  position: relative;
  z-index: 1;
}

.phase-card.is-done {
  border-color: rgba(47, 240, 160, 0.42);
  background: linear-gradient(180deg, #104735, #092c2d);
  box-shadow: inset 0 -2px 0 #2bb887;
}
.phase-card.is-running {
  z-index: 3;
  border-color: rgba(255, 176, 64, 0.62);
  background: linear-gradient(140deg, #5c3d1d, #2b251f);
  box-shadow: inset 0 -2px 0 #c9943c;
}
.phase-card.is-pending {
  border-color: rgba(112, 145, 176, 0.28);
  background: linear-gradient(160deg, #17344f, #0b2036);
  box-shadow: inset 0 -2px 0 rgba(112, 145, 176, 0.35);
}

.phase-card.is-running .phase-accent {
  animation: accent-flow 2.2s ease-in-out infinite;
}
.phase-card.active {
  z-index: 4;
  height: 100%;
  padding-bottom: 24px;
  border-color: transparent;
  border-radius: 10px 10px 0 0;
  background: linear-gradient(180deg, var(--phase-tab-tint), transparent 95%);
  box-shadow: none;
}
.phase-card.active .phase-card-grid { mask-image: linear-gradient(#000, transparent); }
.phase-card.active .phase-accent { bottom: auto; top: 0; height: 2px; }

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

.phase-head .phase-name {
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

.phase-number { font-size: 0.78em; font-weight: 500; color: #a0bdcf; }
.phase-card.active .phase-number { color: var(--phase-link-color); }

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
.is-pending .phase-status { color: #f5fbff; background: rgba(112, 145, 176, 0.14); box-shadow: 0 0 12px rgba(138, 207, 255, 0.16); }

.phase-segments {
  display: grid;
  grid-template-columns: repeat(20, 1fr);
  align-self: center;
  gap: 2px;
  margin: clamp(5px, 0.55vh, 9px) 0;
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
  align-self: end;
  gap: clamp(6px, 0.75vw, 12px);
  min-height: 0;
  color: #f5fbff;
  font-family: "Courier New", monospace;
  font-size: clamp(12px, 1.08vw, 18px);
  font-weight: 700;
  line-height: 1;
  text-shadow: 0 0 10px rgba(28, 222, 255, 0.34);
}

.phase-stats span {
  min-width: 0;
  padding: clamp(1px, 0.22vh, 3px) clamp(5px, 0.55vw, 8px);
  border-radius: 8px;
  background: linear-gradient(180deg, rgba(8, 214, 255, 0.1), rgba(5, 45, 90, 0.14));
  box-shadow: inset 0 0 12px rgba(0, 206, 255, 0.1), 0 0 14px rgba(0, 216, 255, 0.06);
  line-height: 1.05;
  white-space: nowrap;
}

.phase-stats b {
  color: #ffffff;
  font-size: 1.24em;
  text-shadow: 0 0 8px rgba(255, 255, 255, 0.64), 0 0 18px rgba(16, 224, 255, 0.52);
}

.is-done .phase-stats b { text-shadow: 0 0 8px rgba(255, 255, 255, 0.6), 0 0 18px rgba(47, 240, 160, 0.48); }
.is-running .phase-stats b { text-shadow: 0 0 8px rgba(255, 255, 255, 0.6), 0 0 18px rgba(255, 177, 61, 0.52); }
.is-pending .phase-stats b { color: #ffffff; text-shadow: 0 0 8px rgba(255, 255, 255, 0.6), 0 0 18px rgba(138, 207, 255, 0.36); }

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
  --arrow-gap: clamp(40px, 4vw, 68px);
  --node-tag-h: clamp(60px, 7.2vh, 86px);
  position: relative;
  display: flex;
  align-items: flex-start;
  padding: clamp(26px, 3vh, 36px) clamp(24px, 3vw, 60px) clamp(16px, 2vh, 26px);
  min-width: 0;
  min-height: 0;
  border-radius: 0 0 14px 14px;
  overflow: hidden;
}

.flow-board-heading {
  position: absolute;
  top: 14px;
  left: clamp(22px, 2.5vw, 42px);
  right: clamp(22px, 2.5vw, 42px);
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  z-index: 4;
}

/* 板头右侧的实时/轮询信号灯 */
.board-signal {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  height: clamp(24px, 2.6vh, 32px);
  padding: 0 clamp(10px, 1vw, 16px);
  border: 1px solid rgba(103, 232, 249, 0.3);
  border-radius: 999px;
  background: linear-gradient(90deg, rgba(7, 50, 96, 0.72), rgba(3, 18, 38, 0.72));
  color: #9fc6da;
  font-size: clamp(11px, 0.86vw, 14px);
  font-weight: 800;
  letter-spacing: 0.14em;
}

.board-signal.live {
  border-color: rgba(33, 246, 158, 0.42);
  color: #21f69e;
  text-shadow: 0 0 8px rgba(33, 246, 158, 0.42);
}

/* ===== 流程板左下角：演练概览（名称 + 进度） ===== */
.flow-brief {
  position: absolute;
  z-index: 5;
  left: clamp(18px, 2vw, 34px);
  bottom: clamp(14px, 1.6vh, 24px);
  width: clamp(206px, 17.5vw, 296px);
  padding: clamp(11px, 1.1vh, 16px) clamp(13px, 1.1vw, 18px);
  border: 1px solid rgba(103, 232, 249, 0.22);
  border-radius: 12px;
  background:
    linear-gradient(160deg, rgba(7, 42, 76, 0.9), rgba(3, 16, 34, 0.92)),
    repeating-linear-gradient(180deg, rgba(103, 232, 249, 0.03) 0 1px, transparent 1px 14px);
  box-shadow: inset 0 0 22px rgba(0, 150, 220, 0.1), 0 12px 30px rgba(0, 0, 0, 0.3);
  backdrop-filter: blur(3px);
  overflow: hidden;
}

/* 面板顶部光带 */
.flow-brief::before,
.flow-log-panel::before {
  content: "";
  position: absolute;
  top: 0;
  left: 12%;
  right: 12%;
  height: 2px;
  background: linear-gradient(90deg, transparent, rgba(103, 232, 249, 0.85), transparent);
}

.brief-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: clamp(6px, 0.7vh, 10px);
  color: #9fc6da;
  font-size: clamp(11px, 0.8vw, 13px);
  font-weight: 700;
  letter-spacing: 0.16em;
}

.brief-sigil {
  width: 6px;
  height: 6px;
  transform: rotate(45deg);
  background: #52dfff;
  box-shadow: 0 0 8px rgba(82, 223, 255, 0.8);
}

.brief-name {
  margin-bottom: clamp(8px, 1vh, 12px);
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  word-break: break-all;
  color: #f5fbff;
  font-size: clamp(15px, 1.24vw, 21px);
  font-weight: 800;
  line-height: 1.25;
  text-shadow: 0 0 10px rgba(41, 243, 255, 0.28);
}

.brief-progress {
  display: flex;
  align-items: center;
  gap: clamp(10px, 1vw, 16px);
}

/* 进度环：conic-gradient 由 --p（百分比数值）驱动 */
.brief-ring {
  --p: 0;
  position: relative;
  flex: 0 0 auto;
  width: clamp(50px, 4.4vw, 68px);
  aspect-ratio: 1;
  display: grid;
  place-items: center;
  border-radius: 50%;
  background: conic-gradient(from -90deg, #2ff0a0 0 calc(var(--p) * 1%), rgba(103, 232, 249, 0.12) calc(var(--p) * 1%));
  filter: drop-shadow(0 0 10px rgba(47, 240, 160, 0.28));
}

.brief-ring::before {
  content: "";
  position: absolute;
  inset: 5px;
  border-radius: 50%;
  background: radial-gradient(circle at 34% 26%, rgba(12, 52, 84, 0.98), rgba(4, 18, 36, 0.96));
  border: 1px solid rgba(103, 232, 249, 0.2);
}

.brief-ring-val {
  position: relative;
  z-index: 1;
  color: #f5fbff;
  font-size: clamp(14px, 1.15vw, 19px);
  font-weight: 800;
  font-variant-numeric: tabular-nums;
}

.brief-ring-val em {
  margin-left: 1px;
  color: #8fb8cd;
  font-size: 0.62em;
  font-style: normal;
}

.brief-meta {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: clamp(3px, 0.4vh, 6px);
  font-size: clamp(11px, 0.85vw, 13.5px);
}

.brief-status {
  align-self: flex-start;
  padding: 2px 9px;
  border: 1px solid currentColor;
  border-radius: 999px;
  color: #52dfff;
  background: rgba(82, 223, 255, 0.12);
  font-weight: 700;
  line-height: 1.5;
  white-space: nowrap;
}

.brief-status.is-pending { color: #8fb8cd; background: rgba(143, 184, 205, 0.1); }
.brief-status.is-running { color: #ffb13d; background: rgba(255, 177, 61, 0.12); }
.brief-status.is-paused { color: #ffd166; background: rgba(255, 209, 102, 0.12); }
.brief-status.is-completed { color: #2ff0a0; background: rgba(47, 240, 160, 0.13); }
.brief-status.is-terminated { color: #ff8f8f; background: rgba(255, 143, 143, 0.12); }

.brief-count,
.brief-clock {
  overflow: hidden;
  color: #cfe6f5;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.brief-count b {
  color: #52dfff;
  font-size: 1.14em;
  font-variant-numeric: tabular-nums;
}

.brief-clock {
  color: #8fb8cd;
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.06em;
}

.brief-bar {
  margin-top: clamp(9px, 1vh, 13px);
  height: 6px;
  border-radius: 999px;
  background: rgba(50, 102, 132, 0.45);
  overflow: hidden;
}

.brief-bar-fill {
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, #146f90, #12d7f5 62%, #2ff0a0);
  box-shadow: 0 0 12px rgba(18, 215, 245, 0.5);
  transition: width 0.7s cubic-bezier(0.22, 1, 0.36, 1);
}

/* ===== 流程板右下角：执行日志（最新在下） ===== */
.flow-log-panel {
  position: absolute;
  z-index: 5;
  right: clamp(18px, 2vw, 34px);
  bottom: clamp(14px, 1.6vh, 24px);
  display: flex;
  flex-direction: column;
  width: clamp(250px, 22vw, 380px);
  height: clamp(160px, 21vh, 250px);
  padding: clamp(9px, 1vh, 14px) clamp(12px, 1vw, 16px) clamp(8px, 0.9vh, 12px);
  border: 1px solid rgba(103, 232, 249, 0.22);
  border-radius: 12px;
  background: linear-gradient(160deg, rgba(6, 36, 66, 0.9), rgba(3, 14, 30, 0.92));
  box-shadow: inset 0 0 22px rgba(0, 150, 220, 0.1), 0 12px 30px rgba(0, 0, 0, 0.3);
  backdrop-filter: blur(3px);
  overflow: hidden;
}

.flow-log-panel::before {
  background: linear-gradient(90deg, transparent, rgba(47, 240, 160, 0.8), transparent);
}

.log-head {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding-bottom: clamp(6px, 0.7vh, 10px);
  margin-bottom: 4px;
  border-bottom: 1px solid rgba(103, 232, 249, 0.16);
}

.log-title {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: #e8fbff;
  font-size: clamp(12px, 0.9vw, 14.5px);
  font-weight: 800;
  letter-spacing: 0.16em;
}

.log-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #2ff0a0;
  box-shadow: 0 0 8px rgba(47, 240, 160, 0.85);
  animation: log-dot-blink 1.6s ease-in-out infinite;
}

.log-tail {
  color: #7fa8bf;
  font-size: clamp(10px, 0.72vw, 12px);
  letter-spacing: 0.12em;
}

.log-body {
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: rgba(103, 232, 249, 0.3) transparent;
  /* 顶部旧日志淡出，底部最新日志最醒目 */
  mask-image: linear-gradient(180deg, transparent 0, #000 16%, #000 100%);
}

.log-body::-webkit-scrollbar { width: 4px; }
.log-body::-webkit-scrollbar-track { background: transparent; }
.log-body::-webkit-scrollbar-thumb { background: rgba(103, 232, 249, 0.28); border-radius: 2px; }

.log-empty {
  margin: 0;
  padding: 10px 2px;
  color: #7fa8bf;
  font-size: clamp(11px, 0.82vw, 13px);
  text-align: center;
}

.log-row {
  display: grid;
  grid-template-columns: auto auto minmax(0, 1fr);
  gap: 8px;
  align-items: baseline;
  padding: 4px 0;
  border-bottom: 1px dashed rgba(103, 232, 249, 0.08);
  font-size: clamp(11px, 0.82vw, 13.2px);
  line-height: 1.35;
  color: #cfe6f5;
  animation: log-row-in 0.42s ease-out;
}

.log-row:last-child { border-bottom: none; }

.log-time {
  color: #6f97ad;
  font-variant-numeric: tabular-nums;
}

.log-icon {
  color: #52dfff;
  text-shadow: 0 0 8px rgba(82, 223, 255, 0.5);
}

.log-msg {
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.log-row.is-warn .log-icon { color: #ffc861; text-shadow: 0 0 8px rgba(255, 200, 97, 0.5); }
.log-row.is-warn .log-msg { color: #ffe1a3; }
.log-row.is-error .log-icon { color: #ff8f8f; text-shadow: 0 0 8px rgba(255, 143, 143, 0.55); }
.log-row.is-error .log-msg { color: #ffc9c9; }

@keyframes log-dot-blink {
  0%, 100% { opacity: 0.4; transform: scale(0.82); }
  50% { opacity: 1; transform: scale(1); }
}

@keyframes log-row-in {
  from { opacity: 0; transform: translateY(7px); }
  to { opacity: 1; transform: translateY(0); }
}

.flow-empty { width: 100%; text-align: center; color: #83aabd; font-size: 16px; }
.flow-board.is-preview .label-pulse { background: var(--phase-link-color); animation: none; box-shadow: 0 0 8px rgba(82, 223, 255, 0.3); }
.flow-board.is-preview .flow-board-label { color: #b2dceb; }

.flow-board.all-done .flow-board-grid {
  background-image:
    linear-gradient(rgba(74, 222, 128, 0.085) 1px, transparent 1px),
    linear-gradient(90deg, rgba(74, 222, 128, 0.07) 1px, transparent 1px);
}

.flow-board.all-done .flow-board-label {
  border-color: rgba(134, 239, 172, 0.52);
  background:
    linear-gradient(90deg, rgba(20, 83, 45, 0.88), rgba(5, 30, 32, 0.72)),
    rgba(5, 30, 32, 0.72);
  box-shadow:
    inset 0 0 14px rgba(74, 222, 128, 0.14),
    0 0 18px rgba(74, 222, 128, 0.16);
}

.flow-board-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(103, 232, 249, 0.06) 1px, transparent 1px),
    linear-gradient(90deg, rgba(103, 232, 249, 0.045) 1px, transparent 1px);
  background-size: 42px 42px;
  background-position: -1px -1px;
  mask-image: radial-gradient(circle at center, black 0 62%, transparent 90%);
  pointer-events: none;
}

.flow-board-label {
  position: relative;
  z-index: 8;
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 8px;
  height: clamp(26px, 2.8vh, 36px);
  padding: 0 clamp(12px, 1.2vw, 20px);
  border: 1px solid rgba(103, 232, 249, 0.44);
  border-radius: 999px;
  background:
    linear-gradient(90deg, rgba(7, 50, 96, 0.88), rgba(3, 18, 38, 0.72)),
    rgba(3, 18, 38, 0.72);
  color: #e8fbff;
  font-size: clamp(13px, 1.08vw, 17px);
  font-weight: 800;
  line-height: 1;
  letter-spacing: 0.08em;
}

.flow-board-label::before,
.flow-board-label::after {
  content: "";
  position: absolute;
  top: 50%;
  width: 18px;
  height: 1px;
  background: linear-gradient(90deg, transparent, rgba(103, 232, 249, 0.78));
  transform: translateY(-50%);
  pointer-events: none;
}

.flow-board-label::before {
  right: calc(100% + 6px);
}

.flow-board-label::after {
  left: calc(100% + 6px);
  background: linear-gradient(90deg, rgba(103, 232, 249, 0.78), transparent);
}

.label-pulse {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: radial-gradient(circle at 35% 30%, #ffd9a0, #ffb75e 62%, #f08c2e);
  box-shadow: 0 0 8px rgba(255, 183, 94, 0.9), 0 0 18px rgba(255, 152, 66, 0.45);
  animation: label-pulse 1.6s ease-in-out infinite;
}

@keyframes label-pulse {
  0%, 100% { opacity: 0.58; transform: scale(0.86); }
  50% { opacity: 1; transform: scale(1.18); }
}

/* overflow-x: hidden 会使 overflow-y 的 visible 计算为 auto，故视口实际为双向裁剪容器。
   顶部内边距需容纳聚焦节点放大 1.3 倍后外溢的齿轮/涟漪（约 tag_h/2×0.3 + 9px×1.3 ≈ 25px）。 */
.flow-viewport {
  position: relative;
  width: 100%;
  overflow-x: hidden;
  overflow-y: visible;
  margin-top: clamp(4px, 1.4vh, 18px);
  padding: calc(var(--node-tag-h) * 0.16 + 16px) 0 clamp(18px, 2.8vh, 34px);
  mask-image: linear-gradient(90deg, transparent 0, #000 9%, #000 91%, transparent 100%);
}

.flow-track {
  display: flex;
  align-items: flex-start;
  transition: transform 0.7s cubic-bezier(0.22, 1, 0.36, 1);
}

/* 缩放锚点固定在环节标签中线上：任意缩放时标签中心与箭头始终保持同一水平线 */
.flow-node-wrap {
  position: relative;
  flex: 0 0 auto;
  width: clamp(176px, 16.5vw, 248px);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  min-width: 0;
  transform-origin: 50% calc(var(--node-tag-h) / 2);
  transition: transform 0.7s cubic-bezier(0.22, 1, 0.36, 1), opacity 0.7s ease;
}

.flow-node {
  position: relative;
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  min-width: 0;
}

.node-tag {
  position: relative;
  z-index: 2;
  display: flex;
  width: 100%;
  height: var(--node-tag-h);
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  padding: clamp(10px, 1.2vh, 16px) clamp(14px, 1.2vw, 24px);
  border-radius: clamp(10px, 1vw, 16px);
  border: 2px solid rgba(0, 210, 255, 0.64);
  color: #12e4ff;
  font-size: clamp(15px, 1.5em, 24px);
  font-weight: 700;
  background: rgba(4, 31, 55, 0.76);
  box-shadow: 0 0 28px rgba(0, 209, 255, 0.18), inset 0 0 18px rgba(0, 209, 255, 0.12);
  overflow: hidden;
  text-shadow: 0 0 10px rgba(0, 211, 255, 0.24);
}

.node-label {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  text-align: center;
  line-height: 1.35;
  word-break: break-all;
  max-width: 7.6em;
  min-width: 0;
}

.flow-node.is-pending {
  color: #f5fbff;
}

.flow-node.is-pending .node-tag {
  border-color: rgba(78, 119, 157, 0.48);
  color: #f5fbff;
  background: linear-gradient(180deg, rgba(12, 42, 72, 0.86), rgba(6, 26, 48, 0.7));
  box-shadow: 0 0 16px rgba(71, 188, 255, 0.14), inset 0 0 18px rgba(103, 232, 249, 0.08);
  text-shadow: 0 0 8px rgba(255, 255, 255, 0.42), 0 0 16px rgba(71, 188, 255, 0.34);
}

.flow-node.is-done .node-tag {
  border-color: rgba(74, 222, 128, 0.88);
  color: #d1fae5;
  background:
    radial-gradient(circle at 20% 18%, rgba(187, 247, 208, 0.22), transparent 32%),
    linear-gradient(180deg, rgba(22, 101, 52, 0.94), rgba(5, 40, 25, 0.74));
  box-shadow:
    0 0 26px rgba(74, 222, 128, 0.36),
    inset 0 0 22px rgba(74, 222, 128, 0.18),
    inset 0 -4px 0 #4ade80;
  text-shadow: 0 0 12px rgba(74, 222, 128, 0.62);
}

.flow-node.is-done .node-tag::after {
  content: "";
  position: absolute;
  top: 0;
  left: -60%;
  width: 40%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(220, 252, 231, 0.32), transparent);
  transform: skewX(-18deg);
  animation: done-shine 3.6s ease-in-out infinite;
  animation-delay: 0.8s;
  pointer-events: none;
}

.flow-node.is-running .node-tag {
  border-color: rgba(255, 177, 61, 0.92);
  color: #ffe1a3;
  padding-right: clamp(24px, 2.2vw, 38px);
  background:
    radial-gradient(circle at 18% 22%, rgba(255, 235, 170, 0.28), transparent 26%),
    linear-gradient(180deg, rgba(114, 67, 12, 0.94), rgba(48, 31, 17, 0.74));
  box-shadow: 0 0 20px rgba(255, 154, 47, 0.26), inset 0 0 24px rgba(255, 177, 61, 0.16), inset 0 -4px 0 #ffb13d;
  text-shadow: 0 0 10px rgba(255, 177, 61, 0.55);
  animation: node-pulse 2s ease-in-out infinite;
}

.flow-node.is-running .node-tag::before {
  content: "";
  position: absolute;
  inset: 3px;
  border-radius: inherit;
  border: 1px solid rgba(255, 226, 160, 0.45);
  opacity: 0.7;
  pointer-events: none;
}

.flow-node.is-running .node-tag::after {
  content: "";
  position: absolute;
  top: -24%;
  bottom: -24%;
  left: -34%;
  width: 28%;
  background: linear-gradient(90deg, transparent, rgba(255, 244, 210, 0.5), transparent);
  transform: skewX(-18deg);
  opacity: 0;
  pointer-events: none;
  animation: node-scan 2.4s ease-in-out infinite;
}

.node-ripple,
.node-motion {
  position: absolute;
  pointer-events: none;
}

.node-ripple {
  top: -9px;
  left: -9px;
  right: -9px;
  height: calc(var(--node-tag-h) + 18px);
  z-index: 1;
  border-radius: clamp(12px, 1.1vw, 18px);
  border: 1px solid rgba(255, 200, 92, 0.66);
  box-shadow: 0 0 18px rgba(255, 177, 61, 0.34);
  animation: node-ripple 1.9s ease-out infinite;
}

.node-motion {
  top: -9px;
  right: -12px;
  z-index: 3;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: grid;
  place-items: center;
  background: radial-gradient(circle, rgba(255, 214, 118, 0.28), rgba(255, 154, 47, 0.08) 62%, transparent 68%);
  filter: drop-shadow(0 0 10px rgba(255, 177, 61, 0.55));
}

.node-gear {
  position: relative;
  width: 17px;
  height: 17px;
  border-radius: 50%;
  background:
    radial-gradient(circle, rgba(255, 248, 212, 0.98) 0 11%, rgba(255, 214, 118, 0.92) 12% 23%, rgba(255, 154, 47, 0.34) 24% 29%, transparent 30%),
    conic-gradient(from 0deg, #fff1bd 0 10deg, transparent 10deg 30deg, #ffb13d 30deg 45deg, transparent 45deg 62deg, #fff1bd 62deg 76deg, transparent 76deg 96deg, #ffb13d 96deg 112deg, transparent 112deg 136deg, #fff1bd 136deg 150deg, transparent 150deg 170deg, #ffb13d 170deg 186deg, transparent 186deg 210deg, #fff1bd 210deg 224deg, transparent 224deg 248deg, #ffb13d 248deg 264deg, transparent 264deg 288deg, #fff1bd 288deg 302deg, transparent 302deg 326deg, #ffb13d 326deg 342deg, transparent 342deg 360deg);
  box-shadow:
    inset 0 0 0 2px rgba(255, 214, 118, 0.72),
    inset 0 0 8px rgba(255, 244, 207, 0.28),
    0 0 10px rgba(255, 177, 61, 0.46);
  animation: node-gear-spin 1.8s linear infinite;
}

.node-live-dot {
  position: absolute;
  right: 0;
  bottom: 1px;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #fff4cf;
  box-shadow: 0 0 0 3px rgba(255, 177, 61, 0.24), 0 0 12px rgba(255, 226, 160, 0.82);
  animation: node-live-blink 1.05s ease-in-out infinite;
}

@keyframes node-pulse {
  0%, 100% { box-shadow: 0 0 20px rgba(255, 154, 47, 0.26), inset 0 0 24px rgba(255, 177, 61, 0.16), inset 0 -4px 0 #ffb13d; }
  50% { box-shadow: 0 0 30px rgba(255, 154, 47, 0.42), inset 0 0 28px rgba(255, 177, 61, 0.24), inset 0 -4px 0 #ffb13d; }
}

@keyframes done-shine {
  0%, 55% { left: -60%; opacity: 0; }
  65% { opacity: 1; }
  85%, 100% { left: 160%; opacity: 0; }
}

@keyframes node-ripple {
  0% { opacity: 0.76; transform: scale(0.96); }
  70%, 100% { opacity: 0; transform: scale(1.18); }
}

@keyframes node-scan {
  0%, 36% { opacity: 0; transform: translateX(0) skewX(-18deg); }
  48% { opacity: 0.85; }
  72%, 100% { opacity: 0; transform: translateX(520%) skewX(-18deg); }
}

@keyframes node-gear-spin {
  to { transform: rotate(360deg); }
}

@keyframes node-live-blink {
  0%, 100% { opacity: 0.5; transform: scale(0.84); }
  50% { opacity: 1; transform: scale(1); }
}

/* ===== 环节节点下挂的任务步骤清单 ===== */
.node-steps {
  --steps-gap: clamp(12px, 1.6vh, 18px);
  position: relative;
  list-style: none;
  margin: var(--steps-gap) 0 0;
  padding: clamp(9px, 1.1vh, 13px) clamp(10px, 0.9vw, 14px);
  width: 100%;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  gap: clamp(5px, 0.65vh, 8px);
  border: 1px solid rgba(78, 119, 157, 0.34);
  border-radius: 10px;
  background:
    linear-gradient(180deg, rgba(10, 34, 60, 0.88), rgba(4, 18, 36, 0.78)),
    repeating-linear-gradient(0deg, rgba(103, 232, 249, 0.035) 0 1px, transparent 1px 20px);
  box-shadow: inset 0 0 14px rgba(0, 150, 220, 0.08), 0 8px 20px rgba(0, 0, 0, 0.22);
}

/* 标签与步骤清单之间的接口线 */
.node-steps::before {
  content: "";
  position: absolute;
  top: calc(-1 * var(--steps-gap));
  left: 50%;
  width: 1px;
  height: var(--steps-gap);
  background: linear-gradient(180deg, rgba(103, 232, 249, 0.55), rgba(103, 232, 249, 0.06));
}

.node-step {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  min-width: 0;
}

.step-ico {
  flex: 0 0 auto;
  margin-top: 0.45em;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #4e779d;
  box-shadow: 0 0 6px rgba(78, 119, 157, 0.55);
}

/* 步骤名最多两行折行展示，避免长文案被截断遮挡 */
.step-name {
  min-width: 0;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
  word-break: break-all;
  font-size: clamp(12px, 0.92vw, 15.5px);
  line-height: 1.32;
  color: #cfe6f5;
}

.node-step.is-done .step-ico {
  background: #4ade80;
  box-shadow: 0 0 7px rgba(74, 222, 128, 0.6);
}

.node-step.is-done .step-name { color: #a9d8bd; }

/* 已完成步骤：文字右侧的打勾徽章 */
.step-check {
  flex: 0 0 auto;
  margin-left: auto;
  margin-top: 0.3em;
  width: clamp(14px, 1.05vw, 18px);
  aspect-ratio: 1;
  display: grid;
  place-items: center;
  border-radius: 50%;
  border: 1px solid rgba(74, 222, 128, 0.55);
  background:
    radial-gradient(circle at 32% 26%, rgba(220, 252, 231, 0.3), transparent 56%),
    linear-gradient(180deg, rgba(22, 101, 52, 0.92), rgba(6, 46, 28, 0.86));
  box-shadow: 0 0 8px rgba(74, 222, 128, 0.38), inset 0 0 6px rgba(74, 222, 128, 0.22);
  animation: step-check-pop 0.45s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.step-check svg {
  width: 62%;
  fill: none;
  stroke: #bbf7d0;
  stroke-width: 2.2;
  stroke-linecap: round;
  stroke-linejoin: round;
  filter: drop-shadow(0 0 3px rgba(74, 222, 128, 0.8));
}

@keyframes step-check-pop {
  0% { transform: scale(0); opacity: 0; }
  60% { transform: scale(1.18); }
  100% { transform: scale(1); opacity: 1; }
}

.node-step.is-skipped .step-ico {
  background: #6b8f99;
  box-shadow: none;
}

.node-step.is-issue .step-ico {
  background: #f87171;
  box-shadow: 0 0 8px rgba(248, 113, 113, 0.65);
}

.node-step.is-issue .step-name { color: #fecaca; }

.node-step.is-running .step-ico {
  background: #ffb13d;
  box-shadow: 0 0 9px rgba(255, 177, 61, 0.8);
  animation: step-ico-pulse 1.2s ease-in-out infinite;
}

.node-step.is-running .step-name {
  color: #ffe1a3;
  font-weight: 700;
  text-shadow: 0 0 8px rgba(255, 177, 61, 0.4);
}

.node-step.is-more .step-ico {
  width: 6px;
  height: 6px;
  background: transparent;
  border: 1px dashed rgba(120, 160, 190, 0.6);
  box-shadow: none;
}

.node-step.is-more .step-name { color: #8fb3c7; font-weight: 600; }

/* 环节状态对步骤清单的着色 */
.flow-node.is-done .node-steps {
  border-color: rgba(74, 222, 128, 0.32);
  background: linear-gradient(180deg, rgba(10, 46, 34, 0.82), rgba(4, 26, 22, 0.76));
  box-shadow: inset 0 0 14px rgba(74, 222, 128, 0.1), 0 8px 20px rgba(0, 0, 0, 0.22);
}

.flow-node.is-done .node-steps::before {
  background: linear-gradient(180deg, rgba(74, 222, 128, 0.5), rgba(74, 222, 128, 0.05));
}

.flow-node.is-running .node-steps {
  border-color: rgba(255, 177, 61, 0.42);
  background: linear-gradient(180deg, rgba(66, 42, 14, 0.82), rgba(30, 22, 12, 0.76));
  box-shadow: inset 0 0 14px rgba(255, 177, 61, 0.1), 0 0 18px rgba(255, 154, 47, 0.1);
}

.flow-node.is-running .node-steps::before {
  background: linear-gradient(180deg, rgba(255, 177, 61, 0.55), rgba(255, 177, 61, 0.06));
}

@keyframes step-ico-pulse {
  0%, 100% { opacity: 0.55; transform: scale(0.8); }
  50% { opacity: 1; transform: scale(1.15); }
}

.flow-arrow {
  --arrow-c1: #53c7e6;
  --arrow-c2: #57c7ff;
  --arrow-glow: rgba(80, 200, 255, 0.4);
  position: relative;
  flex: 0 0 auto;
  align-self: flex-start;
  /* 对齐环节标签中线（track 顶部对齐 + 固定标签高度） */
  margin-top: calc(var(--node-tag-h) / 2 - 2px);
  width: var(--arrow-gap);
  height: 3px;
  border-radius: 2px;
  background: linear-gradient(90deg, var(--arrow-c1), var(--arrow-c2));
  box-shadow: 0 0 10px var(--arrow-glow);
  z-index: 1;
  transition: margin 0.7s cubic-bezier(0.22, 1, 0.36, 1), opacity 0.7s ease;
}

/* 焊接在左侧卡片边缘的菱形接点：一半压在卡片边框下，视觉上无缝 */
.arrow-port {
  position: absolute;
  left: -3px;
  top: 50%;
  width: 6px;
  height: 6px;
  transform: translateY(-50%) rotate(45deg);
  border-radius: 1.5px;
  background: linear-gradient(135deg, #ffffff, var(--arrow-c2));
  box-shadow: 0 0 8px var(--arrow-glow);
  pointer-events: none;
}

/* 箭头三角：尖端没入右侧卡片边框 */
.flow-arrow::after {
  content: "";
  position: absolute;
  right: -1px;
  top: 50%;
  transform: translateY(-50%);
  border-top: 5px solid transparent;
  border-bottom: 5px solid transparent;
  border-left: 9px solid var(--arrow-c2);
  filter: drop-shadow(0 0 4px var(--arrow-glow));
}

/* 沿箭头全程流动的光点 */
.flow-arrow::before {
  content: "";
  position: absolute;
  top: 50%;
  left: 0;
  width: 6px;
  height: 6px;
  margin-top: -3px;
  border-radius: 50%;
  background: radial-gradient(circle, #fff 0 26%, #bff6ff 52%, transparent 74%);
  filter: drop-shadow(0 0 6px rgba(190, 250, 255, 0.9));
  animation: arrow-dot-flow 1.7s linear infinite;
  pointer-events: none;
}

/* 状态着色：已完成节点的出边（绿） */
.flow-arrow.is-done {
  --arrow-c1: #4ade80;
  --arrow-c2: #22c55e;
  --arrow-glow: rgba(74, 222, 128, 0.6);
}

/* 进行中节点的出边（金） */
.flow-arrow.is-running {
  --arrow-c1: #ffd46a;
  --arrow-c2: #ff9a2f;
  --arrow-glow: rgba(255, 177, 61, 0.62);
}

/* 待执行节点的出边（冷蓝） */
.flow-arrow.is-pending {
  --arrow-c1: #3b7f9f;
  --arrow-c2: #4a9ab8;
  --arrow-glow: rgba(90, 170, 210, 0.3);
}

/* ===== 虚拟起止节点（游戏关卡端点风格） ===== */
/* 固定占位高度与环节标签一致，保证徽章中心对齐箭头中线 */
.flow-node.is-virtual-start,
.flow-node.is-virtual-end {
  height: var(--node-tag-h);
  display: grid;
  place-items: center;
}

.virtual-badge {
  --virtual-glow: rgba(0, 209, 255, 0.3);
  position: relative;
  width: 40%;
  aspect-ratio: 1;
  display: grid;
  place-items: center;
  border-radius: 50%;
  animation: virtual-breathe 3.4s ease-in-out infinite;
}

/* 开始：青色传送门 */
.flow-node.is-virtual-start .virtual-badge {
  border: 2px solid rgba(0, 210, 255, 0.58);
  background:
    radial-gradient(circle at 34% 28%, rgba(103, 232, 249, 0.3), transparent 50%),
    linear-gradient(180deg, rgba(9, 42, 74, 0.92), rgba(3, 20, 40, 0.85));
}

/* 结束：绿色终点 */
.flow-node.is-virtual-end .virtual-badge {
  --virtual-glow: rgba(52, 211, 153, 0.28);
  border: 2px solid rgba(125, 240, 200, 0.52);
  background:
    radial-gradient(circle at 34% 28%, rgba(167, 243, 208, 0.26), transparent 50%),
    linear-gradient(180deg, rgba(10, 48, 46, 0.92), rgba(4, 26, 30, 0.85));
  animation-delay: 1.7s;
}

/* 徽章外旋转虚线环：开始顺时针、结束逆时针，形成首尾呼应 */
.flow-node.is-virtual-start .virtual-badge::before,
.flow-node.is-virtual-end .virtual-badge::before {
  content: "";
  position: absolute;
  inset: -8px;
  border-radius: 50%;
  border: 1px dashed rgba(103, 232, 249, 0.4);
  animation: virtual-spin 16s linear infinite;
  pointer-events: none;
}

.flow-node.is-virtual-end .virtual-badge::before {
  border-color: rgba(125, 240, 200, 0.34);
  animation-direction: reverse;
}

.virtual-glyph {
  display: block;
}

/* 开始：播放三角 */
.virtual-glyph-start {
  width: 0;
  height: 0;
  margin-left: 16%;
  border-top: 11px solid transparent;
  border-bottom: 11px solid transparent;
  border-left: 17px solid #7ef0ff;
  filter: drop-shadow(0 0 8px rgba(126, 240, 255, 0.85));
}

/* 结束：终点格纹旗 */
.virtual-glyph-end {
  width: 48%;
  aspect-ratio: 1;
  border-radius: 3px;
  background: conic-gradient(#e6f7ff 0 25%, #123a54 0 50%, #e6f7ff 0 75%, #123a54 0) 0 0 / 50% 50%;
  box-shadow: inset 0 0 0 1.5px rgba(148, 233, 255, 0.5), 0 0 10px rgba(103, 232, 249, 0.3);
}

/* 徽章下方的小标签：绝对定位避免影响垂直居中 */
.virtual-name {
  position: absolute;
  top: calc(100% + 9px);
  left: 50%;
  transform: translateX(-50%);
  margin-left: 0.21em;
  white-space: nowrap;
  font-size: clamp(11px, 0.9vw, 14px);
  font-weight: 600;
  letter-spacing: 0.42em;
  color: rgba(151, 216, 246, 0.78);
  text-shadow: 0 0 10px rgba(80, 190, 255, 0.35);
  pointer-events: none;
}

.flow-node.is-virtual-end .virtual-name {
  color: rgba(167, 236, 208, 0.78);
  text-shadow: 0 0 10px rgba(52, 211, 153, 0.32);
}

/* 虚拟端点的衔接箭头：弱化冷色，暗示边界 */
.flow-arrow.is-virtual {
  --arrow-c1: rgba(70, 145, 178, 0.6);
  --arrow-c2: rgba(94, 176, 210, 0.6);
  --arrow-glow: rgba(90, 170, 210, 0.22);
}

.flow-arrow.is-virtual::before {
  filter: drop-shadow(0 0 4px rgba(120, 200, 235, 0.45));
  opacity: 0.55;
}

@keyframes virtual-breathe {
  0%, 100% { box-shadow: 0 0 14px var(--virtual-glow), inset 0 0 12px var(--virtual-glow); }
  50% { box-shadow: 0 0 30px var(--virtual-glow), inset 0 0 20px var(--virtual-glow); }
}

@keyframes virtual-spin {
  to { transform: rotate(360deg); }
}

@keyframes arrow-dot-flow {
  0% { left: 0; opacity: 0; }
  15% { opacity: 1; }
  85% { opacity: 1; }
  100% { left: calc(100% - 6px); opacity: 0; }
}

/* ===== 信号条（板头实时状态灯） ===== */
.signal-bars {
  display: inline-grid;
  grid-template-columns: repeat(3, 4px);
  align-items: end;
  gap: 3px;
  height: 16px;
}

.signal-bars i {
  display: block;
  width: 4px;
  height: 6px;
  background: #6f9cb4;
  box-shadow: 0 0 6px rgba(111, 156, 180, 0.4);
  animation: signal-rise 1.4s ease-in-out infinite;
}

.signal-bars i:nth-child(2) { height: 10px; animation-delay: 0.16s; }
.signal-bars i:nth-child(3) { height: 15px; animation-delay: 0.32s; }

.board-signal.live .signal-bars i {
  background: #2ff0a0;
  box-shadow: 0 0 8px rgba(47, 240, 160, 0.42);
}

@keyframes header-scan {
  0% { transform: translateX(-120%); opacity: 0; }
  12% { opacity: 1; }
  58% { opacity: 0.55; }
  100% { transform: translateX(420%); opacity: 0; }
}

@keyframes rect-sweep {
  0% { opacity: 0; transform: translateX(-112%); }
  9% { opacity: 0.62; }
  82% { opacity: 0.62; }
  100% { opacity: 0; transform: translateX(112%); }
}

@keyframes accent-flow {
  0%, 100% { opacity: 0.42; transform: scaleX(0.72); }
  50% { opacity: 0.9; transform: scaleX(1); }
}

@keyframes signal-rise {
  0%, 100% { opacity: 0.35; transform: scaleY(0.64); }
  50% { opacity: 1; transform: scaleY(1); }
}

@media (prefers-reduced-motion: reduce) {
  .phase-card, .flow-track, .flow-node-wrap { transition: none !important; }
  .label-pulse { animation: none !important; }
  .virtual-badge,
  .flow-node.is-virtual-start .virtual-badge::before,
  .flow-node.is-virtual-end .virtual-badge::before { animation: none !important; }
  .header-scanline,
  .command-title::after,
  .title-rail::before,
  .main-rect-sweep,
  .phase-card.is-running .phase-accent,
  .flow-node.is-running .node-tag,
  .flow-node.is-running .node-tag::after,
  .flow-node.is-done .node-tag::after,
  .node-ripple,
  .node-gear,
  .node-live-dot,
  .node-step.is-running .step-ico,
  .step-check,
  .seq-flow,
  .seq-comet,
  .seq-head::before,
  .seq-head::after,
  .log-dot,
  .log-row,
  .flow-arrow::before,
  .signal-bars i {
    animation: none !important;
  }
}

@media (max-width: 1180px) {
  .command-header { grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr); gap: 10px; padding-block: 6px; }
  .header-title-shell { gap: 10px; }
  .title-rail { width: 28px; }
  .title-rail::after { width: 5px; height: 5px; }
  .header-meta { justify-content: flex-end; min-width: 0; }
  .phase-card-strip {
    --phase-gap: 32px;
  }
  .phase-card {
    padding: 7px 9px 6px;
  }
  .phase-head { gap: 5px; }
  .phase-head .phase-name {
    font-size: clamp(14px, 1.55vw, 16px);
    letter-spacing: -0.02em;
  }
  .phase-status {
    padding: 3px 6px;
    font-size: 11px;
  }
  .phase-segments {
    gap: 2px;
    margin: 4px 0;
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
  .flow-brief { width: clamp(184px, 22vw, 240px); }
  .flow-log-panel {
    width: clamp(224px, 27vw, 320px);
    height: clamp(138px, 19vh, 200px);
  }
}

@media (max-width: 1060px) {
  .flow-brief { width: 172px; padding-inline: 11px; }
  .flow-brief .brief-clock { display: none; }
  .flow-log-panel { width: clamp(200px, 26vw, 280px); }
  .phase-card { padding-inline: 8px; }
  .phase-head .phase-name { font-size: 14px; }
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
