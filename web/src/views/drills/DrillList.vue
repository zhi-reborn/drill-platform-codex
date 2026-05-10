<template>
  <div class="drill-list-page">
    <div class="page-header">
      <h1 class="page-title">演练列表</h1>
      <el-button type="primary" @click="$router.push('/drills/create')">
        <el-icon><Plus /></el-icon>
        创建演练
      </el-button>
    </div>
    
    <DataTable
      :columns="columns"
      :data="data"
      :loading="loading"
      pagination
      :total="total"
      @page-change="handlePageChange"
    >
      <template #status="{ row }">
        <DrillStatusBadge :status="row.status" type="drill" />
      </template>
      
      <template #actions="{ row }">
        <el-button link type="primary" size="small">查看</el-button>
        <ActionConfirm
          message="确定要删除此演练吗？"
          danger
          size="small"
          @confirm="handleDelete(row)"
        >
          删除
        </ActionConfirm>
      </template>
    </DataTable>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import DataTable, { type TableColumn } from '@/components/common/DataTable.vue'
import DrillStatusBadge from '@/components/common/DrillStatusBadge.vue'
import ActionConfirm from '@/components/common/ActionConfirm.vue'

const loading = ref(false)
const total = ref(0)

const columns: TableColumn[] = [
  { prop: 'id', label: 'ID', width: 80 },
  { prop: 'title', label: '演练标题', minWidth: 200 },
  { prop: 'category', label: '类型', width: 100 },
  { prop: 'status', label: '状态', width: 100, slot: true },
  { prop: 'created_at', label: '创建时间', width: 180 },
]

const data = ref([])

function handlePageChange({ page, size }: { page: number; size: number }) {
  console.log('Page change:', page, size)
}

function handleDelete(row: Record<string, unknown>) {
  console.log('Delete:', row)
}
</script>

<style lang="scss" scoped>
@use '@/styles/variables' as *;
@use '@/styles/mixins' as *;

.drill-list-page {
  .page-header {
    @include flex-between;
    margin-bottom: $spacing-xl;
    
    .page-title {
      font-size: $font-size-2xl;
      font-weight: $font-weight-bold;
      color: $text-primary;
    }
  }
}
</style>
