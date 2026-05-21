<template>
  <div class="page-container">
    <div class="page-header">
      <div class="header-left">
        <el-button text @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
          返回模板列表
        </el-button>
        <h2 class="page-title">编辑步骤 - {{ templateName }}</h2>
      </div>
      <div class="header-actions">
        <el-button @click="loadTemplateSteps">
          <el-icon><Refresh /></el-icon>
          重新加载
        </el-button>
        <el-button @click="openPhaseManageDialog">
          <el-icon><Setting /></el-icon>
          阶段管理
        </el-button>
        <el-button type="success" @click="openBatchImportDialog">
          <el-icon><Download /></el-icon>
          批量导入
        </el-button>
        <el-button type="primary" @click="handleSaveSteps">
          <el-icon><Check /></el-icon>
          保存步骤
        </el-button>
      </div>
    </div>

    <div class="page-content steps-layout">
      <!-- 左侧：阶段 tab + 步骤树 -->
      <div class="steps-panel">
        <!-- 阶段 tabs -->
        <el-tabs v-model="activePhaseName" class="phase-tabs" type="card">
          <el-tab-pane
            v-for="phase in phases"
            :key="phase.name"
            :name="phase.name"
            :label="phase.name"
          />
          <el-tab-pane v-if="phases.length === 0" name="_empty" label="无阶段" disabled />
        </el-tabs>

        <div class="panel-header">
          <h3>步骤树</h3>
          <el-button type="primary" size="small" @click="openSingleAddDialog" :disabled="!activePhaseName">
            <el-icon><Plus /></el-icon>
            添加步骤
          </el-button>
        </div>
        <div class="panel-body">
          <div v-if="activeSteps.length > 0">
            <el-table
              :data="activeStepsTree"
              border
              size="small"
              row-key="id"
              highlight-current-row
              :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
              default-expand-all
              @current-change="handleRowSelect"
            >
              <el-table-column type="index" label="序号" width="60" align="center" />
              <el-table-column prop="name" label="步骤名称" min-width="120" show-overflow-tooltip>
                <template #default="{ row }">
                  <span :class="{ 'step-selected': selectedStep?.id === row.id }">
                    {{ row.name || '-' }}
                  </span>
                </template>
              </el-table-column>
              <el-table-column prop="step_type" label="类型" width="70" align="center">
                <template #default="{ row }">
                  <el-tag size="small" type="info">{{ getStepTypeLabel(row.step_type) }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="90" align="center" fixed="right">
                <template #default="{ row }">
                  <el-button text type="danger" size="small" @click="removeStepByRow(row)" title="删除">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
          <div v-else class="empty-steps">
            <el-empty :description="phases.length === 0 ? '请先添加阶段' : '暂无步骤，请添加或导入步骤'" :image-size="100">
              <el-button v-if="activePhaseName" type="primary" @click="openSingleAddDialog">
                <el-icon><Plus /></el-icon>
                添加步骤
              </el-button>
              <el-button v-if="activePhaseName" type="success" @click="openBatchImportDialog" style="margin-left: 8px">
                <el-icon><Download /></el-icon>
                批量导入
              </el-button>
            </el-empty>
          </div>
        </div>
      </div>

      <!-- 右侧：步骤详情 -->
      <div class="detail-panel">
        <div class="panel-header">
          <h3 v-if="selectedStep">步骤详情 - {{ selectedStep.name }}</h3>
          <h3 v-else>请选择一个步骤</h3>
          <div v-if="selectedStep" class="panel-actions">
            <el-button text type="primary" size="small" @click="openStepEditDialogForSelected">
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
          </div>
        </div>
        <div class="panel-body">
          <div v-if="selectedStep" class="step-detail">
            <el-descriptions :column="2" border size="default">
              <el-descriptions-item label="步骤名称" :span="2">{{ selectedStep.name }}</el-descriptions-item>
              <el-descriptions-item label="描述" :span="2">{{ selectedStep.description || '-' }}</el-descriptions-item>
              <el-descriptions-item label="父任务Id">{{ selectedStep.parent_step_id || '-' }}</el-descriptions-item>
              <el-descriptions-item label="步骤类型">{{ getStepTypeLabel(selectedStep.step_type) }}</el-descriptions-item>
              <el-descriptions-item label="预计耗时">{{ selectedStep.estimated_duration_minutes ? selectedStep.estimated_duration_minutes + ' 分钟' : '-' }}</el-descriptions-item>
              <el-descriptions-item label="开始偏移">{{ selectedStep.estimated_start_offset ? selectedStep.estimated_start_offset + ' 分钟' : '-' }}</el-descriptions-item>
            </el-descriptions>
            <el-divider>详细信息</el-divider>
            <el-descriptions :column="2" border size="default">
              <el-descriptions-item label="执行角色">{{ selectedStep.default_assignee_role ? (selectedStep.default_assignee_role === 'director' ? '指挥组' : '执行组') : '-' }}</el-descriptions-item>
              <el-descriptions-item label="执行团队">{{ selectedStep.executor_team || '-' }}</el-descriptions-item>
              <el-descriptions-item label="责任部门">{{ selectedStep.responsible_department || '-' }}</el-descriptions-item>
              <el-descriptions-item label="责任人">{{ selectedStep.responsible_person || '-' }}</el-descriptions-item>
              <el-descriptions-item label="执行人">{{ selectedStep.executor || '-' }}</el-descriptions-item>
              <el-descriptions-item label="复核人">{{ selectedStep.reviewer || '-' }}</el-descriptions-item>
              <el-descriptions-item label="任务名称" :span="2">{{ selectedStep.task_name || '-' }}</el-descriptions-item>
              <el-descriptions-item label="子任务描述" :span="2">{{ selectedStep.sub_task || '-' }}</el-descriptions-item>
            </el-descriptions>
          </div>
          <div v-else class="empty-detail">
            <el-empty description="请在左侧选择要查看的步骤" :image-size="80" />
          </div>
        </div>
      </div>
    </div>

    <!-- 阶段管理对话框 -->
    <el-dialog v-model="phaseManageVisible" title="阶段管理" width="500px">
      <div class="phase-manage-list">
        <div v-for="(phase, index) in editablePhases" :key="index" class="phase-manage-item">
          <el-input v-model="phase.name" placeholder="阶段名称，如：准备阶段" size="default" clearable />
          <el-button size="default" type="danger" text @click="removePhase(index)">
            <el-icon><Delete /></el-icon>
          </el-button>
        </div>
      </div>
      <el-button type="primary" plain style="width: 100%; margin-top: 12px" @click="addPhase">
        <el-icon><Plus /></el-icon>
        添加阶段
      </el-button>
      <template #footer>
        <el-button @click="phaseManageVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSavePhases">保存</el-button>
      </template>
    </el-dialog>

    <!-- 批量导入对话框 -->
    <el-dialog v-model="importVisible" title="批量导入步骤" width="520px">
      <div v-if="phases.length === 0" class="import-empty-hint">
        <el-empty description="请先添加阶段" :image-size="60" />
      </div>
      <template v-else>
        <el-form label-width="80px" class="import-form">
          <el-form-item label="目标阶段">
            <el-select v-model="importTargetPhase" placeholder="选择要导入的阶段" style="width: 100%">
              <el-option v-for="phase in phases" :key="phase.name" :label="phase.name" :value="phase.name" />
            </el-select>
          </el-form-item>
        </el-form>
        <div class="excel-upload">
          <el-upload
            ref="uploadRef"
            :before-upload="handleExcelUpload"
            :show-file-list="false"
            accept=".xlsx,.xls"
            class="upload-area"
          >
            <div class="upload-content">
              <el-icon class="upload-icon"><Upload /></el-icon>
              <div class="upload-text">点击或拖拽上传 Excel 文件</div>
              <div class="upload-hint">支持 .xlsx, .xls 格式</div>
            </div>
          </el-upload>
          <div class="template-download">
            <el-button type="info" plain @click="downloadTemplate">
              <el-icon><Download /></el-icon>
              下载 Excel 模板
            </el-button>
          </div>
        </div>
      </template>
    </el-dialog>

    <!-- 单个添加/编辑抽屉 -->
    <el-drawer v-model="singleAddVisible" :title="singleStepEditIndex !== null ? '编辑步骤' : '添加步骤'" size="720px">
      <el-form :model="singleStepForm" label-width="90px" class="single-step-form">
        <el-form-item label="步骤名称" required>
          <el-input v-model="singleStepForm.name" placeholder="请输入步骤名称" maxlength="100" show-word-limit />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="singleStepForm.description" type="textarea" placeholder="步骤描述" :rows="2" maxlength="500" show-word-limit />
        </el-form-item>
        <div class="form-row">
         <el-form-item label="父步骤" class="inline-form-item">
            <el-select v-model="singleStepForm.parent_step_id" clearable placeholder="可选" filterable>
              <el-option
                v-for="opt in formParentStepOptions"
                :key="opt.value"
                :label="opt.label"
                :value="opt.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="步骤类型" class="inline-form-item">
            <el-select v-model="singleStepForm.step_type">
              <el-option label="串行" value="serial" />
              <el-option label="并行" value="parallel" />
            </el-select>
          </el-form-item>
        </div>
        <div class="form-row">
          <el-form-item label="开始偏移" class="inline-form-item">
            <el-input-number v-model="singleStepForm.estimated_start_offset" :min="0" controls-position="right" placeholder="相对启动偏移" />
            <span class="unit-label">分钟</span>
          </el-form-item>
           <el-form-item label="预计耗时" class="inline-form-item">
            <el-input-number v-model="singleStepForm.estimated_duration_minutes" :min="1" :max="1440" controls-position="right" placeholder="可选" />
            <span class="unit-label">分钟</span>
          </el-form-item>
        </div>

        <el-divider>详细信息</el-divider>
         <div class="form-row">
          <el-form-item label="执行角色" class="inline-form-item">
            <el-select v-model="singleStepForm.default_assignee_role" clearable placeholder="可留空">
              <el-option label="指挥组" value="director" />
              <el-option label="执行组" value="executor" />
            </el-select>
          </el-form-item>
           <el-form-item label="执行团队" class="inline-form-item">
            <el-select v-model="singleStepForm.executor_team" clearable placeholder="选择团队" filterable allow-create>
              <el-option v-for="dept in departmentOptions" :key="dept" :label="dept" :value="dept" />
            </el-select>
          </el-form-item>
        </div>
        <div class="form-row">
          <el-form-item label="责任部门" class="inline-form-item">
            <el-input v-model="singleStepForm.responsible_department" placeholder="责任部门" clearable />
          </el-form-item>
          <el-form-item label="责任人" class="inline-form-item">
            <el-input v-model="singleStepForm.responsible_person" placeholder="责任人姓名" clearable />
          </el-form-item>
        </div>
        <div class="form-row">
          <el-form-item label="执行人" class="inline-form-item">
            <el-input v-model="singleStepForm.executor" placeholder="执行人姓名" clearable />
          </el-form-item>
          <el-form-item label="复核人" class="inline-form-item">
            <el-input v-model="singleStepForm.reviewer" placeholder="复核人姓名" clearable />
          </el-form-item>
        </div>

        <el-form-item label="任务名称">
          <el-input v-model="singleStepForm.task_name" placeholder="可独立于步骤名展示的任务名" clearable maxlength="128" show-word-limit />
        </el-form-item>
        <el-form-item label="子任务描述">
          <el-input v-model="singleStepForm.sub_task" type="textarea" placeholder="操作步骤/子任务详细描述" :rows="3" maxlength="2000" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="singleAddVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAddSingleStep">{{ singleStepEditIndex !== null ? '保存修改' : '添加步骤' }}</el-button>
      </template>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh, Plus, Delete, Upload, Download, Check, Edit, ArrowLeft, Setting } from '@element-plus/icons-vue'
import * as XLSX from 'xlsx'
import { userApi } from '@/api'
import { templateApi } from '@/api/modules/template'
import type { StepTemplate, StepType } from '@/types'

const route = useRoute()
const router = useRouter()

const templateId = computed(() => Number(route.params.id))

// ============ 阶段管理 ============

interface PhaseGroup {
  name: string
  steps: StepTemplate[]
}

const phases = ref<PhaseGroup[]>([])
const activePhaseName = ref('')
const phaseManageVisible = ref(false)
const editablePhases = ref<{ name: string }[]>([])

function getPhaseSteps(phaseName: string): StepTemplate[] {
  return phases.value.find(p => p.name === phaseName)?.steps || []
}

function setPhaseSteps(phaseName: string, steps: StepTemplate[]) {
  const phase = phases.value.find(p => p.name === phaseName)
  if (phase) {
    phase.steps = steps
  }
}

// 当前激活阶段的步骤
const activeSteps = computed(() => getPhaseSteps(activePhaseName.value))

// 当前阶段步骤的树形结构
interface StepTreeNode extends StepTemplate {
  children?: StepTreeNode[]
}

const activeStepsTree = computed<StepTreeNode[]>(() => {
  const nodes: StepTreeNode[] = activeSteps.value.map(s => ({ ...s, children: [] }))
  const nodeMap = new Map<number, StepTreeNode>()
  for (const node of nodes) {
    nodeMap.set(node.id, node)
  }
  const roots: StepTreeNode[] = []
  for (const node of nodes) {
    if (node.parent_step_id && nodeMap.has(node.parent_step_id)) {
      const parent = nodeMap.get(node.parent_step_id)!
      parent.children!.push(node)
    } else {
      roots.push(node)
    }
  }
  return roots
})

// 获取所有步骤（扁平化，用于保存）
function getAllSteps(): StepTemplate[] {
  const result: StepTemplate[] = []
  let globalSeq = 0
  for (const phase of phases.value) {
    const orderMap = buildOrderMap(phase.steps)
    for (const node of orderMap.nodes) {
      globalSeq++
      result.push({ ...node, order_index: globalSeq, phase: phase.name })
    }
  }
  return result
}

// 构建树形并返回有序的扁平列表
function buildOrderMap(steps: StepTemplate[]): { nodes: StepTemplate[] } {
  const nodeMap = new Map<number, StepTemplate>()
  const nodes: StepTreeNode[] = steps.map(s => ({ ...s, children: [] }))
  for (const node of nodes) {
    nodeMap.set(node.id, node)
  }
  const roots: StepTreeNode[] = []
  for (const node of nodes) {
    if (node.parent_step_id && nodeMap.has(node.parent_step_id)) {
      nodeMap.get(node.parent_step_id)!.children!.push(node)
    } else {
      roots.push(node)
    }
  }
  const ordered: StepTemplate[] = []
  function traverse(treeNodes: StepTreeNode[]) {
    for (const n of treeNodes) {
      const flat = { ...n } as Record<string, unknown>
      delete flat.children
      ordered.push(flat as unknown as StepTemplate)
      if (n.children && n.children.length > 0) {
        traverse(n.children)
      }
    }
  }
  traverse(roots)
  return { nodes: ordered }
}

// 阶段管理
function openPhaseManageDialog() {
  editablePhases.value = phases.value.map(p => ({ name: p.name }))
  phaseManageVisible.value = true
}

function addPhase() {
  editablePhases.value.push({ name: '' })
}

function removePhase(index: number) {
  const phase = editablePhases.value[index]
  // 如果是已存在的阶段，检查是否有步骤
  if (phases.value.find(p => p.name === phase.name)?.steps.length) {
    ElMessage.warning('该阶段下有步骤，请先移除步骤')
    return
  }
  editablePhases.value.splice(index, 1)
}

function handleSavePhases() {
  // 验证名称
  if (editablePhases.value.some(p => !p.name.trim())) {
    ElMessage.warning('阶段名称不能为空')
    return
  }
  const names = editablePhases.value.map(p => p.name.trim())
  const uniqueNames = new Set(names)
  if (names.length !== uniqueNames.size) {
    ElMessage.warning('阶段名称不能重复')
    return
  }

  // 重建 phases，保留已有步骤
  const newPhases: PhaseGroup[] = []
  const oldPhasesMap = new Map(phases.value.map(p => [p.name, p]))
  for (const name of names) {
    const old = oldPhasesMap.get(name)
    newPhases.push({ name, steps: old ? [...old.steps] : [] })
  }

  // 删除的阶段如果有步骤，合并到第一个阶段
  for (const oldPhase of phases.value) {
    if (!names.includes(oldPhase.name) && oldPhase.steps.length > 0) {
      if (newPhases.length > 0) {
        newPhases[0].steps.push(...oldPhase.steps)
      }
    }
  }

  phases.value = newPhases
  // 如果当前活跃阶段被删除，自动选中第一个
  if (!names.includes(activePhaseName.value) && newPhases.length > 0) {
    activePhaseName.value = newPhases[0].name
  }
  phaseManageVisible.value = false
  ElMessage.success('阶段已保存')
}

// ============ 步骤 CRUD ============

const templateName = ref('')
const selectedStep = ref<StepTemplate | null>(null)
const importVisible = ref(false)
const importTargetPhase = ref('')
const singleAddVisible = ref(false)
const singleStepEditIndex = ref<number | null>(null)
const departmentOptions = ref<string[]>([])

// 表单
const singleStepForm = reactive({
  name: '',
  description: '',
  step_type: 'serial' as StepType,
  timeout_minutes: 5,
  default_assignee_role: 'executor',
  executor_team: '',
  parent_step_id: undefined as number | undefined,
  estimated_duration_minutes: undefined as number | undefined,
  estimated_start_offset: undefined as number | undefined,
  task_name: '',
  sub_task: '',
  responsible_department: '',
  responsible_person: '',
  executor: '',
  reviewer: '',
})

const formParentStepOptions = computed(() => {
  const steps = activeSteps.value
  return steps
    .filter(s => s.id !== (singleStepEditIndex.value !== null ? steps[singleStepEditIndex.value]?.id : undefined))
    .map(s => ({
      value: s.id as number,
      label: `#${s.order_index || ''} ${s.name}`,
    }))
})

// 加载部门
async function loadDepartments() {
  try {
    const depts = await userApi.getDepartments()
    departmentOptions.value = depts
  } catch (error) {
    console.error('Failed to load departments:', error)
  }
}

// 加载模板步骤
async function loadTemplateSteps() {
  try {
    const result = await templateApi.getList({ page: 1, page_size: 1 })
    const template = result.list?.find((t: any) => t.id === templateId.value)
    if (!template) {
      ElMessage.error('模板不存在')
      goBack()
      return
    }
    templateName.value = template.name
    const steps = JSON.parse(JSON.stringify(template.steps || []))
    steps.forEach((s: StepTemplate) => {
      s.description = s.guide_content || s.description || ''
    })

    // 按阶段分组
    const phaseMap = new Map<string, StepTemplate[]>()
    const noPhaseSteps: StepTemplate[] = []
    for (const step of steps) {
      if (step.phase) {
        if (!phaseMap.has(step.phase)) {
          phaseMap.set(step.phase, [])
        }
        phaseMap.get(step.phase)!.push(step)
      } else {
        noPhaseSteps.push(step)
      }
    }

    const newPhases: PhaseGroup[] = []
    for (const [name, stepsList] of phaseMap) {
      newPhases.push({ name, steps: stepsList })
    }
    // 如果没有阶段信息，创建默认阶段
    if (newPhases.length === 0 && noPhaseSteps.length > 0) {
      newPhases.push({ name: '默认阶段', steps: noPhaseSteps })
    }

    phases.value = newPhases
    if (newPhases.length > 0) {
      activePhaseName.value = newPhases[0].name
    }
    selectedStep.value = null
  } catch (error) {
    ElMessage.error('加载步骤失败')
    console.error('Failed to load template steps:', error)
  }
}

// 行选择
function handleRowSelect(row: StepTemplate | undefined) {
  selectedStep.value = row || null
}

// 删除步骤
function removeStepByRow(row: StepTreeNode) {
  const steps = activeSteps.value
  const index = steps.findIndex(s => s.id === row.id)
  if (index >= 0) {
    const removed = steps[index]
    steps.forEach(s => {
      if (s.parent_step_id === removed.id) {
        s.parent_step_id = removed.parent_step_id
      }
    })
    steps.splice(index, 1)
    if (selectedStep.value?.id === removed.id) {
      selectedStep.value = null
    }
  }
}

// 打开添加对话框
function openSingleAddDialog() {
  if (!activePhaseName.value) return
  resetSingleStepForm()
  singleStepEditIndex.value = null
  singleAddVisible.value = true
}

// 重置表单
function resetSingleStepForm() {
  singleStepForm.name = ''
  singleStepForm.description = ''
  singleStepForm.step_type = 'serial'
  singleStepForm.timeout_minutes = 5
  singleStepForm.default_assignee_role = 'executor'
  singleStepForm.executor_team = ''
  singleStepForm.parent_step_id = undefined
  singleStepForm.estimated_duration_minutes = undefined
  singleStepForm.estimated_start_offset = undefined
  singleStepForm.task_name = ''
  singleStepForm.sub_task = ''
  singleStepForm.responsible_department = ''
  singleStepForm.responsible_person = ''
  singleStepForm.executor = ''
  singleStepForm.reviewer = ''
}

// 选中步骤的编辑对话框
function openStepEditDialogForSelected() {
  if (!selectedStep.value) return
  const steps = activeSteps.value
  const index = steps.findIndex(s => s.id === selectedStep.value!.id)
  if (index >= 0) {
    openStepEditDialog(index)
    singleStepForm.parent_step_id = selectedStep.value.parent_step_id
  }
}

// 编辑对话框
function openStepEditDialog(index: number) {
  const step = activeSteps.value[index]
  singleStepForm.name = step.name
  singleStepForm.description = step.description || ''
  singleStepForm.step_type = step.step_type as StepType
  singleStepForm.timeout_minutes = step.timeout_minutes || 5
  singleStepForm.default_assignee_role = step.default_assignee_role || ''
  singleStepForm.executor_team = step.executor_team || ''
  singleStepForm.parent_step_id = step.parent_step_id
  singleStepForm.estimated_duration_minutes = step.estimated_duration_minutes
  singleStepForm.estimated_start_offset = step.estimated_start_offset
  singleStepForm.task_name = step.task_name || ''
  singleStepForm.sub_task = step.sub_task || ''
  singleStepForm.responsible_department = step.responsible_department || ''
  singleStepForm.responsible_person = step.responsible_person || ''
  singleStepForm.executor = step.executor || ''
  singleStepForm.reviewer = step.reviewer || ''
  singleStepEditIndex.value = index
  singleAddVisible.value = true
}

// 编辑步骤（调后端 API 立即生效）
async function handleEditStep() {
  if (!singleStepForm.name.trim()) {
    ElMessage.warning('请输入步骤名称')
    return
  }

  const steps = activeSteps.value
  const step = steps[singleStepEditIndex.value!]

  try {
    const seqVal = (step as any).seq || step.order_index || 1
    await templateApi.updateStep(templateId.value, step.id, {
      name: singleStepForm.name.trim(),
      seq: seqVal,
      step_type: singleStepForm.step_type,
      timeout_minutes: Math.max(5, (singleStepForm.estimated_duration_minutes || 5) * 2),
      guide_content: singleStepForm.description.trim(),
      default_assignee_role: singleStepForm.default_assignee_role,
      executor_team: singleStepForm.executor_team,
      phase: step.phase || activePhaseName.value,
      phase_step: step.phase_step,
      estimated_duration_minutes: singleStepForm.estimated_duration_minutes,
      estimated_start_offset: singleStepForm.estimated_start_offset,
      task_name: singleStepForm.task_name,
      sub_task: singleStepForm.sub_task,
      responsible_department: singleStepForm.responsible_department,
      responsible_person: singleStepForm.responsible_person,
      executor: singleStepForm.executor,
      reviewer: singleStepForm.reviewer,
    })

    // 更新本地状态
    step.name = singleStepForm.name.trim()
    step.description = singleStepForm.description.trim()
    step.step_type = singleStepForm.step_type as StepType
    step.timeout_minutes = Math.max(5, (singleStepForm.estimated_duration_minutes || 5) * 2)
    step.default_assignee_role = singleStepForm.default_assignee_role
    step.executor_team = singleStepForm.executor_team
    step.estimated_duration_minutes = singleStepForm.estimated_duration_minutes
    step.estimated_start_offset = singleStepForm.estimated_start_offset
    step.task_name = singleStepForm.task_name
    step.sub_task = singleStepForm.sub_task
    step.responsible_department = singleStepForm.responsible_department
    step.responsible_person = singleStepForm.responsible_person
    step.executor = singleStepForm.executor
    step.reviewer = singleStepForm.reviewer

    // 同步刷新右侧详情面板
    if (selectedStep.value?.id === step.id) {
      selectedStep.value = { ...step }
    }

    ElMessage.success('步骤已更新')
  } catch (error) {
    ElMessage.error('保存步骤失败')
    console.error('Save step error:', error)
    return
  }

  resetSingleStepForm()
  singleAddVisible.value = false
}

// 添加步骤
function handleAddSingleStep() {
  if (!singleStepForm.name.trim()) {
    ElMessage.warning('请输入步骤名称')
    return
  }

  const steps = activeSteps.value

  if (singleStepEditIndex.value !== null) {
    handleEditStep()
    return
  }

  steps.push({
      id: Date.now(),
      template_id: templateId.value,
      name: singleStepForm.name.trim(),
      description: singleStepForm.description.trim(),
      step_type: singleStepForm.step_type as StepType,
      timeout_minutes: Math.max(5, (singleStepForm.estimated_duration_minutes || 5) * 2),
      default_assignee_role: singleStepForm.default_assignee_role,
      executor_team: singleStepForm.executor_team,
      parent_step_id: singleStepForm.parent_step_id,
      order_index: steps.length + 1,
      created_at: new Date().toISOString(),
      estimated_duration_minutes: singleStepForm.estimated_duration_minutes,
      estimated_start_offset: singleStepForm.estimated_start_offset,
      task_name: singleStepForm.task_name,
      sub_task: singleStepForm.sub_task,
      responsible_department: singleStepForm.responsible_department,
      responsible_person: singleStepForm.responsible_person,
      executor: singleStepForm.executor,
      reviewer: singleStepForm.reviewer,
    })
    ElMessage.success('步骤已添加')

  resetSingleStepForm()
  singleAddVisible.value = false
}

// ============ 保存 ============

async function handleSaveSteps() {
  const allSteps = getAllSteps()
  try {
    await templateApi.updateSteps(templateId.value, allSteps.map(s => ({
      name: s.name,
      seq: s.order_index,
      parent_step_id: s.parent_step_id,
      step_type: s.step_type,
      timeout_minutes: Math.max(5, (s.estimated_duration_minutes || 5) * 2),
      guide_content: s.description || s.guide_content || '',
      default_assignee_role: s.default_assignee_role || '',
      executor_team: s.executor_team || '',
      phase: s.phase || '',
      phase_step: s.phase_step || '',
      estimated_duration_minutes: s.estimated_duration_minutes,
      estimated_start_offset: s.estimated_start_offset,
      task_name: s.task_name || '',
      sub_task: s.sub_task || '',
      responsible_department: s.responsible_department || '',
      responsible_person: s.responsible_person || '',
      executor: s.executor || '',
      reviewer: s.reviewer || '',
    })))
    ElMessage.success('步骤已保存')
    goBack()
  } catch (error) {
    ElMessage.error('保存步骤失败')
    console.error('Save steps error:', error)
  }
}

// ============ 工具函数 ============

function getStepTypeLabel(type: string): string {
  const map: Record<string, string> = {
    serial: '串行',
    parallel: '并行',
  }
  return map[type] || type
}

function goBack() {
  router.push({ name: 'DirectorTemplates' })
}

// ============ Excel 导入 ============

function openBatchImportDialog() {
  if (phases.value.length === 0) {
    ElMessage.warning('请先添加阶段')
    return
  }
  importTargetPhase.value = activePhaseName.value
  importVisible.value = true
}

function downloadTemplate() {
  const header = ['步骤名称', '描述', '步骤类型', '执行角色', '执行团队', '预计耗时 (分)', '责任部门', '责任人', '执行人', '复核人', '任务名称', '子任务描述', '说明']
  const data = [
    header,
    ['检查数据库状态', '检查主库是否正常运行', 'serial', 'executor', '技术部', '10', '技术部', '张三', '李四', '王五', '数据库状态检查', '确认主库正常运行，检查连接池状态', '步骤类型可选值：serial(串行), parallel(并行)'],
    ['切换从库', '将从库提升为主库', 'parallel', 'director', '运维部', '15', '运维部', '赵六', '钱七', '孙八', '主从切换', '停止主库写入，提升从库为主库', '超时时间 = 预计耗时 * 2 (最小 5 分钟)'],
  ]
  const wb = XLSX.utils.book_new()
  const ws = XLSX.utils.aoa_to_sheet(data)
  const colWidths = [
    { wch: 20 }, { wch: 40 }, { wch: 12 }, { wch: 12 }, { wch: 15 },
    { wch: 12 }, { wch: 15 }, { wch: 12 }, { wch: 12 }, { wch: 12 },
    { wch: 20 }, { wch: 40 }, { wch: 50 }
  ]
  ws['!cols'] = colWidths
  XLSX.utils.book_append_sheet(wb, ws, '步骤导入')
  XLSX.writeFile(wb, `步骤导入模板_${templateName.value}_${importTargetPhase.value || '当前阶段'}.xlsx`)
  ElMessage.success('模板已下载')
}

function handleExcelUpload(file: File) {
  if (!importTargetPhase.value) {
    ElMessage.warning('请先选择目标阶段')
    return false
  }

  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      const data = new Uint8Array(e.target?.result as ArrayBuffer)
      const workbook = XLSX.read(data, { type: 'array', cellDates: true })
      const sheetName = workbook.SheetNames[0]
      const sheet = workbook.Sheets[sheetName]
      const rows = XLSX.utils.sheet_to_json(sheet, { header: 1 }) as any[][]

      if (rows.length < 2) {
        ElMessage.warning('Excel 文件内容为空')
        return false
      }

      const steps: StepTemplate[] = []
      const errors: string[] = []
      const targetSteps = getPhaseSteps(importTargetPhase.value)

      // 新表头（无阶段列）
      for (let i = 1; i < rows.length; i++) {
        const row = rows[i]
        const rowNum = i + 1
        const name = String(row[0] || '').trim()
        const description = String(row[1] || '').trim()
        const stepTypeRaw = String(row[2] || '').trim()
        const assigneeRoleRaw = String(row[3] || '').trim()
        const executorTeam = String(row[4] || '').trim()
        const estimatedDuration = parseInt(String(row[5] || '')) || undefined
        const responsibleDepartment = String(row[6] || '').trim()
        const responsiblePerson = String(row[7] || '').trim()
        const executor = String(row[8] || '').trim()
        const reviewer = String(row[9] || '').trim()
        const taskName = String(row[10] || '').trim()
        const subTask = String(row[11] || '').trim()

        if (!name) {
          errors.push(`第${rowNum}行：步骤名称不能为空`)
          continue
        }

        const stepTypeMap: Record<string, string> = {
          '串行': 'serial', '并行': 'parallel',
          'serial': 'serial', 'parallel': 'parallel',
        }
        const stepType = stepTypeMap[stepTypeRaw] || 'serial'

        const assigneeRoleMap: Record<string, string> = {
          '指挥组': 'director', '执行组': 'executor',
          'director': 'director', 'executor': 'executor',
        }
        const assigneeRole = assigneeRoleMap[assigneeRoleRaw.toLowerCase()] || 'executor'

        steps.push({
          id: Date.now() + Math.random(),
          template_id: templateId.value,
          name,
          description,
          step_type: stepType as any,
          timeout_minutes: Math.max(5, Math.max(5, estimatedDuration || 5) * 2),
          default_assignee_role: assigneeRole,
          executor_team: executorTeam,
          order_index: targetSteps.length + steps.length + 1,
          created_at: new Date().toISOString(),
          phase: importTargetPhase.value,
          estimated_duration_minutes: estimatedDuration,
          estimated_start_offset: undefined,
          task_name: taskName,
          sub_task: subTask,
          responsible_department: responsibleDepartment,
          responsible_person: responsiblePerson,
          executor,
          reviewer,
        })
      }

      if (errors.length > 0) {
        ElMessage.warning(errors.join('\n'))
        return false
      }

      if (steps.length > 0) {
        setPhaseSteps(importTargetPhase.value, [...targetSteps, ...steps])
        ElMessage.success(`成功导入 ${steps.length} 个步骤到「${importTargetPhase.value}」阶段`)
        importVisible.value = false
        // 自动切换到导入的阶段
        activePhaseName.value = importTargetPhase.value
      } else if (errors.length === 0) {
        ElMessage.warning('未找到有效数据')
        return false
      }
    } catch {
      ElMessage.error('Excel 文件解析失败')
    }
  }
  reader.readAsArrayBuffer(file)
  return false
}

onMounted(() => {
  loadTemplateSteps()
  loadDepartments()
})
</script>

<style scoped lang="scss">
@use '@/styles/layout' as *;
@use '@/styles/variables' as *;

.page-container {
  @include page-container;
}

.page-header {
  @include page-header;
  display: flex;
  justify-content: space-between;
  align-items: center;

  .header-left {
    display: flex;
    align-items: center;
    gap: $spacing-sm;

    .page-title {
      margin: 0;
      font-size: $font-size-lg;
      font-weight: 600;
      color: $text-primary;
    }
  }

  .header-actions {
    display: flex;
    gap: $spacing-sm;
  }
}

.page-content {
  @include page-content;
}

.steps-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: $spacing-base;
  min-height: calc(100vh - 180px);
}

.steps-panel,
.detail-panel {
  display: flex;
  flex-direction: column;
  background: $bg-secondary;
  border-radius: $radius-base;
  overflow: hidden;

  .phase-tabs {
    padding: 0;
  }

  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: $spacing-sm $spacing-base;
    background: $bg-tertiary;
    border-bottom: 1px solid $border-color;

    h3 {
      margin: 0;
      font-size: $font-size-base;
      font-weight: 600;
      color: $text-primary;
    }

    .panel-actions {
      display: flex;
      gap: $spacing-xs;
    }
  }

  .panel-body {
    flex: 1;
    padding: $spacing-base;
    overflow-y: auto;
  }
}

.steps-panel .panel-body {
  :deep(.el-table) {
    --el-table-border-color: $border-color;
  }
}

.empty-steps {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 300px;
}

.step-detail {
  :deep(.el-descriptions) {
    background: transparent;
  }
}

.empty-detail {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 300px;
}

.step-selected {
  color: var(--el-color-primary);
  font-weight: 600;
}

.single-step-form {
  .form-row {
    display: flex;
    gap: 16px;
  }

  .inline-form-item {
    flex: 1;

    .unit-label {
      margin-left: 8px;
      color: var(--el-text-color-secondary);
      font-size: 12px;
    }
  }
}

.phase-manage-list {
  .phase-manage-item {
    display: flex;
    gap: $spacing-sm;
    align-items: center;
    margin-bottom: $spacing-sm;
  }
}

.import-form {
  margin-bottom: $spacing-base;
}

.import-empty-hint {
  padding: 20px 0;
}

.excel-upload {
  .upload-area {
    :deep(.el-upload) {
      width: 100%;
    }
  }

  .upload-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px 20px;
    border: 2px dashed var(--el-border-color);
    border-radius: $radius-base;
    cursor: pointer;
    transition: border-color 0.2s;

    &:hover {
      border-color: var(--el-color-primary);
    }

    .upload-icon {
      font-size: 48px;
      color: var(--el-text-color-secondary);
      margin-bottom: 12px;
    }

    .upload-text {
      font-size: 16px;
      color: var(--el-text-color-primary);
      margin-bottom: 8px;
    }

    .upload-hint {
      font-size: 12px;
      color: var(--el-text-color-secondary);
    }
  }

  .template-download {
    margin-top: 16px;
    text-align: center;
  }
}
</style>
