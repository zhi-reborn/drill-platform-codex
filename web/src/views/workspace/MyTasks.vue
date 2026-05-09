<template>
  <div class="my-tasks-page">
    <div class="page-header">
      <div class="header-left">
        <h1 class="page-title">我的任务</h1>
        <p class="page-subtitle">查看和处理分配给您的演练任务</p>
      </div>
      <div class="header-right">
        <el-radio-group v-model="filterType" size="default" @change="handleFetchData">
          <el-radio-button label="all">全部</el-radio-button>
          <el-radio-button label="pending">待处理</el-radio-button>
          <el-radio-button label="running">进行中</el-radio-button>
          <el-radio-button label="completed">已完成</el-radio-button>
        </el-radio-group>
      </div>
    </div>

    <div class="page-content">
      <!-- 任务列表 -->
      <div class="task-list" v-loading="loading">
        <div
          v-for="task in tasks"
          :key="task.id"
          class="task-card"
          :class="`task-${task.status}`"
        >
          <div class="task-header">
            <div class="task-info">
              <h3 class="task-name">{{ task.stepName }}</h3>
              <el-link
                type="primary"
                class="task-drill"
                @click="handleViewDrill(task.drillId)"
              >
                {{ task.drillName }}
              </el-link>
            </div>
            <el-tag :type="getStatusType(task.status)" size="large">
              {{ getStatusText(task.status) }}
            </el-tag>
          </div>

          <div class="task-body">
            <div class="task-detail">
              <div class="detail-row">
                <span class="detail-label">演练分类：</span>
                <el-tag size="small">{{ task.category }}</el-tag>
              </div>
              <div class="detail-row">
                <span class="detail-label">创建时间：</span>
                <span>{{ formatDate(task.createdAt) }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">要求完成时间：</span>
                <span :class="{ 'text-warning': isUrgent(task.dueAt) }">
                  {{ formatDate(task.dueAt) }}
                </span>
              </div>
              <div class="detail-row" v-if="task.description">
                <span class="detail-label">任务说明：</span>
                <span class="detail-value">{{ task.description }}</span>
              </div>
            </div>

            <div class="task-progress" v-if="task.status === 'running'">
              <div class="progress-label">
                <span>进度</span>
                <span>{{ task.progress }}%</span>
              </div>
              <el-progress :percentage="task.progress" :show-text="false" />
            </div>
          </div>

          <div class="task-footer">
            <div class="task-actions">
              <el-button
                v-if="task.status === 'pending'"
                type="primary"
                @click="handleStart(task)"
              >
                开始执行
              </el-button>
              <el-button
                v-if="task.status === 'running'"
                type="success"
                @click="handleComplete(task)"
              >
                完成任务
              </el-button>
              <el-button
                v-if="task.status === 'running'"
                type="warning"
                @click="handleIssue(task)"
              >
                报告异常
              </el-button>
              <el-button
                v-if="['pending', 'running'].includes(task.status)"
                @click="handleViewDetail(task)"
              >
                查看详情
              </el-button>
            </div>
            <div class="task-meta">
              <span class="meta-item" v-if="task.timeoutMinutes">
                <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="12" r="10" />
                  <polyline points="12 6 12 12 16 14" />
                </svg>
                时限：{{ task.timeoutMinutes }}分钟
              </span>
              <span class="meta-item">
                <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M22 11.08V12a10 10 0 11-5.93-9.14" />
                  <polyline points="22 4 12 14.01 9 11.01" />
                </svg>
                {{ task.priority === 'high' ? '高优先级' : '普通' }}
              </span>
            </div>
          </div>
        </div>

        <!-- 空状态 -->
        <div v-if="tasks.length === 0" class="empty-state">
          <svg class="empty-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
          </svg>
          <p class="empty-text">暂无任务</p>
          <p class="empty-hint">当有演练任务分配给您时，会显示在这里</p>
        </div>
      </div>

      <!-- 分页 -->
      <div class="pagination-bar" v-if="tasks.length > 0">
        <el-pagination
          v-model:current-page="pagination.currentPage"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleFetchData"
          @current-change="handleFetchData"
        />
      </div>
    </div>

    <!-- 任务详情对话框 -->
    <el-dialog
      v-model="dialogVisible"
      title="任务详情"
      width="600px"
      :close-on-click-modal="false"
    >
      <div v-if="selectedTask" class="dialog-content">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="任务名称">{{ selectedTask.stepName }}</el-descriptions-item>
          <el-descriptions-item label="所属演练">{{ selectedTask.drillName }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(selectedTask.status)">
              {{ getStatusText(selectedTask.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="优先级">
            <el-tag :type="selectedTask.priority === 'high' ? 'danger' : 'info'" size="small">
              {{ selectedTask.priority === 'high' ? '高' : '普通' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatDate(selectedTask.createdAt) }}</el-descriptions-item>
          <el-descriptions-item label="要求完成时间">{{ formatDate(selectedTask.dueAt) }}</el-descriptions-item>
          <el-descriptions-item label="任务说明" label-class-name="full-width">
            {{ selectedTask.description || '无' }}
          </el-descriptions-item>
        </el-descriptions>
      </div>
      <template #footer>
        <el-button @click="dialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'

interface Task {
  id: number
  stepName: string
  drillId: number
  drillName: string
  category: string
  status: 'pending' | 'running' | 'completed' | 'issue' | 'skipped'
  priority: 'normal' | 'high'
  progress: number
  description?: string
  timeoutMinutes?: number
  createdAt: number
  dueAt: number
}

const router = useRouter()

const loading = ref(false)
const filterType = ref('all')
const tasks = ref<Task[]>([])
const dialogVisible = ref(false)
const selectedTask = ref<Task | null>(null)

const pagination = reactive({
  currentPage: 1,
  pageSize: 20,
  total: 0
})

// 模拟数据
const mockTasks: Task[] = [
  {
    id: 1,
    stepName: '主库降级操作',
    drillId: 1,
    drillName: '数据库主从切换演练',
    category: '灾备切换',
    status: 'running',
    priority: 'high',
    progress: 60,
    description: '执行主库降级操作，确保从库可以正常接管',
    timeoutMinutes: 15,
    createdAt: Date.now() - 3600000,
    dueAt: Date.now() + 1800000
  },
  {
    id: 2,
    stepName: '服务降级确认',
    drillId: 2,
    drillName: '支付服务降级演练',
    category: '服务降级',
    status: 'pending',
    priority: 'normal',
    progress: 0,
    timeoutMinutes: 10,
    createdAt: Date.now() - 7200000,
    dueAt: Date.now() + 3600000
  },
  {
    id: 3,
    stepName: '流量切换验证',
    drillId: 1,
    drillName: '数据库主从切换演练',
    category: '灾备切换',
    status: 'completed',
    priority: 'high',
    progress: 100,
    createdAt: Date.now() - 86400000,
    dueAt: Date.now() - 43200000
  }
]

// 获取状态类型
const getStatusType = (status: string) => {
  const typeMap: Record<string, any> = {
    pending: 'info',
    running: 'warning',
    completed: 'success',
    issue: 'danger',
    skipped: ''
  }
  return typeMap[status] || 'info'
}

// 获取状态文本
const getStatusText = (status: string) => {
  const textMap: Record<string, string> = {
    pending: '待处理',
    running: '进行中',
    completed: '已完成',
    issue: '异常',
    skipped: '已跳过'
  }
  return textMap[status] || '未知'
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

// 是否紧急
const isUrgent = (dueAt: number) => {
  return dueAt - Date.now() < 1800000 // 30 分钟内
}

// 加载数据
const handleFetchData = async () => {
  loading.value = true
  try {
    await new Promise(resolve => setTimeout(resolve, 500))
    // 根据筛选条件过滤
    if (filterType.value === 'all') {
      tasks.value = mockTasks
    } else {
      const statusMap: Record<string, string> = {
        all: '',
        pending: 'pending',
        running: 'running',
        completed: 'completed'
      }
      tasks.value = mockTasks.filter(t => {
        if (filterType.value === 'pending') return ['pending', 'running'].includes(t.status)
        return t.status === statusMap[filterType.value]
      })
    }
    pagination.total = tasks.value.length
  } finally {
    loading.value = false
  }
}

// 查看演练
const handleViewDrill = (drillId: number) => {
  router.push(`/console/${drillId}`)
}

// 开始任务
const handleStart = async (task: Task) => {
  try {
    await ElMessageBox.confirm(`确认开始执行任务"${task.stepName}"？`, '提示', {
      type: 'info'
    })
    task.status = 'running'
    ElMessage.success('任务已开始')
  } catch {
    // 取消
  }
}

// 完成任务
const handleComplete = async (task: Task) => {
  try {
    await ElMessageBox.confirm(`确认完成任务"${task.stepName}"？`, '提示', {
      type: 'success'
    })
    task.status = 'completed'
    task.progress = 100
    ElMessage.success('任务已完成')
  } catch {
    // 取消
  }
}

// 报告异常
const handleIssue = async (task: Task) => {
  try {
    await ElMessageBox.prompt('请输入异常描述', '报告异常', {
      type: 'warning',
      inputPlaceholder: '描述遇到的问题...'
    })
    task.status = 'issue'
    ElMessage.warning('异常已报告，等待处理')
  } catch {
    // 取消
  }
}

// 查看详情
const handleViewDetail = (task: Task) => {
  selectedTask.value = task
  dialogVisible.value = true
}

onMounted(() => {
  handleFetchData()
})
</script>

<style scoped>
.my-tasks-page {
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

.page-content {
  background-color: var(--color-muted, #1A1E2F);
  border-radius: 8px;
  border: 1px solid var(--color-border, #334155);
  padding: 20px;
}

.task-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-bottom: 20px;
}

.task-card {
  padding: 20px;
  border-radius: 8px;
  border: 1px solid var(--color-border, #334155);
  background-color: var(--color-secondary, #1E293B);
  transition: all 0.2s ease;
}

.task-card:hover {
  border-color: var(--color-primary, #0F172A);
}

.task-card.task-pending {
  border-left: 3px solid #3B82F6;
}

.task-card.task-running {
  border-left: 3px solid #F59E0B;
}

.task-card.task-completed {
  border-left: 3px solid #22C55E;
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
}

.task-info {
  flex: 1;
}

.task-name {
  margin: 0 0 8px 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--color-foreground, #F8FAFC);
}

.task-drill {
  font-size: 14px;
}

.task-body {
  margin-bottom: 16px;
}

.task-detail {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
}

.detail-row {
  display: flex;
  gap: 12px;
  font-size: 14px;
}

.detail-label {
  color: var(--color-muted-foreground, #94A3B8);
  min-width: 100px;
}

.detail-value {
  color: var(--color-foreground, #F8FAFC);
}

.text-warning {
  color: #F59E0B;
}

.task-progress {
  margin-top: 12px;
}

.progress-label {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  color: var(--color-muted-foreground, #94A3B8);
  margin-bottom: 8px;
}

.task-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 16px;
  border-top: 1px solid var(--color-border, #334155);
}

.task-actions {
  display: flex;
  gap: 8px;
}

.task-meta {
  display: flex;
  gap: 16px;
  font-size: 13px;
  color: var(--color-muted-foreground, #94A3B8);
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

.empty-state {
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
}

.pagination-bar {
  display: flex;
  justify-content: flex-end;
}

.dialog-content {
  padding: 10px 0;
}

:deep(.el-descriptions__label) {
  width: 120px;
}

:deep(.el-descriptions__label.full-width) {
  width: auto;
}

:deep(.el-card),
:deep(.el-dialog) {
  background-color: var(--color-muted, #1A1E2F);
  border: 1px solid var(--color-border, #334155);
}

:deep(.el-dialog__title) {
  color: var(--color-foreground, #F8FAFC);
}

:deep(.el-descriptions) {
  --el-descriptions-body-bg-color: var(--color-secondary, #1E293B);
  --el-descriptions-header-bg-color: var(--color-secondary, #1E293B);
}
</style>
