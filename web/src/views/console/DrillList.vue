<template>
  <div class="drill-list-page">
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">演练列表</h1>
        <p class="page-subtitle">管理和监控所有演练实例</p>
      </div>
      <div class="header-right">
        <el-button type="primary" @click="handleCreate">
          <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19" />
            <line x1="5" y1="12" x2="19" y2="12" />
          </svg>
          新建演练
        </el-button>
      </div>
    </div>

    <div class="page-content">
      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-form :inline="true" :model="filterForm" class="filter-form">
          <el-form-item label="演练名称">
            <el-input
              v-model="filterForm.name"
              placeholder="搜索演练名称"
              clearable
              @clear="handleSearch"
            />
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="filterForm.status" placeholder="全部状态" clearable @change="handleSearch">
              <el-option label="待命" value="pending" />
              <el-option label="执行中" value="running" />
              <el-option label="已暂停" value="paused" />
              <el-option label="已完成" value="completed" />
              <el-option label="已终止" value="terminated" />
            </el-select>
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

      <!-- 数据表格 -->
      <div class="table-container">
        <el-table
          :data="tableData"
          v-loading="loading"
          style="width: 100%"
          @sort-change="handleSortChange"
        >
          <el-table-column prop="id" label="ID" width="80" sortable />
          <el-table-column prop="name" label="演练名称" min-width="200">
            <template #default="{ row }">
              <el-link type="primary" @click="handleView(row)">{{ row.name }}</el-link>
            </template>
          </el-table-column>
          <el-table-column prop="templateName" label="模板" min-width="150" />
          <el-table-column prop="category" label="分类" width="120">
            <template #default="{ row }">
              <el-tag size="small">{{ getCategoryText(row.category) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100" sortable>
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.status)">
                {{ getStatusText(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="progress" label="进度" width="150" sortable>
            <template #default="{ row }">
              <el-progress
                :percentage="getProgressPercent(row)"
                :status="getProgressStatus(row)"
              />
            </template>
          </el-table-column>
          <el-table-column prop="createdAt" label="创建时间" width="180" sortable>
            <template #default="{ row }">
              {{ formatDate(row.createdAt) }}
            </template>
          </el-table-column>
          <el-table-column prop="operator" label="创建人" width="120" />
          <el-table-column label="操作" width="280" fixed="right">
            <template #default="{ row }">
              <el-button
                v-if="row.status === 'running'"
                type="primary"
                size="small"
                @click="handleControl(row)"
              >
                指挥
              </el-button>
              <el-button
                v-if="row.status === 'running'"
                type="warning"
                size="small"
                @click="handlePause(row)"
              >
                暂停
              </el-button>
              <el-button
                v-if="row.status === 'paused'"
                type="success"
                size="small"
                @click="handleResume(row)"
              >
                恢复
              </el-button>
              <el-button
                v-if="['pending', 'paused'].includes(row.status)"
                type="danger"
                size="small"
                @click="handleTerminate(row)"
              >
                终止
              </el-button>
              <el-button
                v-if="row.status === 'completed'"
                type="info"
                size="small"
                @click="handleReport(row)"
              >
                报告
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 分页 -->
      <div class="pagination-bar">
        <el-pagination
          v-model:current-page="pagination.currentPage"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleFetchData"
          @current-change="handleFetchData"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'

interface DrillItem {
  id: number
  name: string
  templateName: string
  category: string
  status: 'pending' | 'running' | 'paused' | 'completed' | 'terminated'
  totalSteps: number
  completedSteps: number
  createdAt: number
  operator: string
}

const router = useRouter()

const loading = ref(false)
const tableData = ref<DrillItem[]>([])

const filterForm = reactive({
  name: '',
  status: '',
  category: ''
})

const pagination = reactive({
  currentPage: 1,
  pageSize: 20,
  total: 0
})

// 模拟数据
const mockData: DrillItem[] = [
  {
    id: 1,
    name: '数据库主从切换演练',
    templateName: 'DB 主从切换模板',
    category: 'disaster_recovery',
    status: 'running',
    totalSteps: 7,
    completedSteps: 3,
    createdAt: Date.now() - 86400000,
    operator: '张三'
  },
  {
    id: 2,
    name: '支付服务降级演练',
    templateName: '服务降级标准流程',
    category: 'degradation',
    status: 'completed',
    totalSteps: 5,
    completedSteps: 5,
    createdAt: Date.now() - 172800000,
    operator: '李四'
  },
  {
    id: 3,
    name: '紧急发布回滚演练',
    templateName: '发布回滚模板',
    category: 'rollback',
    status: 'pending',
    totalSteps: 6,
    completedSteps: 0,
    createdAt: Date.now() - 259200000,
    operator: '王五'
  }
]

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

// 获取进度百分比
const getProgressPercent = (row: DrillItem) => {
  if (row.totalSteps === 0) return 0
  return Math.round((row.completedSteps / row.totalSteps) * 100)
}

// 获取进度状态
const getProgressStatus = (row: DrillItem) => {
  if (row.status === 'completed') return 'success'
  if (row.status === 'terminated') return 'exception'
  if (row.status === 'paused') return 'warning'
  return undefined
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
    // 模拟 API 调用
    await new Promise(resolve => setTimeout(resolve, 500))
    tableData.value = mockData
    pagination.total = mockData.length
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  pagination.currentPage = 1
  handleFetchData()
}

// 重置
const handleReset = () => {
  filterForm.name = ''
  filterForm.status = ''
  filterForm.category = ''
  handleSearch()
}

// 排序变化
const handleSortChange = ({ prop, order }: { prop: string; order: string }) => {
  console.log('排序:', prop, order)
}

// 新建演练
const handleCreate = () => {
  router.push('/console/create')
}

// 查看详情
const handleView = (row: DrillItem) => {
  router.push(`/console/${row.id}`)
}

// 指挥控制
const handleControl = (row: DrillItem) => {
  router.push(`/console/${row.id}`)
}

// 暂停
const handlePause = async (row: DrillItem) => {
  try {
    await ElMessageBox.confirm(`确认暂停演练"${row.name}"？`, '提示', {
      type: 'warning'
    })
    ElMessage.success('演练已暂停')
    handleFetchData()
  } catch {
    // 取消
  }
}

// 恢复
const handleResume = async (row: DrillItem) => {
  try {
    await ElMessageBox.confirm(`确认恢复演练"${row.name}"？`, '提示', {
      type: 'info'
    })
    ElMessage.success('演练已恢复')
    handleFetchData()
  } catch {
    // 取消
  }
}

// 终止
const handleTerminate = async (row: DrillItem) => {
  try {
    await ElMessageBox.confirm(`确认终止演练"${row.name}"？此操作不可恢复`, '警告', {
      type: 'error',
      confirmButtonText: '确认终止',
      confirmButtonClass: 'el-button--danger'
    })
    ElMessage.success('演练已终止')
    handleFetchData()
  } catch {
    // 取消
  }
}

// 查看报告
const handleReport = (row: DrillItem) => {
  router.push(`/report?drillId=${row.id}`)
}

onMounted(() => {
  handleFetchData()
})
</script>

<style scoped>
.drill-list-page {
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
  font-family: 'Fira Sans', sans-serif;
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

.table-container {
  margin-bottom: 20px;
}

.pagination-bar {
  display: flex;
  justify-content: flex-end;
}

/* Element Plus 深色主题适配 */
:deep(.el-table) {
  --el-table-bg-color: transparent;
  --el-table-tr-bg-color: transparent;
  --el-table-header-bg-color: var(--color-secondary, #1E293B);
  --el-table-text-color: var(--color-foreground, #F8FAFC);
  --el-table-header-text-color: var(--color-muted-foreground, #94A3B8);
  --el-table-border-color: var(--color-border, #334155);
  --el-table-row-hover-bg-color: var(--color-secondary, #1E293B);
}

:deep(.el-table th) {
  background-color: var(--color-secondary, #1E293B) !important;
}

:deep(.el-pagination) {
  --el-pagination-text-color: var(--color-foreground, #F8FAFC);
  --el-pagination-button-color: var(--color-foreground, #F8FAFC);
  --el-pagination-hover-color: var(--color-primary, #0F172A);
}
</style>
