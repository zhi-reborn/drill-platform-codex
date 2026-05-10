<template>
  <div class="executor-tasks">
    <div class="page-header">
      <h2 class="page-title">我的任务</h2>
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
              <div class="task-drill-name">{{ task.drill_name }}</div>
              <DrillStatusBadge :status="task.status" type="step" />
            </div>

            <div class="task-card-body">
              <div class="task-step-name">{{ task.step_name }}</div>
              <div class="task-description">{{ task.step_description }}</div>

              <div class="task-meta">
                <div v-if="task.deadline" class="task-deadline">
                  <el-icon><Clock /></el-icon>
                  <span>截止：{{ formatDeadline(task.deadline) }}</span>
                </div>
                <div class="task-assignee">
                  <el-icon><User /></el-icon>
                  <span>{{ task.assigned_to_name }}</span>
                </div>
              </div>
            </div>

            <div class="task-card-footer">
              <el-button
                v-if="task.status === 'pending' || task.status === 'assigned'"
                type="primary"
                class="action-btn"
                @click="goToTaskDetail(task.id)"
              >
                开始执行
              </el-button>
              <el-button
                v-else-if="task.status === 'in_progress'"
                type="primary"
                class="action-btn"
                @click="goToTaskDetail(task.id)"
              >
                继续
              </el-button>
              <el-button
                v-else-if="task.status === 'issued'"
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
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Clock, User } from '@element-plus/icons-vue'
import { taskApi } from '@/api/modules/task'
import type { Task } from '@/types'
import DrillStatusBadge from '@/components/common/DrillStatusBadge.vue'
import EmptyBox from '@/components/common/EmptyBox.vue'
import tasksData from '@/mock/data/tasks.json'

const router = useRouter()

const loading = ref(false)
const tasks = ref<Task[]>([])
const filterStatus = ref<string>('')

const filteredTasks = computed(() => {
  if (!filterStatus.value) {
    return tasks.value
  }
  return tasks.value.filter((task) => task.status === filterStatus.value)
})

const getStatusClass = (status: string) => {
  const classMap: Record<string, string> = {
    in_progress: 'status-in-progress',
    issued: 'status-issued',
    completed: 'status-completed',
  }
  return classMap[status] || ''
}

// 处理筛选变化
function handleFilterChange() {
  // filter is reactive, filteredTasks computed handles it
}

// 格式化截止时间
function formatDeadline(d: string): string {
  if (!d) return '无'
  const date = new Date(d)
  const now = new Date()
  const diff = date.getTime() - now.getTime()
  const hours = Math.floor(diff / (1000 * 60 * 60))

  if (hours < 0) {
    return '已过期'
  }
  if (hours < 1) {
    return '1 小时内'
  }
  return `${hours}小时`
}

// 跳转到任务详情
function goToTaskDetail(taskId: number) {
  router.push(`/executor/tasks/${taskId}`)
}

// 加载数据
async function loadTasks() {
  loading.value = true
  try {
    // 使用 mock 数据
    tasks.value = tasksData as Task[]
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

  .page-title {
    font-size: $font-size-xl;
    font-weight: $font-weight-bold;
    color: $text-primary;
    margin: 0;
  }

  .filter-group {
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
        color: $color-primary;
      }
    }

    :deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
      background: $color-primary;
      border-color: $color-primary;
      color: #fff;
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

      .task-drill-name {
        font-size: $font-size-sm;
        font-weight: $font-weight-semibold;
        color: $text-primary;
        @include ellipsis(1);
        flex: 1;
      }
    }

    .task-card-body {
      flex: 1;
      margin-bottom: $spacing-sm;

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
