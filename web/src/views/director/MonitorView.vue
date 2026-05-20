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
            进度：{{ completedSteps }} / {{ totalSteps }}
          </span>
          <el-progress
            :percentage="progressPercentage"
            :stroke-width="10"
            :status="instance.status === 'completed' ? 'success' : undefined"
          />
        </div>
      </el-card>

      <!-- 控制按钮 -->
      <div class="control-section">
        <!-- 运行中：显示暂停、终止 -->
        <ActionConfirm
          v-if="canPause"
          title="暂停演练"
          message="确定要暂停当前演练吗？"
          type="warning"
          @confirm="handlePause"
        >
          <el-icon><VideoPause /></el-icon>
          暂停
        </ActionConfirm>
        <!-- 已暂停：显示继续 -->
        <ActionConfirm
          v-if="canResume"
          title="继续演练"
          message="确定要继续执行演练吗？"
          type="primary"
          @confirm="handleResume"
        >
          <el-icon><VideoPlay /></el-icon>
          继续
        </ActionConfirm>
        <!-- 待启动：显示开始 -->
        <ActionConfirm
          v-if="canStart"
          title="开始演练"
          message="确定要开始当前演练吗？"
          type="success"
          @confirm="handleStart"
        >
          <el-icon><VideoPlay /></el-icon>
          开始
        </ActionConfirm>
        <!-- 终止按钮：运行中/已暂停/待启动 可终止 -->
        <ActionConfirm
          v-if="canTerminate"
          title="终止演练"
          message="确定要终止当前演练吗？此操作不可恢复！"
          danger
          @confirm="handleTerminate"
        >
          <el-icon><VideoCamera /></el-icon>
          终止
        </ActionConfirm>
        <!-- 已完成/已终止：显示返回 -->
        <el-button v-if="isFinished" @click="router.back()">
          <el-icon><Back /></el-icon>
          返回
        </el-button>
      </div>

      <!-- 步骤列表 -->
      <el-card class="steps-card">
        <template #header>
          <span class="card-title">步骤列表</span>
        </template>
        <el-table :data="drillSteps" style="width: 100%">
          <el-table-column prop="seq" label="序号" width="60" align="center" />
          <el-table-column prop="name" label="步骤名" min-width="180" show-overflow-tooltip />
          <el-table-column prop="step_type" label="类型" width="80" align="center">
            <template #default="{ row }">
              <el-tag :type="getStepTypeTag(row.step_type)" size="small">{{ getStepTypeLabel(row.step_type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100" align="center">
            <template #default="{ row }">
              <DrillStatusBadge :status="row.status" type="step" />
            </template>
          </el-table-column>
          <el-table-column prop="default_assignee_role" label="执行角色" width="80" align="center">
            <template #default="{ row }">
              <el-tag :type="row.default_assignee_role === 'director' ? 'warning' : 'primary'" size="small">
                {{ row.default_assignee_role === 'director' ? '指挥组' : '执行组' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="executor_team" label="执行团队" width="120" align="center" show-overflow-tooltip />
          <el-table-column prop="timeout_minutes" label="超时 (分)" width="80" align="center" />
          <el-table-column label="耗时" width="100" align="center">
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
                <span v-if="log.step_name" class="log-step">步骤：{{ log.step_name }}</span>
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
import { VideoPause, VideoPlay, VideoCamera, Back } from '@element-plus/icons-vue'
import type { DrillInstance, StepInstance } from '@/types'
import DrillStatusBadge from '@/components/common/DrillStatusBadge.vue'
import ActionConfirm from '@/components/common/ActionConfirm.vue'
import { drillApi } from '@/api/modules/drill'

const route = useRoute()
const router = useRouter()

const instance = ref<DrillInstance | null>(null)
const steps = ref<StepInstance[]>([])
const logs = ref<any[]>([])

const drillId = computed(() => {
  const id = route.params.id
  return typeof id === 'string' ? parseInt(id, 10) : 0
})

// API 返回的 steps 已经属于当前 drill
const drillSteps = computed(() => {
  return steps.value.sort((a, b) => a.seq - b.seq)
})

const completedSteps = computed(() => {
  return steps.value.filter(s => s.status === 'completed').length
})

const totalSteps = computed(() => {
  return steps.value.length || 1
})

const progressPercentage = computed(() => {
  return Math.round((completedSteps.value / totalSteps.value) * 100)
})

// 按钮互斥逻辑
const canPause = computed(() => instance.value?.status === 'running')
const canResume = computed(() => instance.value?.status === 'paused')
const canStart = computed(() => instance.value?.status === 'pending')
const canTerminate = computed(() => {
  const status = instance.value?.status
  return status === 'pending' || status === 'running' || status === 'paused'
})
const isFinished = computed(() => {
  const status = instance.value?.status
  return status === 'completed' || status === 'terminated'
})

// 使用 API 返回的真实日志
const drillLogs = computed(() => {
  return logs.value.map(log => ({
    ...log,
    step_name: null,
    operator: log.operator_name,
    remark: log.content,
  }))
})

function calculateDuration(step: StepInstance): string {
  if (!step.start_time || !step.end_time) {
    return '-'
  }
  const start = new Date(step.start_time).getTime()
  const end = new Date(step.end_time).getTime()
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
    start: 'primary',
    pause: 'warning',
    resume: 'primary',
    terminate: 'danger',
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
    start: '演练启动',
    pause: '演练暂停',
    resume: '演练继续',
    terminate: '演练终止',
    step_start: '步骤开始',
    step_complete: '步骤完成',
    step_issue: '步骤异常',
    step_skip: '步骤跳过',
    force_complete: '强制完成',
  }
  return map[action] || action
}

function getStepTypeLabel(stepType: string): string {
  const map: Record<string, string> = {
    serial: '串行',
    parallel: '并行',
    any_of: '任选',
    condition: '条件',
  }
  return map[stepType] || stepType
}

function getStepTypeTag(stepType: string): 'primary' | 'success' | 'warning' | 'info' {
  const map: Record<string, any> = {
    serial: 'primary',
    parallel: 'success',
    any_of: 'warning',
    condition: 'info',
  }
  return map[stepType] || 'info'
}

async function loadDrillData() {
  try {
    // 调用真实 API
    instance.value = await drillApi.getDetail(drillId.value)
    const stepsData = await drillApi.getSteps(drillId.value)
    steps.value = stepsData
    
    // 加载演练日志
    const logsData = await drillApi.getLogs(drillId.value)
    logs.value = logsData
    
    if (!instance.value) {
      ElMessage.error('演练不存在')
      router.back()
    }
  } catch (error) {
    ElMessage.error('加载数据失败')
    console.error('Failed to load drill data:', error)
  }
}

async function handlePause() {
  try {
    await drillApi.pause(drillId.value)
    ElMessage.success('演练已暂停')
    loadDrillData() // 刷新数据
  } catch (error) {
    ElMessage.error('暂停失败')
  }
}

async function handleResume() {
  try {
    await drillApi.resume(drillId.value)
    ElMessage.success('演练已继续')
    loadDrillData() // 刷新数据
  } catch (error) {
    ElMessage.error('继续失败')
  }
}

async function handleStart() {
  try {
    await drillApi.start(drillId.value)
    ElMessage.success('演练已启动')
    loadDrillData() // 刷新数据
  } catch (error) {
    ElMessage.error('启动失败')
  }
}

async function handleTerminate() {
  try {
    await drillApi.terminate(drillId.value)
    ElMessage.success('演练已终止')
    router.push('/director')
  } catch (error) {
    ElMessage.error('终止失败')
  }
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
