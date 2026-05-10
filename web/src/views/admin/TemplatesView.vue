<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">模板管理</h2>
      <el-button @click="loadTemplates">
        <el-icon><Refresh /></el-icon>
        刷新
      </el-button>
    </div>

    <div class="page-content">
      <el-tabs v-model="activeCategory" class="category-tabs">
        <el-tab-pane label="全部" name="all" />
        <el-tab-pane label="灾备切换" name="disaster_recovery" />
        <el-tab-pane label="服务降级" name="degradation" />
        <el-tab-pane label="发布演练" name="release" />
        <el-tab-pane label="安全事件" name="security" />
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
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button text type="primary" @click="viewDetail(row)">
              查看详情
            </el-button>
            <el-button text type="primary" @click="cloneTemplate(row)">
              克隆
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 详情对话框 -->
    <el-dialog v-model="detailVisible" title="模板详情" width="600px">
      <div v-if="selectedTemplate" class="template-detail">
        <h3>{{ selectedTemplate.name }}</h3>
        <p class="description">{{ selectedTemplate.description }}</p>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="分类">
            <el-tag :type="getCategoryTagType(selectedTemplate.category)" size="small">
              {{ getCategoryLabel(selectedTemplate.category) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="版本">{{ selectedTemplate.version }}</el-descriptions-item>
          <el-descriptions-item label="创建人">{{ selectedTemplate.created_by_name }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">
            {{ formatTime(selectedTemplate.updated_at) }}
          </el-descriptions-item>
        </el-descriptions>
        <h4>步骤列表</h4>
        <el-timeline>
          <el-timeline-item
            v-for="step in selectedTemplate.steps"
            :key="step.id"
            :timestamp="`步骤${step.order_index}`"
            placement="top"
          >
            <el-card>
              <h4>{{ step.name }}</h4>
              <p>{{ step.description }}</p>
              <el-tag :type="getStepTypeTag(step.step_type)" size="small">
                {{ getStepTypeLabel(step.step_type) }}
              </el-tag>
              <span class="timeout-info">超时：{{ step.timeout_seconds }}s</span>
            </el-card>
          </el-timeline-item>
        </el-timeline>
      </div>
    </el-dialog>

    <!-- 克隆对话框 -->
    <el-dialog v-model="cloneVisible" title="克隆模板" width="400px">
      <el-form :model="cloneForm" label-width="80px">
        <el-form-item label="新模板名称" required>
          <el-input v-model="cloneForm.newName" placeholder="请输入新模板名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="cloneVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmClone">确认克隆</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import type { DrillTemplate } from '@/types'
import templatesData from '@/mock/data/templates.json'

const activeCategory = ref('all')
const templates = ref<DrillTemplate[]>([])
const detailVisible = ref(false)
const cloneVisible = ref(false)
const selectedTemplate = ref<DrillTemplate | null>(null)
const cloneForm = ref({ newName: '' })

const filteredTemplates = computed(() => {
  if (activeCategory.value === 'all') {
    return templates.value
  }
  return templates.value.filter(t => t.category === activeCategory.value)
})

function getCategoryLabel(category: string): string {
  const map: Record<string, string> = {
    disaster_recovery: '灾备切换',
    degradation: '服务降级',
    release: '发布演练',
    security: '安全事件',
  }
  return map[category] || category
}

function getCategoryTagType(category: string): 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  const map: Record<string, any> = {
    disaster_recovery: 'primary',
    degradation: 'warning',
    release: 'success',
    security: 'danger',
  }
  return map[category] || 'info'
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

function getStepTypeTag(type: string): 'primary' | 'success' | 'warning' | 'info' {
  const map: Record<string, any> = {
    serial: 'primary',
    parallel: 'success',
    any_of: 'warning',
    condition: 'info',
  }
  return map[type] || 'info'
}

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

async function loadTemplates() {
  try {
    templates.value = templatesData as DrillTemplate[]
    ElMessage.success('模板加载成功')
  } catch (error) {
    ElMessage.error('加载模板失败')
    console.error('Failed to load templates:', error)
  }
}

function viewDetail(template: DrillTemplate) {
  selectedTemplate.value = template
  detailVisible.value = true
}

function cloneTemplate(template: DrillTemplate) {
  selectedTemplate.value = template
  cloneForm.value.newName = `${template.name} (副本)`
  cloneVisible.value = true
}

function confirmClone() {
  if (!cloneForm.value.newName.trim()) {
    ElMessage.warning('请输入新模板名称')
    return
  }
  ElMessage.success(`模板「${selectedTemplate.value?.name}」已克隆为「${cloneForm.value.newName}」`)
  cloneVisible.value = false
  selectedTemplate.value = null
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

    .page-title {
      font-size: $font-size-xl;
      font-weight: $font-weight-bold;
      color: $text-primary;
      margin: 0;
    }
  }

  .page-content {
    @include page-content;

    .category-tabs {
      margin-bottom: $spacing-base;
      background: $bg-secondary;
      padding: 0 $spacing-sm;
      border-radius: $radius-base;
    }

    .templates-table {
      background: $bg-secondary;
      border-radius: $radius-base;

      :deep(.el-table__header th) {
        background: $bg-tertiary;
        color: $text-secondary;
      }

      :deep(.el-table__row td) {
        background: $bg-secondary;
        color: $text-primary;
      }

      :deep(.el-table__row--striped td) {
        background: rgba(26, 31, 46, 0.5);
      }
    }
  }
}

.template-detail {
  h3 {
    font-size: $font-size-xl;
    color: $text-primary;
    margin-bottom: $spacing-sm;
  }

  .description {
    color: $text-secondary;
    margin-bottom: $spacing-lg;
    line-height: 1.6;
  }

  h4 {
    font-size: $font-size-md;
    color: $text-primary;
    margin: $spacing-lg 0 $spacing-md;
  }

  .timeout-info {
    margin-left: $spacing-sm;
    font-size: $font-size-sm;
    color: $text-tertiary;
  }
}
</style>
