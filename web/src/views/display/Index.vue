<template>
  <div class="display-page">
    <!-- 顶部状态栏 -->
    <header class="display-header">
      <div class="header-left">
        <h1 class="header-title">生产演练指挥中心</h1>
        <span class="header-subtitle">Drill Command Center</span>
      </div>
      <div class="header-center">
        <div class="current-drill-info" v-if="currentDrill">
          <span class="drill-name">{{ currentDrill.name }}</span>
          <el-tag :type="getDrillStatusType(currentDrill.status)" size="large">
            {{ getDrillStatusText(currentDrill.status) }}
          </el-tag>
        </div>
      </div>
      <div class="header-right">
        <CountdownTimer
          v-if="currentDrill?.endTime"
          :endTime="currentDrill.endTime"
          :show-progress="true"
          :total-duration="currentDrill.totalDuration"
        />
        <span class="current-time">{{ currentTime }}</span>
      </div>
    </header>

    <!-- 主内容区 -->
    <main class="display-main">
      <!-- 左侧：流程图 -->
      <section class="main-panel flow-panel">
        <div class="panel-header">
          <h2>演练流程</h2>
          <div class="panel-stats">
            <span class="stat-item">
              <span class="stat-label">总步骤</span>
              <span class="stat-value">{{ totalSteps }}</span>
            </span>
            <span class="stat-item">
              <span class="stat-label">已完成</span>
              <span class="stat-value completed">{{ completedSteps }}</span>
            </span>
            <span class="stat-item">
              <span class="stat-label">执行中</span>
              <span class="stat-value running">{{ runningSteps }}</span>
            </span>
          </div>
        </div>
        <div class="panel-content">
          <DrillFlowChart
            ref="flowChartRef"
            :nodes="flowNodes"
            :links="flowLinks"
            :width="flowChartWidth"
            :height="flowChartHeight"
          />
        </div>
      </section>

      <!-- 右侧：信息和消息 -->
      <section class="side-panel">
        <!-- 关键指标 -->
        <div class="metrics-panel">
          <div class="metric-card">
            <div class="metric-icon success">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="20 6 9 17 4 12" />
              </svg>
            </div>
            <div class="metric-info">
              <span class="metric-value">{{ successRate }}%</span>
              <span class="metric-label">成功率</span>
            </div>
          </div>
          <div class="metric-card">
            <div class="metric-icon warning">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10" />
                <line x1="12" y1="8" x2="12" y2="12" />
                <line x1="12" y1="16" x2="12.01" y2="16" />
              </svg>
            </div>
            <div class="metric-info">
              <span class="metric-value">{{ issueCount }}</span>
              <span class="metric-label">异常数</span>
            </div>
          </div>
          <div class="metric-card">
            <div class="metric-icon info">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10" />
                <polyline points="12 6 12 12 16 14" />
              </svg>
            </div>
            <div class="metric-info">
              <span class="metric-value">{{ avgDuration }}</span>
              <span class="metric-label">平均耗时 (分钟)</span>
            </div>
          </div>
        </div>

        <!-- 消息滚动 -->
        <div class="messages-panel">
          <MessageTicker
            :messages="messages"
            title="实时消息"
            :max-messages="50"
            @clear="handleClearMessages"
          />
        </div>
      </section>
    </main>

    <!-- 底部：参演人员 -->
    <footer class="display-footer">
      <div class="footer-title">参演人员</div>
      <div class="team-members">
        <UserAvatar
          v-for="member in teamMembers"
          :key="member.id"
          :name="member.name"
          :src="member.avatar"
          :status="member.status"
          :show-status="true"
          size="md"
          :title="`${member.name} - ${member.role}`"
        />
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import DrillFlowChart from '@/components/DrillFlowChart.vue'
import CountdownTimer from '@/components/CountdownTimer.vue'
import MessageTicker from '@/components/MessageTicker.vue'
import UserAvatar from '@/components/UserAvatar.vue'
import { ElTag } from 'element-plus'

interface FlowNode {
  id: string | number
  name: string
  status: 'pending' | 'running' | 'completed' | 'issue' | 'skipped'
  x: number
  y: number
}

interface FlowLink {
  source: string | number
  target: string | number
}

interface MessageItem {
  id: string | number
  type: 'info' | 'success' | 'warning' | 'error'
  content: string
  timestamp: number
}

interface TeamMember {
  id: number
  name: string
  avatar?: string
  status: 'online' | 'busy' | 'away' | 'offline'
  role: string
}

interface DrillInfo {
  id: number
  name: string
  status: 'pending' | 'running' | 'paused' | 'completed' | 'terminated'
  endTime?: number
  totalDuration?: number
}

// 当前演练信息 (模拟数据，实际应从 API 获取)
const currentDrill = ref<DrillInfo | null>({
  id: 1,
  name: '数据库主从切换演练',
  status: 'running',
  endTime: Date.now() + 3600000, // 1 小时后结束
  totalDuration: 7200 // 2 小时
})

// 当前时间
const currentTime = ref('')
const updateTime = () => {
  const now = new Date()
  currentTime.value = now.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  })
}

// 流程图数据 (模拟)
const flowNodes = ref<FlowNode[]>([
  { id: 1, name: '开始', status: 'completed', x: 100, y: 300 },
  { id: 2, name: '备份检查', status: 'completed', x: 250, y: 300 },
  { id: 3, name: '主库降级', status: 'running', x: 400, y: 200 },
  { id: 4, name: '从库升级', status: 'pending', x: 400, y: 400 },
  { id: 5, name: '流量切换', status: 'pending', x: 550, y: 300 },
  { id: 6, name: '验证测试', status: 'pending', x: 700, y: 300 },
  { id: 7, name: '结束', status: 'pending', x: 850, y: 300 }
])

const flowLinks = ref<FlowLink[]>([
  { source: 1, target: 2 },
  { source: 2, target: 3 },
  { source: 2, target: 4 },
  { source: 3, target: 5 },
  { source: 4, target: 5 },
  { source: 5, target: 6 },
  { source: 6, target: 7 }
])

// 流程图尺寸
const flowChartWidth = ref(1000)
const flowChartHeight = ref(500)

// 消息列表
const messages = ref<MessageItem[]>([
  { id: 1, type: 'info', content: '演练已开始，所有参演人员就位', timestamp: Date.now() - 300000 },
  { id: 2, type: 'success', content: '备份检查完成，数据一致性验证通过', timestamp: Date.now() - 240000 },
  { id: 3, type: 'info', content: '步骤【备份检查】已完成，耗时 2 分钟', timestamp: Date.now() - 230000 },
  { id: 4, type: 'info', content: '开始执行步骤【主库降级】', timestamp: Date.now() - 180000 },
  { id: 5, type: 'warning', content: '主库响应延迟升高，持续监控中', timestamp: Date.now() - 60000 }
])

// 参演人员
const teamMembers = ref<TeamMember[]>([
  { id: 1, name: '张三', avatar: '', status: 'online', role: '指挥员' },
  { id: 2, name: '李四', avatar: '', status: 'online', role: 'DBA' },
  { id: 3, name: '王五', avatar: '', status: 'busy', role: '运维' },
  { id: 4, name: '赵六', avatar: '', status: 'online', role: '开发' },
  { id: 5, name: '钱七', avatar: '', status: 'away', role: '测试' }
])

// 统计数据
const totalSteps = computed(() => flowNodes.value.length)
const completedSteps = computed(() => flowNodes.value.filter(n => n.status === 'completed').length)
const runningSteps = computed(() => flowNodes.value.filter(n => n.status === 'running').length)
const successRate = computed(() => {
  const completed = completedSteps.value
  if (completed === 0) return 100
  return Math.round((completed / totalSteps.value) * 100)
})
const issueCount = computed(() => flowNodes.value.filter(n => n.status === 'issue').length)
const avgDuration = computed(() => 3)

// 获取演练状态类型
const getDrillStatusType = (status: string) => {
  const typeMap: Record<string, any> = {
    pending: 'info',
    running: 'success',
    paused: 'warning',
    completed: 'success',
    terminated: 'danger'
  }
  return typeMap[status] || 'info'
}

// 获取演练状态文本
const getDrillStatusText = (status: string) => {
  const textMap: Record<string, string> = {
    pending: '待命',
    running: '执行中',
    paused: '已暂停',
    completed: '已完成',
    terminated: '已终止'
  }
  return textMap[status] || '未知'
}

// 清空消息
const handleClearMessages = () => {
  messages.value = []
}

// 定时器
let timeInterval: number | null = null

onMounted(() => {
  updateTime()
  timeInterval = window.setInterval(updateTime, 1000)
  
  // 模拟 WebSocket 消息推送
  setTimeout(() => {
    messages.value.push({
      id: Date.now(),
      type: 'info',
      content: '步骤【主库降级】执行中，预计剩余 5 分钟',
      timestamp: Date.now()
    })
  }, 5000)
})

onBeforeUnmount(() => {
  if (timeInterval !== null) {
    clearInterval(timeInterval)
  }
})
</script>

<style scoped>
.display-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background-color: var(--color-background, #020617);
  color: var(--color-foreground, #F8FAFC);
  font-family: 'Fira Sans', sans-serif;
  overflow: hidden;
}

/* 顶部状态栏 */
.display-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background-color: var(--color-secondary, #1E293B);
  border-bottom: 1px solid var(--color-border, #334155);
  flex-shrink: 0;
}

.header-left {
  display: flex;
  flex-direction: column;
}

.header-title {
  font-size: 24px;
  font-weight: 700;
  margin: 0;
  color: var(--color-foreground, #F8FAFC);
  font-family: 'Fira Code', monospace;
}

.header-subtitle {
  font-size: 12px;
  color: var(--color-muted-foreground, #94A3B8);
  margin-top: 4px;
}

.header-center {
  flex: 1;
  display: flex;
  justify-content: center;
}

.current-drill-info {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 8px 24px;
  background-color: var(--color-muted, #1A1E2F);
  border-radius: 8px;
  border: 1px solid var(--color-border, #334155);
}

.drill-name {
  font-size: 16px;
  font-weight: 600;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 24px;
}

.current-time {
  font-size: 14px;
  color: var(--color-muted-foreground, #94A3B8);
  font-family: 'Fira Code', monospace;
}

/* 主内容区 */
.display-main {
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 400px;
  gap: 16px;
  padding: 16px;
  overflow: hidden;
}

/* 主面板 */
.main-panel {
  display: flex;
  flex-direction: column;
  background-color: var(--color-muted, #1A1E2F);
  border-radius: 8px;
  border: 1px solid var(--color-border, #334155);
  overflow: hidden;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--color-border, #334155);
}

.panel-header h2 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
}

.panel-stats {
  display: flex;
  gap: 24px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-label {
  font-size: 11px;
  color: var(--color-muted-foreground, #94A3B8);
}

.stat-value {
  font-size: 20px;
  font-weight: 700;
  font-family: 'Fira Code', monospace;
}

.stat-value.completed { color: #22C55E; }
.stat-value.running { color: #3B82F6; }

.panel-content {
  flex: 1;
  padding: 16px;
  overflow: hidden;
}

/* 侧边栏 */
.side-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* 指标面板 */
.metrics-panel {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.metric-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background-color: var(--color-secondary, #1E293B);
  border-radius: 8px;
  border: 1px solid var(--color-border, #334155);
}

.metric-icon {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
}

.metric-icon svg {
  width: 24px;
  height: 24px;
}

.metric-icon.success {
  background-color: rgba(34, 197, 94, 0.1);
  color: #22C55E;
}

.metric-icon.warning {
  background-color: rgba(245, 158, 11, 0.1);
  color: #F59E0B;
}

.metric-icon.info {
  background-color: rgba(59, 130, 246, 0.1);
  color: #3B82F6;
}

.metric-info {
  display: flex;
  flex-direction: column;
}

.metric-value {
  font-size: 24px;
  font-weight: 700;
  font-family: 'Fira Code', monospace;
}

.metric-label {
  font-size: 12px;
  color: var(--color-muted-foreground, #94A3B8);
}

/* 消息面板 */
.messages-panel {
  flex: 1;
  min-height: 0;
}

.messages-panel :deep(.message-ticker) {
  height: 100%;
}

/* 底部参演人员 */
.display-footer {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 24px;
  background-color: var(--color-secondary, #1E293B);
  border-top: 1px solid var(--color-border, #334155);
  flex-shrink: 0;
}

.footer-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-muted-foreground, #94A3B8);
}

.team-members {
  display: flex;
  gap: 8px;
}
</style>
