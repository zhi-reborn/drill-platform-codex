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
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="(row.status_label || row.status) === 'enabled' || row.status === 2 ? 'success' : 'info'" size="small">
              {{ (row.status_label || row.status) === 'enabled' || row.status === 2 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_by_name" label="创建人" width="120" />
        <el-table-column prop="updated_at" label="更新时间" width="160">
          <template #default="{ row }">
            {{ formatTime(row.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="350" fixed="right">
          <template #default="{ row }">
            <el-button text type="primary" @click="openEditDialog(row)">编辑</el-button>
            <el-button text type="primary" @click="goToStepsEditor(row)">编辑步骤</el-button>
            <el-button 
              text 
              :type="(row.status_label || row.status) === 'enabled' || row.status === 2 ? 'warning' : 'success'" 
              @click="handleToggleStatus(row)"
            >
              {{ (row.status_label || row.status) === 'enabled' || row.status === 2 ? '禁用' : '启用' }}
            </el-button>
            <el-button text type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="categoryVisible" title="分类管理" width="600px">
      <div class="category-list">
        <div v-for="(cat, index) in editableCategories" :key="index" class="category-item">
          <div class="category-row">
            <div class="category-field">
              <label>编码</label>
              <el-input v-model="cat.value" size="small" placeholder="英文标识，如 database" :disabled="cat.id !== undefined" />
            </div>
            <div class="category-field">
              <label>名称</label>
              <el-input v-model="cat.label" size="small" placeholder="中文名称，如 数据库" />
            </div>
            <div class="category-field">
              <label>标签类型</label>
              <el-select v-model="cat.tagType" size="small" style="width: 100px">
                <el-option label="默认" value="info" />
                <el-option label="主要" value="primary" />
                <el-option label="成功" value="success" />
                <el-option label="警告" value="warning" />
                <el-option label="危险" value="danger" />
              </el-select>
            </div>
          </div>
          <div class="category-actions">
            <el-button size="small" :disabled="index === 0" @click="moveCategory(index, -1)">
              <el-icon><ArrowUp /></el-icon>
            </el-button>
            <el-button size="small" :disabled="index === editableCategories.length - 1" @click="moveCategory(index, 1)">
              <el-icon><ArrowDown /></el-icon>
            </el-button>
            <el-button size="small" type="danger" @click="removeCategory(index)">
              <el-icon><Delete /></el-icon>
            </el-button>
          </div>
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
import { ref, computed, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh, Plus, Delete, Setting } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import type { DrillTemplate, TemplateCategory } from '@/types'
import { templateApi } from '@/api/modules/template'
import { userApi } from '@/api'

const router = useRouter()

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
  { value: 'database', label: '数据库', tagType: 'info' },
  { value: 'cache', label: '缓存', tagType: 'info' },
  { value: 'mq', label: '消息队列', tagType: 'info' },
]

const activeCategory = ref('all')
const templates = ref<DrillTemplate[]>([])
const categories = ref<CategoryItem[]>([...defaultCategories])
const formVisible = ref(false)
const deleteVisible = ref(false)
const categoryVisible = ref(false)
const importVisible = ref(false)
const editableCategories = ref<CategoryItem[]>([])

const isEditing = ref(false)
const editingId = ref<number | null>(null)
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

const departmentOptions = ref<string[]>([])

async function loadDepartments() {
  try {
    const depts = await userApi.getDepartments()
    departmentOptions.value = depts
  } catch (error) {
    console.error('Failed to load departments:', error)
  }
}

function goToStepsEditor(template: DrillTemplate) {
  router.push({ name: 'DirectorTemplateSteps', params: { id: template.id } })
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

async function loadTemplates() {
  try {
    const params: any = { page: 1, page_size: 100 }
    if (activeCategory.value !== 'all') {
      params.category = activeCategory.value
    }
    const result = await templateApi.getList(params)
    templates.value = result.list || []
  } catch (error) {
    ElMessage.error('加载模板列表失败')
    console.error('Failed to load templates:', error)
  }
}

async function loadCategories() {
  try {
    const result = await templateApi.getCategories()
    categories.value = result.map(c => ({
      value: c.value,
      label: c.label,
      tagType: c.tag_type as any,
    }))
  } catch (error) {
    console.error('Failed to load categories:', error)
  }
}

onMounted(() => {
  loadCategories()
  loadTemplates()
  loadDepartments()
})

function openCategoryDialog() {
  editableCategories.value = JSON.parse(JSON.stringify(categories.value))
  categoryVisible.value = true
}

function addCategory() {
  editableCategories.value.push({
    value: '',
    label: '',
    tagType: 'info',
  })
}

function removeCategory(index: number) {
  const cat = editableCategories.value[index]
  // 如果分类已有 ID（已存在于数据库），检查是否有模板使用
  if (cat.id !== undefined && templates.value.some(t => t.category === cat.value)) {
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

async function handleSaveCategories() {
  // 验证编码（value）不能为空
  for (const cat of editableCategories.value) {
    if (!cat.value || !cat.value.trim()) {
      ElMessage.warning('分类编码不能为空')
      return
    }
    // 验证编码格式（只允许字母、数字、下划线）
    if (!/^[a-zA-Z][a-zA-Z0-9_]*$/.test(cat.value)) {
      ElMessage.warning('分类编码必须以字母开头，只能包含字母、数字和下划线')
      return
    }
  }
  
  // 验证名称不能为空
  for (const cat of editableCategories.value) {
    if (!cat.label || !cat.label.trim()) {
      ElMessage.warning('分类名称不能为空')
      return
    }
  }
  
  // 验证编码不能重复
  const valueCount: Record<string, number> = {}
  for (const cat of editableCategories.value) {
    valueCount[cat.value] = (valueCount[cat.value] || 0) + 1
  }
  for (const [value, count] of Object.entries(valueCount)) {
    if (count > 1) {
      ElMessage.warning(`分类编码 "${value}" 重复`)
      return
    }
  }
  
  // 验证名称不能重复
  const labelCount: Record<string, number> = {}
  for (const cat of editableCategories.value) {
    labelCount[cat.label] = (labelCount[cat.label] || 0) + 1
  }
  for (const [label, count] of Object.entries(labelCount)) {
    if (count > 1) {
      ElMessage.warning(`分类名称 "${label}" 重复`)
      return
    }
  }
  
  try {
    // 保存所有分类，由后端处理排序和删除
    await templateApi.saveCategories(editableCategories.value.map(c => ({
      value: c.value.trim(),
      label: c.label.trim(),
      tag_type: c.tagType,
    })))
    // 重新加载分类
    await loadCategories()
    ElMessage.success('分类已保存')
    categoryVisible.value = false
  } catch (error) {
    ElMessage.error('保存分类失败')
    console.error('Save categories error:', error)
  }
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

async function handleSave() {
  if (!form.name.trim()) {
    ElMessage.warning('请输入模板名称')
    return
  }
  if (!form.category) {
    ElMessage.warning('请选择分类')
    return
  }

  try {
    if (isEditing.value && editingId.value) {
      await templateApi.update(editingId.value, {
        name: form.name,
        category: form.category,
        description: form.description,
      })
      ElMessage.success('模板已更新')
      loadTemplates()
    } else {
      await templateApi.create({
        name: form.name,
        category: form.category,
        description: form.description,
      })
      ElMessage.success('模板已创建')
      loadTemplates()
    }
    formVisible.value = false
  } catch (error) {
    ElMessage.error(isEditing.value ? '更新失败' : '创建失败')
    console.error('Template save error:', error)
  }
}

function handleDelete(template: DrillTemplate) {
  deleteTarget.value = template
  deleteVisible.value = true
}

async function confirmDelete() {
  if (deleteTarget.value) {
    try {
      await templateApi.delete(deleteTarget.value.id)
      ElMessage.success('模板已删除')
      loadTemplates()
    } catch (error) {
      ElMessage.error('删除失败')
      console.error('Template delete error:', error)
    } finally {
      deleteVisible.value = false
      deleteTarget.value = null
    }
  }
  deleteVisible.value = false
  deleteTarget.value = null
}

async function handleToggleStatus(template: DrillTemplate) {
  try {
    await templateApi.toggleStatus(template.id)
    ElMessage.success('状态已更新')
    loadTemplates()
  } catch (error) {
    ElMessage.error('操作失败')
    console.error('Template toggle status error:', error)
  }
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
    align-items: flex-start;
    gap: $spacing-sm;
    padding: $spacing-sm;
    margin-bottom: $spacing-sm;
    background: $bg-tertiary;
    border-radius: $radius-base;

    .category-row {
      flex: 1;
      display: flex;
      gap: $spacing-sm;
      align-items: center;

      .category-field {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 4px;

        label {
          font-size: $font-size-xs;
          color: $text-secondary;
        }
      }
    }

    .category-actions {
      display: flex;
      gap: 4px;
      padding-top: 20px;
    }
  }
}

.add-category {
  text-align: center;
  padding: $spacing-base;
}
</style>