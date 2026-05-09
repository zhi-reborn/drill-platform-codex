<template>
  <div class="template-page">
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">模板管理</h1>
        <p class="page-subtitle">管理和配置演练流程模板</p>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="handleCreate">
          <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19" />
            <line x1="5" y1="12" x2="19" y2="12" />
          </svg>
          新建模板
        </el-button>
      </div>
    </div>

    <div class="page-content">
      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-form :inline="true" :model="filterForm" class="filter-form">
          <el-form-item label="模板名称">
            <el-input
              v-model="filterForm.name"
              placeholder="搜索模板名称"
              clearable
              @clear="handleSearch"
            />
          </el-form-item>
          <el-form-item label="分类">
            <el-select v-model="filterForm.category" placeholder="全部类别" clearable @change="handleSearch">
              <el-option label="灾备切换" value="disaster_recovery" />
              <el-option label="服务降级" value="degradation" />
              <el-option label="发布回滚" value="rollback" />
              <el-option label="安全演练" value="security" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">查询</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 模板卡片列表 -->
      <div class="template-grid" v-loading="loading">
        <div
          v-for="template in templates"
          :key="template.id"
          class="template-card"
          @click="handleView(template)"
        >
          <div class="card-header">
            <h3 class="card-title">{{ template.name }}</h3>
            <el-tag :type="getCategoryType(template.category)" size="small">
              {{ getCategoryText(template.category) }}
            </el-tag>
          </div>
          <p class="card-description">{{ template.description }}</p>
          <div class="card-meta">
            <span class="meta-item">
              <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" />
                <polyline points="14 2 14 8 20 8" />
              </svg>
              {{ template.stepCount }} 个步骤
            </span>
            <span class="meta-item">
              <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10" />
                <polyline points="12 6 12 12 16 14" />
              </svg>
              {{ template.estimatedDuration }} 分钟
            </span>
            <span class="meta-item">
              <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M12 20h9" />
                <path d="M16.5 3.5a2.121 2.121 0 013 3L7 19l-4 1 1-4L16.5 3.5z" />
              </svg>
              更新：{{ formatDate(template.updatedAt) }}
            </span>
          </div>
          <div class="card-actions">
            <el-button size="small" @click.stop="handleEdit(template)">编辑</el-button>
            <el-button size="small" @click.stop="handleCopy(template)">复制</el-button>
            <el-button size="small" type="danger" @click.stop="handleDelete(template)">删除</el-button>
          </div>
        </div>

        <!-- 空状态 -->
        <div v-if="templates.length === 0" class="empty-state">
          <svg class="empty-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" />
            <polyline points="14 2 14 8 20 8" />
            <line x1="16" y1="13" x2="8" y2="13" />
            <line x1="16" y1="17" x2="8" y2="17" />
          </svg>
          <p class="empty-text">暂无模板</p>
          <p class="empty-hint">创建第一个演练模板来开始使用</p>
          <el-button type="primary" @click="handleCreate">新建模板</el-button>
        </div>
      </div>
    </div>

    <!-- 模板详情对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="800px"
      :close-on-click-modal="false"
    >
      <div v-if="selectedTemplate" class="dialog-content">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="模板名称">{{ selectedTemplate.name }}</el-descriptions-item>
          <el-descriptions-item label="分类">
            <el-tag :type="getCategoryType(selectedTemplate.category)">
              {{ getCategoryText(selectedTemplate.category) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="步骤数量">{{ selectedTemplate.stepCount }}</el-descriptions-item>
          <el-descriptions-item label="预计时长">{{ selectedTemplate.estimatedDuration }} 分钟</el-descriptions-item>
          <el-descriptions-item label="创建时间" :span="2">
            {{ formatDate(selectedTemplate.createdAt) }}
          </el-descriptions-item>
          <el-descriptions-item label="模板描述" :span="2">
            {{ selectedTemplate.description || '无' }}
          </el-descriptions-item>
        </el-descriptions>

        <el-divider>流程步骤</el-divider>

        <div class="step-list">
          <div
            v-for="(step, index) in selectedTemplate.steps"
            :key="step.id"
            class="step-item"
          >
            <span class="step-index">{{ index + 1 }}</span>
            <span class="step-name">{{ step.name }}</span>
            <el-tag size="small" :type="getStepType(step.stepType)">
              {{ getStepTypeText(step.stepType) }}
            </el-tag>
            <span class="step-timeout" v-if="step.timeoutMinutes">
              {{ step.timeoutMinutes }}分钟
            </span>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="dialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

interface Template {
  id: number
  name: string
  category: string
  description: string
  stepCount: number
  estimatedDuration: number
  createdAt: number
  updatedAt: number
  steps: Array<{
    id: number
    name: string
    stepType: string
    timeoutMinutes?: number
  }>
}

const loading = ref(false)
const dialogVisible = ref(false)
const selectedTemplate = ref<Template | null>(null)

const filterForm = reactive({
  name: '',
  category: ''
})

// 模拟数据
const templates = ref<Template[]>([
  {
    id: 1,
    name: 'DB 主从切换模板',
    category: 'disaster_recovery',
    description: '数据库主从切换标准流程，适用于 MySQL、PostgreSQL 等数据库',
    stepCount: 7,
    estimatedDuration: 120,
    createdAt: Date.now() - 86400000,
    updatedAt: Date.now() - 3600000,
    steps: [
      { id: 1, name: '开始', stepType: 'serial' },
      { id: 2, name: '备份检查', stepType: 'serial', timeoutMinutes: 10 },
      { id: 3, name: '主库降级', stepType: 'serial', timeoutMinutes: 15 },
      { id: 4, name: '从库升级', stepType: 'serial', timeoutMinutes: 15 },
      { id: 5, name: '流量切换', stepType: 'serial', timeoutMinutes: 10 },
      { id: 6, name: '验证测试', stepType: 'serial', timeoutMinutes: 20 },
      { id: 7, name: '结束', stepType: 'serial' }
    ]
  },
  {
    id: 2,
    name: '服务降级标准流程',
    category: 'degradation',
    description: '服务降级和熔断标准流程，适用于微服务架构',
    stepCount: 5,
    estimatedDuration: 60,
    createdAt: Date.now() - 172800000,
    updatedAt: Date.now() - 86400000,
    steps: [
      { id: 1, name: '开始', stepType: 'serial' },
      { id: 2, name: '流量评估', stepType: 'serial', timeoutMinutes: 5 },
      { id: 3, name: '降级执行', stepType: 'parallel', timeoutMinutes: 10 },
      { id: 4, name: '验证确认', stepType: 'serial', timeoutMinutes: 5 },
      { id: 5, name: '结束', stepType: 'serial' }
    ]
  }
])

const dialogTitle = computed(() => {
  return selectedTemplate.value ? selectedTemplate.value.name : '模板详情'
})

// 获取分类类型
const getCategoryType = (category: string) => {
  const typeMap: Record<string, any> = {
    disaster_recovery: '',
    degradation: 'warning',
    rollback: 'success',
    security: 'danger'
  }
  return typeMap[category] || ''
}

// 获取分类文本
const getCategoryText = (category: string) => {
  const textMap: Record<string, string> = {
    disaster_recovery: '灾备切换',
    degradation: '服务降级',
    rollback: '发布回滚',
    security: '安全演练'
  }
  return textMap[category] || category
}

// 获取步骤类型
const getStepType = (type: string) => {
  const typeMap: Record<string, any> = {
    serial: '',
    parallel: 'warning',
    any_of: 'success',
    condition: 'info'
  }
  return typeMap[type] || ''
}

// 获取步骤类型文本
const getStepTypeText = (type: string) => {
  const textMap: Record<string, string> = {
    serial: '串行',
    parallel: '并行',
    any_of: '或签',
    condition: '条件'
  }
  return textMap[type] || type
}

// 格式化日期
const formatDate = (timestamp: number) => {
  return new Date(timestamp).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// 加载数据
const handleFetchData = async () => {
  loading.value = true
  try {
    await new Promise(resolve => setTimeout(resolve, 500))
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  handleFetchData()
}

// 重置
const handleReset = () => {
  filterForm.name = ''
  filterForm.category = ''
  handleSearch()
}

// 新建
const handleCreate = () => {
  ElMessage.info('新建模板功能开发中')
}

// 查看
const handleView = (template: Template) => {
  selectedTemplate.value = template
  dialogVisible.value = true
}

// 编辑
const handleEdit = (template: Template) => {
  ElMessage.info('编辑模板功能开发中')
}

// 复制
const handleCopy = (template: Template) => {
  ElMessage.info('复制模板功能开发中')
}

// 删除
const handleDelete = async (template: Template) => {
  try {
    await ElMessageBox.confirm(`确认删除模板"${template.name}"？`, '警告', {
      type: 'warning',
      confirmButtonText: '删除',
      confirmButtonClass: 'el-button--danger'
    })
    const index = templates.value.findIndex(t => t.id === template.id)
    if (index !== -1) {
      templates.value.splice(index, 1)
    }
    ElMessage.success('删除成功')
  } catch {
    // 取消
  }
}
</script>

<style scoped>
.template-page {
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

.header-left {
  display: flex;
  flex-direction: column;
}

.page-title {
  font-size: 24px;
  font-weight: 700;
  margin: 0;
  color: var(--color-foreground, #F8FAFC);
}

.page-subtitle {
  font-size: 14px;
  color: var(--color-muted-foreground, #94A3B8);
  margin-top: 4px;
}

.header-right .el-button {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.header-right .icon {
  width: 16px;
  height: 16px;
}

.page-content {
  background-color: var(--color-muted, #1A1E2F);
  border-radius: 8px;
  border: 1px solid var(--color-border, #334155);
  padding: 20px;
}

.filter-bar {
  margin-bottom: 20px;
}

.filter-form .el-form-item {
  margin-bottom: 0;
  margin-right: 16px;
}

.template-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 20px;
}

.template-card {
  padding: 20px;
  border-radius: 8px;
  border: 1px solid var(--color-border, #334155);
  background-color: var(--color-secondary, #1E293B);
  cursor: pointer;
  transition: all 0.2s ease;
}

.template-card:hover {
  border-color: var(--color-primary, #0F172A);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.card-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--color-foreground, #F8FAFC);
}

.card-description {
  font-size: 14px;
  color: var(--color-muted-foreground, #94A3B8);
  margin-bottom: 16px;
  line-height: 1.6;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-meta {
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: var(--color-muted-foreground, #94A3B8);
  margin-bottom: 16px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

.meta-item .icon {
  width: 14px;
  height: 14px;
}

.card-actions {
  display: flex;
  gap: 8px;
  padding-top: 16px;
  border-top: 1px solid var(--color-border, #334155);
}

.empty-state {
  grid-column: 1 / -1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
}

.empty-icon {
  width: 80px;
  height: 80px;
  color: var(--color-border, #334155);
  margin-bottom: 16px;
}

.empty-text {
  font-size: 16px;
  color: var(--color-foreground, #F8FAFC);
  margin-bottom: 8px;
}

.empty-hint {
  font-size: 14px;
  color: var(--color-muted-foreground, #94A3B8);
  margin-bottom: 20px;
}

.dialog-content {
  padding: 10px 0;
}

.step-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.step-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: 6px;
  background-color: var(--color-secondary, #1E293B);
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
  color: var(--color-foreground, #F8FAFC);
}

.step-timeout {
  font-size: 13px;
  color: var(--color-muted-foreground, #94A3B8);
}

:deep(.el-card),
:deep(.el-dialog) {
  background-color: var(--color-muted, #1A1E2F);
  border: 1px solid var(--color-border, #334155);
}

:deep(.el-dialog__title) {
  color: var(--color-foreground, #F8FAFC);
}
</style>
