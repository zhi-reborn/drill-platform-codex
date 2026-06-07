<template>
  <div class="phase-ring" :style="{ width: size + 'px', height: size + 'px' }">
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
      </defs>

      <!-- 中心光晕 -->
      <circle :cx="cx" :cy="cy" :r="innerR" fill="url(#center-glow)" />

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

      <!-- 内圈 6 个环节标签（径向） -->
      <g class="inner-labels">
        <text
          v-for="(label, idx) in innerLabels"
          :key="'il' + idx"
          :x="innerLabelPoints[idx].x"
          :y="innerLabelPoints[idx].y"
          :fill="label.color"
          :font-size="label.size"
          :font-weight="label.weight"
          text-anchor="middle"
          dominant-baseline="middle"
          font-family="'Microsoft YaHei', 'PingFang SC', sans-serif"
        >
          {{ label.text }}
        </text>
      </g>
    </svg>

    <!-- 4 个阶段名称（外侧，HTML 定位） -->
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

    <!-- 中心数字 -->
    <div class="center-content">
      <div class="center-caption">{{ centerCaption }}</div>
      <div class="center-value">
        <span class="num-num">{{ centerNumerator }}</span>
        <span class="num-sep">/</span>
        <span class="num-den">{{ centerDenominator }}</span>
      </div>
      <div class="center-pct">{{ progress }}<span class="pct-unit">%</span></div>
      <div class="center-hint">{{ centerHint }}</div>
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
  centerCaption: string
  centerHint: string
  size?: number
}>()

const size = computed(() => props.size ?? 520)
const cx = computed(() => size.value / 2)
const cy = computed(() => size.value / 2)

// 半径定义
const innerR = computed(() => size.value * 0.18)
const midR = computed(() => size.value * 0.30)   // 内圈标签
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
    // 标签偏移
    const lx = Math.cos(angle) * 80
    const ly = Math.sin(angle) * 22
    items.push({ x, y, lx, ly, angle })
  }
  return items
})

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

// 内圈 6 个标签（径向）
const INNER_LABELS_DEFAULT = [
  '演练复盘与行动',
  '演练启动与人员',
  '基线指标与备份',
  '演练启动与人员',
  '故障定位于影响',
  '业务验收与告警',
]
const innerLabels = computed(() => {
  // 6 个标签从 6 个不同角度摆放（顶部 12 点起算）
  return INNER_LABELS_DEFAULT.map((t, i) => {
    let size = 11
    let weight = 400
    let color = 'rgba(110, 141, 181, 0.85)'
    // 当前阶段高亮（这里简单按数组中含 业务/复盘 的视觉）
    if (t === '业务验收与告警' || t === '基线指标与备份') {
      size = 12
      weight = 600
      color = '#d6e8ff'
    }
    return { text: t, size, weight, color }
  })
})

const innerLabelPoints = computed(() => {
  // 6 个标签的角度（12 点起，顺时针 60° 一格）
  const angles = [-90, -30, 30, 90, 150, 210].map(a => a * Math.PI / 180)
  return angles.map(a => ({
    x: cx.value + midR.value * Math.cos(a),
    y: cy.value + midR.value * Math.sin(a),
  }))
})

function chineseNum(n: number): string {
  return ['一', '二', '三', '四', '五', '六'][n - 1] || String(n)
}
</script>

<style lang="scss" scoped>
.phase-ring {
  position: relative;
  display: flex; align-items: center; justify-content: center;
  font-family: 'Microsoft YaHei', 'PingFang SC', sans-serif;
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
  .phase-label-tag { margin-left: -18px; transform: translateX(-100%); }
  &::before { right: -16px; left: auto; background: linear-gradient(90deg, rgba(0, 212, 255, 0.4), transparent); }
}
.phase-label-3 {
  // 左上
  .phase-label-tag { margin-left: -18px; transform: translateX(-100%); }
  &::before { right: -16px; left: auto; background: linear-gradient(90deg, rgba(0, 212, 255, 0.4), transparent); }
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
  font-size: 13px;
  color: #c7dcff;
  letter-spacing: 3px;
  font-family: 'Orbitron', 'Rajdhani', sans-serif;
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
