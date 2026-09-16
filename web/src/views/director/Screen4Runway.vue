<template>
  <div ref="rootRef" class="screen4-runway" :class="{ 'is-complete': phaseComplete, 'is-motion-paused': motionPaused }">
    <div class="runway-ambient" aria-hidden="true"></div>

    <div class="runway-legend" aria-label="跑道状态图例">
      <span class="legend-kicker" aria-hidden="true">状态</span>
      <span class="legend-item is-done"><i></i>已完成</span>
      <span class="legend-item is-running"><i></i>进行中</span>
      <span class="legend-item is-pending"><i></i>待执行</span>
    </div>

    <div class="deck-cell deck-summary runway-summary" :aria-label="`整体进度，${completedSteps}/${totalSteps} 步骤`">
      <span class="summary-kicker">
        <i></i>
        <span>整体进度</span>
      </span>
      <span class="summary-progress">
        <strong>{{ completedSteps }}</strong>
        <span>/ {{ totalSteps }} 步骤</span>
      </span>
    </div>

    <svg
      class="runway-svg"
      :viewBox="`0 0 1040 ${layout.viewBoxHeight}`"
      preserveAspectRatio="xMidYMid meet"
      role="img"
      aria-label="当前阶段环节能量跑道"
    >
      <defs>
        <linearGradient id="s4-lane-base" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0" stop-color="#163e61" stop-opacity=".52" />
          <stop offset=".5" stop-color="#2c78a5" stop-opacity=".9" />
          <stop offset="1" stop-color="#163e61" stop-opacity=".52" />
        </linearGradient>
        <linearGradient id="s4-lane-complete" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0" stop-color="#0aaa69" />
          <stop offset=".5" stop-color="#24dc8d" />
          <stop offset="1" stop-color="#12bd76" />
        </linearGradient>
        <linearGradient id="s4-lane-active" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0" stop-color="#ffb539" stop-opacity=".25" />
          <stop offset=".48" stop-color="#ffd273" />
          <stop offset="1" stop-color="#ff9f1a" stop-opacity=".2" />
        </linearGradient>
        <linearGradient id="s4-lane-progress" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0" stop-color="#ffc55f" />
          <stop offset=".62" stop-color="#ffd273" />
          <stop offset="1" stop-color="#ffe9ad" />
        </linearGradient>
        <radialGradient id="s4-node-lit-core" cx=".5" cy=".42" r=".62">
          <stop offset="0" stop-color="#e8fff5" />
          <stop offset=".42" stop-color="#54f0b2" />
          <stop offset="1" stop-color="#0fa06c" />
        </radialGradient>
        <radialGradient id="s4-milestone-core">
          <stop offset="0" stop-color="#f8ffff" />
          <stop offset=".35" stop-color="#61f5ff" />
          <stop offset="1" stop-color="#1277bd" />
        </radialGradient>
        <linearGradient id="s4-hub-arc" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0" stop-color="#7ceaff" />
          <stop offset=".52" stop-color="#3fd0f7" />
          <stop offset="1" stop-color="#a4f6ff" />
        </linearGradient>
        <linearGradient id="s4-hub-arc-complete" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0" stop-color="#2af0aa" />
          <stop offset=".55" stop-color="#5df7bd" />
          <stop offset="1" stop-color="#8dffcf" />
        </linearGradient>
        <radialGradient id="s4-hub-plate" cx=".5" cy=".36" r=".68">
          <stop offset="0" stop-color="#0e3d63" />
          <stop offset=".58" stop-color="#062b47" />
          <stop offset="1" stop-color="#031724" />
        </radialGradient>
        <filter id="s4-soft-glow" x="-80%" y="-80%" width="260%" height="260%">
          <feGaussianBlur stdDeviation="5" result="blur" />
          <feMerge>
            <feMergeNode in="blur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
        <filter id="s4-strong-glow" x="-100%" y="-100%" width="300%" height="300%">
          <feGaussianBlur stdDeviation="9" result="blur" />
          <feMerge>
            <feMergeNode in="blur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
        <pattern id="s4-runway-grid" width="24" height="24" patternUnits="userSpaceOnUse">
          <path d="M 24 0 L 0 0 0 24" fill="none" stroke="#57d7ff" stroke-opacity=".08" stroke-width="1" />
        </pattern>
      </defs>

      <rect class="runway-grid" x="18" y="18" width="1004" :height="layout.viewBoxHeight - 36" rx="22" />
      <path class="runway-shadow" :d="layout.trackPath" />
      <path class="runway-rail" :d="layout.trackPath" />
      <path class="runway-dashes" :d="layout.trackPath" />
      <path v-if="completedPath" ref="completePathRef" class="runway-complete-path" :d="completedPath" />
      <path v-if="completedPath" ref="completeCoreRef" class="runway-complete-core" :d="completedPath" />
      <path v-if="completedPath" class="runway-complete-flow" :d="completedPath" />
      <path v-if="activePath" class="runway-active-path" :d="activePath" />
      <path
        v-if="activeHop && activeHop.strokeLength > 0.5"
        :key="activeIndex"
        class="runway-active-progress"
        :d="activePath"
        :style="{ strokeDasharray: `${activeHop.strokeLength} 99999` }"
      />
      <path
        v-if="activeHop && activeHop.strokeLength > 0.5"
        :key="`${activeIndex}-core`"
        class="runway-active-progress-core"
        :d="activePath"
        :style="{ strokeDasharray: `${activeHop.strokeLength} 99999` }"
      />
      <g
        v-if="activeHop && activeHop.strokeLength > 0.5"
        class="runway-progress-head-anchor"
        :style="{ transform: `translate(${activeHop.cursor.x}px, ${activeHop.cursor.y}px)` }"
        aria-hidden="true"
      >
        <g class="runway-progress-head">
          <circle class="progress-head-halo" r="11" />
          <circle class="progress-head-core" r="4.5" />
        </g>
      </g>

      <g
        v-for="indicator in turnIndicators"
        :key="`turn-${indicator.row}`"
        class="turn-indicator"
        :class="{ 'is-energized': indicator.energized }"
        :transform="`translate(${indicator.x} ${indicator.y})`"
        aria-hidden="true"
      >
        <circle r="18" />
        <path class="turn-indicator-arrow" d="M -7 -9 L 0 -2 L 7 -9 M -7 1 L 0 8 L 7 1" />
      </g>

      <g
        v-for="item in nodeItems"
        :key="item.key"
        :class="item.className"
        :transform="`translate(${item.point.x} ${item.point.y})`"
      >
        <title>{{ item.node.name }}</title>
        <circle class="node-orbit node-orbit-outer" r="38" />
        <circle class="node-orbit node-orbit-inner" r="29" />
        <circle class="node-halo" r="25" />
        <circle class="node-core" r="17" />
        <path v-if="item.isCompleted" class="node-check" d="M -7 0 L -2 6 L 9 -7" />
        <path v-else-if="item.node.status === 'issue'" class="node-issue-mark" d="M 0 -8 L 0 3 M 0 9 L 0 10" />
        <g v-else-if="item.node.status === 'running'" class="node-energy" aria-hidden="true">
          <rect v-for="bar in 3" :key="bar" :x="-10 + (bar - 1) * 8" :y="3 - bar * 4" width="5" :height="bar * 5" rx="2" />
        </g>
        <g class="runway-node-label" :transform="`translate(0 ${item.labelLines.length > 1 ? -66 : -51})`">
          <text class="runway-node-name" y="0">
            <tspan
              v-for="(line, lineIndex) in item.labelLines"
              :key="line"
              x="0"
              :dy="lineIndex === 0 ? 0 : 19"
            >{{ line }}</tspan>
          </text>
          <path
            class="runway-node-label-rule"
            :d="item.labelLines.length > 1 ? 'M -24 27 L 24 27' : 'M -24 10 L 24 10'"
          />
        </g>
        <g class="runway-node-count" transform="translate(0 54)">
          <rect x="-39" y="-14" width="78" height="28" rx="14" />
          <text y="6">{{ item.node.completed }}<tspan class="count-divider">/</tspan>{{ item.node.total }}</text>
        </g>
      </g>

      <g
        class="runway-milestone"
        :class="phaseComplete ? 'is-completed' : 'is-pending'"
        :transform="`translate(${milestonePoint.x} ${milestonePoint.y})`"
        role="progressbar"
        aria-valuemin="0"
        aria-valuemax="100"
        :aria-valuenow="phaseProgressPercent"
        :aria-label="`当前阶段进度 ${phaseProgressPercent}%`"
      >
        <g class="milestone-dial" :class="{ 'is-absorbing': dialAbsorbing }">
          <circle class="milestone-scan" r="64" />
          <g class="milestone-ticks" aria-hidden="true">
            <line
              v-for="tick in milestoneTicks"
              :key="`milestone-tick-${tick.angle}`"
              :class="{ 'is-major': tick.major }"
              :transform="`rotate(${tick.angle})`"
              x1="0"
              :y1="tick.major ? -52 : -54.5"
              x2="0"
              :y2="tick.major ? -61 : -58"
            />
          </g>
          <circle class="milestone-track" r="46" />
          <circle
            class="milestone-progress"
            r="46"
            :style="{ strokeDasharray: `${milestoneArcLength} ${MILESTONE_CIRCUMFERENCE}` }"
          />
          <circle class="milestone-plate" r="36" />
          <circle class="milestone-plate-edge" r="30" />
          <circle
            class="milestone-spark"
            cy="-46"
            r="4"
            :style="{ transform: `rotate(${phaseProgressPercent * 3.6}deg)` }"
          />
          <text class="milestone-value" y="12"><tspan>{{ phaseProgressPercent }}</tspan><tspan class="milestone-value-unit" dy="6">%</tspan></text>
        </g>
        <path class="milestone-title-rule" d="M -88 77 L -62 77 M 62 77 L 88 77" />
        <text class="milestone-title" y="82">里程碑</text>
        <text v-if="phaseComplete" class="milestone-status" y="104">阶段达成</text>
      </g>

      <g
        v-if="activeHop && activeHop.strokeLength > 0.5"
        class="runway-baton-anchor"
        :style="{ transform: `translate(${activeHop.cursor.x}px, ${activeHop.cursor.y}px) rotate(${activeHop.angle}deg)` }"
        aria-hidden="true"
      >
        <g class="runway-baton">
          <path class="baton-beam" d="M -42 0 L -14 0" />
          <circle class="baton-tail" cx="-40" cy="0" r="3.5" />
          <path class="baton-body" d="M -30 -7 L -9 -7 L 0 0 L -9 7 L -30 7 Z" />
          <circle class="baton-spark" cx="0" cy="0" r="4.5" />
        </g>
      </g>
    </svg>

    <div class="runway-deck" aria-label="跑道信息栏">
      <div class="deck-cell deck-ticker" aria-label="当前环节任务列表">
        <div class="ticker-head">
          <span class="ticker-node">
            <i></i>
            <span>任务详情</span>
          </span>
          <span class="ticker-count" :aria-label="runningSteps.length ? `当前环节已完成 ${tickerDoneCount} 项，共 ${runningSteps.length} 项` : '暂无当前运行环节'" title="当前环节已完成 / 总任务数">
            <em>已完成</em>
            <strong>{{ runningSteps.length ? tickerDoneCount : '—' }}</strong>
            <span class="ticker-count-divider">/</span>
            <span class="ticker-count-total">{{ runningSteps.length || '—' }}</span>
          </span>
        </div>
        <div v-if="visibleSteps.length" class="ticker-viewport">
          <div class="ticker-track">
            <div class="ticker-sequence">
              <span
                v-for="step in visibleSteps"
                :key="step.id"
                class="ticker-chip"
                :class="`is-${step.status}`"
                :data-step-id="step.id"
                :title="step.name"
              >
                <i class="chip-dot" aria-hidden="true"></i>
                <span class="chip-body">
                  <span class="chip-name">{{ truncateScreen4RunwayText(step.name, 25) }}</span>
                  <span class="chip-meta">
                    <span class="chip-operator"><i class="chip-operator-glyph" aria-hidden="true"></i>{{ tickerOperatorText(step) }}</span>
                    <span class="chip-tag">{{ tickerStatusText(step.status) }}</span>
                  </span>
                </span>
              </span>
            </div>
          </div>
        </div>
        <div v-else-if="allTasksCompleted" class="ticker-complete" aria-hidden="true">
          <i></i>
          <span>当前环节所有任务已完成</span>
        </div>
        <div v-else class="ticker-standby" aria-hidden="true"></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import {
  buildScreen4RunwayLayout,
  getScreen4PhaseStepProgress,
  getScreen4RunwayCompletedPathEnd,
  getScreen4RunwayHopProgress,
  getScreen4RunwayPath,
  isScreen4RunwayNodeCompleted,
  splitScreen4RunwayName,
  truncateScreen4RunwayText,
  type Screen4RunwayNode,
  type Screen4RunwayPoint,
} from './screen4Runway'

interface Screen4RunwayStep {
  id: string
  name: string
  status: string
  assignee?: string
}

const props = defineProps<{
  phaseStatus: string
  nodes: Screen4RunwayNode[]
  completedSteps: number
  totalSteps: number
  runningSteps?: Screen4RunwayStep[]
}>()

const rootRef = ref<HTMLElement | null>(null)
const motionPaused = ref(typeof document !== 'undefined' && document.hidden)
function syncMotionVisibility() { motionPaused.value = document.hidden }
onMounted(() => {
  syncMotionVisibility()
  document.addEventListener('visibilitychange', syncMotionVisibility)
})

const layout = computed(() => buildScreen4RunwayLayout(props.nodes.length))
const sourcePhaseProgress = computed(() => getScreen4PhaseStepProgress(props.nodes))
const displayedPhaseCompleted = ref(sourcePhaseProgress.value.completed)
let pendingCompletionAnimations = 0
const phaseProgressPercent = computed(() => {
  const total = sourcePhaseProgress.value.total
  return total ? Math.round((Math.min(displayedPhaseCompleted.value, total) / total) * 100) : 0
})

watch(sourcePhaseProgress, (next, previous) => {
  if (next.total !== previous.total || next.completed < previous.completed) {
    pendingCompletionAnimations = 0
    displayedPhaseCompleted.value = next.completed
    return
  }
  if (pendingCompletionAnimations === 0) displayedPhaseCompleted.value = next.completed
})

watch(() => props.nodes.map(node => node.id).join('|'), () => {
  pendingCompletionAnimations = 0
  displayedPhaseCompleted.value = sourcePhaseProgress.value.completed
})
const phaseComplete = computed(() => {
  const normalizedStatus = props.phaseStatus.toLowerCase()
  return normalizedStatus === 'completed'
    || normalizedStatus === 'done'
    || (props.nodes.length > 0 && props.nodes.every(isScreen4RunwayNodeCompleted))
})

const completedPathEnd = computed(() => getScreen4RunwayCompletedPathEnd(props.nodes))
const completedPath = computed(() => {
  const end = completedPathEnd.value
  return end !== null && end > 0
    ? getScreen4RunwayPath(layout.value.points, 0, end)
    : ''
})

// 完成段描入：同一推进序列新增一段时只描入增量（旧段保持原位），首现/换阶段整段从起点描入。
const completePathRef = ref<SVGPathElement | null>(null)
const completeCoreRef = ref<SVGPathElement | null>(null)
let prevCompletedEnd: number | null = null
let prevCompletedLength = 0

watch(completedPath, () => {
  // 灯芯与导轨共用同一条路径，描入动画必须逐帧同步。
  const elements = [completePathRef.value, completeCoreRef.value]
    .filter((el): el is SVGPathElement => Boolean(el))
  if (!elements.length) {
    prevCompletedEnd = null
    prevCompletedLength = 0
    return
  }
  const end = completedPathEnd.value
  const total = elements[0].getTotalLength()
  const isExtension = prevCompletedEnd !== null && end === prevCompletedEnd + 1 && total > prevCompletedLength
  const startOffset = isExtension ? total - prevCompletedLength : total
  prevCompletedEnd = end
  prevCompletedLength = total
  if (total <= 0) return
  elements.forEach((el) => {
    el.style.transition = 'none'
    el.style.strokeDasharray = `${total} ${total}`
    el.style.strokeDashoffset = `${startOffset}`
  })
  // 强制回流后再恢复过渡，令描入段平滑推进到位。
  elements.forEach((el) => {
    el.getBoundingClientRect()
    el.style.transition = ''
    el.style.strokeDashoffset = '0'
  })
}, { flush: 'post' })

const activeIndex = computed(() => props.nodes.findIndex(node => (
  !isScreen4RunwayNodeCompleted(node)
  && (node.status === 'running' || node.completed > 0)
)))
const activePath = computed(() => {
  if (activeIndex.value < 0 || activeIndex.value >= layout.value.points.length - 1) return ''
  return getScreen4RunwayPath(layout.value.points, activeIndex.value, activeIndex.value + 1)
})

const activeNode = computed(() => (activeIndex.value >= 0 ? props.nodes[activeIndex.value] ?? null : null))
const activeRatio = computed(() => {
  const node = activeNode.value
  if (!node || node.total <= 0) return 0
  return Math.min(1, Math.max(0, node.completed / node.total))
})
const activeHop = computed(() => {
  if (activeIndex.value < 0 || activeIndex.value >= layout.value.points.length - 1) return null
  // 环节间跑道按当前运行环节的步骤完成比例推进。
  return getScreen4RunwayHopProgress(layout.value.points, activeIndex.value, activeIndex.value + 1, activeRatio.value)
})

// 终点仪表盘：进度环半径与周长，弧长按阶段步骤比例推进。
const MILESTONE_RING_RADIUS = 46
const MILESTONE_CIRCUMFERENCE = 2 * Math.PI * MILESTONE_RING_RADIUS
// 每 6° 一格刻度，每 30° 一个主刻度。
const milestoneTicks = Array.from({ length: 60 }, (_, index) => ({
  angle: index * 6,
  major: index % 5 === 0,
}))

const milestonePoint = computed<Screen4RunwayPoint>(() => (
  layout.value.points[props.nodes.length] ?? layout.value.points[layout.value.points.length - 1]
))
const milestoneArcLength = computed(() => (
  MILESTONE_CIRCUMFERENCE * Math.min(1, Math.max(0, phaseProgressPercent.value / 100))
))

const nodeItems = computed(() => props.nodes.map((node, index) => ({
  key: node.id,
  node,
  point: layout.value.points[index],
  labelLines: splitScreen4RunwayName(node.name),
  isCompleted: isScreen4RunwayNodeCompleted(node),
  className: [
    'runway-node',
    isScreen4RunwayNodeCompleted(node) ? 'is-completed' : '',
    index === activeIndex.value ? 'is-running' : '',
    node.status === 'issue' ? 'is-issue' : '',
    node.status === 'pending' && index !== activeIndex.value ? 'is-pending' : '',
  ],
})))

const turnIndicators = computed(() => layout.value.points.flatMap((point, index, points) => {
  const nextPoint = points[index + 1]
  if (!nextPoint || nextPoint.row === point.row) return []

  // 完成段已越过该折返点（路径覆盖到 index+1）时，折返点随之导通。
  const end = completedPathEnd.value
  return [{
    row: point.row,
    x: point.direction === 'right' ? 998 : 42,
    y: (point.y + nextPoint.y) / 2,
    energized: end !== null && end >= index + 1,
  }]
}))

// ===== 底部任务传送带 =====

const runningSteps = computed<Screen4RunwayStep[]>(() => props.runningSteps ?? [])
const tickerDoneCount = computed(() => runningSteps.value.filter(step => isAbsorbedStatus(step.status)).length)

const tickerStatusPriority: Record<string, number> = {
  running: 0,
  issue: 1,
  pending: 2,
}

// 静态任务队列：执行中优先，异常其次，待执行任务保持原业务顺序紧随其后。
const visibleSteps = computed(() => runningSteps.value
  .filter(step => !isAbsorbedStatus(step.status))
  .sort((left, right) => (
    (tickerStatusPriority[left.status] ?? 3) - (tickerStatusPriority[right.status] ?? 3)
  )))

// 当前环节任务收束判定：
// 1. 运行环节的任务全部完成（收束瞬间）；
// 2. 无运行环节但已有环节收束（环节推进间隙 / 阶段末尾）——
//    环节状态切换后 runningSteps 会立即指向下一环节或清空，仅靠条件 1 宣告会一闪而过甚至不出现。
const allTasksCompleted = computed(() => {
  if (visibleSteps.value.length > 0) return false
  if (runningSteps.value.length > 0) return true
  return props.nodes.some(isScreen4RunwayNodeCompleted)
})

const tickerStatusLabels: Record<string, string> = {
  done: '已完成',
  skipped: '已跳过',
  running: '执行中',
  issue: '异常',
  pending: '待执行',
}

function tickerStatusText(status: string) {
  return tickerStatusLabels[status] ?? '待执行'
}

function isAbsorbedStatus(status: string) {
  return status === 'done' || status === 'skipped'
}

// 操作人行：未指派时静默降级，不打断卡片节奏。
function tickerOperatorText(step: Screen4RunwayStep) {
  return step.assignee?.trim() || '未指派'
}

// ===== 任务完成 → 飘入终点百分数环 =====

const dialAbsorbing = ref(false)
let dialAbsorbTimer: ReturnType<typeof setTimeout> | null = null

function prefersReducedMotion() {
  return window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ?? false
}

function playTaskCompletions(steps: Screen4RunwayStep[]) {
  const root = rootRef.value
  const dial = root?.querySelector('.milestone-dial')
  if (!root || !dial || motionPaused.value || prefersReducedMotion()) return
  const rootRect = root.getBoundingClientRect()
  const dialRect = dial.getBoundingClientRect()
  // 父组件在状态写入的同步时刻调用，卡片尚未从任务栏移除，可准确记录起飞位置。
  const launches = steps.flatMap(step => {
    const card = root.querySelector<HTMLElement>(`[data-step-id="${CSS.escape(step.id)}"]`)
    return card ? [{ step, card, from: card.getBoundingClientRect() }] : []
  })
  pendingCompletionAnimations += launches.length
  launches.forEach((launch, index) => {
    window.setTimeout(() => {
      if (motionPaused.value || rootRef.value !== root) return releasePendingMilestoneProgress()
      spawnAbsorbFlyer(root, launch.step, launch.card, launch.from, rootRect, dialRect)
    }, index * 190)
  })
}

function spawnAbsorbFlyer(
  root: HTMLElement,
  step: Screen4RunwayStep,
  card: HTMLElement,
  from: DOMRect,
  rootRect: DOMRect,
  dialRect: DOMRect,
) {
  const flyer = document.createElement('div')
  flyer.className = 'runway-flyer'

  const signal = document.createElement('span')
  signal.className = 'runway-flyer-signal'
  signal.textContent = '✓'
  const copy = document.createElement('span')
  copy.className = 'runway-flyer-copy'
  const title = document.createElement('strong')
  title.textContent = step.name
  const meta = document.createElement('small')
  meta.textContent = tickerOperatorText(step)
  copy.append(title, meta)
  const tag = document.createElement('span')
  tag.className = 'runway-flyer-tag'
  tag.textContent = '已完成'
  flyer.append(signal, copy, tag)
  root.appendChild(flyer)

  const startX = from.left + from.width / 2 - rootRect.left
  const startY = from.top + from.height / 2 - rootRect.top
  const hubX = dialRect.left + dialRect.width / 2 - rootRect.left
  const hubY = dialRect.top + dialRect.height / 2 - rootRect.top
  const showOnRight = hubX + 230 < rootRect.width
  const showX = hubX + (showOnRight ? 126 : -126)
  const showY = hubY
  const controlX = (startX + showX) / 2
  const controlY = Math.min(startY, showY) - 72
  const startTransform = `translate(${startX}px, ${startY}px) translate(-50%, -50%) scale(.96) rotate(-3deg)`
  flyer.style.transform = startTransform
  flyer.style.width = `${Math.min(228, Math.max(176, from.width))}px`

  card.animate([
    { opacity: 1, transform: 'translateY(0) scale(1)' },
    { opacity: 0, transform: 'translateY(5px) scale(.94)' },
  ], { duration: 280, easing: 'ease-in', fill: 'forwards' })

  const flightFrames: Keyframe[] = []
  for (let index = 0; index <= 20; index += 1) {
    const progress = index / 20
    const rest = 1 - progress
    const x = rest * rest * startX + 2 * rest * progress * controlX + progress * progress * showX
    const y = rest * rest * startY + 2 * rest * progress * controlY + progress * progress * showY
    flightFrames.push({
      transform: `translate(${x}px, ${y}px) translate(-50%, -50%) scale(${.96 - progress * .14}) rotate(${-3 * rest}deg)`,
      offset: progress,
    })
  }
  const flight = flyer.animate(flightFrames, {
    duration: 760,
    easing: 'cubic-bezier(.12, .62, .24, 1)',
    fill: 'forwards',
  })
  const trailTimer = window.setInterval(() => spawnRunwayTrail(root, flyer, rootRect), 48)

  flight.onfinish = () => {
    clearInterval(trailTimer)
    flyer.classList.add('is-highlight')
    const done = spawnRunwayDoneBanner(root, step.name, showX, showY + 43)
    const parked = `translate(${showX}px, ${showY}px) translate(-50%, -50%) scale(.82)`
    flyer.animate([
      { transform: parked },
      { transform: `translate(${showX}px, ${showY}px) translate(-50%, -50%) scale(.87)`, offset: .5 },
      { transform: parked },
    ], { duration: 1050, easing: 'ease-in-out', fill: 'forwards' })

    window.setTimeout(() => {
      done.animate([
        { opacity: 1, transform: 'translate(-50%, 0)' },
        { opacity: 0, transform: 'translate(-50%, -8px)' },
      ], { duration: 260, easing: 'ease-in', fill: 'forwards' }).onfinish = () => done.remove()
      flyer.animate([
        { transform: parked, opacity: 1, filter: 'blur(0)' },
        { transform: `translate(${hubX}px, ${hubY}px) translate(-50%, -50%) scale(.12)`, opacity: .05, filter: 'blur(4px)' },
      ], { duration: 440, easing: 'cubic-bezier(.55, 0, .85, .4)', fill: 'forwards' }).onfinish = () => {
        flyer.remove()
        triggerRunwayAbsorption(root, hubX, hubY)
        commitMilestoneProgress()
        pulseMilestoneDial()
      }
    }, 1050)
  }
  flight.oncancel = () => {
    clearInterval(trailTimer)
    flyer.remove()
    releasePendingMilestoneProgress()
  }
}

function commitMilestoneProgress() {
  const target = sourcePhaseProgress.value.completed
  displayedPhaseCompleted.value = Math.min(target, displayedPhaseCompleted.value + 1)
  releasePendingMilestoneProgress()
}

function releasePendingMilestoneProgress() {
  pendingCompletionAnimations = Math.max(0, pendingCompletionAnimations - 1)
  if (pendingCompletionAnimations === 0) {
    displayedPhaseCompleted.value = sourcePhaseProgress.value.completed
  }
}

function spawnRunwayDoneBanner(root: HTMLElement, taskName: string, x: number, y: number) {
  const done = document.createElement('div')
  done.className = 'runway-fly-done'
  const icon = document.createElement('i')
  icon.textContent = '✓'
  const text = document.createElement('span')
  text.textContent = `「${taskName}」已完成`
  done.append(icon, text)
  done.style.left = `${x}px`
  done.style.top = `${y}px`
  root.appendChild(done)
  done.animate([
    { opacity: 0, transform: 'translate(-50%, -10px) scale(.9)' },
    { opacity: 1, transform: 'translate(-50%, 0) scale(1)' },
  ], { duration: 340, easing: 'cubic-bezier(.2, 1.4, .4, 1)', fill: 'forwards' })
  return done
}

function spawnRunwayTrail(root: HTMLElement, flyer: HTMLElement, rootRect: DOMRect) {
  const rect = flyer.getBoundingClientRect()
  const dot = document.createElement('i')
  dot.className = 'runway-fly-trail'
  const size = 3 + Math.random() * 3
  dot.style.width = `${size}px`
  dot.style.height = `${size}px`
  dot.style.left = `${rect.left + rect.width / 2 - rootRect.left}px`
  dot.style.top = `${rect.top + rect.height / 2 - rootRect.top}px`
  root.appendChild(dot)
  const driftX = (Math.random() - .5) * 18
  const driftY = 5 + Math.random() * 12
  dot.animate([
    { opacity: .95, transform: 'translate(-50%, -50%) scale(1)' },
    { opacity: 0, transform: `translate(calc(-50% + ${driftX}px), calc(-50% + ${driftY}px)) scale(.18)` },
  ], { duration: 680, easing: 'ease-out' }).onfinish = () => dot.remove()
}

function triggerRunwayAbsorption(root: HTMLElement, x: number, y: number) {
  for (const [index, size] of [250, 170].entries()) {
    const ring = document.createElement('i')
    ring.className = 'runway-hub-shockwave'
    ring.style.left = `${x}px`
    ring.style.top = `${y}px`
    root.appendChild(ring)
    ring.animate([
      { width: '34px', height: '34px', opacity: .9, borderWidth: '3px' },
      { width: `${size}px`, height: `${size}px`, opacity: 0, borderWidth: '1px' },
    ], { duration: 620 + index * 160, easing: 'cubic-bezier(.2, .8, .3, 1)' }).onfinish = () => ring.remove()
  }

  for (let index = 0; index < 14; index += 1) {
    const angle = index / 14 * Math.PI * 2 + Math.random() * .2
    const distance = 34 + Math.random() * 48
    const particle = document.createElement('i')
    particle.className = 'runway-burst-particle'
    particle.style.left = `${x}px`
    particle.style.top = `${y}px`
    root.appendChild(particle)
    particle.animate([
      { transform: 'translate(-50%, -50%) scale(1)', opacity: 1 },
      { transform: `translate(calc(-50% + ${Math.cos(angle) * distance}px), calc(-50% + ${Math.sin(angle) * distance}px)) scale(.12)`, opacity: 0 },
    ], { duration: 520 + Math.random() * 220, easing: 'ease-out' }).onfinish = () => particle.remove()
  }
}

defineExpose({ playTaskCompletions })

// 环体吸收脉冲：缩放 + 辉光闪烁，与进度弧推进同步。
function pulseMilestoneDial() {
  dialAbsorbing.value = false
  if (dialAbsorbTimer) clearTimeout(dialAbsorbTimer)
  requestAnimationFrame(() => {
    dialAbsorbing.value = true
  })
  dialAbsorbTimer = setTimeout(() => {
    dialAbsorbing.value = false
  }, 700)
}

onUnmounted(() => {
  document.removeEventListener('visibilitychange', syncMotionVisibility)
  if (dialAbsorbTimer) clearTimeout(dialAbsorbTimer)
})
</script>

<style scoped lang="scss">
.screen4-runway {
  position: relative;
  font-family: inherit;
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
  isolation: isolate;
  contain: paint style;
  border-radius: 18px;
  border: 1px solid rgba(65, 188, 238, 0.22);
  background:
    radial-gradient(ellipse at 50% 46%, rgba(15, 85, 141, 0.06), transparent 68%),
    linear-gradient(180deg, rgba(3, 20, 40, 0.06), rgba(2, 15, 32, 0.18));
  box-shadow: inset 0 0 42px rgba(0, 6, 18, 0.16);
}

.runway-ambient {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background: radial-gradient(ellipse at 50% 60%, rgba(27, 94, 150, 0.05), transparent 72%);
}

// 状态图例：左上角 HUD 状态铭牌，与底部信息栏同一套玻璃质感。
// 与右上整体进度同 top、同高，构成严格同一水平线上的镜像双锚点。
.runway-legend {
  position: absolute;
  z-index: 4;
  top: 12px;
  left: 12px;
  box-sizing: border-box;
  height: 38px;
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 0 20px 0 17px;
  border: 1px solid rgba(65, 188, 238, 0.26);
  border-radius: 10px;
  background: linear-gradient(105deg, rgba(4, 26, 45, 0.85), rgba(6, 34, 52, 0.58));
  box-shadow:
    inset 0 1px 0 rgba(140, 224, 255, 0.12),
    inset 0 0 18px rgba(38, 196, 242, 0.05),
    0 10px 26px rgba(0, 7, 18, 0.38);
  backdrop-filter: blur(8px);
  animation: legend-arrive .55s cubic-bezier(.2, .8, .25, 1) both;

  // 左缘能量栏：青→绿渐变锚定"状态"语义。
  &::before {
    content: '';
    position: absolute;
    left: 0;
    top: 10px;
    bottom: 10px;
    width: 3px;
    border-radius: 999px;
    background: linear-gradient(180deg, #37ecb0, #3fd0f7);
    box-shadow: 0 0 10px rgba(56, 231, 167, 0.55);
  }
}

// 图例字号与跑道环节节点名称（18px）同规格，铭牌信息与跑道主体信息同级可读。
.legend-kicker {
  padding-right: 13px;
  border-right: 1px solid rgba(65, 188, 238, 0.18);
  color: #9dd9e8;
  font-size: 18px;
  font-weight: 700;
  letter-spacing: .2em;
}

.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: #8fb6cc;
  font-size: 18px;
  font-weight: 700;
  letter-spacing: .5px;
  white-space: nowrap;

  i {
    width: 11px;
    height: 11px;
    border-radius: 50%;
    box-shadow: 0 0 10px currentColor;
  }

  &.is-done i { color: #38e7a7; background: currentColor; }

  &.is-running i {
    color: #ffb43d;
    background: currentColor;
    animation: ticker-blink 1.4s ease-in-out infinite;
  }

  &.is-pending i { color: #4e98c8; background: currentColor; }
}

.runway-deck {
  position: absolute;
  z-index: 4;
  right: 12px;
  bottom: 10px;
  left: 12px;
  display: flex;
  align-items: stretch;
  gap: 10px;
  // 两行任务卡片需要更高的舞台：传送带整体加高，跑道图形相应上移让位。
  height: 68px;
}

.deck-cell {
  display: flex;
  align-items: center;
  min-width: 0;
  padding: 0 12px;
  border: 1px solid rgba(67, 212, 255, 0.26);
  border-radius: 8px;
  background: linear-gradient(110deg, rgba(4, 37, 61, 0.92), rgba(3, 26, 47, 0.66));
  box-shadow: inset 3px 0 0 rgba(48, 221, 178, 0.7), inset 0 0 18px rgba(38, 196, 242, 0.05), 0 10px 30px rgba(0, 7, 18, 0.3);
  backdrop-filter: blur(8px);
}

.deck-summary {
  flex: 0 0 auto;
  gap: 8px;
}

// 整体进度独立悬浮在右上角：与左侧状态铭牌同线同高，左缘/右缘能量条镜像呼应。
.runway-summary {
  position: absolute;
  z-index: 4;
  top: 12px;
  right: 12px;
  box-sizing: border-box;
  width: fit-content;
  max-width: calc(100% - 430px);
  height: 38px;
  justify-content: flex-start;
  border-color: rgba(65, 188, 238, 0.26);
  border-radius: 10px;
  background: linear-gradient(255deg, rgba(4, 26, 45, 0.85), rgba(6, 34, 52, 0.58));
  box-shadow:
    inset 0 1px 0 rgba(140, 224, 255, .12),
    inset 0 0 18px rgba(38, 196, 242, .05),
    0 10px 26px rgba(0, 7, 18, .38);
  animation: legend-arrive .55s cubic-bezier(.2, .8, .25, 1) both;

  // 右缘能量栏：与图例左栏镜像，绿→青渐变回扣"进度"语义。
  &::before {
    content: '';
    position: absolute;
    right: 0;
    top: 10px;
    bottom: 10px;
    width: 3px;
    border-radius: 999px;
    background: linear-gradient(180deg, #3fd0f7, #37ecb0);
    box-shadow: 0 0 10px rgba(56, 231, 167, 0.55);
  }
}

// “整体进度”主标与跑道节点名称（18px）同级；数字同规格、单位层级降一档保持节奏。
.summary-kicker {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 5px;
  min-width: 0;
  overflow: visible;
  color: #9dd9e8;
  font-size: 18px;
  font-weight: 700;
  letter-spacing: .08em;
  white-space: nowrap;

  i {
    flex: 0 0 auto;
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: #37ecb0;
    box-shadow: 0 0 10px #37ecb0;
  }
}

.summary-progress {
  display: flex;
  align-items: baseline;
  gap: 4px;
  white-space: nowrap;
  font-family: inherit;
  font-variant-numeric: tabular-nums;

  strong {
    color: #58f0b6;
    font-size: 18px;
    font-weight: 800;
    line-height: 1;
    text-shadow: 0 0 10px rgba(56, 231, 167, .42);
  }

  span {
    color: #d9f3ff;
    font-size: 18px;
    font-weight: 700;
    line-height: 1;
  }
}

.deck-ticker {
  position: relative;
  flex: 1 1 auto;
  gap: 12px;
  padding-right: 0;
}

.ticker-head {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 9px;
  padding-right: 12px;
  border-right: 1px solid rgba(67, 212, 255, 0.18);
}

// 任务栏头部主标与跑道节点名称（18px）同规格，栏首信息与跑道主体信息同级可读。
.ticker-node {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  max-width: 200px;
  overflow: hidden;
  color: #d7f4ff;
  font-size: 18px;
  font-weight: 700;
  white-space: nowrap;

  > span {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  i {
    flex: 0 0 auto;
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: currentColor;
    box-shadow: 0 0 8px currentColor;
    animation: ticker-blink 1.4s ease-in-out infinite;
  }
}

.ticker-count {
  display: inline-flex;
  align-items: baseline;
  gap: 6px;
  min-width: 28px;
  padding: 6px 10px;
  border: 1px solid rgba(79, 204, 218, .22);
  border-radius: 7px;
  background: linear-gradient(120deg, rgba(24, 113, 125, .18), rgba(5, 30, 49, .5));
  color: #75aac4;
  font-family: inherit;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;

  strong {
    color: #58f0b6;
    font-size: 18px;
    font-weight: 800;
    line-height: 1;
    text-shadow: 0 0 8px rgba(56, 231, 167, .24);
  }

  em {
    color: #8cb6c7;
    font-family: inherit;
    font-size: 11px;
    font-style: normal;
  }
}

.ticker-count-divider {
  color: #527d94;
  font-size: 14px;
}

.ticker-count-total {
  color: #bdd8e5;
  font-size: 16px;
  font-weight: 600;
}

.ticker-viewport {
  position: relative;
  flex: 1 1 auto;
  min-width: 0;
  overflow: hidden;
}

.ticker-track {
  display: flex;
  align-items: center;
  width: 100%;
  min-width: 0;
  height: 100%;
  contain: paint;
}

.ticker-sequence {
  display: flex;
  align-items: center;
  flex: 1 1 auto;
  min-width: 0;
  gap: 10px;
  overflow: hidden;
}

// 任务卡片：两行"任务铭牌"——上行任务名，下行操作人 + 状态徽章。
// 左缘状态能量轨与图例/进度栏的能量栏同一设计语言，卡片按状态换色。
.ticker-chip {
  position: relative;
  display: inline-flex;
  align-items: flex-start;
  flex: 0 1 auto;
  gap: 9px;
  min-width: 0;
  max-width: 430px;
  padding: 9px 14px 8px 16px;
  border: 1px solid rgba(88, 148, 186, 0.28);
  border-radius: 12px;
  color: #a8cbe0;
  background: linear-gradient(163deg, rgba(9, 38, 62, 0.92), rgba(4, 21, 38, 0.8));
  box-shadow:
    inset 0 1px 0 rgba(140, 224, 255, 0.08),
    0 8px 22px rgba(0, 7, 18, 0.35);
  white-space: nowrap;
  overflow: hidden;

  &::before {
    content: '';
    position: absolute;
    left: 0;
    top: 8px;
    bottom: 8px;
    width: 3px;
    border-radius: 999px;
    background: #4e98c8;
    box-shadow: 0 0 8px rgba(78, 152, 200, 0.5);
  }

  .chip-dot {
    flex: 0 0 auto;
    margin-top: 5px;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #4e98c8;
  }

  .chip-body {
    display: flex;
    flex: 0 1 auto;
    flex-direction: column;
    gap: 3px;
    min-width: 0;
  }

  .chip-name {
    flex: 0 1 auto;
    min-width: 0;
    max-width: 348px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: #d7ecf7;
    font-size: 15px;
    font-weight: 700;
    line-height: 18px;
    letter-spacing: .02em;
  }

  .chip-meta {
    display: inline-flex;
    align-items: center;
    gap: 9px;
    min-width: 0;
    line-height: 14px;
  }

  .chip-operator {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    min-width: 0;
    max-width: 220px;
    overflow: hidden;
    text-overflow: ellipsis;
    color: #8fb9d0;
    font-size: 12px;

    .chip-operator-glyph {
      flex: 0 0 auto;
      width: 9px;
      height: 9px;
      border-radius: 50%;
      background: radial-gradient(circle at 35% 32%, #bfe8fa, #3f7ea6 78%);
      box-shadow: 0 0 6px rgba(80, 200, 240, 0.4);
    }
  }

  .chip-tag {
    flex: 0 0 auto;
    padding: 1px 8px;
    border: 1px solid rgba(111, 147, 168, 0.42);
    border-radius: 999px;
    color: #7fb2cf;
    font-size: 11px;
    letter-spacing: .06em;
  }

  &.is-done {
    border-color: rgba(43, 224, 159, 0.32);

    &::before { background: #38e7a7; box-shadow: 0 0 8px rgba(56, 231, 167, 0.6); }

    .chip-dot { background: #38e7a7; box-shadow: 0 0 8px rgba(56, 231, 167, 0.6); }
    .chip-tag { color: #58c9a4; border-color: rgba(88, 201, 164, 0.42); }
  }

  &.is-skipped {
    opacity: .68;

    &::before { background: #6d93ab; box-shadow: none; }
    .chip-dot { background: #6d93ab; }
  }

  &.is-issue {
    border-color: rgba(255, 98, 110, 0.45);

    &::before { background: #ff626e; box-shadow: 0 0 8px rgba(255, 98, 110, 0.55); }

    .chip-dot { background: #ff626e; box-shadow: 0 0 8px rgba(255, 98, 110, 0.6); }
    .chip-tag { color: #ff8b95; border-color: rgba(255, 139, 149, 0.45); }
  }

  &.is-running {
    border-color: rgba(255, 180, 61, 0.55);
    background: linear-gradient(163deg, rgba(66, 45, 12, 0.55), rgba(26, 19, 8, 0.78));
    box-shadow:
      inset 0 1px 0 rgba(255, 226, 173, 0.14),
      0 0 18px rgba(255, 171, 45, 0.16),
      0 8px 22px rgba(0, 7, 18, 0.35);

    &::before {
      background: linear-gradient(180deg, #ffd273, #ff9f1a);
      box-shadow: 0 0 10px rgba(255, 180, 61, 0.65);
    }

    .chip-name { color: #ffe3ad; }

    .chip-dot {
      background: #ffb43d;
      box-shadow: 0 0 9px rgba(255, 180, 61, 0.8);
      animation: ticker-blink 1.2s ease-in-out infinite;
    }

    .chip-tag {
      color: #ffca70;
      border-color: rgba(255, 202, 112, 0.5);
      background: rgba(255, 180, 61, 0.12);
      font-weight: 700;
    }
  }
}

// 传送带空载待机：以缓慢漂移的虚线车道替代文字占位，静默表达"待命中"。
.ticker-standby {
  position: relative;
  flex: 1 1 auto;
  height: 3px;
  margin: 0 22px;
  border-radius: 999px;
  background: repeating-linear-gradient(90deg, rgba(112, 178, 208, .32) 0 7px, transparent 7px 20px);
  mask-image: linear-gradient(90deg, transparent, #000 14%, #000 86%, transparent);
  animation: standby-drift 2.8s linear infinite;

  // 待命光珠：沿空载车道巡游的微光，保持系统"心跳"。
  &::after {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    width: 26px;
    height: 3px;
    border-radius: 999px;
    background: linear-gradient(90deg, transparent, rgba(140, 225, 255, .9), transparent);
    box-shadow: 0 0 8px rgba(120, 215, 255, .45);
    opacity: 0;
    animation: standby-bead 3.6s cubic-bezier(.45, .05, .55, .95) infinite;
  }
}

// 全部完成：在分隔线右侧以导通的对勾徽记宣告环节收束，与完成链路同色系。
.ticker-complete {
  display: flex;
  align-items: center;
  gap: 9px;
  padding-left: 6px;
  color: #9fe8cd;
  font-size: 18px;
  font-weight: 700;
  letter-spacing: 1px;
  white-space: nowrap;
  animation: legend-arrive .45s cubic-bezier(.2, .8, .25, 1) both;

  i {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex: 0 0 auto;
    width: 17px;
    height: 17px;
    border: 1px solid rgba(66, 240, 164, .55);
    border-radius: 50%;
    background: rgba(13, 66, 51, .45);
    box-shadow: 0 0 10px rgba(56, 231, 167, .35);

    // 纯 CSS 对勾，与节点上的完成盖印同语义。
    &::after {
      content: '';
      width: 6px;
      height: 3.5px;
      border-left: 1.5px solid #42f0a4;
      border-bottom: 1.5px solid #42f0a4;
      transform: rotate(-45deg) translateY(-.5px);
    }
  }
}

.runway-svg {
  position: absolute;
  inset: 2px 4px 82px;
  width: calc(100% - 8px);
  height: calc(100% - 84px);
  // 图形重心偏下（里程碑标题/计数延伸至底部车道下方），整体上移使其在信息栏上方视觉居中；
  // 底部传送带加高后同步抬高跑道，保留呼吸间距。
  transform: translateY(clamp(-54px, -5.5vh, -34px)) scale(1.08);
  transform-origin: center;
  overflow: visible;
  contain: paint;
  font-family: inherit;
}

.runway-grid {
  fill: url('#s4-runway-grid');
  stroke: rgba(68, 195, 239, 0.08);
  opacity: 0.14;
}

.runway-shadow,
.runway-rail,
.runway-dashes,
.runway-complete-path,
.runway-complete-core,
.runway-complete-flow,
.runway-active-path,
.runway-active-progress,
.runway-active-progress-core {
  fill: none;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.runway-shadow {
  stroke: rgba(0, 4, 12, 0.72);
  stroke-width: 27;
}

.runway-rail {
  stroke: url('#s4-lane-base');
  stroke-width: 15;
  filter: url('#s4-soft-glow');
}

.runway-dashes {
  stroke: rgba(146, 221, 247, 0.42);
  stroke-width: 2;
  stroke-dasharray: 2 13;
  animation: runway-drift 3.2s linear infinite;
}

.runway-complete-path {
  stroke: url('#s4-lane-complete');
  // 完整覆盖 15px 蓝轨，完成区间读作一条连续绿色链路。
  stroke-width: 15;
  filter: drop-shadow(0 0 8px rgba(29, 236, 147, .72)) drop-shadow(0 0 16px rgba(15, 189, 115, .38));
  // 新增段由脚本 FLIP 描入（旧段保持原位），无需整体淡入。
  transition: stroke-dashoffset .75s cubic-bezier(.3, .75, .3, 1);
}

.runway-complete-core {
  // 饱和绿灯芯保持导通感，避免近白高光把完成轨冲成青白色。
  stroke: #42f0a4;
  stroke-width: 3;
  opacity: .96;
  filter: drop-shadow(0 0 5px rgba(66, 240, 164, .8));
  transition: stroke-dashoffset .75s cubic-bezier(.3, .75, .3, 1);
}

.runway-complete-flow {
  // 沿导通链巡游的能量光珠：0.1 长度 + 圆头端帽 = 圆点，周期 52.1 与 keyframe 严格一致保证无缝循环。
  stroke: #42f0a4;
  stroke-width: 5.5;
  stroke-dasharray: 0.1 52;
  filter: none;
  animation: completed-energy-flow 1.5s linear infinite;
}

.runway-active-path {
  stroke: url('#s4-lane-active');
  // 前方路径用细虚线，与实心推进条拉开粗细层级（实线=已推进，细虚=去路）。
  stroke-width: 4;
  stroke-dasharray: 4 15;
  opacity: .24;
  filter: none;
  animation: active-energy 1.8s linear infinite;
}

.runway-active-progress {
  stroke: url('#s4-lane-progress');
  stroke-width: 13;
  filter: drop-shadow(0 0 7px rgba(255, 176, 54, .82)) drop-shadow(0 0 15px rgba(255, 143, 25, .42));
  // 每完成一步，能量条平滑推进到新的比例位置。
  transition: stroke-dasharray .55s cubic-bezier(.3, .75, .3, 1);
}

.runway-active-progress-core {
  // 与外层填充共用同一比例裁剪，形成清晰连续的橙金灯芯。
  stroke: #ffe29a;
  stroke-width: 3;
  filter: drop-shadow(0 0 5px rgba(255, 218, 128, .9));
  transition: stroke-dasharray .55s cubic-bezier(.3, .75, .3, 1);
}

.runway-progress-head-anchor {
  transition: transform .55s cubic-bezier(.3, .75, .3, 1);
}

.runway-progress-head {
  transform-box: fill-box;
  transform-origin: center;
  will-change: transform, opacity;
  animation: progress-head-pulse 1.6s ease-in-out infinite;
}

.progress-head-halo {
  fill: rgba(255, 187, 71, .14);
  stroke: rgba(255, 214, 130, .5);
  stroke-width: 1;
}

.progress-head-core {
  fill: #ffe9ad;
  filter: none;
}

.turn-indicator {
  circle {
    fill: rgba(5, 35, 56, 0.92);
    stroke: rgba(77, 198, 237, 0.5);
  }

  path {
    fill: none;
    stroke: #72dfff;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
    filter: url('#s4-soft-glow');
  }
}

// 完成链越过折返点：半透明绿罩让导轨与光珠透出，读作链路穿针而过。
.turn-indicator.is-energized {
  circle {
    fill: rgba(7, 34, 26, 0.55);
    stroke: rgba(61, 240, 174, 0.72);
  }

  path {
    stroke: #a9ffd6;
  }
}

.runway-node {
  --node-color: #4e98c8;
  --node-soft: rgba(48, 134, 188, 0.22);
  color: var(--node-color);
  animation: node-arrive .55s cubic-bezier(.2, .8, .25, 1) both;
}

.runway-node.is-completed {
  --node-color: #38e7a7;
  --node-soft: rgba(43, 224, 159, 0.25);
}

// 完成节点核心点亮：与绿色导轨连成一条完整能量链，节点读作链上导通的灯。
.runway-node.is-completed .node-core {
  fill: url('#s4-node-lit-core');
  stroke: #a4ffd9;
  stroke-width: 1.5;
}

// 亮核心上的对勾压深色，读作盖印确认。
.runway-node.is-completed .node-check {
  stroke: #04301f;
}

.runway-node.is-completed .node-orbit-inner {
  stroke-width: 2;
  opacity: .9;
}

// 外层虚线轨道环缓慢旋转，表达“持续带电”。
.runway-node.is-completed .node-orbit-outer {
  animation: node-orbit-spin 12s linear infinite;
}

.runway-node.is-running {
  --node-color: #ffb43d;
  --node-soft: rgba(255, 171, 45, 0.32);
}

.runway-node.is-running .node-orbit-outer {
  animation: node-pulse 1.8s ease-in-out infinite;
}

.runway-node.is-issue {
  --node-color: #ff626e;
  --node-soft: rgba(255, 77, 91, 0.27);
}

.runway-node.is-pending {
  opacity: .8;
}

.node-orbit,
.node-halo,
.node-core {
  vector-effect: non-scaling-stroke;
}

.node-orbit {
  fill: none;
  stroke: currentColor;
}

.node-orbit-outer {
  stroke-width: 1;
  stroke-dasharray: 2 6;
  opacity: .5;
}

.node-orbit-inner {
  stroke-width: 1.5;
  opacity: .78;
}

.node-halo {
  fill: var(--node-soft);
  stroke: currentColor;
  stroke-width: 5;
  filter: url('#s4-soft-glow');
}

.node-core {
  fill: #071f35;
  stroke: currentColor;
  stroke-width: 2;
}

.node-check,
.node-issue-mark {
  fill: none;
  stroke: currentColor;
  stroke-width: 3;
  stroke-linecap: round;
  stroke-linejoin: round;
  filter: url('#s4-soft-glow');
}

.node-energy {
  fill: currentColor;
  filter: url('#s4-soft-glow');

  rect {
    transform-box: fill-box;
    transform-origin: center bottom;
    animation: energy-bar .9s ease-in-out infinite alternate;
  }

  rect:nth-child(2) { animation-delay: -.25s; }
  rect:nth-child(3) { animation-delay: -.5s; }
}

.runway-node-label-rule {
  fill: none;
  stroke: currentColor;
  stroke-width: 1;
  stroke-linecap: round;
  opacity: .28;
}

.runway-node-name {
  fill: #d8f3ff;
  font-size: 18px;
  font-weight: 700;
  letter-spacing: .8px;
  text-anchor: middle;
  text-shadow: 0 2px 8px rgba(0, 7, 15, .75);
}

.runway-node-count {
  rect {
    fill: rgba(3, 27, 46, .92);
    stroke: currentColor;
    stroke-opacity: .38;
  }

  text {
    fill: currentColor;
    font-family: inherit;
    font-variant-numeric: tabular-nums;
    font-size: 17px;
    font-weight: 700;
    letter-spacing: 1px;
    text-anchor: middle;
  }

  .count-divider {
    opacity: .42;
  }
}

.runway-milestone {
  color: #49dff5;
}

.milestone-scan {
  fill: none;
  stroke: rgba(73, 217, 249, .5);
  stroke-width: 1;
  stroke-dasharray: 2 9;
  transform-box: fill-box;
  transform-origin: center;
  will-change: transform;
  animation: milestone-scan-spin 16s linear infinite;
}

.milestone-ticks line {
  stroke: rgba(103, 213, 245, .3);
  stroke-width: 1;
  stroke-linecap: round;

  &.is-major {
    stroke: rgba(147, 235, 255, .58);
    stroke-width: 1.5;
  }
}

.milestone-track {
  fill: none;
  stroke: rgba(62, 152, 199, .26);
  stroke-width: 7;
}

.milestone-progress {
  fill: none;
  stroke: url('#s4-hub-arc');
  stroke-width: 7;
  stroke-linecap: round;
  filter: url('#s4-soft-glow');
  transform: rotate(-90deg);
  animation: path-arrive .8s ease both;
  // 每完成一步，终点进度环平滑伸展。
  transition: stroke-dasharray .6s cubic-bezier(.3, .75, .3, 1);
}

.milestone-plate {
  fill: url('#s4-hub-plate');
  stroke: rgba(73, 221, 255, .48);
  stroke-width: 1;
}

.milestone-plate-edge {
  fill: none;
  stroke: rgba(214, 249, 255, .1);
  stroke-width: 1;
}

.milestone-spark {
  fill: url('#s4-milestone-core');
  filter: url('#s4-strong-glow');
  // 光点沿进度环随比例移动。
  transition: transform .6s cubic-bezier(.3, .75, .3, 1);
}

.milestone-value {
  fill: #58e8ff;
  font-family: inherit;
  font-variant-numeric: tabular-nums;
  font-size: 34px;
  font-weight: 700;
  text-anchor: middle;
  filter: url('#s4-soft-glow');
}

.milestone-value-unit {
  font-size: 13px;
}

.milestone-title-rule {
  fill: none;
  stroke: currentColor;
  stroke-opacity: .32;
  stroke-width: 1;
  stroke-linecap: round;
}

.milestone-title,
.milestone-status {
  text-anchor: middle;
}

.milestone-title {
  fill: #e3f8ff;
  font-size: 16px;
  font-weight: 700;
  letter-spacing: 3px;
}

.milestone-status {
  fill: currentColor;
  font-size: 12px;
  letter-spacing: 1px;
}

.runway-milestone.is-completed {
  color: #39e8aa;

  .milestone-scan {
    stroke: rgba(72, 235, 186, .42);
  }

  .milestone-ticks line {
    stroke: rgba(88, 235, 195, .3);

    &.is-major {
      stroke: rgba(150, 248, 214, .6);
    }
  }

  .milestone-progress {
    stroke: url('#s4-hub-arc-complete');
  }

  .milestone-plate {
    stroke: rgba(69, 237, 178, .55);
  }

  .milestone-value {
    fill: #65f2bd;
  }
}

.runway-baton-anchor {
  transform-box: view-box;
  transform-origin: 0 0;
  transition: transform .55s cubic-bezier(.3, .75, .3, 1);
}

.runway-baton {
  color: #ffc55f;
  filter: none;
  animation: baton-hover 1.5s ease-in-out infinite;
}

// 环体内层组：吸收脉冲以局部原点（环节圆心）缩放。
.milestone-dial.is-absorbing {
  animation: milestone-absorb .66s cubic-bezier(.2, .85, .3, 1);
}

.baton-beam {
  fill: none;
  stroke: currentColor;
  stroke-width: 2;
  stroke-dasharray: 4 6;
  animation: active-energy .8s linear infinite;
}

.baton-tail,
.baton-spark { fill: #fff0bc; }
.baton-body { fill: currentColor; }

@keyframes runway-drift {
  to { stroke-dashoffset: -30; }
}

@keyframes active-energy {
  to { stroke-dashoffset: -42; }
}

@keyframes completed-energy-flow {
  // -52.1 = 光珠 dash 周期(0.1+52)，保证循环无缝。
  to { stroke-dashoffset: -52.1; }
}

@keyframes progress-head-pulse {
  0%, 100% { opacity: .58; transform: scale(.86); }
  50% { opacity: 1; transform: scale(1.12); }
}

@keyframes path-arrive {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes node-arrive {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes node-pulse {
  0%, 100% { opacity: .45; transform: scale(.96); }
  50% { opacity: 1; transform: scale(1.12); }
}

// -64 = 轨道环虚线周期(2+6) 的 8 倍，保证旋转无缝。
@keyframes node-orbit-spin {
  to { stroke-dashoffset: -64; }
}

@keyframes energy-bar {
  from { transform: scaleY(.48); opacity: .58; }
  to { transform: scaleY(1); opacity: 1; }
}

@keyframes baton-hover {
  0%, 100% { opacity: .62; }
  50% { opacity: 1; }
}

@keyframes milestone-scan-spin {
  to { transform: rotate(360deg); }
}

@keyframes ticker-blink {
  0%, 100% { opacity: 1; }
  50% { opacity: .35; }
}

// -20px = 待机车道虚线周期(7+13)，保证漂移无缝循环。
@keyframes standby-drift {
  to { background-position: -20px 0; }
}

@keyframes standby-bead {
  from { left: -30px; opacity: 0; }
  18%, 82% { opacity: 1; }
  to { left: 100%; opacity: 0; }
}

@keyframes legend-arrive {
  from { opacity: 0; transform: translateY(-8px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes milestone-absorb {
  0% { transform: scale(1); }
  38% {
    transform: scale(1.14);
    filter: drop-shadow(0 0 16px rgba(103, 232, 249, .85)) drop-shadow(0 0 26px rgba(47, 240, 160, .5));
  }
  100% { transform: scale(1); }
}

@media (max-width: 1280px) {
  .runway-deck {
    gap: 8px;
    height: 64px;
  }

  // 窄屏下图例收窄而非隐藏：字号与桌面端保持一致，仅收紧留白（高度不变以维持双锚点同线）。
  .runway-legend {
    gap: 10px;
    padding: 0 14px 0 13px;
  }

  .legend-kicker {
    padding-right: 9px;
  }

  .runway-svg {
    inset: 0 2px 78px;
    width: calc(100% - 4px);
    height: calc(100% - 78px);
  }

  .runway-dashes { animation-duration: 4.8s; }
  .runway-complete-flow { animation-duration: 2.4s; }
  .runway-active-path { animation-duration: 2.6s; }
  .runway-node.is-completed .node-orbit-outer { animation-duration: 18s; }
  .ticker-standby { animation-duration: 4.2s; }
  .ticker-standby::after { animation-duration: 5.2s; }
}

.screen4-runway.is-motion-paused *,
.screen4-runway.is-motion-paused *::before,
.screen4-runway.is-motion-paused *::after {
  animation-play-state: paused !important;
}

@media (prefers-reduced-motion: reduce) {
  .runway-complete-path,
  .runway-complete-core,
  .runway-baton-anchor {
    transition: none !important;
  }

  .runway-legend,
  .legend-item.is-running i,
  .runway-dashes,
  .runway-complete-flow,
  .runway-active-path,
  .runway-progress-head,
  .runway-node,
  .node-orbit-outer,
  .node-energy rect,
  .milestone-scan,
  .milestone-dial.is-absorbing,
  .ticker-node i,
  .ticker-chip.is-running .chip-dot,
  .ticker-standby,
  .ticker-standby::after,
  .ticker-complete,
  .runway-baton,
  .baton-beam {
    animation: none !important;
  }
}

</style>

<style lang="scss">
// 飘入元素由脚本动态创建（不携带 scoped 标记），使用全局样式。
.runway-flyer {
  position: absolute;
  top: 0;
  left: 0;
  z-index: 40;
  box-sizing: border-box;
  display: flex;
  align-items: center;
  gap: 9px;
  min-height: 54px;
  padding: 8px 10px;
  border: 1px solid rgba(73, 221, 255, .68);
  border-radius: 10px;
  color: #dffbff;
  background:
    linear-gradient(110deg, rgba(13, 66, 73, .96), rgba(5, 27, 43, .97)),
    repeating-linear-gradient(90deg, rgba(103, 232, 249, .05) 0 1px, transparent 1px 18px);
  box-shadow:
    0 0 0 1px rgba(73, 221, 255, .12),
    0 0 22px rgba(48, 215, 242, .44),
    0 12px 30px rgba(0, 7, 18, .52),
    inset 3px 0 #2ee8e0,
    inset 0 0 18px rgba(32, 190, 203, .14);
  font-family: inherit;
  pointer-events: none;
  will-change: transform, opacity, filter;
  overflow: hidden;

  &::after {
    content: '';
    position: absolute;
    inset: 0;
    background: linear-gradient(105deg, transparent 28%, rgba(198, 252, 255, .22) 48%, transparent 68%);
    transform: translateX(-120%);
    animation: runway-flyer-scan 1.1s ease-in-out infinite;
  }

  &.is-highlight {
    border-color: rgba(66, 240, 164, .85);
    box-shadow:
      0 0 0 1px rgba(66, 240, 164, .24),
      0 0 28px rgba(47, 240, 160, .58),
      0 0 56px rgba(47, 240, 160, .25),
      inset 3px 0 #42f0a4,
      inset 0 0 20px rgba(47, 240, 160, .18);
  }
}

.runway-flyer-signal {
  position: relative;
  z-index: 1;
  flex: 0 0 auto;
  display: grid;
  place-items: center;
  width: 23px;
  height: 23px;
  border: 1px solid rgba(125, 255, 211, .7);
  border-radius: 50%;
  color: #063326;
  background: radial-gradient(circle at 35% 30%, #dffff3, #42f0a4 56%, #13a875);
  box-shadow: 0 0 12px rgba(66, 240, 164, .7);
  font-size: 13px;
  font-weight: 900;
}

.runway-flyer-copy {
  position: relative;
  z-index: 1;
  flex: 1 1 auto;
  min-width: 0;

  strong,
  small {
    display: block;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    color: #edfffa;
    font-size: 14px;
    line-height: 1.25;
    text-shadow: 0 0 8px rgba(66, 240, 164, .34);
  }

  small {
    margin-top: 3px;
    color: #85bdd0;
    font-size: 10px;
  }
}

.runway-flyer-tag {
  position: relative;
  z-index: 1;
  flex: 0 0 auto;
  padding: 2px 7px;
  border: 1px solid rgba(66, 240, 164, .5);
  border-radius: 999px;
  color: #aaffda;
  background: rgba(34, 197, 94, .13);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: .08em;
}

.runway-fly-done {
  position: absolute;
  z-index: 41;
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 6px 11px 6px 7px;
  border: 1px solid rgba(66, 240, 164, .56);
  border-radius: 999px;
  color: #caffea;
  background: linear-gradient(135deg, rgba(9, 55, 43, .96), rgba(4, 24, 37, .96));
  box-shadow: 0 0 18px rgba(47, 240, 160, .36);
  font-family: inherit;
  font-size: 11px;
  font-weight: 700;
  white-space: nowrap;
  pointer-events: none;

  i {
    display: grid;
    place-items: center;
    width: 17px;
    height: 17px;
    border-radius: 50%;
    color: #073324;
    background: #42f0a4;
    box-shadow: 0 0 9px rgba(66, 240, 164, .7);
    font-size: 10px;
    font-style: normal;
  }
}

.runway-fly-trail,
.runway-burst-particle,
.runway-hub-shockwave {
  position: absolute;
  z-index: 39;
  pointer-events: none;
}

.runway-fly-trail {
  border-radius: 50%;
  background: radial-gradient(circle, #efffff 0 18%, #58e8ff 42%, transparent 74%);
  box-shadow: 0 0 7px rgba(73, 221, 255, .78);
  transform: translate(-50%, -50%);
}

.runway-hub-shockwave {
  box-sizing: border-box;
  border-style: solid;
  border-color: rgba(73, 221, 255, .72);
  border-radius: 50%;
  box-shadow: 0 0 20px rgba(73, 221, 255, .5), inset 0 0 18px rgba(47, 240, 160, .3);
  transform: translate(-50%, -50%);
}

.runway-burst-particle {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: radial-gradient(circle, #fff4c9, #ffc55f 42%, #42f0a4 76%);
  box-shadow: 0 0 9px rgba(255, 197, 95, .74), 0 0 15px rgba(66, 240, 164, .36);
  transform: translate(-50%, -50%);
}

@keyframes runway-flyer-scan {
  to { transform: translateX(120%); }
}

@media (prefers-reduced-motion: reduce) {
  .runway-flyer,
  .runway-fly-done,
  .runway-fly-trail,
  .runway-hub-shockwave,
  .runway-burst-particle {
    display: none !important;
  }
}
</style>
