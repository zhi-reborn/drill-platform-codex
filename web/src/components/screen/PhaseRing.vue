<template>
  <div
    ref="phaseRingRef"
    class="phase-ring"
    :class="[
      phaseDirectionDown ? 'phase-dir-down' : 'phase-dir-up',
      `phase-stage-${safeCurrentIndex}`,
      { 'phase-ring-compact': isCompact },
    ]"
    :style="{
      width: containerSize.width + 'px',
      height: containerSize.height + 'px',
      '--connector-y': connectorY,
    }"
  >
    <div class="relay-runway">
      <div class="runway-head">
        <div class="phase-tag">
          <div class="phase-tag-inner">
            <div class="phase-tag-name">{{ currentPhaseLabel }}</div>
          </div>
          <div class="phase-tag-scan"></div>
        </div>
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
          <radialGradient id="target-core" cx="50%" cy="42%" r="62%">
            <stop offset="0" stop-color="#fff5d6" />
            <stop offset=".45" stop-color="#ffd36f" />
            <stop offset="1" stop-color="#e8891c" />
          </radialGradient>
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
          <linearGradient id="baton-fill" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0" stop-color="#6b2a08" />
            <stop offset=".18" stop-color="#b85316" />
            <stop offset=".5" stop-color="#ff8426" />
            <stop offset=".82" stop-color="#b85316" />
            <stop offset="1" stop-color="#6b2a08" />
          </linearGradient>
          <linearGradient id="baton-core" x1="0" y1="0" x2="1" y2="0">
            <stop offset="0" stop-color="rgba(255, 224, 162, 0)" />
            <stop offset=".5" stop-color="rgba(255, 245, 214, 0.85)" />
            <stop offset="1" stop-color="rgba(255, 224, 162, 0)" />
          </linearGradient>
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
              v-else-if="node.visualStatus === 'completed' && !node.isFinish"
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
            <!-- 终点靶心：最后一个环节的节点圆环由靶心取代；先于接力棒绘制，抵达终点时箭头压在靶心上呈"命中目标" -->
            <g
              v-if="node.isFinish"
              class="finish-target"
              :class="{ 'finish-target-done': node.visualStatus === 'completed' }"
            >
              <circle r="26" class="finish-target-board" />
              <circle r="21.5" class="finish-target-scan" />
              <circle r="16.5" class="finish-target-ring finish-target-ring-outer" />
              <circle r="10.5" class="finish-target-ring finish-target-ring-mid" />
              <circle r="4.2" class="finish-target-bullseye" />
            </g>
            <circle
              v-else
              r="19"
              :class="['node-core', `node-core-${node.visualStatus}`]"
            />
            <g
              v-if="node.visualStatus === 'running'"
              class="baton-group"
              :class="`flow-${node.flowDir}`"
            >
              <!-- 能量尾迹：位于 baton 后方（与流向相反），逐渐衰减，强化方向感 -->
              <circle
                v-for="(t, i) in node.trail"
                :key="'trail-' + i"
                class="baton-trail-dot"
                :class="`baton-trail-${i + 1}`"
                :cx="t.cx"
                cy="0"
                :r="t.r"
              />
              <rect x="-40" y="-18" width="80" height="36" rx="18" class="baton" />
              <rect x="-32" y="-9" width="64" height="18" rx="9" class="baton-core" />
              <!-- 箭头随跑道流向：偶数行向右，奇数行向左 -->
              <path
                :d="node.flowDir === 'left' ? 'M20 0 H-16 M-4 -10 L-18 0 L-4 10' : 'M-20 0 H16 M4 -10 L18 0 L4 10'"
                class="baton-arrow"
              />
            </g>
            <text
              class="node-label"
              :class="{ 'node-label-multiline': node.lines.length > 1 }"
              :style="{ fontSize: node.labelFontSize + 'px' }"
              :y="node.labelY"
              :fill="node.labelColor"
            >
              <tspan
                v-for="(line, li) in node.lines"
                :key="li"
                x="0"
                :dy="li === 0 ? 0 : node.labelLineHeight"
              >{{ line }}</tspan>
            </text>
          </g>
        </g>
      </svg>
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
// 节点距跑道端点的内边距：缩小后节点更舒展，相邻标签间距更大
const LANE_NODE_PADDING = 48
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
  size?: number
  fullscreen?: boolean
}>()

const size = computed(() => props.size ?? 520)
const isCompact = computed(() => size.value < 460)
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
  const lastIndex = currentNodes.value.length - 1
  return currentNodes.value.map((name, index) => {
    const status = currentStatuses.value[index]
    const visualStatus = visualStatusOf(status, index)
    const p = trackPoints.value[index]
    const { row } = locateLaneNode(index)
    // 跑道流向：偶数行（0,2…）由左向右，奇数行（1,3…）由右向左
    const flowDir = row % 2 === 1 ? 'left' : 'right'
    const trailSign = flowDir === 'left' ? 1 : -1
    const lines = splitName(name)
    const isMulti = lines.length > 1
    // 多行时字号略缩，但不低于 stage-name 的最小字号（14px CSS → SVG 单位约 18px）
    // 字号收一点 + 节点内边距缩小，让相邻标签留出更舒展的呼吸间距
    const labelFontSize = isMulti
      ? (isCompact.value ? 22 : 25)
      : (isCompact.value ? 24 : 28)
    const labelLineHeight = isCompact.value ? 24 : 29
    // 末行基线固定在 -32（与原单行一致），多出的行向上延伸
    const labelY = -32 - (lines.length - 1) * labelLineHeight
    return {
      ...p,
      index,
      lines,
      visualStatus,
      labelAbove: shouldLabelAbove(index, p),
      labelColor: '#f2f8ff',
      labelFontSize,
      labelY,
      labelLineHeight,
      flowDir,
      // 最后一个环节为跑道终点，节点圆环由靶心取代
      isFinish: index === lastIndex,
      // 尾迹位于 baton 后方（与流向相反），半径与透明度逐级衰减
      trail: [
        { cx: trailSign * 50, r: 3.4 },
        { cx: trailSign * 60, r: 2.5 },
        { cx: trailSign * 68, r: 1.7 },
      ],
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

// 将环节名处理为最多 2 行的展示文本
// 1) 超过 12 字时截断为前 11 字 + "…"，保持 2 行 5+7 布局与宽度安全
// 2) 在中点附近寻找功能词（的/与/和/及）作为断点，断在词后让语义自然
// 如 "监控告警的实时触发与确认" → ["监控告警的","实时触发与确认"]
// 如 "监控告警的实时触发与确认及响应处置" → ["监控告警的","实时触发与确…"]
function splitName(name: string): string[] {
  // 2 行最多容纳 12 字（5+7），超出部分以省略号收尾
  const MAX_LEN = 12
  const text = name.length > MAX_LEN
    ? name.slice(0, MAX_LEN - 1) + '…'
    : name
  const len = text.length
  if (len <= 4) return [text]
  if (len <= 8) {
    // 首行 4 字，剩余移至第二行（如 "系统基线检查" → ["系统基线","检查"]）
    return [text.slice(0, 4), text.slice(4)]
  }
  // 超长名称：拆成 2 行
  // 在中点附近（±2）优先寻找功能词断点，让首行落在修饰语后、动作前
  const breakers = '的与和及'
  const mid = Math.floor(len / 2)
  const candidates = [mid, mid - 1, mid + 1, mid - 2, mid + 2]
    .filter(p => p >= 3 && p <= len - 3)
  for (const pos of candidates) {
    if (breakers.includes(text[pos - 1])) {
      return [text.slice(0, pos), text.slice(pos)]
    }
  }
  // 无功能词：从中点等分
  return [text.slice(0, mid), text.slice(mid)]
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
  min-height: 56px;
  pointer-events: none;
}

// 左上角当前阶段标签：切角 HUD 面板 + 扫光
.phase-tag {
  --chamfer: 12px;
  position: relative;
  padding: 1px; /* 渐变描边厚度 */
  background: linear-gradient(135deg, rgba(45, 228, 255, 0.65), rgba(45, 228, 255, 0.12) 45%, rgba(45, 228, 255, 0.45));
  clip-path: polygon(
    var(--chamfer) 0,
    100% 0,
    100% calc(100% - var(--chamfer)),
    calc(100% - var(--chamfer)) 100%,
    0 100%,
    0 var(--chamfer)
  );
  box-shadow: 0 0 18px rgba(45, 228, 255, 0.16);
  animation: phase-tag-in 0.7s cubic-bezier(0.22, 1, 0.36, 1) both;
}

.phase-tag-inner {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 8px 22px 9px 16px;
  background:
    linear-gradient(90deg, rgba(45, 228, 255, 0.14), rgba(45, 228, 255, 0) 55%),
    linear-gradient(160deg, rgba(10, 28, 62, 0.92), rgba(8, 24, 56, 0.78));
  clip-path: polygon(
    var(--chamfer) 0,
    100% 0,
    100% calc(100% - var(--chamfer)),
    calc(100% - var(--chamfer)) 100%,
    0 100%,
    0 var(--chamfer)
  );
  overflow: hidden;
}

.phase-tag-name {
  font-size: clamp(17px, 1.55em, 27px);
  font-weight: 900;
  letter-spacing: 4px;
  line-height: 1.15;
  font-family: 'Microsoft YaHei', 'PingFang SC', sans-serif;
  background: linear-gradient(180deg, #ffffff 35%, #a5ecff 90%);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
  filter: drop-shadow(0 0 9px rgba(45, 228, 255, 0.4));
}

.phase-tag-scan {
  position: absolute;
  top: 0;
  bottom: 0;
  left: -40%;
  width: 34%;
  background: linear-gradient(100deg, transparent, rgba(151, 240, 255, 0.16), rgba(151, 240, 255, 0.32), rgba(151, 240, 255, 0.16), transparent);
  transform: skewX(-14deg);
  animation: phase-tag-scan 4.2s ease-in-out infinite;
  pointer-events: none;
}

@keyframes phase-tag-in {
  from { opacity: 0; transform: translateY(-10px); filter: blur(3px); }
  to { opacity: 1; transform: translateY(0); filter: blur(0); }
}

@keyframes phase-tag-scan {
  0%, 55% { left: -40%; opacity: 0; }
  60% { opacity: 1; }
  90% { left: 110%; opacity: 0.9; }
  100% { left: 110%; opacity: 0; }
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
  padding-top: 24px;
  padding-bottom: 16px;
  background:
    radial-gradient(circle at 50% 42%, rgba(54, 245, 255, 0.2), transparent 46%),
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
  font-size: 38px;
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
  margin-left: 2px;
  font-size: 19px;
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

/* 终点靶心：靶盘（覆盖跑道端帽）+ 旋转虚线扫描环 + 三层同心环 + 呼吸靶心 */
.finish-target-board {
  fill: rgba(7, 19, 40, 0.95);
  stroke: rgba(255, 214, 130, 0.4);
  stroke-width: 1.5;
  filter: drop-shadow(0 0 8px rgba(255, 180, 74, 0.28));
}

.finish-target-scan {
  fill: none;
  stroke: rgba(255, 214, 130, 0.55);
  stroke-width: 1.6;
  stroke-dasharray: 10 7;
  transform-origin: center;
  transform-box: fill-box;
  animation: finish-scan-spin 6s linear infinite;
}

.finish-target-ring-outer {
  fill: rgba(255, 180, 74, 0.07);
  stroke: #ffd36f;
  stroke-width: 2.4;
  filter: drop-shadow(0 0 6px rgba(255, 180, 74, 0.5));
}

.finish-target-ring-mid {
  fill: none;
  stroke: rgba(45, 228, 255, 0.72);
  stroke-width: 1.6;
}

.finish-target-bullseye {
  fill: url(#target-core);
  transform-origin: center;
  transform-box: fill-box;
  filter: drop-shadow(0 0 5px rgba(255, 189, 98, 0.85));
  animation: target-bullseye-pulse 1.8s ease-in-out infinite;
}

.finish-target-done {
  .finish-target-board {
    fill: rgba(6, 32, 26, 0.95);
    stroke: rgba(115, 255, 192, 0.45);
    filter: drop-shadow(0 0 8px rgba(73, 255, 166, 0.32));
  }

  .finish-target-scan {
    stroke: rgba(115, 255, 192, 0.6);
  }

  .finish-target-ring-outer {
    stroke: #73ffc0;
    filter: drop-shadow(0 0 6px rgba(73, 255, 166, 0.5));
  }

  .finish-target-ring-mid {
    stroke: rgba(125, 255, 198, 0.65);
  }

  .finish-target-bullseye {
    fill: #7dffc6;
    filter: drop-shadow(0 0 6px rgba(73, 255, 166, 0.85));
  }
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

.phase-ring-compact {
  .runway-head {
    top: 10px;
    left: 14px;
    right: 28px;
    min-height: 46px;
  }

  .phase-tag {
    --chamfer: 9px;

    .phase-tag-inner {
      padding: 6px 18px 7px 12px;
    }

    .phase-tag-name {
      font-size: clamp(14px, 1.3em, 21px);
      letter-spacing: 3px;
    }
  }

  .head-stats {
    min-height: 44px;
    gap: 7px;
    padding: 4px 10px;
    border-radius: 12px;
  }

  .progress-hub {
    top: 28px;
    width: 94px;
    height: 94px;
  }

  .hub-glow {
    width: 132px;
    height: 132px;
  }

  .hub-rings {
    width: 92px;
    height: 92px;
  }

  .hub-core {
    width: 72px;
    height: 72px;
    padding-top: 19px;
    padding-bottom: 13px;
  }

  .hub-num {
    font-size: 30px;
  }

  .hub-unit {
    font-size: 15px;
  }

  .runway-svg {
    inset: 104px 6px 58px;
    height: calc(100% - 162px);
  }

  .node-label {
    font-size: 26px;
    stroke-width: 2.4px;
  }
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

.baton-trail-dot {
  fill: #ffbd62;
  filter: drop-shadow(0 0 4px rgba(255, 189, 98, 0.65));
  transform-box: fill-box;
  transform-origin: center;
  animation: trail-breathe 1.4s ease-in-out infinite;
}

.baton-trail-1 { opacity: 0.74; animation-delay: 0s; }
.baton-trail-2 { opacity: 0.44; animation-delay: -0.45s; }
.baton-trail-3 { opacity: 0.2; animation-delay: -0.9s; }

.baton {
  fill: url(#baton-fill);
  stroke: #ffdd9a;
  stroke-width: 2.6;
  filter: drop-shadow(0 0 7px rgba(255, 132, 38, 0.72));
}

.baton-core {
  fill: url(#baton-core);
}

.baton-arrow {
  fill: none;
  stroke: #fff5d6;
  stroke-width: 3;
  stroke-linecap: round;
  stroke-linejoin: round;
  filter: drop-shadow(0 0 4px rgba(255, 245, 214, 0.8));
}

// 跑道流向驱动 baton 摆动方向：向右行进时右摆，向左行进时左摆
.baton-group.flow-right .baton,
.baton-group.flow-right .baton-core,
.baton-group.flow-right .baton-arrow {
  animation: baton-run-right 1.05s ease-in-out infinite;
}

.baton-group.flow-left .baton,
.baton-group.flow-left .baton-core,
.baton-group.flow-left .baton-arrow {
  animation: baton-run-left 1.05s ease-in-out infinite;
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

// 多行标签字号略小，相应收细描边以保持清晰
.node-label-multiline {
  stroke-width: 2.5px;
}

@keyframes finish-scan-spin {
  to { transform: rotate(360deg); }
}

@keyframes target-bullseye-pulse {
  0%, 100% { transform: scale(0.86); }
  50% { transform: scale(1.12); }
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

@keyframes baton-run-right {
  50% { transform: translateX(6px) scale(1.04); }
}

@keyframes baton-run-left {
  50% { transform: translateX(-6px) scale(1.04); }
}

@keyframes trail-breathe {
  0%, 100% { transform: scale(0.78); }
  50% { transform: scale(1.22); }
}

@keyframes node-ring {
  from { opacity: 0.85; transform: scale(0.65); }
  to { opacity: 0; transform: scale(1.78); }
}

@media (prefers-reduced-motion: reduce) {
  .relay-runway::before,
  .runway-svg-turn-pip,
  .baton,
  .baton-arrow,
  .baton-core,
  .baton-trail-dot,
  .node-pulse-ring,
  .finish-target-scan,
  .finish-target-bullseye,
  .progress-num,
  .foot-tag-dot {
    animation: none;
  }
}
</style>
