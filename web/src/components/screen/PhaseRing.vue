<template>
  <div class="phase-ring" :style="{ width: containerSize.width + 'px', height: containerSize.height + 'px' }">
    <div class="ring-inner" :style="{ width: (size + PAD_X * 2) + 'px', height: (size + PAD_Y_TOP + PAD_Y_BOTTOM) + 'px' }">
    <svg :viewBox="`0 0 ${size + PAD_X * 2} ${size + PAD_Y_TOP + PAD_Y_BOTTOM}`" :width="size + PAD_X * 2" :height="size + PAD_Y_TOP + PAD_Y_BOTTOM" class="ring-svg">
      <defs>
        <linearGradient id="grad-active" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#fff4cf" />
          <stop offset="44%" stop-color="#ffb44a" />
          <stop offset="100%" stop-color="#ff6900" />
        </linearGradient>
        <linearGradient id="grad-done" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#0066cc" />
          <stop offset="100%" stop-color="#00a0c8" />
        </linearGradient>
        <linearGradient id="grad-pending" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#1a3a6a" />
          <stop offset="100%" stop-color="#0a1f3a" />
        </linearGradient>
        <radialGradient id="center-glow" cx="50%" cy="50%" r="50%">
          <stop offset="0%" stop-color="rgba(0, 212, 255, 0.25)" />
          <stop offset="60%" stop-color="rgba(0, 80, 160, 0.12)" />
          <stop offset="100%" stop-color="rgba(0, 30, 80, 0)" />
        </radialGradient>
        <filter id="glow" x="-50%" y="-50%" width="200%" height="200%">
          <feGaussianBlur stdDeviation="3" result="b" />
          <feMerge>
            <feMergeNode in="b" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
        <filter id="glow-strong" x="-100%" y="-100%" width="300%" height="300%">
          <feGaussianBlur stdDeviation="6" result="b" />
          <feMerge>
            <feMergeNode in="b" />
            <feMergeNode in="b" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
        <!-- 雷达扫描渐变（35° 扇形，从前缘亮到后缘透明） -->
        <linearGradient id="radar-sweep-grad" gradientUnits="userSpaceOnUse"
          :x1="cx + outerR * 1.05 * Math.cos(Math.PI / 2 - Math.PI / 7)"
          :y1="cy - outerR * 1.05 * Math.sin(Math.PI / 2 - Math.PI / 7)"
          :x2="cx + outerR * 1.05 * Math.cos(Math.PI / 2)"
          :y2="cy - outerR * 1.05 * Math.sin(Math.PI / 2)">
          <stop offset="0%" stop-color="#00d4ff" stop-opacity="0.38" />
          <stop offset="40%" stop-color="#00d4ff" stop-opacity="0.15" />
          <stop offset="100%" stop-color="#00d4ff" stop-opacity="0" />
        </linearGradient>
      </defs>

      <!-- 中心光晕 -->
      <circle :cx="cx" :cy="cy" :r="innerR" fill="url(#center-glow)" />

      <!-- 同心装饰环 -->
      <g class="concentric-rings" opacity="0.18">
        <circle :cx="cx" :cy="cy" :r="innerR * 1.4" fill="none" stroke="#00d4ff" stroke-width="0.5" stroke-dasharray="3 6" />
        <circle :cx="cx" :cy="cy" :r="(innerR + segR) / 2" fill="none" stroke="#00d4ff" stroke-width="0.5" />
        <circle :cx="cx" :cy="cy" :r="segR" fill="none" stroke="#00d4ff" stroke-width="0.5" stroke-dasharray="2 4" />
        <circle :cx="cx" :cy="cy" :r="outerR * 0.92" fill="none" stroke="#00d4ff" stroke-width="0.3" stroke-dasharray="1 5" />
      </g>

      <!-- 径向网格线（12 条，每 30°） -->
      <g class="radial-grid" opacity="0.10">
        <line v-for="i in 12" :key="'rl' + i"
          :x1="cx + innerR * Math.cos(((i - 1) / 12) * Math.PI * 2)"
          :y1="cy + innerR * Math.sin(((i - 1) / 12) * Math.PI * 2)"
          :x2="cx + outerR * 0.96 * Math.cos(((i - 1) / 12) * Math.PI * 2)"
          :y2="cy + outerR * 0.96 * Math.sin(((i - 1) / 12) * Math.PI * 2)"
          stroke="#00d4ff" stroke-width="0.5" />
      </g>

      <!-- 外圈装饰刻度（240 条短线） -->
      <g class="ticks">
        <line
          v-for="i in 120"
          :key="'tk' + i"
          :x1="cx + (outerR + 4) * Math.cos(((i - 1) / 120) * Math.PI * 2)"
          :y1="cy + (outerR + 4) * Math.sin(((i - 1) / 120) * Math.PI * 2)"
          :x2="cx + (outerR + (i % 5 === 0 ? 12 : 8)) * Math.cos(((i - 1) / 120) * Math.PI * 2)"
          :y2="cy + (outerR + (i % 5 === 0 ? 12 : 8)) * Math.sin(((i - 1) / 120) * Math.PI * 2)"
          :stroke="i % 15 === 0 ? 'rgba(0, 212, 255, 0.5)' : 'rgba(0, 212, 255, 0.18)'"
          :stroke-width="i % 15 === 0 ? 1.5 : 0.8"
        />
      </g>

      <!-- 阶段分段环（4 段弧） -->
      <g class="phase-segments" filter="url(#glow)">
        <path
          v-for="(seg, idx) in segmentPaths"
          :key="'seg' + idx"
          :d="seg.d"
          :stroke="seg.color"
          :stroke-width="12"
          fill="none"
          :stroke-linecap="'butt'"
          :opacity="seg.opacity"
        />
      </g>

      <!-- 环节节点进度弧（外圈小弧形，长度反映进度） -->
      <g class="node-progress-arcs">
        <path
          v-for="(arc, idx) in nodeProgressArcs"
          :key="'npa' + idx"
          :d="arc.d"
          :stroke="arc.color"
          :stroke-width="5"
          fill="none"
          stroke-linecap="butt"
          :opacity="arc.opacity"
          :class="{ 'arc-running': arc.isRunning }"
        />
      </g>

      <!-- 活跃段流光动画层 -->
      <g v-if="activeSegmentPath" class="flow-light-group">
        <path
          :d="activeSegmentPath"
          stroke="#ffb44a"
          stroke-width="14"
          fill="none"
          stroke-linecap="butt"
          class="flow-light-streak"
        />
      </g>

      <!-- 活跃段粒子（沿弧线运动） -->
      <g v-if="activeSegmentPath" class="flow-particles">
        <circle r="3.5" fill="#fff4cf" opacity="0.9" class="flow-particle flow-particle-1">
          <animateMotion :dur="activeSegDur" repeatCount="indefinite" :path="activeSegmentPath" />
        </circle>
        <circle r="2.5" fill="#ff9a2f" opacity="0.7" class="flow-particle flow-particle-2">
          <animateMotion :dur="activeSegDur" begin="0.4s" repeatCount="indefinite" :path="activeSegmentPath" />
        </circle>
        <circle r="2" fill="#ffe0a0" opacity="0.6" class="flow-particle flow-particle-3">
          <animateMotion :dur="activeSegDur" begin="0.8s" repeatCount="indefinite" :path="activeSegmentPath" />
        </circle>
      </g>

      <!-- 雷达扫描光束 -->
      <g class="radar-sweep">
        <path
          :d="`M ${cx} ${cy} L ${cx} ${cy - outerR * 1.02} A ${outerR * 1.02} ${outerR * 1.02} 0 0 1 ${cx + outerR * 1.02 * Math.sin(Math.PI / 7)} ${cy - outerR * 1.02 * Math.cos(Math.PI / 7)} Z`"
          fill="url(#radar-sweep-grad)"
        />
        <line
          :x1="cx" :y1="cy"
          :x2="cx" :y2="cy - outerR * 1.02"
          stroke="#00d4ff" stroke-width="1.8" opacity="0.7"
        />
      </g>

      <!-- 当前阶段进度弧（在分段环外侧） -->
      <g v-if="progressPath" class="progress-arc" filter="url(#glow)">
        <path
          :d="progressPath"
          stroke="url(#grad-active)"
          stroke-width="4"
          fill="none"
          stroke-linecap="round"
        />
      </g>

    </svg>

    <!-- 4 个阶段节点（外侧 4 角，圆形，HTML 定位） -->
    <div
      v-for="(p, idx) in phasePoints"
      :key="'lbl' + idx"
      class="phase-node-label"
      :class="['phase-node-' + idx, idx === currentIndex ? 'phase-node-active' : '', idx < currentIndex ? 'phase-node-done' : '']"
      :style="{
        left: (p.x / (size + PAD_X * 2) * 100) + '%',
        top: (p.y / (size + PAD_Y_TOP + PAD_Y_BOTTOM) * 100) + '%',
      }"
    >
      <span class="phase-node-text">阶段{{ chineseNum(idx + 1) }}</span>
    </div>

    <!-- 环节标签（环形外侧，围绕各阶段弧段分布） -->
    <svg :viewBox="`0 0 ${size + PAD_X * 2} ${size + PAD_Y_TOP + PAD_Y_BOTTOM}`" :width="size + PAD_X * 2" :height="size + PAD_Y_TOP + PAD_Y_BOTTOM" class="ring-svg ring-svg-overlay">
      <line
        v-for="(lp, idx) in ringLabelLines"
        :key="'rll' + idx"
        :x1="lp.x1" :y1="lp.y1"
        :x2="lp.x2" :y2="lp.y2"
        :stroke="ringLabels[idx].isRunning ? 'rgba(255, 180, 74, 0.58)' : 'rgba(0, 212, 255, 0.2)'"
        :stroke-width="ringLabels[idx].isRunning ? 1.5 : 1"
        :class="{ 'label-line-running': ringLabels[idx].isRunning }"
      />
      <!-- 环节节点圆点 -->
      <circle
        v-for="(lp, idx) in ringLabelLines"
        :key="'rlc' + idx"
        :cx="lp.x2" :cy="lp.y2"
        :r="ringLabels[idx].isRunning ? 4 : 2.5"
        :fill="ringLabels[idx].isRunning ? '#ffb44a' : '#00d4ff'"
        :opacity="ringLabels[idx].isRunning ? 0.95 : 0.5"
        :class="{ 'label-node-running': ringLabels[idx].isRunning }"
        :filter="ringLabels[idx].isRunning ? 'url(#glow-strong)' : ''"
      />
      <!-- 运行中环节节点波纹效果 -->
      <g v-for="(rp, idx) in runningRipples" :key="'rp' + idx" :transform="`translate(${rp.cx}, ${rp.cy})`">
        <circle r="4" fill="none" stroke="#ffb44a" stroke-width="2" class="ripple ripple-1" />
        <circle r="4" fill="none" stroke="rgba(255, 180, 74, 0.6)" stroke-width="1.5" class="ripple ripple-2" />
        <circle r="4" fill="none" stroke="rgba(255, 180, 74, 0.3)" stroke-width="1" class="ripple ripple-3" />
      </g>
      <!-- 运行中节点光晕环 -->
      <circle
        v-for="(rp, idx) in runningRipples"
        :key="'rh' + idx"
        :cx="rp.cx" :cy="rp.cy"
        r="8"
        fill="none"
        stroke="rgba(255, 180, 74, 0.25)"
        stroke-width="1"
        class="running-halo"
      />
    </svg>

      <div
        v-for="(lp, idx) in ringLabelPositions"
        :key="'rl' + idx"
        class="ring-outer-label"
        :class="[
          { 'ring-outer-label-active': ringLabels[idx].active },
          { 'ring-outer-label-running': ringLabels[idx].isRunning },
          { 'ring-outer-label-flipped': lp.flipped },
          'ring-outer-label-phase-' + ringLabels[idx].phaseIdx,
        ]"
        :style="{
          left: lp.leftPct + '%',
          top: lp.topPct + '%',
          '--label-rotate': lp.rotate + 'deg',
        }"
      >
        <span class="ring-outer-label-content">
          <span v-if="ringLabels[idx].isRunning" class="running-indicator">
            <span class="running-dot"></span>
            <span class="running-dot-ripple"></span>
            <span class="running-dot-ripple delay"></span>
          </span>
          <span class="ring-outer-label-text">{{ ringLabels[idx].text }}</span>
        </span>
      </div>

    <!-- 中心数字 -->
    <div class="center-content">
      <div class="center-caption">演练总进度</div>
      <div class="center-pct">{{ progress }}<span class="pct-unit">%</span></div>
      <div class="center-value">
        <span class="num-num">{{ centerNumerator }}</span>
        <span class="num-sep">/</span>
        <span class="num-den">{{ centerDenominator }}</span>
      </div>
      <div class="center-hint">{{ centerHint }}</div>
    </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  phases: string[]
  phaseNames: string[][]
  phaseNodeStatuses: { status: string; progress: number }[][]
  currentIndex: number
  progress: number
  centerNumerator: number
  centerDenominator: number
  centerHint: string
  size?: number
}>()

const size = computed(() => props.size ?? 520)
const PAD_X = 90  // 左右预留空间（加大以容纳更多标签）
const PAD_Y_TOP = 85  // 顶部预留（加大以容纳倾斜标签）
const PAD_Y_BOTTOM = 100  // 底部预留（标签文字需要更多空间）
const containerSize = computed(() => ({
  width: size.value + PAD_X * 2,
  height: size.value + PAD_Y_TOP + PAD_Y_BOTTOM,
}))
const cx = computed(() => size.value / 2 + PAD_X)
const cy = computed(() => size.value / 2 + PAD_Y_TOP)

// 半径定义
const innerR = computed(() => size.value * 0.18)
const segR = computed(() => size.value * 0.36)   // 分段环
const outerR = computed(() => size.value * 0.46) // 刻度

// === 按阶段环节节点数量比例分配弧度 ===
const segCount = computed(() => props.phases.length || 4)

const segAngles = computed(() => {
  const count = segCount.value
  const MIN_SEG_RAD = 0.5  // 每段最小弧度（约28.6°），确保阶段至少可见
  const TOTAL_GAP = 0.06 * count
  const AVAILABLE_RAD = Math.PI * 2 - TOTAL_GAP

  const counts = props.phases.map((_, i) => Math.max(props.phaseNames?.[i]?.length || 1, 1))
  const totalWeight = counts.reduce((a, b) => a + b, 0)
  // 先按比例分配
  let angles = counts.map(c => (c / totalWeight) * AVAILABLE_RAD)
  // 确保每段至少 MIN_SEG_RAD：从大段中扣除差额
  const deficit = angles.filter(a => a < MIN_SEG_RAD).reduce((acc, a) => acc + (MIN_SEG_RAD - a), 0)
  const bigTotal = angles.filter(a => a >= MIN_SEG_RAD).reduce((acc, a) => acc + a, 0)
  angles = angles.map(a => {
    if (a < MIN_SEG_RAD) return MIN_SEG_RAD
    return a - deficit * (a / bigTotal)
  })
  // 安全兜底
  angles = angles.map(a => Math.max(a, MIN_SEG_RAD))
  return angles
})

// 每段弧的起止角度（弧度），与 segmentPaths 对齐
const segRanges = computed(() => {
  const startOffset = -Math.PI / 2 + Math.PI / 4 - segAngles.value[0] / 2
  const gap = 0.06
  const ranges: { a1: number; a2: number; mid: number }[] = []
  let cursor = startOffset
  for (let i = 0; i < segCount.value; i++) {
    const realA1 = cursor + gap / 2
    const realA2 = cursor + gap / 2 + segAngles.value[i]
    ranges.push({ a1: realA1, a2: realA2, mid: (realA1 + realA2) / 2 })
    cursor = realA2 + gap / 2
  }
  return ranges
})

// 4 个相位点的位置（弧段端点，与环上的节点圆点重合）
const phasePoints = computed(() => {
  const items = []
  const r = outerR.value
  for (let i = 0; i < segCount.value; i++) {
    const angle = segRanges.value[i].mid
    const x = cx.value + r * Math.cos(angle)
    const y = cy.value + r * Math.sin(angle)
    items.push({ x, y, angle })
  }
  return items
})

// 活跃段的弧线路径（用于流光 + 粒子动画）
const activeSegmentPath = computed(() => {
  if (props.currentIndex < 0 || props.currentIndex >= segCount.value) return ''
  return segmentPaths.value[props.currentIndex]?.d || ''
})

// 活跃段粒子运动周期
const activeSegDur = computed(() => '2.4s')

// 4 段弧路径（按比例分配弧度）
const segmentPaths = computed(() => {
  const r = segR.value
  return Array.from({ length: segCount.value }).map((_, i) => {
    const { a1, a2 } = segRanges.value[i]
    const x1 = cx.value + r * Math.cos(a1)
    const y1 = cy.value + r * Math.sin(a1)
    const x2 = cx.value + r * Math.cos(a2)
    const y2 = cy.value + r * Math.sin(a2)
    const large = (a2 - a1) > Math.PI ? 1 : 0
    let color = 'url(#grad-pending)'
    let opacity = 0.4
    if (i < props.currentIndex) { color = 'url(#grad-done)'; opacity = 0.85 }
    if (i === props.currentIndex) { color = 'url(#grad-active)'; opacity = 1 }
    return {
      d: `M ${x1} ${y1} A ${r} ${r} 0 ${large} 1 ${x2} ${y2}`,
      color, opacity,
    }
  })
})

// 当前阶段进度弧
const progressPath = computed(() => {
  const r = segR.value + 16
  if (props.currentIndex < 0 || props.currentIndex >= segCount.value) return ''
  const { a1 } = segRanges.value[props.currentIndex]
  const span = (props.progress / 100) * segAngles.value[props.currentIndex]
  const a2 = a1 + span
  if (props.progress <= 0) return ''
  const x1 = cx.value + r * Math.cos(a1)
  const y1 = cy.value + r * Math.sin(a1)
  const x2 = cx.value + r * Math.cos(a2)
  const y2 = cy.value + r * Math.sin(a2)
  const large = (a2 - a1) > Math.PI ? 1 : 0
  return `M ${x1} ${y1} A ${r} ${r} 0 ${large} 1 ${x2} ${y2}`
})

// 环节节点进度弧：每个环节节点在环外侧有独立的小弧形，长度反映进度
const nodeProgressArcs = computed(() => {
  const r = outerR.value + 18 // 在刻度环外侧

  const arcs: {
    d: string
    color: string
    opacity: number
    status: string
    isRunning: boolean
  }[] = []

  for (let pi = 0; pi < segCount.value; pi++) {
    const nodeStatuses = props.phaseNodeStatuses?.[pi] || []
    const names = props.phaseNames?.[pi] || []
    if (names.length === 0) continue

    // 该阶段弧段的角度范围
    const { a1: segStart, a2: segEnd } = segRanges.value[pi]
    const segAngle = segEnd - segStart

    // 每个环节节点在弧段内均分空间
    const nodeAngle = segAngle / names.length
    const nodeGap = 0.03 // 环节节点间的小间隙

    for (let ni = 0; ni < names.length; ni++) {
      const nodeStart = segStart + ni * nodeAngle + nodeGap / 2
      const nodeEnd = segStart + (ni + 1) * nodeAngle - nodeGap / 2
      const nodeSpan = nodeEnd - nodeStart

      const ns = nodeStatuses[ni] || { status: 'pending', progress: 0 }
      const progress = ns.progress / 100
      const isRunning = ns.status === 'running'
      const isDone = ns.status === 'completed'
      const isIssue = ns.status === 'issue' || ns.status === 'timeout'

      // 背景弧（完整弧段，暗色）
      const bx1 = cx.value + r * Math.cos(nodeStart)
      const by1 = cy.value + r * Math.sin(nodeStart)
      const bx2 = cx.value + r * Math.cos(nodeEnd)
      const by2 = cy.value + r * Math.sin(nodeEnd)
      const bLarge = nodeSpan > Math.PI ? 1 : 0
      const bgPath = `M ${bx1} ${by1} A ${r} ${r} 0 ${bLarge} 1 ${bx2} ${by2}`

      // 进度弧（按进度比例截取）
      const progressEnd = nodeStart + nodeSpan * progress
      let fgPath = ''
      if (progress > 0) {
        const fx1 = bx1
        const fy1 = by1
        const fx2 = cx.value + r * Math.cos(progressEnd)
        const fy2 = cy.value + r * Math.sin(progressEnd)
        const fSpan = progressEnd - nodeStart
        const fLarge = fSpan > Math.PI ? 1 : 0
        fgPath = `M ${fx1} ${fy1} A ${r} ${r} 0 ${fLarge} 1 ${fx2} ${fy2}`
      }

      // 颜色
      let fgColor = '#1a3a6a'
      let fgOpacity = 0.3
      if (isDone) { fgColor = '#00a0c8'; fgOpacity = 0.85 }
      else if (isRunning) { fgColor = '#ff9a2f'; fgOpacity = 1 }
      else if (isIssue) { fgColor = '#ff4444'; fgOpacity = 0.9 }
      else if (progress > 0) { fgColor = '#0088bb'; fgOpacity = 0.6 }

      arcs.push({
        d: bgPath,
        color: 'rgba(26, 58, 106, 0.25)',
        opacity: 1,
        status: ns.status,
        isRunning: false,
      })

      if (fgPath) {
        arcs.push({
          d: fgPath,
          color: fgColor,
          opacity: fgOpacity,
          status: ns.status,
          isRunning,
        })
      }
    }
  }

  return arcs
})

// 外圈环节标签：从各阶段的 phaseNames 实时获取，围绕对应阶段的弧段分布
const ringLabels = computed(() => {
  const labels: { text: string; active: boolean; phaseIdx: number; angleDeg: number; isRunning: boolean }[] = []

  for (let pi = 0; pi < segCount.value; pi++) {
    const names = props.phaseNames?.[pi] || []
    const nodeStatuses = props.phaseNodeStatuses?.[pi] || []
    if (names.length === 0) continue
    // 该阶段的弧段角度范围
    const { a1: segStartRad, a2: segEndRad } = segRanges.value[pi]
    const segStartDeg = segStartRad * 180 / Math.PI
    const segEndDeg = segEndRad * 180 / Math.PI

    // 在弧段内均匀分布标签
    const step = names.length > 1 ? (segEndDeg - segStartDeg) / (names.length + 1) : 0
    for (let ni = 0; ni < names.length; ni++) {
      const angleDeg = names.length === 1
        ? (segStartDeg + segEndDeg) / 2
        : segStartDeg + step * (ni + 1)
      const ns = nodeStatuses[ni] || { status: 'pending', progress: 0 }
      labels.push({
        text: names[ni],
        active: pi === props.currentIndex,
        phaseIdx: pi,
        angleDeg,
        isRunning: ns.status === 'running',
      })
    }
  }
  return labels
})

const labelR = computed(() => outerR.value + 30)

const ringLabelPositions = computed(() => {
  const totalW = size.value + PAD_X * 2
  const totalH = size.value + PAD_Y_TOP + PAD_Y_BOTTOM
  return ringLabels.value.map((lbl) => {
    const a = lbl.angleDeg * Math.PI / 180
    const cosA = Math.cos(a)
    const sinA = Math.sin(a)
    // 所有标签与中心环保持相同距离
    const r = labelR.value
    const px = cx.value + r * cosA
    const py = cy.value + r * sinA
    const flipped = cosA < -0.08
    // 顶部区域标签倾斜优化：避免文字垂直朝上被遮挡
    let rotate = flipped ? lbl.angleDeg + 180 : lbl.angleDeg
    const TOP_MIN = -135
    const TOP_MAX = -45
    if (!flipped && lbl.angleDeg > TOP_MIN && lbl.angleDeg < TOP_MAX) {
      // 顶部右侧：限制旋转角不超过 -50°（向右倾斜）
      rotate = Math.max(rotate, -50)
    }
    if (flipped) {
      const flippedAngle = lbl.angleDeg + 180
      if (flippedAngle > 135 && flippedAngle < 225) {
        // 顶部左侧：限制旋转角不超过 230°（向左倾斜）
        rotate = Math.min(rotate, 230)
      }
    }
    return {
      leftPct: (px / totalW) * 100,
      topPct: (py / totalH) * 100,
      rotate,
      flipped,
    }
  })
})

// 连接线：从标签到环外缘
const ringLabelLines = computed(() => {
  return ringLabels.value.map((lbl) => {
    const a = lbl.angleDeg * Math.PI / 180
    const rLabel = labelR.value - 4
    const rRing = outerR.value + 8   // 环端（稍超出刻度）
    return {
      x1: cx.value + rLabel * Math.cos(a),
      y1: cy.value + rLabel * Math.sin(a),
      x2: cx.value + rRing * Math.cos(a),
      y2: cy.value + rRing * Math.sin(a),
    }
  })
})

// 运行中环节节点的位置信息（用于波纹动画）
const runningRipples = computed(() => {
  return ringLabels.value
    .map((lbl) => {
      if (!lbl.isRunning) return null
      const a = lbl.angleDeg * Math.PI / 180
      const rRing = outerR.value + 8
      return {
        cx: cx.value + rRing * Math.cos(a),
        cy: cy.value + rRing * Math.sin(a),
      }
    })
    .filter(Boolean) as { cx: number; cy: number }[]
})

function chineseNum(n: number): string {
  return ['一', '二', '三', '四', '五', '六'][n - 1] || String(n)
}
</script>

<style lang="scss" scoped>
.phase-ring {
  display: flex; align-items: center; justify-content: center;
  font-family: 'Microsoft YaHei', 'PingFang SC', sans-serif;
}
.ring-inner {
  position: relative;
  flex-shrink: 0;
}
.ring-svg {
  position: absolute; top: 0; left: 0;
  pointer-events: none;
}
.pulse-node {
  animation: node-pulse 1.6s ease-in-out infinite;
  transform-origin: center;
  transform-box: fill-box;
}
@keyframes node-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.55; filter: drop-shadow(0 0 8px #ff7a00); }
}

// 活跃段流光动画
.flow-light-streak {
  stroke-dasharray: 30 200;
  stroke-dashoffset: 0;
  animation: flow-streak 2.4s linear infinite;
  will-change: stroke-dashoffset;
}
@keyframes flow-streak {
  from { stroke-dashoffset: 230; }
  to { stroke-dashoffset: 0; }
}

// 环节节点进度弧 - 运行中动画
.arc-running {
  animation: arc-pulse 1.5s ease-in-out infinite;
  filter: drop-shadow(0 0 4px rgba(255, 154, 47, 0.6));
}
@keyframes arc-pulse {
  0%, 100% { opacity: 1; filter: drop-shadow(0 0 4px rgba(255, 154, 47, 0.6)); }
  50% { opacity: 0.7; filter: drop-shadow(0 0 8px rgba(255, 154, 47, 0.9)); }
}

// 活跃段粒子光晕
.flow-particle {
  filter: drop-shadow(0 0 4px rgba(255, 180, 74, 0.8));
}
.flow-particle-1 {
  animation: particle-glow 1.2s ease-in-out infinite;
}
.flow-particle-2 {
  animation: particle-glow 1.2s ease-in-out infinite 0.3s;
}
.flow-particle-3 {
  animation: particle-glow 1.2s ease-in-out infinite 0.6s;
}
@keyframes particle-glow {
  0%, 100% { opacity: 0.9; }
  50% { opacity: 0.4; }
}

// 雷达扫描旋转（由浏览器合成器线程处理，无需 GPU 硬件）
.radar-sweep {
  transform-origin: center;
  transform-box: view-box;
  animation: radar-rotate 10s linear infinite;
  will-change: transform;
}
@keyframes radar-rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

// 同心环微脉冲
.concentric-rings {
  animation: ring-pulse 6s ease-in-out infinite;
}
@keyframes ring-pulse {
  0%, 100% { opacity: 0.18; }
  50% { opacity: 0.09; }
}

// 无障碍：减少动画
@media (prefers-reduced-motion: reduce) {
  .radar-sweep { animation: none; opacity: 0.3; }
  .concentric-rings { animation: none; }
  .flow-light-streak { animation: none; stroke-dasharray: none; }
  .flow-particle { display: none; }
  .pulse-node { animation: none; }
  .arc-running { animation: none; }
  .label-node-running { animation: none; }
  .label-line-running { animation: none; }
  .ripple-1, .ripple-2, .ripple-3 { animation: none; display: none; }
  .running-halo { animation: none; }
  .running-dot { animation: none; }
  .running-dot-ripple { animation: none; display: none; }
  .ring-outer-label-running .ring-outer-label-content { animation: none; }
  .ring-outer-label-running .ring-outer-label-dot { animation: none; }
  .ring-outer-label-running .dot-core,
  .ring-outer-label-running .dot-orbit,
  .ring-outer-label-running .ring-outer-label-text::after,
  .ring-outer-label-running .running-tag,
  .ring-outer-label-running .running-tag-dot { animation: none !important; }
  .ring-outer-label-running .dot-ripple { display: none !important; }
}

// 阶段圆形节点
.phase-node-label {
  position: absolute;
  transform: translate(-50%, -50%);
  pointer-events: none;
  z-index: 5;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(4, 18, 48, 0.85);
  border: 2px solid rgba(0, 212, 255, 0.45);
  box-shadow: 0 0 12px rgba(0, 212, 255, 0.2);
  transition: all 0.3s ease;
}
.phase-node-text {
  font-family: 'Microsoft YaHei', 'PingFang SC', sans-serif;
  font-size: 11px;
  font-weight: 700;
  color: #8cb8e0;
  letter-spacing: 1px;
  white-space: nowrap;
}
.phase-node-done {
  border-color: rgba(0, 160, 200, 0.6);
  background: rgba(0, 40, 70, 0.75);
  .phase-node-text { color: #5aafcc; }
}
.phase-node-active {
  border-color: #ff9a2f;
  background: rgba(80, 38, 8, 0.88);
  box-shadow: 0 0 22px rgba(255, 122, 0, 0.45), inset 0 0 10px rgba(255, 122, 0, 0.12);
  animation: phase-node-pulse 2s ease-in-out infinite;
  .phase-node-text { color: #ffd699; }
}
@keyframes phase-node-pulse {
  0%, 100% { box-shadow: 0 0 22px rgba(255, 122, 0, 0.45), inset 0 0 10px rgba(255, 122, 0, 0.12); }
  50% { box-shadow: 0 0 30px rgba(255, 122, 0, 0.65), inset 0 0 14px rgba(255, 122, 0, 0.18); }
}

.ring-svg-overlay {
  z-index: 1;
  pointer-events: none;
}

// 径向环节标签：文字方向接近圆环半径，避免顶部/底部横向堆叠
.label-node-running {
  animation: label-node-pulse 1.4s ease-in-out infinite;
  filter: drop-shadow(0 0 8px rgba(255, 180, 74, 0.85));
}
@keyframes label-node-pulse {
  0%, 100% { opacity: 0.95; }
  50% { opacity: 0.55; }
}

// 运行中连接线动画
.label-line-running {
  animation: line-flow 1.2s ease-in-out infinite;
}
@keyframes line-flow {
  0%, 100% { opacity: 0.6; stroke-dashoffset: 0; }
  50% { opacity: 1; stroke-dashoffset: 4; }
}

// SVG 波纹扩散动画
.ripple {
  transform-origin: center;
  transform-box: fill-box;
  fill: none;
}
.ripple-1 {
  animation: ripple-expand 2s ease-out infinite;
}
.ripple-2 {
  animation: ripple-expand 2s ease-out infinite 0.5s;
}
.ripple-3 {
  animation: ripple-expand 2s ease-out infinite 1s;
}
@keyframes ripple-expand {
  0% {
    r: 4;
    opacity: 0.9;
    stroke-width: 2.5;
  }
  60% {
    opacity: 0.3;
  }
  100% {
    r: 22;
    opacity: 0;
    stroke-width: 0.5;
  }
}

// 运行中节点呼吸光晕
.running-halo {
  animation: halo-breathe 1.8s ease-in-out infinite;
  transform-origin: center;
  transform-box: fill-box;
}
@keyframes halo-breathe {
  0%, 100% { r: 8; opacity: 0.35; stroke-width: 1.5; }
  50% { r: 12; opacity: 0.1; stroke-width: 0.5; }
}
.ring-outer-label {
  position: absolute;
  z-index: 4;
  width: 0;
  height: 0;
  transform: translate(-50%, -50%) rotate(var(--label-rotate, 0deg));
  transform-origin: center;
  pointer-events: none;
}
.ring-outer-label-content {
  position: absolute;
  left: 8px;
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  align-items: center;
  max-width: 96px;
  padding: 1px 5px;
  background: linear-gradient(90deg, rgba(4, 18, 48, 0.75), rgba(4, 18, 48, 0));
}
.ring-outer-label-flipped .ring-outer-label-content {
  left: auto;
  right: 8px;
  background: linear-gradient(270deg, rgba(4, 18, 48, 0.75), rgba(4, 18, 48, 0));
}
.ring-outer-label-text {
  display: block;
  max-width: 86px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: rgba(178, 216, 246, 0.72);
  font-size: 10px;
  line-height: 1.1;
  letter-spacing: 0.2px;
  text-shadow: 0 0 8px rgba(2, 14, 36, 0.95);
}
.ring-outer-label-active .ring-outer-label-text {
  color: rgba(220, 240, 255, 0.92);
}
.ring-outer-label-running {
  z-index: 7;
  .ring-outer-label-content {
    background: linear-gradient(90deg, rgba(84, 42, 8, 0.88), rgba(84, 42, 8, 0));
    border-radius: 3px;
    padding: 2px 6px;
    box-shadow: 0 0 12px rgba(255, 154, 47, 0.35);
    animation: label-bg-pulse 1.8s ease-in-out infinite;
  }
  .ring-outer-label-text {
    color: #ffe2b5;
    font-weight: 700;
    text-shadow: 0 0 6px rgba(255, 154, 47, 0.6);
  }
}
@keyframes label-bg-pulse {
  0%, 100% { box-shadow: 0 0 12px rgba(255, 154, 47, 0.35); background: linear-gradient(90deg, rgba(84, 42, 8, 0.88), rgba(84, 42, 8, 0)); }
  50% { box-shadow: 0 0 20px rgba(255, 154, 47, 0.55); background: linear-gradient(90deg, rgba(100, 48, 8, 0.95), rgba(84, 42, 8, 0)); }
}
.ring-outer-label-running.ring-outer-label-flipped .ring-outer-label-content {
  background: linear-gradient(270deg, rgba(84, 42, 8, 0.88), rgba(84, 42, 8, 0));
  animation: label-bg-pulse-flipped 1.8s ease-in-out infinite;
}
@keyframes label-bg-pulse-flipped {
  0%, 100% { box-shadow: 0 0 12px rgba(255, 154, 47, 0.35); background: linear-gradient(270deg, rgba(84, 42, 8, 0.88), rgba(84, 42, 8, 0)); }
  50% { box-shadow: 0 0 20px rgba(255, 154, 47, 0.55); background: linear-gradient(270deg, rgba(100, 48, 8, 0.95), rgba(84, 42, 8, 0)); }
}

// 运行指示器（标签前的圆点 + 波纹）
.running-indicator {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 8px;
  height: 8px;
  margin-right: 4px;
  flex-shrink: 0;
}
.running-dot {
  position: absolute;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #ffb44a;
  box-shadow: 0 0 6px rgba(255, 154, 47, 0.9);
  animation: dot-core-pulse 1.4s ease-in-out infinite;
}
@keyframes dot-core-pulse {
  0%, 100% { transform: scale(1); opacity: 1; }
  50% { transform: scale(0.7); opacity: 0.7; }
}
.running-dot-ripple {
  position: absolute;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  border: 1.5px solid rgba(255, 180, 74, 0.7);
  animation: dot-ripple-expand 2s ease-out infinite;
}
.running-dot-ripple.delay {
  animation-delay: 0.7s;
}
@keyframes dot-ripple-expand {
  0% { transform: scale(1); opacity: 0.8; }
  100% { transform: scale(3.5); opacity: 0; }
}

// 中心内容
.center-content {
  position: absolute;
  top: 50%; left: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
  z-index: 3;
  display: flex; flex-direction: column; align-items: center; gap: 6px;
}
.center-caption {
  font-size: 14px;
  color: #a0c4f0;
  letter-spacing: 4px;
  font-family: 'Microsoft YaHei', 'PingFang SC', sans-serif;
  font-weight: 500;
  text-shadow: 0 0 10px rgba(0, 120, 255, 0.3);
  position: relative;
  padding-bottom: 8px;
  &::after {
    content: '';
    position: absolute;
    bottom: 0;
    left: 50%;
    transform: translateX(-50%);
    width: 48px;
    height: 1px;
    background: linear-gradient(90deg, transparent, rgba(0, 212, 255, 0.5), transparent);
  }
}
.center-value {
  display: flex; align-items: center; gap: 4px;
  font-family: 'Share Tech Mono', monospace;
  color: #00d4ff;
  text-shadow: 0 0 8px rgba(0, 212, 255, 0.4);
  .num-num { font-size: 18px; }
  .num-sep { font-size: 16px; opacity: 0.5; }
  .num-den { font-size: 16px; opacity: 0.7; }
}
.center-pct {
  font-family: 'Share Tech Mono', monospace;
  font-size: 78px;
  font-weight: 900;
  line-height: 1;
  color: #ffffff;
  text-shadow:
    0 4px 0 rgba(255, 122, 0, 0.9),
    0 0 18px rgba(255, 122, 0, 0.7),
    0 0 42px rgba(38, 118, 255, 0.5);
  .pct-unit { font-size: 34px; opacity: 0.95; margin-left: 2px; }
}
.center-hint {
  font-size: 11px;
  color: #6e8db5;
  letter-spacing: 1px;
  margin-top: 2px;
}
</style>
