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
      <!-- 连接线：从环节标签到环边缘 -->
      <line
        v-for="(lp, idx) in ringLabelLines"
        :key="'rll' + idx"
        :x1="lp.x1" :y1="lp.y1"
        :x2="lp.x2" :y2="lp.y2"
        stroke="rgba(0, 212, 255, 0.25)"
        stroke-width="1"
        stroke-dasharray="3 3"
      />
      <!-- 连接端点小圆 -->
      <circle
        v-for="(lp, idx) in ringLabelLines"
        :key="'rlc' + idx"
        :cx="lp.x2" :cy="lp.y2"
        r="2.5"
        fill="#00d4ff"
        opacity="0.5"
      />
    </svg>
    <div
      v-for="(lp, idx) in ringLabelPositions"
      :key="'rl' + idx"
      class="ring-outer-label"
      :class="[
        { 'ring-outer-label-active': ringLabels[idx].active },
        { 'ring-outer-label-running': ringLabels[idx].isRunning },
        'ring-outer-label-phase-' + ringLabels[idx].phaseIdx,
        lp.align === 'left' ? 'label-align-left' : lp.align === 'right' ? 'label-align-right' : 'label-align-center',
      ]"
      :style="{
        left: lp.leftPct + '%',
        top: lp.topPct + '%',
      }"
    >
      <span class="ring-outer-label-dot">
        <!-- 运行中: 双层波纹 + 旋转扫描 -->
        <template v-if="ringLabels[idx].isRunning">
          <span class="dot-ripple dot-ripple-1"></span>
          <span class="dot-ripple dot-ripple-2"></span>
          <span class="dot-ripple dot-ripple-3"></span>
          <span class="dot-core"></span>
          <span class="dot-orbit"></span>
        </template>
      </span>
      <span class="ring-outer-label-text">
        {{ ringLabels[idx].text }}
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
const PAD_X = 70  // 左右预留空间
const PAD_Y_TOP = 60  // 顶部预留（阶段节点在上方时较小）
const PAD_Y_BOTTOM = 90  // 底部预留（标签文字需要更多空间）
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

// 4 个相位点的位置（弧段端点，与环上的节点圆点重合）
const phasePoints = computed(() => {
  const items = []
  const r = outerR.value
  for (let i = 0; i < 4; i++) {
    // 与 segmentPaths 的角度一致：4 个节点在 45°, 135°, 225°, 315°
    const angle = -Math.PI / 2 + (Math.PI / 2) * i + Math.PI / 4
    const x = cx.value + r * Math.cos(angle)
    const y = cy.value + r * Math.sin(angle)
    items.push({ x, y, angle })
  }
  return items
})

// 活跃段的弧线路径（用于流光 + 粒子动画）
const activeSegmentPath = computed(() => {
  if (props.currentIndex < 0 || props.currentIndex >= 4) return ''
  return segmentPaths.value[props.currentIndex]?.d || ''
})

// 活跃段粒子运动周期
const activeSegDur = computed(() => '2.4s')

// 4 段弧路径
const segmentPaths = computed(() => {
  const r = segR.value
  const gap = 0.06 // 段间隙（弧度）
  const segCount = 4
  const totalSpan = Math.PI * 2
  const segSpan = totalSpan / segCount
  const startOffset = -Math.PI / 2 + Math.PI / 4 - segSpan / 2 // 让第一段对称
  return Array.from({ length: segCount }).map((_, i) => {
    const a1 = startOffset + i * segSpan + gap / 2
    const a2 = startOffset + (i + 1) * segSpan - gap / 2
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
  const segCount = 4
  const segSpan = (Math.PI * 2) / segCount
  const startOffset = -Math.PI / 2 + Math.PI / 4 - segSpan / 2
  const a1 = startOffset + props.currentIndex * segSpan
  const span = (props.progress / 100) * segSpan
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
  const segCount = props.phases.length || 4
  const segSpan = (Math.PI * 2) / segCount
  const startOffset = -Math.PI / 2 + Math.PI / 4 - segSpan / 2
  const gap = 0.06 // 与 segmentPaths 一致的段间隙

  const arcs: {
    d: string
    color: string
    opacity: number
    status: string
    isRunning: boolean
  }[] = []

  for (let pi = 0; pi < segCount; pi++) {
    const nodeStatuses = props.phaseNodeStatuses?.[pi] || []
    const names = props.phaseNames?.[pi] || []
    if (names.length === 0) continue

    // 该阶段弧段的角度范围
    const segStart = startOffset + pi * segSpan + gap / 2
    const segEnd = startOffset + (pi + 1) * segSpan - gap / 2
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

// 外圈环节标签：从各阶段的 phaseNames 实时获取，均匀围绕对应阶段的弧段
const ringLabels = computed(() => {
  const labels: { text: string; active: boolean; phaseIdx: number; angleDeg: number; isRunning: boolean }[] = []
  const segCount = props.phases.length || 4
  const segSpan = 360 / segCount
  const startOffsetDeg = -90 + 90 / segCount - segSpan / 2 // 与 segmentPaths 的 startOffset 对应

  for (let pi = 0; pi < segCount; pi++) {
    const names = props.phaseNames?.[pi] || []
    const nodeStatuses = props.phaseNodeStatuses?.[pi] || []
    if (names.length === 0) continue
    // 每个阶段的弧段角度范围
    const segStartDeg = startOffsetDeg + pi * segSpan
    const segEndDeg = segStartDeg + segSpan
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

// 标签放置半径（比阶段圆形节点再外一些，避免重叠）
const labelR = computed(() => outerR.value + 52)

// 标签 HTML 定位（百分比）
const ringLabelPositions = computed(() => {
  const totalW = size.value + PAD_X * 2
  const totalH = size.value + PAD_Y_TOP + PAD_Y_BOTTOM
  return ringLabels.value.map((lbl) => {
    const a = lbl.angleDeg * Math.PI / 180
    const r = labelR.value
    const px = cx.value + r * Math.cos(a)
    const py = cy.value + r * Math.sin(a)
    let align: 'left' | 'right' | 'center' = 'center'
    if (Math.cos(a) > 0.2) align = 'left'
    else if (Math.cos(a) < -0.2) align = 'right'
    return {
      leftPct: (px / totalW) * 100,
      topPct: (py / totalH) * 100,
      align,
    }
  })
})

// 连接线：从标签到环外缘
const ringLabelLines = computed(() => {
  return ringLabels.value.map((lbl) => {
    const a = lbl.angleDeg * Math.PI / 180
    const rLabel = labelR.value - 16 // 标签端（稍内缩）
    const rRing = outerR.value + 8   // 环端（稍超出刻度）
    return {
      x1: cx.value + rLabel * Math.cos(a),
      y1: cy.value + rLabel * Math.sin(a),
      x2: cx.value + rRing * Math.cos(a),
      y2: cy.value + rRing * Math.sin(a),
    }
  })
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
  animation: radar-rotate 5s linear infinite;
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

// 外圈环节标签（雷达环形布局）
.ring-outer-label {
  position: absolute;
  z-index: 4;
  transform: translate(-50%, -50%);
  pointer-events: none;
  display: flex;
  align-items: center;
  gap: 6px;
  white-space: nowrap;
}
.label-align-center {
  transform: translate(-50%, -50%);
}
.label-align-left {
  transform: translate(4px, -50%);
}
.label-align-right {
  transform: translate(calc(-100% - 4px), -50%);
  flex-direction: row-reverse;
}
.ring-outer-label-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: rgba(0, 212, 255, 0.5);
  flex-shrink: 0;
  box-shadow: 0 0 4px rgba(0, 212, 255, 0.3);
}
.ring-outer-label-text {
  font-size: 12px;
  color: rgba(180, 210, 240, 0.85);
  letter-spacing: 1.5px;
  font-weight: 400;
  text-shadow: 0 0 6px rgba(0, 30, 80, 0.8);
}
.ring-outer-label-active .ring-outer-label-dot {
  background: #00d4ff;
  box-shadow: 0 0 8px rgba(0, 212, 255, 0.6);
}
.ring-outer-label-active .ring-outer-label-text {
  color: #d6e8ff;
  font-weight: 600;
}

// 运行中的环节节点 - 醒目动态高亮（雷达波纹 + 旋转扫描点 + 流光下划线 + RUNNING 徽章）
.ring-outer-label-running {
  z-index: 6;

  .ring-outer-label-dot {
    position: relative;
    width: 14px;
    height: 14px;
    background: transparent !important;
    box-shadow: none !important;
    overflow: visible;
    flex-shrink: 0;
  }

  // 中心实心点（始终可见，做呼吸缩放）
  .dot-core {
    position: absolute;
    top: 50%; left: 50%;
    width: 8px; height: 8px;
    margin: -4px 0 0 -4px;
    border-radius: 50%;
    background: radial-gradient(circle at 30% 30%, #ffe5b8, #ff9a2f 60%, #ff5800);
    box-shadow:
      0 0 8px rgba(255, 154, 47, 0.95),
      0 0 16px rgba(255, 100, 0, 0.55);
    animation: dot-core-breath 1.4s ease-in-out infinite;
    z-index: 3;
  }

  // 三层向外扩散的波纹环（依次延迟）
  .dot-ripple {
    position: absolute;
    top: 50%; left: 50%;
    width: 14px; height: 14px;
    margin: -7px 0 0 -7px;
    border: 2px solid rgba(255, 154, 47, 0.85);
    border-radius: 50%;
    opacity: 0;
    animation: dot-ripple-out 2.1s cubic-bezier(0.22, 0.61, 0.36, 1) infinite;
    pointer-events: none;
  }
  .dot-ripple-1 { animation-delay: 0s; }
  .dot-ripple-2 { animation-delay: 0.7s; }
  .dot-ripple-3 { animation-delay: 1.4s; }

  // 围绕节点旋转的小光点（轨道扫描）
  .dot-orbit {
    position: absolute;
    top: 50%; left: 50%;
    width: 24px; height: 24px;
    margin: -12px 0 0 -12px;
    border-radius: 50%;
    animation: dot-orbit-rotate 2.4s linear infinite;
    pointer-events: none;
    z-index: 2;

    &::before {
      content: '';
      position: absolute;
      top: -2px; left: 50%;
      width: 4px; height: 4px;
      margin-left: -2px;
      border-radius: 50%;
      background: #ffd699;
      box-shadow: 0 0 6px #ff9a2f, 0 0 12px rgba(255, 154, 47, 0.7);
    }
  }

  .ring-outer-label-text {
    position: relative;
    color: #fff3da !important;
    font-weight: 700 !important;
    letter-spacing: 1.5px;
    text-shadow:
      0 0 6px rgba(255, 154, 47, 0.9),
      0 0 14px rgba(255, 100, 0, 0.6),
      0 0 24px rgba(255, 60, 0, 0.35) !important;

    // 文字底部流光横线
    &::after {
      content: '';
      position: absolute;
      left: 0; right: 0; bottom: -3px;
      height: 1px;
      background: linear-gradient(
        90deg,
        transparent,
        rgba(255, 154, 47, 0.9),
        rgba(255, 220, 150, 1),
        rgba(255, 154, 47, 0.9),
        transparent
      );
      background-size: 200% 100%;
      animation: text-underline-flow 2s linear infinite;
    }
  }

  // RUNNING 徽章
  .running-tag {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    margin-left: 6px;
    padding: 1px 6px 1px 5px;
    font-size: 9px;
    font-weight: 700;
    letter-spacing: 1px;
    color: #ff9a2f;
    background: rgba(255, 100, 0, 0.12);
    border: 1px solid rgba(255, 154, 47, 0.55);
    border-radius: 2px;
    vertical-align: middle;
    text-shadow: none;
    animation: running-tag-pulse 1.6s ease-in-out infinite;

    .running-tag-dot {
      width: 4px;
      height: 4px;
      border-radius: 50%;
      background: #ff9a2f;
      box-shadow: 0 0 4px #ff9a2f;
      animation: running-tag-blink 0.9s ease-in-out infinite;
    }

    .running-tag-text {
      font-family: 'Share Tech Mono', monospace;
    }
  }
}

@keyframes dot-core-breath {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.25); }
}

@keyframes dot-ripple-out {
  0% { transform: scale(0.6); opacity: 0.9; border-width: 2px; }
  70% { opacity: 0.25; border-width: 1px; }
  100% { transform: scale(3.4); opacity: 0; border-width: 1px; }
}

@keyframes dot-orbit-rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@keyframes text-underline-flow {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

@keyframes running-tag-pulse {
  0%, 100% {
    border-color: rgba(255, 154, 47, 0.55);
    box-shadow: 0 0 0 rgba(255, 154, 47, 0);
  }
  50% {
    border-color: rgba(255, 200, 100, 0.95);
    box-shadow: 0 0 8px rgba(255, 154, 47, 0.5);
  }
}

@keyframes running-tag-blink {
  0%, 60%, 100% { opacity: 1; }
  30% { opacity: 0.25; }
}

// 已完成阶段的标签（phaseIdx < currentIndex）
.ring-outer-label[class*="ring-outer-label-phase-"]:not(.ring-outer-label-active) {
  .ring-outer-label-dot {
    background: rgba(0, 160, 200, 0.4);
    box-shadow: 0 0 4px rgba(0, 160, 200, 0.2);
  }
  .ring-outer-label-text {
    color: rgba(140, 180, 220, 0.6);
  }
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
