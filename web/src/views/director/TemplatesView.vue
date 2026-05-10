<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">模板管理</h2>
      <div class="header-actions">
        <el-button @click="loadTemplates">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-button @click="openCategoryDialog">
          <el-icon><Setting /></el-icon>
          分类管理
        </el-button>
        <el-button type="primary" @click="openCreateDialog">
          <el-icon><Plus /></el-icon>
          新建模板
        </el-button>
      </div>
    </div>

    <div class="page-content">
      <el-tabs v-model="activeCategory" class="category-tabs">
        <el-tab-pane label="全部" name="all" />
        <el-tab-pane v-for="cat in categories" :key="cat.value" :label="cat.label" :name="cat.value" />
      </el-tabs>

      <el-table :data="filteredTemplates" style="width: 100%" class="templates-table">
        <el-table-column prop="name" label="模板名" min-width="200" />
        <el-table-column prop="category" label="分类" width="120">
          <template #default="{ row }">
            <el-tag :type="getCategoryTagType(row.category)" size="small">
              {{ getCategoryLabel(row.category) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="version" label="版本" width="100" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'published' ? 'success' : 'info'" size="small">
              {{ row.status === 'published' ? '已发布' : '草稿' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_by_name" label="创建人" width="120" />
        <el-table-column prop="updated_at" label="更新时间" width="160">
          <template #default="{ row }">
            {{ formatTime(row.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button text type="primary" @click="openEditDialog(row)">编辑</el-button>
            <el-button text type="primary" @click="openStepsDialog(row)">编辑步骤</el-button>
            <el-button text type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="categoryVisible" title="分类管理" width="500px">
      <div class="category-list">
        <div v-for="(cat, index) in editableCategories" :key="cat.value" class="category-item">
          <el-tag :type="cat.tagType" size="small" style="min-width: 80px; text-align: center;">
            {{ cat.label }}
          </el-tag>
          <el-input v-model="cat.label" size="small" class="category-input" placeholder="分类名称" />
          <el-button-group>
            <el-button size="small" :disabled="index === 0" @click="moveCategory(index, -1)">
              <el-icon><ArrowUp /></el-icon>
            </el-button>
            <el-button size="small" :disabled="index === editableCategories.length - 1" @click="moveCategory(index, 1)">
              <el-icon><ArrowDown /></el-icon>
            </el-button>
            <el-button size="small" type="danger" @click="removeCategory(index)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </el-button-group>
        </div>
      </div>
      <div class="add-category">
        <el-button type="primary" plain @click="addCategory">
          <el-icon><Plus /></el-icon>
          添加分类
        </el-button>
      </div>
      <template #footer>
        <el-button @click="categoryVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveCategories">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="formVisible" :title="isEditing ? '编辑模板' : '新建模板'" width="500px">
      <el-form :model="form" label-width="80px">
        <el-form-item label="模板名称" required>
          <el-input v-model="form.name" placeholder="请输入模板名称" />
        </el-form-item>
        <el-form-item label="分类" required>
          <el-select v-model="form.category" placeholder="请选择分类">
            <el-option v-for="cat in categories" :key="cat.value" :label="cat.label" :value="cat.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入模板描述" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="stepsVisible" :title="`编辑步骤 - ${editingTemplateName}`" size="900px">
      <template #header>
        <div class="steps-drawer-header">
          <span>编辑步骤 - {{ editingTemplateName }}</span>
          <div class="header-right">
            <el-button type="primary" @click="openBatchImportDialog">
              <el-icon><Download /></el-icon>
              批量导入
            </el-button>
          </div>
        </div>
      </template>
      <div class="steps-editor">
        <div v-if="editingSteps.length > 0" class="steps-table">
          <el-table :data="editingSteps" border size="small" row-key="id">
            <el-table-column type="index" label="序号" width="60" align="center" />
            <el-table-column prop="name" label="步骤名称" min-width="150">
              <template #default="{ row }">
                {{ row.name || '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="description" label="描述" min-width="180" show-overflow-tooltip>
              <template #default="{ row }">
                {{ row.description || '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="step_type" label="类型" width="80" align="center">
              <template #default="{ row }">
                {{ getStepTypeLabel(row.step_type) }}
              </template>
            </el-table-column>
            <el-table-column prop="timeout_seconds" label="超时(秒)" width="90" align="center" />
            <el-table-column prop="assignee" label="操作人" width="100" show-overflow-tooltip>
              <template #default="{ row }">
                {{ row.assignee || '-' }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="160" align="center" fixed="right">
              <template #default="{ $index }">
                <el-button-group size="small">
                  <el-button text type="primary" @click="openStepEditDialog($index)" title="编辑">
                    <el-icon><Edit /></el-icon>
                  </el-button>
                  <el-button :disabled="$index === 0" text @click="moveStep($index, -1)" title="上移">
                    <el-icon><Top /></el-icon>
                  </el-button>
                  <el-button :disabled="$index === editingSteps.length - 1" text @click="moveStep($index, 1)" title="下移">
                    <el-icon><Bottom /></el-icon>
                  </el-button>
                  <el-button text type="danger" @click="removeStep($index)" title="删除">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </el-button-group>
              </template>
            </el-table-column>
          </el-table>
          <div class="single-add-wrapper">
            <el-button type="primary" plain @click="openSingleAddDialog">
              <el-icon><Plus /></el-icon>
              单个添加
            </el-button>
          </div>
        </div>
        <div v-else class="steps-empty">
          <el-empty description="暂无步骤">
            <el-button type="primary" @click="openBatchImportDialog">
              <el-icon><Download /></el-icon>
              导入步骤
            </el-button>
          </el-empty>
        </div>
      </div>
      <template #footer>
        <div class="drawer-footer">
          <el-button @click="stepsVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSaveSteps">保存步骤</el-button>
        </div>
      </template>
    </el-drawer>

    <el-dialog v-model="importVisible" title="批量导入步骤" width="520px">
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
    </el-dialog>

    <el-dialog v-model="singleAddVisible" title="添加步骤" width="520px">
      <el-form :model="singleStepForm" label-width="90px" class="single-step-form">
        <el-form-item label="步骤名称" required>
          <el-input v-model="singleStepForm.name" placeholder="请输入步骤名称" maxlength="100" show-word-limit />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="singleStepForm.description" type="textarea" placeholder="步骤描述" :rows="2" maxlength="500" show-word-limit />
        </el-form-item>
        <div class="form-row">
          <el-form-item label="步骤类型">
            <el-select v-model="singleStepForm.step_type">
              <el-option label="串行" value="serial" />
              <el-option label="并行" value="parallel" />
              <el-option label="任选" value="any_of" />
              <el-option label="条件" value="condition" />
            </el-select>
          </el-form-item>
          <el-form-item label="超时时间">
            <el-input-number v-model="singleStepForm.timeout_seconds" :min="30" :max="3600" controls-position="right" />
          </el-form-item>
        </div>
        <el-form-item label="操作人">
          <el-input v-model="singleStepForm.assignee" placeholder="填写操作人" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="singleAddVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAddSingleStep">添加步骤</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="deleteVisible" title="确认删除" width="400px">
      <p>确定要删除模板「{{ deleteTarget?.name }}」吗？此操作不可撤销。</p>
      <template #footer>
        <el-button @click="deleteVisible = false">取消</el-button>
        <el-button type="danger" @click="confirmDelete">确认删除</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Plus, Delete, Setting, Upload, Download, Top, Bottom, Edit } from '@element-plus/icons-vue'
import * as XLSX from 'xlsx'
import { useAuthStore } from '@/stores/auth'
import type { DrillTemplate, StepTemplate, TemplateCategory, StepType } from '@/types'
import templatesData from '@/mock/data/templates.json'

interface CategoryItem {
  value: string
  label: string
  tagType: 'primary' | 'success' | 'warning' | 'danger' | 'info'
}

const authStore = useAuthStore()

const defaultCategories: CategoryItem[] = [
  { value: 'disaster_recovery', label: '灾备切换', tagType: 'primary' },
  { value: 'degradation', label: '服务降级', tagType: 'warning' },
  { value: 'release', label: '发布演练', tagType: 'success' },
  { value: 'security', label: '安全事件', tagType: 'danger' },
]

const activeCategory = ref('all')
const templates = ref<DrillTemplate[]>([])
const categories = ref<CategoryItem[]>([...defaultCategories])
const formVisible = ref(false)
const stepsVisible = ref(false)
const deleteVisible = ref(false)
const categoryVisible = ref(false)
const importVisible = ref(false)
const editableCategories = ref<CategoryItem[]>([])

const isEditing = ref(false)
const editingId = ref<number | null>(null)
const editingSteps = ref<StepTemplate[]>([])
const editingTemplateId = ref<number | null>(null)
const editingTemplateName = ref('')
const deleteTarget = ref<DrillTemplate | null>(null)

const filteredTemplates = computed(() => {
  if (activeCategory.value === 'all') return templates.value
  return templates.value.filter(t => t.category === activeCategory.value)
})

const form = reactive({
  name: '',
  category: 'disaster_recovery' as TemplateCategory,
  description: '',
})

const singleStepForm = reactive({
  name: '',
  description: '',
  step_type: 'serial' as StepType,
  timeout_seconds: 300,
  assignee: '',
})

const singleStepEditIndex = ref<number | null>(null)
const singleAddVisible = ref(false)

function openBatchImportDialog() {
  importVisible.value = true
}

function openSingleAddDialog() {
  resetSingleStepForm()
  singleStepEditIndex.value = null
  singleAddVisible.value = true
}

function resetSingleStepForm() {
  singleStepForm.name = ''
  singleStepForm.description = ''
  singleStepForm.step_type = 'serial'
  singleStepForm.timeout_seconds = 300
  singleStepForm.assignee = ''
}

function handleAddSingleStep() {
  if (!singleStepForm.name.trim()) {
    ElMessage.warning('请输入步骤名称')
    return
  }

  if (singleStepEditIndex.value !== null) {
    // 编辑模式
    const step = editingSteps.value[singleStepEditIndex.value]
    step.name = singleStepForm.name.trim()
    step.description = singleStepForm.description.trim()
    step.step_type = singleStepForm.step_type as StepType
    step.timeout_seconds = singleStepForm.timeout_seconds
    step.assignee = singleStepForm.assignee.trim()
    ElMessage.success('步骤已更新')
  } else {
    // 新增模式
    editingSteps.value.push({
      id: Date.now(),
      template_id: editingTemplateId.value || 0,
      name: singleStepForm.name.trim(),
      description: singleStepForm.description.trim(),
      step_type: singleStepForm.step_type as StepType,
      timeout_seconds: singleStepForm.timeout_seconds,
      assignee: singleStepForm.assignee.trim(),
      order_index: editingSteps.value.length + 1,
      created_at: new Date().toISOString(),
    })
    ElMessage.success('步骤已添加')
  }

  resetSingleStepForm()
  singleAddVisible.value = false
}

function moveStep(index: number, direction: number) {
  const newIndex = index + direction
  if (newIndex < 0 || newIndex >= editingSteps.value.length) return
  const temp = editingSteps.value[index]
  editingSteps.value[index] = editingSteps.value[newIndex]
  editingSteps.value[newIndex] = temp
  editingSteps.value.forEach((s: StepTemplate, i: number) => { s.order_index = i + 1 })
}

function openStepEditDialog(index: number) {
  const step = editingSteps.value[index]
  singleStepForm.name = step.name
  singleStepForm.description = step.description || ''
  singleStepForm.step_type = step.step_type as StepType
  singleStepForm.timeout_seconds = step.timeout_seconds
  singleStepForm.assignee = step.assignee || ''
  singleStepEditIndex.value = index
  singleAddVisible.value = true
}

function getCategoryLabel(value: string): string {
  const cat = categories.value.find(c => c.value === value)
  return cat?.label || value
}

function getCategoryTagType(value: string): 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  const cat = categories.value.find(c => c.value === value)
  return cat?.tagType || 'info'
}

function getStepTypeLabel(type: string): string {
  const map: Record<string, string> = {
    serial: '串行',
    parallel: '并行',
    any_of: '任选',
    condition: '条件',
  }
  return map[type] || type
}

function formatTime(dateStr: string): string {
  return new Date(dateStr).toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function loadTemplates() {
  templates.value = JSON.parse(JSON.stringify(templatesData)) as DrillTemplate[]
  ElMessage.success('模板加载成功')
}

function openCategoryDialog() {
  editableCategories.value = JSON.parse(JSON.stringify(categories.value))
  categoryVisible.value = true
}

function addCategory() {
  const newValue = `custom_${Date.now()}` as TemplateCategory
  editableCategories.value.push({
    value: newValue,
    label: '新分类',
    tagType: 'info',
  })
}

function removeCategory(index: number) {
  const cat = editableCategories.value[index]
  if (templates.value.some(t => t.category === cat.value)) {
    ElMessage.warning('该分类下有模板，请先移除或转移模板')
    return
  }
  editableCategories.value.splice(index, 1)
}

function moveCategory(index: number, direction: number) {
  const newIndex = index + direction
  if (newIndex < 0 || newIndex >= editableCategories.value.length) return
  const temp = editableCategories.value[index]
  editableCategories.value[index] = editableCategories.value[newIndex]
  editableCategories.value[newIndex] = temp
}

function handleSaveCategories() {
  const labels = editableCategories.value.map(c => c.label)
  if (new Set(labels).size !== labels.length) {
    ElMessage.warning('分类名称不能重复')
    return
  }
  for (const cat of editableCategories.value) {
    if (!cat.label.trim()) {
      ElMessage.warning('分类名称不能为空')
      return
    }
  }
  categories.value = JSON.parse(JSON.stringify(editableCategories.value))
  ElMessage.success('分类已保存')
  categoryVisible.value = false
}

function openCreateDialog() {
  isEditing.value = false
  editingId.value = null
  form.name = ''
  form.category = (categories.value[0]?.value || 'disaster_recovery') as TemplateCategory
  form.description = ''
  formVisible.value = true
}

function openEditDialog(template: DrillTemplate) {
  isEditing.value = true
  editingId.value = template.id
  form.name = template.name
  form.category = template.category
  form.description = template.description || ''
  formVisible.value = true
}

function handleSave() {
  if (!form.name.trim()) {
    ElMessage.warning('请输入模板名称')
    return
  }
  if (!form.category) {
    ElMessage.warning('请选择分类')
    return
  }

  if (isEditing.value && editingId.value) {
    const idx = templates.value.findIndex(t => t.id === editingId.value)
    if (idx !== -1) {
      templates.value[idx] = {
        ...templates.value[idx],
        name: form.name,
        category: form.category,
        description: form.description,
        updated_at: new Date().toISOString(),
      }
      ElMessage.success('模板已更新')
    }
  } else {
    const newTemplate: DrillTemplate = {
      id: Math.max(...templates.value.map(t => t.id), 0) + 1,
      name: form.name,
      category: form.category,
      description: form.description,
      version: '1.0.0',
      status: 'draft',
      created_by: authStore.user?.id || 1,
      created_by_name: authStore.userName || '当前用户',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      steps: [],
    }
    templates.value.push(newTemplate)
    ElMessage.success('模板已创建')
  }
  formVisible.value = false
}

function openStepsDialog(template: DrillTemplate) {
  editingTemplateId.value = template.id
  editingTemplateName.value = template.name
  editingSteps.value = JSON.parse(JSON.stringify(template.steps || []))
  stepsVisible.value = true
}

function removeStep(index: number) {
  editingSteps.value.splice(index, 1)
  editingSteps.value.forEach((s: StepTemplate, i: number) => { s.order_index = i + 1 })
}

function handleSaveSteps() {
  const template = templates.value.find(t => t.id === editingTemplateId.value)
  if (template) {
    template.steps = editingSteps.value
    template.updated_at = new Date().toISOString()
    ElMessage.success('步骤已保存')
  }
  stepsVisible.value = false
}

function downloadTemplate() {
  const header = ['步骤名称', '描述', '步骤类型', '超时时间(秒)', '操作人']
  const wb = XLSX.utils.book_new()
  const ws = XLSX.utils.aoa_to_sheet([header])
  const colWidths = [
    { wch: 20 }, { wch: 30 }, { wch: 10 }, { wch: 12 }, { wch: 15 }
  ]
  ws['!cols'] = colWidths
  XLSX.utils.book_append_sheet(wb, ws, '步骤导入')
  XLSX.writeFile(wb, `步骤导入模板_${editingTemplateName.value}.xlsx`)
  ElMessage.success('模板已下载')
}

function handleExcelUpload(file: File) {
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

      for (let i = 1; i < rows.length; i++) {
        const row = rows[i]
        const rowNum = i + 1
        const name = String(row[0] || '').trim()
        const description = String(row[1] || '').trim()
        const stepTypeRaw = String(row[2] || '').trim()
        const timeoutSeconds = parseInt(String(row[3] || '300')) || 300
        const assignee = String(row[4] || '').trim()

        if (!name) {
          errors.push(`第${rowNum}行：步骤名称不能为空`)
          continue
        }

        const stepTypeMap: Record<string, string> = {
          '串行': 'serial', '并行': 'parallel', '任选': 'any_of', '条件': 'condition',
          'serial': 'serial', 'parallel': 'parallel', 'any_of': 'any_of', 'condition': 'condition',
        }
        const stepType = stepTypeMap[stepTypeRaw] || 'serial'

        steps.push({
          id: Date.now() + Math.random(),
          template_id: editingTemplateId.value || 0,
          name,
          description,
          step_type: stepType as any,
          timeout_seconds: Math.min(3600, Math.max(30, timeoutSeconds)),
          assignee,
          order_index: editingSteps.value.length + steps.length + 1,
          created_at: new Date().toISOString(),
        })
      }

      if (errors.length > 0) {
        ElMessage.warning(errors.join('\n'))
      }

      if (steps.length > 0) {
        editingSteps.value.push(...steps)
        ElMessage.success(`成功导入 ${steps.length} 个步骤`)
      } else if (errors.length === 0) {
        ElMessage.warning('未找到有效数据')
      }
    } catch {
      ElMessage.error('Excel 文件解析失败')
    }
  }
  reader.readAsArrayBuffer(file)
  return false
}

function handleDelete(template: DrillTemplate) {
  deleteTarget.value = template
  deleteVisible.value = true
}

function confirmDelete() {
  if (deleteTarget.value) {
    templates.value = templates.value.filter(t => t.id !== deleteTarget.value!.id)
    ElMessage.success('模板已删除')
  }
  deleteVisible.value = false
  deleteTarget.value = null
}

loadTemplates()
</script>

<style scoped lang="scss">
@use '@/styles/layout' as *;
@use '@/styles/variables' as *;

.page-container {
  @include page-container;

  .page-header {
    @include page-header;
  }
}

.page-content {
  @include page-content;

  .category-tabs {
    margin-bottom: $spacing-base;
  }

  .templates-table {
    :deep(.el-table__row td) {
      background: $bg-secondary;
      color: $text-primary;
    }
  }
}

.header-actions {
  display: flex;
  gap: $spacing-sm;
}

.category-list {
  .category-item {
    display: flex;
    align-items: center;
    gap: $spacing-base;
    padding: $spacing-sm;
    margin-bottom: $spacing-sm;
    background: $bg-tertiary;
    border-radius: $radius-base;

    .category-input {
      flex: 1;
    }
  }
}

.add-category {
  text-align: center;
  padding: $spacing-base;
}

.steps-drawer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.steps-editor {
  .steps-table {
    max-height: calc(80vh - 200px);
    overflow-y: auto;
  }

  .steps-empty {
    padding: 60px 0;
  }

  .single-add-wrapper {
    display: flex;
    justify-content: center;
    padding: 16px 0;
  }
}

.drawer-footer {
  display: flex;
  justify-content: flex-end;
  gap: $spacing-sm;
}

.import-tabs {
  :deep(.el-tabs__header) {
    margin-bottom: 16px;
  }
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

.single-step-form {
  .form-row {
    display: flex;
    gap: $spacing-base;

    .el-form-item {
      flex: 1;
    }
  }
}

.import-content {
  .import-steps {
    margin-top: $spacing-base;
    border-top: 1px solid $border-color;
    padding-top: $spacing-base;

    .import-steps-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: $spacing-base;
    }

    .import-step-item {
      padding: $spacing-sm;
      border-radius: $radius-sm;

      .step-info {
        display: flex;
        align-items: center;
        gap: $spacing-sm;

        .step-name {
          color: $text-primary;
        }
      }
    }

    .no-steps {
      text-align: center;
      color: $text-secondary;
      padding: $spacing-lg;
    }
  }
}
</style>