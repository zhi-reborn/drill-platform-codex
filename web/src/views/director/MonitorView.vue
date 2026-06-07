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
                v-if="step.default_assignee_role === 'director' && !isParentStep(step)"
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
                v-if="!isParentStep(step)"
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
                v-if="!isParentStep(step)"
                :title="`强制完成：${step.name}`"
                message="强制将此步骤标记为完成，是否继续？"
                @confirm="handleForceComplete(step)"
                :disabled="instance?.status === 'paused'"
              >
                <el-icon><CircleCheck /></el-icon>
                强制完成
              </ActionConfirm>
              <span v-if="isParentStep(step)" class="parent-hint">
                父任务 · 子任务全部完成后自动完成
              </span>
            </div>
          </el-card>
        </div>
      </el-card>

      <!-- 步骤列表 (按 phase 分组) -->
      <el-card class="steps-card">
        <template #header>
          <span class="card-title">步骤列表</span>
        </template>
        <el-tabs v-model="activePhase" type="card" class="phase-tabs">
          <el-tab-pane
            v-for="(group, phase) in phaseGroups"
            :key="phase"
            :name="phase"
          >
            <template #label>
              <span class="tab-label">
                {{ phase }}
                <el-tag size="small" type="info">{{ group.length }} 步</el-tag>
              </span>
            </template>
            <el-table :data="group" row-key="id" :tree-props="{ children: 'children' }" style="width: 100%" size="small">
              <el-table-column prop="seq" label="序号" width="55" align="center" />
              <el-table-column prop="name" label="步骤名" min-width="150" show-overflow-tooltip />
              <el-table-column prop="step_type" label="类型" width="70" align="center">
                <template #default="{ row }">
                  <el-tag :type="getStepTypeTag(row.step_type)" size="small">{{ getStepTypeLabel(row.step_type) }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="status" label="状态" width="140" align="center">
                <template #default="{ row }">
                  <template v-if="row.children && row.children.length > 0">
                    {{ getStepStatusText(row).text }}
                  </template>
                  <template v-else>
                    <DrillStatusBadge :status="row.status" type="step" />
                  </template>
                </template>
              </el-table-column>
              <el-table-column prop="estimated_duration_minutes" label="预计耗时" width="80" align="center">
                <template #default="{ row }">
                  {{ row.estimated_duration_minutes ? `${row.estimated_duration_minutes}m` : '-' }}
                </template>
              </el-table-column>
              <el-table-column prop="default_assignee_role" label="角色" width="75" align="center">
                <template #default="{ row }">
                  <el-tag :type="row.default_assignee_role === 'director' ? 'warning' : 'primary'" size="small">
                    {{ row.default_assignee_role === 'director' ? '指挥组' : '执行组' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="attributes" label="操作人" width="90" align="center" show-overflow-tooltip>
                <template #default="{ row }">
                  {{ row.attributes?.operator || '-' }}
                </template>
              </el-table-column>
              <el-table-column prop="attributes" label="复核人" width="90" align="center" show-overflow-tooltip>
                <template #default="{ row }">
                  {{ row.attributes?.reviewer || '-' }}
                </template>
              </el-table-column>
              <el-table-column label="实际耗时" width="85" align="center">
                <template #default="{ row }">
                  {{ calculateDuration(row) }}
                </template>
              </el-table-column>
              <el-table-column label="操作" width="220" align="center" fixed="right">
                <template #default="{ row }">
                  <el-button type="primary" link size="small" @click="showStepDetail(row)">详情</el-button>
                  <template v-if="!isParentStep(row)">
                    <el-popconfirm
                      v-if="row.status === 'pending' && row.default_assignee_role === 'director'"
                      title="确认开始此步骤？"
                      @confirm="handleStartStep(row)"
                    >
                      <template #reference>
                        <el-button type="primary" link size="small">开始</el-button>
                      </template>
                    </el-popconfirm>
                    <el-button
                      v-if="['pending', 'running'].includes(row.status)"
                      type="success"
                      link
                      size="small"
                      @click="handleDirectorComplete(row)"
                    >完成</el-button>
                    <el-popconfirm
                      v-if="['pending', 'running', 'timeout'].includes(row.status)"
                      title="确认跳过此步骤？"
                      @confirm="handleSkipStep(row)"
                    >
                      <template #reference>
                        <el-button type="warning" link size="small">跳过</el-button>
                      </template>
                    </el-popconfirm>
                    <el-popconfirm
                      v-if="['pending', 'running', 'timeout'].includes(row.status)"
                      title="确认强制完成此步骤？"
                      @confirm="handleForceComplete(row)"
                    >
                      <template #reference>
                        <el-button type="danger" link size="small">强制完成</el-button>
                      </template>
                    </el-popconfirm>
                    <el-popconfirm
                      v-if="['timeout', 'completed', 'skipped'].includes(row.status)"
                      title="确认重新派发此步骤？"
                      @confirm="handleResumeTask(row)"
                    >
                      <template #reference>
                        <el-button type="primary" link size="small">重新派发</el-button>
                      </template>
                    </el-popconfirm>
                  </template>
                </template>
              </el-table-column>
            </el-table>
          </el-tab-pane>
        </el-tabs>
      </el-card>

      <!-- 步骤详情弹窗 -->
      <el-dialog v-model="detailVisible" :title="`步骤详情：${selectedStep?.name || ''}`" width="720px" destroy-on-close>
        <template v-if="selectedStep">
          <div class="detail-toolbar">
            <el-button v-if="!detailEditing" type="primary" link @click="enterEditMode">
              <el-icon><Edit /></el-icon>编辑
            </el-button>
            <template v-else>
              <el-button @click="cancelEditMode">取消</el-button>
              <el-button type="primary" :loading="detailSaving" @click="saveStepInfo">保存</el-button>
            </template>
          </div>
          <el-form v-if="!detailEditing" :column="2" border size="small" label-position="left" label-width="100px">
            <el-row :gutter="0">
              <el-col :span="12">
                <el-form-item label="序号">{{ selectedStep.seq }}</el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="状态">
                  <DrillStatusBadge :status="selectedStep.status" type="step" />
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="步骤类型">
                  <el-tag :type="getStepTypeTag(selectedStep.step_type)" size="small">{{ getStepTypeLabel(selectedStep.step_type) }}</el-tag>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="执行角色">
                  <el-tag :type="selectedStep.default_assignee_role === 'director' ? 'warning' : 'primary'" size="small">
                    {{ selectedStep.default_assignee_role === 'director' ? '指挥组' : '执行组' }}
                  </el-tag>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="阶段">{{ selectedStep.phase || '未分类' }}</el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="阶段内步骤">{{ selectedStep.phase_step || '-' }}</el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="预计耗时">{{ selectedStep.estimated_duration_minutes ? `${selectedStep.estimated_duration_minutes} 分钟` : '-' }}</el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="预计启动偏移">{{ selectedStep.estimated_start_offset ? `${selectedStep.estimated_start_offset} 秒` : '-' }}</el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="超时时间(分钟)">{{ selectedStep.timeout_minutes || '-' }}</el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="执行团队">{{ selectedStep.executor_team || '-' }}</el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="责任部门">{{ selectedStep.attributes?.responsible_department || '-' }}</el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="配合部门">{{ selectedStep.attributes?.cooperating_department || '-' }}</el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="操作人">{{ selectedStep.attributes?.operator || '-' }}</el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="复核人">{{ selectedStep.attributes?.reviewer || '-' }}</el-form-item>
              </el-col>
              <el-col :span="24">
                <el-form-item label="分配人员">{{ selectedStep.assignee_names || '-' }}</el-form-item>
              </el-col>
              <el-col :span="24">
                <el-form-item label="操作说明">{{ selectedStep.attributes?.operation_guide || '-' }}</el-form-item>
              </el-col>
              <el-col :span="24">
                <el-form-item label="验证方式">{{ selectedStep.attributes?.verification_method || '-' }}</el-form-item>
              </el-col>
              <el-col :span="24">
                <el-form-item label="兜底措施">{{ selectedStep.attributes?.fallback_measures || '-' }}</el-form-item>
              </el-col>
              <el-col :span="8">
                <el-form-item label="实际耗时">{{ calculateDuration(selectedStep) }}</el-form-item>
              </el-col>
              <el-col :span="8">
                <el-form-item label="开始时间">{{ selectedStep.start_time ? formatTime(selectedStep.start_time) : '-' }}</el-form-item>
              </el-col>
              <el-col :span="8">
                <el-form-item label="结束时间">{{ selectedStep.end_time ? formatTime(selectedStep.end_time) : '-' }}</el-form-item>
              </el-col>
            </el-row>
          </el-form>
          <el-form v-else ref="editFormRef" :model="editForm" label-position="left" label-width="100px" size="default">
            <el-form-item label="责任部门">
              <el-input v-model="editForm.attributes.responsible_department" placeholder="责任部门" clearable />
            </el-form-item>
            <el-form-item label="配合部门">
              <el-input v-model="editForm.attributes.cooperating_department" placeholder="配合部门" clearable />
            </el-form-item>
            <el-form-item label="操作人">
              <el-autocomplete
                :model-value="editForm.attributes.operator"
                @update:model-value="(v: string) => (editForm.attributes.operator = v ?? '')"
                :fetch-suggestions="userQuerySearch"
                placeholder="输入或选择操作人"
                clearable
                style="width: 100%"
                @focus="ensureUserOptions"
              />
            </el-form-item>
            <el-form-item label="复核人">
              <el-autocomplete
                :model-value="editForm.attributes.reviewer"
                @update:model-value="(v: string) => (editForm.attributes.reviewer = v ?? '')"
                :fetch-suggestions="userQuerySearch"
                placeholder="输入或选择复核人"
                clearable
                style="width: 100%"
                @focus="ensureUserOptions"
              />
            </el-form-item>
            <el-form-item label="操作说明">
              <el-input v-model="editForm.attributes.operation_guide" type="textarea" :rows="3" placeholder="操作说明" />
            </el-form-item>
            <el-form-item label="验证方式">
              <el-input v-model="editForm.attributes.verification_method" type="textarea" :rows="2" placeholder="验证方式" />
            </el-form-item>
            <el-form-item label="兜底措施">
              <el-input v-model="editForm.attributes.fallback_measures" type="textarea" :rows="2" placeholder="兜底措施" />
            </el-form-item>
            <el-form-item label="备注">
              <el-input v-model="editForm.remark" type="textarea" :rows="2" placeholder="本次编辑备注" />
            </el-form-item>
          </el-form>
        </template>
      </el-dialog>

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
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Edit } from '@element-plus/icons-vue'
import { VideoPause, VideoPlay, VideoCamera, Back, Warning, DArrowRight, CircleCheck, RefreshRight, CircleCheckFilled } from '@element-plus/icons-vue'
import type { DrillInstance, StepInstance } from '@/types'
import DrillStatusBadge from '@/components/common/DrillStatusBadge.vue'
import ActionConfirm from '@/components/common/ActionConfirm.vue'
import { drillApi } from '@/api/modules/drill'
import { userApi } from '@/api/modules/user'
import { useAuthStore } from '@/stores/auth'

const activePhase = ref<string>('')
const selectedStep = ref<StepInstance | null>(null)
const detailVisible = ref(false)
const detailEditing = ref(false)
const detailSaving = ref(false)
const editFormRef = ref()

const editForm = reactive({
  timeout_minutes: 0,
  executor_team: '',
  assignee_names: '',
  remark: '',
  attributes: {
    operation_guide: '',
    verification_method: '',
    fallback_measures: '',
    responsible_department: '',
    cooperating_department: '',
    operator: '',
    reviewer: '',
  },
})

// 用户列表(用于操作人/复核人模糊匹配)
const userOptions = ref<{ value: string; label: string }[]>([])
let userOptionsLoaded = false

async function ensureUserOptions() {
  if (userOptionsLoaded) return
  try {
    const res = await userApi.getList({ page: 1, page_size: 200 })
    const items = (res as any)?.items || (res as any)?.data?.items || []
    userOptions.value = items.map((u: any) => {
      const name = u.real_name || u.RealName || u.username || u.Username || ''
      return { value: name, label: name }
    }).filter((o: any) => o.value)
    userOptionsLoaded = true
  } catch (e) {
    // 静默失败,降级为普通输入框
    userOptions.value = []
  }
}

function userQuerySearch(queryString: string, cb: any) {
  if (!queryString) {
    cb(userOptions.value)
    return
  }
  const q = queryString.toLowerCase()
  const results = userOptions.value.filter(u => u.value.toLowerCase().includes(q))
  cb(results)
}

const route = useRoute()
const router = useRouter()

const instance = ref<DrillInstance | null>(null)
const steps = ref<StepInstance[]>([])
const logs = ref<any[]>([])

// WebSocket
let ws: WebSocket | null = null
let refreshTimer: number | null = null

const drillId = computed(() => {
  const id = route.params.id
  const parsed = typeof id === 'string' ? parseInt(id, 10) : 0
  return Number.isFinite(parsed) && parsed > 0 ? parsed : 0
})

const isValidDrill = computed(() => drillId.value > 0)

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

// 按 phase 分组（支持树形结构）
const phaseGroups = computed(() => {
  const groups: Record<string, (StepInstance & { children?: StepInstance[] })[]> = {}
  const sorted = steps.value.sort((a, b) => a.seq - b.seq)
  const stepMap = new Map<number, StepInstance & { children?: StepInstance[] }>()

  // 第一步：初始化所有步骤到 map
  for (const step of sorted) {
    stepMap.set(step.id, { ...step })
  }

  // 第二步：建立父子关系
  for (const step of sorted) {
    const node = stepMap.get(step.id)!
    if (step.parent_step_id && stepMap.has(step.parent_step_id)) {
      const parent = stepMap.get(step.parent_step_id)!
      if (!parent.children) {
        parent.children = []
      }
      parent.children.push(node)
    }
  }

  // 第三步：按 phase 分组，仅处理根步骤（tree roots）
  for (const step of sorted) {
    const node = stepMap.get(step.id)!
    // 只处理根步骤（没有 parent_step_id 或 parent 不存在）
    if (step.parent_step_id && stepMap.has(step.parent_step_id)) {
      continue
    }
    const phase = node.phase || '未分类'
    if (!groups[phase]) {
      groups[phase] = []
    }
    groups[phase].push(node)
  }

  return groups
})

// 初始化 activePhase，默认选中第一个 phase
function initActivePhase() {
  const phases = Object.keys(phaseGroups.value)
  if (phases.length > 0) {
    activePhase.value = phases[0]
  }
}

// 父步骤状态聚合显示
function getStepStatusText(row: StepInstance & { children?: StepInstance[] }): { text: string; isParent: boolean } {
  if (!row.children || row.children.length === 0) {
    return { text: row.status, isParent: false }
  }
  const total = row.children.length
  const completed = row.children.filter(c => c.status === 'completed' || c.status === 'skipped').length
  return { text: `${completed}/${total} 子任务已完成`, isParent: true }
}

// 检查是否为父步骤（有其他步骤的 parent_step_id 指向它）
function isParentStep(step: StepInstance): boolean {
  return steps.value.some(s => s.parent_step_id === step.id)
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

function parseStepAttributes(attributes: StepInstance['attributes'] | string | null | undefined) {
  if (!attributes) return {}
  if (typeof attributes !== 'string') return attributes
  try {
    return JSON.parse(attributes)
  } catch {
    return {}
  }
}

function getStepOperationId(step: StepInstance): number {
  return step.step_template_id || step.id
}

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

// WebSocket 实时刷新 — 区分事件类型，步骤事件增量更新
let logDebounceTimer: number | null = null
let componentDestroyed = false

function connectWebSocket() {
  if (componentDestroyed) return
  if (ws) ws.close()
  if (!isValidDrill.value) return
  const wsProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
  const authStore = useAuthStore()
  const wsUrl = `${wsProtocol}://${window.location.host}/ws/control/${drillId.value}?token=${authStore.token}`

  ws = new WebSocket(wsUrl)
  ws.onmessage = (ev: MessageEvent) => {
    if (componentDestroyed) return
    try {
      const msg = JSON.parse(ev.data)
      handleWSMessage(msg)
    } catch {
      // 解析失败时兜底全量刷新
      loadDrillData()
    }
  }
  ws.onerror = () => {
    if (componentDestroyed) return
    startFallbackPolling()
  }
  ws.onclose = () => {
    if (componentDestroyed) return
    if (instance.value?.status === 'running') {
      startFallbackPolling()
    }
  }
}

function handleWSMessage(msg: { event_type?: string; event?: string; payload?: any }) {
  const event = msg.event_type || msg.event || ''
  const payload = msg.payload || {}

  // 心跳忽略
  if (event === 'ping' || event === 'pong') return

  // 步骤事件：增量更新本地步骤数据
  if (event.startsWith('step_')) {
    patchLocalStep(payload)
    // 日志防抖刷新（300ms 内合并多次步骤变更）
    if (logDebounceTimer) clearTimeout(logDebounceTimer)
    logDebounceTimer = window.setTimeout(() => {
      logDebounceTimer = null
      refreshLogs()
    }, 300)
    return
  }

  // 演练状态变更：刷新实例详情 + 日志
  if (event.startsWith('drill_')) {
    refreshInstanceDetail()
    refreshLogs()
    return
  }

  // 其他事件：只刷新日志
  refreshLogs()
}

// 增量更新本地步骤数据（不调 API）
function patchLocalStep(payload: any) {
  const stepId = payload.step_id || payload.stepId
  if (!stepId) return

  const idx = steps.value.findIndex(s => s.id === stepId)
  if (idx === -1) return

  const newStatus = payload.new_status || payload.newStatus
  if (!newStatus) return

  const step = { ...steps.value[idx] }
  step.status = newStatus
  if (payload.start_time) step.start_time = payload.start_time
  if (payload.end_time) step.end_time = payload.end_time
  if (payload.timeout_at) step.timeout_at = payload.timeout_at
  if (payload.assignee_names) step.assignee_names = payload.assignee_names
  if (payload.remark) step.remark = payload.remark
  if (payload.issue_desc) step.issue_desc = payload.issue_desc

  const newSteps = [...steps.value]
  newSteps[idx] = step
  steps.value = newSteps
}

// 只刷新实例详情（1 个 API）
async function refreshInstanceDetail() {
  try {
    instance.value = await drillApi.getDetail(drillId.value)
  } catch { /* 静默失败 */ }
}

// 只刷新日志（1 个 API）
async function refreshLogs() {
  try {
    logs.value = await drillApi.getLogs(drillId.value)
  } catch { /* 静默失败 */ }
}

function startFallbackPolling() {
  if (componentDestroyed) return
  if (refreshTimer) clearInterval(refreshTimer)
  refreshTimer = window.setInterval(() => {
    if (componentDestroyed) {
      stopFallbackPolling()
      return
    }
    if (instance.value?.status === 'running') {
      loadDrillData()
    } else {
      stopFallbackPolling()
    }
  }, 5000)
}

function stopFallbackPolling() {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

async function loadDrillData() {
  if (componentDestroyed) return
  if (!isValidDrill.value) {
    router.replace('/director/drills')
    return
  }
  try {
    instance.value = await drillApi.getDetail(drillId.value)
    if (componentDestroyed) return
    if (!instance.value) {
      ElMessage.error('演练不存在')
      router.back()
      return
    }
    let stepsData = await drillApi.getSteps(drillId.value)
    if (componentDestroyed) return
    // 解析 attributes JSON 字符串为对象
    stepsData = stepsData.map((step: StepInstance) => ({
      ...step,
      attributes: parseStepAttributes(step.attributes),
    }))
    steps.value = stepsData

    const logsData = await drillApi.getLogs(drillId.value)
    if (componentDestroyed) return
    logs.value = logsData

    // 初始化展开所有 phase 面板
    initActivePhase()

    // 连接 WebSocket，演练运行中自动刷新状态
    connectWebSocket()
  } catch (error: any) {
    if (componentDestroyed) return
    if (error?.response?.status === 404) {
      ElMessage.error('演练不存在或已删除')
      router.back()
    } else {
      ElMessage.error(error?.message || '加载数据失败')
    }
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

async function handleStartStep(step: StepInstance) {
  try {
    await drillApi.startStep(drillId.value, getStepOperationId(step))
    ElMessage.success('步骤已开始')
    loadDrillData()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '操作失败')
  }
}

async function handleSkipStep(step: StepInstance) {
  try {
    await drillApi.skipStep(drillId.value, getStepOperationId(step), 'director skipped')
    ElMessage.success('步骤已跳过')
    loadDrillData()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '操作失败')
  }
}

async function handleResumeTask(step: StepInstance) {
  try {
    await drillApi.resumeTask(drillId.value, getStepOperationId(step))
    ElMessage.success('任务已重新派发')
    loadDrillData()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '操作失败')
  }
}

async function handleDirectorComplete(step: StepInstance) {
  try {
    // pending 状态下后端 CompleteStep 会校验失败,自动降级为强制完成
    if (step.status === 'pending') {
      await drillApi.forceCompleteStep(drillId.value, getStepOperationId(step), `指挥组完成任务：${step.name}`)
      ElMessage.success(`步骤「${step.name}」已完成`)
    } else {
      await drillApi.completeStep(drillId.value, getStepOperationId(step), '指挥组完成任务')
      ElMessage.success('步骤已完成')
    }
    loadDrillData()
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '操作失败')
  }
}

async function handleForceComplete(step: StepInstance) {
  try {
    await drillApi.forceCompleteStep(drillId.value, getStepOperationId(step), `指挥组强制完成步骤：${step.name}`)
    ElMessage.success(`步骤「${step.name}」已强制完成`)
    loadDrillData()
  } catch (error) {
    ElMessage.error('强制完成失败')
  }
}

function showStepDetail(step: StepInstance) {
  selectedStep.value = step
  detailEditing.value = false
  detailVisible.value = true
}

function enterEditMode() {
  if (!selectedStep.value) return
  const s = selectedStep.value
  // 逐字段赋值(避免直接替换 reactive 对象丢失响应性)
  editForm.timeout_minutes = s.timeout_minutes || 0
  editForm.executor_team = s.executor_team || ''
  editForm.assignee_names = s.assignee_names || ''
  editForm.remark = ''
  const attrs = s.attributes || {}
  editForm.attributes.operation_guide = attrs.operation_guide || ''
  editForm.attributes.verification_method = attrs.verification_method || ''
  editForm.attributes.fallback_measures = attrs.fallback_measures || ''
  editForm.attributes.responsible_department = attrs.responsible_department || ''
  editForm.attributes.cooperating_department = attrs.cooperating_department || ''
  editForm.attributes.operator = attrs.operator || ''
  editForm.attributes.reviewer = attrs.reviewer || ''
  detailEditing.value = true
  ensureUserOptions()
}

function cancelEditMode() {
  detailEditing.value = false
}

async function saveStepInfo() {
  if (!selectedStep.value) return
  const s = selectedStep.value
  const opId = getStepOperationId(s)
  detailSaving.value = true
  try {
    // reactive 对象直接读取,无需 .value
    const cleanedAttrs: Record<string, string> = {}
    Object.entries(editForm.attributes).forEach(([k, v]) => {
      if (v && String(v).trim()) cleanedAttrs[k] = String(v).trim()
    })
    console.log('[StepEdit] cleanedAttrs =', cleanedAttrs, 'remark =', editForm.remark)
    await drillApi.updateStepInfo(drillId.value, opId, {
      attributes: cleanedAttrs,
      remark: editForm.remark,
    })
    ElMessage.success('步骤信息已保存')
    detailEditing.value = false
    await loadDrillData()
    // 强制清缓存:重新从最新数据里拿一份
    const refreshed = steps.value.find(x => x.id === s.id)
    if (refreshed) {
      // 显式拷贝,避免引用旧对象
      selectedStep.value = { ...refreshed }
    }
  } catch (error: any) {
    console.error('[StepEdit] save error', error)
    ElMessage.error(error.response?.data?.message || '保存失败')
  } finally {
    detailSaving.value = false
  }
}

onMounted(() => {
  componentDestroyed = false
  loadDrillData()
})

onUnmounted(() => {
  componentDestroyed = true
  if (ws) {
    ws.close()
    ws = null
  }
  stopFallbackPolling()
})
</script>

<style scoped lang="scss">
@use '@/styles/layout' as *;
@use '@/styles/variables' as *;

.detail-toolbar {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px dashed #e6e8eb;
}

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
            color: $color-accent;
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

        .parent-hint {
          font-size: $font-size-xs;
          color: $text-tertiary;
          font-style: italic;
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

  .phase-tabs {
    .tab-label {
      display: flex;
      align-items: center;
      gap: $spacing-xs;
    }
  }

  .sub-task-content {
    white-space: pre-wrap;
    word-break: break-all;
    max-height: 200px;
    overflow-y: auto;
    font-size: $font-size-xs;
    line-height: 1.6;
    color: $text-secondary;
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
