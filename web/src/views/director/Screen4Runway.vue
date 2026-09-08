<template>
  <div class="screen4-runway" :class="{ 'is-complete': phaseComplete }">
    <div class="runway-ambient" aria-hidden="true"></div>
    <div class="runway-summary">
      <span class="summary-kicker">
        <i></i>
        {{ phaseName }}
      </span>
      <span class="summary-progress">
        <strong>{{ progress.completed }}</strong>
        <span>/ {{ progress.total }} 环节</span>
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
          <stop offset="0" stop-color="#2af0aa" />
          <stop offset=".55" stop-color="#5df7bd" />
          <stop offset="1" stop-color="#20d991" />
        </linearGradient>
        <linearGradient id="s4-lane-active" x1="0" y1="0" x2="1" y2="0">
          <stop offset="0" stop-color="#ffb539" stop-opacity=".25" />
          <stop offset=".48" stop-color="#ffd273" />
          <stop offset="1" stop-color="#ff9f1a" stop-opacity=".2" />
        </linearGradient>
        <radialGradient id="s4-milestone-core">
          <stop offset="0" stop-color="#f8ffff" />
          <stop offset=".35" stop-color="#61f5ff" />
          <stop offset="1" stop-color="#1277bd" />
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
      <path v-if="completedPath" class="runway-complete-path" :d="completedPath" />
      <path v-if="activePath" class="runway-active-path" :d="activePath" />

      <g
        v-for="indicator in turnIndicators"
        :key="`turn-${indicator.row}`"
        class="turn-indicator"
        :transform="`translate(${indicator.x} ${indicator.y})`"
        aria-hidden="true"
      >
        <circle r="18" />
        <path :d="indicator.direction === 'right' ? 'M -7 -6 L 2 0 L -7 6 M 1 -6 L 10 0 L 1 6' : 'M 7 -6 L -2 0 L 7 6 M -1 -6 L -10 0 L -1 6'" />
      </g>

      <g
        v-for="item in visualItems"
        :key="item.key"
        :class="item.className"
        :transform="`translate(${item.point.x} ${item.point.y})`"
      >
        <template v-if="item.kind === 'node'">
          <circle class="node-orbit node-orbit-outer" r="38" />
          <circle class="node-orbit node-orbit-inner" r="29" />
          <circle class="node-halo" r="25" />
          <circle class="node-core" r="17" />
          <path v-if="item.node.status === 'done' || item.node.status === 'skipped'" class="node-check" d="M -7 0 L -2 6 L 9 -7" />
          <path v-else-if="item.node.status === 'issue'" class="node-issue-mark" d="M 0 -8 L 0 3 M 0 9 L 0 10" />
          <g v-else-if="item.node.status === 'running'" class="node-energy" aria-hidden="true">
            <rect v-for="bar in 3" :key="bar" :x="-10 + (bar - 1) * 8" :y="3 - bar * 4" width="5" :height="bar * 5" rx="2" />
          </g>
          <text class="runway-node-index" y="-48">{{ String(item.index + 1).padStart(2, '0') }}</text>
          <text class="runway-node-name" y="49">
            <tspan
              v-for="(line, lineIndex) in item.labelLines"
              :key="line"
              x="0"
              :dy="lineIndex === 0 ? 0 : 20"
            >{{ line }}</tspan>
          </text>
          <g class="runway-node-count" :transform="`translate(0 ${item.labelLines.length > 1 ? 91 : 72})`">
            <rect x="-39" y="-13" width="78" height="26" rx="13" />
            <text y="5">{{ item.node.completed }}/{{ item.node.total }}</text>
          </g>
        </template>

        <template v-else>
          <circle class="milestone-scan" r="43" />
          <circle class="milestone-ring milestone-ring-outer" r="34" />
          <circle class="milestone-ring milestone-ring-inner" r="23" />
          <circle class="milestone-core" r="10" />
          <path class="milestone-cross" d="M -34 0 L 34 0 M 0 -34 L 0 34" />
          <text class="milestone-title" y="57">里程碑</text>
          <text class="milestone-status" y="79">{{ phaseComplete ? '阶段达成' : '等待到达' }}</text>
        </template>
      </g>

      <g
        v-if="activePoint"
        class="runway-baton"
        :transform="`translate(${activePoint.x} ${activePoint.y}) rotate(${activePoint.direction === 'right' ? 0 : 180})`"
        aria-hidden="true"
      >
        <path class="baton-beam" d="M 31 0 L 67 0" />
        <circle class="baton-tail" cx="33" cy="0" r="4" />
        <path class="baton-body" d="M 43 -7 L 68 -7 L 79 0 L 68 7 L 43 7 Z" />
        <circle class="baton-spark" cx="78" cy="0" r="5" />
      </g>
    </svg>

    <div class="runway-legend" aria-label="跑道状态图例">
      <span><i class="is-done"></i>已完成</span>
      <span><i class="is-running"></i>进行中</span>
      <span><i class="is-pending"></i>待执行</span>
      <span><i class="is-issue"></i>异常</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  buildScreen4RunwayLayout,
  getScreen4RunwayPath,
  getScreen4RunwayProgress,
  splitScreen4RunwayName,
  type Screen4RunwayNode,
  type Screen4RunwayPoint,
} from './screen4Runway'

const props = defineProps<{
  phaseName: string
  phaseStatus: string
  nodes: Screen4RunwayNode[]
}>()

const terminalStatuses = new Set(['done', 'skipped', 'issue'])

const layout = computed(() => buildScreen4RunwayLayout(props.nodes.length))
const progress = computed(() => getScreen4RunwayProgress(props.nodes.map(node => node.status)))
const phaseComplete = computed(() => {
  const normalizedStatus = props.phaseStatus.toLowerCase()
  return normalizedStatus === 'completed'
    || normalizedStatus === 'done'
    || (props.nodes.length > 0 && progress.value.completed === progress.value.total)
})

const completedEndIndex = computed(() => {
  let endIndex = -1
  for (const [index, node] of props.nodes.entries()) {
    if (!terminalStatuses.has(node.status)) break
    endIndex = index
  }
  return endIndex
})

const completedPath = computed(() => (
  completedEndIndex.value > 0
    ? getScreen4RunwayPath(layout.value.points, 0, completedEndIndex.value)
    : ''
))

const activeIndex = computed(() => props.nodes.findIndex(node => node.status === 'running'))
const activePoint = computed<Screen4RunwayPoint | null>(() => (
  activeIndex.value >= 0 ? layout.value.points[activeIndex.value] || null : null
))
const activePath = computed(() => {
  if (activeIndex.value < 0 || activeIndex.value >= layout.value.points.length - 1) return ''
  return getScreen4RunwayPath(layout.value.points, activeIndex.value, activeIndex.value + 1)
})

const visualItems = computed(() => {
  const nodeItems = props.nodes.map((node, index) => ({
    key: node.id,
    kind: 'node' as const,
    node,
    index,
    point: layout.value.points[index],
    labelLines: splitScreen4RunwayName(node.name),
    className: [
      'runway-node',
      terminalStatuses.has(node.status) && node.status !== 'issue' ? 'is-completed' : '',
      node.status === 'running' ? 'is-running' : '',
      node.status === 'issue' ? 'is-issue' : '',
      node.status === 'pending' ? 'is-pending' : '',
    ],
  }))
  const milestonePoint = layout.value.points[props.nodes.length]

  return [
    ...nodeItems,
    {
      key: 'screen4-runway-milestone',
      kind: 'milestone' as const,
      point: milestonePoint,
      className: ['runway-milestone', phaseComplete.value ? 'is-completed' : 'is-pending'],
    },
  ]
})

const turnIndicators = computed(() => layout.value.points.flatMap((point, index, points) => {
  const nextPoint = points[index + 1]
  if (!nextPoint || nextPoint.row === point.row) return []

  return [{
    row: point.row,
    x: point.direction === 'right' ? 998 : 42,
    y: (point.y + nextPoint.y) / 2,
    direction: nextPoint.direction,
  }]
}))
</script>

<style scoped lang="scss">
.screen4-runway {
  position: relative;
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
  isolation: isolate;
  border-radius: 18px;
  border: 1px solid rgba(65, 188, 238, 0.22);
  background:
    radial-gradient(circle at 50% 46%, rgba(15, 111, 157, 0.16), transparent 42%),
    linear-gradient(180deg, rgba(1, 18, 34, 0.2), rgba(0, 13, 27, 0.52));
  box-shadow: inset 0 0 42px rgba(0, 6, 18, 0.42);
}

.runway-ambient {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background:
    linear-gradient(90deg, transparent 49.8%, rgba(83, 218, 255, 0.05) 50%, transparent 50.2%),
    linear-gradient(0deg, transparent 49.8%, rgba(83, 218, 255, 0.04) 50%, transparent 50.2%);
  background-size: 56px 56px;
  mask-image: linear-gradient(to bottom, transparent, #000 16%, #000 84%, transparent);
}

.runway-summary {
  position: absolute;
  z-index: 2;
  top: 14px;
  right: 18px;
  display: flex;
  align-items: center;
  gap: 14px;
  min-width: 220px;
  height: 42px;
  padding: 0 14px;
  border: 1px solid rgba(67, 212, 255, 0.28);
  border-radius: 8px;
  color: #bfeaff;
  background: linear-gradient(110deg, rgba(4, 37, 61, 0.9), rgba(3, 26, 47, 0.64));
  box-shadow: inset 3px 0 0 #30ddb2, 0 10px 30px rgba(0, 7, 18, 0.28);
  backdrop-filter: blur(8px);
}

.summary-kicker {
  display: flex;
  align-items: center;
  gap: 7px;
  max-width: 130px;
  overflow: hidden;
  font-size: 14px;
  white-space: nowrap;
  text-overflow: ellipsis;

  i {
    flex: 0 0 auto;
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: #37ecb0;
    box-shadow: 0 0 10px #37ecb0;
  }
}

.summary-progress {
  display: flex;
  align-items: baseline;
  gap: 4px;
  padding-left: 13px;
  border-left: 1px solid rgba(83, 210, 255, 0.18);
  font-family: 'DIN Alternate', 'Arial Narrow', sans-serif;

  strong {
    color: #f4fcff;
    font-size: 22px;
    line-height: 1;
  }

  span {
    color: #75aac4;
    font-size: 12px;
  }
}

.runway-svg {
  position: absolute;
  inset: 8px 10px 20px;
  width: calc(100% - 20px);
  height: calc(100% - 28px);
  overflow: visible;
  font-family: 'Microsoft YaHei', 'PingFang SC', sans-serif;
}

.runway-grid {
  fill: url('#s4-runway-grid');
  stroke: rgba(68, 195, 239, 0.08);
}

.runway-shadow,
.runway-rail,
.runway-dashes,
.runway-complete-path,
.runway-active-path {
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
  stroke-width: 9;
  filter: url('#s4-soft-glow');
  animation: path-arrive .8s ease both;
}

.runway-active-path {
  stroke: url('#s4-lane-active');
  stroke-width: 9;
  stroke-dasharray: 26 16;
  filter: url('#s4-strong-glow');
  animation: active-energy 1.15s linear infinite;
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
    filter: url('#s4-soft-glow');
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

.runway-node-index {
  fill: currentColor;
  font-family: 'DIN Alternate', 'Arial Narrow', sans-serif;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 2px;
  text-anchor: middle;
  opacity: .84;
}

.runway-node-name {
  fill: #d8f3ff;
  font-size: 16px;
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
    font-family: 'DIN Alternate', 'Arial Narrow', sans-serif;
    font-size: 14px;
    font-weight: 700;
    letter-spacing: 1px;
    text-anchor: middle;
  }
}

.runway-milestone {
  color: #49dff5;

  &.is-completed {
    color: #39e8aa;

    .milestone-scan {
      fill: rgba(50, 228, 166, .16);
    }
  }
}

.milestone-scan {
  fill: rgba(52, 198, 238, .12);
  stroke: currentColor;
  stroke-width: 1;
  stroke-dasharray: 5 5;
  filter: url('#s4-soft-glow');
  animation: milestone-scan 6s linear infinite;
  transform-box: fill-box;
  transform-origin: center;
}

.milestone-ring {
  fill: rgba(2, 22, 40, .88);
  stroke: currentColor;
  filter: url('#s4-soft-glow');
}

.milestone-ring-outer { stroke-width: 2; }
.milestone-ring-inner { stroke-width: 1; opacity: .65; }

.milestone-core {
  fill: url('#s4-milestone-core');
  filter: url('#s4-strong-glow');
}

.milestone-cross {
  fill: none;
  stroke: currentColor;
  stroke-width: 1;
  opacity: .5;
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

.runway-baton {
  color: #ffc55f;
  filter: url('#s4-strong-glow');
  animation: baton-hover 1.5s ease-in-out infinite;
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

.runway-legend {
  position: absolute;
  left: 20px;
  bottom: 12px;
  z-index: 2;
  display: flex;
  align-items: center;
  gap: 17px;
  color: #6390a8;
  font-size: 11px;
  letter-spacing: .5px;

  span {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }

  i {
    width: 6px;
    height: 6px;
    border-radius: 50%;
    box-shadow: 0 0 7px currentColor;
  }

  .is-done { color: #38e7a7; background: currentColor; }
  .is-running { color: #ffb43d; background: currentColor; }
  .is-pending { color: #4e98c8; background: currentColor; }
  .is-issue { color: #ff626e; background: currentColor; }
}

@keyframes runway-drift {
  to { stroke-dashoffset: -30; }
}

@keyframes active-energy {
  to { stroke-dashoffset: -42; }
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

@keyframes energy-bar {
  from { transform: scaleY(.48); opacity: .58; }
  to { transform: scaleY(1); opacity: 1; }
}

@keyframes milestone-scan {
  to { transform: rotate(360deg); }
}

@keyframes baton-hover {
  0%, 100% { opacity: .62; }
  50% { opacity: 1; }
}

@media (max-width: 1280px) {
  .runway-summary {
    min-width: 190px;
    transform: scale(.9);
    transform-origin: top right;
  }

  .runway-legend {
    gap: 11px;
    transform: scale(.9);
    transform-origin: bottom left;
  }
}

@media (prefers-reduced-motion: reduce) {
  .runway-dashes,
  .runway-active-path,
  .runway-node,
  .node-orbit-outer,
  .node-energy rect,
  .milestone-scan,
  .runway-baton,
  .baton-beam {
    animation: none !important;
  }
}
</style>
