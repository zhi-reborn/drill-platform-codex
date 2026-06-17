<template>
  <div class="page-container">
    <div class="page-header">
      <h1 class="page-title">演练实例</h1>
      <el-button type="primary" @click="$router.push('/director/create')">
        <el-icon><Plus /></el-icon>
        创建演练
      </el-button>
    </div>

    <div class="page-content">
      <div class="table-toolbar">
        <div class="toolbar-left">
          <el-select v-model="filterStatus" placeholder="状态筛选" clearable style="width: 120px">
            <el-option label="待启动" value="pending" />
            <el-option label="执行中" value="running" />
            <el-option label="已暂停" value="paused" />
            <el-option label="已完成" value="completed" />
            <el-option label="已终止" value="terminated" />
          </el-select>
        </div>
        <div class="toolbar-right">
          <el-input v-model="keyword" placeholder="搜索演练名称" clearable style="width: 200px" @input="loadDrills">
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </div>
      </div>

      <el-table :data="drills" border v-loading="loading" style="width: 100%">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="name" label="演练名称" min-width="180" show-overflow-tooltip />
        <el-table-column prop="template_name" label="模板" min-width="150" show-overflow-tooltip />
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="getDrillStatusType(row.status)" size="small">
              {{ getDrillStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="progress_pct" label="进度" width="100" align="center">
          <template #default="{ row }">
            <el-progress :percentage="row.progress_pct" :status="row.status === 'completed' ? 'success' : undefined" />
          </template>
        </el-table-column>
        <el-table-column prop="created_by_name" label="创建人" width="100" />
        <el-table-column prop="created_at" label="创建时间" width="160">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="300" fixed="right">
          <template #default="{ row }">
            <el-button text type="primary" size="small" @click="viewMonitor(row)">
              详情
            </el-button>
            <el-button text type="success" size="small" @click="viewScreen(row)">
              大屏
            </el-button>
            <el-button text type="warning" size="small" @click="viewScreen2(row)">
              <el-icon><DataBoard /></el-icon>
              大屏2
            </el-button>
            <el-button text type="danger" size="small" @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.page_size"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadDrills"
          @current-change="loadDrills"
        />
      </div>
    </div>

    <el-dialog v-model="deleteVisible" title="确认删除" width="400px">
      <p>确定要删除演练「{{ deleteTarget?.name }}」吗？此操作不可撤销。</p>
      <template #footer>
        <el-button @click="deleteVisible = false">取消</el-button>
        <el-button type="danger" @click="confirmDelete">确认删除</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { DataBoard } from '@element-plus/icons-vue'
import type { DrillInstance } from '@/types'
import { drillApi } from '@/api/modules/drill'
import { DRILL_STATUS_LABELS } from '@/types/instance'

const router = useRouter()

const loading = ref(false)
const drills = ref<DrillInstance[]>([])
const filterStatus = ref('')
const keyword = ref('')
const deleteVisible = ref(false)
const deleteTarget = ref<DrillInstance | null>(null)

const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
})

async function loadDrills() {
  loading.value = true
  try {
    const params: any = {
      page: pagination.page,
      page_size: pagination.page_size,
    }
    if (filterStatus.value) {
      params.status = filterStatus.value
    }
    if (keyword.value) {
      params.keyword = keyword.value
    }
    const result = await drillApi.getList(params)
    drills.value = result.list || []
    pagination.total = result.total || 0
  } catch (error) {
    ElMessage.error('加载演练列表失败')
    console.error('Failed to load drills:', error)
  } finally {
    loading.value = false
  }
}

function getDrillStatusLabel(status: string): string {
  return DRILL_STATUS_LABELS[status as keyof typeof DRILL_STATUS_LABELS] || status
}

function getDrillStatusType(status: string): 'success' | 'warning' | 'danger' | 'info' {
  const map: Record<string, 'success' | 'warning' | 'danger' | 'info'> = {
    pending: 'info',
    running: 'warning',
    paused: 'warning',
    completed: 'success',
    terminated: 'danger',
  }
  return map[status] || 'info'
}

function formatTime(time: string): string {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function viewMonitor(row: DrillInstance) {
  router.push(`/director/monitor/${row.id}`)
}

function viewScreen(row: DrillInstance) {
  router.push(`/screen/${row.id}`)
}

function viewScreen2(row: DrillInstance) {
  router.push(`/director/screen/${row.id}`)
}

function handleDelete(row: DrillInstance) {
  deleteTarget.value = row
  deleteVisible.value = true
}

async function confirmDelete() {
  if (!deleteTarget.value) return
  try {
    await drillApi.delete(deleteTarget.value.id)
    ElMessage.success('演练已删除')
    loadDrills()
  } catch (error) {
    ElMessage.error('删除失败')
    console.error('Delete drill error:', error)
  } finally {
    deleteVisible.value = false
    deleteTarget.value = null
  }
}

onMounted(() => {
  loadDrills()
})
</script>

<style scoped lang="scss">
@use '@/styles/layout' as *;
@use '@/styles/variables' as *;

.page-container {
  @include page-container;

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: $spacing-base;

    .page-title {
      font-size: $font-size-xl;
      font-weight: $font-weight-bold;
      color: $text-primary;
      margin: 0;
    }
  }

  .page-content {
    @include page-content;

    .table-toolbar {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: $spacing-base;
    }

    .pagination-container {
      display: flex;
      justify-content: flex-end;
      margin-top: $spacing-base;
    }
  }
}
</style>
