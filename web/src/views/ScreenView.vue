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
        <svg class="header-frame" viewBox="0 0 1200 74" preserveAspectRatio="none" aria-hidden="true">
          <defs>
            <linearGradient id="header-line-grad" x1="0%" y1="0%" x2="100%" y2="0%">
              <stop offset="0%" stop-color="#148cff" stop-opacity="0" />
              <stop offset="12%" stop-color="#11bfff" stop-opacity="0.72" />
              <stop offset="34%" stop-color="#13d8ff" stop-opacity="1" />
              <stop offset="50%" stop-color="#75eaff" stop-opacity="0.92" />
              <stop offset="66%" stop-color="#13d8ff" stop-opacity="1" />
              <stop offset="88%" stop-color="#11bfff" stop-opacity="0.72" />
              <stop offset="100%" stop-color="#148cff" stop-opacity="0" />
            </linearGradient>
            <filter id="header-line-glow" x="-8%" y="-130%" width="116%" height="360%">
              <feGaussianBlur stdDeviation="3.4" result="blur" />
              <feMerge>
                <feMergeNode in="blur" />
                <feMergeNode in="SourceGraphic" />
              </feMerge>
            </filter>
          </defs>
          <path class="header-frame-line header-frame-line-shadow" d="M26 15 H122 L144 31 H326 L362 57 H838 L874 31 H1056 L1078 15 H1174" />
          <path class="header-frame-line" d="M26 15 H122 L144 31 H326 L362 57 H838 L874 31 H1056 L1078 15 H1174" />
          <path class="header-frame-line header-frame-line-inner" d="M312 35 L356 66 H844 L888 35" />
          <path class="header-frame-cap" d="M392 66 H808" />
        </svg>
        <div class="header-title-block">
          <h1 class="drill-title">应急指挥中心</h1>
        </div>
        <div class="header-meta">
        </div>
        <button class="btn-icon" :class="{ active: isFullscreenLike }" @click="toggleFullscreen" title="全屏切换">
          <FullScreen :size="16" />
        </button>
        <div class="header-pulse-line" />
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

        <div class="kpi-card">
          <span class="kpi-orb" />
          <div class="kpi-copy">
            <span class="kpi-label-zh">总耗时</span>
          </div>
          <div class="kpi-value-row kpi-queue-row">
            <span class="kpi-value-num">{{ totalDurationText }}</span>
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
              :progress="progressPercent"
              :center-numerator="completedCount"
              :center-denominator="totalCount"
              :center-hint="`阶段${chineseNum(displayPhaseIndex + 1)} · ${displayPhaseName}`"
              :instance-name="currentDrill.name"
              :size="ringSize"
              :fullscreen="isFullscreenLike"
            />
          </div>
        </section>

        <!-- RIGHT: Active steps -->
        <section class="panel panel-right">
          <div class="sub-panel sub-warn">
            <div class="panel-header">
              <span class="panel-deco-corner tl" />
              <span class="panel-deco-corner tr" />
              <span class="panel-title-zh">执行中步骤</span>
              <span class="panel-realtime">
                <span class="rt-dot" />
                实时
              </span>
              <div class="panel-scan-line" />
            </div>
            <div
              class="panel-body warn-list"
              ref="warnListRef"
              :style="{
                '--visible-alert-count': Math.max(visibleAlerts.length, 1),
                '--alert-card-gap': `${alertCardGap}px`,
              }"
            >
              <div
                v-for="(alert, ai) in visibleAlerts"
                :key="ai"
                class="alert-card"
                :class="'alert-' + alert.level"
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
                <!-- 层级路径：环节名 - 任务名 -->
                <div class="alert-hierarchy">
                  <span class="hierarchy-phase">{{ alert.parentPhase }}</span>
                  <span v-if="alert.directParent !== '--'" class="hierarchy-dash">—</span>
                  <span v-if="alert.directParent !== '--'" class="hierarchy-task">{{ alert.directParent }}</span>
                </div>
              </div>
              <div v-if="activeAlerts.length === 0" class="empty-tip">暂无活跃步骤</div>
              <div v-else-if="visibleAlerts.length < activeAlerts.length" ref="moreTipRef" class="more-tip">
                还有 {{ pendingCount }} 个步骤待执行...
              </div>
            </div>
          </div>
        </section>
      </main>

      <!-- Footer decorations -->
      <footer class="screen-footer" />

      <!-- 任务完成弹窗 -->
      <Transition name="cyber-modal">
        <div v-if="completionModal.visible" class="completion-overlay" @click="dismissCompletionModal">
          <div class="completion-seal" @click.stop>
            <span class="seal-corner tl" />
            <span class="seal-corner tr" />
            <span class="seal-corner bl" />
            <span class="seal-corner br" />
            <div class="seal-scan" />
            <div class="seal-grid" />

            <div class="seal-sigil">
              <span class="sigil-ring sigil-ring-1" />
              <span class="sigil-ring sigil-ring-2" />
              <span class="sigil-ring sigil-ring-3" />
              <svg class="sigil-hex" viewBox="0 0 100 100" aria-hidden="true">
                <polygon points="50,6 90,29 90,71 50,94 10,71 10,29" />
              </svg>
              <svg class="sigil-check" viewBox="0 0 48 48" aria-hidden="true">
                <path d="M13 25 L20 32 L35 15" />
              </svg>
            </div>

            <div class="seal-body">
              <div class="seal-title">任务完成</div>
              <div class="seal-step">{{ completionModal.stepName }}</div>
              <div v-if="completionModal.phaseName" class="seal-path">
                <span class="path-phase">{{ completionModal.phaseName }}</span>
                <template v-if="completionModal.directParent && completionModal.directParent !== '--'">
                  <span class="path-sep">›</span>
                  <span class="path-task">{{ completionModal.directParent }}</span>
                </template>
              </div>
            </div>

            <div class="seal-progress">
              <div :key="completionModal.key" class="seal-progress-fill" />
            </div>
          </div>
        </div>
      </Transition>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onBeforeUnmount, watch } from 'vue'
import type { ComponentPublicInstance } from 'vue'
import { useRoute } from 'vue-router'
import { CircleClose, FullScreen } from '@element-plus/icons-vue'
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
const moreTipRef = ref<HTMLElement | null>(null)

// 任务完成弹窗
const completionModal = ref({
  visible: false,
  key: 0,
  stepName: '',
  phaseName: '',
  directParent: '',
  timer: null as ReturnType<typeof setTimeout> | null,
})

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

// 任务完成弹窗：展示步骤名称、阶段路径，3.5s 后自动关闭
function showCompletionModal(stepId: number) {
  const step = drillSteps.value.find(s => s.id === stepId)
  if (!step) return
  if (completionModal.value.timer) {
    clearTimeout(completionModal.value.timer)
  }
  completionModal.value.key += 1
  completionModal.value.stepName = step.name
  completionModal.value.phaseName = findParentPhase(stepId)
  completionModal.value.directParent = findDirectParent(stepId)
  completionModal.value.visible = true
  completionModal.value.timer = setTimeout(() => {
    completionModal.value.visible = false
    completionModal.value.timer = null
  }, 3500)
}

function dismissCompletionModal() {
  if (completionModal.value.timer) {
    clearTimeout(completionModal.value.timer)
    completionModal.value.timer = null
  }
  completionModal.value.visible = false
}

// === KPI 计算 ===
const completedCount = computed(() =>
  leafSteps.value.filter(s => ['completed', 'skipped', 'timeout', 'issue'].includes(s.status)).length
)
const totalCount = computed(() => leafSteps.value.length)
const pendingCount = computed(() => leafSteps.value.filter(s => s.status === 'pending').length)
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
      if (leaves.length === 0) return { status: child.status, progress: 0 }

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
      return { status, progress: Math.min(progress, 100) }
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

  const running: Array<{
    title: string
    operator: string
    team: string
    parentPhase: string
    directParent: string
    statusLabel: string
    level: 'warn' | 'info' | 'danger'
    seq: number
  }> = []

  const pending: typeof running = []

  // 进行中步骤（只看叶子步骤）
  leafSteps.value
    .filter(s => s.status === 'running')
    .forEach(s => {
      const attrs = typeof (s as any).attributes === 'string'
        ? safeParseJSON((s as any).attributes)
        : (s as any).attributes
      const operatorName = attrs?.operator
      running.push({
        title: s.name,
        operator: operatorName || '',
        team: s.executor_team || '运维部',
        parentPhase: findParentPhase(s.id),
        directParent: findDirectParent(s.id),
        statusLabel: '执行中',
        level: 'warn',
        seq: s.seq,
      })
    })

  // 待执行步骤（始终展示，排在运行中之后）
  leafSteps.value
    .filter(s => s.status === 'pending')
    .forEach((s) => {
      const attrs = typeof (s as any).attributes === 'string'
        ? safeParseJSON((s as any).attributes)
        : (s as any).attributes
      const operatorName = attrs?.operator
      pending.push({
        title: s.name,
        operator: operatorName || '',
        team: s.executor_team || '运维部',
        parentPhase: findParentPhase(s.id),
        directParent: findDirectParent(s.id),
        statusLabel: '待执行',
        level: 'info',
        seq: s.seq,
      })
    })

  // 先按类型分组排序（running/issue/timeout 在前，pending 在后），再按 seq 排序
  const sortedRunning = running.sort((a, b) => a.seq - b.seq)
  const sortedPending = pending.sort((a, b) => a.seq - b.seq)
  return [...sortedRunning, ...sortedPending]
})

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
}

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
function logActionLabel(action: string): string {
  const map: Record<string, string> = {
    complete: '完成', step_complete: '完成',
    issue: '异常', step_issue: '异常',
    timeout: '超时', step_timeout: '超时',
    force_complete: '强制完成',
    skip: '跳过', step_skip: '跳过',
    start: '启动', step_start: '启动',
    pause: '暂停', drill_paused: '暂停',
    resume: '恢复', drill_resumed: '恢复',
    drill_started: '演练启动',
    drill_completed: '演练完成',
    drill_terminated: '演练终止',
  }
  return map[action] || action
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
    const drill = await drillApi.getDetail(drillId.value)
    if (componentDestroyed) return
    currentDrill.value = drill

    const steps = await drillApi.getSteps(drillId.value)
    if (componentDestroyed) return
    drillSteps.value = steps.sort((a, b) => a.seq - b.seq)

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
      const data = JSON.parse(event.data)
      const eventType = normalizeWsEvent(data.type || data.event_type || data.event)
      if (!eventType) return

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

  // 任务完成弹窗（step_complete 为归一化后的事件名）
  if (eventType === 'step_complete') {
    showCompletionModal(stepId)
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

onMounted(() => {
  componentDestroyed = false
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
  if (completionModal.value.timer) {
    clearTimeout(completionModal.value.timer)
    completionModal.value.timer = null
  }
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

  .header-frame-line,
  .header-frame-cap {
    fill: none;
    stroke: url(#header-line-grad);
    stroke-linecap: square;
    stroke-linejoin: miter;
    vector-effect: non-scaling-stroke;
    filter: url(#header-line-glow);
  }

  .header-frame-line {
    stroke-width: 3;
  }

  .header-frame-line-shadow {
    stroke-width: 10;
    opacity: 0.16;
  }

  .header-frame-line-inner {
    stroke-width: 2;
    opacity: 0.72;
  }

  .header-frame-cap {
    stroke-width: 2;
    opacity: 0.8;
  }

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
    border-bottom: 1px solid rgba(0, 214, 255, 0.36);
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
      bottom: -1px;
      width: 44%;
      height: 2px;
      background: linear-gradient(90deg, transparent, rgba(33, 229, 255, 0.42), rgba(117, 234, 255, 0.82), rgba(33, 229, 255, 0.42), transparent);
      box-shadow:
        0 0 10px rgba(0, 213, 255, 0.5),
        0 0 14px rgba(29, 210, 255, 0.38);
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
    z-index: 2;
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
    z-index: 2;
    background: transparent; border: 1px solid $line;
    color: $neon; width: 34px; height: 34px;
    display: flex; align-items: center; justify-content: center;
    cursor: pointer; border-radius: 2px;
    transition: all 0.2s;
    &:hover,
    &.active { border-color: $neon; box-shadow: 0 0 10px $neon-soft; background: rgba(0, 212, 255, 0.1); }
  }
  .header-pulse-line {
    position: absolute;
    bottom: 0; left: 0;
    width: 100%; height: 2px;
    background: linear-gradient(90deg, transparent, $neon, transparent);
    transform: scaleX(0);
    animation: header-pulse 4.8s ease-in-out infinite;
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
  .txt-completed { color: $neon; }
  .txt-terminated { color: $danger; }
  .txt-pending { color: $text-dim; }
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
  min-height: 0;
  .sub-panel {
    position: relative;
    background: transparent; border: 0;
    display: flex; flex-direction: column;
    min-height: 0;
    flex: 1;
  }
}

// ===== Alerts =====
.warn-list {
  --visible-alert-count: 5;
  --alert-card-gap: 9px;
  display: flex; flex-direction: column; gap: clamp(8px, 0.95vh, 11px);
  overflow: hidden;
  flex: 1;
  min-height: 0;
  padding-bottom: clamp(8px, 0.95vh, 12px);
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

    // 运行中卡片微光
    &.alert-warn {
      border-color: rgba(255, 182, 72, 0.34);
      box-shadow: inset 0 0 22px rgba(255, 140, 40, 0.08);
    }
    &.alert-danger {
      border-color: rgba(255, 77, 106, 0.42);
      box-shadow: inset 0 0 22px rgba(255, 77, 106, 0.1);
    }
    &.alert-info {
      border-color: rgba(0, 212, 255, 0.28);
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
  .alert-hierarchy {
    display: flex; align-items: center; gap: 0;
    min-height: 20px;
    padding-top: clamp(5px, 0.6vh, 7px);
    padding-bottom: 1px;
    border-top: 1px solid rgba(0, 212, 255, 0.1);
    font-size: clamp(12px, 0.78vw, 14px);
    line-height: 1.2;
    min-width: 0;
    overflow: hidden;
    .hierarchy-phase {
      font-weight: 600;
      color: rgba(0, 212, 255, 0.85);
      line-height: 1.2;
      white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
      max-width: 45%;
    }
    .hierarchy-dash {
      color: rgba(200, 220, 245, 0.6); margin: 0 6px;
      font-weight: 400;
      line-height: 1.2;
    }
    .hierarchy-task {
      font-weight: 600;
      color: rgba(226, 240, 255, 0.86);
      line-height: 1.2;
      white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
      max-width: 45%;
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

@keyframes indicator-pulse {
  0%, 100% { opacity: 1; box-shadow: 0 0 6px rgba(255, 182, 72, 0.7); }
  50% { opacity: 0.5; box-shadow: 0 0 10px rgba(255, 182, 72, 0.9); }
}

@keyframes badge-danger-flash {
  0%, 100% { border-color: rgba(255, 77, 106, 0.3); }
  50% { border-color: rgba(255, 77, 106, 0.8); box-shadow: 0 0 6px rgba(255, 77, 106, 0.4); }
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
    gap: clamp(6px, 0.8vh, 9px);

    .alert-card {
      min-height: 84px;
      max-height: 112px;
      padding-top: 8px;
      padding-bottom: 8px;
      gap: 5px;
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

    .alert-hierarchy {
      font-size: 9px;
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
  .sigil-ring { animation: none !important; }
  .seal-scan { animation: none !important; display: none; }
  .sigil-check path { animation: none !important; stroke-dashoffset: 0 !important; }
  .sigil-hex polygon { animation: none !important; }
}

// ===== 任务完成弹窗 =====
.completion-overlay {
  position: fixed;
  inset: 0;
  z-index: 9000;
  display: flex;
  align-items: center;
  justify-content: center;
  background:
    radial-gradient(circle at 50% 50%, rgba(0, 80, 60, 0.12), transparent 52%),
    rgba(2, 6, 16, 0.72);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  cursor: pointer;
}

.completion-seal {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 18px;
  min-width: 340px;
  max-width: 480px;
  padding: 36px 52px 30px;
  border: 1px solid rgba(0, 255, 156, 0.42);
  border-radius: 4px;
  background:
    linear-gradient(135deg, rgba(0, 40, 28, 0.72), rgba(4, 14, 34, 0.92)),
    radial-gradient(circle at 50% 0%, rgba(0, 255, 156, 0.1), transparent 60%);
  box-shadow:
    0 0 0 1px rgba(0, 212, 255, 0.12),
    0 0 36px rgba(0, 255, 156, 0.18),
    0 0 72px rgba(0, 212, 255, 0.1),
    inset 0 0 32px rgba(0, 255, 156, 0.06);
  cursor: default;
  overflow: hidden;
}

// 四角 HUD 框
.seal-corner {
  position: absolute;
  width: 18px;
  height: 18px;
  border: 2px solid $neon;
  pointer-events: none;
  opacity: 0;
  animation: corner-snap 0.4s 0.15s ease-out forwards;

  &.tl { top: -1px; left: -1px; border-right: 0; border-bottom: 0; }
  &.tr { top: -1px; right: -1px; border-left: 0; border-bottom: 0; }
  &.bl { bottom: -1px; left: -1px; border-right: 0; border-top: 0; }
  &.br { bottom: -1px; right: -1px; border-left: 0; border-top: 0; }
}

@keyframes corner-snap {
  from { opacity: 0; transform: scale(1.6); }
  to { opacity: 1; transform: scale(1); }
}

// 扫描线
.seal-scan {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, transparent, rgba(0, 255, 156, 0.7), transparent);
  box-shadow: 0 0 12px rgba(0, 255, 156, 0.5);
  animation: seal-scan-sweep 2.2s 0.3s ease-in-out infinite;
  pointer-events: none;
  z-index: 2;
}

@keyframes seal-scan-sweep {
  0% { top: 0; opacity: 0; }
  8% { opacity: 1; }
  50% { top: calc(100% - 3px); opacity: 0.4; }
  58% { opacity: 0; }
  100% { top: 0; opacity: 0; }
}

// 网格纹理
.seal-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(0, 255, 156, 0.035) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 255, 156, 0.035) 1px, transparent 1px);
  background-size: 28px 28px;
  mask-image: radial-gradient(circle at center, #000 30%, transparent 80%);
  -webkit-mask-image: radial-gradient(circle at center, #000 30%, transparent 80%);
  pointer-events: none;
}

// 成功徽记：六边形盾牌 + 雷达环 + 自绘对勾
.seal-sigil {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 88px;
  height: 88px;
}

.sigil-ring {
  position: absolute;
  top: 50%;
  left: 50%;
  width: 64px;
  height: 64px;
  border: 1.5px solid rgba(0, 255, 156, 0.55);
  border-radius: 50%;
  transform: translate(-50%, -50%) scale(0.5);
  opacity: 0;
  animation: sigil-ring-expand 2s ease-out infinite;
}

.sigil-ring-1 { animation-delay: 0s; }
.sigil-ring-2 { animation-delay: 0.66s; }
.sigil-ring-3 { animation-delay: 1.32s; }

@keyframes sigil-ring-expand {
  0% { transform: translate(-50%, -50%) scale(0.5); opacity: 0.9; border-color: rgba(0, 255, 156, 0.7); }
  100% { transform: translate(-50%, -50%) scale(1.7); opacity: 0; border-color: rgba(0, 212, 255, 0.1); }
}

.sigil-hex {
  position: relative;
  width: 64px;
  height: 64px;
  filter: drop-shadow(0 0 14px rgba(0, 255, 156, 0.4));
  z-index: 2;

  polygon {
    fill: rgba(0, 255, 156, 0.1);
    stroke: rgba(0, 255, 156, 0.65);
    stroke-width: 2;
    stroke-linejoin: round;
    animation: sigil-hex-glow 2s ease-in-out infinite;
  }
}

@keyframes sigil-hex-glow {
  0%, 100% { stroke: rgba(0, 255, 156, 0.55); fill: rgba(0, 255, 156, 0.08); }
  50% { stroke: rgba(0, 255, 156, 0.85); fill: rgba(0, 255, 156, 0.16); }
}

.sigil-check {
  position: absolute;
  width: 38px;
  height: 38px;
  z-index: 3;

  path {
    fill: none;
    stroke: $ok;
    stroke-width: 5;
    stroke-linecap: round;
    stroke-linejoin: round;
    stroke-dasharray: 42;
    stroke-dashoffset: 42;
    animation: sigil-check-draw 0.55s 0.25s ease-out forwards;
    filter: drop-shadow(0 0 6px rgba(0, 255, 156, 0.8));
  }
}

@keyframes sigil-check-draw {
  to { stroke-dashoffset: 0; }
}

// 文本区
.seal-body {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
}

.seal-title {
  font-family: $font-cn;
  font-size: clamp(20px, 1.6em, 26px);
  font-weight: 900;
  letter-spacing: 6px;
  color: #ffffff;
  text-shadow:
    0 0 12px rgba(0, 255, 156, 0.7),
    0 0 24px rgba(0, 212, 255, 0.3);
  padding-left: 6px;
}

.seal-step {
  font-family: $font-cn;
  font-size: clamp(16px, 1.25em, 21px);
  font-weight: 800;
  color: #2cf8d8;
  text-shadow: 0 0 10px rgba(44, 248, 216, 0.5);
  text-align: center;
  max-width: 380px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.seal-path {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 2px;
  font-family: $font-mono;
  font-size: clamp(11px, 0.9em, 14px);
  color: rgba(160, 205, 245, 0.7);

  .path-phase {
    color: rgba(0, 212, 255, 0.85);
    font-weight: 700;
  }
  .path-sep {
    color: rgba(120, 160, 200, 0.5);
  }
  .path-task {
    color: rgba(200, 220, 245, 0.7);
    max-width: 160px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

// 自动关闭进度条
.seal-progress {
  position: relative;
  z-index: 1;
  width: 100%;
  height: 3px;
  margin-top: 4px;
  border-radius: 2px;
  background: rgba(0, 255, 156, 0.1);
  overflow: hidden;
}

.seal-progress-fill {
  height: 100%;
  border-radius: 2px;
  background: linear-gradient(90deg, $ok, $neon);
  box-shadow: 0 0 8px rgba(0, 255, 156, 0.6);
  animation: seal-progress-drain 3.5s linear forwards;
}

@keyframes seal-progress-drain {
  from { width: 100%; }
  to { width: 0%; }
}

// Vue Transition: overlay 淡入淡出 + seal 弹性缩放
.cyber-modal-enter-active {
  transition: opacity 0.35s ease-out;

  .completion-seal {
    transition: transform 0.5s cubic-bezier(0.34, 1.56, 0.64, 1), opacity 0.35s ease-out;
    transition-delay: 0.05s;
  }
}

.cyber-modal-leave-active {
  transition: opacity 0.25s ease-in;

  .completion-seal {
    transition: transform 0.25s ease-in, opacity 0.25s ease-in;
  }
}

.cyber-modal-enter-from {
  opacity: 0;

  .completion-seal {
    opacity: 0;
    transform: scale(0.82) translateY(12px);
  }
}

.cyber-modal-leave-to {
  opacity: 0;

  .completion-seal {
    opacity: 0;
    transform: scale(0.94);
  }
}

@media (max-width: 540px) {
  .completion-seal {
    min-width: unset;
    width: calc(100vw - 48px);
    max-width: unset;
    padding: 28px 28px 24px;
    gap: 14px;
  }

  .seal-sigil {
    width: 72px;
    height: 72px;
  }

  .sigil-hex {
    width: 54px;
    height: 54px;
  }

  .sigil-check {
    width: 32px;
    height: 32px;
  }

  .seal-step {
    max-width: 100%;
  }
}
</style>
