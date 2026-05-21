<template>
  <div class="page-container">
    <div v-if="instance" class="page-main">
      <!-- 顶部：演练信息 + 控制 -->
      <el-card class="header-card">
        <div class="header-content">
          <div class="header-left">
            <h2 class="drill-name">{{ instance.name }}</h2>
            <DrillStatusBadge :status="instance.status" type="drill" class="status-badge" />
          </div>
          <div class="header-right">
            <div class="progress-wrap">
              <span class="progress-label">进度 {{ progressPercentage }}%</span>
              <el-progress
                :percentage="progressPercentage"
                :stroke-width="8"
                :status="instance.status === 'completed' ? 'success' : undefined"
              />
            </div>
            <div class="control-buttons">
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
              <el-button v-if="isFinished" @click="router.back()">
                <el-icon><Back /></el-icon>
                返回
              </el-button>
            </div>
          </div>
        </div>
      </el-card>

      <!-- 异常/超时步骤 -->
      <el-card v-if="issueSteps.length > 0" class="issue-card" shadow="never">
        <template #header>
          <div class="issue-header">
            <el-icon class="issue-icon"><Warning /></el-icon>
            <span class="issue-title">异常/超时 ({{ issueSteps.length }})</span>
          </div>
        </template>
        <div class="issue-list">
        <el-alert
          v-for="step in issueSteps"
          :key="step.id"
          :title="step.name"
          :description="getIssueDescription(step)"
          :type="step.status === 'timeout' ? 'warning' : 'error'"
          :closable="false"
          show-icon
          class="issue-item"
        >
          <template #default>
            <div class="issue-actions">
              <ActionConfirm
                :title="`重新派发：${step.name}`"
                message="将步骤状态重置为运行中并重新计时超时，是否继续？"
                type="warning"
                @confirm="handleResumeTask(step)"
                :disabled="instance?.status === 'paused'"
              >
                <el-icon><RefreshRight /></el-icon>
                重新派发
              </ActionConfirm>
              <ActionConfirm
                :title="`跳过步骤：${step.name}`"
                message="跳过此步骤后，该步骤将标记为已跳过，是否继续？"
                type="warning"
                @confirm="handleSkipStep(step)"
                :disabled="instance?.status === 'paused'"
              >
                <el-icon><DArrowRight /></el-icon>
                跳过
              </ActionConfirm>
              <ActionConfirm
                :title="`强制完成：${step.name}`"
                message="强制将此步骤标记为完成，是否继续？"
                @confirm="handleForceComplete(step)"
                :disabled="instance?.status === 'paused'"
              >
                <el-icon><CircleCheck /></el-icon>
                强制完成
              </ActionConfirm>
            </div>
          </template>
        </el-alert>
        </div>
      </el-card>

      <!-- 当前步骤 (running) -->
      <el-card v-if="runningSteps.length > 0" class="running-card">
        <template #header>
          <div class="running-header">
            <el-icon class="running-icon"><VideoCamera /></el-icon>
            <span class="running-title">当前步骤 ({{ runningSteps.length }})</span>
          </div>
        </template>
        <div class="running-grid">
          <el-card
            v-for="step in runningSteps"
            :key="step.id"
            class="step-card"
            shadow="hover"
          >
            <div class="step-info">
              <div class="step-name">
                <span class="step-seq">#{{ step.seq }}</span>
                {{ step.name }}
              </div>
              <el-descriptions :column="1" size="small" class="step-meta">
                <el-descriptions-item label="执行角色">
                  <el-tag size="small" :type="step.default_assignee_role === 'director' ? 'warning' : 'primary'">
                    {{ step.default_assignee_role === 'director' ? '指挥组' : '执行组' }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item v-if="step.assignee_names" label="执行人">
                  {{ step.assignee_names }}
                </el-descriptions-item>
                <el-descriptions-item v-if="step.timeout_minutes" label="超时">
                  {{ step.timeout_minutes }} 分钟
                  <span v-if="step.timeout_at" class="timeout-countdown">({{ getCountdown(step.timeout_at) }})</span>
                </el-descriptions-item>
                <el-descriptions-item v-if="step.executor_team" label="团队">
                  {{ step.executor_team }}
                </el-descriptions-item>
              </el-descriptions>
            </div>
            <div class="step-actions">
              <ActionConfirm
                v-if="step.default_assignee_role === 'director'"
                :title="`完成任务：${step.name}`"
                message="确认已完成此步骤？"
                type="success"
                @confirm="handleDirectorComplete(step)"
                :disabled="instance?.status === 'paused'"
              >
                <el-icon><CircleCheckFilled /></el-icon>
                完成任务
              </ActionConfirm>
              <ActionConfirm
                :title="`跳过步骤：${step.name}`"
                message="跳过此步骤后，该步骤将标记为已跳过，是否继续？"
                type="warning"
                @confirm="handleSkipStep(step)"
                :disabled="instance?.status === 'paused'"
              >
                <el-icon><DArrowRight /></el-icon>
                跳过
              </ActionConfirm>
              <ActionConfirm
                :title="`强制完成：${step.name}`"
                message="强制将此步骤标记为完成，是否继续？"
                @confirm="handleForceComplete(step)"
                :disabled="instance?.status === 'paused'"
              >
                <el-icon><CircleCheck /></el-icon>
                强制完成
              </ActionConfirm>
            </div>
          </el-card>
        </div>
      </el-card>

      <!-- 步骤列表 -->
      <el-card class="steps-card">
        <template #header>
          <span class="card-title">步骤列表</span>
        </template>
        <el-table :data="drillStepTree" row-key="id" :tree-props="{ children: 'children' }" style="width: 100%">
          <el-table-column prop="seq" label="序号" width="60" align="center" />
          <el-table-column prop="name" label="步骤名" min-width="180" show-overflow-tooltip />
          <el-table-column prop="step_type" label="类型" width="80" align="center">
            <template #default="{ row }">
              <el-tag :type="getStepTypeTag(row.step_type)" size="small">{{ getStepTypeLabel(row.step_type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="160" align="center">
            <template #default="{ row }">
              <template v-if="row.children && row.children.length > 0">
                {{ getStepStatusText(row).text }}
              </template>
              <template v-else>
                <DrillStatusBadge :status="row.status" type="step" />
              </template>
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
import { VideoPause, VideoPlay, VideoCamera, Back, Warning, DArrowRight, CircleCheck, RefreshRight, CircleCheckFilled } from '@element-plus/icons-vue'
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

const drillSteps = computed(() => {
  return steps.value.sort((a, b) => a.seq - b.seq)
})

// 将扁平步骤数据转换为树形结构
const drillStepTree = computed(() => {
  const sorted = steps.value.sort((a, b) => a.seq - b.seq)
  const stepMap = new Map<number, StepInstance & { children?: StepInstance[] }>()

  // 第一步：初始化所有步骤到 map
  for (const step of sorted) {
    stepMap.set(step.id, { ...step })
  }

  const roots: (StepInstance & { children?: StepInstance[] })[] = []

  // 第二步：建立父子关系
  for (const step of sorted) {
    const node = stepMap.get(step.id)!
    if (step.parent_step_id && stepMap.has(step.parent_step_id)) {
      const parent = stepMap.get(step.parent_step_id)!
      if (!parent.children) {
        parent.children = []
      }
      parent.children.push(node)
    } else {
      roots.push(node)
    }
  }

  return roots
})

// 父步骤状态聚合显示
function getStepStatusText(row: StepInstance & { children?: StepInstance[] }): { text: string; isParent: boolean } {
  if (!row.children || row.children.length === 0) {
    return { text: row.status, isParent: false }
  }
  const total = row.children.length
  const completed = row.children.filter(c => c.status === 'completed' || c.status === 'skipped').length
  return { text: `${completed}/${total} 子任务已完成`, isParent: true }
}

const completedSteps = computed(() => {
  return steps.value.filter(s => s.status === 'completed' || s.status === 'skipped').length
})

const totalSteps = computed(() => {
  return steps.value.length || 1
})

const progressPercentage = computed(() => {
  return Math.round((completedSteps.value / totalSteps.value) * 100)
})

const runningSteps = computed(() => {
  return steps.value
    .filter(s => s.status === 'running')
    .sort((a, b) => a.seq - b.seq)
})

const issueSteps = computed(() => {
  return steps.value
    .filter(s => s.status === 'timeout' || s.status === 'issue')
    .sort((a, b) => a.seq - b.seq)
})

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
    step_timeout: 'danger',
    force_complete: 'warning',
    resume_task: 'primary',
    timeout: 'danger',
    skip: 'info',
  }
  return map[action] || 'info'
}

function getLogActionLabel(action: string): string {
  const map: Record<string, string> = {
    start: '演练启动',
    pause: '演练暂停',
    resume: '演练继续',
    complete: '演练完成',
    terminate: '演练终止',
    step_start: '步骤开始',
    step_complete: '步骤完成',
    step_issue: '步骤异常',
    step_skip: '步骤跳过',
    step_timeout: '步骤超时',
    force_complete: '强制完成',
    resume_task: '重新派发',
    timeout: '系统超时',
    skip: '手动跳过',
    director_complete: '指挥组完成任务',
    step_complete_director: '指挥组完成任务',
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

function getIssueDescription(step: StepInstance): string {
  if (step.status === 'timeout') {
    return `步骤已超时（限制 ${step.timeout_minutes || '?'} 分钟）`
  }
  if (step.status === 'issue') {
    return '步骤执行异常，需要处理'
  }
  return ''
}

function getCountdown(timeoutAt: string): string {
  if (!timeoutAt) return ''
  const now = Date.now()
  const target = new Date(timeoutAt).getTime()
  const diff = Math.max(0, Math.floor((target - now) / 1000))
  if (diff <= 0) return '已超时'
  const mins = Math.floor(diff / 60)
  const secs = diff % 60
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

async function loadDrillData() {
  try {
    instance.value = await drillApi.getDetail(drillId.value)
    const stepsData = await drillApi.getSteps(drillId.value)
    steps.value = stepsData

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
    loadDrillData()
  } catch (error) {
    ElMessage.error('暂停失败')
  }
}

async function handleResume() {
  try {
    await drillApi.resume(drillId.value)
    ElMessage.success('演练已继续')
    loadDrillData()
  } catch (error) {
    ElMessage.error('继续失败')
  }
}

async function handleStart() {
  try {
    await drillApi.start(drillId.value)
    ElMessage.success('演练已启动')
    loadDrillData()
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

async function handleSkipStep(step: StepInstance) {
  try {
    await drillApi.skipStep(drillId.value, step.step_template_id, 'director skipped')
    ElMessage.success('步骤已跳过')
    loadDrillData()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '操作失败')
  }
}

async function handleResumeTask(step: StepInstance) {
  try {
    await drillApi.resumeTask(drillId.value, step.step_template_id)
    ElMessage.success('任务已重新派发')
    loadDrillData()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '操作失败')
  }
}

async function handleDirectorComplete(step: StepInstance) {
  try {
    await drillApi.completeStep(drillId.value, step.step_template_id, '指挥组完成任务')
    ElMessage.success('步骤已完成')
    loadDrillData()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '操作失败')
  }
}

async function handleForceComplete(step: StepInstance) {
  try {
    await drillApi.forceCompleteStep(drillId.value, step.step_template_id, `指挥组强制完成步骤：${step.name}`)
    ElMessage.success(`步骤「${step.name}」已强制完成`)
    loadDrillData()
  } catch (error) {
    ElMessage.error('强制完成失败')
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

  .page-main {
    @include page-content;
  }

  .header-card {
    @include card-compact;
    margin-bottom: $spacing-base;

    :deep(.el-card__body) {
      padding: $spacing-base;
    }

    .header-content {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      gap: $spacing-base;
      flex-wrap: wrap;

      .header-left {
        display: flex;
        align-items: center;
        gap: $spacing-sm;

        .drill-name {
          font-size: $font-size-xl;
          font-weight: $font-weight-bold;
          color: $text-primary;
          margin: 0;
        }
      }

      .header-right {
        display: flex;
        flex-direction: column;
        align-items: flex-end;
        gap: $spacing-sm;

        .progress-wrap {
          min-width: 240px;

          .progress-label {
            display: block;
            font-size: $font-size-xs;
            color: $text-secondary;
            margin-bottom: 4px;
          }
        }

        .control-buttons {
          display: flex;
          gap: $spacing-xs;
          flex-wrap: wrap;

          :deep(.el-button) {
            display: flex;
            align-items: center;
            gap: 6px;
          }
        }
      }
    }
  }

  .issue-card {
    margin-bottom: $spacing-base;
    border-color: rgba(245, 108, 108, 0.3);

    :deep(.el-card__body) {
      padding: $spacing-sm;
    }

    .issue-header {
      display: flex;
      align-items: center;
      gap: $spacing-xs;

      .issue-icon {
        color: #f56c6c;
        font-size: $font-size-base;
      }

      .issue-title {
        font-size: $font-size-base;
        font-weight: $font-weight-semibold;
        color: #f56c6c;
      }
    }

  .issue-list {
    display: flex;
    flex-direction: column;
    gap: $spacing-sm;

    .issue-item {
      :deep(.el-alert__content) {
        padding-right: $spacing-sm;
      }

      .issue-actions {
        display: flex;
        gap: $spacing-sm;
        margin-top: $spacing-sm;

        :deep(.el-button) {
          display: flex;
          align-items: center;
          gap: 4px;
        }
      }
    }
  }
    }
  }

  .running-card {
    @include card-compact;
    margin-bottom: $spacing-base;

    .running-header {
      display: flex;
      align-items: center;
      gap: $spacing-xs;

      .running-icon {
        color: #67c23a;
        font-size: $font-size-base;
      }

      .running-title {
        font-size: $font-size-base;
        font-weight: $font-weight-semibold;
        color: $text-primary;
      }
    }

    .running-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
      gap: $spacing-base;
    }

    .step-card {
      :deep(.el-card__body) {
        padding: $spacing-sm;
      }

      .step-info {
        margin-bottom: $spacing-sm;

        .step-name {
          font-size: $font-size-base;
          font-weight: $font-weight-semibold;
          color: $text-primary;
          margin-bottom: $spacing-xs;

          .step-seq {
            color: $color-primary;
            margin-right: $spacing-xs;
          }
        }

        .step-meta {
          :deep(.el-descriptions__label) {
            font-size: $font-size-xs;
            color: $text-tertiary;
          }

          :deep(.el-descriptions__content) {
            font-size: $font-size-xs;
            color: $text-secondary;
          }

          .timeout-countdown {
            color: #e6a23c;
            font-weight: $font-weight-semibold;
          }
        }
      }

      .step-actions {
        display: flex;
        gap: $spacing-xs;

        :deep(.el-button) {
          display: flex;
          align-items: center;
          gap: 4px;
          flex: 1;
        }
      }
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
</style>
