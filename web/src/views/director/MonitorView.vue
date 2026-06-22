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

      <!-- 异常步骤 -->
      <el-card v-if="issueSteps.length > 0" class="issue-card" shadow="never">
        <template #header>
          <div class="issue-header">
            <el-icon class="issue-icon"><Warning /></el-icon>
            <span class="issue-title">异常 ({{ issueSteps.length }})</span>
          </div>
        </template>
        <div class="issue-list">
        <el-alert
          v-for="step in issueSteps"
          :key="step.id"
          :title="step.name"
          :description="getIssueDescription(step)"
          type="error"
          :closable="false"
          show-icon
          class="issue-item"
        >
          <template #default>
            <div class="issue-actions">
              <ActionConfirm
                :title="`重新派发：${step.name}`"
                message="将步骤状态重置为运行中，是否继续？"
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
                <el-descriptions-item v-if="getStepOperator(step)" label="操作人">
                  {{ getStepOperator(step) }}
                </el-descriptions-item>
              </el-descriptions>
            </div>
            <div class="step-actions">
              <el-button
                v-if="step.default_assignee_role === 'director' && !isParentStep(step)"
                type="success"
                size="small"
                :disabled="instance?.status === 'paused'"
                @click="handleDirectorComplete(step)"
              >
                <el-icon><CircleCheckFilled /></el-icon>
                完成任务
              </el-button>
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

      <!-- 步骤列表 (树状展示) -->
      <el-card class="steps-card">
        <template #header>
          <div class="steps-card-header">
            <div class="steps-card-title-area">
              <span class="card-title">步骤列表</span>
              <div class="depth-legend">
                <span class="legend-item legend-phase">阶段</span>
                <span class="legend-item legend-link">环节</span>
                <span class="legend-item legend-task">任务</span>
                <span class="legend-item legend-step">步骤</span>
              </div>
            </div>
            <div class="steps-card-actions">
              <el-button size="small" @click="expandAllSteps">全部展开</el-button>
              <el-button size="small" @click="collapseAllSteps">全部折叠</el-button>
            </div>
          </div>
        </template>
        <el-table
          ref="stepsTableRef"
          :data="drillStepTree"
          row-key="id"
          :tree-props="{ children: 'children' }"
          :default-expand-all="true"
          :row-class-name="stepRowClassName"
          style="width: 100%"
          size="small"
        >
          <el-table-column prop="name" label="步骤名" min-width="260" show-overflow-tooltip class-name="name-col">
            <template #default="{ row }">
              <div class="step-name-cell">
                <span
                  class="depth-badge"
                  :class="`depth-badge-${getStepDepth(row)}`"
                >
                  <span class="depth-label">{{ getDepthLabel(getStepDepth(row)) }}</span>
                  <span class="depth-num">{{ getSiblingIndex(row) }}</span>
                </span>
                <span :class="['step-name-text', `name-depth-${getStepDepth(row)}`]">{{ row.name }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="step_type" label="类型" width="70" align="center">
            <template #default="{ row }">
              <el-tag v-if="row.step_type" :type="getStepTypeTag(row.step_type)" size="small">{{ getStepTypeLabel(row.step_type) }}</el-tag>
              <template v-else>-</template>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="140" align="center">
            <template #default="{ row }">
              <template v-if="isParentStep(row)">
                <span class="parent-status-text">{{ getStepStatusText(row).text }}</span>
              </template>
              <template v-else>
                <DrillStatusBadge :status="row.status" type="step" />
              </template>
            </template>
          </el-table-column>
          <el-table-column prop="default_assignee_role" label="角色" width="75" align="center">
            <template #default="{ row }">
              <template v-if="isParentStep(row)">-</template>
              <template v-else>
                <el-tag :type="row.default_assignee_role === 'director' ? 'warning' : 'primary'" size="small">
                  {{ row.default_assignee_role === 'director' ? '指挥组' : '执行组' }}
                </el-tag>
              </template>
            </template>
          </el-table-column>
          <el-table-column prop="attributes" label="操作人" width="90" align="center" show-overflow-tooltip>
            <template #default="{ row }">
              {{ isParentStep(row) ? '-' : (row.attributes?.operator || '-') }}
            </template>
          </el-table-column>
          <el-table-column prop="attributes" label="复核人" width="90" align="center" show-overflow-tooltip>
            <template #default="{ row }">
              {{ isParentStep(row) ? '-' : (row.attributes?.reviewer || '-') }}
            </template>
          </el-table-column>
          <el-table-column label="实际耗时" width="85" align="center">
            <template #default="{ row }">
              {{ isParentStep(row) ? '-' : calculateDuration(row) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="220" align="center" fixed="right">
            <template #default="{ row }">
              <template v-if="!row._isGroup">
                <el-button type="primary" link size="small" @click="showStepDetail(row)">详情</el-button>
                <template v-if="!isParentStep(row)">
                  <el-tooltip
                    v-if="row.status === 'pending'"
                    :content="getStartDisabledReason(row)"
                    :disabled="canStartStep(row)"
                    placement="top"
                  >
                    <span>
                      <el-button
                        type="primary"
                        link
                        size="small"
                        :disabled="!canStartStep(row)"
                        @click="handleStartStep(row)"
                      >开始</el-button>
                    </span>
                  </el-tooltip>
                  <el-button
                    v-if="row.status === 'running'"
                    type="success"
                    link
                    size="small"
                    @click="handleDirectorComplete(row)"
                  >完成</el-button>
                  <el-button
                    v-if="row.status === 'issue'"
                    type="success"
                    link
                    size="small"
                    @click="confirmStepAction('步骤异常，确认强制完成？', () => handleForceComplete(row))"
                  >
                    完成
                  </el-button>
                  <el-button
                    v-if="['pending', 'running'].includes(row.status)"
                    type="warning"
                    link
                    size="small"
                    @click="confirmStepAction('确认跳过此步骤？', () => handleSkipStep(row))"
                  >
                    跳过
                  </el-button>
                  <el-button
                    v-if="['pending', 'running'].includes(row.status)"
                    type="danger"
                    link
                    size="small"
                    @click="confirmStepAction('确认强制完成此步骤？', () => handleForceComplete(row))"
                  >
                    强制完成
                  </el-button>
                  <el-button
                    v-if="['completed', 'skipped'].includes(row.status)"
                    type="primary"
                    link
                    size="small"
                    @click="confirmStepAction('确认重新派发此步骤？', () => handleResumeTask(row))"
                  >
                    重新派发
                  </el-button>
                </template>
                <span v-else class="parent-hint">
                  子任务全部完成后自动完成
                </span>
              </template>
            </template>
          </el-table-column>
        </el-table>
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
                <el-form-item label="执行团队">{{ selectedStep.executor_team || '-' }}</el-form-item>
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
            <el-form-item label="操作人">
              <el-autocomplete
                :model-value="editForm.attributes.operator"
                @update:model-value="(v: string | number) => (editForm.attributes.operator = String(v ?? ''))"
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
                @update:model-value="(v: string | number) => (editForm.attributes.reviewer = String(v ?? ''))"
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
import { ref, shallowRef, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Edit } from '@element-plus/icons-vue'
import { VideoPause, VideoPlay, VideoCamera, Back, Warning, DArrowRight, CircleCheck, RefreshRight, CircleCheckFilled } from '@element-plus/icons-vue'
import type { DrillInstance, StepInstance, StepStatus } from '@/types'
import DrillStatusBadge from '@/components/common/DrillStatusBadge.vue'
import ActionConfirm from '@/components/common/ActionConfirm.vue'
import { drillApi } from '@/api/modules/drill'
import { userApi } from '@/api/modules/user'
import { useAuthStore } from '@/stores/auth'

const stepsTableRef = ref()
const selectedStep = ref<StepInstance | null>(null)
const detailVisible = ref(false)
const detailEditing = ref(false)
const detailSaving = ref(false)
const editFormRef = ref()

const editForm = reactive({
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
const steps = shallowRef<StepInstance[]>([])
const logs = shallowRef<any[]>([])

// WebSocket
let ws: WebSocket | null = null
let refreshTimer: number | null = null

const drillId = computed(() => {
  const id = route.params.id
  const parsed = typeof id === 'string' ? parseInt(id, 10) : 0
  return Number.isFinite(parsed) && parsed > 0 ? parsed : 0
})

const isValidDrill = computed(() => drillId.value > 0)

const sortedSteps = computed(() => {
  return [...steps.value].sort((a, b) => a.seq - b.seq)
})

const drillSteps = computed(() => sortedSteps.value)

const stepById = computed(() => {
  const map = new Map<number, StepInstance>()
  for (const step of sortedSteps.value) {
    map.set(step.id, step)
  }
  return map
})

const childrenByParentId = computed(() => {
  const map = new Map<number, StepInstance[]>()
  for (const step of sortedSteps.value) {
    if (!step.parent_step_id) continue
    const children = map.get(step.parent_step_id) || []
    children.push(step)
    map.set(step.parent_step_id, children)
  }
  return map
})

const parentIdSet = computed(() => new Set(childrenByParentId.value.keys()))

// 将扁平步骤数据转换为树形结构
// 优先使用 parent_step_id 建立真实的层级关系
// 仅当数据无层级关系时，才退回到 phase/phase_step 虚拟分组
const drillStepTree = computed(() => {
  const sorted = sortedSteps.value

  // 建立 step → node 映射
  const stepMap = new Map<number, any>()
  for (const step of sorted) {
    stepMap.set(step.id, { ...step })
  }

  // 通过 parent_step_id 建立父子关系
  const childIds = new Set<number>()
  for (const step of sorted) {
    if (step.parent_step_id && stepMap.has(step.parent_step_id)) {
      const parent = stepMap.get(step.parent_step_id)!
      if (!parent.children) parent.children = []
      parent.children.push(stepMap.get(step.id)!)
      childIds.add(step.id)
    }
  }

  // 根步骤
  const rootSteps = sorted.filter(s => !childIds.has(s.id)).map(s => stepMap.get(s.id)!)

  // 判断数据是否已有真实层级：有 parent_step_id 关系的步骤数 > 0
  const hasHierarchy = childIds.size > 0

  if (hasHierarchy) {
    // 数据本身有层级结构，直接用 parent_step_id 构建的树
    return rootSteps
  }

  // 无层级数据：退回 phase → phase_step 虚拟分组
  const phaseMap = new Map<string, any>()
  const roots: any[] = []
  let virtualSeq = -1

  for (const step of rootSteps) {
    const phase = step.phase || '未分类'

    if (!phaseMap.has(phase)) {
      const phaseNode = {
        id: `__phase__${phase}`,
        name: phase,
        seq: virtualSeq--,
        status: '',
        step_type: 'phase',
        timeout_minutes: 0,
        default_assignee_role: '',
        estimated_duration_minutes: 0,
        attributes: {},
        start_time: null,
        end_time: null,
        timeout_at: null,
        assignee_names: '',
        remark: '',
        issue_desc: '',
        executor_team: '',
        step_template_id: 0,
        drill_instance_id: 0,
        children: [] as any[],
        _isGroup: true,
        _groupType: 'phase' as const,
      }
      phaseMap.set(phase, phaseNode)
      roots.push(phaseNode)
    }

    const phaseNode = phaseMap.get(phase)!
    const phaseStep = step.phase_step || '默认'

    let phaseStepNode = phaseNode.children.find(
      (c: any) => c._isGroup && c._groupType === 'phaseStep' && c.name === phaseStep
    )
    if (!phaseStepNode) {
      phaseStepNode = {
        id: `__phaseStep__${phase}__${phaseStep}`,
        name: phaseStep,
        seq: virtualSeq--,
        status: '',
        step_type: 'phaseStep',
        timeout_minutes: 0,
        default_assignee_role: '',
        estimated_duration_minutes: 0,
        attributes: {},
        start_time: null,
        end_time: null,
        timeout_at: null,
        assignee_names: '',
        remark: '',
        issue_desc: '',
        executor_team: '',
        step_template_id: 0,
        drill_instance_id: 0,
        children: [] as any[],
        _isGroup: true,
        _groupType: 'phaseStep' as const,
      }
      phaseNode.children.push(phaseStepNode)
    }

    phaseStepNode.children.push(step)
  }

  return roots
})

// 全部展开/折叠
function expandAllSteps() {
  toggleExpandAll(drillStepTree.value, true)
}

function collapseAllSteps() {
  toggleExpandAll(drillStepTree.value, false)
}

function toggleExpandAll(data: any[], expand: boolean) {
  data.forEach(item => {
    stepsTableRef.value?.toggleRowExpansion(item, expand)
    if (item.children && item.children.length > 0) {
      toggleExpandAll(item.children, expand)
    }
  })
}

// 父步骤/分组状态聚合显示
function getStepStatusText(row: any): { text: string; isParent: boolean } {
  return stepTreeMetaMap.value.get(row.id)?.statusText || { text: row.status, isParent: false }
}

// 检查是否为父步骤（虚拟分组节点 或 有子步骤的真实步骤）
function isParentStep(step: any): boolean {
  if (step._isGroup) return true
  return stepTreeMetaMap.value.get(step.id)?.isParent || parentIdSet.value.has(step.id)
}

// 终态集合：已完成/已跳过/异常
const TERMINAL_STATUSES: StepStatus[] = ['completed', 'skipped', 'issue']
const DEPENDENCY_SATISFIED_STATUSES: StepStatus[] = ['completed', 'issue']
const terminalStatusSet = new Set<StepStatus>(TERMINAL_STATUSES)
const dependencySatisfiedStatusSet = new Set<StepStatus>(DEPENDENCY_SATISFIED_STATUSES)

type StepTreeMeta = {
  depth: number
  siblingIndex: number
  isParent: boolean
  statusText: { text: string; isParent: boolean }
}

const stepTreeMetaMap = computed(() => {
  const metaMap = new Map<number | string, StepTreeMeta>()

  function visit(nodes: any[], depth: number) {
    let total = 0
    let completed = 0
    let running = 0

    nodes.forEach((node, index) => {
      const hasChildren = Boolean(node.children?.length)
      const childSummary = hasChildren ? visit(node.children, depth + 1) : null
      const nodeTotal = childSummary?.total ?? 1
      const nodeCompleted = childSummary?.completed ?? (terminalStatusSet.has(node.status) ? 1 : 0)
      const nodeRunning = childSummary?.running ?? (node.status === 'running' ? 1 : 0)
      const isParent = Boolean(node._isGroup || hasChildren)
      const statusText = isParent
        ? {
            text: nodeRunning > 0
              ? `${nodeCompleted}/${nodeTotal} 已完成 · ${nodeRunning} 进行中`
              : `${nodeCompleted}/${nodeTotal} ${node._isGroup ? '已完成' : '子任务已完成'}`,
            isParent: true,
          }
        : { text: node.status, isParent: false }

      metaMap.set(node.id, {
        depth,
        siblingIndex: index + 1,
        isParent,
        statusText,
      })

      total += nodeTotal
      completed += nodeCompleted
      running += nodeRunning
    })

    return { total, completed, running }
  }

  visit(drillStepTree.value, 0)
  return metaMap
})

// 判断步骤是否可以手动开始
function canStartStep(step: any): boolean {
  return !getStartDisabledReason(step)
}

const stepStartDisabledReasonMap = computed(() => {
  const map = new Map<number, string>()
  for (const step of sortedSteps.value) {
    map.set(step.id, resolveStartDisabledReason(step))
  }
  return map
})

// 获取步骤不能开始的原因，返回空字符串表示可以开始
function getStartDisabledReason(step: any): string {
  if (typeof step.id === 'number') {
    return stepStartDisabledReasonMap.value.get(step.id) ?? resolveStartDisabledReason(step)
  }
  return resolveStartDisabledReason(step)
}

function resolveStartDisabledReason(step: any): string {
  if (instance.value?.status !== 'running') return '演练未处于执行中状态'
  if (step.status !== 'pending') return '步骤不在待执行状态'
  if (isParentStep(step)) return '父步骤不可手动开始'

  const ownDependencyReason = getStepDependencyDisabledReason(step)
  if (ownDependencyReason) return ownDependencyReason

  // 检查父步骤链是否可被激活：pending 祖先允许由后端自动启动，但祖先自身的依赖必须已满足
  let ancestor = step
  while (ancestor.parent_step_id) {
    const parent = stepById.value.get(ancestor.parent_step_id)
    if (!parent) break
    if (parent.status === 'pending') {
      const parentDependencyReason = getStepDependencyDisabledReason(parent)
      if (parentDependencyReason) {
        return `父步骤未满足开始条件：${parent.name}（${parentDependencyReason}）`
      }
    } else if (parent.status !== 'running' && parent.status !== 'completed') {
      return `父步骤未启动：${parent.name}`
    }
    ancestor = parent
  }

  return ''
}

function getStepDependencyDisabledReason(step: any): string {
  // 检查 pre_step_ids 中的前序步骤是否全部完成
  const preStepIds = step.pre_step_ids || []
  if (preStepIds.length > 0) {
    const pendingPreSteps: string[] = []
    preStepIds.forEach((preId: number) => {
      const preStep = stepById.value.get(preId)
      if (!preStep) {
        // 前序步骤未找到，视为未完成
        pendingPreSteps.push(`步骤#${preId}`)
      } else if (!dependencySatisfiedStatusSet.has(preStep.status)) {
        pendingPreSteps.push(preStep.name)
      }
    })
    if (pendingPreSteps.length > 0) {
      return `前序步骤未完成：${pendingPreSteps.join('、')}`
    }
  }

  // 串行步骤兜底检查：如果 pre_step_ids 为空但同级存在更早的未完成串行步骤，也应禁用
  if (preStepIds.length === 0 && step.parent_step_id) {
    const parent = stepById.value.get(step.parent_step_id)
    if (parent?.step_type === 'parallel') {
      return ''
    }
    const siblings = childrenByParentId.value.get(step.parent_step_id) || []
    const earlierPendingSibling = siblings.find(s =>
      s.id !== step.id && s.seq < step.seq && s.step_type === 'serial' && !dependencySatisfiedStatusSet.has(s.status)
    )
    if (earlierPendingSibling) {
      return `前序步骤未完成：${earlierPendingSibling.name}`
    }
  }

  return ''
}

// 计算步骤在树中的深度（0=根/阶段, 1=环节, 2=任务, 3=步骤）
function getStepDepth(step: any): number {
  const meta = stepTreeMetaMap.value.get(step.id)
  if (meta) return meta.depth
  let depth = 0
  let current = step
  while (current.parent_step_id) {
    depth++
    const parent = stepById.value.get(current.parent_step_id)
    if (!parent) break
    current = parent
  }
  return depth
}

function getSiblingIndex(step: any): number {
  return stepTreeMetaMap.value.get(step.id)?.siblingIndex || 0
}

// 深度→层级标签
function getDepthLabel(depth: number): string {
  const labels = ['阶段', '环节', '任务', '步骤']
  return labels[Math.min(depth, labels.length - 1)] || '步骤'
}

// 深度→标签颜色类型
function getDepthTagType(depth: number): 'primary' | 'success' | 'warning' | 'info' | 'danger' {
  const types: Record<number, 'primary' | 'success' | 'warning' | 'info' | 'danger'> = {
    0: 'primary', 1: 'success', 2: 'warning', 3: 'info',
  }
  return types[Math.min(depth, 3)] || 'info'
}

// el-table 行类名：按深度添加样式
function stepRowClassName({ row }: { row: any }): string {
  const depth = getStepDepth(row)
  return `step-depth-${Math.min(depth, 3)}`
}

const completedSteps = computed(() => {
  return sortedSteps.value.filter(s => terminalStatusSet.has(s.status)).length
})

const totalSteps = computed(() => {
  return steps.value.length || 1
})

const progressPercentage = computed(() => {
  return Math.round((completedSteps.value / totalSteps.value) * 100)
})

const runningSteps = computed(() => {
  return sortedSteps.value.filter(s => s.status === 'running')
})

const issueSteps = computed(() => {
  return sortedSteps.value.filter(s => s.status === 'issue')
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
  return logs.value
    .filter(log => !['timeout', 'step_timeout'].includes(log.action))
    .map(log => ({
      ...log,
      step_name: null,
      operator: log.operator_name,
      remark: typeof log.content === 'string' ? log.content.replace(/超时/g, '结束') : log.content,
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

function parsePreStepIds(preStepIds: number[] | string | null | undefined): number[] {
  if (!preStepIds) return []
  if (Array.isArray(preStepIds)) return preStepIds
  if (typeof preStepIds === 'string') {
    try {
      const parsed = JSON.parse(preStepIds)
      return Array.isArray(parsed) ? parsed : []
    } catch {
      return []
    }
  }
  return []
}

function normalizeStepForMonitor(step: StepInstance): StepInstance {
  if (step.status !== 'timeout') return step
  return {
    ...step,
    status: 'running',
    timeout_at: null,
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
    force_complete: 'warning',
    resume_task: 'primary',
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
    force_complete: '强制完成',
    resume_task: '重新派发',
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
  if (step.status === 'issue') {
    return '步骤执行异常，需要处理'
  }
  return ''
}

function getStepOperator(step: any): string {
  if (!step) return ''
  let attrs = step.attributes
  if (typeof attrs === 'string') {
    try { attrs = JSON.parse(attrs) } catch { return '' }
  }
  return attrs?.operator || ''
}

// WebSocket 实时刷新 — 区分事件类型，步骤事件增量更新
let logDebounceTimer: number | null = null
let drillRefreshTimer: number | null = null
let drillDataLoading = false
let drillDataReloadQueued = false
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
      scheduleDrillDataRefresh()
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

  // 步骤事件：全量刷新步骤数据（步骤完成会级联更新父步骤状态）
  if (event.startsWith('step_')) {
    scheduleDrillDataRefresh()
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

function scheduleDrillDataRefresh(delay = 120) {
  if (componentDestroyed) return
  if (drillRefreshTimer) {
    clearTimeout(drillRefreshTimer)
  }
  drillRefreshTimer = window.setTimeout(() => {
    drillRefreshTimer = null
    loadDrillData()
  }, delay)
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
  step.status = newStatus === 'timeout' ? 'running' : newStatus
  if (payload.start_time) step.start_time = payload.start_time
  if (payload.end_time) step.end_time = payload.end_time
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
  if (drillDataLoading) {
    drillDataReloadQueued = true
    return
  }
  if (drillRefreshTimer) {
    clearTimeout(drillRefreshTimer)
    drillRefreshTimer = null
  }
  drillDataLoading = true
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
    // 解析 attributes JSON 字符串为对象，解析 pre_step_ids 为数组
    stepsData = stepsData.map((step: StepInstance) => ({
      ...normalizeStepForMonitor(step),
      attributes: parseStepAttributes(step.attributes),
      pre_step_ids: parsePreStepIds(step.pre_step_ids),
    }))
    steps.value = stepsData

    const logsData = await drillApi.getLogs(drillId.value)
    if (componentDestroyed) return
    logs.value = logsData

    // 连接 WebSocket，演练运行中自动刷新状态
    // 仅在 WebSocket 未连接时建立连接，避免刷新数据时重连导致循环
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      connectWebSocket()
    }
  } catch (error: any) {
    if (componentDestroyed) return
    if (error?.response?.status === 404) {
      ElMessage.error('演练不存在或已删除')
      router.back()
    } else {
      ElMessage.error(error?.message || '加载数据失败')
    }
    console.error('Failed to load drill data:', error)
  } finally {
    drillDataLoading = false
    if (drillDataReloadQueued && !componentDestroyed) {
      drillDataReloadQueued = false
      scheduleDrillDataRefresh(0)
    }
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
    scheduleDrillDataRefresh(0)
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '操作失败')
  }
}

async function confirmStepAction(message: string, action: () => Promise<void>) {
  try {
    await ElMessageBox.confirm(message, '确认操作', {
      confirmButtonText: '确认',
      cancelButtonText: '取消',
      type: 'warning',
    })
    await action()
  } catch {
    // 取消操作或弹窗关闭，不提示错误
  }
}

async function handleSkipStep(step: StepInstance) {
  try {
    await drillApi.skipStep(drillId.value, getStepOperationId(step), 'director skipped')
    ElMessage.success('步骤已跳过')
    scheduleDrillDataRefresh(0)
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || '操作失败')
  }
}

async function handleResumeTask(step: StepInstance) {
  try {
    await drillApi.resumeTask(drillId.value, getStepOperationId(step))
    ElMessage.success('任务已重新派发')
    scheduleDrillDataRefresh(0)
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
    scheduleDrillDataRefresh(0)
  } catch (error: any) {
    ElMessage.error(error.response?.data?.message || error.message || '操作失败')
  }
}

async function handleForceComplete(step: StepInstance) {
  try {
    await drillApi.forceCompleteStep(drillId.value, getStepOperationId(step), `指挥组强制完成步骤：${step.name}`)
    ElMessage.success(`步骤「${step.name}」已强制完成`)
    scheduleDrillDataRefresh(0)
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
  if (drillRefreshTimer) {
    clearTimeout(drillRefreshTimer)
    drillRefreshTimer = null
  }
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

  .steps-card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: $spacing-base;
    flex-wrap: wrap;

    .steps-card-title-area {
      display: flex;
      align-items: center;
      gap: $spacing-lg;
    }

    .depth-legend {
      display: flex;
      gap: $spacing-xs;
    }

    .legend-item {
      display: inline-flex;
      align-items: center;
      font-size: 11px;
      padding: 2px 8px;
      border-radius: 3px;
      color: #fff;
      font-weight: $font-weight-medium;
    }

    .legend-phase { background: var(--el-color-primary); }
    .legend-link { background: var(--el-color-success); }
    .legend-task { background: var(--el-color-warning); }
    .legend-step { background: #64748B; }

    .steps-card-actions {
      display: flex;
      gap: 4px;
    }
  }

  // 深度徽章（步骤名列内）
  .depth-badge {
    display: inline-flex; align-items: center; gap: 0;
    margin-right: 8px; flex-shrink: 0;
    height: 20px; border-radius: 3px;
    font-size: 11px; font-weight: 700;
    letter-spacing: 0.5px; white-space: nowrap;
    overflow: hidden;
    line-height: 1;

    .depth-label {
      padding: 0 5px;
    }
    .depth-num {
      display: inline-flex; align-items: center; justify-content: center;
      min-width: 16px; height: 20px;
      padding: 0 3px;
      font-size: 10px; font-weight: 800;
      font-family: $font-family-mono;
    }
  }

  // 各层级配色
  .depth-badge-0 { background: rgba(24, 144, 255, 0.1); color: #1890ff; .depth-num { background: #1890ff; color: #fff; } }
  .depth-badge-1 { background: rgba(82, 196, 26, 0.1); color: #52c41a; .depth-num { background: #52c41a; color: #fff; } }
  .depth-badge-2 { background: rgba(250, 173, 20, 0.1); color: #faad14; .depth-num { background: #faad14; color: #fff; } }
  .depth-badge-3 { background: rgba(100, 116, 139, 0.12); color: #64748b; .depth-num { background: #64748b; color: #fff; } }

  .step-name-text {
    font-size: 13px;
  }

  .name-depth-0 {
    font-weight: $font-weight-bold;
    color: var(--el-color-primary-dark-2);
  }

  .name-depth-1 {
    font-weight: $font-weight-semibold;
    color: var(--el-color-success-dark-2);
  }

  .name-depth-2 {
    font-weight: $font-weight-medium;
    color: $text-secondary;
  }

  .name-depth-3 {
    font-weight: $font-weight-regular;
    color: $text-tertiary;
    font-size: 12px;
  }

  // 表格行按深度的视觉区分
  :deep(.el-table) {
    // 展开箭头样式优化
    .el-table__expand-icon {
      font-size: 13px;
      color: $text-tertiary;
      transition: color 0.2s, transform 0.2s;

      &:hover {
        color: var(--el-color-primary);
      }
    }

    .el-table__expand-icon--expanded {
      color: var(--el-color-primary);
    }

    // 步骤名列布局
    .name-col .cell {
      display: flex;
      align-items: center;
    }

    .step-name-cell {
      display: flex;
      align-items: center;
      gap: 8px;
      overflow: hidden;
    }

    // 步骤名列左侧色带
    .step-depth-0 .name-col {
      border-left: 3px solid var(--el-color-primary);
    }
    .step-depth-1 .name-col {
      border-left: 3px solid var(--el-color-success);
    }
    .step-depth-2 .name-col {
      border-left: 3px solid var(--el-color-warning);
    }
    .step-depth-3 .name-col {
      border-left: 3px solid var(--el-color-info-light-5);
    }

    // 行背景色按深度微弱区分
    .step-depth-0 td { background: rgba(24, 144, 255, 0.04); }
    .step-depth-1 td { background: rgba(82, 196, 26, 0.03); }
    .step-depth-2 td { background: rgba(250, 173, 20, 0.02); }
  }

  .parent-status-text {
    font-size: 12px;
    color: $text-secondary;
  }

  .parent-hint {
    font-size: 11px;
    color: $text-tertiary;
    font-style: italic;
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
