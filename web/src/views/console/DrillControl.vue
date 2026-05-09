<template>
  <div class="drill-control-page">
    <div class="page-header">
      <div class="header-left">
        <el-page-header @back="handleBack">
          <template #content>
            <div class="breadcrumb">
              <span class="drill-name">{{ drillInfo.name }}</span>
              <el-tag :type="getStatusType(drillInfo.status)" class="status-tag">
                {{ getStatusText(drillInfo.status) }}
              </el-tag>
            </div>
          </template>
        </el-page-header>
      </div>
      <div class="header-right">
        <CountdownTimer
          v-if="drillInfo.endTime"
          :endTime="drillInfo.endTime"
          :show-progress="true"
          :total-duration="drillInfo.totalDuration"
        />
      </div>
    </div>

    <div class="page-content">
      <!-- 左侧：流程图和步骤 -->
      <div class="main-section">
        <!-- 流程图 -->
        <el-card class="flow-card">
          <template #header>
            <div class="card-header">
              <span>演练流程</span>
              <div class="flow-stats">
                <span class="stat">总步骤：{{ totalSteps }}</span>
                <span class="stat completed">已完成：{{ completedSteps }}</span>
                <span class="stat running">执行中：{{ runningSteps }}</span>
              </div>
            </div>
          </template>
          <DrillFlowChart
            ref="flowChartRef"
            :nodes="flowNodes"
            :links="flowLinks"
            :width="800"
            :height="450"
          />
        </el-card>

        <!-- 步骤列表 -->
        <el-card class="steps-card">
          <template #header>
            <span>步骤详情</span>
          </template>
          <div class="step-list">
            <div
              v-for="(step, index) in steps"
              :key="step.id"
              class="step-item"
              :class="`step-${step.status}`"
              @click="handleStepClick(step)"
            >
              <div class="step-header">
                <span class="step-index">{{ index + 1 }}</span>
                <span class="step-name">{{ step.name }}</span>
                <el-tag :type="getStepStatusType(step.status)" size="small">
                  {{ getStepStatusText(step.status) }}
                </el-tag>
              </div>
              <div class="step-info">
                <span class="step-assignee" v-if="step.assignee">
                  <UserAvatar :name="step.assignee" size="xs" />
                  {{ step.assignee }}
                </span>
                <span class="step-timeout" v-if="step.timeoutMinutes">
                  时限：{{ step.timeoutMinutes }}分钟
                </span>
              </div>
              <div class="step-actions" v-if="canOperate(step)">
                <el-button
                  v-if="step.status === 'running'"
                  type="success"
                  size="small"
                  @click.stop="handleCompleteStep(step)"
                >
                  完成
                </el-button>
                <el-button
                  v-if="step.status === 'running'"
                  type="warning"
                  size="small"
                  @click.stop="handleIssueStep(step)"
                >
                  异常
                </el-button>
                <el-button
                  v-if="step.status === 'pending' && step.allowSkip"
                  type="info"
                  size="small"
                  @click.stop="handleSkipStep(step)"
                >
                  跳过
                </el-button>
              </div>
            </div>
          </div>
        </el-card>
      </div>

      <!-- 右侧：消息和控制台 -->
      <div class="side-section">
        <!-- 消息滚动 -->
        <el-card class="messages-card">
          <template #header>
            <span>实时消息</span>
          </template>
          <MessageTicker
            :messages="messages"
            title=""
            :max-messages="100"
            @clear="handleClearMessages"
          />
        </el-card>

        <!-- 干预控制台 -->
        <el-card class="control-card">
          <template #header>
            <span>指挥控制</span>
          </template>
          <div class="control-actions">
            <el-button
              v-if="drillInfo.status === 'running'"
              type="warning"
              class="control-btn"
              @click="handlePause"
            >
              暂停演练
            </el-button>
            <el-button
              v-if="drillInfo.status === 'paused'"
              type="success"
              class="control-btn"
              @click="handleResume"
            >
              恢复演练
            </el-button>
            <el-button
              type="danger"
              class="control-btn"
              :disabled="drillInfo.status === 'completed'"
              @click="handleTerminate"
            >
              终止演练
            </el-button>
          </div>

          <el-divider />

          <div class="control-form">
            <div class="form-title">人工干预</div>
            <el-form :model="interveneForm" label-position="top" size="default">
              <el-form-item label="选择步骤">
                <el-select
                  v-model="interveneForm.stepId"
                  placeholder="选择步骤"
                  style="width: 100%"
                >
                  <el-option
                    v-for="step in steps"
                    :key="step.id"
                    :label="step.name"
                    :value="step.id"
                    :disabled="step.status === 'completed'"
                  />
                </el-select>
              </el-form-item>
              <el-form-item label="操作类型">
                <el-select
                  v-model="interveneForm.action"
                  placeholder="选择操作"
                  style="width: 100%"
                >
                  <el-option label="强制完成" value="force_complete" />
                  <el-option label="标记异常" value="mark_issue" />
                  <el-option label="跳过步骤" value="skip" />
                  <el-option label="重新执行" value="retry" />
                </el-select>
              </el-form-item>
              <el-form-item label="备注">
                <el-input
                  v-model="interveneForm.remark"
                  type="textarea"
                  :rows="3"
                  placeholder="请输入干预原因"
                />
              </el-form-item>
              <el-button type="primary" @click="handleIntervene" style="width: 100%">
                执行干预
              </el-button>
            </el-form>
          </div>
        </el-card>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import DrillFlowChart from '@/components/DrillFlowChart.vue'
import CountdownTimer from '@/components/CountdownTimer.vue'
import MessageTicker from '@/components/MessageTicker.vue'
import UserAvatar from '@/components/UserAvatar.vue'

interface FlowNode {
  id: string | number
  name: string
  status: 'pending' | 'running' | 'completed' | 'issue' | 'skipped'
  x: number
  y: number
}

interface FlowLink {
  source: string | number
  target: string | number
}

interface Step {
  id: number
  name: string
  status: 'pending' | 'running' | 'completed' | 'issue' | 'skipped'
  assignee?: string
  timeoutMinutes?: number
  allowSkip?: boolean
}

interface MessageItem {
  id: string | number
  type: 'info' | 'success' | 'warning' | 'error'
  content: string
  timestamp: number
}

interface DrillInfo {
  id: number
  name: string
  status: 'pending' | 'running' | 'paused' | 'completed' | 'terminated'
  endTime?: number
  totalDuration?: number
}

const router = useRouter()
const route = useRoute()

const drillInfo = ref<DrillInfo>({
  id: 1,
  name: '数据库主从切换演练',
  status: 'running',
  endTime: Date.now() + 3600000,
  totalDuration: 7200
})

// 流程图数据
const flowNodes = ref<FlowNode[]>([
  { id: 1, name: '开始', status: 'completed', x: 100, y: 200 },
  { id: 2, name: '备份检查', status: 'completed', x: 250, y: 200 },
  { id: 3, name: '主库降级', status: 'running', x: 400, y: 100 },
  { id: 4, name: '从库升级', status: 'pending', x: 400, y: 300 },
  { id: 5, name: '流量切换', status: 'pending', x: 550, y: 200 },
  { id: 6, name: '验证测试', status: 'pending', x: 700, y: 200 },
  { id: 7, name: '结束', status: 'pending', x: 850, y: 200 }
])

const flowLinks = ref<FlowLink[]>([
  { source: 1, target: 2 },
  { source: 2, target: 3 },
  { source: 2, target: 4 },
  { source: 3, target: 5 },
  { source: 4, target: 5 },
  { source: 5, target: 6 },
  { source: 6, target: 7 }
])

// 步骤列表
const steps = ref<Step[]>([
  { id: 1, name: '开始', status: 'completed', assignee: '张三', allowSkip: false },
  { id: 2, name: '备份检查', status: 'completed', assignee: '李四', timeoutMinutes: 10, allowSkip: false },
  { id: 3, name: '主库降级', status: 'running', assignee: '王五', timeoutMinutes: 15, allowSkip: false },
  { id: 4, name: '从库升级', status: 'pending', assignee: '赵六', timeoutMinutes: 15, allowSkip: true },
  { id: 5, name: '流量切换', status: 'pending', assignee: '钱七', timeoutMinutes: 10, allowSkip: false },
  { id: 6, name: '验证测试', status: 'pending', assignee: '张三', timeoutMinutes: 20, allowSkip: false },
  { id: 7, name: '结束', status: 'pending', assignee: '张三', allowSkip: false }
])

// 消息列表
const messages = ref<MessageItem[]>([
  { id: 1, type: 'info', content: '演练已开始', timestamp: Date.now() - 300000 },
  { id: 2, type: 'success', content: '步骤【备份检查】已完成', timestamp: Date.now() - 240000 },
  { id: 3, type: 'info', content: '开始执行【主库降级】', timestamp: Date.now() - 180000 }
])

// 干预表单
const interveneForm = reactive({
  stepId: 0,
  action: '',
  remark: ''
})

// 统计数据
const totalSteps = computed(() => steps.value.length)
const completedSteps = computed(() => steps.value.filter(s => s.status === 'completed').length)
const runningSteps = computed(() => steps.value.filter(s => s.status === 'running').length)

// 获取状态类型
const getStatusType = (status: string) => {
  const typeMap: Record<string, any> = {
    pending: 'info',
    running: 'success',
    paused: 'warning',
    completed: 'success',
    terminated: 'danger'
  }
  return typeMap[status] || 'info'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const textMap: Record<string, string> = {
    pending: '待命',
    running: '执行中',
    paused: '已暂停',
    completed: '已完成',
    terminated: '已终止'
  }
  return textMap[status] || '未知'
}

// 获取步骤状态类型
const getStepStatusType = (status: string) => {
  const typeMap: Record<string, any> = {
    pending: 'info',
    running: 'warning',
    completed: 'success',
    issue: 'danger',
    skipped: ''
  }
  return typeMap[status] || 'info'
}

// 获取步骤状态文本
const getStepStatusText = (status: string) => {
  const textMap: Record<string, string> = {
    pending: '待命',
    running: '执行中',
    completed: '已完成',
    issue: '异常',
    skipped: '已跳过'
  }
  return textMap[status] || '未知'
}

// 是否可以操作
const canOperate = (step: Step) => {
  return ['running', 'pending'].includes(step.status)
}

// 点击步骤
const handleStepClick = (step: Step) => {
  console.log('点击步骤:', step)
}

// 完成步骤
const handleCompleteStep = async (step: Step) => {
  try {
    await ElMessageBox.confirm(`确认完成步骤"${step.name}"？`, '提示', {
      type: 'info'
    })
    step.status = 'completed'
    ElMessage.success('步骤已完成')
    messages.value.push({
      id: Date.now(),
      type: 'success',
      content: `步骤【${step.name}】已完成`,
      timestamp: Date.now()
    })
  } catch {
    // 取消
  }
}

// 标记异常
const handleIssueStep = async (step: Step) => {
  try {
    await ElMessageBox.confirm(`标记步骤"${step.name}"为异常？`, '警告', {
      type: 'warning'
    })
    step.status = 'issue'
    ElMessage.warning('步骤已标记为异常')
    messages.value.push({
      id: Date.now(),
      type: 'error',
      content: `步骤【${step.name}】报告异常`,
      timestamp: Date.now()
    })
  } catch {
    // 取消
  }
}

// 跳过步骤
const handleSkipStep = async (step: Step) => {
  try {
    await ElMessageBox.confirm(`确认跳过步骤"${step.name}"？`, '提示', {
      type: 'info'
    })
    step.status = 'skipped'
    ElMessage.info('步骤已跳过')
    messages.value.push({
      id: Date.now(),
      type: 'info',
      content: `步骤【${step.name}】已跳过`,
      timestamp: Date.now()
    })
  } catch {
    // 取消
  }
}

// 暂停演练
const handlePause = async () => {
  try {
    await ElMessageBox.confirm('确认暂停演练？', '提示', {
      type: 'warning'
    })
    drillInfo.value.status = 'paused'
    ElMessage.warning('演练已暂停')
  } catch {
    // 取消
  }
}

// 恢复演练
const handleResume = async () => {
  try {
    await ElMessageBox.confirm('确认恢复演练？', '提示', {
      type: 'info'
    })
    drillInfo.value.status = 'running'
    ElMessage.success('演练已恢复')
  } catch {
    // 取消
  }
}

// 终止演练
const handleTerminate = async () => {
  try {
    await ElMessageBox.confirm('确认终止演练？此操作不可恢复', '警告', {
      type: 'error',
      confirmButtonText: '确认终止',
      confirmButtonClass: 'el-button--danger'
    })
    drillInfo.value.status = 'terminated'
    ElMessage.error('演练已终止')
  } catch {
    // 取消
  }
}

// 执行干预
const handleIntervene = async () => {
  if (!interveneForm.stepId || !interveneForm.action) {
    ElMessage.warning('请选择步骤和操作类型')
    return
  }

  try {
    await ElMessageBox.confirm('确认执行干预操作？', '警告', {
      type: 'warning'
    })

    const step = steps.value.find(s => s.id === interveneForm.stepId)
    if (step) {
      if (interveneForm.action === 'force_complete') {
        step.status = 'completed'
      } else if (interveneForm.action === 'mark_issue') {
        step.status = 'issue'
      } else if (interveneForm.action === 'skip') {
        step.status = 'skipped'
      }
    }

    ElMessage.success('干预成功')
    messages.value.push({
      id: Date.now(),
      type: 'warning',
      content: `指挥员干预：${interveneForm.action} 步骤 ${step?.name}`,
      timestamp: Date.now()
    })

    interveneForm.stepId = 0
    interveneForm.action = ''
    interveneForm.remark = ''
  } catch {
    // 取消
  }
}

// 清空消息
const handleClearMessages = () => {
  messages.value = []
}

// 返回
const handleBack = () => {
  router.back()
}

onMounted(() => {
  // 这里应该从 API 获取演练详情
  console.log('演练 ID:', route.params.id)
})
</script>

<style scoped>
.drill-control-page {
  padding: 24px;
  background-color: var(--color-background, #020617);
  min-height: 100vh;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 12px;
}

.drill-name {
  font-size: 16px;
  font-weight: 600;
  color: var(--color-foreground, #F8FAFC);
}

.page-content {
  display: grid;
  grid-template-columns: 1fr 400px;
  gap: 20px;
}

.main-section,
.side-section {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.flow-stats {
  display: flex;
  gap: 16px;
  font-size: 13px;
}

.stat {
  color: var(--color-muted-foreground, #94A3B8);
}

.stat.completed {
  color: #22C55E;
}

.stat.running {
  color: #3B82F6;
}

.step-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.step-item {
  padding: 16px;
  border-radius: 8px;
  border: 1px solid var(--color-border, #334155);
  background-color: var(--color-secondary, #1E293B);
  cursor: pointer;
  transition: all 0.2s ease;
}

.step-item:hover {
  border-color: var(--color-primary, #0F172A);
}

.step-item.step-completed {
  border-left: 3px solid #22C55E;
}

.step-item.step-running {
  border-left: 3px solid #3B82F6;
}

.step-item.step-issue {
  border-left: 3px solid #EF4444;
}

.step-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.step-index {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background-color: var(--color-muted, #1A1E2F);
  color: var(--color-foreground, #F8FAFC);
  font-size: 13px;
  font-weight: 600;
  flex-shrink: 0;
}

.step-name {
  flex: 1;
  font-weight: 500;
  color: var(--color-foreground, #F8FAFC);
}

.step-info {
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: var(--color-muted-foreground, #94A3B8);
  margin-bottom: 8px;
}

.step-assignee {
  display: flex;
  align-items: center;
  gap: 6px;
}

.step-actions {
  display: flex;
  gap: 8px;
}

.messages-card {
  min-height: 300px;
}

.messages-card :deep(.message-ticker) {
  height: 300px;
}

.control-actions {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.control-btn {
  width: 100%;
}

.control-form {
  margin-top: 16px;
}

.form-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 12px;
  color: var(--color-foreground, #F8FAFC);
}

:deep(.el-card) {
  background-color: var(--color-muted, #1A1E2F);
  border: 1px solid var(--color-border, #334155);
}

:deep(.el-card__header) {
  background-color: var(--color-secondary, #1E293B);
  border-bottom: 1px solid var(--color-border, #334155);
}

:deep(.el-card__header span) {
  color: var(--color-foreground, #F8FAFC);
  font-weight: 600;
}

@media (max-width: 1200px) {
  .page-content {
    grid-template-columns: 1fr;
  }
}
</style>
