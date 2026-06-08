<template>
  <div class="phase-ring" :style="{ width: containerSize + 'px', height: containerSize + 'px' }">
    <div class="ring-inner" :style="{ width: size + 'px', height: size + 'px' }">
    <svg :viewBox="`0 0 ${size} ${size}`" :width="size" :height="size" class="ring-svg">
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

      <!-- 4 个阶段的连接点（圆点） -->
      <g class="phase-nodes">
        <circle
          v-for="(p, idx) in phasePoints"
          :key="'p' + idx"
          :cx="p.x"
          :cy="p.y"
          :r="idx === currentIndex ? 8 : 5"
          :fill="idx === currentIndex ? '#ff7a00' : idx < currentIndex ? '#00d4ff' : '#1a3a6a'"
          :stroke="idx === currentIndex ? '#fff4cf' : 'rgba(0, 212, 255, 0.4)'"
          stroke-width="2"
          :class="idx === currentIndex ? 'pulse-node' : ''"
        />
      </g>

    </svg>

    <!-- 4 个阶段名称（外侧 4 角，HTML 定位） -->
    <div
      v-for="(p, idx) in phasePoints"
      :key="'lbl' + idx"
      class="phase-label"
      :class="['phase-label-' + idx, idx === currentIndex ? 'phase-label-active' : '', idx < currentIndex ? 'phase-label-done' : '']"
      :style="{
        left: (p.x / size * 100) + '%',
        top: (p.y / size * 100) + '%',
        '--lx': (p.lx) + 'px',
        '--ly': (p.ly) + 'px',
      }"
    >
      <div class="phase-label-tag">
        <span class="phase-label-num">阶段{{ chineseNum(idx + 1) }}</span>
        <span class="phase-label-name">{{ phases[idx] || '--' }}</span>
      </div>
    </div>

    <!-- 6 个环节标签（环形外侧，雷达布局） -->
    <svg :viewBox="`0 0 ${size} ${size}`" :width="size" :height="size" class="ring-svg ring-svg-overlay">
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
        lp.align === 'left' ? 'label-align-left' : lp.align === 'right' ? 'label-align-right' : 'label-align-center',
      ]"
      :style="{
        left: lp.leftPct + '%',
        top: lp.topPct + '%',
      }"
    >
      <span class="ring-outer-label-dot" />
      <span class="ring-outer-label-text">{{ ringLabels[idx].text }}</span>
    </div>

    <!-- 中心数字 -->
    <div class="center-content">
      <div class="center-caption">演练总进度</div>
      <div class="center-value">
        <span class="num-num">{{ centerNumerator }}</span>
        <span class="num-sep">/</span>
        <span class="num-den">{{ centerDenominator }}</span>
      </div>
      <div class="center-pct">{{ progress }}<span class="pct-unit">%</span></div>
      <div class="center-hint">{{ centerHint }}</div>
    </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  phases: string[]
  currentIndex: number
  progress: number
  centerNumerator: number
  centerDenominator: number
  centerHint: string
  size?: number
}>()

const size = computed(() => props.size ?? 520)
const PAD = 50 // 外圈标签预留空间
const containerSize = computed(() => size.value + PAD * 2)
const cx = computed(() => size.value / 2)
const cy = computed(() => size.value / 2)

// 半径定义
const innerR = computed(() => size.value * 0.18)
const segR = computed(() => size.value * 0.36)   // 分段环
const outerR = computed(() => size.value * 0.46) // 刻度

// 4 个相位点的位置（外侧 4 个角）
const phasePoints = computed(() => {
  const list = props.phases.length || 4
  const items = []
  for (let i = 0; i < 4; i++) {
    // 4 个相位 = 90° 间隔，从右上 - 右下 - 左下 - 左上
    const angle = -Math.PI / 2 + (Math.PI / 2) * i + Math.PI / 4 // 起始 45°
    const r = outerR.value
    const x = cx.value + r * Math.cos(angle)
    const y = cy.value + r * Math.sin(angle)
    // 标签偏移（左侧系数较小，补偿 translateX(-100%) 造成的额外左偏）
    const lx = Math.cos(angle) * (Math.cos(angle) > 0 ? 80 : 70)
    const ly = Math.sin(angle) * 22
    items.push({ x, y, lx, ly, angle })
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

// 外圈 6 个环节标签（环形外侧，雷达布局）
const RING_LABELS = [
  '业务验收与告警',
  '演练复盘与行动',
  '基线指标与备份',
  '故障定位与影响',
  '演练启动与人员',
  '执行恢复与流量',
]
// 6 个标签角度（12 点起，顺时针 60° 间隔）
const RING_LABEL_ANGLES_DEG = [-90, -30, 30, 90, 150, 210]

const ringLabels = computed(() => {
  return RING_LABELS.map((text, i) => {
    // 高亮当前阶段相关标签
    const active = i === 0 // 默认第一个高亮
    return { text, active }
  })
})

// 标签放置半径（比外圈刻度再外一些）
const labelR = computed(() => outerR.value + 28)

// 标签 HTML 定位（百分比）
const ringLabelPositions = computed(() => {
  return RING_LABEL_ANGLES_DEG.map((deg) => {
    const a = deg * Math.PI / 180
    const r = labelR.value
    const px = cx.value + r * Math.cos(a)
    const py = cy.value + r * Math.sin(a)
    // 文字对齐：右侧标签左对齐，左侧标签右对齐，顶部/底部居中
    let align: 'left' | 'right' | 'center' = 'center'
    if (Math.cos(a) > 0.2) align = 'left'
    else if (Math.cos(a) < -0.2) align = 'right'
    return {
      leftPct: (px / size.value) * 100,
      topPct: (py / size.value) * 100,
      align,
    }
  })
})

// 连接线：从标签到环外缘
const ringLabelLines = computed(() => {
  return RING_LABEL_ANGLES_DEG.map((deg) => {
    const a = deg * Math.PI / 180
    const rLabel = labelR.value - 12 // 标签端（稍内缩）
    const rRing = outerR.value + 6   // 环端（稍超出刻度）
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
}

.phase-label {
  position: absolute;
  transform: translate(calc(-50% + var(--lx)), calc(-50% + var(--ly)));
  pointer-events: none;
  z-index: 2;
  &::before {
    content: '';
    position: absolute;
    left: -16px; top: 50%;
    width: 14px; height: 1px;
    background: linear-gradient(90deg, transparent, rgba(0, 212, 255, 0.4));
  }
}
.phase-label-tag {
  display: flex; flex-direction: column;
  background: rgba(0, 30, 80, 0.55);
  border: 1px solid rgba(0, 212, 255, 0.3);
  padding: 4px 10px;
  white-space: nowrap;
  position: relative;
  &::before, &::after {
    content: '';
    position: absolute; width: 6px; height: 6px;
    border: 1px solid #00d4ff;
  }
  &::before { top: -1px; left: -1px; border-right: 0; border-bottom: 0; }
  &::after { bottom: -1px; right: -1px; border-left: 0; border-top: 0; }
  .phase-label-num {
    font-family: 'Share Tech Mono', 'Consolas', monospace;
    font-size: 11px;
    color: #00d4ff;
    letter-spacing: 1px;
  }
  .phase-label-name {
    font-size: 13px;
    color: #d6e8ff;
    font-weight: 600;
    letter-spacing: 1px;
  }
}
.phase-label-done .phase-label-tag {
  border-color: rgba(0, 212, 255, 0.55);
  .phase-label-name { color: #b8d8ff; }
}
.phase-label-active .phase-label-tag {
  border-color: #ff9a2f;
  background: rgba(93, 48, 10, 0.72);
  box-shadow: 0 0 18px rgba(255, 122, 0, 0.38);
  &::before, &::after { border-color: #ff9a2f; }
  .phase-label-num { color: #ffb44a; }
  .phase-label-name { color: #ffffff; }
}

// 4 个角的标签方向修正
.phase-label-0 {
  // 右上
  transform: translate(0, calc(-50% + var(--ly)));
  .phase-label-tag { margin-left: 18px; }
  &::before { left: -18px; }
}
.phase-label-1 {
  // 右下
  .phase-label-tag { margin-left: 18px; }
  &::before { left: -18px; }
}
.phase-label-2 {
  // 左下
  .phase-label-tag { margin-left: -4px; transform: translateX(-100%); }
  &::before { right: -16px; left: auto; background: linear-gradient(90deg, rgba(0, 212, 255, 0.4), transparent); }
}
.phase-label-3 {
  // 左上
  .phase-label-tag { margin-left: -4px; transform: translateX(-100%); }
  &::before { right: -16px; left: auto; background: linear-gradient(90deg, rgba(0, 212, 255, 0.4), transparent); }
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
  font-size: 12px;
  color: rgba(255, 180, 74, 0.85);
  letter-spacing: 2px;
  margin-top: 6px;
  font-weight: 600;
  text-shadow: 0 0 8px rgba(255, 122, 0, 0.35);
  position: relative;
  padding: 3px 14px;
  background: rgba(255, 122, 0, 0.08);
  border: 1px solid rgba(255, 122, 0, 0.25);
  white-space: nowrap;
  &::before, &::after {
    content: '';
    position: absolute;
    width: 4px; height: 4px;
    border: 1px solid rgba(255, 180, 74, 0.5);
  }
  &::before { top: -1px; left: -1px; border-right: 0; border-bottom: 0; }
  &::after { bottom: -1px; right: -1px; border-left: 0; border-top: 0; }
}
</style>
