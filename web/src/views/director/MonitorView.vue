<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">{{ instance?.name || '演练监控' }}</h2>
      <DrillStatusBadge v-if="instance" :status="instance.status" type="drill" />
    </div>

    <div v-if="instance" class="page-content">
      <!-- 顶部信息 -->
      <el-card class="info-card">
        <div class="progress-section">
          <span class="progress-label">
            进度：{{ instance.completed_steps }} / {{ instance.total_steps }}
          </span>
          <el-progress
            :percentage="Math.round(instance.completed_steps / instance.total_steps * 100)"
            :stroke-width="10"
            :status="instance.status === 'completed' ? 'success' : undefined"
          />
        </div>
      </el-card>

      <!-- 控制按钮 -->
      <div class="control-section">
        <ActionConfirm
          title="暂停演练"
          message="确定要暂停当前演练吗？"
          type="warning"
          @confirm="handlePause"
        >
          <el-icon><VideoPause /></el-icon>
          暂停
        </ActionConfirm>
        <ActionConfirm
          title="继续演练"
          message="确定要继续执行演练吗？"
          type="primary"
          @confirm="handleResume"
        >
          <el-icon><VideoPlay /></el-icon>
          继续
        </ActionConfirm>
        <ActionConfirm
          title="终止演练"
          message="确定要终止当前演练吗？此操作不可恢复！"
          danger
          @confirm="handleTerminate"
        >
          <el-icon><VideoCamera /></el-icon>
          终止
        </ActionConfirm>
      </div>

      <!-- 步骤列表 -->
      <el-card class="steps-card">
        <template #header>
          <span class="card-title">步骤列表</span>
        </template>
        <el-table :data="drillSteps" style="width: 100%">
          <el-table-column prop="order_index" label="序号" width="80" />
          <el-table-column prop="step_name" label="步骤名" min-width="200" />
          <el-table-column prop="status" label="状态" width="120">
            <template #default="{ row }">
              <DrillStatusBadge :status="row.status" type="step" />
            </template>
          </el-table-column>
          <el-table-column prop="assignee_name" label="执行人" width="120" />
          <el-table-column label="耗时" width="120">
            <template #default="{ row }">
              {{ calculateDuration(row) }}
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 操作日志 -->
      <el-card class="logs-card">
        <template #header>
          <span class="card-title">操作日志</span>
        </template>
        <el-timeline>
          <el-timeline-item
            v-for="log in drillLogs"
            :key="log.id"
            :timestamp="formatTime(log.created_at)"
            placement="top"
          >
            <el-card>
              <div class="log-content">
                <el-tag :type="getLogTypeTag(log.action)" size="small">
                  {{ getLogActionLabel(log.action) }}
                </el-tag>
                <span class="log-step">步骤：{{ log.step_name }}</span>
                <span v-if="log.operator" class="log-operator">操作人：{{ log.operator }}</span>
              </div>
              <p v-if="log.remark" class="log-remark">{{ log.remark }}</p>
            </el-card>
          </el-timeline-item>
        </el-timeline>
        <div v-if="drillLogs.length === 0" class="empty-tip">
          暂无操作日志
        </div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { VideoPause, VideoPlay, VideoCamera } from '@element-plus/icons-vue'
import type { DrillInstance, StepInstance } from '@/types'
import DrillStatusBadge from '@/components/common/DrillStatusBadge.vue'
import ActionConfirm from '@/components/common/ActionConfirm.vue'
import instancesData from '@/mock/data/instances.json'
import stepsData from '@/mock/data/steps.json'

const route = useRoute()
const router = useRouter()

const instance = ref<DrillInstance | null>(null)
const steps = ref<StepInstance[]>([])

const drillId = computed(() => {
  const id = route.params.id
  return typeof id === 'string' ? parseInt(id, 10) : 0
})

const drillSteps = computed(() => {
  return steps.value.filter(s => s.drill_id === drillId.value).sort((a, b) => a.order_index - b.order_index)
})

// 模拟日志数据（实际应从 API 获取）
const drillLogs = computed(() => {
  const logs: any[] = []
  drillSteps.value.forEach((step, index) => {
    if (step.started_at) {
      logs.push({
        id: `log-${step.id}-start`,
        action: 'step_start',
        step_name: step.step_name,
        operator: step.assignee_name,
        created_at: step.started_at,
        remark: `开始执行步骤`,
      })
    }
    if (step.completed_at) {
      logs.push({
        id: `log-${step.id}-complete`,
        action: 'step_complete',
        step_name: step.step_name,
        operator: step.assignee_name,
        created_at: step.completed_at,
        remark: step.result_json ? `执行结果：${step.result_json}` : '步骤执行完成',
      })
    }
  })
  return logs.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
})

function calculateDuration(step: StepInstance): string {
  if (!step.started_at || !step.completed_at) {
    return '-'
  }
  const start = new Date(step.started_at).getTime()
  const end = new Date(step.completed_at).getTime()
  const diff = Math.floor((end - start) / 1000)
  if (diff < 60) {
    return `${diff}s`
  }
  const mins = Math.floor(diff / 60)
  const secs = diff % 60
  return `${mins}m ${secs}s`
}

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

function getLogTypeTag(action: string): 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  const map: Record<string, any> = {
    step_start: 'primary',
    step_complete: 'success',
    step_issue: 'danger',
    step_skip: 'info',
    force_complete: 'warning',
  }
  return map[action] || 'info'
}

function getLogActionLabel(action: string): string {
  const map: Record<string, string> = {
    step_start: '步骤开始',
    step_complete: '步骤完成',
    step_issue: '步骤异常',
    step_skip: '步骤跳过',
    force_complete: '强制完成',
  }
  return map[action] || action
}

async function loadDrillData() {
  try {
    instance.value = instancesData.find(i => i.id === drillId.value) as DrillInstance || null
    steps.value = stepsData as StepInstance[]
    if (!instance.value) {
      ElMessage.error('演练不存在')
      router.back()
    }
  } catch (error) {
    ElMessage.error('加载数据失败')
    console.error('Failed to load drill data:', error)
  }
}

function handlePause() {
  ElMessage.success('演练已暂停')
}

function handleResume() {
  ElMessage.success('演练已继续')
}

function handleTerminate() {
  ElMessage.success('演练已终止')
  router.push('/director/dashboard')
}

onMounted(() => {
  loadDrillData()
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

    .info-card {
      @include card-compact;
      margin-bottom: $spacing-base;

      .progress-section {
        .progress-label {
          display: block;
          font-size: $font-size-sm;
          color: $text-secondary;
          margin-bottom: $spacing-sm;
        }
      }
    }

    .control-section {
      display: flex;
      gap: $spacing-sm;
      margin-bottom: $spacing-base;

      :deep(.el-button) {
        display: flex;
        align-items: center;
        gap: 6px;
      }
    }

    .steps-card,
    .logs-card {
      @include card-compact;
      margin-bottom: $spacing-base;

      .card-title {
        font-size: $font-size-base;
        font-weight: $font-weight-semibold;
        color: $text-primary;
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

    .logs-card {
      .log-content {
        display: flex;
        align-items: center;
        gap: $spacing-sm;
        flex-wrap: wrap;
        font-size: $font-size-sm;

        .log-step,
        .log-operator {
          color: $text-secondary;
        }
      }

      .log-remark {
        margin-top: $spacing-xs;
        font-size: $font-size-sm;
        color: $text-tertiary;
        line-height: 1.6;
      }

      .empty-tip {
        text-align: center;
        color: $text-tertiary;
        padding: $spacing-base;
      }
    }
  }
}
</style>
