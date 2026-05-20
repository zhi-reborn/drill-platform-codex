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
          <el-descriptions-item v-if="task.executor_team" label="执行组">
            <el-tag type="info">{{ task.executor_team }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="分配给">
            {{ assignedNames || '未分配' }}
          </el-descriptions-item>
          <el-descriptions-item v-if="task.timeout_at" label="超时时间">
            {{ formatDeadline(task.timeout_at) }}
          </el-descriptions-item>
          <el-descriptions-item v-if="task.start_time" label="开始时间">
            {{ formatTime(task.start_time) }}
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
            v-if="task.status === 'running'"
            type="success"
            @click="handleComplete"
          >
            <el-icon><CircleCheck /></el-icon>
            完成
          </el-button>
          <el-button
            v-if="task.status === 'running'"
            type="danger"
            @click="handleReportIssue"
          >
            <el-icon><Warning /></el-icon>
            上报异常
          </el-button>
          <el-tag v-if="task.status !== 'running'" type="info">
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
import type { StepInstance } from '@/types/instance'

const route = useRoute()
const router = useRouter()

const loading = ref(false)
const task = ref<StepInstance | null>(null)

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
    task.value = await taskApi.getById(stepId.value)
  } catch (error) {
    ElMessage.error('加载任务失败')
    console.error('Failed to load task:', error)
  } finally {
    loading.value = false
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
