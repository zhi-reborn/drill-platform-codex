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
              <el-card class="drill-card" @click="selectDrill(drill.id)">
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
                  <el-button type="primary" size="small">
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
        </div>
        <div class="filter-group">
          <el-radio-group v-model="filterStatus" size="default" @change="handleFilterChange">
            <el-radio-button value="">全部</el-radio-button>
            <el-radio-button value="pending">待执行</el-radio-button>
            <el-radio-button value="assigned">已分配</el-radio-button>
            <el-radio-button value="in_progress">执行中</el-radio-button>
            <el-radio-button value="completed">已完成</el-radio-button>
            <el-radio-button value="issued">异常</el-radio-button>
          </el-radio-group>
        </div>
      </div>

      <div v-loading="loading" class="tasks-container">
        <EmptyBox v-if="!loading && filteredTasks.length === 0" title="暂无任务" description="当前没有分配给您的任务" />

        <el-row v-else :gutter="20" class="tasks-grid">
          <el-col v-for="task in filteredTasks" :key="task.id" :xs="24" :sm="12" :lg="8" :xl="6">
            <div class="task-card" :class="getStatusClass(task.status)">
              <div class="task-card-header">
                <div class="task-step-name">{{ task.name }}</div>
                <DrillStatusBadge :status="task.status" type="step" />
              </div>

              <div class="task-card-body">
                <div class="task-meta-row">
                  <el-tag v-if="task.default_assignee_role === 'director'" size="small" type="warning">指挥组</el-tag>
                  <el-tag v-else-if="task.default_assignee_role" size="small" type="primary">{{ task.default_assignee_role }}</el-tag>
                  <el-tag v-if="task.executor_team" size="small" type="info">{{ task.executor_team }}</el-tag>
                  <span v-if="task.phase_step" class="task-phase">{{ task.phase_step }}</span>
                </div>

                <div class="task-meta">
                  <div v-if="task.estimated_duration_minutes" class="task-duration">
                    预计耗时：{{ task.estimated_duration_minutes }} 分钟
                  </div>
                  <div v-if="task.timeout_at" class="task-deadline">
                    <el-icon><Clock /></el-icon>
                    <span>截止：{{ formatDeadline(task.timeout_at) }}</span>
                  </div>
                </div>
              </div>

              <div class="task-card-footer">
                <el-button
                  v-if="task.status === 'pending'"
                  type="info"
                  class="action-btn"
                  disabled
                >
                  等待中
                </el-button>
                <el-button
                  v-else-if="task.status === 'running' && !isParentTask(task)"
                  type="primary"
                  class="action-btn"
                  @click="goToTaskDetail(task.id)"
                >
                  开始执行
                </el-button>
                <el-button
                  v-else-if="task.status === 'running'"
                  class="action-btn"
                  disabled
                >
                  子任务执行中
                </el-button>
                <el-button
                  v-else-if="task.status === 'issue'"
                  type="warning"
                  class="action-btn"
                  @click="goToTaskDetail(task.id)"
                >
                  查看异常
                </el-button>
                <el-button v-else disabled class="action-btn">
                  已完成
                </el-button>
              </div>
            </div>
          </el-col>
        </el-row>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Clock, User, Monitor, ArrowLeft, ArrowRight } from '@element-plus/icons-vue'
import type { StepInstance } from '@/types/instance'
import type { DrillInstance } from '@/types'
import DrillStatusBadge from '@/components/common/DrillStatusBadge.vue'
import EmptyBox from '@/components/common/EmptyBox.vue'
import { taskApi } from '@/api/modules/task'
import { drillApi } from '@/api/modules/drill'

const router = useRouter()

const loading = ref(false)
const tasks = ref<StepInstance[]>([])
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
  return result
})

const recentActivity = ref<any[]>([])

const getStatusClass = (status: string) => {
  const classMap: Record<string, string> = {
    running: 'status-in-progress',
    issue: 'status-issued',
    completed: 'status-completed',
    timeout: 'status-issued',
  }
  return classMap[status] || ''
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
  return tasks.value.some((t: StepInstance) => t.parent_step_id === task.id)
}

// 选择演练
function selectDrill(drillId: number) {
  selectedDrillId.value = drillId
  currentDrill.value = instances.value.find(i => i.id === drillId) || null
}

// 返回演练列表
function backToDrillList() {
  selectedDrillId.value = null
  currentDrill.value = null
  filterStatus.value = ''
}

// 处理筛选变化
function handleFilterChange() {
  // filter is reactive, filteredTasks computed handles it
}

// 跳转到任务详情
function goToTaskDetail(taskId: number) {
  router.push(`/executor/tasks/${taskId}`)
}

// 查看大屏
function viewScreen(drillId: number) {
  router.push(`/screen/${drillId}`)
}

// 加载数据
async function loadTasks() {
  loading.value = true
  try {
    // 加载我的任务
    tasks.value = await taskApi.getMyTasks()
    
    // 加载演练列表
    const drillResult = await drillApi.getList({ page: 1, page_size: 50 })
    instances.value = drillResult.list
    
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
  } catch (error) {
    ElMessage.error('加载任务失败')
    console.error('Failed to load tasks:', error)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadTasks()
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

.tasks-grid {
  .task-card {
    @include card-compact;
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 180px;
    border-left-width: 4px;
    transition: transform 0.2s, box-shadow 0.2s;

    &:hover {
      transform: translateY(-2px);
      box-shadow: $shadow-md;
    }

    &.status-in-progress {
      border-left-color: #55c3d3;
    }

    &.status-issued {
      border-left-color: #da3633;
    }

    &.status-completed {
      border-left-color: #2ea043;
    }

    .task-card-header {
      @include flex-between;
      margin-bottom: $spacing-sm;
      gap: $spacing-xs;

      .task-step-name {
        font-size: $font-size-base;
        font-weight: $font-weight-semibold;
        color: $text-primary;
        @include ellipsis(1);
        flex: 1;
      }
    }

    .task-card-body {
      flex: 1;
      margin-bottom: $spacing-sm;

      .task-meta-row {
        display: flex;
        flex-wrap: wrap;
        gap: 6px;
        margin-bottom: $spacing-xs;
        align-items: center;

        .task-phase {
          font-size: $font-size-xs;
          color: $text-tertiary;
          margin-left: 4px;
        }
      }

      .task-step-name {
        font-size: $font-size-base;
        font-weight: $font-weight-medium;
        color: $text-primary;
        margin-bottom: $spacing-xs;
      }

      .task-description {
        font-size: $font-size-sm;
        color: $text-secondary;
        margin-bottom: $spacing-sm;
        @include ellipsis(2);
        line-height: 1.6;
      }

      .task-meta {
        display: flex;
        flex-direction: column;
        gap: $spacing-xs;
        font-size: $font-size-xs;
        color: $text-tertiary;

        .task-duration {
          color: $text-secondary;
        }

        .task-deadline,
        .task-assignee {
          display: flex;
          align-items: center;
          gap: 4px;

          .el-icon {
            font-size: 14px;
          }
        }

        .task-deadline {
          color: $color-warning;
        }
      }
    }

    .task-card-footer {
      border-top: 1px solid $border-color-light;
      padding-top: $spacing-sm;

      .action-btn {
        width: 100%;
      }
    }
  }
}
</style>
