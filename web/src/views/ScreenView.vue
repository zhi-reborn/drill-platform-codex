<template>
  <ScreenLayout>
    <div class="screen-content">
      <!-- 顶部 KPI 栏 (80px) -->
      <MetricsBar :metrics="kpiMetrics" />

      <!-- 主体区域 -->
      <div class="screen-body">
        <!-- 左侧面板 25%：活跃演练 -->
        <StatusPanel :drills="activeDrills" />

        <!-- 中间区域 55%：图表 -->
        <div class="charts-area">
          <!-- 第一行：三个小图 200px -->
          <div class="chart-row chart-row--top">
            <div class="chart-card">
              <h3 class="chart-title">演练成功率</h3>
              <GaugeChart
                :data="{ name: '成功率', value: stats.success_rate, max: 100 }"
                height="160px"
              />
            </div>
            <div class="chart-card">
              <h3 class="chart-title">步骤状态分布</h3>
              <PieChart :data="stepStatusData" height="160px" />
            </div>
            <div class="chart-card">
              <h3 class="chart-title">各分类演练数</h3>
              <BarChart :data="categoryData" height="160px" />
            </div>
          </div>
          <!-- 第二行：大图 300px -->
          <div class="chart-row chart-row--bottom">
            <div class="chart-card chart-card--wide">
              <h3 class="chart-title">演练步骤时间线</h3>
              <TimelineChart :data="timelineData" height="280px" />
            </div>
          </div>
        </div>

        <!-- 右侧面板 20%：告警流 -->
        <AlertFeed :alerts="alerts" />
      </div>
    </div>
  </ScreenLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import ScreenLayout from '@/components/screen/ScreenLayout.vue'
import MetricsBar from '@/components/screen/MetricsBar.vue'
import StatusPanel from '@/components/screen/StatusPanel.vue'
import AlertFeed from '@/components/screen/AlertFeed.vue'
import GaugeChart from '@/components/charts/GaugeChart.vue'
import PieChart from '@/components/charts/PieChart.vue'
import BarChart from '@/components/charts/BarChart.vue'
import TimelineChart from '@/components/charts/TimelineChart.vue'
import type { DrillInstance } from '@/types'

// Mock data imports
import dashboardData from '@/mock/data/dashboard.json'
import instancesData from '@/mock/data/instances.json'
import stepsData from '@/mock/data/steps.json'
import notificationsData from '@/mock/data/notifications.json'

const refreshTimer = ref<number>()
const stats = ref(dashboardData.stats)

// === KPI ===
const kpiMetrics = computed(() => [
  { label: '总演练', value: stats.value.total_drills, icon: 'Monitor' },
  { label: '活跃', value: stats.value.active_drills, icon: 'VideoPlay' },
  { label: '成功率', value: `${stats.value.success_rate}%`, icon: 'CircleCheck' },
  { label: '失败率', value: `${stats.value.failure_rate}%`, icon: 'CircleClose' },
  { label: '平均耗时', value: `${Math.round(stats.value.avg_step_duration_seconds / 60)}min`, icon: 'Timer' },
  { label: '团队在线', value: `${stats.value.team_online_count}/${stats.value.team_total_count}`, icon: 'User' },
])

// === 活跃演练 ===
const activeDrills = computed(() => {
  return (instancesData as DrillInstance[]).filter(d => d.status === 'running' || d.status === 'paused')
})

// === 步骤状态分布 ===
const stepStatusData = computed(() => {
  const steps = stepsData as Array<Record<string, unknown>>
  const counts: Record<string, number> = {}
  steps.forEach(s => {
    const status = s.status as string
    counts[status] = (counts[status] || 0) + 1
  })
  return Object.entries(counts).map(([name, value]) => ({ name, value }))
})

// === 各分类演练数 ===
const categoryData = computed(() => {
  const labels: Record<string, string> = {
    disaster_recovery: '灾备切换',
    degradation: '服务降级',
    release: '发布演练',
    security: '安全事件',
  }
  return dashboardData.by_category.map((c: { category: string; count: number }) => ({
    name: labels[c.category] || c.category,
    value: c.count,
  }))
})

// === 演练时间线 ===
const timelineData = computed(() => {
  const grouped: Record<number, Array<Record<string, unknown>>> = {}
  const steps = stepsData as Array<Record<string, unknown>>
  steps.forEach(s => {
    const drillId = s.drill_id as number
    if (!grouped[drillId]) grouped[drillId] = []
    grouped[drillId].push(s)
  })
  return Object.entries(grouped).map(([drillId, items]) => ({
    name: `演练 #${drillId}`,
    items: items.map(s => ({
      name: s.step_name as string,
      startTime: (s.started_at as string) || '',
      endTime: (s.completed_at as string) || '',
      status: (s.status as 'pending' | 'running' | 'completed' | 'timeout' | 'skipped' | 'issue') || 'pending',
    })),
  }))
})

// === 告警流 (丰富数据) ===
const alerts = computed(() => {
  const items: Array<{ id: number; level: 'info' | 'warning' | 'error'; message: string; icon: string; created_at: string }> = []

  // 1. 从 steps 中提取 issue 和 timeout
  const steps = stepsData as Array<Record<string, unknown>>
  steps.forEach((s, i) => {
    if (s.status === 'issue') {
      items.push({
        id: i,
        level: 'error',
        message: `步骤「${s.step_name}」异常: ${s.error_message || '未知错误'}`,
        icon: 'Warning',
        created_at: (s.started_at as string) || '2024-12-20T10:04:00Z',
      })
    }
    if (s.status === 'timeout') {
      items.push({
        id: i + 1000,
        level: 'warning',
        message: `步骤「${s.step_name}」超时`,
        icon: 'Clock',
        created_at: (s.started_at as string) || '2024-12-20T10:04:00Z',
      })
    }
  })

  // 2. 从通知中补充
  const notifs = notificationsData as Array<Record<string, unknown>>
  notifs.forEach((n, i) => {
    const type = n.type as string
    if (type === 'drill_started') {
      items.push({
        id: i + 2000,
        level: 'info',
        message: `演练「${n.title}」已启动`,
        icon: 'VideoPlay',
        created_at: n.created_at as string,
      })
    }
    if (type === 'drill_completed') {
      items.push({
        id: i + 3000,
        level: 'info',
        message: `演练「${n.title}」全部通过`,
        icon: 'CircleCheck',
        created_at: n.created_at as string,
      })
    }
    if (type === 'drill_paused') {
      items.push({
        id: i + 4000,
        level: 'warning',
        message: `演练「${n.title}」已暂停`,
        icon: 'VideoPause',
        created_at: n.created_at as string,
      })
    }
    if (type === 'drill_terminated') {
      items.push({
        id: i + 5000,
        level: 'error',
        message: `演练「${n.title}」已终止`,
        icon: 'CircleClose',
        created_at: n.created_at as string,
      })
    }
  })

  items.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
  return items.slice(0, 30)
})

// 自动刷新
function refreshData() {
  stats.value = dashboardData.stats
}

onMounted(() => {
  refreshTimer.value = window.setInterval(refreshData, 5000)
})

onBeforeUnmount(() => {
  if (refreshTimer.value) clearInterval(refreshTimer.value)
})
</script>

<style lang="scss" scoped>
@use '@/styles/variables' as *;

.screen-content {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

.screen-body {
  flex: 1;
  display: grid;
  grid-template-columns: 25% 55% 20%;
  overflow: hidden;
}

.charts-area {
  padding: 8px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.chart-row {
  display: flex;
  gap: 8px;

  &--top {
    flex: 0 0 auto;
  }

  &--bottom {
    flex: 1;
  }
}

.chart-card {
  flex: 1;
  background: $bg-secondary;
  border: 1px solid $border-color;
  border-radius: $radius-base;
  padding: 10px 12px;
  min-height: 0;
  overflow: hidden;

  &--wide {
    flex: 1 1 100%;
  }

  .chart-title {
    font-size: $font-size-xs;
    font-weight: $font-weight-semibold;
    color: $text-secondary;
    margin: 0 0 6px 0;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }
}
</style>
