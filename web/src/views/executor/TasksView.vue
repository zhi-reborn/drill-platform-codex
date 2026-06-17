<template>
  <div class="executor-tasks">
    <!-- 页面 1: 活跃演练列表 -->
    <div v-if="!selectedDrillId" class="drill-list-page">
      <div class="page-header">
        <h2 class="page-title">我的任务</h2>
      </div>

      <div class="page-content">
        <!-- 统计卡片 -->
        <el-row :gutter="20" class="stats-row">
          <el-col :span="6">
            <el-card class="stat-card">
              <div class="stat-label">我的任务</div>
              <div class="stat-value">{{ myTasksCount }}</div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card class="stat-card">
              <div class="stat-label">待执行</div>
              <div class="stat-value pending">{{ pendingTasksCount }}</div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card class="stat-card">
              <div class="stat-label">执行中</div>
              <div class="stat-value in-progress">{{ inProgressTasksCount }}</div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card class="stat-card">
              <div class="stat-label">已完成</div>
              <div class="stat-value completed">{{ completedTasksCount }}</div>
            </el-card>
          </el-col>
        </el-row>

        <!-- 活跃演练列表 -->
        <el-card class="section-card">
          <template #header>
            <span class="card-title">活跃演练 - 选择演练查看任务</span>
          </template>
          <el-row :gutter="20">
            <el-col
              v-for="drill in activeDrillsWithTasks"
              :key="drill.id"
              :xs="24"
              :sm="12"
              :lg="8"
            >
              <el-card class="drill-card" @click="goToDrillTasks(drill.id)">
                <div class="drill-header">
                  <span class="drill-name">{{ drill.name }}</span>
                  <DrillStatusBadge :status="drill.drillStatus" type="drill" />
                </div>
                <div class="drill-progress">
                  <el-progress
                    :percentage="Math.round(drill.completedSteps / drill.totalSteps * 100)"
                    :stroke-width="8"
                  />
                </div>
                <div class="drill-tasks-summary">
                  <div class="task-stat">
                    <span class="label">我的任务</span>
                    <span class="value">{{ drill.myTasksCount }}</span>
                  </div>
                  <div class="task-stat">
                    <span class="label">待执行</span>
                    <span class="value pending">{{ drill.pendingTasksCount }}</span>
                  </div>
                </div>
                <div class="drill-actions">
                  <el-button type="success" size="small" @click.stop="viewScreen(drill.id)">
                    <el-icon><Monitor /></el-icon>
                    大屏
                  </el-button>
                  <el-button type="warning" size="small" @click.stop="viewScreen2(drill.id)">
                    <el-icon><DataBoard /></el-icon>
                    大屏2
                  </el-button>
                  <el-button type="primary" size="small" @click.stop.prevent="goToDrillTasks(drill.id)">
                    <el-icon><ArrowRight /></el-icon>
                    查看任务
                  </el-button>
                </div>
              </el-card>
            </el-col>
          </el-row>
          <div v-if="activeDrillsWithTasks.length === 0" class="empty-tip">
            暂无活跃演练
          </div>
        </el-card>

        <!-- 最近活动 -->
        <el-card class="section-card">
          <template #header>
            <span class="card-title">最近活动</span>
          </template>
          <el-table :data="recentActivity" stripe style="width: 100%">
            <el-table-column prop="type" label="类型" width="120">
              <template #default="{ row }">
                <el-tag :type="getActivityTypeTag(row.type)" size="small">
                  {{ getActivityLabel(row.type) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="drill_name" label="演练名称" min-width="180" />
            <el-table-column prop="operator" label="操作人" width="100" />
            <el-table-column prop="created_at" label="时间" width="160">
              <template #default="{ row }">
                {{ formatTime(row.created_at) }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </div>
    </div>

    <!-- 页面 2: 选中演练的任务列表 -->
    <div v-else class="tasks-detail-page">
      <div class="page-header">
        <div class="breadcrumb" @click="backToDrillList">
          <el-icon><ArrowLeft /></el-icon>
          <span>返回演练列表</span>
        </div>
        <h2 class="page-title">{{ currentDrill?.name || '任务列表' }}</h2>
        <div class="header-actions">
          <el-button type="success" @click="viewScreen(selectedDrillId)">
            <el-icon><Monitor /></el-icon>
            大屏
          </el-button>
          <el-button type="warning" @click="viewScreen2(selectedDrillId)">
            <el-icon><DataBoard /></el-icon>
            大屏2
          </el-button>
        </div>
        <div class="filter-group">
          <el-radio-group v-model="filterStatus" size="default" @change="handleFilterChange">
            <el-radio-button value="">全部</el-radio-button>
            <el-radio-button value="pending">待执行</el-radio-button>
            <el-radio-button value="running">执行中</el-radio-button>
            <el-radio-button value="completed">已完成</el-radio-button>
            <el-radio-button value="issue">异常</el-radio-button>
          </el-radio-group>
        </div>
      </div>

      <div v-loading="loading" class="tasks-container">
        <EmptyBox v-if="!loading && filteredTasks.length === 0" title="暂无任务" description="当前没有分配给您的任务" />

        <div v-else class="flow-list">
          <template v-for="(group, gIdx) in groupedTasks" :key="group.phase">
            <!-- 阶段分隔标题 -->
            <div class="phase-header" :class="{ 'is-first': gIdx === 0 }">
              <div class="phase-icon">
                <svg v-if="group.activeCount > 0" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" width="16" height="16"><path fill="currentColor" d="M512 64a448 448 0 1 1 0 896 448 448 0 0 1 0-896zM288 512a38.4 38.4 0 0 0 0 76.8h376.32l-100.352 107.52a38.4 38.4 0 0 0 55.04 53.76l168.96-180.48a38.4 38.4 0 0 0 0-53.76l-168.96-180.48a38.4 38.4 0 0 0-55.04 53.76L664.32 512H288z"/></svg>
                <svg v-else-if="group.doneCount === group.totalCount" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" width="16" height="16"><path fill="currentColor" d="M512 896a384 384 0 1 0 0-768 384 384 0 0 0 0 768m0 64a448 448 0 1 1 0-896 448 448 0 0 1 0 896"/><path fill="currentColor" d="M745.344 361.344a32 32 0 0 1 45.312 45.312l-288 288a32 32 0 0 1-45.312 0l-160-160a32 32 0 1 1 45.312-45.312L480 626.752z"/></svg>
                <span v-else class="phase-dot"></span>
              </div>
              <span class="phase-name">{{ group.phase }}</span>
              <span class="phase-stats">{{ group.doneCount }}/{{ group.totalCount }}</span>
            </div>
            <!-- 阶段内的任务列表 -->
            <template
              v-for="(task, tIdx) in group.tasks"
              :key="task.id"
            >
              <!-- 容器步骤：渲染为分节标题（depth-0 阶段容器由 phase-header 展示，跳过） -->
              <div
                v-if="(task as any)._isContainer && getStepDepth(task) > 0"
                class="section-header"
                :class="[`depth-${getStepDepth(task)}`]"
                :style="{ marginLeft: getFlowIndent(task) }"
              >
                <div class="section-badge">
                  <span class="section-badge-label">{{ getStepDepthLabel(task) }}</span>
                </div>
                <span class="section-name">{{ task.name }}</span>
                <span class="section-child-count">{{ getSectionChildCount(task) }}</span>
                <DrillStatusBadge v-if="task.status !== 'pending'" :status="task.status" type="step" />
              </div>
              <!-- 叶子步骤：渲染为 flow card -->
              <div
                v-else
                class="flow-item"
                :class="[getStatusClass(task.status), { 'is-last': tIdx === group.tasks.length - 1 && gIdx === groupedTasks.length - 1 }]"
                :style="{ marginLeft: getFlowIndent(task) }"
              >
              <div class="flow-rail">
                <div class="flow-dot" :class="task.status">
                  <svg v-if="task.status === 'completed'" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" width="12" height="12"><path fill="currentColor" d="M745.344 361.344a32 32 0 0 1 45.312 45.312l-288 288a32 32 0 0 1-45.312 0l-160-160a32 32 0 1 1 45.312-45.312L480 626.752z"/></svg>
                  <svg v-else-if="task.status === 'running'" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" width="12" height="12"><path fill="currentColor" d="M288 512a38.4 38.4 0 0 0 0 76.8h448a38.4 38.4 0 0 0 0-76.8H288z"/></svg>
                  <svg v-else-if="task.status === 'timeout'" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" width="12" height="12"><path fill="currentColor" d="M512 64a448 448 0 1 1 0 896 448 448 0 0 1 0-896m0 393.664L407.936 353.6a38.4 38.4 0 1 0-54.336 54.336L457.664 512 353.6 616.064a38.4 38.4 0 1 0 54.336 54.336L512 566.336 616.064 670.4a38.4 38.4 0 1 0 54.336-54.336L566.336 512 670.4 407.936a38.4 38.4 0 1 0-54.336-54.336z"/></svg>
                  <svg v-else-if="task.status === 'issue'" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" width="12" height="12"><path fill="currentColor" d="M512 64a448 448 0 1 1 0 896 448 448 0 0 1 0-896zm-32 232v288a32 32 0 0 0 64 0V296a32 32 0 0 0-64 0zm0 416a32 32 0 1 0 64 0 32 32 0 0 0-64 0z"/></svg>
                  <span v-else class="dot-inner"></span>
                </div>
                <div class="flow-line"></div>
              </div>
              <div class="flow-content">
                <div class="flow-card" :class="[getStatusClass(task.status)]" @click="goToTaskDetail(task.id)">
                  <div class="flow-card-header">
                    <div class="flow-step-name">{{ task.name }}</div>
                    <DrillStatusBadge :status="task.status" type="step" />
                  </div>
                  <div class="flow-card-body">
                    <div class="flow-meta-row">
                      <el-tag v-if="task.default_assignee_role === 'director'" size="small" type="warning">指挥组</el-tag>
                      <el-tag v-else-if="task.default_assignee_role" size="small" type="primary">{{ task.default_assignee_role }}</el-tag>
                      <el-tag v-if="task.executor_team" size="small" type="info">{{ task.executor_team }}</el-tag>
                      <el-tag v-if="getTaskOperator(task)" size="small" type="success">
                        操作人：{{ getTaskOperator(task) }}
                      </el-tag>
                    </div>
                    <div class="flow-meta">
                      <div v-if="task.estimated_duration_minutes" class="flow-duration">
                        预计耗时：{{ task.estimated_duration_minutes }} 分钟
                      </div>
                      <div v-if="task.timeout_at" class="flow-deadline">
                        <el-icon><Clock /></el-icon>
                        <span>截止：{{ formatDeadline(task.timeout_at) }}</span>
                      </div>
                    </div>
                  </div>
                  <div class="flow-card-footer" @click.stop>
                    <el-button
                      v-if="task.status === 'pending' && !canStartTask(task)"
                      type="info"
                      size="small"
                      disabled
                    >
                      等待前序完成
                    </el-button>
                    <el-button
                      v-else-if="task.status === 'pending' && canStartTask(task)"
                      type="primary"
                      size="small"
                      @click="goToTaskDetail(task.id)"
                    >
                      开始执行
                    </el-button>
                    <el-button
                      v-else-if="task.status === 'running' && !isParentTask(task)"
                      type="success"
                      size="small"
                      @click="goToTaskDetail(task.id)"
                    >
                      <el-icon><CircleCheck /></el-icon>
                      完成
                    </el-button>
                    <el-button
                      v-else-if="task.status === 'running'"
                      size="small"
                      disabled
                    >
                      子任务执行中
                    </el-button>
                    <el-button
                      v-else-if="task.status === 'issue'"
                      type="warning"
                      size="small"
                      @click="goToTaskDetail(task.id)"
                    >
                      查看异常
                    </el-button>
                    <el-button
                      v-else-if="task.status === 'timeout'"
                      type="danger"
                      size="small"
                      @click="goToTaskDetail(task.id)"
                    >
                      已超时
                    </el-button>
                    <el-button
                      v-else-if="task.status === 'skipped'"
                      type="info"
                      size="small"
                      disabled
                    >
                      已跳过
                    </el-button>
                    <el-button v-else size="small" disabled>
                      已完成
                    </el-button>
                  </div>
                </div>
              </div>
            </div>
          </template>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Clock, Monitor, ArrowLeft, ArrowRight, CircleCheck, DataBoard } from '@element-plus/icons-vue'
import type { StepAttributes } from '@/types/template'
import type { StepInstance, StepStatus } from '@/types/instance'
import type { DrillInstance } from '@/types'
import DrillStatusBadge from '@/components/common/DrillStatusBadge.vue'
import EmptyBox from '@/components/common/EmptyBox.vue'
import { taskApi } from '@/api/modules/task'
import { drillApi } from '@/api/modules/drill'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()

const loading = ref(false)
const tasks = ref<StepInstance[]>([])
const drillFlowSteps = ref<StepInstance[]>([])
const instances = ref<DrillInstance[]>([])
const filterStatus = ref<string>('')

// 选中的演练 ID
const selectedDrillId = ref<number | null>(null)
const currentDrill = ref<DrillInstance | null>(null)

// 统计
const myTasksCount = computed(() => tasks.value.length)
const pendingTasksCount = computed(() => tasks.value.filter((t: StepInstance) => t.status === 'pending').length)
const inProgressTasksCount = computed(() => tasks.value.filter((t: StepInstance) => t.status === 'running').length)
const completedTasksCount = computed(() => tasks.value.filter((t: StepInstance) => t.status === 'completed').length)

const workflowSteps = computed(() => drillFlowSteps.value.length > 0 ? drillFlowSteps.value : tasks.value)

const workflowStepById = computed(() => {
  const map = new Map<number, StepInstance>()
  for (const step of workflowSteps.value) {
    map.set(step.id, step)
  }
  return map
})

const workflowChildrenByParentId = computed(() => {
  const map = new Map<number, StepInstance[]>()
  for (const step of workflowSteps.value) {
    if (!step.parent_step_id) continue
    const children = map.get(step.parent_step_id) || []
    children.push(step)
    map.set(step.parent_step_id, children)
  }
  return map
})

const parentTaskIdSet = computed(() => new Set(workflowChildrenByParentId.value.keys()))

const workflowDepthMap = computed(() => {
  const map = new Map<number, number>()
  for (const step of workflowSteps.value) {
    let depth = 0
    let pid = step.parent_step_id
    const visited = new Set<number>()
    while (pid && !visited.has(pid)) {
      visited.add(pid)
      depth++
      pid = workflowStepById.value.get(pid)?.parent_step_id
    }
    map.set(step.id, depth)
  }
  return map
})

const sectionChildCountMap = computed(() => {
  const map = new Map<number, number>()
  for (const [parentId, children] of workflowChildrenByParentId.value) {
    const leafCount = children.filter(child => !parentTaskIdSet.value.has(child.id)).length
    map.set(parentId, leafCount)
  }
  return map
})

const taskOperatorMap = computed(() => {
  const map = new Map<number, string>()
  for (const task of tasks.value) {
    const attrs = task.attributes as StepAttributes | string | null | undefined
    if (!attrs) continue
    if (typeof attrs === 'string') {
      try {
        const parsed = JSON.parse(attrs)
        if (parsed?.operator) map.set(task.id, parsed.operator)
      } catch { /* ignore */ }
    } else if (attrs.operator) {
      map.set(task.id, attrs.operator)
    }
  }
  return map
})

// 活跃演练（带任务统计）
const activeDrillsWithTasks = computed(() => {
  return instances.value
    .filter(i => i.status === 'running' || i.status === 'paused')
    .map(drill => {
      const drillTasks = tasks.value.filter((t: StepInstance) => t.drill_instance_id === drill.id)
      return {
        id: drill.id,
        name: drill.name,
        drillStatus: drill.status,
        completedSteps: drill.progress_pct,
        totalSteps: 100,
        myTasksCount: drillTasks.length,
        pendingTasksCount: drillTasks.filter((t: StepInstance) => t.status === 'pending' || t.status === 'running').length,
      }
    })
})

const filteredTasks = computed(() => {
  let result = tasks.value
  if (selectedDrillId.value) {
    result = result.filter((t: StepInstance) => t.drill_instance_id === selectedDrillId.value)
  }
  if (filterStatus.value) {
    result = result.filter((t: StepInstance) => t.status === filterStatus.value)
  }
  return sortTasksByFlowOrder(result, drillFlowSteps.value)
})

const groupedTasks = computed(() => {
  const sorted = filteredTasks.value
  const idSet = new Set<number>()
  for (const t of sorted) idSet.add(t.id)
  const parentSet = new Set<number>()
  for (const t of sorted) {
    if (t.parent_step_id && idSet.has(t.parent_step_id)) {
      parentSet.add(t.parent_step_id)
    }
  }

  const groups: { phase: string; tasks: (StepInstance & { _isContainer?: boolean })[]; doneCount: number; activeCount: number; totalCount: number }[] = []
  let currentGroup: typeof groups[0] | null = null

  for (const task of sorted) {
    const phase = task.phase || task.phase_step || '未分类'
    if (!currentGroup || currentGroup.phase !== phase) {
      currentGroup = { phase, tasks: [], doneCount: 0, activeCount: 0, totalCount: 0 }
      groups.push(currentGroup)
    }
    const isContainer = parentSet.has(task.id)
    currentGroup.tasks.push({ ...task, _isContainer: isContainer })
    if (!isContainer) {
      currentGroup.totalCount++
      if (['completed', 'skipped'].includes(task.status)) currentGroup.doneCount++
      if (task.status === 'running') currentGroup.activeCount++
    }
  }
  return groups
})

function flattenTreeOrder(taskList: StepInstance[]): StepInstance[] {
  if (taskList.length === 0) return []

  const idMap = new Map<number, StepInstance>()
  for (const t of taskList) {
    idMap.set(t.id, t)
  }

  // 构建子节点映射
  const childrenMap = new Map<number, StepInstance[]>()
  const roots: StepInstance[] = []

  for (const t of taskList) {
    if (t.parent_step_id && idMap.has(t.parent_step_id)) {
      if (!childrenMap.has(t.parent_step_id)) childrenMap.set(t.parent_step_id, [])
      childrenMap.get(t.parent_step_id)!.push(t)
    } else {
      roots.push(t)
    }
  }

  // 子节点按 seq 排序
  for (const children of childrenMap.values()) {
    children.sort((a, b) => (a.seq || 0) - (b.seq || 0))
  }
  roots.sort((a, b) => (a.seq || 0) - (b.seq || 0))

  // 前序遍历
  const result: StepInstance[] = []
  function traverse(node: StepInstance) {
    result.push(node)
    const children = childrenMap.get(node.id)
    if (children) {
      for (const child of children) {
        traverse(child)
      }
    }
  }

  for (const root of roots) {
    traverse(root)
  }
  return result
}

function sortTasksByFlowOrder(visibleTasks: StepInstance[], flowSteps: StepInstance[]): StepInstance[] {
  if (visibleTasks.length === 0) return []
  if (flowSteps.length === 0) return flattenTreeOrder(visibleTasks)

  const visibleMap = new Map<number, StepInstance>()
  for (const task of visibleTasks) visibleMap.set(task.id, task)

  const ordered: StepInstance[] = []
  const orderedIds = new Set<number>()
  for (const step of flattenTreeOrder(flowSteps)) {
    const visibleTask = visibleMap.get(step.id)
    if (visibleTask) {
      ordered.push(visibleTask)
      orderedIds.add(visibleTask.id)
    }
  }

  for (const task of flattenTreeOrder(visibleTasks)) {
    if (!orderedIds.has(task.id)) {
      ordered.push(task)
      orderedIds.add(task.id)
    }
  }
  return ordered
}

const recentActivity = ref<any[]>([])

const getStatusClass = (status: string) => {
  const classMap: Record<string, string> = {
    running: 'status-in-progress',
    issue: 'status-issued',
    completed: 'status-completed',
    timeout: 'status-issued',
    skipped: 'status-completed',
  }
  return classMap[status] || ''
}

function getStepDepth(task: StepInstance): number {
  return workflowDepthMap.value.get(task.id) || 0
}

function getFlowIndent(task: StepInstance): string {
  return `${Math.max(getStepDepth(task) - 1, 0) * 20}px`
}

function getStepDepthLabel(task: StepInstance): string {
  const depth = getStepDepth(task)
  const labels: Record<number, string> = { 0: '阶段', 1: '环节', 2: '任务' }
  return labels[depth] ?? '分组'
}

function getSectionChildCount(task: StepInstance): string {
  const leafCount = sectionChildCountMap.value.get(task.id) || 0
  return leafCount > 0 ? `${leafCount} 项` : ''
}

function getTaskOperator(task: StepInstance): string {
  return taskOperatorMap.value.get(task.id) || ''
}

function getActivityTypeTag(type: string): 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  const map: Record<string, any> = {
    start: 'primary',
    pause: 'warning',
    resume: 'primary',
    terminate: 'danger',
    drill_start: 'primary',
    drill_complete: 'success',
    drill_terminate: 'danger',
    drill_pause: 'warning',
    drill_resume: 'primary',
    step_start: 'info',
    step_complete: 'success',
  }
  return map[type] || 'info'
}

function getActivityLabel(type: string): string {
  const map: Record<string, string> = {
    start: '演练启动',
    pause: '演练暂停',
    resume: '演练恢复',
    terminate: '演练终止',
    drill_start: '演练开始',
    drill_complete: '演练完成',
    drill_terminate: '演练终止',
    drill_pause: '演练暂停',
    drill_resume: '演练恢复',
    step_start: '步骤开始',
    step_complete: '步骤完成',
  }
  return map[type] || type
}

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function formatDeadline(d: string): string {
  if (!d) return '无'
  const date = new Date(d)
  const now = new Date()
  const diff = date.getTime() - now.getTime()
  const mins = Math.floor(diff / (1000 * 60))

  if (mins < 0) {
    return '已过期'
  }
  if (mins < 60) {
    return `${mins}分钟`
  }
  const hours = Math.floor(mins / 60)
  if (hours < 1) {
    return '1 小时内'
  }
  return `${hours}小时`
}

// 检查是否为父任务（有其他任务的 parent_step_id 指向它）
function isParentTask(task: StepInstance): boolean {
  return parentTaskIdSet.value.has(task.id)
}

// 终态集合
const DEPENDENCY_SATISFIED_STATUSES: StepStatus[] = ['completed', 'timeout', 'issue']
const dependencySatisfiedStatusSet = new Set<StepStatus>(DEPENDENCY_SATISFIED_STATUSES)

// 解析 pre_step_ids（API 返回的是 JSON 字符串）
function parsePreStepIds(preStepIds: number[] | string | null | undefined): number[] {
  if (!preStepIds) return []
  if (Array.isArray(preStepIds)) return preStepIds
  if (typeof preStepIds === 'string') {
    try {
      const parsed = JSON.parse(preStepIds)
      return Array.isArray(parsed) ? parsed : []
    } catch {
      return []
    }
  }
  return []
}

function parseStepAttributes(attributes: StepAttributes | string | null | undefined): StepAttributes {
  if (!attributes) return {}
  if (typeof attributes !== 'string') return attributes
  try {
    return JSON.parse(attributes)
  } catch {
    return {}
  }
}

// 判断任务是否满足开始条件（前序步骤全部完成，父步骤已开始）
function canStartTask(task: StepInstance): boolean {
  if (task.status !== 'pending') return false
  if (isParentTask(task)) return false

  // 检查父步骤链是否已激活：所有祖先必须处于 running 状态
  let currentAncestorId = task.parent_step_id
  while (currentAncestorId) {
    const parent = workflowStepById.value.get(currentAncestorId)
    if (!parent) break
    if (parent.status !== 'running' && parent.status !== 'completed') return false
    currentAncestorId = parent.parent_step_id
  }

  // 检查前序步骤
  const preStepIds = task.pre_step_ids || []
  if (preStepIds.length > 0) {
    const allPreDone = preStepIds.every((preId: number) => {
      const preStep = workflowStepById.value.get(preId)
      // 前序步骤未找到，视为未完成
      if (!preStep) return false
      return dependencySatisfiedStatusSet.has(preStep.status)
    })
    if (!allPreDone) return false
  }

  // 串行步骤兜底检查：如果 pre_step_ids 为空但同级存在更早的未完成串行步骤，也不能开始
  if (preStepIds.length === 0 && task.parent_step_id) {
    const parent = workflowStepById.value.get(task.parent_step_id)
    if (parent?.step_type === 'parallel') {
      return true
    }
    const siblings = workflowChildrenByParentId.value.get(task.parent_step_id) || []
    const earlierPendingSibling = siblings.find((t: StepInstance) =>
      t.id !== task.id && t.seq < task.seq && t.step_type === 'serial' && !dependencySatisfiedStatusSet.has(t.status)
    )
    if (earlierPendingSibling) return false
  }

  return true
}

// 选择演练
function selectDrill(drillId: number) {
  selectedDrillId.value = drillId
  currentDrill.value = instances.value.find(i => i.id === drillId) || null
  loadDrillFlowSteps(drillId)
}

function goToDrillTasks(drillId: number) {
  selectDrill(drillId)
  router.replace({ path: '/executor', query: { drill_id: String(drillId) } })
}

// 返回演练列表
function backToDrillList() {
  selectedDrillId.value = null
  currentDrill.value = null
  drillFlowSteps.value = []
  filterStatus.value = ''
  router.replace({ path: '/executor' })
}

// 处理筛选变化
function handleFilterChange() {
  // filter is reactive, filteredTasks computed handles it
}

// 跳转到任务详情
function goToTaskDetail(taskId: number) {
  const task = workflowStepById.value.get(taskId) || tasks.value.find((t: StepInstance) => t.id === taskId)
  router.push({
    path: `/executor/tasks/${taskId}`,
    query: task ? { parent: isParentTask(task) ? '1' : '0' } : undefined,
  })
}

// 查看大屏
function viewScreen(drillId: number) {
  router.push(`/screen/${drillId}`)
}

// 查看大屏2
function viewScreen2(drillId: number | null) {
  if (drillId) router.push(`/executor/screen/${drillId}`)
}

// 加载数据
async function loadTasks(options: { silent?: boolean; lightweight?: boolean } = {}): Promise<void> {
  if (!options.silent) loading.value = true
  try {
    // 加载我的任务
    const rawTasks = await taskApi.getMyTasks()
    // 解析 pre_step_ids 字符串为数组
    tasks.value = rawTasks.map((t: StepInstance) => ({
      ...t,
      pre_step_ids: parsePreStepIds(t.pre_step_ids),
      attributes: parseStepAttributes(t.attributes),
    }))

    if (!options.lightweight) {
      // 加载演练列表
      const drillResult = await drillApi.getList({ page: 1, page_size: 50 })
      instances.value = drillResult.list
    }

    if (selectedDrillId.value) {
      await loadDrillFlowSteps(selectedDrillId.value)
    }

    if (!options.lightweight) {
      // 加载最近活动（从演练日志）
      const allLogs: any[] = []
      for (const drill of instances.value.slice(0, 10)) {
        try {
          const logs = await drillApi.getLogs(drill.id)
          logs.forEach((log: any) => {
            allLogs.push({
              type: log.action,
              drill_name: drill.name,
              operator: log.operator_name || '流程引擎',
              created_at: log.created_at,
            })
          })
        } catch (e) {
          // 忽略错误
        }
      }
      recentActivity.value = allLogs
        .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
        .slice(0, 5)
    }
  } catch (error) {
    ElMessage.error('加载任务失败')
    console.error('Failed to load tasks:', error)
  } finally {
    if (!options.silent) loading.value = false
  }
}

async function loadDrillFlowSteps(drillId: number): Promise<void> {
  try {
    const rawSteps = await drillApi.getSteps(drillId)
    drillFlowSteps.value = rawSteps.map((t: StepInstance) => ({
      ...t,
      pre_step_ids: parsePreStepIds(t.pre_step_ids),
      attributes: parseStepAttributes(t.attributes),
    }))
  } catch (error) {
    console.error('Failed to load drill flow steps:', error)
  }
}

// WebSocket 实时同步
let wsConnections: WebSocket[] = []
let componentDestroyed = false
let refreshTimer: number | null = null
let taskDataLoading = false
let queuedTaskRefresh: { lightweight: boolean } | null = null

async function refreshTasks(options: { silent?: boolean; lightweight?: boolean } = {}) {
  const lightweight = Boolean(options.lightweight)
  if (taskDataLoading) {
    queuedTaskRefresh = {
      lightweight: queuedTaskRefresh ? queuedTaskRefresh.lightweight && lightweight : lightweight,
    }
    return
  }

  taskDataLoading = true
  try {
    await loadTasks(options)
  } finally {
    taskDataLoading = false
    if (queuedTaskRefresh && !componentDestroyed) {
      const next = queuedTaskRefresh
      queuedTaskRefresh = null
      scheduleTaskRefresh(next)
    }
  }
}

function scheduleTaskRefresh(options: { lightweight?: boolean } = {}) {
  if (refreshTimer !== null) {
    window.clearTimeout(refreshTimer)
  }
  refreshTimer = window.setTimeout(() => {
    refreshTimer = null
    refreshTasks({ silent: true, lightweight: options.lightweight }).then(() => {
      if (!options.lightweight) {
        connectAllDrills()
      }
    })
  }, 300)
}

function patchLocalStep(stepId: number, payload: any, eventType: string) {
  const newStatus = payload.new_status || mapEventToStatus(eventType)
  const patchList = (list: StepInstance[]) => {
    const idx = list.findIndex((t: StepInstance) => t.id === stepId)
    if (idx === -1) return list
    const next = [...list]
    next[idx] = {
      ...next[idx],
      ...(newStatus ? { status: newStatus as StepStatus } : {}),
      ...(payload.remark ? { remark: payload.remark } : {}),
      ...(payload.issue_desc ? { issue_desc: payload.issue_desc } : {}),
      ...(payload.start_time ? { start_time: payload.start_time } : {}),
      ...(payload.end_time ? { end_time: payload.end_time } : {}),
      ...(payload.timeout_at ? { timeout_at: payload.timeout_at } : {}),
    }
    return next
  }

  tasks.value = patchList(tasks.value)
  drillFlowSteps.value = patchList(drillFlowSteps.value)
}

function connectDrillWS(drillId: number) {
  const authStore = useAuthStore()
  const wsProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
  const wsUrl = `${wsProtocol}://${window.location.host}/ws/control/${drillId}?token=${authStore.token}`
  const ws = new WebSocket(wsUrl)
  ws.onmessage = (event) => {
    if (componentDestroyed) return
    try {
      const data = JSON.parse(event.data)
      const eventType = data.type || data.event_type
      if (!eventType) return
      const payload = data.payload || data.data || data
      // 增量更新步骤状态
      if (eventType.startsWith('step_') && payload) {
        const stepId = Number(payload.step_id ?? payload.id)
        if (stepId) {
          patchLocalStep(stepId, payload, eventType)
        }
        scheduleTaskRefresh({ lightweight: Boolean(selectedDrillId.value) })
      }
      // 演练状态变化时全量刷新
      if (eventType.startsWith('drill_')) {
        scheduleTaskRefresh({ lightweight: false })
      }
    } catch { /* ignore */ }
  }
  ws.onerror = () => { /* ignore */ }
  wsConnections.push(ws)
}

function mapEventToStatus(eventType: string): string {
  const map: Record<string, string> = {
    step_started: 'running',
    step_complete: 'completed',
    step_issue: 'issue',
    step_skipped: 'skipped',
    step_timeout: 'timeout',
  }
  return map[eventType] || ''
}

function connectAllDrills() {
  // 为每个活跃演练建立 WebSocket 连接
  const activeDrills = instances.value.filter(i => i.status === 'running' || i.status === 'paused')
  for (const drill of activeDrills) {
    if (!wsConnections.some(ws => ws.url.includes(`/ws/control/${drill.id}`))) {
      connectDrillWS(drill.id)
    }
  }
}

function disconnectAllWS() {
  wsConnections.forEach(ws => {
    try { ws.close() } catch { /* ignore */ }
  })
  wsConnections = []
}

onMounted(() => {
  loadTasks().then(() => {
    const queryDrillId = Number(route.query.drill_id)
    if (queryDrillId) {
      selectDrill(queryDrillId)
    }
    connectAllDrills()
  })
})

onBeforeUnmount(() => {
  componentDestroyed = true
  if (refreshTimer !== null) {
    window.clearTimeout(refreshTimer)
  }
  disconnectAllWS()
})
</script>

<style lang="scss" scoped>
@use '@/styles/layout' as *;
@use '@/styles/variables' as *;
@use '@/styles/mixins' as *;

.executor-tasks {
  @include page-container;
}

.page-header {
  @include page-header;
  flex-wrap: wrap;
  gap: $spacing-sm;

  .breadcrumb {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: $font-size-sm;
    color: $text-secondary;
    cursor: pointer;
    padding: 8px 12px;
    border-radius: $radius-base;
    transition: all 0.2s;

    &:hover {
      background: $bg-tertiary;
      color: $text-primary;
    }

    .el-icon {
      font-size: 14px;
    }
  }

  .page-title {
    font-size: $font-size-xl;
    font-weight: $font-weight-bold;
    color: $text-primary;
    margin: 0;
  }

  .header-actions {
    margin-left: auto;
  }

  .filter-group {
    width: 100%;
    margin-top: $spacing-sm;

    :deep(.el-radio-group) {
      display: flex;
      flex-wrap: wrap;
      gap: $spacing-xs;
    }

    :deep(.el-radio-button__inner) {
      background: $bg-secondary;
      border-color: $border-color;
      color: $text-secondary;

      &:hover {
        color: $color-accent;
      }
    }

    :deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
      background: $color-accent;
      border-color: $color-accent;
      color: #fff;
    }
  }
}

.page-content {
  @include page-content;

  .stats-row {
    margin-bottom: $spacing-base;

    .stat-card {
      @include stat-card;
      text-align: center;

      .stat-value {
        font-size: $font-size-xl;
        font-weight: $font-weight-bold;
        color: $text-primary;

        &.pending {
          color: $color-warning;
        }

        &.in-progress {
          color: $color-info;
        }

        &.completed {
          color: $color-success;
        }
      }
    }
  }

  .section-card {
    @include card-compact;
    margin-bottom: $spacing-base;

    .card-title {
      font-size: $font-size-base;
      font-weight: $font-weight-semibold;
      color: $text-primary;
    }

    .drill-card {
      background: $bg-tertiary;
      border-color: $border-color-light;
      margin-bottom: $spacing-sm;
      cursor: pointer;
      transition: all 0.2s;

      &:hover {
        transform: translateY(-2px);
        box-shadow: $shadow-md;
        border-color: $color-accent;
      }

      .drill-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: $spacing-sm;

        .drill-name {
          font-size: $font-size-base;
          font-weight: $font-weight-semibold;
          color: $text-primary;
        }
      }

      .drill-progress {
        margin-bottom: $spacing-sm;
      }

      .drill-tasks-summary {
        display: flex;
        gap: $spacing-base;
        margin-bottom: $spacing-sm;
        padding: $spacing-sm 0;
        border-top: 1px solid $border-color-light;
        border-bottom: 1px solid $border-color-light;

        .task-stat {
          display: flex;
          flex-direction: column;
          align-items: center;

          .label {
            font-size: $font-size-xs;
            color: $text-secondary;
            margin-bottom: 2px;
          }

          .value {
            font-size: $font-size-lg;
            font-weight: $font-weight-bold;
            color: $text-primary;

            &.pending {
              color: $color-warning;
            }
          }
        }
      }

      .drill-actions {
        display: flex;
        gap: $spacing-xs;
        justify-content: flex-end;
      }
    }

    .empty-tip {
      text-align: center;
      color: $text-tertiary;
      padding: $spacing-base;
    }

    :deep(.el-table) {
      background: $bg-secondary;
      color: $text-primary;

      .el-table__header th {
        background: $bg-tertiary;
        color: $text-secondary;
      }

      .el-table__row td {
        background: $bg-secondary;
        border-color: $border-color-light;
      }

      .el-table__row--striped td {
        background: rgba(26, 31, 46, 0.5);
      }
    }
  }
}

.tasks-container {
  min-height: 400px;
}

.flow-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.phase-header {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  padding: $spacing-md 0 $spacing-sm;
  border-bottom: 1px solid $border-color-light;
  margin-bottom: $spacing-sm;

  &.is-first {
    padding-top: 0;
  }

  .phase-icon {
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    flex-shrink: 0;
    color: $text-tertiary;

    svg {
      filter: none;
    }

    .phase-dot {
      width: 10px;
      height: 10px;
      border-radius: 50%;
      background: $border-color;
    }
  }

  .phase-name {
    font-size: $font-size-md;
    font-weight: $font-weight-bold;
    color: $text-primary;
    flex: 1;
  }

  .phase-stats {
    font-size: $font-size-xs;
    color: $text-tertiary;
    background: $bg-tertiary;
    padding: 2px 8px;
    border-radius: 10px;
  }
}

.section-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 14px;
  margin: 10px 0 4px 0;
  background: linear-gradient(90deg, rgba(24, 144, 255, 0.08) 0%, rgba(24, 144, 255, 0.01) 70%, transparent 100%);
  border-left: 3px solid $color-accent;
  border-radius: 0 $radius-base $radius-base 0;

  &.depth-1 {
    margin-left: 0;
  }

  &.depth-2 {
    margin-left: 20px;
    border-left-color: $color-success;
    background: linear-gradient(90deg, rgba(82, 196, 26, 0.07) 0%, rgba(82, 196, 26, 0.01) 70%, transparent 100%);
  }

  .section-badge {
    flex-shrink: 0;
    padding: 2px 8px;
    border-radius: 4px;
    font-size: $font-size-xs;
    font-weight: 500;
    background: rgba(24, 144, 255, 0.12);
    color: $color-accent;
    letter-spacing: 0.04em;

    .depth-2 & {
      background: rgba(82, 196, 26, 0.12);
      color: $color-success;
    }
  }

  .section-name {
    flex: 1;
    font-size: $font-size-base;
    font-weight: 600;
    color: $text-primary;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .section-child-count {
    font-size: $font-size-xs;
    color: $text-tertiary;
    background: $bg-tertiary;
    padding: 1px 8px;
    border-radius: 10px;
    flex-shrink: 0;
  }
}

.flow-item {
  display: flex;
  gap: $spacing-base;
  min-height: 0;

  .flow-rail {
    display: flex;
    flex-direction: column;
    align-items: center;
    width: 32px;
    flex-shrink: 0;

    .flow-dot {
      width: 28px;
      height: 28px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      flex-shrink: 0;
      z-index: 1;
      background: $bg-tertiary;
      border: 2px solid $border-color;
      color: $text-tertiary;
      transition: all 0.3s;

      .dot-inner {
        width: 8px;
        height: 8px;
        border-radius: 50%;
        background: $text-tertiary;
      }

      &.completed {
        background: $color-success-bg;
        border-color: $color-success;
        color: $color-success;
      }

      &.running {
        background: $color-accent-bg;
        border-color: $color-accent;
        color: $color-accent;
        animation: dot-pulse 2s ease-in-out infinite;
      }

      &.timeout {
        background: $color-error-bg;
        border-color: $color-error;
        color: $color-error;
      }

      &.issue {
        background: $color-warning-bg;
        border-color: $color-warning;
        color: $color-warning;
      }

      &.pending {
        background: $bg-tertiary;
        border-color: $border-color;
        color: $text-tertiary;
      }
    }

    .flow-line {
      width: 2px;
      flex: 1;
      min-height: 8px;
      background: $border-color-light;
    }
  }

  &.is-last .flow-rail .flow-line {
    display: none;
  }

  .flow-content {
    flex: 1;
    padding-bottom: $spacing-base;
  }

  .flow-card {
    background: $bg-secondary;
    border: 1px solid $border-color-light;
    border-radius: $radius-md;
    padding: $spacing-md $spacing-base;
    transition: all 0.2s;
    border-left: 3px solid $border-color;
    cursor: pointer;

    &:hover {
      box-shadow: $shadow-md;
      border-color: $color-accent-border;
    }

    &.status-in-progress {
      border-left-color: $color-accent;
    }

    &.status-issued {
      border-left-color: $color-error;
    }

    &.status-completed {
      border-left-color: $color-success;
      opacity: 0.75;
    }

    .flow-card-header {
      display: flex;
      align-items: center;
      gap: $spacing-sm;
      margin-bottom: $spacing-sm;

      .flow-step-name {
        font-size: $font-size-base;
        font-weight: $font-weight-semibold;
        color: $text-primary;
        flex: 1;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }

    .flow-card-body {
      margin-bottom: $spacing-sm;

      .flow-meta-row {
        display: flex;
        flex-wrap: wrap;
        gap: 6px;
        margin-bottom: $spacing-xs;
        align-items: center;

        .flow-phase {
          font-size: $font-size-xs;
          color: $text-tertiary;
          margin-left: 4px;
        }
      }

      .flow-meta {
        display: flex;
        flex-wrap: wrap;
        gap: $spacing-md;
        font-size: $font-size-xs;
        color: $text-tertiary;

        .flow-duration {
          color: $text-secondary;
        }

        .flow-deadline {
          display: flex;
          align-items: center;
          gap: 4px;
          color: $color-warning;

          .el-icon {
            font-size: 14px;
          }
        }
      }
    }

    .flow-card-footer {
      display: flex;
      align-items: center;
      gap: $spacing-xs;
      padding-top: $spacing-sm;
      border-top: 1px solid $border-color-light;
    }
  }
}

@keyframes dot-pulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba($color-accent, 0.3); }
  50% { box-shadow: 0 0 0 6px rgba($color-accent, 0); }
}
</style>
