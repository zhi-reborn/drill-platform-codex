<template>
  <div class="page-container">
    <div class="page-header">
      <h2 class="page-title">全部演练</h2>
    </div>

    <div class="page-content">
      <div class="table-header">
        <el-radio-group v-model="statusFilter" class="status-filter" @change="handleFilterChange">
          <el-radio-button value="">全部</el-radio-button>
          <el-radio-button value="pending">待启动</el-radio-button>
          <el-radio-button value="running">执行中</el-radio-button>
          <el-radio-button value="paused">已暂停</el-radio-button>
          <el-radio-button value="completed">已完成</el-radio-button>
          <el-radio-button value="terminated">已终止</el-radio-button>
        </el-radio-group>
        <div class="total-info">共 {{ total }} 条记录</div>
      </div>

      <el-table :data="filteredInstances" style="width: 100%" class="drills-table" v-loading="loading">
        <el-table-column prop="name" label="演练名" min-width="200" />
        <el-table-column prop="template_name" label="模板" width="180" />
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <DrillStatusBadge :status="row.status" type="drill" />
          </template>
        </el-table-column>
        <el-table-column prop="created_by_name" label="创建人" width="120" />
        <el-table-column label="进度" width="150">
          <template #default="{ row }">
            <el-progress
              :percentage="row.progress_pct || 0"
              :stroke-width="6"
              :status="row.status === 'completed' ? 'success' : undefined"
            />
          </template>
        </el-table-column>
        <el-table-column prop="started_at" label="开始时间" width="180">
          <template #default="{ row }">
            {{ row.started_at ? formatTime(row.started_at) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button text type="primary" @click="viewDetail(row)">
              查看详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="table-pagination">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="loadInstances"
          @size-change="loadInstances"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { DrillInstance } from '@/types'
import DrillStatusBadge from '@/components/common/DrillStatusBadge.vue'
import { drillApi } from '@/api/modules/drill'

const router = useRouter()
const statusFilter = ref('')
const instances = ref<DrillInstance[]>([])
const total = ref(0)
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)

const filteredInstances = computed(() => {
  if (!statusFilter.value) {
    return instances.value
  }
  return instances.value.filter(i => i.status === statusFilter.value)
})

function formatTime(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

async function loadInstances() {
  loading.value = true
  try {
    const data = await drillApi.getList({ page: page.value, page_size: pageSize.value, status: statusFilter.value || undefined })
    instances.value = data.list || []
    total.value = data.total || 0
  } catch (error: any) {
    ElMessage.error(error.message || '加载演练列表失败')
    console.error('Failed to load instances:', error)
  } finally {
    loading.value = false
  }
}

function viewDetail(instance: DrillInstance) {
  router.push(`/director/monitor/${instance.id}`)
}

function handleFilterChange() {
  page.value = 1
  loadInstances()
}

onMounted(() => {
  loadInstances()
})
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

    .table-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: $spacing-base;

      .status-filter {
        flex: 1;
      }

      .total-info {
        font-size: $font-size-sm;
        color: $text-secondary;
        margin-left: $spacing-md;
      }
    }

    .table-pagination {
      display: flex;
      justify-content: flex-end;
      margin-top: $spacing-base;
    }

    .drills-table {
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
</style>
