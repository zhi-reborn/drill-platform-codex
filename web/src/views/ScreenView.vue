<template>
  <div ref="screenRootRef" class="screen-root">
    <!-- Background layers -->
    <div class="bg-grid" />
    <div class="bg-scan" />
    <div class="bg-vignette" />

    <!-- 漂浮微光粒子 -->
    <div class="bg-particles">
      <div v-for="i in 4" :key="'bp' + i" class="bg-particle" :class="'bp-' + i" />
    </div>

    <!-- Loading state -->
    <div v-if="loading" class="overlay-state">
      <div class="loader">
        <div class="loader-ring" />
        <p class="loader-text">正在连接演练数据...</p>
      </div>
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="overlay-state error">
      <div class="error-content">
        <el-icon :size="48"><CircleClose /></el-icon>
        <p>{{ error }}</p>
        <el-button type="primary" @click="handleRetry">重试</el-button>
      </div>
    </div>

    <!-- Main content -->
    <div v-else-if="currentDrill" class="screen-content" :class="{ 'screen-content-fallback-fullscreen': fallbackFullscreen }">
      <!-- ========== HEADER ========== -->
      <header class="screen-header">
        <svg class="header-frame" viewBox="0 0 1200 82" preserveAspectRatio="none" aria-hidden="true">
          <defs>
            <linearGradient id="header-line-grad" x1="0%" y1="0%" x2="100%" y2="0%">
              <stop offset="0%" stop-color="#148cff" stop-opacity="0" />
              <stop offset="12%" stop-color="#24d7ff" stop-opacity="0.9" />
              <stop offset="34%" stop-color="#42efff" stop-opacity="1" />
              <stop offset="50%" stop-color="#d7fbff" stop-opacity="1" />
              <stop offset="66%" stop-color="#42efff" stop-opacity="1" />
              <stop offset="88%" stop-color="#24d7ff" stop-opacity="0.9" />
              <stop offset="100%" stop-color="#148cff" stop-opacity="0" />
            </linearGradient>
            <linearGradient id="header-flow-grad" x1="0%" y1="0%" x2="100%" y2="0%">
              <stop offset="0%" stop-color="#ffffff" stop-opacity="0" />
              <stop offset="42%" stop-color="#9bf8ff" stop-opacity="0.1" />
              <stop offset="50%" stop-color="#ffffff" stop-opacity="1" />
              <stop offset="58%" stop-color="#9bf8ff" stop-opacity="0.1" />
              <stop offset="100%" stop-color="#ffffff" stop-opacity="0" />
            </linearGradient>
            <filter id="header-line-glow" x="-8%" y="-130%" width="116%" height="360%">
              <feGaussianBlur stdDeviation="4.2" result="blur" />
              <feMerge>
                <feMergeNode in="blur" />
                <feMergeNode in="SourceGraphic" />
              </feMerge>
            </filter>
          </defs>
          <path class="header-frame-line header-frame-line-shadow" d="M26 18 H156 L178 36 H402 L422 70 H462" />
          <path class="header-frame-line header-frame-line-shadow" d="M738 70 H778 L798 36 H1018 L1040 18 H1174" />
          <path class="header-frame-line" d="M26 18 H156 L178 36 H402 L422 70 H462" />
          <path class="header-frame-line" d="M738 70 H778 L798 36 H1018 L1040 18 H1174" />
          <path class="header-frame-tray" d="M462 70 H738" />
          <path class="header-frame-lower-tray header-frame-lower-tray-glow" d="M400 34 L417 64 H783 L800 34" />
          <path class="header-frame-lower-tray" d="M410 34 L426 62 H774 L790 34" />
          <path class="header-frame-tip header-frame-tip-glow" d="M422 70 H462" />
          <path class="header-frame-tip header-frame-tip-glow" d="M738 70 H778" />
          <path class="header-frame-tip-soft header-frame-tip-soft-glow" d="M410 34 L426 62" />
          <path class="header-frame-tip-soft header-frame-tip-soft-glow" d="M774 62 L790 34" />
          <path class="header-frame-tip" d="M422 70 H462" />
          <path class="header-frame-tip" d="M738 70 H778" />
          <path class="header-frame-tip-soft" d="M410 34 L426 62" />
          <path class="header-frame-tip-soft" d="M774 62 L790 34" />
          <path class="header-flow header-flow-left header-flow-delay-1" d="M410 34 L426 62 H600" />
          <path class="header-flow header-flow-right header-flow-delay-1" d="M790 34 L774 62 H600" />
          <path class="header-flow header-flow-left header-flow-delay-2" d="M422 70 H600" />
          <path class="header-flow header-flow-right header-flow-delay-2" d="M778 70 H600" />
          <path class="header-flow header-flow-left header-flow-delay-3" d="M462 70 H600" />
          <path class="header-flow header-flow-right header-flow-delay-3" d="M738 70 H600" />
        </svg>
        <div class="header-title-block">
          <h1 class="drill-title">应急指挥中心</h1>
        </div>
        <div class="header-meta">
        </div>
        <button
          class="btn-icon btn-fullscreen-mark"
          :class="{ active: isFullscreenLike }"
          @click="toggleFullscreen"
          :title="isFullscreenLike ? '退出全屏' : '全屏切换'"
          :aria-label="isFullscreenLike ? '退出全屏' : '进入全屏'"
        >
          <img
            v-if="screenBrand.fullscreenIconImage"
            class="fullscreen-mark fullscreen-mark-image"
            :src="screenBrand.fullscreenIconImage"
            alt=""
            aria-hidden="true"
            @error="handleBrandIconError"
          />
          <span v-else class="fullscreen-mark" aria-hidden="true">{{ screenBrand.fullscreenIconText }}</span>
        </button>
        <div class="header-pulse-line" />
        <span class="header-brand">{{ screenBrand.companyName }}</span>
      </header>

      <!-- ========== TOP KPI ROW ========== -->
      <section class="kpi-row">
        <div class="kpi-card">
          <span class="kpi-orb" />
          <div class="kpi-copy">
            <span class="kpi-label-zh">演练状态</span>
          </div>
          <div class="kpi-value-row kpi-status-row">
            <span class="status-dot" :class="'dot-' + (currentDrill.status || '')" />
            <span class="kpi-value-text" :class="'txt-' + (currentDrill.status || '')">{{ drillStatusLabel }}</span>
          </div>
        </div>

        <div class="kpi-card kpi-progress-card">
          <span class="kpi-orb" />
          <div class="kpi-copy">
            <span class="kpi-label-zh">整体进度</span>
          </div>
          <div class="kpi-value-row kpi-progress-row">
            <div class="progress-ring-wrap">
              <svg class="progress-ring-svg" viewBox="0 0 100 100">
                <defs>
                  <linearGradient id="ring-grad" x1="0%" y1="0%" x2="100%" y2="100%">
                    <stop offset="0%" stop-color="#fff" stop-opacity="0.95" />
                    <stop offset="60%" stop-color="#2cf8d8" stop-opacity="1" />
                    <stop offset="100%" stop-color="#00d4aa" stop-opacity="1" />
                  </linearGradient>
                </defs>
                <circle class="progress-ring-bg" cx="50" cy="50" r="46" />
                <circle class="progress-ring-fill"
                  cx="50" cy="50" r="46"
                  :style="{ strokeDashoffset: 289.03 * (1 - progressPercent / 100) }" />
              </svg>
              <span class="progress-ring-text">{{ progressPercent }}<small>%</small></span>
            </div>
            <div class="progress-node-block">
              <div class="node-count-row">
                <span class="kpi-value-num">{{ completedCount }}</span>
                <span class="node-separator">/</span>
                <span class="node-total">{{ totalCount }}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="kpi-card kpi-alert-card" :class="'alert-level-' + alertLevel">
          <span class="kpi-orb kpi-orb-alert" :class="'orb-alert-' + alertLevel"></span>
          <div class="kpi-copy">
            <span class="kpi-label-zh">异常任务</span>
          </div>
          <div class="kpi-value-row kpi-alert-row">
            <div class="alert-value-block">
              <span class="kpi-value-num alert-num" :class="'num-' + alertLevel">{{ abnormalTaskCount }}</span>
            </div>
          </div>
          <!-- 异常任务详情提示 -->
          <div v-if="abnormalTaskCount > 0" class="alert-detail-hint">
            <span class="hint-dot"></span>
            <span class="hint-text">查看步骤列表获取详情</span>
          </div>
        </div>

        <div class="kpi-card">
          <span class="kpi-orb" />
          <div class="kpi-copy">
            <span class="kpi-label-zh">当前阶段</span>
          </div>
          <div class="kpi-value-row">
            <span class="kpi-value-text">阶段{{ chineseNum(currentPhaseIndex + 1) }}</span>
          </div>
        </div>
      </section>

      <!-- ========== MAIN GRID ========== -->
      <main class="screen-main">
        <!-- LEFT: Stage overview -->
        <section class="panel panel-stages">
          <div class="panel-header">
            <span class="panel-deco-corner tl" />
            <span class="panel-deco-corner tr" />
            <span class="panel-title-zh">演练阶段总览</span>
            <span class="panel-badge">{{ stages.length }}</span>
            <div class="panel-scan-line" />
          </div>
          <div class="panel-body stages-list">
            <div
              v-for="(stage, idx) in stages"
              :key="idx"
              class="stage-card"
              :class="['stage-' + stage.status, { 'stage-current': idx === displayPhaseIndex }]"
              role="tab"
              :aria-selected="idx === displayPhaseIndex"
              tabindex="0"
              @click="selectPhase(idx)"
              @keydown.enter.prevent="selectPhase(idx)"
              @keydown.space.prevent="selectPhase(idx)"
            >
              <div class="stage-card-top">
                <div class="stage-name-block">
                  <span class="stage-name">{{ stage.name }}</span>
                </div>
                <span class="stage-badge" :class="'badge-' + stage.status">
                  {{ stageBadgeLabel(stage.status) }}
                </span>
              </div>
              <div class="stage-segments">
                <span
                  v-for="(seg, si) in stage.segments"
                  :key="si"
                  class="segment"
                  :class="'seg-' + seg"
                />
              </div>
              <div class="stage-card-bottom">
                <span class="stage-meta">
                  <span class="meta-key">环节</span>
                  <span class="meta-val">{{ stage.completedPhases }} / {{ stage.totalPhases }}</span>
                </span>
                <span class="stage-meta">
                  <span class="meta-key">步骤</span>
                  <span class="meta-val">{{ stage.completedSteps }} / {{ stage.totalSteps }}</span>
                </span>
              </div>
            </div>
          </div>
        </section>

        <!-- CENTER: Phase ring -->
        <section ref="centerPanelRef" class="panel panel-center">
          <div class="center-stage" :class="`center-stage-${displayPhaseIndex}`">
            <PhaseRing
              :phases="ringPhases"
              :phase-names="ringPhaseNames"
              :phase-node-statuses="ringPhaseNodeStatuses"
              :phase-statuses="ringPhaseStatuses"
              :current-index="displayPhaseIndex"
              :progress="displayPhaseProgress"
              :center-numerator="displayPhaseStepNum"
              :center-denominator="displayPhaseStepDen"
              :center-hint="`阶段${chineseNum(displayPhaseIndex + 1)} · ${displayPhaseName}`"
              :size="ringSize"
              :fullscreen="isFullscreenLike"
            />
          </div>
        </section>

        <!-- RIGHT: Active steps + Exec log -->
        <section class="panel panel-right">
          <div class="sub-panel sub-warn">
            <div class="panel-header">
              <span class="panel-deco-corner tl" />
              <span class="panel-deco-corner tr" />
              <span class="panel-title-zh">当前环节待完成的任务</span>
              <span class="panel-realtime">
                <span class="rt-dot" />
                实时
              </span>
              <div class="panel-scan-line" />
            </div>
            <div class="panel-body warn-list">
              <!-- 上部：执行中步骤卡片（自适应可见数量） -->
              <div class="alert-scroll" ref="warnListRef" :style="{
                '--visible-alert-count': Math.max(visibleAlerts.length, 1),
                '--alert-card-gap': `${alertCardGap}px`,
              }">
                <div
                  v-for="(alert, ai) in visibleAlerts"
                  :key="alert.stepId"
                  class="alert-card"
                  :class="'alert-' + alert.level"
                  :data-step-id="alert.stepId"
                  :ref="el => setAlertCardRef(el, ai)"
                >
                  <!-- 顶部：状态指示条 + 标题行 -->
                  <div class="alert-head">
                    <span class="alert-indicator" />
                    <span class="alert-title">{{ alert.title }}</span>
                    <span class="alert-status-badge" :class="'badge-' + alert.level">{{ alert.statusLabel }}</span>
                  </div>
                  <!-- 元数据行 -->
                  <div class="alert-foot">
                    <span v-if="alert.operator" class="alert-meta">
                      <span class="meta-icon">◈</span>
                      <span class="meta-label">操作人</span>
                      <span class="meta-val operator-val">{{ alert.operator }}</span>
                    </span>
                  </div>
                  <!-- 剩余时间进度条：默认超时 10 分钟，仅大屏展示，不参与流程流转；
                       待执行任务不计时，显示"未开始"占位 -->
                  <div
                    class="alert-countdown"
                    :class="{
                      'is-urgent': alert.startedAt && alertCountdownPercent(alert) >= 100,
                      'is-idle': !alert.startedAt,
                    }"
                  >
                    <div class="countdown-bar">
                      <div
                        v-if="alert.startedAt"
                        class="countdown-fill"
                        :style="{ width: Math.min(alertCountdownPercent(alert), 100) + '%' }"
                      />
                    </div>
                    <span class="countdown-text">{{ alert.startedAt ? alertCountdownText(alert) : '未开始' }}</span>
                  </div>
                </div>
                <div v-if="activeAlerts.length === 0" class="empty-tip">当前环节任务已全部完成</div>
                <div v-else-if="visibleAlerts.length < activeAlerts.length" ref="moreTipRef" class="more-tip">
                  还有 {{ activeAlerts.length - visibleAlerts.length }} 个任务...
                </div>
              </div>
            </div>
          </div>
          <!-- 下部：执行日志流（独立面板，标题与"执行中步骤"同款） -->
          <div class="sub-panel sub-log">
            <div class="panel-header">
              <span class="panel-deco-corner tl" />
              <span class="panel-deco-corner tr" />
              <span class="panel-title-zh">执行日志</span>
              <span class="panel-realtime">
                <span class="rt-dot" />
                实时
              </span>
              <div class="panel-scan-line" />
            </div>
            <div class="panel-body exec-log">
              <div class="exec-log-view" ref="execLogRef">
                <!-- 贴底驻留模式：最新日志始终在最下面，产生新日志时整体上移一条 -->
                <div class="exec-log-track">
                  <div
                    v-for="(log, li) in execLogLines"
                    :key="`${log.id}-${li}`"
                    class="log-line"
                    :class="['log-' + logActionClass(log.action), { 'is-newest': li === execLogLines.length - 1 }]"
                  >
                    <span class="log-time">{{ formatTime(log.created_at) }}</span>
                    <span class="log-step">{{ resolveStepName(log) }}</span>
                    <span class="log-action">{{ logActionLabel(log.action) }}</span>
                  </div>
                </div>
                <div v-if="!execLogLines.length" class="log-empty">暂无执行日志</div>
              </div>
            </div>
          </div>
        </section>
      </main>

      <!-- Footer decorations -->
      <footer class="screen-footer" />
    </div>

    <!-- 任务完成流式动画层：飞行的任务卡片 ghost + 汇聚粒子 + 圆环吸收特效 -->
    <div ref="flyLayerRef" class="task-fly-layer" aria-hidden="true"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onBeforeUnmount, watch } from 'vue'
import type { ComponentPublicInstance } from 'vue'
import { useRoute } from 'vue-router'
import { CircleClose } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import type { StepInstance, StepInstanceLog, DrillInstance, StepStatus, DrillStatus } from '@/types/instance'
import { drillApi } from '@/api/modules/drill'
import { useAuthStore } from '@/stores/auth'
import PhaseRing from '@/components/screen/PhaseRing.vue'

const route = useRoute()
const screenRootRef = ref<HTMLElement | null>(null)
const fallbackFullscreen = ref(false)
const isNativeFullscreen = ref(false)
const isFullscreenLike = computed(() => isNativeFullscreen.value || fallbackFullscreen.value)
type ScreenBrandConfig = {
  companyName?: string
  fullscreenIconText?: string
  fullscreenIconImage?: string
}

const DEFAULT_SCREEN_BRAND = {
  companyName: '东风科技有限公司',
  fullscreenIconText: 'N',
  fullscreenIconImage: '',
}

const screenBrand = ref({ ...DEFAULT_SCREEN_BRAND })
const loading = ref(true)
const error = ref<string | null>(null)
const viewportWidth = ref(window.innerWidth)
const viewportHeight = ref(window.innerHeight)
const centerPanelRef = ref<HTMLElement | null>(null)
const centerPanelWidth = ref(0)
const centerPanelHeight = ref(0)
let centerPanelResizeObserver: ResizeObserver | null = null

let ws: WebSocket | null = null
let refreshTimer: number | null = null
let dataRefreshTimer: number | null = null
let dataLoading = false
let dataReloadQueued = false
let timeTimer: number | null = null
let componentDestroyed = false

// 顶部时间
const systemTime = ref(formatSystemTime(new Date()))

// 当前路由 drill id
const drillId = computed(() => {
  const id = route.params.id
  return typeof id === 'string' ? parseInt(id, 10) : null
})

// 数据
const currentDrill = ref<DrillInstance | null>(null)
const drillSteps = ref<StepInstance[]>([])
const recentLogs = ref<StepInstanceLog[]>([])
const warnListRef = ref<HTMLElement | null>(null)
const alertCardRefs = ref<HTMLElement[]>([])
const execLogRef = ref<HTMLElement | null>(null)
const execLogViewHeight = ref(96)
const EXEC_LOG_LINE_HEIGHT = 26
const moreTipRef = ref<HTMLElement | null>(null)

// 任务完成流式动画
const flyLayerRef = ref<HTMLElement | null>(null)

// === 树形步骤辅助 ===
// 构建父子映射，支持任意深度嵌套（阶段→环节→任务→操作步骤）
const childMap = computed(() => {
  const map = new Map<number, StepInstance[]>()
  for (const s of drillSteps.value) {
    const pid = s.parent_step_id
    if (pid) {
      const list = map.get(pid) || []
      list.push(s)
      map.set(pid, list)
    }
  }
  return map
})

// 递归收集某步骤子树中的所有叶子节点
function collectLeaves(stepId: number): StepInstance[] {
  const children = childMap.value.get(stepId)
  if (!children || children.length === 0) {
    // 自身就是叶子
    const self = drillSteps.value.find(s => s.id === stepId)
    return self ? [self] : []
  }
  const leaves: StepInstance[] = []
  for (const c of children) {
    leaves.push(...collectLeaves(c.id))
  }
  return leaves
}

// 叶子步骤：无子节点的步骤（实际执行的操作步骤）
const leafSteps = computed(() => {
  const allSteps = drillSteps.value
  if (allSteps.length === 0) return []
  const hasChild = new Set<number>()
  for (const s of allSteps) {
    if (s.parent_step_id) hasChild.add(s.parent_step_id)
  }
  const leaves = allSteps.filter(s => !hasChild.has(s.id))
  return leaves.length > 0 ? leaves : allSteps
})

// 根步骤（阶段）
const rootSteps = computed(() => {
  const allSteps = drillSteps.value
  const hasParent = new Set<number>()
  for (const s of allSteps) {
    if (s.parent_step_id) hasParent.add(s.id)
  }
  // 没有 parent_step_id 的就是根
  return allSteps.filter(s => !s.parent_step_id).sort((a, b) => a.seq - b.seq)
})

// 从叶子步骤向上追溯，找到所属环节（根步骤的直接子节点）
function findParentPhase(stepId: number): string {
  const stepMap = new Map<number, StepInstance>()
  for (const s of drillSteps.value) stepMap.set(s.id, s)
  const rootIds = new Set(rootSteps.value.map(r => r.id))
  let cur = stepMap.get(stepId)
  let last = cur
  while (cur && !rootIds.has(cur.id)) {
    last = cur
    cur = cur.parent_step_id ? stepMap.get(cur.parent_step_id) : undefined
  }
  // last 是根的直接子节点（环节），cur 是根节点本身
  if (last && last.id !== stepId) return last.name
  // 叶子本身就是根步骤
  return cur?.name || '--'
}

// 从叶子步骤向上追溯，找到直接父节点名称（任务）
function findDirectParent(stepId: number): string {
  const stepMap = new Map<number, StepInstance>()
  for (const s of drillSteps.value) stepMap.set(s.id, s)
  const step = stepMap.get(stepId)
  if (!step?.parent_step_id) return '--'
  const parent = stepMap.get(step.parent_step_id)
  return parent?.name || '--'
}

function detectNewlyCompletedLeafSteps(previousStatuses: Map<number, StepStatus>, steps: StepInstance[]): number[] {
  if (previousStatuses.size === 0) return []

  const hasChild = new Set<number>()
  for (const step of steps) {
    if (step.parent_step_id) hasChild.add(step.parent_step_id)
  }

  return steps
    .filter(step => !hasChild.has(step.id))
    .filter(step => step.status === 'completed' && previousStatuses.get(step.id) !== 'completed')
    .map(step => step.id)
}

// 任务完成流式动画：对应任务卡片 ghost 从右侧飞向中心百分数圆环，
// 营造"信息汇聚、流入圆环汇总"的视觉效果。
// 调用时机：applyStepEvent 在更新 drillSteps 之后同步调用本函数，
// 此时 Vue 尚未 patch DOM，原 alert-card 仍在原位，可精确捕获其位置。
function playTaskFlowAnimation(stepId: number) {
  const step = drillSteps.value.find(s => s.id === stepId)
  if (!step) return
  const screenRoot = screenRootRef.value
  const flyLayer = flyLayerRef.value
  if (!screenRoot || !flyLayer) return

  // 圆环汇流点（hub-core）
  const hubCore = screenRoot.querySelector<HTMLElement>('.progress-hub .hub-core')
  if (!hubCore) return
  const hubGlow = screenRoot.querySelector<HTMLElement>('.progress-hub .hub-glow')

  const screenRect = screenRoot.getBoundingClientRect()
  const hubRect = hubCore.getBoundingClientRect()
  const hubCx = hubRect.left + hubRect.width / 2 - screenRect.left
  const hubCy = hubRect.top + hubRect.height / 2 - screenRect.top

  // 任务卡片元素（DOM 尚未 patch，原卡片仍在）
  const cardEl = screenRoot.querySelector<HTMLElement>(`[data-step-id="${stepId}"]`)
  if (!cardEl) {
    // 卡片不在可见区域（已滚出 / 属于其他阶段 / 父级步骤无卡片）：
    // 仅触发圆环脉冲反馈，不生成飞行 ghost，避免出现"凭空飞入"的卡片
    pulseHub(hubCore, hubGlow)
    return
  }
  const cardRect = cardEl.getBoundingClientRect()
  const cardW = Math.max(160, Math.min(cardRect.width, 280))
  const fromX = cardRect.left + cardRect.width / 2 - screenRect.left
  const fromY = cardRect.top + cardRect.height / 2 - screenRect.top
  // 原卡片立即淡出下沉，制造"被抽离"的错觉
  cardEl.animate(
    [
      { opacity: 1, transform: 'translateY(0) scale(1)' },
      { opacity: 0, transform: `translateY(6px) scale(${0.94})` },
    ],
    { duration: 320, easing: 'cubic-bezier(0.4, 0, 0.7, 0.2)', fill: 'forwards' },
  )

  // 构建 ghost 卡片
  const ghost = document.createElement('div')
  ghost.className = 'fly-ghost'
  ghost.style.width = cardW + 'px'

  const shell = document.createElement('div')
  shell.className = 'fly-ghost-shell'

  const bar = document.createElement('div')
  bar.className = 'fly-ghost-bar'

  const body = document.createElement('div')
  body.className = 'fly-ghost-body'

  const title = document.createElement('div')
  title.className = 'fly-ghost-title'
  title.textContent = step.name

  const path = document.createElement('div')
  path.className = 'fly-ghost-path'
  const phaseName = findParentPhase(stepId)
  path.textContent = phaseName && phaseName !== '--' ? phaseName : '任务完成'

  body.appendChild(title)
  body.appendChild(path)

  const tag = document.createElement('div')
  tag.className = 'fly-ghost-tag'
  tag.textContent = '已完成'

  shell.appendChild(bar)
  shell.appendChild(body)
  shell.appendChild(tag)
  ghost.appendChild(shell)

  // 飞行动画参数：慢速平移滑入，卡片保持较大尺寸飞抵圆环
  const flightMs = 1600
  const startScale = 0.94
  const endScale = 0.42

  // 初始位置内联设置，避免 WAAPI 首帧前的瞬闪（ghost 默认在 0,0）
  const startTransform = `translate(${fromX}px, ${fromY}px) translate(-50%, -50%) scale(${startScale}) rotate(-3deg)`
  ghost.style.transform = startTransform
  ghost.style.opacity = '1'
  flyLayer.appendChild(ghost)

  // 二次贝塞尔轨迹：平缓弧线滑向圆环，营造"平移"感
  const cpX = (fromX + hubCx) / 2
  const cpY = (fromY + hubCy) / 2 - 55
  const bezier = (t: number) => {
    const u = 1 - t
    return {
      x: u * u * fromX + 2 * u * t * cpX + t * t * hubCx,
      y: u * u * fromY + 2 * u * t * cpY + t * t * hubCy,
    }
  }

  const steps = 28
  const keyframes: Keyframe[] = []
  for (let i = 0; i <= steps; i++) {
    const t = i / steps
    const p = bezier(t)
    const scale = startScale + (endScale - startScale) * t
    // 末段渐隐：贴近圆环时才被"吸收"，平移过程保持清晰
    const opacity = t < 0.9 ? 1 : Math.max(0, 1 - (t - 0.9) / 0.1)
    keyframes.push({
      transform: `translate(${p.x}px, ${p.y}px) translate(-50%, -50%) scale(${scale}) rotate(${(t - 0.5) * 4}deg)`,
      opacity,
      offset: t,
    })
  }
  const flight = ghost.animate(keyframes, {
    duration: flightMs,
    easing: 'cubic-bezier(0.4, 0.05, 0.3, 1)',
    fill: 'forwards',
  })

  // 飞行尾迹粒子
  let trailTimer: number | null = window.setInterval(() => {
    const rect = ghost.getBoundingClientRect()
    spawnTrailDot(flyLayer, rect.left + rect.width / 2 - screenRect.left, rect.top + rect.height / 2 - screenRect.top)
  }, 64)

  flight.onfinish = () => {
    if (trailTimer !== null) {
      clearInterval(trailTimer)
      trailTimer = null
    }
    ghost.remove()
    triggerHubAbsorption(flyLayer, hubCx, hubCy)
    pulseHub(hubCore, hubGlow)
  }
}

// 尾迹粒子：随 ghost 飞行路径散落的青色光点，逐渐衰减
function spawnTrailDot(layer: HTMLElement, x: number, y: number) {
  const dot = document.createElement('div')
  dot.className = 'fly-trail-dot'
  const size = 3 + Math.random() * 4
  dot.style.width = size + 'px'
  dot.style.height = size + 'px'
  dot.style.left = x + 'px'
  dot.style.top = y + 'px'
  layer.appendChild(dot)
  const dx = (Math.random() - 0.5) * 18
  const dy = (Math.random() - 0.5) * 18 + 6
  dot.animate(
    [
      { opacity: 0.95, transform: 'translate(-50%, -50%) scale(1)' },
      { opacity: 0, transform: `translate(calc(-50% + ${dx}px), calc(-50% + ${dy}px)) scale(0.2)` },
    ],
    { duration: 720, easing: 'ease-out' },
  ).onfinish = () => dot.remove()
}

// 圆环吸收特效：冲击波环 + 放射状粒子迸发
function triggerHubAbsorption(layer: HTMLElement, x: number, y: number) {
  // 冲击波环（双层）
  const ring1 = document.createElement('div')
  ring1.className = 'hub-shockwave hub-shockwave-1'
  ring1.style.left = x + 'px'
  ring1.style.top = y + 'px'
  layer.appendChild(ring1)
  ring1.animate(
    [
      { width: '48px', height: '48px', opacity: 0.9, borderWidth: '3px' },
      { width: '300px', height: '300px', opacity: 0, borderWidth: '1px' },
    ],
    { duration: 820, easing: 'cubic-bezier(0.2, 0.8, 0.3, 1)' },
  ).onfinish = () => ring1.remove()

  const ring2 = document.createElement('div')
  ring2.className = 'hub-shockwave hub-shockwave-2'
  ring2.style.left = x + 'px'
  ring2.style.top = y + 'px'
  layer.appendChild(ring2)
  ring2.animate(
    [
      { width: '32px', height: '32px', opacity: 0.75, borderWidth: '2px' },
      { width: '200px', height: '200px', opacity: 0, borderWidth: '1px' },
    ],
    { duration: 640, easing: 'cubic-bezier(0.2, 0.8, 0.3, 1)' },
  ).onfinish = () => ring2.remove()

  // 放射状粒子迸发
  const count = 16
  for (let i = 0; i < count; i++) {
    const angle = (i / count) * Math.PI * 2 + Math.random() * 0.25
    const dist = 36 + Math.random() * 56
    const p = document.createElement('div')
    p.className = 'hub-burst-particle'
    p.style.left = x + 'px'
    p.style.top = y + 'px'
    layer.appendChild(p)
    const dx = Math.cos(angle) * dist
    const dy = Math.sin(angle) * dist
    p.animate(
      [
      { transform: 'translate(-50%, -50%) scale(1)', opacity: 1 },
        { transform: `translate(calc(-50% + ${dx}px), calc(-50% + ${dy}px)) scale(0.15)`, opacity: 0 },
      ],
      { duration: 560 + Math.random() * 260, easing: 'cubic-bezier(0.15, 0.7, 0.3, 1)' },
    ).onfinish = () => p.remove()
  }
}

// 圆环波动：卡片"扔进"圆环后的阻尼震荡——
// 放大 → 回落 → 二次微涨 → 复原，模拟水面涟漪式的能量衰减
function pulseHub(hubCore: HTMLElement, hubGlow: HTMLElement | null) {
  hubCore.animate(
    [
      { transform: 'scale(1)', filter: 'brightness(1) saturate(1)' },
      { transform: 'scale(1.32)', filter: 'brightness(1.9) saturate(1.45)', offset: 0.22 },
      { transform: 'scale(1.05)', filter: 'brightness(1.25) saturate(1.1)', offset: 0.5 },
      { transform: 'scale(1.14)', filter: 'brightness(1.5) saturate(1.2)', offset: 0.72 },
      { transform: 'scale(1)', filter: 'brightness(1) saturate(1)' },
    ],
    { duration: 1100, easing: 'cubic-bezier(0.2, 0.8, 0.3, 1)' },
  )
  if (hubGlow) {
    hubGlow.animate(
      [
        { transform: 'translate(-50%, -50%) scale(1)', opacity: 1 },
        { transform: 'translate(-50%, -50%) scale(1.55)', opacity: 1, offset: 0.22 },
        { transform: 'translate(-50%, -50%) scale(1.12)', opacity: 1, offset: 0.5 },
        { transform: 'translate(-50%, -50%) scale(1.28)', opacity: 1, offset: 0.72 },
        { transform: 'translate(-50%, -50%) scale(1)', opacity: 1 },
      ],
      { duration: 1100, easing: 'cubic-bezier(0.2, 0.8, 0.3, 1)' },
    )
  }
}

// === KPI 计算 ===
const completedCount = computed(() =>
  leafSteps.value.filter(s => ['completed', 'skipped', 'timeout', 'issue'].includes(s.status)).length
)
const totalCount = computed(() => leafSteps.value.length)
const progressPercent = computed(() => {
  if (totalCount.value === 0) return 0
  return Math.round((completedCount.value / totalCount.value) * 100)
})

// 演练开始时间（兼容 start_time / started_at 两种字段名）
const drillStartTime = computed(() =>
  (currentDrill.value as any)?.start_time || (currentDrill.value as any)?.started_at || null
)

// 驱动实时刷新和运行中状态计算
const elapsedSeconds = ref(0)
const totalDurationText = computed(() => {
  if (!drillStartTime.value) return '--:--:--'
  const h = Math.floor(elapsedSeconds.value / 3600)
  const m = Math.floor((elapsedSeconds.value % 3600) / 60)
  const s = elapsedSeconds.value % 60
  return [h, m, s].map(n => String(n).padStart(2, '0')).join(':')
})

// === 异常任务数指标 ===
// 异常任务：状态为 issue 的任务
const abnormalTaskCount = computed(() => {
  if (!drillSteps.value || drillSteps.value.length === 0) return 0
  // 计算异常任务：issue 状态的步骤
  const issueCount = drillSteps.value.filter(s => s.status === 'issue').length
  return issueCount
})

// 告警等级：safe (0个) -> warning (1-2个) -> danger (≥3个)
const alertLevel = computed(() => {
  const count = abnormalTaskCount.value
  if (count === 0) return 'safe'
  if (count <= 2) return 'warning'
  return 'danger'
})

// 告警状态标签
const alertLabel = computed(() => {
  const labels: Record<string, string> = {
    safe: '运行正常',
    warning: '轻微异常',
    danger: '需要关注'
  }
  return labels[alertLevel.value] || '运行正常'
})

// 状态标签
const drillStatusLabel = computed(() => {
  const map: Record<string, string> = {
    running: '执行中', paused: '已暂停', completed: '已完成', terminated: '已终止', pending: '待启动',
  }
  return map[currentDrill.value?.status || ''] || '未知'
})

// === 阶段（stages） ===
// 将步骤均匀分成 4 个阶段；如果演练有 8 步且要演示 6/8，可以从 completedSteps 截取
const STAGE_NAMES = [
  '业务验收与告警',
  '演练复盘与行动',
  '演练启动与人员',
  '基线指标与备份',
]

const buildStageSegments = (finished: number, total: number, active: boolean) => {
  const segCount = 18
  if (total <= 0) return Array.from({ length: segCount }, () => 'empty')
  const doneCount = finished >= total ? segCount : Math.floor((finished / total) * segCount)
  return Array.from({ length: segCount }, (_, i) => {
    if (i < doneCount) return 'done'
    if (active && finished < total && i === doneCount) return 'active'
    return 'todo'
  })
}

const stages = computed(() => {
  // 依赖 elapsedSeconds 每秒刷新，使未结束阶段的截止时间实时跳动
  const _tick = elapsedSeconds.value
  const allSteps = drillSteps.value
  if (allSteps.length === 0) return []

  const roots = rootSteps.value
  const hasHierarchy = roots.length > 0 && leafSteps.value.length < allSteps.length

  if (hasHierarchy) {
    // 有层级结构：每个根步骤就是一个阶段
    return roots.map((root, idx) => {
      const directChildren = childMap.value.get(root.id) || []
      const leaves = collectLeaves(root.id)
      // 终态判断（completed/skipped/timeout/issue 都视为已结束）
      const isTerminal = (s: StepInstance) => ['completed', 'skipped', 'timeout', 'issue'].includes(s.status)
      // 环节统计：基于每个环节下的叶子步骤完成情况判断
      const completedPhases = directChildren.filter(c => {
        const phaseLeaves = collectLeaves(c.id)
        return phaseLeaves.length > 0 && phaseLeaves.every(l => isTerminal(l))
      }).length
      const totalPhases = directChildren.length
      // 步骤统计：叶子节点（实际操作步骤）
      const finishedLeaves = leaves.filter(s => isTerminal(s)).length
      const totalLeaves = leaves.length
      const running = leaves.some(s => s.status === 'running')
      const hasIssue = leaves.some(s => s.status === 'issue' || s.status === 'timeout')
      const allDone = leaves.every(s => isTerminal(s)) && totalLeaves > 0
      // 部分完成：有已完成的叶子步骤，但不是全部
      const hasProgress = finishedLeaves > 0 && !allDone
      const status = allDone ? 'done' : (running ? 'running' : (hasIssue ? 'issue' : (hasProgress ? 'running' : 'pending')))
      // 时间范围
      const started = leaves.find(s => s.start_time)?.start_time
      const ended = [...leaves].reverse().find(s => s.end_time)?.end_time
      let endStr: string | null = ended ?? null
      if (!endStr) {
        const t = leaves[0]?.timeout_minutes
        if (t) endStr = new Date(Date.now() + t * 60000).toISOString()
      }
      const timeRange = `${formatHM(started)} / ${formatHM(endStr)}`
      const segs = buildStageSegments(finishedLeaves, totalLeaves, running)
      const team = leaves.find(s => s.executor_team)?.executor_team
        || leaves.find(s => s.assignee_names)?.assignee_names
        || '运维部'
      return {
        name: root.name || STAGE_NAMES[idx % STAGE_NAMES.length],
        status,
        timeRange,
        completedPhases,
        totalPhases,
        completedSteps: finishedLeaves,
        totalSteps: totalLeaves,
        segments: segs,
        team,
        phaseNames: directChildren.map(c => c.name),
      }
    })
  }

  // 无层级：均匀分成 4 个阶段
  const total = allSteps.length
  const stageCount = Math.min(4, total)
  const perStage = Math.ceil(total / stageCount)
  const isTerminal = (s: StepInstance) => ['completed', 'skipped', 'timeout', 'issue'].includes(s.status)
  return Array.from({ length: stageCount }).map((_, idx) => {
    const slice = allSteps.slice(idx * perStage, (idx + 1) * perStage)
    const finished = slice.filter(s => isTerminal(s)).length
    const running = slice.some(s => s.status === 'running')
    const hasIssue = slice.some(s => s.status === 'issue' || s.status === 'timeout')
    const allDone = slice.every(s => isTerminal(s)) && slice.length > 0
    const hasProgress = finished > 0 && !allDone
    const status = allDone ? 'done' : (running ? 'running' : (hasIssue ? 'issue' : (hasProgress ? 'running' : 'pending')))
    const started = slice.find(s => s.start_time)?.start_time
    const ended = slice.find(s => s.end_time)?.end_time
    let endStr: string | null = ended ?? null
    if (!endStr) {
      const t = slice[0]?.timeout_minutes
      if (t) endStr = new Date(Date.now() + t * 60000).toISOString()
    }
    const timeRange = `${formatHM(started)} / ${formatHM(endStr)}`
    const segs = buildStageSegments(finished, slice.length, running)
    const team = slice.find(s => s.executor_team)?.executor_team
      || slice.find(s => s.assignee_names)?.assignee_names
      || '运维部'
    return {
      name: STAGE_NAMES[idx % STAGE_NAMES.length],
      status,
      timeRange,
      completedPhases: 0,
      totalPhases: 0,
      completedSteps: finished,
      totalSteps: slice.length,
      segments: segs,
      team,
      phaseNames: [] as string[],
    }
  })
})

// 当前阶段 index（高亮 + 中心环）
const currentPhaseIndex = computed(() => {
  const i = stages.value.findIndex(s => s.status === 'running')
  if (i >= 0) return i
  // 如果都在 done，按最后一个 done
  const lastDone = stages.value.map(s => s.status).lastIndexOf('done')
  if (lastDone >= 0) return Math.min(lastDone, stages.value.length - 1)
  return 0
})

const selectedPhaseIndex = ref<number | null>(null)
const displayPhaseIndex = computed(() => {
  const count = stages.value.length
  if (count === 0) return 0
  const selected = selectedPhaseIndex.value
  if (selected == null) return currentPhaseIndex.value
  return Math.max(0, Math.min(selected, count - 1))
})

const currentPhaseName = computed(() => stages.value[currentPhaseIndex.value]?.name || '演练启动')
const displayPhaseName = computed(() => stages.value[displayPhaseIndex.value]?.name || '演练启动')

const currentPhaseProgress = computed(() => {
  const s = stages.value[currentPhaseIndex.value]
  if (!s) return { num: 0, den: 0 }
  return { num: s.completedSteps, den: s.totalSteps }
})

// 当前展示阶段的步骤进度（与左侧阶段卡片"步骤"口径一致，用于中心环 hub）
// 注意：与全局 progressPercent 区分——此处仅统计当前展示阶段的叶子步骤
const displayPhaseProgress = computed(() => {
  const s = stages.value[displayPhaseIndex.value]
  if (!s || s.totalSteps === 0) return 0
  return Math.round((s.completedSteps / s.totalSteps) * 100)
})
const displayPhaseStepNum = computed(() => stages.value[displayPhaseIndex.value]?.completedSteps ?? 0)
const displayPhaseStepDen = computed(() => stages.value[displayPhaseIndex.value]?.totalSteps ?? 0)

function selectPhase(index: number) {
  if (index < 0 || index >= stages.value.length) return
  selectedPhaseIndex.value = index
}

// 阶段环需要的相位名称 + 各阶段环节名称
const ringPhases = computed(() => {
  return stages.value.map(s => s.name)
})

// 每个阶段的环节名称列表
const ringPhaseNames = computed(() => {
  return stages.value.map(s => s.phaseNames || [])
})

// 当前进行中的环节名（取跑道当前阶段里状态为 running 的环节节点），
// 无运行环节时回退：首个未完成环节 → 当前阶段名
const currentRunningNodeName = computed(() => {
  const names = ringPhaseNames.value[displayPhaseIndex.value] || []
  const statuses = ringPhaseNodeStatuses.value[displayPhaseIndex.value] || []
  const runningIdx = statuses.findIndex(s => s?.status === 'running')
  if (runningIdx >= 0 && names[runningIdx]) return names[runningIdx]
  const pendingIdx = statuses.findIndex(s => s?.status !== 'completed')
  if (pendingIdx >= 0 && names[pendingIdx]) return names[pendingIdx]
  return currentPhaseName.value
})

const ringPhaseStatuses = computed(() => {
  return stages.value.map(s => s.status)
})

// 每个阶段中每个环节节点的状态信息
const ringPhaseNodeStatuses = computed(() => {
  const _tick = elapsedSeconds.value
  const nowMs = Date.now()
  const isTerminal = (s: StepInstance) => ['completed', 'skipped', 'timeout', 'issue'].includes(s.status)

  return rootSteps.value.map(root => {
    const directChildren = childMap.value.get(root.id) || []
    return directChildren.map(child => {
      const leaves = collectLeaves(child.id)
      if (leaves.length === 0) {
        // 无子任务的环节：以自身状态计 0/1 或 1/1
        const selfDone = isTerminal(child) ? 1 : 0
        return { status: child.status, progress: selfDone * 100, completed: selfDone, total: 1 }
      }

      const totalLeaves = leaves.length
      const finishedLeaves = leaves.filter(s => isTerminal(s)).length
      const isRunning = leaves.some(s => s.status === 'running')
      const isDone = leaves.every(s => isTerminal(s)) && totalLeaves > 0
      const hasIssue = leaves.some(s => s.status === 'issue' || s.status === 'timeout')

      let progress = 0
      if (isDone) {
        progress = 100
      } else if (totalLeaves > 0) {
        // 只计算已完成任务的进度比例（不包含运行中任务的时间进度估算）
        progress = Math.round((finishedLeaves / totalLeaves) * 100)
      }

      const status = isDone ? 'completed' : hasIssue ? 'issue' : isRunning ? 'running' : 'pending'
      return { status, progress: Math.min(progress, 100), completed: finishedLeaves, total: totalLeaves }
    })
  })
})

const ringSize = computed(() => {
  const panelHeight = centerPanelHeight.value || Math.max(260, viewportHeight.value - 204)
  const maxRingFromH = Math.floor((panelHeight - 24) / 0.82)
  // 基于中间面板宽度限制，避免非全屏时按整屏宽度估算导致内容偏左裁切
  const panelWidth = centerPanelWidth.value || viewportWidth.value
  const maxRingFromW = Math.floor((panelWidth - 4) / 1.72)
  return Math.min(700, Math.max(230, maxRingFromW), Math.max(230, maxRingFromH))
})

function measureCenterPanel() {
  if (!centerPanelRef.value) return
  centerPanelWidth.value = centerPanelRef.value.clientWidth
  centerPanelHeight.value = centerPanelRef.value.clientHeight
}

// === 告警 ===
// 从步骤的"进行中"或异常中推算
function safeParseJSON(str: string): Record<string, any> | null {
  try { return JSON.parse(str) } catch { return null }
}

const activeAlerts = computed(() => {
  // 依赖 elapsedSeconds 使 computed 每秒重算
  const _now = elapsedSeconds.value

  const items: Array<{
    stepId: number
    title: string
    operator: string
    team: string
    parentPhase: string
    directParent: string
    statusLabel: string
    level: 'warn' | 'info' | 'danger' | 'ok' | 'pending'
    seq: number
    startedAt: string | null
    timeoutMin: number
  }> = []

  // 展示当前运行环节下的全部叶子任务（含各状态），与跑道当前环节节点联动
  const names = ringPhaseNames.value[displayPhaseIndex.value] || []
  const statuses = ringPhaseNodeStatuses.value[displayPhaseIndex.value] || []
  const runningIdx = statuses.findIndex(s => s?.status === 'running')
  const nodeIdx = runningIdx >= 0 ? runningIdx : statuses.findIndex(s => s?.status !== 'completed')
  if (nodeIdx < 0 || !names[nodeIdx]) return items

  // 定位环节节点对应的步骤（根步骤的直接子节点）
  const root = rootSteps.value[displayPhaseIndex.value]
  const nodeStep = root ? (childMap.value.get(root.id) || [])[nodeIdx] : undefined
  if (!nodeStep) return items

  const levelOf = (status: string): 'warn' | 'info' | 'danger' | 'ok' | 'pending' => {
    if (status === 'running') return 'warn'
    if (status === 'issue' || status === 'timeout') return 'danger'
    if (status === 'completed' || status === 'done' || status === 'skipped') return 'ok'
    return 'pending'
  }
  const labelOf = (status: string): string => {
    if (status === 'running') return '执行中'
    if (status === 'issue') return '异常'
    if (status === 'timeout') return '超时'
    if (status === 'completed' || status === 'done') return '已完成'
    if (status === 'skipped') return '已跳过'
    return '待执行'
  }

  // 只保留未完成任务（已完成/已跳过视为完成，不进入列表）
  const isIncomplete = (status: string) => !['completed', 'done', 'skipped'].includes(status)

  collectLeaves(nodeStep.id)
    .filter(s => isIncomplete(s.status))
    .forEach(s => {
    const attrs = typeof (s as any).attributes === 'string'
      ? safeParseJSON((s as any).attributes)
      : (s as any).attributes
    items.push({
      stepId: s.id,
      title: s.name,
      operator: attrs?.operator || '',
      team: s.executor_team || '运维部',
      parentPhase: findParentPhase(s.id),
      directParent: findDirectParent(s.id),
      statusLabel: labelOf(s.status),
      level: levelOf(s.status),
      seq: s.seq,
      startedAt: (s as any).start_time || null,
      timeoutMin: (s as any).timeout_minutes || COUNTDOWN_DEFAULT_MIN,
    })
  })

  // 按流程顺序排序
  return items.sort((a, b) => a.seq - b.seq)
})

// ===== 执行中步骤剩余时间进度条（仅大屏展示，不影响流程流转） =====
// 超时时间默认 10 分钟；无 start_time 时以当前时间为起点
const COUNTDOWN_DEFAULT_MIN = 10
// 无 start_time 的步骤以组件加载时刻为计时起点
const elapsedBase = Date.now()

function alertCountdownPercent(alert: { startedAt: string | null, timeoutMin: number }): number {
  const start = alert.startedAt ? new Date(alert.startedAt).getTime() : elapsedBase
  // 进度条固定按 10 分钟窗口推进：满 10 分钟即 100% 不再增长（不影响右侧耗时文本继续计时）
  const total = COUNTDOWN_DEFAULT_MIN * 60 * 1000
  const used = Math.max(Date.now() - start, 0)
  return Math.min((used / total) * 100, 100)
}

// 展示已耗时：一直显示"已耗时"文案，超时后仅进度条转红脉冲提醒，文本继续计时
// 时长自适应格式：<1h 用 m:ss，<1d 用 h:mm:ss，≥1d 用 d天h小时
function alertCountdownText(alert: { startedAt: string | null, timeoutMin: number }): string {
  const start = alert.startedAt ? new Date(alert.startedAt).getTime() : elapsedBase
  const used = Math.max(Date.now() - start, 0)
  const s = Math.floor((used % 60000) / 1000)
  const m = Math.floor((used % 3600000) / 60000)
  const h = Math.floor((used % 86400000) / 3600000)
  const d = Math.floor(used / 86400000)
  let time: string
  if (d >= 1) time = `${d}天${h}小时`
  else if (h >= 1) time = `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
  else time = `${m}:${String(s).padStart(2, '0')}`
  return `已耗时 ${time}`
}

// 可见步骤数量：按容器和实际卡片尺寸自适应，避免在不同屏幕上写死展示数量
const ALERT_CARD_FALLBACK_HEIGHT = 106
const ALERT_CARD_FALLBACK_GAP = 9
const MORE_TIP_FALLBACK_HEIGHT = 32
const containerHeight = ref(0)
const alertCardHeight = ref(ALERT_CARD_FALLBACK_HEIGHT)
const alertCardGap = ref(ALERT_CARD_FALLBACK_GAP)
const moreTipHeight = ref(MORE_TIP_FALLBACK_HEIGHT)
let warnListResizeObserver: ResizeObserver | null = null

function setAlertCardRef(el: Element | ComponentPublicInstance | null, index: number) {
  if (el instanceof HTMLElement) alertCardRefs.value[index] = el
}

function measureWarnList() {
  const list = warnListRef.value
  if (!list) return

  const style = window.getComputedStyle(list)
  const paddingY = parseFloat(style.paddingTop || '0') + parseFloat(style.paddingBottom || '0')
  containerHeight.value = Math.max(0, list.clientHeight - paddingY)
  alertCardGap.value = parseFloat(style.rowGap || style.gap || '') || ALERT_CARD_FALLBACK_GAP

  const firstCard = alertCardRefs.value.find(Boolean)
  if (firstCard) {
    const cardStyle = window.getComputedStyle(firstCard)
    const minHeight = parseFloat(cardStyle.minHeight || '') || 0
    alertCardHeight.value = Math.max(minHeight, ALERT_CARD_FALLBACK_HEIGHT)
  }
  if (moreTipRef.value) moreTipHeight.value = moreTipRef.value.getBoundingClientRect().height || MORE_TIP_FALLBACK_HEIGHT
  if (execLogRef.value) execLogViewHeight.value = execLogRef.value.clientHeight || 96
}

// 执行日志：贴底驻留展示（时间正序，最新在最下）；
// 超出一屏时旧日志在顶部被裁剪，产生新日志时整体上移一条
const execLogLines = computed(() => {
  const lines = [...recentLogs.value].reverse()
  const maxVisible = Math.max(1, Math.floor(execLogViewHeight.value / EXEC_LOG_LINE_HEIGHT) + 2)
  return lines.slice(Math.max(0, lines.length - maxVisible))
})

const visibleAlertCount = computed(() => {
  // 依赖 elapsedSeconds 使其每秒重算
  const _t = elapsedSeconds.value
  const available = containerHeight.value
  if (!available) return 5
  const cardHeight = Math.max(alertCardHeight.value, 1)
  const gap = alertCardGap.value
  const firstPass = Math.max(1, Math.floor((available + gap) / (cardHeight + gap)))
  const hasMore = activeAlerts.value.length > firstPass
  const reserved = hasMore ? moreTipHeight.value + gap : 0
  const count = Math.floor((available - reserved + gap) / (cardHeight + gap))
  return Math.min(Math.max(count, 1), activeAlerts.value.length)
})
const visibleAlerts = computed(() => activeAlerts.value.slice(0, visibleAlertCount.value))

// === 日志（已有数据） ===
function logActionClass(action: string): string {
  if (!action) return 'step'
  if (action.includes('issue') || action.includes('timeout')) return 'danger'
  if (action.includes('skip')) return 'skip'
  if (action.includes('force')) return 'force'
  if (action.includes('start') || action.includes('resume')) return 'ok'
  if (action.includes('complete') || action.includes('terminate')) return 'step'
  return 'step'
}
// 日志动作统一收敛为两种状态：完成（含各类完成变体）/ 异常（含超时）
function logActionLabel(action: string): string {
  if (action.includes('issue') || action.includes('timeout')) return '异常'
  return '完成'
}

// 根据 log 中的 step_instance_id 在 drillSteps 中查节点名称
function resolveStepName(log: StepInstanceLog): string {
  const id = log?.step_instance_id
  if (!id) return currentDrill.value?.name || '演练'
  const step = drillSteps.value.find(s => s.id === id)
  if (step?.name) return step.name
  return `步骤 #${id}`
}
function stageBadgeLabel(status: string): string {
  const map: Record<string, string> = {
    done: '已完成', running: '进行中', pending: '待开始', issue: '异常',
  }
  return map[status] || status
}

function chineseNum(n: number): string {
  return ['一', '二', '三', '四', '五', '六'][n - 1] || String(n)
}

function formatTime(dateStr: string | null | undefined): string {
  if (!dateStr) return '--:--:--'
  const d = new Date(dateStr)
  return [d.getHours(), d.getMinutes(), d.getSeconds()].map(n => String(n).padStart(2, '0')).join(':')
}
function formatHM(dateStr: string | null | undefined): string {
  if (!dateStr) return '--:--'
  const d = new Date(dateStr)
  return [d.getHours(), d.getMinutes()].map(n => String(n).padStart(2, '0')).join(':')
}
function formatSystemTime(d: Date): string {
  return `${d.getFullYear()}.${String(d.getMonth() + 1).padStart(2, '0')}.${String(d.getDate()).padStart(2, '0')} ${[d.getHours(), d.getMinutes(), d.getSeconds()].map(n => String(n).padStart(2, '0')).join(':')}`
}

// 计时：每秒钟刷新 systemTime / elapsed
function tick() {
  const now = new Date()
  systemTime.value = formatSystemTime(now)
  const started = drillStartTime.value
  if (started) {
    const start = new Date(started).getTime()
    if (!isNaN(start)) {
      const ended = (currentDrill.value as any)?.end_time || (currentDrill.value as any)?.completed_at
      const end = ended
        ? new Date(ended).getTime()
        : now.getTime()
      elapsedSeconds.value = Math.max(0, Math.round((end - start) / 1000))
    }
  }
  // 更新容器高度（用于截断计算）
  measureWarnList()
}

// 数据加载
async function loadData() {
  if (dataLoading) {
    dataReloadQueued = true
    return
  }
  if (!drillId.value) {
    error.value = '无效的演练 ID'
    loading.value = false
    return
  }
  dataLoading = true
  try {
    const previousStepStatuses = new Map(drillSteps.value.map(step => [step.id, step.status]))

    const drill = await drillApi.getDetail(drillId.value)
    if (componentDestroyed) return
    currentDrill.value = drill

    const steps = await drillApi.getSteps(drillId.value)
    if (componentDestroyed) return
    const sortedSteps = steps.sort((a, b) => a.seq - b.seq)
    const newlyCompletedStepIds = detectNewlyCompletedLeafSteps(previousStepStatuses, sortedSteps)
    drillSteps.value = sortedSteps
    newlyCompletedStepIds.forEach(playTaskFlowAnimation)

    const logs = await drillApi.getLogs(drillId.value)
    if (componentDestroyed) return
    recentLogs.value = logs.slice(0, 30)

    loading.value = false
    error.value = null
    tick()
    // 仅在 WebSocket 未连接时建立连接，避免刷新数据时重连导致循环
    if (!ws || ws.readyState === WebSocket.CLOSING || ws.readyState === WebSocket.CLOSED) {
      connectWebSocket()
    }
  } catch (err: any) {
    if (componentDestroyed) return
    error.value = err.message || '加载数据失败'
    console.error('Failed to load drill data:', err)
    loading.value = false
  } finally {
    dataLoading = false
    if (dataReloadQueued && !componentDestroyed) {
      dataReloadQueued = false
      scheduleDataRefresh()
    }
  }
}
function handleRetry() {
  loadData()
}

// WebSocket
const REFRESH_EVENTS = new Set([
  'step_change', 'drill_status',
  'step_start', 'step_started', 'step_complete', 'step_completed', 'step_issue',
  'step_skip', 'step_skipped', 'step_timeout', 'step_force_complete',
  'drill_started', 'drill_paused', 'drill_resumed', 'drill_completed', 'drill_terminated',
  'timeout_warning', 'timeout_alert',
])
const LOG_EVENTS: Record<string, string> = {
  step_start: 'step_start',
  step_started: 'step_start',
  step_complete: 'step_complete',
  step_completed: 'step_complete',
  step_issue: 'step_issue',
  step_skip: 'step_skip',
  step_skipped: 'step_skip',
  step_timeout: 'step_timeout',
  step_force_complete: 'force_complete',
  drill_started: 'drill_started',
  drill_paused: 'drill_paused',
  drill_resumed: 'drill_resumed',
  drill_completed: 'drill_completed',
  drill_terminated: 'drill_terminated',
}

function connectWebSocket() {
  if (componentDestroyed) return
  if (ws && (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING)) return
  if (ws) ws.close()
  const wsProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
  const authStore = useAuthStore()
  const wsUrl = `${wsProtocol}://${window.location.host}/ws/control/${drillId.value}?token=${authStore.token}`
  ws = new WebSocket(wsUrl)
  ws.onmessage = (event) => {
    if (componentDestroyed) return
    try {
      // 后端 WritePump 会把消息打包成 JSON 数组批量发送，需逐条处理
      const parsed = JSON.parse(event.data)
      const messages = Array.isArray(parsed) ? parsed : [parsed]
      for (const data of messages) {
        const eventType = normalizeWsEvent(data.type || data.event_type || data.event)
        if (!eventType) continue

        const payload = data.payload || data.data || data

        // 步骤事件：增量更新 + 推入本地日志
        if (eventType.startsWith('step_')) {
          applyStepEvent(eventType, payload)
        } else if (eventType.startsWith('drill_')) {
          applyDrillEvent(eventType, payload)
        }

        // 合并刷新确保级联状态正确（延迟执行，让增量先生效）
        if (eventType.startsWith('step_') || eventType.startsWith('drill_') || REFRESH_EVENTS.has(eventType)) {
          scheduleDataRefresh()
        }
      }
    } catch (e) { /* ignore */ }
  }
  ws.onerror = () => {
    if (componentDestroyed) return
    startFallbackPolling()
  }
  ws.onclose = () => {
    if (componentDestroyed) return
    if (currentDrill.value?.status === 'running') startFallbackPolling()
  }
}

function scheduleDataRefresh(delay = 160) {
  if (componentDestroyed) return
  if (dataRefreshTimer) clearTimeout(dataRefreshTimer)
  dataRefreshTimer = window.setTimeout(() => {
    dataRefreshTimer = null
    loadData()
  }, delay)
}

// 增量应用步骤事件,避免每次都重拉 3 个 API
function applyStepEvent(eventType: string, payload: any) {
  if (!payload) return
  const stepId = Number(payload.step_id ?? payload.stepId ?? payload.id ?? payload.step_instance_id)
  if (!stepId) {
    scheduleDataRefresh()
    return
  }
  const idx = drillSteps.value.findIndex(s => s.id === stepId)
  if (idx === -1) {
    // 找不到对应步骤,降级为全量刷新
    scheduleDataRefresh()
    return
  }
  const target = { ...drillSteps.value[idx] }
  const newStatus = payload.new_status || mapEventToStatus(eventType)
  if (newStatus) target.status = newStatus as StepStatus
  if (payload.start_time) target.start_time = payload.start_time
  if (payload.end_time) target.end_time = payload.end_time
  if (payload.timeout_at) target.timeout_at = payload.timeout_at
  if (payload.remark) target.remark = payload.remark
  if (payload.issue_desc) target.issue_desc = payload.issue_desc
  if (payload.assignee_names) target.assignee_names = payload.assignee_names

  const nextSteps = [...drillSteps.value]
  nextSteps[idx] = target
  drillSteps.value = nextSteps

  // 推入一条本地日志
  const logAction = LOG_EVENTS[eventType] || eventType
  const newLog: StepInstanceLog = {
    id: Date.now(),
    step_instance_id: stepId,
    action: logAction,
    operator_id: 0,
    operator_name: payload.executor || '流程引擎',
    content: payload.remark || payload.comment || payload.issue_desc || '',
    created_at: new Date().toISOString(),
  }
  recentLogs.value = [newLog, ...recentLogs.value].slice(0, 30)

  // 任务完成流式动画（step_complete 为归一化后的事件名）
  // 仅对叶子步骤（实际操作步骤）触发；父级步骤（环节/阶段）的级联完成事件
  // 不触发，避免一次完成产生多个 ghost（1-2-1-2 问题的根因）
  if (eventType === 'step_complete') {
    const isLeafStep = leafSteps.value.some(s => s.id === stepId)
    if (isLeafStep) playTaskFlowAnimation(stepId)
  }

  // 重新计算 KPI
  recomputeKpis()
}

function applyDrillEvent(eventType: string, payload: any) {
  if (!payload) return
  if (currentDrill.value) {
    const newStatus = payload.new_status || mapEventToStatus(eventType)
    if (newStatus) currentDrill.value.status = newStatus as DrillStatus
  }
  const newLog: StepInstanceLog = {
    id: Date.now(),
    step_instance_id: null,
    action: LOG_EVENTS[eventType] || eventType,
    operator_id: 0,
    operator_name: payload.operator || '流程引擎',
    content: payload.remark || '',
    created_at: new Date().toISOString(),
  }
  recentLogs.value = [newLog, ...recentLogs.value].slice(0, 30)
  recomputeKpis()
}

function mapEventToStatus(eventType: string): string {
  const map: Record<string, string> = {
    step_start: 'running',
    step_started: 'running',
    step_complete: 'completed',
    step_completed: 'completed',
    step_force_complete: 'completed',
    step_issue: 'issue',
    step_skip: 'skipped',
    step_skipped: 'skipped',
    step_timeout: 'timeout',
    drill_started: 'running',
    drill_paused: 'paused',
    drill_resumed: 'running',
    drill_completed: 'completed',
    drill_terminated: 'terminated',
  }
  return map[eventType] || ''
}

function normalizeWsEvent(eventType: string): string {
  const aliases: Record<string, string> = {
    step_start: 'step_started',
    step_skip: 'step_skipped',
    step_completed: 'step_complete',
    step_force_complete: 'step_complete',
  }
  return aliases[eventType] || eventType
}

function recomputeKpis() {
  // 触发 drillSteps / currentDrill 的依赖计算
  // Vue 的响应式系统会自动重算
  tick()
}

function startFallbackPolling() {
  if (componentDestroyed) return
  if (refreshTimer) clearInterval(refreshTimer)
  refreshTimer = window.setInterval(() => {
    if (componentDestroyed) {
      stopFallbackPolling()
      return
    }
    if (currentDrill.value?.status === 'running') scheduleDataRefresh(0)
  }, 5000)
}

function stopFallbackPolling() {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

// 全屏
async function toggleFullscreen() {
  if (document.fullscreenElement || fallbackFullscreen.value) {
    if (document.fullscreenElement) await document.exitFullscreen()
    fallbackFullscreen.value = false
    return
  }

  const target = screenRootRef.value || document.documentElement
  try {
    await target.requestFullscreen()
  } catch (err) {
    console.warn('native fullscreen failed, fallback to page fullscreen:', err)
    fallbackFullscreen.value = true
    ElMessage.info('当前预览容器不允许浏览器全屏，已切换为页面内全屏')
  }
}

function handleFullscreenChange() {
  isNativeFullscreen.value = Boolean(document.fullscreenElement)
  if (document.fullscreenElement) fallbackFullscreen.value = false
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && fallbackFullscreen.value) {
    fallbackFullscreen.value = false
  }
}

function cleanBrandValue(value: unknown): string {
  return typeof value === 'string' ? value.trim() : ''
}

async function loadScreenBrand() {
  try {
    const response = await fetch('/screen-brand.json', { cache: 'no-store' })
    if (!response.ok) return
    const config = await response.json() as ScreenBrandConfig
    screenBrand.value = {
      companyName: cleanBrandValue(config.companyName) || DEFAULT_SCREEN_BRAND.companyName,
      fullscreenIconText: cleanBrandValue(config.fullscreenIconText) || DEFAULT_SCREEN_BRAND.fullscreenIconText,
      fullscreenIconImage: cleanBrandValue(config.fullscreenIconImage),
    }
  } catch (err) {
    console.warn('load screen brand config failed:', err)
  }
}

function handleBrandIconError() {
  screenBrand.value = {
    ...screenBrand.value,
    fullscreenIconImage: '',
  }
}

onMounted(() => {
  componentDestroyed = false
  loadScreenBrand()
  loadData()
  window.addEventListener('resize', handleResize)
  window.addEventListener('keydown', handleKeydown)
  document.addEventListener('fullscreenchange', handleFullscreenChange)
  nextTick(() => {
    measureCenterPanel()
    if (centerPanelRef.value && typeof ResizeObserver !== 'undefined') {
      centerPanelResizeObserver = new ResizeObserver(measureCenterPanel)
      centerPanelResizeObserver.observe(centerPanelRef.value)
    }
    measureWarnList()
    if (warnListRef.value && typeof ResizeObserver !== 'undefined') {
      warnListResizeObserver = new ResizeObserver(measureWarnList)
      warnListResizeObserver.observe(warnListRef.value)
    }
  })
  timeTimer = window.setInterval(tick, 1000)
})
onBeforeUnmount(() => {
  componentDestroyed = true
  window.removeEventListener('resize', handleResize)
  window.removeEventListener('keydown', handleKeydown)
  document.removeEventListener('fullscreenchange', handleFullscreenChange)
  centerPanelResizeObserver?.disconnect()
  centerPanelResizeObserver = null
  warnListResizeObserver?.disconnect()
  warnListResizeObserver = null
  if (timeTimer) clearInterval(timeTimer)
  if (dataRefreshTimer) clearTimeout(dataRefreshTimer)
  stopFallbackPolling()
  if (ws) { ws.close(); ws = null }
})

function handleResize() {
  viewportWidth.value = window.innerWidth
  viewportHeight.value = window.innerHeight
  measureCenterPanel()
  measureWarnList()
}

watch(visibleAlerts, () => {
  nextTick(measureWarnList)
}, { flush: 'post' })
</script>

<style lang="scss" scoped>
@use '@/styles/variables' as *;

// ===== 主题色 =====
$bg-deep: #040a1a;
$bg-mid: #061029;
$bg-card: rgba(8, 24, 56, 0.72);
$line: rgba(0, 212, 255, 0.18);
$line-strong: rgba(0, 212, 255, 0.5);
$neon: #00d4ff;
$neon-dim: #00a0c8;
$neon-soft: rgba(0, 212, 255, 0.15);
$ok: #00ff9c;
$warn: #ffb648;
$danger: #ff4d6a;
$text: #d6e8ff;
$text-dim: #a9c7ec;
$text-mute: #7f9fc7;

$font-display: 'Microsoft YaHei', 'PingFang SC', 'Segoe UI', Arial, sans-serif;
$font-mono: Consolas, Menlo, Monaco, 'Courier New', monospace;
$font-cn: 'Microsoft YaHei', 'PingFang SC', 'Hiragino Sans GB', Arial, sans-serif;

.screen-root {
  position: relative;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  /* 统一缩放基准：所有 vw 字号基于此，保持比例一致 */
  font-size: clamp(14px, 0.92vw, 17px);
  background:
    radial-gradient(circle at 50% 48%, rgba(0, 96, 205, 0.18), transparent 34%),
    radial-gradient(circle at 74% 68%, rgba(255, 122, 0, 0.08), transparent 18%),
    linear-gradient(180deg, #061229 0%, #020815 100%);
  color: $text;
  font-family: $font-cn;
  user-select: none;

  &::before {
    content: '';
    position: absolute;
    inset: 8px 12px 8px;
    border: 1px solid rgba(134, 181, 255, 0.45);
    box-shadow:
      inset 0 0 42px rgba(38, 118, 255, 0.1),
      0 0 28px rgba(38, 118, 255, 0.08);
    pointer-events: none;
    z-index: 1;
  }

  &::after {
    content: '';
    position: absolute;
    inset: 0;
    background-image:
      linear-gradient(30deg, rgba(93, 151, 240, 0.08) 12%, transparent 12.5%, transparent 87%, rgba(93, 151, 240, 0.08) 87.5%, rgba(93, 151, 240, 0.08)),
      linear-gradient(150deg, rgba(93, 151, 240, 0.08) 12%, transparent 12.5%, transparent 87%, rgba(93, 151, 240, 0.08) 87.5%, rgba(93, 151, 240, 0.08)),
      linear-gradient(30deg, rgba(93, 151, 240, 0.08) 12%, transparent 12.5%, transparent 87%, rgba(93, 151, 240, 0.08) 87.5%, rgba(93, 151, 240, 0.08)),
      linear-gradient(150deg, rgba(93, 151, 240, 0.08) 12%, transparent 12.5%, transparent 87%, rgba(93, 151, 240, 0.08) 87.5%, rgba(93, 151, 240, 0.08)),
      linear-gradient(60deg, rgba(40, 99, 180, 0.08) 25%, transparent 25.5%, transparent 75%, rgba(40, 99, 180, 0.08) 75%, rgba(40, 99, 180, 0.08)),
      linear-gradient(60deg, rgba(40, 99, 180, 0.08) 25%, transparent 25.5%, transparent 75%, rgba(40, 99, 180, 0.08) 75%, rgba(40, 99, 180, 0.08));
    background-position: 0 0, 0 0, 18px 32px, 18px 32px, 0 0, 18px 32px;
    background-size: 36px 64px;
    opacity: 0.14;
    mask-image: radial-gradient(circle at center, #000 0%, transparent 82%);
    pointer-events: none;
  }
}

// ===== 背景层 =====
.bg-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(64, 141, 255, 0.08) 1px, transparent 1px),
    linear-gradient(90deg, rgba(64, 141, 255, 0.08) 1px, transparent 1px);
  background-size: 64px 64px;
  transform: perspective(560px) rotateX(58deg) translateY(-120px) scale(1.2);
  transform-origin: 50% 0;
  mask-image: linear-gradient(180deg, transparent, #000 18%, transparent 92%);
  pointer-events: none;
  z-index: 0;
}
.bg-scan {
  position: absolute;
  inset: 0;
  background: repeating-linear-gradient(
    0deg,
    transparent 0,
    transparent 2px,
    rgba(0, 212, 255, 0.02) 3px,
    transparent 4px
  );
  pointer-events: none;
  z-index: 0;
  opacity: 0.72;
}
.bg-vignette {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse at top, rgba(0, 212, 255, 0.06), transparent 60%),
    radial-gradient(ellipse at bottom, rgba(0, 80, 160, 0.1), transparent 70%);
  pointer-events: none;
  z-index: 0;
}

// ===== 漂浮微光粒子 =====
.bg-particles {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 0;
  overflow: hidden;
}
.bg-particle {
  position: absolute;
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: rgba(0, 212, 255, 0.6);
  box-shadow: 0 0 5px rgba(0, 212, 255, 0.36);
}
.bp-1 { left: 8%; top: 15%; animation: float-particle 12s ease-in-out infinite; }
.bp-2 { left: 22%; top: 70%; animation: float-particle 15s ease-in-out infinite 2s; width: 3px; height: 3px; }
.bp-3 { left: 45%; top: 25%; animation: float-particle 18s ease-in-out infinite 4s; background: rgba(255, 180, 74, 0.5); box-shadow: 0 0 5px rgba(255, 180, 74, 0.26); }
.bp-4 { left: 65%; top: 80%; animation: float-particle 14s ease-in-out infinite 1s; width: 3px; height: 3px; }
@keyframes float-particle {
  0%, 100% { transform: translate(0, 0); opacity: 0.6; }
  25% { transform: translate(15px, -20px); opacity: 1; }
  50% { transform: translate(-10px, -35px); opacity: 0.4; }
  75% { transform: translate(20px, -15px); opacity: 0.8; }
}

// ===== Overlay =====
.overlay-state {
  position: fixed; inset: 0;
  display: flex; align-items: center; justify-content: center;
  z-index: 100; background: $bg-deep;
  &.error .error-content {
    display: flex; flex-direction: column; align-items: center; gap: $spacing-base;
    color: $text-dim;
    .el-icon { color: $danger; }
    p { font-size: $font-size-sm; }
  }
}
.loader {
  text-align: center;
  .loader-ring {
    width: 48px; height: 48px; margin: 0 auto $spacing-base;
    border: 2px solid $line; border-top-color: $neon;
    border-radius: 50%; animation: spin 1s linear infinite;
  }
  .loader-text {
    color: $text-dim; font-size: $font-size-xs;
    letter-spacing: 3px; text-transform: uppercase;
    font-family: $font-display;
  }
}
@keyframes spin { to { transform: rotate(360deg); } }

// ===== 屏幕主容器 =====
.screen-content {
  position: relative;
  z-index: 2;
  display: flex;
  flex-direction: column;
  height: 100vh;
  padding: 12px 18px 6px;
  gap: 8px;
}

.screen-content-fallback-fullscreen {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background:
    radial-gradient(circle at 50% 48%, rgba(0, 96, 205, 0.18), transparent 34%),
    radial-gradient(circle at 74% 68%, rgba(255, 122, 0, 0.08), transparent 18%),
    linear-gradient(180deg, #061229 0%, #020815 100%);
}

// ===== HEADER =====
.screen-header {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 74px;
  background:
    radial-gradient(ellipse at 50% 12%, rgba(37, 132, 255, 0.34), transparent 32%),
    linear-gradient(180deg, rgba(21, 66, 127, 0.34), rgba(6, 24, 57, 0.12) 68%, rgba(0, 212, 255, 0.04)),
    linear-gradient(90deg, rgba(13, 58, 124, 0.28), rgba(4, 18, 44, 0.08) 36%, rgba(4, 18, 44, 0.08) 64%, rgba(13, 58, 124, 0.28));
  border: 0;
  padding: 0 64px;
  box-shadow:
    inset 0 1px 0 rgba(115, 191, 255, 0.36),
    inset 0 -1px 0 rgba(44, 144, 255, 0.38),
    0 8px 28px rgba(0, 56, 120, 0.18);
  overflow: hidden;

  &::before {
    content: '';
    position: absolute;
    inset: 2px 0 auto;
    height: 54px;
    background:
      radial-gradient(ellipse at 50% 18%, rgba(0, 136, 255, 0.3), transparent 38%),
      linear-gradient(90deg, transparent, rgba(0, 212, 255, 0.08) 24%, rgba(45, 130, 255, 0.18) 50%, rgba(0, 212, 255, 0.08) 76%, transparent);
    pointer-events: none;
  }

  &::after {
    content: '';
    position: absolute;
    left: 30px;
    right: 30px;
    bottom: 0;
    height: 1px;
    background: linear-gradient(90deg, transparent, rgba(0, 212, 255, 0.18) 16%, rgba(78, 166, 255, 0.5) 50%, rgba(0, 212, 255, 0.18) 84%, transparent);
    pointer-events: none;
  }

  .header-frame {
    position: absolute;
    inset: 0 24px;
    z-index: 1;
    width: calc(100% - 48px);
    height: 100%;
    pointer-events: none;
  }

  .header-frame-line {
    fill: none;
    stroke: url(#header-line-grad);
    stroke-linecap: square;
    stroke-linejoin: miter;
    vector-effect: non-scaling-stroke;
    filter: url(#header-line-glow);
    pointer-events: none;
  }

  .header-frame-line {
    stroke-width: 2.4;
  }

  .header-frame-line-shadow {
    stroke-width: 8;
    opacity: 0.14;
  }

  .header-frame-line-inner {
    stroke-width: 2.2;
    opacity: 0.76;
  }

  .header-frame-tray {
    fill: none;
    stroke: url(#header-line-grad);
    stroke-width: 2.6;
    stroke-linecap: square;
    vector-effect: non-scaling-stroke;
    filter: url(#header-line-glow);
    opacity: 1;
    pointer-events: none;
  }

  .header-frame-lower-tray {
    fill: none;
    stroke: url(#header-line-grad);
    stroke-width: 2.2;
    stroke-linecap: square;
    stroke-linejoin: miter;
    vector-effect: non-scaling-stroke;
    filter: url(#header-line-glow);
    opacity: 0.92;
    pointer-events: none;
  }

  .header-frame-lower-tray-glow {
    stroke-width: 5.2;
    opacity: 0.26;
  }

  .header-frame-tip {
    fill: none;
    stroke: #d4f9ff;
    stroke-width: 3.6;
    stroke-linecap: round;
    vector-effect: non-scaling-stroke;
    filter: url(#header-line-glow);
    opacity: 0.96;
    pointer-events: none;
  }

  .header-frame-tip-glow {
    stroke-width: 7;
    opacity: 0.22;
  }

  .header-frame-tip-soft {
    fill: none;
    stroke: url(#header-line-grad);
    stroke-width: 2.2;
    stroke-linecap: square;
    stroke-linejoin: miter;
    vector-effect: non-scaling-stroke;
    filter: url(#header-line-glow);
    opacity: 0.92;
    pointer-events: none;
  }

  .header-frame-tip-soft-glow {
    stroke-width: 5.2;
    opacity: 0.26;
  }

  .header-flow {
    fill: none;
    stroke: url(#header-flow-grad);
    stroke-width: 3.2;
    stroke-linecap: round;
    stroke-linejoin: round;
    vector-effect: non-scaling-stroke;
    opacity: 0;
    stroke-dasharray: 16 292;
    pointer-events: none;
    animation: header-flow-to-center 4.8s cubic-bezier(0.42, 0, 0.22, 1) infinite;
  }

  .header-flow-delay-1 { animation-delay: 0s; }
  .header-flow-delay-2 { animation-delay: 0.34s; }
  .header-flow-delay-3 { animation-delay: 0.68s; }

  .header-title-block {
    position: absolute;
    left: 50%;
    top: 0;
    z-index: 2;
    width: min(430px, 42vw);
    min-width: 300px;
    height: 52px;
    transform: translateX(-50%);
    justify-content: center;
    text-align: center;
    display: flex;
    align-items: center;
    pointer-events: auto;
    background:
      radial-gradient(ellipse at 50% 58%, rgba(0, 218, 255, 0.16), transparent 66%),
      linear-gradient(90deg, transparent, rgba(0, 192, 255, 0.12) 22%, rgba(83, 215, 255, 0.22) 50%, rgba(0, 192, 255, 0.12) 78%, transparent);
    border-top: 0;
    border-bottom: 0;
    box-shadow:
      0 0 28px rgba(0, 128, 255, 0.16),
      inset 0 0 22px rgba(0, 114, 255, 0.12);

    &::after {
      content: '';
      position: absolute;
      left: 50%;
      transform: translateX(-50%);
      pointer-events: none;
    }

    &::after {
      display: none;
    }

    .drill-title {
      font-family: $font-cn;
      font-size: clamp(38px, 3vw, 42px);
      font-weight: 900;
      letter-spacing: 5px;
      margin: 0;
      padding-left: 6px;
      color: #ffffff;
      text-shadow: 0 0 10px rgba(64, 170, 255, 0.8);
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      line-height: 1;
      pointer-events: auto;
    }
    .drill-title-en {
      display: block;
      margin-top: 0;
      font-family: $font-display;
      font-size: clamp(8px, 0.9em, 15px);
      font-weight: 700;
      letter-spacing: 7px;
      color: rgba(194, 214, 255, 0.66);
      white-space: nowrap;
    }
  }
  .header-meta {
    margin-left: auto;
    position: relative;
    z-index: 4;
    display: flex; align-items: center; gap: 10px;
    font-family: $font-mono;
    .meta-label { color: $text-dim; font-size: 0.88em; letter-spacing: 2px; font-weight: 700; }
    .meta-divider { color: $text-mute; }
    .meta-value {
      color: #ecf6ff;
      font-size: clamp(14px, 1.35em, 22px);
      font-weight: 700;
      text-shadow: 0 0 8px rgba(95, 171, 255, 0.6);
      letter-spacing: 1px;
    }
  }
  .btn-icon {
    position: absolute;
    right: 34px;
    top: 50%;
    transform: translateY(-50%);
    z-index: 4;
    background: transparent; border: 1px solid $line;
    color: $neon; width: 34px; height: 34px;
    display: flex; align-items: center; justify-content: center;
    cursor: pointer; border-radius: 2px;
    transition: all 0.2s;
    &:hover,
    &.active { border-color: $neon; box-shadow: 0 0 10px $neon-soft; background: rgba(0, 212, 255, 0.1); }
  }

  .btn-fullscreen-mark {
    top: auto;
    right: 22px;
    bottom: 11px;
    width: 44px;
    height: 36px;
    padding: 0;
    transform: none;
    border: 0;
    background: transparent;
    box-shadow: none;

    &:hover,
    &.active {
      border: 0;
      background: transparent;
      box-shadow: none;
    }
  }

  .fullscreen-mark {
    width: 42px;
    height: 28px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-family: $font-display;
    font-size: 31px;
    font-weight: 900;
    font-style: italic;
    line-height: 1;
    letter-spacing: 0;
    color: rgba(255, 255, 255, 0.98);
    text-shadow:
      0 0 4px rgba(255, 255, 255, 0.58),
      0 0 9px rgba(46, 225, 255, 0.5);
    filter:
      drop-shadow(0 0 4px rgba(255, 255, 255, 0.5))
      drop-shadow(0 0 7px rgba(46, 225, 255, 0.38));
  }

  .fullscreen-mark-image {
    display: block;
    object-fit: contain;
  }
  .header-pulse-line {
    position: absolute;
    bottom: 0; left: 0;
    width: 100%; height: 2px;
    background: linear-gradient(90deg, transparent, $neon, transparent);
    transform: scaleX(0);
    animation: header-pulse 4.8s ease-in-out infinite;
  }

  .header-brand {
    position: absolute;
    right: 68px;
    bottom: 12px;
    z-index: 4;
    font-family: $font-cn;
    font-size: 16px;
    font-weight: 700;
    letter-spacing: 2px;
    color: #f0f7ff;
    text-shadow:
      0 0 8px rgba(240, 247, 255, 0.42),
      0 0 14px rgba(0, 212, 255, 0.28);
    pointer-events: none;
    user-select: none;
  }
}
@keyframes deco-flicker {
  0%, 100% { opacity: 0.7; }
  50% { opacity: 0.35; }
}
@keyframes header-pulse {
  0% { transform: scaleX(0); opacity: 0; }
  50% { transform: scaleX(1); opacity: 1; }
  100% { transform: scaleX(0); opacity: 0; }
}
@keyframes header-flow-to-center {
  0% {
    opacity: 0;
    stroke-dashoffset: 260;
  }
  16% { opacity: 0.78; }
  54% {
    opacity: 0.86;
    stroke-dashoffset: 0;
  }
  72% { opacity: 0; }
  100% {
    opacity: 0;
    stroke-dashoffset: -54;
  }
}
@media (prefers-reduced-motion: reduce) {
  .header-flow {
    animation: none;
    opacity: 0.18;
    stroke-dashoffset: 0;
  }
}

// ===== KPI Row =====
.kpi-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  height: 88px;
  flex-shrink: 0;
  padding: 0;
}
.kpi-card {
  position: relative;
  background:
    linear-gradient(90deg, rgba(14, 50, 112, 0.82), rgba(10, 26, 57, 0.36)),
    linear-gradient(180deg, rgba(53, 138, 255, 0.18), transparent);
  border: 1px solid rgba(105, 165, 255, 0.38);
  padding: 12px 18px 10px 80px;
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(72px, max-content);
  grid-template-rows: 1fr;
  column-gap: 10px;
  align-items: center;
  overflow: hidden;
  transition: all 0.3s;
  clip-path: polygon(0 0, calc(100% - 18px) 0, 100% 18px, 100% 100%, 18px 100%, 0 calc(100% - 18px));

  // 角装饰
  &::before, &::after {
    content: '';
    position: absolute; width: 10px; height: 10px;
    border: 1px solid $neon;
    animation: corner-flicker 4s ease-in-out infinite;
  }
  &::before { top: -1px; left: -1px; border-right: 0; border-bottom: 0; }
  &::after { bottom: -1px; right: -1px; border-left: 0; border-top: 0; animation-delay: 2s; }

  .kpi-orb {
    position: absolute;
    left: 24px;
    top: 14px;
    width: 34px;
    height: 34px;
    border-radius: 50%;
    background:
      radial-gradient(circle at 36% 34%, #9ffcff 0 14%, #00d4ff 15% 28%, rgba(20, 255, 189, 0.85) 29% 45%, rgba(0, 74, 165, 0.7) 46% 100%);
    box-shadow:
      0 0 14px rgba(0, 212, 255, 0.85),
      0 0 36px rgba(0, 212, 255, 0.22);

    &::before {
      content: '';
      position: absolute;
      inset: 7px -10px;
      border: 1px solid rgba(0, 212, 255, 0.64);
      border-radius: 50%;
      transform: rotate(-8deg);
    }

    &::after {
      content: '';
      position: absolute;
      left: 12px;
      bottom: -17px;
      width: 15px;
      height: 15px;
      border-radius: 50%;
      border: 2px solid rgba(0, 212, 255, 0.54);
      box-shadow: 0 0 10px rgba(0, 212, 255, 0.55);
    }
  }
  .kpi-copy {
    grid-column: 1;
    grid-row: 1;
    display: flex;
    align-items: center;
    min-width: 0;
    height: 24px;
    overflow: hidden;
  }
  .kpi-label-zh {
    display: block;
    position: relative;
    max-width: 100%;
    font-size: clamp(13px, 1.15em, 19px);
    line-height: 1;
    font-weight: 900;
    color: #f0f7ff;
    letter-spacing: 0.6px;
    padding-bottom: 5px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    &::after {
      content: '';
      position: absolute;
      left: 0;
      bottom: 0;
      width: 24px;
      height: 2px;
      border-radius: 1px;
      background: linear-gradient(90deg, #2cf8d8, transparent);
    }
  }
  .kpi-label-en {
    font-family: $font-display;
    display: block;
    margin-top: 6px;
    font-size: 10px;
    font-weight: 600;
    color: rgba(196, 214, 255, 0.55);
    letter-spacing: 4px;
    opacity: 1;
  }
  .kpi-value-row {
    grid-column: 2;
    grid-row: 1;
    justify-self: end;
    align-self: center;
    margin-top: 0;
    max-width: 100%;
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 6px;
    min-width: 0;
    overflow: hidden;
    &.mono { gap: 2px; }
  }
  .kpi-status-row {
    align-items: center;
  }
  .kpi-progress-row {
    align-items: center;
  }
  .kpi-value-num {
    font-family: $font-mono;
    font-size: clamp(20px, 2.1em, 36px);
    font-weight: 800;
    color: #2cf8d8;
    text-shadow: 0 0 12px rgba(44, 248, 216, 0.42);
    line-height: 1;
  }
  .kpi-value-unit {
    font-family: $font-cn; font-size: clamp(13px, 1.1em, 19px); color: #c8fff5; opacity: 0.92; font-weight: 800;
  }
  .kpi-value-sep {
    font-family: $font-mono; font-size: clamp(15px, 1.6em, 28px); color: #2cf8d8; opacity: 0.75;
    transform: translateY(-2px);
  }
  .kpi-value-text {
    display: block;
    max-width: clamp(66px, 5.8vw, 112px);
    font-size: clamp(15px, 1.4em, 23px); font-weight: 900; color: #ff8a1f;
    text-align: right;
    text-shadow: 0 0 12px rgba(255, 122, 0, 0.5);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .kpi-node-count {
    margin-left: 10px;
    color: #dff6ff;
    font-family: $font-mono;
    font-size: 14px;
  }

  // === Progress card specific: ring + node count layout ===
  &.kpi-progress-card {
    .kpi-value-row.kpi-progress-row {
      align-items: center;
      gap: 12px;
    }
  }
  .kpi-queue-row {
    align-items: center;
    gap: 10px;
  }
  .progress-ring-wrap {
    position: relative;
    width: 56px;
    height: 56px;
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .progress-ring-svg {
    width: 100%;
    height: 100%;
    transform: rotate(-90deg);
  }
  .progress-ring-bg {
    fill: none;
    stroke: rgba(255, 255, 255, 0.12);
    stroke-width: 6;
  }
  .progress-ring-fill {
    fill: none;
    stroke: url(#ring-grad);
    stroke-width: 6;
    stroke-linecap: round;
    stroke-dasharray: 289.03;
    transition: stroke-dashoffset 0.8s ease;
  }
  .progress-ring-text {
    position: absolute;
    font-family: $font-mono;
    font-size: clamp(16px, 1.4em, 24px);
    font-weight: 800;
    color: #fff;
    line-height: 1;
    small {
      font-size: 0.65em;
      font-weight: 600;
      opacity: 0.75;
    }
  }
  .progress-node-block {
    display: flex;
    align-items: center;
  }
  .node-count-row {
    display: flex;
    align-items: baseline;
    gap: 3px;
    .node-separator {
      font-family: $font-mono;
      font-size: 14px;
      color: rgba(255, 255, 255, 0.45);
    }
    .node-total {
      font-family: $font-mono;
      font-size: 16px;
      color: rgba(255, 255, 255, 0.55);
      font-weight: 600;
    }
  }
  .status-dot {
    width: 10px; height: 10px; border-radius: 50%;
    background: $ok; box-shadow: 0 0 10px $ok;
    animation: pulse 1.6s ease-in-out infinite;
    &.dot-paused { background: $warn; box-shadow: 0 0 10px $warn; }
    &.dot-completed { background: $neon; box-shadow: 0 0 10px $neon; }
    &.dot-terminated { background: $danger; box-shadow: 0 0 10px $danger; animation: none; }
    &.dot-pending { background: $text-mute; box-shadow: none; animation: none; }
  }
  .txt-running { color: $ok; }
  .txt-paused { color: $warn; }
  .txt-completed {
    color: #2cf8d8;
    text-shadow: 0 0 12px rgba(44, 248, 216, 0.6), 0 0 24px rgba(0, 255, 156, 0.35);
  }
  .txt-terminated { color: $danger; }
  .txt-pending { color: $text-dim; }

  // === Alert Card - Abnormal Tasks ===
  &.kpi-alert-card {
    // 去掉边框与切角，让数字不受框线包围
    border: none;
    clip-path: none;
    &::before, &::after { display: none; }

    // 保持与其他卡片一致的背景与双列布局，确保标签水平线对齐
    .kpi-alert-row {
      grid-column: 2;
      grid-row: 1;
      justify-self: center; // 偏右但不贴边
      align-self: center;
    }

    // 安全状态 - orb 与其他卡片保持一致（青蓝色）
    &.alert-level-safe {
      .alert-value-block {
        background: transparent;
        border: none;
        box-shadow: none;
      }

      .kpi-orb-alert {
        background:
          radial-gradient(circle at 36% 34%, #9ffcff 0 14%, #00d4ff 15% 28%, rgba(20, 255, 189, 0.85) 29% 45%, rgba(0, 74, 165, 0.7) 46% 100%);
        box-shadow:
          0 0 14px rgba(0, 212, 255, 0.85),
          0 0 36px rgba(0, 212, 255, 0.22);
      }
    }

    // 警告状态 - 黄色 orb（背景不变）
    &.alert-level-warning {
      .kpi-orb-alert {
        background:
          radial-gradient(circle at 36% 34%, #fff5bd 0 14%, #ffd36f 15% 28%, rgba(255, 180, 74, 0.85) 29% 45%, rgba(130, 80, 20, 0.7) 46% 100%);
        box-shadow:
          0 0 14px rgba(255, 180, 74, 0.85),
          0 0 36px rgba(255, 180, 74, 0.28);
        animation: orb-pulse-yellow 2s ease-in-out infinite;
      }
    }

    // 危险状态 - 红色 orb（背景不变）
    &.alert-level-danger {
      .kpi-orb-alert {
        background:
          radial-gradient(circle at 36% 34%, #ff9f9f 0 14%, #ff6b6b 15% 28%, rgba(255, 85, 85, 0.85) 29% 45%, rgba(150, 30, 30, 0.7) 46% 100%);
        box-shadow:
          0 0 14px rgba(255, 85, 85, 0.85),
          0 0 36px rgba(255, 85, 85, 0.28);
        animation: orb-pulse-red 1.6s ease-in-out infinite;
      }
    }
  }
}

// Alert Value Styles
.alert-value-block {
  display: flex;
  align-items: center;
  justify-content: flex-end; // 靠右，与右栏 center 对齐形成偏右效果
  width: 100%;
}

.alert-num {
  font-size: clamp(30px, 2.8em, 46px);
  font-weight: 900;
  letter-spacing: -1px;
  background: transparent;
  border: none;
  box-shadow: none;

  &.num-safe {
    color: #7dffc6;
    text-shadow: none;
  }

  &.num-warning {
    color: #ffd36f;
    text-shadow: 0 0 16px rgba(255, 180, 74, 0.72);
  }

  &.num-danger {
    color: #ff5c5c;
    text-shadow: 0 0 20px rgba(255, 85, 85, 0.88);
    animation: num-pulse-red 1.2s ease-in-out infinite;
  }
}

.alert-detail-hint {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 4px;
  padding: 6px 10px;
  background: rgba(255, 85, 85, 0.08);
  border-radius: 6px;
  border: 1px solid rgba(255, 85, 85, 0.18);
}

.hint-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #ff5c5c;
  box-shadow: 0 0 8px rgba(255, 85, 85, 0.65);
  animation: hint-dot-pulse 1.4s ease-in-out infinite;
}

.hint-text {
  font-size: 12px;
  color: rgba(255, 200, 200, 0.85);
  letter-spacing: 0.5px;
}

@keyframes orb-pulse-yellow {
  0%, 100% {
    box-shadow: 0 0 14px rgba(255, 180, 74, 0.85), 0 0 36px rgba(255, 180, 74, 0.28);
  }
  50% {
    box-shadow: 0 0 20px rgba(255, 180, 74, 1), 0 0 46px rgba(255, 180, 74, 0.45);
  }
}

@keyframes orb-pulse-red {
  0%, 100% {
    box-shadow: 0 0 14px rgba(255, 85, 85, 0.85), 0 0 36px rgba(255, 85, 85, 0.28);
  }
  50% {
    box-shadow: 0 0 22px rgba(255, 85, 85, 1), 0 0 50px rgba(255, 85, 85, 0.55);
  }
}

@keyframes num-pulse-red {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.05); }
}

@keyframes hint-dot-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.5; transform: scale(0.85); }
}
@keyframes corner-flicker {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

// ===== MAIN GRID =====
.screen-main {
  display: grid;
  grid-template-columns: minmax(220px, 17vw) minmax(0, 1.18fr) minmax(280px, 24vw);
  gap: 16px;
  flex: 1;
  min-height: 0;
  padding: 0;
}

.screen-root:fullscreen .screen-main,
.screen-content-fallback-fullscreen .screen-main {
  grid-template-columns: minmax(220px, 17vw) minmax(0, 1.18fr) minmax(280px, 24vw);
  gap: 16px;
}

.panel {
  position: relative;
  display: flex; flex-direction: column;
  background: $bg-card;
  border: 0;
  overflow: hidden;
}
.panel-header {
  position: relative;
  display: flex; align-items: center; gap: 8px;
  height: 44px;
  padding: 0 16px 0 28px;
  background:
    linear-gradient(90deg, rgba(0, 116, 255, 0.62), rgba(8, 29, 67, 0.38) 52%, transparent 100%);
  border-bottom: 0;
  flex-shrink: 0;
  overflow: hidden;
  .panel-deco-corner {
    position: absolute; width: 8px; height: 8px; border-color: $neon; border-style: solid; border-width: 0;
    &.tl { top: 0; left: 0; border-top-width: 2px; border-left-width: 2px; }
    &.tr { top: 0; right: 0; border-top-width: 2px; border-right-width: 2px; }
  }
  .panel-title-zh {
    font-size: clamp(16px, 1.6em, 28px);
    font-weight: 900;
    color: #ffffff;
    letter-spacing: 2px;
    text-shadow: 0 0 10px rgba(64, 170, 255, 0.8);
    // 标题承载动态环节名，超长时省略并加提示，避免挤压右侧"实时"标签
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .panel-title-en {
    font-family: $font-display;
    font-size: 10px; color: $neon; opacity: 0.7;
    letter-spacing: 2px;
  }
  .panel-badge {
    margin-left: auto;
    background: $neon-soft; border: 1px solid $neon;
    color: $neon; font-family: $font-mono;
    font-size: 16px; line-height: 1; padding: 0 8px; height: 24px;
    display: inline-flex; align-items: center;
  }
  .panel-realtime {
    margin-left: auto;
    display: inline-flex; align-items: center; gap: 4px;
    font-size: 17px; line-height: 1;
    transform: scaleY(0.94);
    transform-origin: center;
    color: $ok; font-family: $font-display;
    letter-spacing: 1px;
    .rt-dot {
      width: 6px; height: 6px; border-radius: 50%;
      background: $ok; box-shadow: 0 0 6px $ok;
      animation: pulse 1.4s ease-in-out infinite;
    }
  }
  // 面板流光扫描线
  .panel-scan-line {
    position: absolute;
    top: 0; left: -100%;
    width: 60%;
    height: 100%;
    background: linear-gradient(90deg, transparent, rgba(0, 212, 255, 0.15), transparent);
    animation: panel-scan 7s ease-in-out infinite;
    pointer-events: none;
  }
}
@keyframes panel-scan {
  0% { transform: translateX(0); }
  100% { transform: translateX(350%); }
}
.panel-body {
  flex: 1;
  overflow-y: auto;
  padding: 14px;
  min-height: 0;
  &::-webkit-scrollbar { width: 4px; }
  &::-webkit-scrollbar-track { background: transparent; }
  &::-webkit-scrollbar-thumb { background: $line-strong; border-radius: 2px; }
}
.empty-tip {
  text-align: center; color: $text-mute;
  font-size: 12px; padding: 30px 0;
  font-family: $font-mono;
}
.more-tip {
  text-align: center; color: #f5f9ff;
  font-size: 18px; padding: 8px 0 16px;
  font-family: $font-mono;
  letter-spacing: 1px;
  border-top: 1px dashed $line;
  text-shadow: 0 0 12px rgba(0, 180, 255, 0.55);
}

// ===== Left stages =====
.panel-stages {
  background: transparent;
  min-height: 0;
  .stages-list {
    display: flex; flex-direction: column; gap: 16px;
    min-height: 0;
    flex: 1;
    overflow-y: auto;
    padding-right: 34px;
    scroll-snap-type: y proximity;
  }
  .stage-card {
    position: relative;
    background: linear-gradient(90deg, rgba(15, 32, 62, 0.94), rgba(9, 22, 48, 0.76));
    border: 1px solid rgba(106, 157, 215, 0.34);
    border-left: 5px solid rgba(160, 200, 242, 0.7);
    padding: 14px 16px;
    flex: 0 0 calc((100% - 48px) / 4);
    min-height: 112px;
    scroll-snap-align: start;
    display: flex; flex-direction: column; gap: 10px;
    justify-content: center;
    transition: all 0.2s;
    cursor: pointer;

    &:focus-visible {
      outline: 2px solid rgba(45, 228, 255, 0.78);
      outline-offset: 2px;
    }

    &.stage-done {
      background: linear-gradient(90deg, rgba(9, 58, 42, 0.94), rgba(7, 31, 35, 0.78));
      border-color: rgba(52, 255, 151, 0.46);
      border-left-color: #34ff97;
      box-shadow: 0 0 18px rgba(52, 255, 151, 0.18), inset 0 0 18px rgba(52, 255, 151, 0.09);
    }
    &.stage-running {
      background: linear-gradient(90deg, rgba(86, 43, 8, 0.95), rgba(34, 25, 20, 0.78));
      border-color: #ff9a2f;
      border-left-color: #ff7a00;
      box-shadow: inset 0 0 18px rgba(255, 122, 0, 0.12);
    }
    &.stage-current {
      transform: translateX(10px);
      border-right: 1px solid rgba(255, 224, 162, 0.7);
      background:
        linear-gradient(90deg, rgba(86, 43, 8, 0.96), rgba(12, 38, 66, 0.84)),
        radial-gradient(circle at 100% 50%, rgba(45, 228, 255, 0.22), transparent 34%);
      box-shadow:
        16px 0 28px rgba(45, 228, 255, 0.12),
        inset 0 0 22px rgba(255, 122, 0, 0.13);

      &::after {
        content: '';
        position: absolute;
        top: 50%;
        right: -19px;
        width: 48px;
        height: 8px;
        border-radius: 999px;
        background:
          linear-gradient(90deg, rgba(45, 228, 255, 0.18), rgba(125, 255, 198, 0.88) 55%, rgba(125, 255, 198, 0.96)) 0 50% / 100% 2px no-repeat;
        box-shadow: 0 0 12px rgba(45, 228, 255, 0.58), 0 0 18px rgba(125, 255, 198, 0.18);
        transform: translateY(-50%);
        z-index: 1;
      }

      &::before {
        content: '';
        position: absolute;
        top: 50%;
        right: -23px;
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background:
          radial-gradient(circle at 50% 50%, #7dffc6 0 3px, rgba(45, 228, 255, 0.86) 3.2px 4px, transparent 4.2px);
        clip-path: inset(0 50% 0 0);
        box-shadow: 0 0 10px rgba(125, 255, 198, 0.72), 0 0 18px rgba(45, 228, 255, 0.42);
        transform: translateY(-50%);
        z-index: 2;
      }
    }
    &.stage-current.stage-done {
      border-color: rgba(52, 255, 151, 0.5);
      border-left-color: #34ff97;
      border-right-color: rgba(125, 255, 198, 0.7);
      background:
        linear-gradient(90deg, rgba(9, 68, 46, 0.96), rgba(7, 36, 38, 0.84)),
        radial-gradient(circle at 100% 50%, rgba(45, 228, 255, 0.16), transparent 34%);
      box-shadow:
        0 0 20px rgba(52, 255, 151, 0.22),
        16px 0 24px rgba(45, 228, 255, 0.1),
        inset 0 0 20px rgba(52, 255, 151, 0.12);
    }
    &.stage-issue { border-left-color: $danger; }
  }
  .stage-card-top {
    display: flex; align-items: center; justify-content: space-between; gap: 12px;
    min-width: 0;
    flex-shrink: 0;
    .stage-name-block { display: flex; align-items: baseline; gap: 8px; min-width: 0; flex: 1; }
    .stage-index {
      font-family: $font-mono; font-size: 17px;
      color: #ffffff; font-weight: 900;
      text-shadow: 0 0 8px rgba(255, 255, 255, 0.42);
    }
    .stage-name {
      min-width: 0;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      font-size: clamp(14px, 1.2em, 20px); color: #f2f8ff; font-weight: 900;
    }
  }
  .stage-badge {
    flex-shrink: 0;
    font-size: clamp(12px, 0.95em, 16px); padding: 5px 12px;
    border-radius: 2px; letter-spacing: 1.4px;
    font-weight: 900;
    &.badge-done {
      background: rgba(52, 255, 151, 0.18);
      color: #d8ffe9;
      border: 1px solid rgba(52, 255, 151, 0.72);
      box-shadow: 0 0 12px rgba(52, 255, 151, 0.28), inset 0 0 10px rgba(52, 255, 151, 0.12);
      text-shadow: 0 0 8px rgba(52, 255, 151, 0.62);
    }
    &.badge-running { background: rgba(255, 122, 0, 0.14); color: #ffc179; border: 1px solid rgba(255, 122, 0, 0.62); }
    &.badge-pending { background: rgba(110, 141, 181, 0.15); color: $text-dim; border: 1px solid $line; }
    &.badge-issue { background: rgba(255, 77, 106, 0.18); color: $danger; border: 1px solid rgba(255, 77, 106, 0.4); }
  }
  .stage-segments {
    display: flex; gap: 3px; height: 9px;
    flex-shrink: 0;
    .segment { flex: 1; background: rgba(255, 255, 255, 0.05); }
    .segment.seg-done { background: linear-gradient(180deg, #b8ffd4, #34ff97 48%, #00b86b); box-shadow: 0 0 6px rgba(52, 255, 151, 0.54); }
    .segment.seg-active { background: linear-gradient(180deg, #ff9a2f, #ff7a00); box-shadow: 0 0 6px #ff7a00; animation: pulse 1.2s ease-in-out infinite; }
    .segment.seg-todo { background: rgba(110, 141, 181, 0.25); }
    .segment.seg-empty { background: transparent; border: 1px dashed rgba(110, 141, 181, 0.2); }
  }
  .stage-card-bottom {
    display: flex; justify-content: space-between;
    flex-shrink: 0;
    .stage-meta {
      display: flex; flex-direction: column; gap: 1px;
      .meta-key {
        font-size: clamp(13px, 1.02em, 17px);
        color: rgba(220, 248, 255, 0.98);
        letter-spacing: 1.8px;
        font-weight: 800;
        text-shadow: 0 0 8px rgba(0, 212, 255, 0.42);
      }
      .meta-val { font-family: $font-mono; font-size: clamp(14px, 1.1em, 18px); color: #f1fbff; font-weight: 800; }
    }
  }
}

// ===== Center phase ring =====
.panel-center {
  background:
    radial-gradient(circle at center, rgba(4, 18, 49, 0.38), transparent 74%);
  position: relative;
  overflow: hidden;
  &::before, &::after {
    content: '';
    position: absolute; width: 14px; height: 14px;
    border: 1px solid $neon;
  }
  &::before { top: -1px; left: -1px; border-right: 0; border-bottom: 0; }
  &::after { bottom: -1px; right: -1px; border-left: 0; border-top: 0; }
}
.center-stage {
  flex: 1;
  display: flex; align-items: stretch; justify-content: center;
  position: relative;
  min-height: 0;
  padding-left: 0;
  padding-inline: 0;
  &::before, &::after {
    content: '';
    position: absolute;
    inset: 10px 14px;
    pointer-events: none;
    border: 1px solid rgba(0, 212, 255, 0.08);
    background:
      linear-gradient(rgba(0, 212, 255, 0.045) 1px, transparent 1px),
      linear-gradient(90deg, rgba(0, 212, 255, 0.045) 1px, transparent 1px);
    background-size: 44px 44px;
  }
  &.center-stage-0::after,
  &.center-stage-1::after {
    display: none;
  }
  &.center-stage-2::after,
  &.center-stage-3::after {
    display: none;
  }
}


// ===== Right column =====
.panel-right {
  display: flex; flex-direction: column;
  background: transparent; border: none; padding: 0;
  gap: clamp(8px, 1vh, 14px);
  min-height: 0;
  .sub-panel {
    position: relative;
    background: transparent; border: 0;
    display: flex; flex-direction: column;
    min-height: 0;
    flex: 1;
  }
  // 执行日志为固定占比的独立面板，不与执行中步骤争抢空间
  .sub-log {
    flex: 0 0 clamp(178px, 24vh, 258px);
  }
}

// ===== Alerts =====
.warn-list {
  display: flex; flex-direction: column;
  overflow: hidden;
  flex: 1;
  min-height: 0;
  .alert-scroll {
    flex: 1;
    min-height: 0;
    display: flex; flex-direction: column;
    gap: var(--alert-card-gap, 9px);
    overflow: hidden;
  }
  .alert-card {
    position: relative;
    background: linear-gradient(135deg, rgba(15, 30, 58, 0.96), rgba(8, 18, 40, 0.9));
    border: 1px solid rgba(0, 212, 255, 0.24);
    border-radius: 3px;
    padding: clamp(10px, 1vh, 13px) clamp(12px, 0.9vw, 16px) clamp(10px, 1vh, 13px) clamp(14px, 1vw, 18px);
    display: grid;
    grid-template-rows: auto auto auto;
    align-content: center;
    gap: clamp(6px, 0.7vh, 8px);
    overflow: hidden;
    flex: 1 1 calc((100% - (var(--visible-alert-count) - 1) * var(--alert-card-gap)) / var(--visible-alert-count));
    min-height: 102px;
    max-height: 132px;

    // 顶部状态指示条（替代左边框）
    &::before {
      content: '';
      position: absolute;
      top: 0; left: 0; right: 0;
      height: 2px;
      background: linear-gradient(90deg, $warn, rgba(255, 182, 72, 0.2));
    }

    &.alert-danger::before {
      background: linear-gradient(90deg, $danger, rgba(255, 77, 106, 0.2));
    }
    &.alert-info::before {
      background: linear-gradient(90deg, $neon, rgba(0, 212, 255, 0.15));
    }

    // 运行中卡片突出显示：橙色辉光描边 + 呼吸动效，与跑道运行节点同色系
    &.alert-warn {
      border-color: #ff9a2f;
      background: linear-gradient(135deg, rgba(58, 28, 6, 0.96), rgba(20, 12, 8, 0.92));
      box-shadow:
        inset 0 0 22px rgba(255, 122, 0, 0.14),
        0 0 16px rgba(255, 154, 47, 0.28);
      animation: alert-running-breathe 2.2s ease-in-out infinite;
      &::before {
        background: linear-gradient(90deg, #ffc179, rgba(255, 154, 47, 0.35));
        height: 3px;
        box-shadow: 0 0 8px rgba(255, 154, 47, 0.8);
      }
      .alert-indicator {
        background: #ffb75e;
        box-shadow: 0 0 8px rgba(255, 183, 94, 0.95);
      }
    }
    &.alert-danger {
      border-color: rgba(255, 77, 106, 0.42);
      box-shadow: inset 0 0 22px rgba(255, 77, 106, 0.1);
    }
    &.alert-info {
      border-color: rgba(0, 212, 255, 0.28);
    }
    // 已完成 / 待执行卡片配色（与跑道节点完成绿/待定蓝呼应）
    &.alert-ok {
      border-color: rgba(73, 255, 166, 0.4);
      &::before { background: linear-gradient(90deg, #49ffa6, rgba(73, 255, 166, 0.15)); }
    }
    &.alert-pending {
      border-color: rgba(123, 167, 201, 0.3);
      &::before { background: linear-gradient(90deg, #7ba7c9, rgba(123, 167, 201, 0.12)); }
      // 待执行指示灯：灰蓝色呼吸，与待执行徽标同色系（覆盖默认橘黄）
      .alert-indicator {
        background: #7ba7c9;
        box-shadow: 0 0 6px rgba(123, 167, 201, 0.65);
        animation: indicator-pulse-cool 2.4s ease-in-out infinite;
      }
    }
  }
  .alert-head {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr) auto;
    align-items: center;
    gap: 8px;
    .alert-indicator {
      width: 8px; height: 8px;
      border-radius: 50%;
      background: $warn;
      box-shadow: 0 0 6px rgba(255, 182, 72, 0.7);
      flex-shrink: 0;
      animation: indicator-pulse 1.6s ease-in-out infinite;
    }
    .alert-title {
      min-width: 0;
      font-size: clamp(16px, 1vw, 20px); color: #f5f9ff; font-weight: 900;
      white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
      letter-spacing: 0;
    }
    .alert-status-badge {
      font-family: $font-cn; font-size: clamp(12px, 0.8vw, 14px); line-height: 1; font-weight: 900;
      padding: 3px 8px; letter-spacing: 0.6px;
      border-radius: 2px; white-space: nowrap; flex-shrink: 0;
      &.badge-warn { color: $warn; background: rgba(255, 182, 72, 0.12); border: 1px solid rgba(255, 182, 72, 0.3); }
      &.badge-danger { color: $danger; background: rgba(255, 77, 106, 0.1); border: 1px solid rgba(255, 77, 106, 0.3); animation: badge-danger-flash 1s ease-in-out infinite; }
      &.badge-info { color: $neon; background: rgba(0, 212, 255, 0.08); border: 1px solid rgba(0, 212, 255, 0.25); }
      &.badge-ok { color: #7dffc6; background: rgba(73, 255, 166, 0.1); border: 1px solid rgba(73, 255, 166, 0.32); }
      &.badge-pending { color: $text-dim; background: rgba(110, 141, 181, 0.12); border: 1px solid $line; }
    }
  }
  .alert-foot {
    display: flex;
    align-items: center;
    gap: clamp(12px, 1vw, 18px);
    row-gap: 2px;
    flex-wrap: wrap;
    min-height: 18px;
    overflow: hidden;
    .alert-meta {
      flex: 0 0 auto;
      min-width: max-content;
      display: flex; align-items: center; gap: 3px;
      .meta-icon {
        font-size: 10px;
        color: rgba(0, 212, 255, 0.7);
        line-height: 1;
      }
      .meta-label {
        position: relative;
        display: inline-flex;
        align-items: center;
        font-size: clamp(12px, 0.82vw, 14px);
        color: rgba(222, 246, 255, 0.96);
        letter-spacing: 1.2px;
        font-weight: 800;
        white-space: nowrap;
        text-shadow: 0 0 8px rgba(0, 212, 255, 0.38);
        &::after {
          content: '';
          width: 1px;
          height: 12px;
          margin: 0 7px;
          border-radius: 1px;
          background: linear-gradient(180deg, transparent, rgba(44, 248, 216, 0.9), transparent);
          box-shadow: 0 0 8px rgba(44, 248, 216, 0.55);
        }
      }
      .meta-val {
        font-family: $font-mono;
        font-size: clamp(12px, 0.82vw, 14px);
        color: rgba(232, 244, 255, 0.95);
        font-weight: 800;
        max-width: clamp(104px, 8vw, 150px);
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
      .operator-val {
        color: rgba(44, 248, 216, 0.95);
      }
    }
  }
  // 剩余时间进度条：横向能量条 + 等宽字体倒计时
  .alert-countdown {
    display: flex; align-items: center; gap: 10px;
    min-height: 20px;
    padding-top: clamp(5px, 0.6vh, 7px);
    padding-bottom: 2px;
    border-top: 1px solid rgba(0, 212, 255, 0.1);

    .countdown-bar {
      flex: 1;
      height: 6px;
      border-radius: 3px;
      background: rgba(8, 22, 42, 0.9);
      border: 1px solid rgba(0, 212, 255, 0.18);
      overflow: hidden;
      position: relative;
    }
    .countdown-fill {
      height: 100%;
      border-radius: 2px;
      background: linear-gradient(90deg, #00d4ff, #2cf7d8);
      box-shadow: 0 0 8px rgba(0, 212, 255, 0.65);
      transition: width 0.6s ease;
    }
    .countdown-text {
      flex-shrink: 0;
      font-family: $font-mono;
      font-size: clamp(11px, 0.72vw, 12.5px);
      font-weight: 700;
      letter-spacing: 0.5px;
      color: rgba(44, 248, 216, 0.95);
      text-shadow: 0 0 8px rgba(44, 248, 216, 0.35);
      white-space: nowrap;
    }

    // 超时态：满格后保持青蓝渐变并轻微呼吸辉光（仅视觉提醒，不改变任务状态）
    &.is-urgent {
      .countdown-fill {
        box-shadow: 0 0 12px rgba(0, 212, 255, 0.85);
        animation: countdown-urgent-pulse 1.2s ease-in-out infinite;
      }
    }

    // 待执行态：不计时，进度条空置，文本弱化显示"未开始"
    &.is-idle {
      .countdown-text {
        color: $text-dim;
        text-shadow: none;
      }
    }
  }
  .empty-tip,
  .more-tip {
    min-height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    color: rgba(160, 205, 245, 0.62);
    font-family: $font-mono;
    font-size: clamp(14px, 0.95vw, 16px);
    letter-spacing: 0.08em;
    border-top: 1px dashed rgba(0, 212, 255, 0.16);
    background: linear-gradient(90deg, transparent, rgba(0, 120, 190, 0.08), transparent);
  }
  .more-tip {
    margin-top: auto;
  }
}

// ===== 执行日志（独立面板的日志视图区） =====
.exec-log {
  display: flex; flex-direction: column;
  padding: 8px 4px 8px 2px;

  .exec-log-view {
    flex: 1;
    min-height: 0;
    overflow: hidden;
    position: relative;
    border: 1px solid rgba(0, 212, 255, 0.16);
    border-radius: 3px;
    background:
      linear-gradient(180deg, rgba(5, 15, 32, 0.9), rgba(3, 10, 24, 0.95));

    // 顶部渐隐遮罩：超出一屏时最旧的日志在顶部淡出
    &::before {
      content: '';
      position: absolute;
      top: 0; left: 0; right: 0;
      height: 16px;
      z-index: 1;
      pointer-events: none;
      background: linear-gradient(180deg, rgba(3, 10, 24, 0.95), transparent);
    }
  }

  // 贴底驻留：日志整体沉底，最新一条永远贴着可视区底部
  .exec-log-track {
    position: absolute;
    inset: 0;
    display: flex; flex-direction: column;
    justify-content: flex-end;
  }

  .log-line {
    display: flex; align-items: center; gap: 10px;
    height: 26px;
    padding: 0 12px;
    font-size: clamp(12px, 0.78vw, 13.5px);
    line-height: 1;
    white-space: nowrap;

    // 偶数行微弱底色，终端行码感
    &:nth-child(even) {
      background: rgba(0, 212, 255, 0.035);
    }

    // 最新一条：入场动效（轻微下落淡入），配合列表整体上移形成"新日志挤入"的节奏
    &.is-newest {
      animation: log-line-in 0.45s cubic-bezier(0.22, 0.9, 0.32, 1);
    }

    .log-time {
      font-family: $font-mono;
      font-size: 11.5px;
      font-weight: 700;
      color: rgba(120, 175, 215, 0.75);
      letter-spacing: 0.5px;
      flex-shrink: 0;
    }
    .log-step {
      // 占满剩余空间并允许收缩：任务名过长时省略号截断，不与右侧动作徽标重叠
      flex: 1 1 auto;
      min-width: 0;
      overflow: hidden; text-overflow: ellipsis;
      font-weight: 700;
      color: rgba(226, 240, 255, 0.92);
    }
    .log-action {
      flex-shrink: 0;
      font-family: $font-cn;
      font-size: 11px;
      font-weight: 900;
      letter-spacing: 1px;
      padding: 2.5px 7px;
      border-radius: 2px;
      line-height: 1;
    }
  }

  // 动作语义配色：启动绿 / 完成青 / 异常超时红 / 跳过灰 / 强制橙
  .log-ok .log-action { color: #55ffb0; background: rgba(73, 255, 166, 0.1); border: 1px solid rgba(73, 255, 166, 0.28); }
  .log-step .log-action { color: #6fd8ff; background: rgba(0, 212, 255, 0.09); border: 1px solid rgba(0, 212, 255, 0.25); }
  .log-danger .log-action { color: #ff6b8a; background: rgba(255, 77, 106, 0.1); border: 1px solid rgba(255, 77, 106, 0.32); }
  .log-skip .log-action { color: rgba(178, 196, 220, 0.85); background: rgba(150, 175, 205, 0.08); border: 1px solid rgba(150, 175, 205, 0.25); }
  .log-force .log-action { color: #ffb547; background: rgba(255, 182, 72, 0.1); border: 1px solid rgba(255, 182, 72, 0.3); }

  .log-empty {
    position: absolute;
    inset: 0;
    display: flex; align-items: center; justify-content: center;
    font-family: $font-mono;
    font-size: 13px;
    letter-spacing: 0.1em;
    color: rgba(150, 195, 235, 0.5);
  }
}

// 新日志入场：轻微下落 + 淡入，模拟"挤入"列表底部
@keyframes log-line-in {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes countdown-urgent-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.45; }
}

@keyframes indicator-pulse {
  0%, 100% { opacity: 1; box-shadow: 0 0 6px rgba(255, 182, 72, 0.7); }
  50% { opacity: 0.5; box-shadow: 0 0 10px rgba(255, 182, 72, 0.9); }
}

// 待执行指示灯呼吸：灰蓝色、节奏更缓（区别于执行中的活跃橘黄）
@keyframes indicator-pulse-cool {
  0%, 100% { opacity: 1; box-shadow: 0 0 5px rgba(123, 167, 201, 0.55); }
  50% { opacity: 0.45; box-shadow: 0 0 8px rgba(123, 167, 201, 0.75); }
}

@keyframes badge-danger-flash {
  0%, 100% { border-color: rgba(255, 77, 106, 0.3); }
  50% { border-color: rgba(255, 77, 106, 0.8); box-shadow: 0 0 6px rgba(255, 77, 106, 0.4); }
}

// 执行中任务卡片呼吸：外辉光强弱缓慢起伏，营造"进行中"的活跃感
@keyframes alert-running-breathe {
  0%, 100% { box-shadow: inset 0 0 22px rgba(255, 122, 0, 0.14), 0 0 14px rgba(255, 154, 47, 0.22); }
  50% { box-shadow: inset 0 0 26px rgba(255, 122, 0, 0.2), 0 0 24px rgba(255, 154, 47, 0.42); }
}

// ===== Footer =====
.screen-footer {
  display: flex; justify-content: space-between;
  padding: 0;
  height: 16px;
  flex-shrink: 0;
  font-family: $font-display; font-size: 10px;
  color: $text-mute; letter-spacing: 2px;
}

@media (max-height: 820px) {
  .screen-root {
    font-size: clamp(12px, 0.82vw, 15px);
  }

  .screen-content {
    padding: 8px 12px 4px;
    gap: 6px;
  }

  .screen-root::before {
    inset: 6px 8px 6px;
  }

  .screen-header {
    height: 58px;

    .header-title-block {
      height: 42px;

      .drill-title {
        font-size: clamp(28px, 2.4vw, 34px);
        letter-spacing: 3px;
      }
    }
  }

  .kpi-row {
    height: 74px;
    gap: 10px;
  }

  .kpi-card {
    padding-top: 8px;
    padding-bottom: 8px;

    .kpi-orb {
      width: 28px;
      height: 28px;
    }

    .progress-ring-wrap {
      width: 42px;
      height: 42px;
    }
  }

  .screen-main {
    gap: 10px;
  }

  .panel-header {
    height: 38px;
  }

  .panel-body {
    padding: 10px;
  }

  .panel-stages {
    .stages-list {
      gap: 10px;
      padding-right: 16px;
    }

    .stage-card {
      flex-basis: calc((100% - 30px) / 4);
      min-height: 82px;
      padding: 9px 12px;
      gap: 7px;
    }
  }

  .warn-list {
    .alert-card {
      min-height: 84px;
      max-height: 112px;
      padding-top: 8px;
      padding-bottom: 8px;
      gap: 5px;
    }
  }

  .sub-log {
    flex-basis: clamp(150px, 21vh, 210px);

    .log-line {
      height: 22px;
      font-size: 12px;
    }
  }

  .screen-footer {
    height: 8px;
  }
}

@media (max-width: 1080px) {
  .kpi-row {
    height: 82px;
    gap: 10px;
  }

  .kpi-card {
    padding: 10px 12px 8px 58px;
    column-gap: 8px;

    .kpi-orb {
      left: 16px;
      top: 14px;
      width: 28px;
      height: 28px;

      &::before {
        inset: 6px -8px;
      }

      &::after {
        left: 10px;
        bottom: -14px;
        width: 12px;
        height: 12px;
      }
    }

    .kpi-copy {
      left: 58px;
    }

    .kpi-label-zh {
      font-size: 0.82em;
      letter-spacing: 0;
    }

    .kpi-value-row {
      max-width: calc(100% - 82px);
      gap: 4px;
    }

    &.kpi-progress-card {
      .kpi-value-row.kpi-progress-row {
        gap: 6px;
      }
    }

    .kpi-value-num {
      font-size: 1.3em;
    }

    .kpi-value-text {
      font-size: 1.05em;
    }

    .progress-ring-wrap {
      width: 38px;
      height: 38px;
    }

    .progress-ring-text {
      font-size: 0.82em;
      small {
        font-size: 0.65em;
      }
    }

    .node-count-row {
      gap: 2px;

      .kpi-value-num {
        font-size: 18px;
      }

      .node-total {
        font-size: 12px;
      }

      .node-separator {
        font-size: 11px;
      }
    }
  }
}

@media (max-width: 1000px) {
  .screen-content {
    padding: 10px 12px 4px;
    gap: 6px;
  }

  .screen-root::before {
    inset: 6px 8px 6px;
  }

  .screen-header {
    height: 50px;
    padding: 0 12px;

    .header-frame {
      inset: 0 14px;
      width: calc(100% - 28px);
    }

    .header-title-block {
      top: 3px;
      width: min(320px, 56vw);
      min-width: 220px;
      height: 36px;

      .drill-title {
        font-size: 24px;
        letter-spacing: 2px;
        padding-left: 2px;
      }

      .drill-title-en {
        font-size: 8px;
        letter-spacing: 4px;
        max-width: 260px;
        overflow: hidden;
      }
    }

    .header-meta {
      gap: 5px;

      .meta-label,
      .meta-divider {
        display: none;
      }

      .meta-value {
        font-size: 12px;
      }
    }

    .btn-icon {
      right: 18px;
      width: 26px; height: 26px;
    }

    .btn-fullscreen-mark {
      right: 28px;
      bottom: 5px;
      width: 42px;
      height: 32px;
      border: 0;
      background: transparent;
      box-shadow: none;
    }

    .fullscreen-mark {
      width: 39px;
      height: 26px;
    }

    .header-brand {
      display: none;
    }
  }

  .kpi-row {
    gap: 6px;
    height: 76px;
    padding: 0;
  }

  .kpi-card {
    padding: 10px 8px 8px 54px;
    column-gap: 4px;

    .kpi-orb {
      left: 14px;
      top: 50%;
      width: 28px;
      height: 28px;

      &::before {
        inset: 5px -7px;
      }

      &::after {
        left: 9px;
        bottom: -13px;
        width: 10px;
        height: 10px;
      }
    }

    .kpi-copy {
      left: 54px;
    }

    .kpi-label-zh {
      font-size: 0.72em;
      letter-spacing: 0;
      line-height: 1.1;
      padding-bottom: 4px;
      &::after {
        width: 16px;
        height: 1.5px;
      }
    }

    .kpi-label-en {
      font-size: 7px;
      letter-spacing: 2px;
      line-height: 1.2;
    }

    .kpi-value-row {
      max-width: calc(100% - 72px);
    }

    .kpi-value-num {
      font-size: 1.1em;
    }

    .kpi-value-sep {
      font-size: 0.85em;
    }

    .kpi-value-unit {
      font-size: 0.68em;
    }

    .kpi-value-text {
      font-size: 0.9em;
      white-space: nowrap;
    }

    .kpi-node-count {
      margin-left: 4px;
      font-size: 8px;
      white-space: nowrap;
    }

    .progress-ring-wrap {
      width: 40px;
      height: 40px;
    }
    .progress-ring-text {
      font-size: 0.72em;
      small {
        font-size: 0.65em;
      }
    }
    .node-count-row {
      .kpi-value-num {
        font-size: 16px;
      }
      .node-total {
        font-size: 12px;
      }
      .node-separator {
        font-size: 11px;
      }
    }
  }

  .screen-main {
    grid-template-columns: 174px minmax(320px, 1fr) 168px;
    gap: 10px;
    padding: 0;
  }

  .panel-header {
    height: 34px;
    padding: 0 8px 0 18px;

    .panel-title-zh {
      font-size: 0.85em;
      letter-spacing: 1px;
    }

    .panel-title-en {
      font-size: 7px;
      letter-spacing: 2px;
    }
  }

  .panel-body {
    padding: 8px;
  }

  .panel-stages {
    .stages-list {
      gap: 8px;
      padding-right: 6px;
    }

    .stage-card {
      padding: 6px 8px 7px;
      flex-basis: calc((100% - 24px) / 4);
      min-height: 72px;
      gap: 5px;
    }

    .stage-card-top {
      .stage-name-block {
        gap: 5px;
      }

      .stage-index {
        font-size: 12px;
      }

      .stage-name {
        font-size: 0.65em;
        max-width: 86px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }

    .stage-badge {
      font-size: 0.56em;
      padding: 1px 5px;
    }

    .stage-segments {
      height: 6px;
      gap: 2px;
    }

    .stage-card-bottom {
      .stage-meta {
        .meta-key {
          font-size: 0.68em;
        }

        .meta-val {
          font-size: 0.68em;
        }
      }
    }
  }

  .warn-list {
    .alert-card {
      padding: 6px 8px 6px 10px;
    }

    .alert-head {
      .alert-title {
        font-size: 10px;
      }

      .alert-indicator {
        width: 4px; height: 4px;
      }

      .alert-status-badge {
        font-size: 10px;
        padding: 1px 4px;
      }
    }

    .alert-foot {
      .alert-meta {
        .meta-icon { display: none; }
        .meta-label { font-size: 11px; }
        .meta-val { font-size: 11px; }
      }
    }

    .alert-countdown {
      .countdown-text {
        font-size: 9px;
      }
    }
  }

  .screen-footer {
    font-size: 8px;
    letter-spacing: 1px;
    padding: 0;
  }
}

// ===== 无障碍：减少动画 =====
@media (prefers-reduced-motion: reduce) {
  .bg-particle { animation: none !important; opacity: 0.3; }
  .panel-scan-line { animation: none !important; display: none; }
  .header-pulse-line { animation: none !important; display: none; }
  .kpi-card::before, .kpi-card::after { animation: none !important; }
  .status-dot { animation: none !important; }
  .rt-dot { animation: none !important; }
  .bg-scan { animation: none !important; }
  .segment.seg-active { animation: none !important; }
  .fly-ghost, .fly-trail-dot, .hub-shockwave, .hub-burst-particle { display: none !important; }
  .alert-card.alert-warn,
  .alert-card.alert-pending .alert-indicator,
  .log-line.is-newest { animation: none !important; }
}

// ===== 任务完成流式动画 =====
// 全屏覆盖层：承载飞行 ghost、尾迹粒子、圆环吸收特效
.task-fly-layer {
  position: absolute;
  inset: 0;
  z-index: 8000;
  pointer-events: none;
  overflow: hidden;
}

// 飞行 ghost：任务卡片的"信息分身"，沿弧线飞向圆环
.fly-ghost {
  position: absolute;
  top: 0;
  left: 0;
  will-change: transform, opacity;
  pointer-events: none;
  filter: drop-shadow(0 0 14px rgba(45, 228, 255, 0.55)) drop-shadow(0 0 28px rgba(0, 255, 156, 0.22));
}

.fly-ghost-shell {
  position: relative;
  display: flex;
  align-items: center;
  gap: 10px;
  box-sizing: border-box;
  width: 100%;
  padding: 10px 14px 10px 11px;
  border: 1px solid rgba(45, 228, 255, 0.62);
  border-radius: 6px;
  background:
    linear-gradient(135deg, rgba(8, 38, 70, 0.94), rgba(4, 18, 42, 0.94));
  box-shadow:
    0 0 0 1px rgba(0, 255, 156, 0.16),
    0 0 22px rgba(45, 228, 255, 0.4),
    0 0 44px rgba(0, 255, 156, 0.18),
    inset 0 0 18px rgba(45, 228, 255, 0.12);
  overflow: hidden;

  // 顶部高亮扫描线，强化"数据封装"质感
  &::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    height: 1px;
    background: linear-gradient(90deg, transparent, rgba(45, 228, 255, 0.9), transparent);
    box-shadow: 0 0 6px rgba(45, 228, 255, 0.6);
  }

  // 底部网格暗纹
  &::after {
    content: '';
    position: absolute;
    inset: 0;
    background-image:
      linear-gradient(rgba(45, 228, 255, 0.06) 1px, transparent 1px),
      linear-gradient(90deg, rgba(45, 228, 255, 0.06) 1px, transparent 1px);
    background-size: 18px 18px;
    mask-image: linear-gradient(120deg, transparent, #000 60%, transparent);
    -webkit-mask-image: linear-gradient(120deg, transparent, #000 60%, transparent);
    pointer-events: none;
  }
}

// 左侧能量条（完成态：青绿渐变）
.fly-ghost-bar {
  flex: 0 0 4px;
  align-self: stretch;
  min-height: 28px;
  border-radius: 2px;
  background: linear-gradient(180deg, #2cf8d8, #00d4aa 60%, #09b86d);
  box-shadow:
    0 0 8px rgba(44, 248, 216, 0.75),
    0 0 14px rgba(0, 255, 156, 0.4);
}

.fly-ghost-body {
  position: relative;
  z-index: 1;
  flex: 1 1 auto;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.fly-ghost-title {
  font-family: $font-cn;
  font-size: 14px;
  font-weight: 800;
  line-height: 1.25;
  color: #eaf6ff;
  text-shadow: 0 0 6px rgba(45, 228, 255, 0.45);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.fly-ghost-path {
  font-family: $font-mono;
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.5px;
  color: rgba(45, 228, 255, 0.82);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

// 右侧"已完成"标签
.fly-ghost-tag {
  position: relative;
  z-index: 1;
  flex: 0 0 auto;
  padding: 3px 8px;
  border-radius: 3px;
  font-family: $font-mono;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 1px;
  color: #baffdd;
  background: rgba(0, 255, 156, 0.14);
  border: 1px solid rgba(0, 255, 156, 0.42);
  text-shadow: 0 0 4px rgba(0, 255, 156, 0.55);
}

// 飞行尾迹粒子：青色光点，随路径散落并衰减
.fly-trail-dot {
  position: absolute;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(120, 240, 255, 0.98) 0%, rgba(45, 228, 255, 0.5) 55%, transparent 100%);
  box-shadow: 0 0 6px rgba(45, 228, 255, 0.65);
  pointer-events: none;
  will-change: transform, opacity;
  transform: translate(-50%, -50%);
}

// 圆环吸收：冲击波环（双层，外层青绿 + 内层金白）
.hub-shockwave {
  position: absolute;
  transform: translate(-50%, -50%);
  border-radius: 50%;
  border-style: solid;
  pointer-events: none;
  will-change: width, height, opacity, border-width;
  box-sizing: border-box;
}

.hub-shockwave-1 {
  border-color: rgba(45, 228, 255, 0.72);
  box-shadow:
    0 0 20px rgba(45, 228, 255, 0.55),
    inset 0 0 20px rgba(0, 255, 156, 0.32);
}

.hub-shockwave-2 {
  border-color: rgba(255, 224, 162, 0.6);
  box-shadow:
    0 0 14px rgba(255, 180, 74, 0.45),
    inset 0 0 14px rgba(255, 224, 162, 0.22);
}

// 圆环吸收：放射状粒子迸发（金白 → 青绿渐变）
.hub-burst-particle {
  position: absolute;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: radial-gradient(circle, #fff5d6 0%, #ffd36f 35%, #ff8426 70%, transparent 100%);
  box-shadow:
    0 0 8px rgba(255, 180, 74, 0.7),
    0 0 14px rgba(255, 132, 38, 0.35);
  pointer-events: none;
  will-change: transform, opacity;
  transform: translate(-50%, -50%);
}
</style>
