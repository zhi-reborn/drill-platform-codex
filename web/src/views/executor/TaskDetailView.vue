<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">任务详情</h2>
      <el-button @click="router.back()">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
    </div>

    <div v-loading="loading" v-if="task" class="page-content">
      <el-card class="info-card">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="演练名称">{{ task.drill_instance?.name || '' }}</el-descriptions-item>
          <el-descriptions-item label="步骤名称">{{ task.name }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <DrillStatusBadge :status="task.status" type="step" />
          </el-descriptions-item>
          <el-descriptions-item label="预计耗时">
            {{ task.estimated_duration_minutes ? `${task.estimated_duration_minutes} 分钟` : '-' }}
          </el-descriptions-item>
          <el-descriptions-item label="执行角色">
            <el-tag :type="task.default_assignee_role === 'director' ? 'warning' : 'primary'" size="small">
              {{ task.default_assignee_role === 'director' ? '指挥组' : '执行组' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item v-if="task.executor_team" label="执行组">
            <el-tag type="info">{{ task.executor_team }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item v-if="task.phase_step" label="环节">
            {{ task.phase_step }}
          </el-descriptions-item>
          <el-descriptions-item v-if="task.timeout_minutes" label="超时限制">
            {{ task.timeout_minutes }} 分钟
          </el-descriptions-item>
          <el-descriptions-item v-if="task.timeout_at" label="超时时间">
            {{ formatDeadline(task.timeout_at) }}
          </el-descriptions-item>
          <el-descriptions-item v-if="task.start_time" label="开始时间">
            {{ formatTime(task.start_time) }}
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 操作信息 -->
      <el-card v-if="task.attributes" class="detail-card">
        <template #header>
          <span class="card-title">详细信息</span>
        </template>
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item v-if="task.attributes.responsible_department" label="责任部门">
            {{ task.attributes.responsible_department }}
          </el-descriptions-item>
          <el-descriptions-item v-if="task.attributes.cooperating_department" label="配合部门">
            {{ task.attributes.cooperating_department }}
          </el-descriptions-item>
          <el-descriptions-item v-if="task.attributes.responsible_team" label="责任团队">
            {{ task.attributes.responsible_team }}
          </el-descriptions-item>
          <el-descriptions-item v-if="task.attributes.operator" label="操作人">
            {{ task.attributes.operator }}
          </el-descriptions-item>
          <el-descriptions-item v-if="task.attributes.reviewer" label="复核人">
            {{ task.attributes.reviewer }}
          </el-descriptions-item>
          <el-descriptions-item v-if="task.attributes.operation_guide" label="操作说明" :span="2">
            {{ task.attributes.operation_guide }}
          </el-descriptions-item>
          <el-descriptions-item v-if="task.attributes.verification_method" label="验证方式" :span="2">
            {{ task.attributes.verification_method }}
          </el-descriptions-item>
          <el-descriptions-item v-if="task.attributes.worst_case_analysis" label="最坏影响分析" :span="2">
            {{ task.attributes.worst_case_analysis }}
          </el-descriptions-item>
          <el-descriptions-item v-if="task.attributes.fallback_measures" label="兜底措施" :span="2">
            {{ task.attributes.fallback_measures }}
          </el-descriptions-item>
          <el-descriptions-item v-if="task.attributes.remark" label="备注" :span="2">
            {{ task.attributes.remark }}
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <el-card v-if="task.remark" class="detail-card">
        <template #header>
          <span class="card-title">执行备注</span>
        </template>
        <p>{{ task.remark }}</p>
      </el-card>

      <el-card class="action-card">
        <template #header>
          <span class="card-title">执行操作</span>
        </template>
        <el-form :model="actionForm" label-width="80px">
          <el-form-item label="备注">
            <el-input
              v-model="actionForm.remark"
              type="textarea"
              :rows="4"
              placeholder="请输入执行备注（可选）"
            />
          </el-form-item>
        </el-form>
        <div class="action-buttons">
          <el-button
            v-if="task.status === 'pending' && !isParentTask"
            type="primary"
            @click="handleStart"
          >
            开始执行
          </el-button>
          <el-button
            v-if="task.status === 'running' && !isParentTask"
            type="success"
            @click="handleComplete"
          >
            <el-icon><CircleCheck /></el-icon>
            完成
          </el-button>
          <el-button
            v-if="task.status === 'running' && !isParentTask"
            type="danger"
            @click="handleReportIssue"
          >
            <el-icon><Warning /></el-icon>
            上报异常
          </el-button>
          <span v-if="task.status === 'running' && isParentTask" class="parent-task-hint">
            父任务 · 子任务全部完成后自动完成
          </span>
          <el-tag v-if="task.status !== 'running' && task.status !== 'pending'" type="info">
            {{ task.status === 'completed' ? '已完成' : task.status }}
          </el-tag>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, CircleCheck, Warning } from '@element-plus/icons-vue'
import DrillStatusBadge from '@/components/common/DrillStatusBadge.vue'
import { taskApi } from '@/api/modules/task'
import { drillApi } from '@/api/modules/drill'
import type { StepInstance } from '@/types/instance'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const task = ref<StepInstance | null>(null)
const isParentTask = ref(false)

const actionForm = ref({
  remark: '',
})

const stepId = computed(() => {
  const id = route.params.id
  return typeof id === 'string' ? parseInt(id, 10) : 0
})

const assignedNames = computed(() => {
  if (!task.value) return '未分配'
  if (task.value.assignee_names) return task.value.assignee_names
  if (task.value.executor_team) return task.value.executor_team
  return '未分配'
})

function formatDeadline(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
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

async function loadTask() {
  if (!stepId.value) return
  loading.value = true
  try {
    const data = await taskApi.getById(stepId.value)
    // 解析 attributes JSON 字符串
    if (data && typeof data.attributes === 'string') {
      try {
        data.attributes = JSON.parse(data.attributes)
      } catch { /* ignore parse error */ }
    }
    task.value = data

    // 检查是否为父任务：加载该演练所有步骤，确认是否有子步骤指向本任务
    if (data && data.drill_instance_id) {
      try {
        const allSteps = await drillApi.getSteps(data.drill_instance_id)
        isParentTask.value = allSteps.some(
          (s: StepInstance) => s.parent_step_id === data.id
        )
      } catch {
        isParentTask.value = false
      }
    }
  } catch (error) {
    ElMessage.error('加载任务失败')
    console.error('Failed to load task:', error)
  } finally {
    loading.value = false
  }
}

async function handleStart() {
  if (!task.value) return
  try {
    await taskApi.start(stepId.value)
    ElMessage.success('任务已开始')
    await loadTask()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || error.message || '操作失败')
  }
}

async function handleComplete() {
  if (!task.value) return
  try {
    await ElMessageBox.confirm('确定要标记此任务为完成状态吗？', '确认完成', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await taskApi.complete(stepId.value, actionForm.value.remark)
    ElMessage.success('任务已完成')
    await loadTask()
  } catch (error: any) {
    if (error === 'cancel') return
    ElMessage.error(error.response?.data?.message || error.message || '操作失败')
  }
}

async function handleReportIssue() {
  if (!task.value) return
  if (!actionForm.value.remark.trim()) {
    ElMessage.warning('请填写异常说明')
    return
  }
  try {
    await ElMessageBox.confirm('确定要上报此任务异常吗？', '确认上报', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await taskApi.reportIssue(stepId.value, actionForm.value.remark)
    ElMessage.success('异常已上报')
    await loadTask()
  } catch (error: any) {
    if (error === 'cancel') return
    ElMessage.error(error.response?.data?.message || error.message || '上报失败')
  }
}

onMounted(() => {
  loadTask()
})
</script>

<style scoped lang="scss">
@use '@/styles/layout' as *;
@use '@/styles/variables' as *;

.page-container {
  @include page-container;

  .page-header {
    @include page-header;

    .page-title {
      font-size: $font-size-xl;
      font-weight: $font-weight-bold;
      color: $text-primary;
      margin: 0;
    }
  }

  .page-content {
    @include page-content;
    display: flex;
    flex-direction: column;
    gap: $spacing-base;

    .info-card,
    .detail-card,
    .action-card {
      @include card-compact;

      .card-title {
        font-size: $font-size-base;
        font-weight: $font-weight-semibold;
        color: $text-primary;
      }

      :deep(.el-descriptions__label) {
        background: $bg-tertiary;
        color: $text-secondary;
      }

      :deep(.el-descriptions__content) {
        background: $bg-secondary;
        color: $text-primary;
      }
    }

    .action-card {
      .action-buttons {
        display: flex;
        gap: $spacing-sm;
        margin-top: $spacing-base;
        align-items: center;

        :deep(.el-button) {
          display: flex;
          align-items: center;
          gap: 6px;
        }

        .parent-task-hint {
          font-size: $font-size-sm;
          color: $text-tertiary;
          font-style: italic;
        }
      }

      :deep(.el-textarea__inner) {
        background: $bg-tertiary;
        border-color: $border-color;
        color: $text-primary;
      }
    }
  }
}
</style>
