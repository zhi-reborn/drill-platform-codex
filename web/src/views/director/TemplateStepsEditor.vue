<template>
  <div class="page-container">
    <div class="page-header">
      <div class="header-left">
        <el-button text @click="goBack">
          <el-icon>
            <ArrowLeft />
          </el-icon>
          返回模板列表
        </el-button>
        <h2 class="page-title">编辑步骤 - {{ templateName }}</h2>
      </div>
      <div class="header-actions">
        <el-button @click="loadTemplateSteps">
          <el-icon>
            <Refresh />
          </el-icon>
          重新加载
        </el-button>
        <el-button @click="openPhaseManageDialog">
          <el-icon>
            <Setting />
          </el-icon>
          阶段管理
        </el-button>
        <el-button type="success" @click="openBatchImportDialog">
          <el-icon>
            <Download />
          </el-icon>
          批量导入
        </el-button>
        <el-button type="warning" @click="exportSteps">
          <el-icon>
            <Upload />
          </el-icon>
          导出步骤
        </el-button>
        <el-button type="primary" @click="handleSaveSteps">
          <el-icon>
            <Check />
          </el-icon>
          保存步骤
        </el-button>
      </div>
    </div>

    <div class="page-content steps-layout">
      <!-- 左侧：阶段 tab + 步骤树 -->
      <div class="steps-panel">
        <!-- 阶段 tabs -->
        <el-tabs v-model="activePhaseName" class="phase-tabs" type="card">
          <el-tab-pane v-for="phase in phases" :key="phase.name" :name="phase.name" :label="phase.name" />
          <el-tab-pane v-if="phases.length === 0" name="_empty" label="无阶段" disabled />
        </el-tabs>

        <div class="panel-header">
          <h3>步骤树</h3>
          <el-button type="primary" size="small" @click="openSingleAddDialog" :disabled="!activePhaseName">
            <el-icon>
              <Plus />
            </el-icon>
            添加步骤
          </el-button>
        </div>
        <div class="panel-body">
          <div v-if="activeSteps.length > 0">
            <draggable v-model="rootStepList" :animation="200" item-key="id" handle=".drag-handle"
              ghost-class="drag-ghost" class="draggable-list" @start="dragging = true" @end="onDragEnd">
              <template #item="{ element: root }">
                <div class="step-group">
                  <!-- 根步骤行 -->
                  <div class="step-row" :class="{ 'step-selected': selectedStep?.id === root.id }"
                    @click="handleStepSelect(root)">
                    <span class="drag-handle">⠿</span>
                    <span class="seq-badge">{{ computeSEQ(root) }}</span>
                    <button v-if="root.hasChildren" class="expand-btn" @click.stop="toggleCollapse(root.id)">
                      {{ isCollapsed(root.id) ? '▶' : '▼' }}
                    </button>
                    <span v-else class="expand-placeholder"></span>
                    <span class="step-name">{{ root.name || '-' }}</span>
                    <el-tag size="small" type="info">{{ getStepTypeLabel(root.step_type) }}</el-tag>
                    <el-button text type="danger" size="small" @click.stop="removeStepByRow(root)" title="删除">
                      <el-icon>
                        <Delete />
                      </el-icon>
                    </el-button>
                  </div>
                  <!-- 子步骤行 -->
                  <template v-if="!isCollapsed(root.id)">
                    <draggable v-model="root.children" :animation="200" item-key="id" handle=".drag-handle"
                      ghost-class="drag-ghost" :group="{ name: 'children-group', pull: false, put: false }"
                      class="children-list" @start="dragging = true" @end="onDragEnd">
                      <template #item="{ element: child }">
                        <div class="step-row child-row" :class="{ 'step-selected': selectedStep?.id === child.id }"
                          @click="handleStepSelect(child)">
                          <span class="drag-handle">⠿</span>
                          <span class="seq-badge">{{ computeSEQ(child) }}</span>
                          <button v-if="child.hasChildren" class="expand-btn" @click.stop="toggleCollapse(child.id)">
                            {{ isCollapsed(child.id) ? '▶' : '▼' }}
                          </button>
                          <span v-else class="expand-placeholder"></span>
                          <span class="step-name">{{ child.name || '-' }}</span>
                          <el-tag size="small" type="info">{{ getStepTypeLabel(child.step_type) }}</el-tag>
                          <el-button text type="danger" size="small" @click.stop="removeStepByRow(child)" title="删除">
                            <el-icon>
                              <Delete />
                            </el-icon>
                          </el-button>
                        </div>
                      </template>
                    </draggable>
                  </template>
                </div>
              </template>
            </draggable>
          </div>
          <div v-else class="empty-steps">
            <el-empty :description="phases.length === 0 ? '请先添加阶段' : '暂无步骤，请添加或导入步骤'" :image-size="100">
              <el-button v-if="activePhaseName" type="primary" @click="openSingleAddDialog">
                <el-icon>
                  <Plus />
                </el-icon>
                添加步骤
              </el-button>
              <el-button v-if="activePhaseName" type="success" @click="openBatchImportDialog" style="margin-left: 8px">
                <el-icon>
                  <Download />
                </el-icon>
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
              <el-icon>
                <Edit />
              </el-icon>
              编辑
            </el-button>
          </div>
        </div>
        <div class="panel-body">
          <div v-if="selectedStep" class="step-detail">
            <el-descriptions :column="2" border size="default">
              <el-descriptions-item label="步骤名称" :span="2">{{ selectedStep.name }}</el-descriptions-item>
              <el-descriptions-item label="描述" :span="2">{{ selectedStep.description || '-' }}</el-descriptions-item>
              <el-descriptions-item label="父步骤序号">{{ selectedStep.parent_seq_display || '-' }}</el-descriptions-item>
              <el-descriptions-item label="步骤类型">{{ getStepTypeLabel(selectedStep.step_type) }}</el-descriptions-item>
              <el-descriptions-item label="环节">{{ selectedStep.phase_step || '-' }}</el-descriptions-item>
              <el-descriptions-item label="预计耗时">{{ selectedStep.estimated_duration_minutes ?
                selectedStep.estimated_duration_minutes + ' 分钟' : '-' }}</el-descriptions-item>
              <el-descriptions-item label="开始偏移">{{ selectedStep.estimated_start_offset ?
                selectedStep.estimated_start_offset
                + ' 分钟' : '-' }}</el-descriptions-item>
            </el-descriptions>
            <el-divider>责任与扩展</el-divider>
            <el-descriptions :column="2" border size="default">
              <el-descriptions-item label="执行角色">{{ selectedStep.default_assignee_role ?
                (selectedStep.default_assignee_role
                  === 'director' ? '指挥组' : '执行组') : '-' }}</el-descriptions-item>
              <el-descriptions-item label="执行团队">{{ selectedStep.executor_team || '-' }}</el-descriptions-item>
              <el-descriptions-item label="责任部门">{{ selectedStep.attributes?.responsible_department || '-'
                }}</el-descriptions-item>
              <el-descriptions-item label="配合部门">{{ selectedStep.attributes?.cooperating_department || '-'
                }}</el-descriptions-item>
              <el-descriptions-item label="责任团队">{{ selectedStep.attributes?.responsible_team || '-'
                }}</el-descriptions-item>
              <el-descriptions-item label="操作人">{{ selectedStep.attributes?.operator || '-' }}</el-descriptions-item>
              <el-descriptions-item label="复核人">{{ selectedStep.attributes?.reviewer || '-' }}</el-descriptions-item>
            </el-descriptions>
            <el-divider>说明与预案</el-divider>
            <el-descriptions :column="1" border size="default">
              <el-descriptions-item label="操作说明">{{ selectedStep.attributes?.operation_guide || '-'
                }}</el-descriptions-item>
              <el-descriptions-item label="验证方式">{{ selectedStep.attributes?.verification_method || '-'
                }}</el-descriptions-item>
              <el-descriptions-item label="最坏影响分析">{{ selectedStep.attributes?.worst_case_analysis || '-'
                }}</el-descriptions-item>
              <el-descriptions-item label="兜底措施">{{ selectedStep.attributes?.fallback_measures || '-'
                }}</el-descriptions-item>
              <el-descriptions-item label="备注">{{ selectedStep.attributes?.remark || '-' }}</el-descriptions-item>
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
            <el-icon>
              <Delete />
            </el-icon>
          </el-button>
        </div>
      </div>
      <el-button type="primary" plain style="width: 100%; margin-top: 12px" @click="addPhase">
        <el-icon>
          <Plus />
        </el-icon>
        添加阶段
      </el-button>
      <template #footer>
        <el-button @click="phaseManageVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSavePhases">保存</el-button>
      </template>
    </el-dialog>

    <!-- 批量导入对话框 -->
    <el-dialog v-model="importVisible" title="批量导入步骤" width="520px">
      <div class="excel-upload">
        <el-upload ref="uploadRef" :before-upload="handleExcelUpload" :show-file-list="false" accept=".xlsx,.xls"
          class="upload-area">
          <div class="upload-content">
            <el-icon class="upload-icon">
              <Upload />
            </el-icon>
            <div class="upload-text">点击或拖拽上传 Excel 文件</div>
            <div class="upload-hint">支持 .xlsx, .xls 格式</div>
          </div>
        </el-upload>
        <div class="template-download">
          <el-button type="info" plain @click="downloadTemplate">
            <el-icon>
              <Download />
            </el-icon>
            下载 Excel 模板
          </el-button>
        </div>
      </div>
    </el-dialog>

    <!-- 单个添加/编辑抽屉 -->
    <el-drawer v-model="singleAddVisible" :title="singleStepEditIndex !== null ? '编辑步骤' : '添加步骤'" size="720px">
      <el-form :model="singleStepForm" label-width="90px" class="single-step-form">
        <el-form-item label="步骤名称" required>
          <el-input v-model="singleStepForm.name" placeholder="请输入步骤名称" maxlength="100" show-word-limit />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="singleStepForm.description" type="textarea" placeholder="步骤描述" :rows="2" maxlength="500"
            show-word-limit />
        </el-form-item>
        <div class="form-row">
          <el-form-item label="父步骤" class="inline-form-item">
            <el-select v-model="singleStepForm.parent_step_id" clearable placeholder="可选" filterable>
              <el-option v-for="opt in formParentStepOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
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
          <el-form-item label="环节" class="inline-form-item">
            <el-input v-model="singleStepForm.phase_step" placeholder="如：初始化、主流程" clearable />
          </el-form-item>
          <el-form-item label="开始偏移" class="inline-form-item">
            <el-input-number v-model="singleStepForm.estimated_start_offset" :min="0" controls-position="right"
              placeholder="相对启动偏移" />
            <span class="unit-label">分钟</span>
          </el-form-item>
        </div>
        <div class="form-row">
          <el-form-item label="预计耗时" class="inline-form-item">
            <el-input-number v-model="singleStepForm.estimated_duration_minutes" :min="1" :max="1440"
              controls-position="right" placeholder="可选" />
            <span class="unit-label">分钟</span>
          </el-form-item>
          <el-form-item label="超时时间" class="inline-form-item">
            <el-input-number v-model="singleStepForm.timeout_minutes" :min="1" :max="1440" controls-position="right"
              placeholder="5" />
            <span class="unit-label">分钟</span>
          </el-form-item>
        </div>

        <el-divider>执行权限</el-divider>
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

        <el-divider>责任信息</el-divider>
        <div class="form-row">
          <el-form-item label="责任部门" class="inline-form-item">
            <el-input v-model="singleStepForm.attributes.responsible_department" placeholder="责任部门" clearable />
          </el-form-item>
          <el-form-item label="配合部门" class="inline-form-item">
            <el-input v-model="singleStepForm.attributes.cooperating_department" placeholder="配合部门" clearable />
          </el-form-item>
        </div>
        <div class="form-row">
          <el-form-item label="责任团队" class="inline-form-item">
            <el-input v-model="singleStepForm.attributes.responsible_team" placeholder="责任团队" clearable />
          </el-form-item>
          <el-form-item label="操作人" class="inline-form-item">
            <el-input v-model="singleStepForm.attributes.operator" placeholder="操作人姓名" clearable />
          </el-form-item>
        </div>
        <el-form-item label="复核人">
          <el-input v-model="singleStepForm.attributes.reviewer" placeholder="复核人姓名" clearable />
        </el-form-item>

        <el-divider>说明与预案</el-divider>
        <el-form-item label="操作说明">
          <el-input v-model="singleStepForm.attributes.operation_guide" type="textarea" placeholder="操作步骤详细说明" :rows="3"
            maxlength="2000" show-word-limit />
        </el-form-item>
        <el-form-item label="验证方式">
          <el-input v-model="singleStepForm.attributes.verification_method" type="textarea" placeholder="如何验证操作成功"
            :rows="2" maxlength="1000" show-word-limit />
        </el-form-item>
        <el-form-item label="最坏影响分析">
          <el-input v-model="singleStepForm.attributes.worst_case_analysis" type="textarea" placeholder="操作失败的最大影响"
            :rows="2" maxlength="1000" show-word-limit />
        </el-form-item>
        <el-form-item label="兜底措施">
          <el-input v-model="singleStepForm.attributes.fallback_measures" type="textarea" placeholder="失败后如何恢复"
            :rows="2" maxlength="1000" show-word-limit />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="singleStepForm.attributes.remark" placeholder="其他备注信息" clearable />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="singleAddVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAddSingleStep">{{ singleStepEditIndex !== null ? '保存修改' : '添加步骤'
          }}</el-button>
      </template>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh, Plus, Delete, Upload, Download, Check, Edit, ArrowLeft, Setting } from '@element-plus/icons-vue'
import draggable from 'vuedraggable'
import * as XLSX from 'xlsx'
import { userApi } from '@/api'
import { templateApi } from '@/api/modules/template'
import type { StepTemplate, StepType, StepAttributes } from '@/types'

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

// ============ 折叠状态管理（localStorage） ============

const storageKey = computed(() => `drill-steps-collapse-${templateId.value}`)

function loadCollapsed(): Set<number> {
  try {
    const raw = localStorage.getItem(storageKey.value)
    return new Set(raw ? JSON.parse(raw) : [])
  } catch { return new Set() }
}

const collapsedIds = ref<Set<number>>(loadCollapsed())

watch(
  collapsedIds,
  () => localStorage.setItem(storageKey.value, JSON.stringify([...collapsedIds.value])),
  { deep: true }
)

function isCollapsed(id: number): boolean { return collapsedIds.value.has(id) }
function toggleCollapse(id: number) {
  const next = new Set(collapsedIds.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  collapsedIds.value = next
}

// ============ 树形结构 ============

interface StepTreeNode extends StepTemplate {
  children?: StepTreeNode[]
  hasChildren?: boolean
}

const rootStepList = ref<StepTreeNode[]>([])
let _seqCounter = 0

function buildTreeFromSteps(steps: StepTemplate[]): StepTreeNode[] {
  const nodes: StepTreeNode[] = steps.map(s => ({ ...s, children: [], hasChildren: false }))
  const nodeMap = new Map<number, StepTreeNode>()
  for (const node of nodes) {
    nodeMap.set(node.id, node)
  }
  const roots: StepTreeNode[] = []
  for (const node of nodes) {
    // 兼容 parent_step_id 为 0 或 null 的情况
    const parentId = node.parent_step_id
    if (parentId && parentId > 0 && nodeMap.has(parentId)) {
      const parent = nodeMap.get(parentId)!
      parent.children!.push(node)
      parent.hasChildren = true
    } else {
      roots.push(node)
    }
  }
  return roots
}

function recomputeSEQ(tree: StepTreeNode[]): number {
  _seqCounter = 0
  function walk(nodes: StepTreeNode[]) {
    for (const node of nodes) {
      _seqCounter++
      node.order_index = _seqCounter
      const flat = getPhaseSteps(activePhaseName.value)?.find(s => s.id === node.id)
      if (flat) flat.order_index = _seqCounter
      if (node.children?.length) {
        walk(node.children)
      }
    }
  }
  walk(tree)
  return _seqCounter
}

function syncTreeToFlatList(tree: StepTreeNode[]) {
  const phase = phases.value.find(p => p.name === activePhaseName.value)
  if (!phase) return
  const newSteps: StepTemplate[] = []
  function flatten(nodes: StepTreeNode[]) {
    for (const node of nodes) {
      const { children, hasChildren, ...flat } = node
      newSteps.push(flat as unknown as StepTemplate)
      if (node.children?.length) flatten(node.children)
    }
  }
  flatten(tree)
  // 替换原始组
  phase.steps = newSteps
}

function computeSEQ(step: StepTreeNode): number {
  return step.order_index ?? 0
}

function buildAndSyncTree() {
  const steps = getPhaseSteps(activePhaseName.value)
  if (!steps || steps.length === 0) {
    rootStepList.value = []
    return
  }
  const tree = buildTreeFromSteps(steps)
  recomputeSEQ(tree)
  rootStepList.value = tree
  // 强制 Vue 响应式更新
  nextTick(() => {
    rootStepList.value = [...rootStepList.value]
  })
}

// 监听 activePhaseName 变化，重建根步骤列表
watch(activePhaseName, buildAndSyncTree, { flush: 'sync' })

// 监听阶段数量变化
watch(
  () => phases.value.map(p => p.name).join('|'),
  buildAndSyncTree,
  { flush: 'sync' }
)

// 构建当前树形结构的步骤序号映射（ID -> order_index）
const stepSeqMap = computed(() => {
  const map: Record<number, number> = {}
  const traverse = (nodes: StepTreeNode[]) => {
    for (const n of nodes) {
      if (n.id !== undefined) map[n.id as number] = n.order_index ?? 0
      if (n.children?.length) traverse(n.children)
    }
  }
  traverse(rootStepList.value)
  return map
})

// ============ 拖拽排序 ============

const dragging = ref(false)

function handleStepSelect(step: StepTemplate) {
  let parentSeq = '-'
  if (step.parent_step_id) {
    const seq = stepSeqMap.value[step.parent_step_id]
    if (seq && seq > 0) {
      parentSeq = `#${seq}`
    } else {
      const parent = activeSteps.value.find(s => s.id === step.parent_step_id)
      if (parent?.order_index) {
        parentSeq = `#${parent.order_index}`
      } else if (step.parent_step_id) {
        parentSeq = String(step.parent_step_id)
      }
    }
  }
  selectedStep.value = {
    ...step,
    attributes: { ...((step as any).attributes || {}) },
    parent_seq_display: parentSeq
  }
}

function onDragEnd() {
  recomputeSEQ(rootStepList.value)
  syncTreeToFlatList(rootStepList.value)
  dragging.value = false
}

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
  const nodeMap = new Map<number, StepTreeNode>()
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
const selectedStep = ref<(StepTemplate & { parent_seq_display?: string }) | null>(null)
const importVisible = ref(false)
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
  phase_step: '' as string,
  attributes: {} as StepAttributes,
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
    const template = await templateApi.getById(templateId.value)
    if (!template) {
      ElMessage.error('模板不存在')
      goBack()
      return
    }
    templateName.value = template.name
    const steps = JSON.parse(JSON.stringify(template.steps || []))
    steps.forEach((s: StepTemplate) => {
      s.description = s.guide_content || s.description || ''
      // 确保 attributes 是对象（兼容后端返回的 JSON 字符串）
      if (s.attributes && typeof s.attributes === 'string') {
        try { s.attributes = JSON.parse(s.attributes) } catch { s.attributes = {} }
      } else if (!s.attributes) {
        s.attributes = {}
      }
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
    // 在 phases 和 activePhaseName 都设置完后强制重建
    buildAndSyncTree()
    selectedStep.value = null
  } catch (error) {
    ElMessage.error('加载步骤失败')
    console.error('Failed to load template steps:', error)
  }
}

// 行选择（保留兼容旧代码引用）
function handleRowSelect(_row: StepTemplate | undefined) { }

// 删除步骤
function removeStepByRow(row: StepTreeNode) {
  const steps = getPhaseSteps(activePhaseName.value)
  if (!steps) return
  const index = steps.findIndex(s => s.id === row.id)
  if (index >= 0) {
    const removed = steps[index]
    // 子步骤提升为父步骤的子步骤
    steps.forEach(s => {
      if (s.parent_step_id === removed.id) {
        s.parent_step_id = removed.parent_step_id
      }
    })
    steps.splice(index, 1)
    if (selectedStep.value?.id === removed.id) {
      selectedStep.value = null
    }
    // 同步更新树形结构
    buildAndSyncTree()
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
  singleStepForm.phase_step = ''
  singleStepForm.attributes = {}
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
  singleStepForm.description = step.description || step.guide_content || ''
  singleStepForm.step_type = step.step_type as StepType
  singleStepForm.timeout_minutes = step.timeout_minutes || 5
  singleStepForm.default_assignee_role = step.default_assignee_role || ''
  singleStepForm.executor_team = step.executor_team || ''
  singleStepForm.parent_step_id = step.parent_step_id
  singleStepForm.estimated_duration_minutes = step.estimated_duration_minutes
  singleStepForm.estimated_start_offset = step.estimated_start_offset
  singleStepForm.phase_step = step.phase_step || ''
  singleStepForm.attributes = parseAttributes(step.attributes)
  singleStepEditIndex.value = index
  singleAddVisible.value = true
}

function parseAttributes(val: unknown): Record<string, string | undefined> {
  if (!val) return {}
  if (typeof val === 'string') {
    try { return JSON.parse(val) } catch { return {} }
  }
  // 返回浅拷贝，防止编辑表单时直接影响原数据
  return { ...(val as Record<string, string | undefined>) }
}

// 编辑步骤（调后端 API 立即生效）
async function handleEditStep() {
  if (!singleStepForm.name.trim()) {
    ElMessage.warning('请输入步骤名称')
    return
  }

  const steps = activeSteps.value
  const step = steps[singleStepEditIndex.value!]

  // 更新本地状态
  step.name = singleStepForm.name.trim()
  step.description = singleStepForm.description.trim()
  step.step_type = singleStepForm.step_type as StepType
  step.order_index = singleStepEditIndex.value! + 1
  step.timeout_minutes = singleStepForm.timeout_minutes
  step.default_assignee_role = singleStepForm.default_assignee_role
  step.executor_team = singleStepForm.executor_team
  step.parent_step_id = singleStepForm.parent_step_id
  step.estimated_duration_minutes = singleStepForm.estimated_duration_minutes
  step.estimated_start_offset = singleStepForm.estimated_start_offset
  step.phase_step = singleStepForm.phase_step || ''
  step.attributes = { ...singleStepForm.attributes }

  // 仅已保存步骤（真实数据库 ID）调用 API
  const realId = step.id && Number.isInteger(step.id)
  if (realId) {
    try {
      const seqVal = singleStepEditIndex.value! + 1
      await templateApi.updateStep(templateId.value, step.id, {
        name: singleStepForm.name.trim(),
        seq: seqVal,
        step_type: singleStepForm.step_type,
        timeout_minutes: singleStepForm.timeout_minutes,
        guide_content: singleStepForm.description.trim(),
        default_assignee_role: singleStepForm.default_assignee_role,
        executor_team: singleStepForm.executor_team,
        phase: step.phase || activePhaseName.value,
        phase_step: singleStepForm.phase_step || '',
        estimated_duration_minutes: singleStepForm.estimated_duration_minutes,
        estimated_start_offset: singleStepForm.estimated_start_offset,
        attributes: JSON.stringify(singleStepForm.attributes),
      })
      ElMessage.success('步骤已更新')
    } catch (error) {
      ElMessage.error('保存步骤失败')
      console.error('Save step error:', error)
      return
    }
  } else {
    const isNew = singleStepEditIndex.value === steps.length - 1 || !step.order_index
    ElMessage.success(isNew ? '步骤已添加' : '步骤已更新')
  }

  // 强制刷新右侧详情面板（直接赋新值确保视图更新）
  let parentSeq = '-'
  if (step.parent_step_id) {
    const seq = stepSeqMap.value[step.parent_step_id]
    if (seq && seq > 0) {
      parentSeq = `#${seq}`
    } else {
      const parent = activeSteps.value.find(s => s.id === step.parent_step_id)
      if (parent?.order_index) {
        parentSeq = `#${parent.order_index}`
      } else if (step.parent_step_id) {
        parentSeq = String(step.parent_step_id)
      }
    }
  }
  selectedStep.value = {
    ...step,
    attributes: { ...step.attributes },
    parent_seq_display: parentSeq
  }

  resetSingleStepForm()
  singleAddVisible.value = false
  buildAndSyncTree()
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
    timeout_minutes: singleStepForm.timeout_minutes,
    default_assignee_role: singleStepForm.default_assignee_role,
    executor_team: singleStepForm.executor_team,
    parent_step_id: singleStepForm.parent_step_id,
    order_index: steps.length + 1,
    created_at: new Date().toISOString(),
    estimated_duration_minutes: singleStepForm.estimated_duration_minutes,
    estimated_start_offset: singleStepForm.estimated_start_offset,
    phase_step: singleStepForm.phase_step || '',
    attributes: { ...singleStepForm.attributes },
  })
  ElMessage.success('步骤已添加')

  resetSingleStepForm()
  singleAddVisible.value = false
  buildAndSyncTree()
}

// ============ 保存 ============

async function handleSaveSteps() {
  const allSteps = getAllSteps()
  // Build position map for parent step remapping (1-based index)
  const idToPos = new Map<number, number>()
  allSteps.forEach((s, idx) => { if (s.id) idToPos.set(s.id, idx + 1) })

  try {
    await templateApi.updateSteps(templateId.value, allSteps.map((s, idx) => {
      const payload: {
        name: string; seq: number; step_type: string; timeout_minutes?: number;
        guide_content?: string; default_assignee_role?: string; executor_team?: string;
        phase?: string; phase_step?: string; parent_step_id?: number;
        estimated_duration_minutes?: number; estimated_start_offset?: number;
        attributes?: string; id?: number;
      } = {
        name: s.name,
        seq: s.order_index || (idx + 1),
        step_type: s.step_type || 'serial',
        timeout_minutes: s.timeout_minutes || 5,
        guide_content: s.description || s.guide_content || '',
        default_assignee_role: s.default_assignee_role || '',
        executor_team: s.executor_team || '',
        phase: s.phase || '',
        phase_step: s.phase_step || '',
        estimated_duration_minutes: s.estimated_duration_minutes,
        estimated_start_offset: s.estimated_start_offset,
        attributes: typeof s.attributes === 'string' ? s.attributes : JSON.stringify(s.attributes || {}),
      }
      if (s.id && Number.isInteger(s.id)) {
        payload.id = s.id
      }
      if (s.parent_step_id && s.parent_step_id > 0) {
        const parentPos = idToPos.get(s.parent_step_id)
        if (parentPos) payload.parent_step_id = parentPos
      }
      return payload
    }))
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
  importVisible.value = true
}

function downloadTemplate() {
  const header = ['阶段', '环节', '父步骤名称', '步骤名称', '描述', '步骤类型', '预计耗时 (分)', '超时时间 (分)', '执行角色', '执行团队', '责任部门', '配合部门', '责任团队', '操作人', '复核人', '操作说明', '验证方式', '最坏影响分析', '兜底措施', '备注']
  const colWidths = [
    { wch: 14 }, { wch: 12 }, { wch: 20 }, { wch: 20 }, { wch: 30 }, { wch: 12 }, { wch: 12 }, { wch: 12 },
    { wch: 10 }, { wch: 12 }, { wch: 12 }, { wch: 12 }, { wch: 12 }, { wch: 12 }, { wch: 12 },
    { wch: 40 }, { wch: 30 }, { wch: 30 }, { wch: 30 }, { wch: 20 },
  ]

  // 按阶段分组示例数据，每个阶段对应一个 sheet
  const phasesData: Record<string, any[][]> = {
    '准备阶段': [
      header,
      ['准备阶段', '初始化', '', '检查数据库状态', '检查主库是否正常运行', '串行', '10', '5', '执行组', '技术部', '技术部', '', 'DBA组', '李四', '王五', '连接主库检查状态，确认连接池正常', '执行 SHOW SLAVE STATUS 确认从库状态', '主库不可用导致业务中断', '切换从库为主库', ''],
      ['准备阶段', '初始化', '检查数据库状态', '检查主库连接', '检查主库连接池状态', '串行', '5', '5', '执行组', '技术部', '技术部', '', 'DBA组', '李四', '', '连接主库确认连接池正常', '', '', '', ''],
    ],
    '执行阶段': [
      header,
      ['执行阶段', '主流程', '', '切换从库', '将从库提升为主库', '并行', '15', '5', '指挥组', '运维部', '运维部', '', '运维组', '钱七', '孙八', '停止主库写入，提升从库，更新应用配置', '应用连接新主库验证读写正常', '数据不一致或丢失', '回退到原主库', '注意备份数据'],
    ],
  }

  const wb = XLSX.utils.book_new()
  for (const [phaseName, rows] of Object.entries(phasesData)) {
    const ws = XLSX.utils.aoa_to_sheet(rows)
    ws['!cols'] = colWidths
    // sheet 名最长 31 字符
    const sheetName = phaseName.length > 31 ? phaseName.slice(0, 31) : phaseName
    XLSX.utils.book_append_sheet(wb, ws, sheetName)
  }
  XLSX.writeFile(wb, `步骤导入模板_${templateName.value}.xlsx`)
  ElMessage.success('模板已下载')
}

// 步骤类型和角色映射（英文 → 中文），保证导出的 Excel 可被导入识别
const stepTypeLabels: Record<string, string> = {
  serial: '串行',
  parallel: '并行',
}

const assigneeRoleLabels: Record<string, string> = {
  director: '指挥组',
  executor: '执行组',
}

function exportSteps() {
  // 收集所有阶段的所有步骤，按阶段分组
  const phaseStepsMap = new Map<string, StepTemplate[]>()
  for (const phase of phases.value) {
    if (phase.steps.length > 0) {
      phaseStepsMap.set(phase.name, [...phase.steps])
    }
  }

  if (phaseStepsMap.size === 0) {
    ElMessage.warning('当前模板没有步骤可导出')
    return
  }

  // 构建 ID → 步骤名称映射表（用于父步骤名称查表）
  const idToName = new Map<number, string>()
  for (const [, steps] of phaseStepsMap) {
    for (const step of steps) {
      if (step.id) idToName.set(step.id, step.name)
    }
  }

  const header = ['阶段', '环节', '父步骤名称', '步骤名称', '描述', '步骤类型', '预计耗时 (分)', '超时时间 (分)', '执行角色', '执行团队', '责任部门', '配合部门', '责任团队', '操作人', '复核人', '操作说明', '验证方式', '最坏影响分析', '兜底措施', '备注']
  const colWidths = [
    { wch: 14 }, { wch: 12 }, { wch: 20 }, { wch: 20 }, { wch: 30 }, { wch: 12 }, { wch: 12 }, { wch: 12 },
    { wch: 10 }, { wch: 12 }, { wch: 12 }, { wch: 12 }, { wch: 12 }, { wch: 12 }, { wch: 12 },
    { wch: 40 }, { wch: 30 }, { wch: 30 }, { wch: 30 }, { wch: 20 },
  ]

  const wb = XLSX.utils.book_new()
  let totalSteps = 0

  for (const [phaseName, steps] of phaseStepsMap) {
    const data: any[][] = [header]
    for (const step of steps) {
      const attrs = step.attributes || {}
      const parentName = step.parent_step_id && step.parent_step_id > 0
        ? (idToName.get(step.parent_step_id) || '')
        : ''
      data.push([
        phaseName,
        step.phase_step || '',
        parentName,
        step.name || '',
        step.description || step.guide_content || '',
        stepTypeLabels[step.step_type] || step.step_type || '串行',
        step.estimated_duration_minutes ?? '',
        step.timeout_minutes ?? 5,
        assigneeRoleLabels[step.default_assignee_role || ''] || step.default_assignee_role || '执行组',
        step.executor_team || '',
        attrs.responsible_department || '',
        attrs.cooperating_department || '',
        attrs.responsible_team || '',
        attrs.operator || '',
        attrs.reviewer || '',
        attrs.operation_guide || '',
        attrs.verification_method || '',
        attrs.worst_case_analysis || '',
        attrs.fallback_measures || '',
        attrs.remark || '',
      ])
    }
    const ws = XLSX.utils.aoa_to_sheet(data)
    ws['!cols'] = colWidths
    // sheet 名最长 31 字符
    const sheetName = phaseName.length > 31 ? phaseName.slice(0, 31) : phaseName
    XLSX.utils.book_append_sheet(wb, ws, sheetName)
    totalSteps += steps.length
  }

  XLSX.writeFile(wb, `步骤导出_${templateName.value}_${new Date().toISOString().slice(0, 10)}.xlsx`)
  ElMessage.success(`成功导出 ${totalSteps} 个步骤到 ${phaseStepsMap.size} 个 Sheet`)
}

function handleExcelUpload(file: File) {
  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      const data = new Uint8Array(e.target?.result as ArrayBuffer)
      const workbook = XLSX.read(data, { type: 'array', cellDates: true })
      const errors: string[] = []

      // 构建全局名称 → ID 映射表（所有已存在的阶段 + 步骤）
      const nameToIdMap = new Map<string, number>()
      for (const phase of phases.value) {
        for (const s of phase.steps) {
          if (s.id && s.name) nameToIdMap.set(s.name, s.id)
        }
      }

      // 阶段 → 步骤列表
      const phaseStepsMap = new Map<string, StepTemplate[]>()
      let globalOrder = 1

      // 遍历所有 Sheet，每个 Sheet 对应一个阶段
      for (const sheetName of workbook.SheetNames) {
        const sheet = workbook.Sheets[sheetName]
        const rows = XLSX.utils.sheet_to_json(sheet, { header: 1 }) as any[][]

        if (rows.length < 2) continue // 跳过空 Sheet

        // 20 列表头
        for (let i = 1; i < rows.length; i++) {
          const row = rows[i]
          const rowNum = i + 1
          const phase = String(row[0] || '').trim()
          const phaseStepName = String(row[1] || '').trim()
          const parentStepName = String(row[2] || '').trim()
          const name = String(row[3] || '').trim()
          const description = String(row[4] || '').trim()
          const stepTypeRaw = String(row[5] || '').trim()
          const estimatedDuration = parseInt(String(row[6] || '')) || undefined
          const timeoutMinutes = parseInt(String(row[7] || '')) || 5
          const assigneeRoleRaw = String(row[8] || '').trim()
          const executorTeam = String(row[9] || '').trim()
          const responsibleDepartment = String(row[10] || '').trim()
          const cooperatingDepartment = String(row[11] || '').trim()
          const responsibleTeam = String(row[12] || '').trim()
          const operator = String(row[13] || '').trim()
          const reviewer = String(row[14] || '').trim()
          const operationGuide = String(row[15] || '').trim()
          const verificationMethod = String(row[16] || '').trim()
          const worstCaseAnalysis = String(row[17] || '').trim()
          const fallbackMeasures = String(row[18] || '').trim()
          const remark = String(row[19] || '').trim()

          if (!name) {
            errors.push(`「${sheetName}」第${rowNum}行：步骤名称不能为空`)
            continue
          }
          if (!phase) {
            errors.push(`「${sheetName}」第${rowNum}行：阶段不能为空`)
            continue
          }

          // 检查步骤名称是否重复
          if (nameToIdMap.has(name)) {
            errors.push(`「${sheetName}」第${rowNum}行：步骤名称「${name}」与已有步骤重复`)
            continue
          }

          // 解析父子关系
          let parentStepId: number | undefined
          if (parentStepName) {
            const parentId = nameToIdMap.get(parentStepName)
            if (parentId) {
              parentStepId = parentId
            } else {
              errors.push(`「${sheetName}」第${rowNum}行：父步骤「${parentStepName}」不存在，请确保父步骤在该步骤之前出现`)
              continue
            }
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

          const newId = Date.now() + Math.random()
          const newStep: StepTemplate = {
            id: newId,
            template_id: templateId.value,
            parent_step_id: parentStepId,
            name,
            description,
            step_type: stepType as any,
            timeout_minutes: timeoutMinutes,
            default_assignee_role: assigneeRole,
            executor_team: executorTeam || '',
            order_index: globalOrder++,
            created_at: new Date().toISOString(),
            phase,
            phase_step: phaseStepName || '',
            estimated_duration_minutes: estimatedDuration,
            attributes: {
              responsible_department: responsibleDepartment || undefined,
              cooperating_department: cooperatingDepartment || undefined,
              responsible_team: responsibleTeam || undefined,
              operator: operator || undefined,
              reviewer: reviewer || undefined,
              operation_guide: operationGuide || undefined,
              verification_method: verificationMethod || undefined,
              worst_case_analysis: worstCaseAnalysis || undefined,
              fallback_measures: fallbackMeasures || undefined,
              remark: remark || undefined,
            },
          }

          if (!phaseStepsMap.has(phase)) {
            phaseStepsMap.set(phase, [])
          }
          phaseStepsMap.get(phase)!.push(newStep)

          // 将新步骤加入映射表，后续行可以引用它作为父步骤
          nameToIdMap.set(name, newId)
        }
      } // end sheet loop

      if (errors.length > 0) {
        ElMessage.warning(errors.join('\n'))
        return false
      }

      // 统计总导入步骤数
      let totalImported = 0
      for (const [, steps] of phaseStepsMap) {
        totalImported += steps.length
      }
      if (totalImported === 0) {
        ElMessage.warning('未找到有效数据')
        return false
      }

      // 将导入的步骤分配到对应阶段
      for (const [phaseName, newSteps] of phaseStepsMap) {
        const existingPhase = phases.value.find(p => p.name === phaseName)
        if (existingPhase) {
          existingPhase.steps = [...existingPhase.steps, ...newSteps]
        } else {
          phases.value.push({ name: phaseName, steps: newSteps })
        }
      }

      // 激活第一个有数据的阶段
      if (phaseStepsMap.size > 0) {
        const firstPhase = phaseStepsMap.keys().next().value
        if (firstPhase) activePhaseName.value = firstPhase
      }
      buildAndSyncTree()
      ElMessage.success(`成功导入 ${totalImported} 个步骤到 ${phaseStepsMap.size} 个阶段`)
      importVisible.value = false
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
  .draggable-list {
    display: flex;
    flex-direction: column;
  }

  .step-group {
    margin-bottom: 2px;
  }

  .step-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 10px;
    background: white;
    border: 1px solid transparent;
    border-radius: 4px;
    cursor: pointer;
    transition: all 0.15s;

    &:hover {
      background: #ecf5ff;
      border-color: var(--el-color-primary-light-7);
    }
  }

  .child-row {
    margin-left: 32px;
    border-left: 3px solid var(--el-color-primary-light-7);
    background: #f7f8fa;
    min-height: 36px;
    line-height: 1;

    .step-name {
      color: #606266;
      line-height: 1;
    }

    &:hover {
      background: #ecf5ff;
      border-color: var(--el-color-primary);
    }
  }

  .drag-handle {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    cursor: grab;
    color: #c0c4cc;
    font-size: 14px;
    width: 16px;
    height: 24px;
    user-select: none;
    line-height: 1;
    flex-shrink: 0;

    &:hover {
      color: var(--el-color-primary);
    }

    &:active {
      cursor: grabbing;
    }
  }

  .seq-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 24px;
    border-radius: 50%;
    background: var(--el-color-primary);
    color: white;
    font-size: 11px;
    font-weight: 600;
    flex-shrink: 0;
  }

  .expand-btn {
    background: none;
    border: none;
    cursor: pointer;
    color: #606266;
    font-size: 10px;
    width: 16px;
    text-align: center;
    padding: 0;
    flex-shrink: 0;

    &:hover {
      color: var(--el-color-primary);
    }
  }

  .expand-placeholder {
    width: 16px;
    flex-shrink: 0;
  }

  .step-name {
    flex: 1;
    font-size: 13px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .drag-ghost {
    background: rgba(var(--el-color-primary-rgb, 64, 158, 255), 0.15);
    border: 1px dashed var(--el-color-primary);
    border-radius: 4px;
    opacity: 0.6;
  }

  .children-list {
    display: flex;
    flex-direction: column;
    gap: 2px;
    margin-top: 2px;
    margin-left: 4px;
    border-left: 2px solid $border-color;
    padding-left: 8px;
    padding-top: 2px;
  }

  :deep(.el-tag) {
    flex-shrink: 0;
  }

  :deep(.el-button) {
    flex-shrink: 0;
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
