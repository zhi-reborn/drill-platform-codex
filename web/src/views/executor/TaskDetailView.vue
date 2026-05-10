<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">任务详情</h2>
      <el-button @click="router.back()">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
    </div>

    <div v-if="task" class="page-content">
      <!-- 任务信息 -->
      <el-card class="info-card">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="演练名称">{{ task.drill_name }}</el-descriptions-item>
          <el-descriptions-item label="步骤名称">{{ task.step_name }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <DrillStatusBadge :status="task.status" type="step" />
          </el-descriptions-item>
          <el-descriptions-item label="执行人">{{ task.assigned_to_name }}</el-descriptions-item>
          <el-descriptions-item v-if="task.deadline" label="截止时间">
            {{ formatDeadline(task.deadline) }}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatTime(task.created_at) }}
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 步骤详情 -->
      <el-card class="detail-card">
        <template #header>
          <span class="card-title">步骤详情</span>
        </template>
        <div class="step-detail">
          <div class="detail-section">
            <h4>步骤描述</h4>
            <p class="description">{{ task.step_description }}</p>
          </div>
          <div v-if="task.script" class="detail-section">
            <h4>脚本代码</h4>
            <pre class="code-block">{{ task.script }}</pre>
          </div>
          <div v-if="stepInstance?.error_message" class="detail-section error">
            <h4>错误信息</h4>
            <p class="error-message">{{ stepInstance.error_message }}</p>
          </div>
        </div>
      </el-card>

      <!-- 操作表单 -->
      <el-card class="action-card">
        <template #header>
          <span class="card-title">执行操作</span>
        </template>
        <el-form :model="actionForm" label-width="80px">
          <el-form-item label="执行结果">
            <el-input
              v-model="actionForm.result"
              type="textarea"
              :rows="4"
              placeholder="请输入执行结果（可选）"
            />
          </el-form-item>
          <el-form-item label="备注">
            <el-input
              v-model="actionForm.remark"
              type="textarea"
              :rows="2"
              placeholder="请输入备注说明（可选）"
            />
          </el-form-item>
        </el-form>
        <div class="action-buttons">
          <ActionConfirm
            title="完成任务"
            message="确定要标记此任务为完成状态吗？"
            type="success"
            @confirm="handleComplete"
          >
            <el-icon><CircleCheck /></el-icon>
            完成
          </ActionConfirm>
          <ActionConfirm
            title="上报异常"
            message="确定要上报此任务异常吗？"
            type="danger"
            @confirm="handleReportIssue"
          >
            <el-icon><Warning /></el-icon>
            上报异常
          </ActionConfirm>
          <ActionConfirm
            title="跳过任务"
            message="确定要跳过此任务吗？"
            type="warning"
            @confirm="handleSkip"
          >
            <el-icon><CircleClose /></el-icon>
            跳过
          </ActionConfirm>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft, CircleCheck, Warning, CircleClose } from '@element-plus/icons-vue'
import type { Task, StepInstance } from '@/types'
import DrillStatusBadge from '@/components/common/DrillStatusBadge.vue'
import ActionConfirm from '@/components/common/ActionConfirm.vue'
import tasksData from '@/mock/data/tasks.json'
import stepsData from '@/mock/data/steps.json'

const route = useRoute()
const router = useRouter()

const task = ref<Task | null>(null)
const stepInstance = ref<StepInstance | null>(null)

const actionForm = ref({
  result: '',
  remark: '',
})

const taskId = computed(() => {
  const id = route.params.id
  return typeof id === 'string' ? parseInt(id, 10) : 0
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

async function loadTaskData() {
  try {
    task.value = (tasksData.find(t => t.id === taskId.value) as Task) || null
    if (task.value) {
      stepInstance.value = (stepsData.find(s => s.id === task.value!.step_id) as StepInstance) || null
    }
    if (!task.value) {
      ElMessage.error('任务不存在')
      router.back()
    }
  } catch (error) {
    ElMessage.error('加载数据失败')
    console.error('Failed to load task data:', error)
  }
}

function handleComplete() {
  ElMessage.success('任务已完成')
  router.back()
}

function handleReportIssue() {
  if (!actionForm.value.remark.trim()) {
    ElMessage.warning('请填写异常说明')
    return
  }
  ElMessage.success('异常已上报')
  router.back()
}

function handleSkip() {
  ElMessage.success('任务已跳过')
  router.back()
}

onMounted(() => {
  loadTaskData()
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
