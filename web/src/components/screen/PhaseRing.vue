<template>
  <div
    ref="phaseRingRef"
    class="phase-ring"
    :class="[phaseDirectionDown ? 'phase-dir-down' : 'phase-dir-up', `phase-stage-${safeCurrentIndex}`]"
    :style="{
      width: containerSize.width + 'px',
      height: containerSize.height + 'px',
      '--connector-y': connectorY,
    }"
  >
    <div class="relay-runway">
      <div class="runway-head">
        <div class="runway-title">当前阶段所属环节</div>
        <div class="head-stats">
          <div class="stats-block stats-done">
            <span class="stats-num">{{ phaseCompletedCount }}</span>
            <span class="stats-label">已完成</span>
          </div>
          <span class="stats-sep">/</span>
          <div class="stats-block">
            <span class="stats-num">{{ currentNodes.length }}</span>
            <span class="stats-label">环节</span>
          </div>
        </div>
      </div>

      <!-- 数据汇聚点 - 阶段整体进度 -->
      <div class="progress-hub" :class="{ 'is-done': progress >= 100 }">
        <!-- 数据汇流粒子 -->
        <div class="flow-particles">
          <div class="flow-particle flow-particle-1"></div>
          <div class="flow-particle flow-particle-2"></div>
          <div class="flow-particle flow-particle-3"></div>
          <div class="flow-particle flow-particle-4"></div>
          <div class="flow-particle flow-particle-5"></div>
          <div class="flow-particle flow-particle-6"></div>
        </div>
        <div class="hub-glow"></div>
        <div class="hub-rings">
          <div class="hub-ring hub-ring-1"></div>
          <div class="hub-ring hub-ring-2"></div>
          <div class="hub-ring hub-ring-3"></div>
        </div>
        <div class="hub-core">
          <span class="hub-num">{{ progress }}</span>
          <span class="hub-unit">%</span>
        </div>
      </div>

      <svg ref="runwaySvgRef" class="runway-svg" viewBox="0 0 1040 500" preserveAspectRatio="xMidYMid meet" aria-label="当前阶段接力能量跑道">
        <defs>
          <linearGradient id="lane-base" x1="0" y1="0" x2="1" y2="0">
            <stop offset="0" stop-color="#123050" />
            <stop offset=".5" stop-color="#23618d" />
            <stop offset="1" stop-color="#123050" />
          </linearGradient>
          <linearGradient id="lane-aura" x1="0" y1="0" x2="1" y2="0">
            <stop offset="0" stop-color="#23f0ff" stop-opacity=".12" />
            <stop offset=".5" stop-color="#6bd6ff" stop-opacity=".32" />
            <stop offset="1" stop-color="#ffbd62" stop-opacity=".22" />
          </linearGradient>
          <linearGradient id="lane-done" x1="0" y1="0" x2="1" y2="0">
            <stop offset="0" stop-color="#09b86d" />
            <stop offset="1" stop-color="#73ffc0" />
          </linearGradient>
          <linearGradient id="lane-active" x1="0" y1="0" x2="1" y2="0">
            <stop offset="0" stop-color="#ff8426" />
            <stop offset="1" stop-color="#ffe0a2" />
          </linearGradient>
          <linearGradient id="crown-gold" x1="0" y1="-20" x2="0" y2="14" gradientUnits="userSpaceOnUse">
            <stop offset="0" stop-color="#fff5bd" />
            <stop offset=".28" stop-color="#ffd36f" />
            <stop offset=".58" stop-color="#b86b16" />
            <stop offset="1" stop-color="#f2b64d" />
          </linearGradient>
          <linearGradient id="crown-ridge" x1="-26" y1="0" x2="26" y2="0" gradientUnits="userSpaceOnUse">
            <stop offset="0" stop-color="#7a3f0b" />
            <stop offset=".2" stop-color="#fff0a8" />
            <stop offset=".5" stop-color="#d8861f" />
            <stop offset=".8" stop-color="#fff0a8" />
            <stop offset="1" stop-color="#7a3f0b" />
          </linearGradient>
          <radialGradient id="crown-gem" cx="50%" cy="35%" r="65%">
            <stop offset="0" stop-color="#ffffff" />
            <stop offset=".34" stop-color="#7ff6ff" />
            <stop offset="1" stop-color="#0b77a4" />
          </radialGradient>
          <filter id="crown-glow" x="-60%" y="-70%" width="220%" height="240%">
            <feGaussianBlur stdDeviation="1.5" result="blur" />
            <feMerge>
              <feMergeNode in="blur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
          <filter id="relay-glow" x="-40%" y="-40%" width="180%" height="180%">
            <feGaussianBlur stdDeviation="4" result="blur" />
            <feMerge>
              <feMergeNode in="blur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
          <filter id="milestone-glow" x="-80%" y="-80%" width="260%" height="260%">
            <feGaussianBlur stdDeviation="6" result="blur" />
            <feMerge>
              <feMergeNode in="blur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
        </defs>

        <path
          :d="trackPath"
          fill="none"
          stroke="url(#lane-aura)"
          stroke-width="60"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="lane-aura"
        />
        <path
          :d="trackPath"
          fill="none"
          stroke="rgba(8, 23, 45, .96)"
          stroke-width="46"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
        <path
          :d="trackPath"
          fill="none"
          stroke="url(#lane-base)"
          stroke-width="24"
          stroke-linecap="round"
          stroke-linejoin="round"
          opacity=".82"
        />
        <path
          :d="trackPath"
          fill="none"
          stroke="rgba(226, 247, 255, .28)"
          stroke-width="2"
          stroke-linecap="round"
          stroke-dasharray="22 28"
          class="lane-dash"
        />
        <g class="runway-svg-turn-pips" aria-hidden="true">
          <rect
            v-for="pip in turnPips"
            :key="pip.key"
            class="runway-svg-turn-pip"
            :class="`runway-svg-turn-pip-${pip.side}`"
            :x="pip.x - 9"
            :y="pip.y - 2"
            width="18"
            height="4"
            rx="2"
          />
        </g>
        <path
          v-if="donePath"
          :d="donePath"
          fill="none"
          stroke="url(#lane-done)"
          stroke-width="10"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
        <path
          v-if="activePath"
          :d="activePath"
          fill="none"
          stroke="url(#lane-active)"
          stroke-width="10"
          stroke-linecap="round"
          stroke-linejoin="round"
        />

        <g font-family="Microsoft YaHei, PingFang SC, sans-serif" text-anchor="middle">
          <g
            v-for="node in visibleNodes"
            :key="node.index"
            :transform="`translate(${node.x} ${node.y})`"
            :class="['runway-node', `runway-node-${node.visualStatus}`]"
          >
            <circle
              v-if="node.visualStatus === 'running'"
              r="30"
              fill="none"
              stroke="#ffbd62"
              stroke-width="3"
              class="node-pulse-ring"
            />
            <circle
              v-else-if="node.visualStatus === 'completed'"
              r="25"
              fill="rgba(73, 255, 166, .12)"
              stroke="rgba(73, 255, 166, .35)"
              stroke-width="1.5"
            />
            <circle
              v-else-if="node.visualStatus === 'issue'"
              r="25"
              fill="rgba(255, 85, 85, .11)"
              stroke="rgba(255, 85, 85, .38)"
              stroke-width="1.5"
            />
            <rect
              v-if="node.visualStatus === 'running'"
              x="-38"
              y="-17"
              width="76"
              height="34"
              rx="17"
              class="baton"
            />
            <circle
              v-else
              r="19"
              :class="['node-core', `node-core-${node.visualStatus}`]"
            />
            <path
              v-if="node.visualStatus === 'completed'"
              d="M-8 0 l6 6 12-14"
              fill="none"
              stroke="#baffdd"
              stroke-width="3"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
            <path
              v-else-if="node.visualStatus === 'running'"
              d="M-18 0 H16 M5 -9 L17 0 L5 9"
              fill="none"
              stroke="#fff1cb"
              stroke-width="3"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="baton-arrow"
            />
            <text
              class="node-label"
              :y="node.labelAbove ? -32 : 44"
              :fill="node.labelColor"
            >
              {{ node.shortName }}
            </text>
          </g>

          <g :transform="`translate(${finishPoint.x} ${finishPoint.y})`" class="finish-node" :class="{ 'finish-node-done': isCurrentPhaseDone }">
            <circle r="34" fill="none" stroke="rgba(255, 214, 130, .34)" stroke-width="2" class="finish-beacon-ring finish-beacon-ring-a" />
            <circle r="27" fill="none" stroke="rgba(45, 228, 255, .28)" stroke-width="1.6" class="finish-beacon-ring finish-beacon-ring-b" />
            <g class="finish-sparks">
              <line x1="0" y1="-35" x2="0" y2="-42" />
              <line x1="28" y1="-28" x2="34" y2="-34" />
              <line x1="38" y1="0" x2="46" y2="0" />
              <line x1="28" y1="28" x2="34" y2="34" />
              <line x1="-28" y1="28" x2="-34" y2="34" />
              <line x1="-38" y1="0" x2="-46" y2="0" />
              <line x1="-28" y1="-28" x2="-34" y2="-34" />
            </g>
            <line
              x1="0"
              y1="-38"
              x2="0"
              y2="-29"
              stroke="rgba(255, 213, 129, .72)"
              stroke-width="2.4"
              stroke-linecap="round"
              class="finish-tether"
            />
            <circle r="34" fill="rgba(255, 180, 74, .11)" stroke="rgba(255, 214, 130, .45)" stroke-width="2" class="finish-disc" />
            <circle r="24" fill="none" stroke="rgba(45, 228, 255, .42)" stroke-width="1.5" stroke-dasharray="12 8" class="finish-orbit" />
            <path class="finish-crown-shadow" d="M-27 -5 L-20 15 H20 L27 -5 L14 1 L7 -15 L0 -3 L-7 -15 L-14 1 Z" />
            <path
              class="finish-crown"
              d="M-27 -5 L-20 15 H20 L27 -5 L14 1 L7 -15 L0 -3 L-7 -15 L-14 1 Z"
            />
            <path class="finish-crown-ridge" d="M-19 14 H19 M-14 1 L-20 15 M0 -3 V15 M14 1 L20 15" />
            <ellipse class="finish-crown-gem finish-crown-gem-main" cx="0" cy="-3" rx="3.5" ry="4.4" />
            <circle class="finish-crown-gem" cx="-7" cy="-15" r="3" />
            <circle class="finish-crown-gem" cx="7" cy="-15" r="3" />
            <circle class="finish-crown-gem" cx="-27" cy="-5" r="2.8" />
            <circle class="finish-crown-gem" cx="27" cy="-5" r="2.8" />
            <text class="finish-text" y="75">{{ isCurrentPhaseDone ? '完成' : '里程碑' }}</text>
          </g>
        </g>
      </svg>

      <div class="runway-foot">
        <div class="foot-status">
          <span class="status-dot"></span>
          <span class="status-text">{{ instanceName || '未命名演练' }}</span>
        </div>
        <div class="foot-phase-name">
          <span class="phase-name-tag">当前阶段</span>
          <span class="phase-name-text">{{ currentPhaseLabel }}</span>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'

type NodeStatus = { status: string; progress: number }
type TrackPoint = { x: number; y: number }

const LANE_CAPACITY = 4
const LANE_LEFT = 118
const LANE_RIGHT = 900
const LANE_NODE_PADDING = 56
const TURN_LEFT = 46
const TURN_RIGHT = 998
const DASH_VISUAL_OFFSET_Y = -3

const props = defineProps<{
  phases: string[]
  phaseNames: string[][]
  phaseNodeStatuses: NodeStatus[][]
  phaseStatuses?: string[]
  currentIndex: number
  progress: number
  centerNumerator: number
  centerDenominator: number
  centerHint: string
  instanceName?: string
  size?: number
}>()

const size = computed(() => props.size ?? 520)
const containerSize = computed(() => ({
  width: Math.max(480, Math.round(size.value * 1.72)),
  height: '100%',
}))

const currentNodes = computed(() => props.phaseNames?.[props.currentIndex] || [])
const currentStatuses = computed(() => props.phaseNodeStatuses?.[props.currentIndex] || [])
const safeCurrentIndex = computed(() => Math.max(0, Math.min(props.currentIndex, 3)))
const phaseDirectionDown = computed(() => safeCurrentIndex.value < 2)
const phaseRingRef = ref<HTMLElement | null>(null)
const runwaySvgRef = ref<SVGSVGElement | null>(null)
const connectorY = ref('62px')
let connectorResizeObserver: ResizeObserver | null = null

const currentPhaseLabel = computed(() => {
  return props.phases?.[safeCurrentIndex.value] || props.centerHint || '当前阶段'
})

const laneY = computed(() => {
  const startY = 72
  const endY = 428
  const rowCount = Math.max(1, Math.ceil(Math.max(currentNodes.value.length, 1) / LANE_CAPACITY))
  if (rowCount === 1) return [startY]
  return Array.from({ length: rowCount }, (_, row) => (
    Math.round(startY + ((endY - startY) * row) / (rowCount - 1))
  ))
})

function updateConnectorY() {
  const svg = runwaySvgRef.value
  const phaseRing = phaseRingRef.value
  if (!svg || !phaseRing) return

  const scale = Math.min(svg.clientWidth / 1040, svg.clientHeight / 500)
  if (!Number.isFinite(scale) || scale <= 0) return

  const svgTop = svg.getBoundingClientRect().top - phaseRing.getBoundingClientRect().top
  const yOffset = (svg.clientHeight - 500 * scale) / 2
  const y = svgTop + yOffset + laneY.value[0] * scale + DASH_VISUAL_OFFSET_Y
  connectorY.value = `${Math.round(y * 10) / 10}px`
}

onMounted(() => {
  nextTick(updateConnectorY)
  if (runwaySvgRef.value && 'ResizeObserver' in window) {
    connectorResizeObserver = new ResizeObserver(updateConnectorY)
    connectorResizeObserver.observe(runwaySvgRef.value)
  }
})

onBeforeUnmount(() => {
  connectorResizeObserver?.disconnect()
  connectorResizeObserver = null
})

watch(laneY, () => nextTick(updateConnectorY), { flush: 'post' })

const isCurrentPhaseDone = computed(() => {
  const status = props.phaseStatuses?.[props.currentIndex]
  if (status === 'done' || status === 'completed') return true
  return currentNodes.value.length > 0 && currentNodes.value.every((_, idx) => isDone(currentStatuses.value[idx]))
})

const phaseCompletedCount = computed(() => {
  return currentNodes.value.filter((_, idx) => isDone(currentStatuses.value[idx])).length
})

// 当前运行环节的进度（用于其他逻辑）
const completionPercent = computed(() => {
  const statuses = currentStatuses.value
  if (statuses.length === 0) return 0

  // 找到当前正在运行的环节节点
  const runningNode = statuses.find(s => s?.status === 'running')
  if (runningNode) {
    // 显示当前运行环节的进度
    return Math.min(100, runningNode.progress || 0)
  }

  // 没有运行中的环节时显示 0
  return 0
})

const activeNodeIndex = computed(() => {
  const running = currentStatuses.value.findIndex(s => s?.status === 'running')
  if (running >= 0) return running
  const firstPending = currentStatuses.value.findIndex(s => !isDone(s))
  if (firstPending >= 0) return firstPending
  return Math.max(0, currentNodes.value.length - 1)
})

const trackPoints = computed<TrackPoint[]>(() => {
  const count = Math.max(currentNodes.value.length, 1)
  return Array.from({ length: count }, (_, i) => pointAt(i))
})

const laneNodeCounts = computed(() => {
  const count = Math.max(currentNodes.value.length, 1)
  const rowCount = Math.max(1, Math.ceil(count / LANE_CAPACITY))
  const baseCount = Math.floor(count / rowCount)
  const extraCount = count % rowCount
  return Array.from({ length: rowCount }, (_, row) => baseCount + (row < extraCount ? 1 : 0))
})

const trackPath = computed(() => {
  const lanes = laneY.value
  let d = `M ${LANE_LEFT} ${lanes[0]} H ${LANE_RIGHT}`
  for (let row = 1; row < lanes.length; row += 1) {
    const prevY = lanes[row - 1]
    const y = lanes[row]
    const midY = Math.round((prevY + y) / 2)
    if (row % 2 === 1) {
      d += ` Q ${TURN_RIGHT} ${prevY} ${TURN_RIGHT} ${midY} Q ${TURN_RIGHT} ${y} ${LANE_RIGHT} ${y} H ${LANE_LEFT}`
    } else {
      d += ` Q ${TURN_LEFT} ${prevY} ${TURN_LEFT} ${midY} Q ${TURN_LEFT} ${y} ${LANE_LEFT} ${y} H ${LANE_RIGHT}`
    }
  }
  return d
})

const turnPips = computed(() => {
  return laneY.value.slice(1).map((y, index) => ({
    key: `turn-${index}`,
    x: index % 2 === 0 ? TURN_RIGHT : TURN_LEFT,
    y: Math.round((laneY.value[index] + y) / 2),
    side: index % 2 === 0 ? 'right' : 'left',
  }))
})

const visibleNodes = computed(() => {
  return currentNodes.value.map((name, index) => {
    const status = currentStatuses.value[index]
    const visualStatus = visualStatusOf(status, index)
    const p = trackPoints.value[index]
    return {
      ...p,
      index,
      shortName: shorten(name),
      visualStatus,
      labelAbove: shouldLabelAbove(index, p),
      labelColor: '#f2f8ff',
    }
  })
})

const donePath = computed(() => {
  const indexes = currentStatuses.value
    .map((status, idx) => (isDone(status) ? idx : -1))
    .filter(idx => idx >= 0)
  return indexes.length > 1 ? trackPathThroughIndexes(indexes) : ''
})

const activePath = computed(() => {
  const activeIdx = activeNodeIndex.value
  if (activeIdx <= 0) return ''
  return trackPathThroughIndexes([activeIdx - 1, activeIdx])
})

const finishPoint = computed(() => {
  const lastPoint = trackPoints.value[trackPoints.value.length - 1] || { x: LANE_RIGHT, y: laneY.value[laneY.value.length - 1] }
  return {
    x: lastPoint.x,
    y: lastPoint.y + 86,
  }
})

function pointAt(index: number): TrackPoint {
  const { row, laneIndex, laneCount } = locateLaneNode(index)
  const y = laneY.value[row] ?? laneY.value[laneY.value.length - 1]
  const reversed = row % 2 === 1
  return lanePoint(
    laneIndex,
    laneCount,
    reversed ? LANE_RIGHT : LANE_LEFT,
    reversed ? LANE_LEFT : LANE_RIGHT,
    y,
  )
}

function locateLaneNode(index: number): { row: number; laneIndex: number; laneCount: number } {
  let offset = index
  for (let row = 0; row < laneNodeCounts.value.length; row += 1) {
    const laneCount = laneNodeCounts.value[row]
    if (offset < laneCount) {
      return { row, laneIndex: offset, laneCount }
    }
    offset -= laneCount
  }
  const row = laneNodeCounts.value.length - 1
  return { row, laneIndex: Math.max(0, laneNodeCounts.value[row] - 1), laneCount: laneNodeCounts.value[row] }
}

function lanePoint(index: number, count: number, startX: number, endX: number, y: number): TrackPoint {
  const direction = Math.sign(endX - startX) || 1
  const laneStart = startX + direction * LANE_NODE_PADDING
  const laneEnd = endX - direction * LANE_NODE_PADDING
  const ratio = count > 1 ? index / (count - 1) : 0.5
  const x = laneStart + (laneEnd - laneStart) * ratio
  return { x, y }
}

function trackPathThroughIndexes(indexes: number[]): string {
  const first = trackPoints.value[indexes[0]]
  if (!first) return ''

  let d = `M ${first.x} ${first.y}`
  for (let i = 1; i < indexes.length; i += 1) {
    d += trackPathSegment(indexes[i - 1], indexes[i])
  }
  return d
}

function trackPathSegment(fromIndex: number, toIndex: number): string {
  if (toIndex <= fromIndex) {
    const to = trackPoints.value[toIndex]
    return to ? ` L ${to.x} ${to.y}` : ''
  }

  let d = ''
  for (let index = fromIndex; index < toIndex; index += 1) {
    d += adjacentTrackPathSegment(index, index + 1)
  }
  return d
}

function adjacentTrackPathSegment(fromIndex: number, toIndex: number): string {
  const from = trackPoints.value[fromIndex]
  const to = trackPoints.value[toIndex]
  if (!from || !to) return ''

  const fromLane = locateLaneNode(fromIndex)
  const toLane = locateLaneNode(toIndex)
  if (fromLane.row === toLane.row) {
    return ` L ${to.x} ${to.y}`
  }

  const lanes = laneY.value
  const prevY = lanes[fromLane.row]
  const y = lanes[toLane.row]
  const midY = Math.round((prevY + y) / 2)
  const turnX = fromLane.row % 2 === 0 ? TURN_RIGHT : TURN_LEFT
  const laneEndX = fromLane.row % 2 === 0 ? LANE_RIGHT : LANE_LEFT
  return ` L ${laneEndX} ${prevY} Q ${turnX} ${prevY} ${turnX} ${midY} Q ${turnX} ${y} ${laneEndX} ${y} L ${to.x} ${to.y}`
}

function shouldLabelAbove(_index: number, _point: TrackPoint): boolean {
  return true
}

function visualStatusOf(status: NodeStatus | undefined, index: number): 'completed' | 'running' | 'issue' | 'pending' {
  if (isDone(status)) return 'completed'
  if (status?.status === 'issue' || status?.status === 'timeout') return 'issue'
  if (status?.status === 'running' || index === activeNodeIndex.value) return 'running'
  return 'pending'
}

function isDone(status: NodeStatus | undefined): boolean {
  return status?.status === 'completed' || status?.status === 'done'
}

function shorten(name: string): string {
  return name.length > 5 ? `${name.slice(0, 5)}…` : name
}

</script>

<style lang="scss" scoped>
.phase-ring {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: stretch;
  justify-content: center;
  max-width: 100%;
  min-height: 0;
  font-family: 'Microsoft YaHei', 'PingFang SC', Arial, sans-serif;
}

.phase-ring::before {
  content: '';
  position: absolute;
  left: -4px;
  top: var(--connector-y, 62px);
  width: 48px;
  height: 8px;
  transform: translateY(-50%);
  border-radius: 999px;
  background:
    radial-gradient(circle at 4px 50%, #7dffc6 0 3px, rgba(45, 228, 255, 0.86) 3.2px 4px, transparent 4.2px),
    linear-gradient(
      90deg,
      rgba(45, 228, 255, 0.18),
      rgba(45, 228, 255, 0.68) 19%,
      rgba(125, 255, 198, 0.94) 50%,
      rgba(45, 228, 255, 0.68) 81%,
      rgba(45, 228, 255, 0.08)
    ) 0 50% / 100% 2px no-repeat;
  box-shadow:
    0 0 8px rgba(45, 228, 255, 0.58),
    0 0 16px rgba(125, 255, 198, 0.2);
  z-index: 4;
}

.phase-ring::after {
  content: '';
  position: absolute;
  left: -4px;
  top: var(--connector-y, 62px);
  width: 8px;
  height: 8px;
  transform: translateY(-50%);
  border-radius: 50%;
  background: #7dffc6;
  border: 1px solid rgba(45, 228, 255, 0.86);
  box-shadow:
    0 0 10px rgba(125, 255, 198, 0.72),
    0 0 18px rgba(45, 228, 255, 0.42);
  z-index: 5;
}

.relay-runway {
  position: relative;
  width: calc(100% - 8px);
  height: 100%;
  min-height: 0;
  overflow: hidden;
  border: 1px solid rgba(45, 228, 255, 0.32);
  border-radius: 18px;
  background:
    linear-gradient(90deg, rgba(45, 228, 255, 0.06) 1px, transparent 1px),
    linear-gradient(180deg, rgba(45, 228, 255, 0.04) 1px, transparent 1px),
    radial-gradient(ellipse at 55% 54%, rgba(0, 212, 255, 0.2), transparent 56%),
    linear-gradient(180deg, rgba(4, 24, 56, 0.78), rgba(2, 10, 28, 0.9));
  background-size: 44px 100%, 100% 34px, auto, auto;
  box-shadow:
    inset 0 0 36px rgba(0, 212, 255, 0.12),
    inset 0 -46px 90px rgba(0, 10, 26, 0.72),
    0 18px 46px rgba(0, 0, 0, 0.18);
}

.relay-runway::after {
  content: '';
  position: absolute;
  inset: 10px;
  border: 1px solid rgba(45, 228, 255, 0.12);
  border-radius: 14px;
  clip-path: polygon(0 0, 18% 0, 18% 2px, 82% 2px, 82% 0, 100% 0, 100% 100%, 0 100%);
  pointer-events: none;
}

.relay-runway::before {
  content: '';
  position: absolute;
  inset: -24% -34%;
  background:
    conic-gradient(from 92deg at 50% 50%, transparent, rgba(45, 228, 255, 0.12), transparent 34%, rgba(45, 228, 255, 0.08), transparent 62%);
  opacity: 0.28;
  pointer-events: none;
}

.runway-head {
  position: absolute;
  top: 14px;
  left: 16px;
  right: 40px;
  z-index: 3;
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 56px;
  pointer-events: none;
}

.runway-title {
  padding: 5px 12px 6px;
  color: #ffffff;
  font-size: clamp(16px, 1.6em, 28px);
  font-weight: 900;
  letter-spacing: 2px;
  border-left: 3px solid #2de4ff;
  background: linear-gradient(90deg, rgba(45, 228, 255, 0.16), rgba(45, 228, 255, 0.02));
  box-shadow: 0 0 18px rgba(45, 228, 255, 0.08);
  text-shadow: 0 0 10px rgba(64, 170, 255, 0.8);
}

// 右上角统计信息
.head-stats {
  box-sizing: border-box;
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-height: 56px;
  padding: 5px 14px;
  border: 1px solid rgba(45, 228, 255, 0.26);
  border-radius: 14px;
  background: rgba(8, 24, 56, 0.52);
  box-shadow: 0 0 14px rgba(45, 228, 255, 0.14);
}

.head-stats .stats-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.head-stats .stats-num {
  font-family: Consolas, Menlo, Monaco, 'Courier New', monospace;
  font-size: clamp(18px, 1.6em, 26px);
  font-weight: 900;
  color: #eaf8ff;
  text-shadow: 0 0 12px rgba(45, 228, 255, 0.48);
  font-variant-numeric: tabular-nums;
  line-height: 1;
}

.head-stats .stats-done .stats-num {
  color: #7dffc6;
  text-shadow: 0 0 14px rgba(73, 255, 166, 0.55);
}

.head-stats .stats-label {
  font-size: clamp(11px, 0.9em, 15px);
  color: rgba(211, 239, 255, 0.78);
  font-weight: 700;
  letter-spacing: 0.5px;
  line-height: 1.1;
  white-space: nowrap;
}

.head-stats .stats-sep {
  font-family: Consolas, Menlo, Monaco, 'Courier New', monospace;
  font-size: clamp(16px, 1.4em, 22px);
  color: rgba(45, 228, 255, 0.55);
  font-weight: 600;
  align-self: center;
  margin: 0 2px;
}

// 数据汇聚点 - 阶段进度
.progress-hub {
  position: absolute;
  top: 38px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 5;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 118px;
  height: 118px;
  pointer-events: none;
  isolation: isolate;
}

.hub-glow {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 170px;
  height: 170px;
  border-radius: 50%;
  background:
    radial-gradient(circle, rgba(45, 228, 255, 0.26) 0%, rgba(45, 228, 255, 0.12) 34%, transparent 66%);
  animation: hub-glow-pulse 3s ease-in-out infinite;
  z-index: 0;
}

.hub-rings {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 114px;
  height: 114px;
  z-index: 1;
}

.hub-ring {
  position: absolute;
  inset: 0;
  border-radius: 50%;
  box-sizing: border-box;
  animation: hub-ring-orbit 6s linear infinite;
}

.hub-ring-1 {
  border: 1px solid transparent;
  background:
    linear-gradient(rgba(3, 18, 42, 0.86), rgba(3, 18, 42, 0.86)) padding-box,
    conic-gradient(from -28deg, transparent 0 22deg, rgba(45, 228, 255, 0.8) 22deg 118deg, transparent 118deg 164deg, rgba(45, 228, 255, 0.42) 164deg 214deg, transparent 214deg 360deg) border-box;
  animation-delay: 0s;
  opacity: 0.92;
}

.hub-ring-2 {
  inset: 8px;
  border: 1px dashed rgba(115, 220, 255, 0.34);
  animation-delay: -2s;
  animation-direction: reverse;
}

.hub-ring-3 {
  inset: 17px;
  border: 1px solid rgba(45, 228, 255, 0.24);
  box-shadow: inset 0 0 18px rgba(45, 228, 255, 0.1);
  animation-delay: -4s;
}

.hub-core {
  position: relative;
  z-index: 2;
  display: flex;
  align-items: baseline;
  justify-content: center;
  width: 88px;
  height: 88px;
  padding-top: 22px;
  padding-bottom: 18px;
  background:
    radial-gradient(circle at 50% 42%, rgba(54, 245, 255, 0.18), transparent 44%),
    radial-gradient(circle at center, rgba(18, 92, 210, 0.3) 0%, transparent 72%),
    linear-gradient(145deg, rgba(8, 30, 60, 0.96), rgba(5, 15, 38, 0.86));
  border: 1px solid rgba(45, 228, 255, 0.52);
  border-radius: 50%;
  box-shadow:
    0 0 26px rgba(45, 228, 255, 0.28),
    inset 0 0 18px rgba(45, 228, 255, 0.15),
    inset 0 -12px 22px rgba(2, 8, 24, 0.5);
  overflow: hidden;
}

.hub-core::before {
  content: '';
  position: absolute;
  inset: 8px;
  border-radius: 50%;
  border: 1px solid rgba(211, 245, 255, 0.12);
  background:
    linear-gradient(90deg, transparent 48%, rgba(45, 228, 255, 0.12) 49% 51%, transparent 52%),
    linear-gradient(0deg, transparent 48%, rgba(45, 228, 255, 0.1) 49% 51%, transparent 52%);
  opacity: 0.76;
}

.hub-core::after {
  content: '';
  position: absolute;
  inset: 3px;
  border-radius: 50%;
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.18), transparent 34%);
  opacity: 0.48;
}

.hub-num {
  position: relative;
  z-index: 1;
  font-family: Consolas, Menlo, Monaco, 'Courier New', monospace;
  font-size: 36px;
  font-weight: 900;
  color: #2de4ff;
  line-height: 1;
  text-shadow:
    0 0 14px rgba(45, 228, 255, 0.82),
    0 0 28px rgba(45, 228, 255, 0.42);
  font-variant-numeric: tabular-nums;
  letter-spacing: 0;
  animation: hub-num-glow 2.8s ease-in-out infinite;
}

.hub-unit {
  position: relative;
  z-index: 1;
  font-family: Consolas, Menlo, Monaco, 'Courier New', monospace;
  align-self: baseline;
  margin-left: 1px;
  font-size: 18px;
  font-weight: 900;
  color: rgba(142, 237, 255, 0.84);
  line-height: 1;
}

// 数据汇流粒子 - 溪流汇入大海的效果
.flow-particles {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 1;
  pointer-events: none;
}

.flow-particle {
  position: absolute;
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(45, 228, 255, 0.92) 0%, rgba(45, 228, 255, 0.42) 50%, transparent 100%);
  box-shadow: 0 0 8px rgba(45, 228, 255, 0.65);
  animation: flow-to-hub 3.5s ease-in-out infinite;
}

.flow-particle-1 {
  top: 15%;
  left: 22%;
  animation-delay: 0s;
}

.flow-particle-2 {
  top: 25%;
  right: 18%;
  animation-delay: -0.6s;
}

.flow-particle-3 {
  top: 38%;
  left: 12%;
  animation-delay: -1.2s;
}

.flow-particle-4 {
  top: 42%;
  right: 14%;
  animation-delay: -1.8s;
}

.flow-particle-5 {
  bottom: 28%;
  left: 18%;
  animation-delay: -2.4s;
}

.flow-particle-6 {
  bottom: 22%;
  right: 22%;
  animation-delay: -3s;
}

@keyframes flow-to-hub {
  0% {
    opacity: 0.15;
    transform: scale(0.6);
  }
  25% {
    opacity: 0.85;
    transform: scale(1.1);
  }
  50% {
    opacity: 1;
    transform: scale(1.2) translate(20%, 15%);
  }
  75% {
    opacity: 0.9;
    transform: scale(1) translate(40%, 30%);
  }
  100% {
    opacity: 0;
    transform: scale(0.3) translate(50%, 50%);
  }
}

.progress-hub.is-done {
  .hub-glow {
    background:
      radial-gradient(circle, rgba(73, 255, 166, 0.28) 0%, rgba(73, 255, 166, 0.12) 34%, transparent 66%),
      radial-gradient(circle, rgba(45, 228, 255, 0.16) 0%, transparent 54%);
  }
  
  .hub-core {
    border-color: rgba(73, 255, 166, 0.55);
    box-shadow:
      0 0 26px rgba(73, 255, 166, 0.34),
      0 0 42px rgba(45, 228, 255, 0.18),
      inset 0 0 20px rgba(73, 255, 166, 0.18),
      inset 0 -12px 22px rgba(2, 8, 24, 0.5);
  }
  
  .hub-num {
    color: #7dffc6;
    text-shadow:
      0 0 20px rgba(73, 255, 166, 0.85),
      0 0 40px rgba(73, 255, 166, 0.55);
    animation: none;
  }
  
  .hub-ring-1 {
    background:
      linear-gradient(rgba(3, 18, 42, 0.86), rgba(3, 18, 42, 0.86)) padding-box,
      conic-gradient(from -28deg, transparent 0 18deg, rgba(73, 255, 166, 0.82) 18deg 130deg, transparent 130deg 176deg, rgba(45, 228, 255, 0.58) 176deg 232deg, transparent 232deg 360deg) border-box;
  }
  
  .hub-ring-2 {
    border-color: rgba(73, 255, 166, 0.35);
  }
}

@keyframes hub-glow-pulse {
  0%, 100% { opacity: 0.6; transform: translate(-50%, -50%) scale(1); }
  50% { opacity: 1; transform: translate(-50%, -50%) scale(1.15); }
}

@keyframes hub-ring-orbit {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@keyframes hub-num-glow {
  0%, 100% { opacity: 0.92; }
  50% { opacity: 1; }
}

.runway-progress {
  position: relative;
  display: inline-flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  border: 1px solid #ff9a2f;
  border-radius: 10px;
  background: linear-gradient(135deg, rgba(86, 43, 8, 0.95), rgba(34, 25, 20, 0.78));
  box-shadow:
    inset 0 0 20px rgba(255, 122, 0, 0.12);
  pointer-events: none;
  transition: border-color 0.3s ease, box-shadow 0.3s ease;
}

.runway-progress.is-done {
  border-color: rgba(73, 255, 166, 0.46);
  box-shadow:
    inset 0 0 18px rgba(73, 255, 166, 0.14),
    0 0 22px rgba(73, 255, 166, 0.18);
}

.progress-label {
  color: rgba(255, 224, 162, 0.86);
  font-size: clamp(14px, 1.2em, 20px);
  font-weight: 900;
  letter-spacing: 1.2px;
  white-space: nowrap;
}

.progress-value {
  display: inline-flex;
  align-items: baseline;
  gap: 2px;
  line-height: 1;
}

.progress-num {
  color: #ffe0a4;
  font-family: Consolas, Menlo, Monaco, 'Courier New', monospace;
  font-size: clamp(14px, 1.2em, 20px);
  font-weight: 900;
  letter-spacing: 1px;
  text-shadow: 0 0 14px rgba(255, 180, 74, 0.6);
  font-variant-numeric: tabular-nums;
  min-width: 2ch;
  text-align: right;
  animation: progress-pulse 2.4s ease-in-out infinite;
}

.progress-unit {
  color: rgba(255, 224, 162, 0.9);
  font-family: Consolas, Menlo, Monaco, 'Courier New', monospace;
  font-size: clamp(14px, 1.2em, 20px);
  font-weight: 900;
}

.runway-progress.is-done .progress-num {
  color: #7dffc6;
  text-shadow: 0 0 14px rgba(73, 255, 166, 0.62);
  animation: none;
}

.progress-bar {
  position: relative;
  display: block;
  width: 120px;
  height: 4px;
  border-radius: 2px;
  background: rgba(8, 23, 45, 0.72);
  overflow: hidden;
  box-shadow: inset 0 0 4px rgba(0, 0, 0, 0.6);
}

.progress-bar-fill {
  position: absolute;
  inset: 0 auto 0 0;
  border-radius: 2px;
  background: linear-gradient(90deg, #2de4ff, #ffbd62);
  box-shadow: 0 0 8px rgba(255, 180, 74, 0.7), 0 0 12px rgba(45, 228, 255, 0.5);
  transition: width 0.6s cubic-bezier(0.4, 0, 0.2, 1);
}

.runway-progress.is-done .progress-bar-fill {
  background: linear-gradient(90deg, #2de4ff, #73ffc0);
  box-shadow: 0 0 8px rgba(73, 255, 166, 0.7), 0 0 12px rgba(45, 228, 255, 0.5);
}

// 跑道底部 - 信息栏
.runway-foot {
  position: absolute;
  left: 24px;
  right: 24px;
  bottom: 16px;
  z-index: 3;
  display: flex;
  align-items: center;
  justify-content: space-between;
  pointer-events: none;
}

.foot-status {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  color: rgba(211, 239, 255, 0.92);
  font-size: clamp(13px, 1.1em, 18px);
  font-weight: 800;
  letter-spacing: 1px;
  border: 1px solid rgba(45, 228, 255, 0.32);
  border-radius: 12px;
  background: rgba(8, 24, 56, 0.58);
  box-shadow: 0 0 12px rgba(45, 228, 255, 0.12);
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #2de4ff;
  box-shadow: 0 0 8px rgba(45, 228, 255, 0.85);
  animation: status-pulse 1.8s ease-in-out infinite;
}

.status-text {
  max-width: min(320px, 38vw);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: 'Microsoft YaHei', 'PingFang SC', Arial, sans-serif;
}

.foot-phase-name {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  border: 1px solid rgba(45, 228, 255, 0.28);
  border-radius: 12px;
  background: rgba(8, 24, 56, 0.58);
  box-shadow:
    0 0 12px rgba(45, 228, 255, 0.1),
    inset 0 0 16px rgba(45, 228, 255, 0.04);
}

.phase-name-tag {
  font-size: clamp(13px, 1.1em, 18px);
  color: rgba(211, 239, 255, 0.92);
  font-weight: 800;
  letter-spacing: 1px;
  white-space: nowrap;
  font-family: 'Microsoft YaHei', 'PingFang SC', Arial, sans-serif;

  &::after {
    content: '';
    display: inline-block;
    width: 1px;
    height: 12px;
    margin-left: 8px;
    background: linear-gradient(180deg, transparent, rgba(45, 228, 255, 0.5), transparent);
    vertical-align: middle;
  }
}

.phase-name-text {
  font-size: clamp(13px, 1.1em, 18px);
  color: #ff9a2f;
  font-weight: 800;
  letter-spacing: 1px;
  font-family: 'PingFang SC', 'Microsoft YaHei', sans-serif;
  text-shadow: 0 0 10px rgba(255, 154, 47, 0.5);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 12em;
}

@keyframes status-pulse {
  0%, 100% { opacity: 0.85; transform: scale(1); }
  50% { opacity: 1; transform: scale(1.2); box-shadow: 0 0 12px rgba(45, 228, 255, 1); }
}

@keyframes progress-pulse {
  0%, 100% { text-shadow: 0 0 10px rgba(255, 180, 74, 0.45); }
  50% { text-shadow: 0 0 18px rgba(255, 180, 74, 0.78), 0 0 24px rgba(255, 180, 74, 0.32); }
}

.runway-svg {
  position: absolute;
  inset: 128px 6px 70px;
  width: calc(100% - 18px);
  height: calc(100% - 198px);
  overflow: visible;
  z-index: 1;
}

.lane-aura {
  opacity: 0.72;
}

.lane-dash {
  opacity: 0.58;
  stroke: rgba(211, 245, 255, 0.54);
}

.runway-svg-turn-pip {
  fill: rgba(211, 245, 255, 0.72);
  opacity: 0.18;
  transform-box: fill-box;
  transform-origin: center;
  animation: runway-svg-turn-pip 2.4s ease-in-out infinite;
}

.runway-svg-turn-pip-left {
  animation-delay: -1.2s;
}

.baton {
  fill: #6b3514;
  stroke: #ffdd9a;
  stroke-width: 3;
  animation: baton-move 1.05s ease-in-out infinite;
}

.baton-arrow {
  animation: baton-move 1.05s ease-in-out infinite;
}

.node-pulse-ring {
  animation: node-ring 1.28s ease-out infinite;
  transform-origin: center;
  transform-box: fill-box;
}

.node-core {
  stroke-width: 3;
}

.node-core-completed {
  fill: #073c31;
  stroke: #55ffb0;
}

.node-core-pending {
  fill: #102845;
  stroke: #567fa7;
}

.node-core-issue {
  fill: #4a1721;
  stroke: #ff666a;
}

.node-label {
  font-size: 31px;
  font-weight: 900;
  letter-spacing: 0;
  paint-order: stroke fill;
  stroke: rgba(2, 10, 24, 0.62);
  stroke-width: 3px;
}

.finish-node {
  opacity: 0.82;
  .finish-text {
    fill: #dcecff;
    font-size: 31px;
    font-weight: 900;
    letter-spacing: 0;
    paint-order: stroke fill;
    stroke: rgba(2, 10, 24, 0.62);
    stroke-width: 3px;
  }
}

.finish-beacon-ring {
  transform-origin: center;
  transform-box: fill-box;
  animation: finish-beacon-ring 2.2s ease-out infinite;
}

.finish-beacon-ring-b {
  animation-delay: -1.1s;
}

.finish-sparks {
  stroke: rgba(255, 224, 162, 0.82);
  stroke-width: 2.5;
  stroke-linecap: round;
  animation: finish-sparks-flash 1.8s ease-in-out infinite;
  transform-origin: center;
  transform-box: fill-box;
}

.finish-tether {
  animation: finish-tether-pulse 1.6s ease-in-out infinite;
}

.finish-disc {
  stroke-width: 2.4px;
}

.finish-crown-shadow {
  fill: rgba(25, 12, 2, 0.62);
  transform: translate(2px, 4px);
}

.finish-crown {
  fill: url(#crown-gold);
  stroke: #fff0a8;
  stroke-width: 2.1;
  stroke-linejoin: round;
  paint-order: stroke fill;
}

.finish-crown-ridge {
  fill: none;
  stroke: url(#crown-ridge);
  stroke-width: 1.35;
  stroke-linecap: round;
  stroke-linejoin: round;
  opacity: 0.8;
}

.finish-crown-gem {
  fill: url(#crown-gem);
  stroke: rgba(255, 255, 255, 0.82);
  stroke-width: 0.8;
}

.finish-crown-gem-main {
  fill: #e53e2f;
  stroke: #ffd7b0;
}

.finish-orbit {
  transform-origin: center;
  transform-box: fill-box;
  animation: finish-orbit-spin 5s linear infinite;
}

.finish-orbit-outer {
  animation: finish-orbit-pulse 2.2s ease-in-out infinite;
}

.finish-node-done {
  opacity: 1;
  animation: finish-beacon 1.5s ease-in-out infinite;
  .finish-text {
    fill: #ffe1a5;
  }
}

@keyframes panel-sweep {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@keyframes runway-svg-turn-pip {
  0%, 100% {
    opacity: 0.22;
    transform: scaleX(0.82);
  }
  45% {
    opacity: 0.72;
    transform: scaleX(1.28);
  }
}

@keyframes baton-move {
  50% { transform: translateX(8px) scale(1.04); }
}

@keyframes node-ring {
  from { opacity: 0.85; transform: scale(0.65); }
  to { opacity: 0; transform: scale(1.78); }
}

@keyframes finish-beacon {
  50% {
    opacity: 0.88;
  }
}

@keyframes finish-beacon-ring {
  from {
    opacity: 0.82;
    transform: scale(0.72);
  }
  to {
    opacity: 0;
    transform: scale(1.42);
  }
}

@keyframes finish-sparks-flash {
  0%, 100% {
    opacity: 0.34;
    transform: scale(0.9);
  }
  48% {
    opacity: 0.95;
    transform: scale(1.04);
  }
}

@keyframes finish-tether-pulse {
  50% { opacity: 0.52; stroke-width: 3.4; }
}

@keyframes finish-orbit-spin {
  to { transform: rotate(360deg); }
}

@keyframes finish-orbit-pulse {
  50% { opacity: 0.38; transform: scale(1.12); }
}

@media (prefers-reduced-motion: reduce) {
  .relay-runway::before,
  .runway-svg-turn-pip,
  .baton,
  .baton-arrow,
  .node-pulse-ring,
  .finish-node-done,
  .finish-beacon-ring,
  .finish-sparks,
  .finish-tether,
  .finish-orbit,
  .finish-orbit-outer,
  .progress-num,
  .foot-tag-dot {
    animation: none;
  }
}
</style>
