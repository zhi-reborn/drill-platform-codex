<template>
  <div class="data-table">
    <el-table
      :data="data"
      v-loading="loading"
      stripe
      border
      style="width: 100%"
      @selection-change="handleSelectionChange"
    >
      <el-table-column v-if="selectable" type="selection" width="48" />
      <el-table-column
        v-for="col in columns"
        :key="col.prop"
        :prop="col.prop"
        :label="col.label"
        :width="col.width"
        :min-width="col.minWidth"
        :formatter="col.formatter"
        :sortable="col.sortable"
        :align="col.align || 'left'"
      >
        <template v-if="col.slot" #default="scope">
          <slot :name="col.prop" :row="scope.row" />
        </template>
      </el-table-column>
      <el-table-column v-if="$slots.actions" label="操作" width="180" fixed="right" align="center">
        <template #default="scope">
          <slot name="actions" :row="scope.row" />
        </template>
      </el-table-column>
    </el-table>
    
    <div v-if="pagination" class="table-pagination">
      <el-pagination
        v-model:current-page="currentPage"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="emit('page-change', { page: currentPage, size: pageSize })"
        @size-change="emit('page-change', { page: 1, size: pageSize })"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

export interface TableColumn {
  prop: string
  label: string
  width?: number | string
  minWidth?: number | string
  sortable?: boolean | string
  align?: 'left' | 'center' | 'right'
  formatter?: (row: Record<string, unknown>) => string
  slot?: boolean
}

interface Props {
  columns: TableColumn[]
  data: Record<string, unknown>[]
  loading?: boolean
  selectable?: boolean
  pagination?: boolean
  total?: number
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  selectable: false,
  pagination: false,
  total: 0
})

const emit = defineEmits<{
  'selection-change': [selection: Record<string, unknown>[]]
  'page-change': [info: { page: number; size: number }]
}>()

const currentPage = ref(1)
const pageSize = ref(20)

function handleSelectionChange(selection: Record<string, unknown>[]) {
  emit('selection-change', selection)
}
</script>

<style lang="scss" scoped>
@use '@/styles/variables' as *;

.data-table {
  .table-pagination {
    margin-top: $spacing-base;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
