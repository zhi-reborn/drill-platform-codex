<template>
  <el-button
    :type="danger ? 'danger' : type"
    :size="size"
    :disabled="disabled"
    @click="handleConfirm"
  >
    <slot />
  </el-button>
</template>

<script setup lang="ts">
import { ElMessageBox } from 'element-plus'

interface Props {
  title?: string
  message: string
  danger?: boolean
  type?: 'primary' | 'success' | 'warning' | 'danger' | 'info' | 'text' | 'default'
  size?: 'default' | 'small' | 'large'
  disabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  title: '',
  danger: false,
  type: 'primary',
  size: 'default',
  disabled: false
})

const emit = defineEmits<{ confirm: [] }>()

async function handleConfirm() {
  try {
    await ElMessageBox.confirm(
      props.message,
      props.title || (props.danger ? '危险操作确认' : '操作确认'),
      {
        confirmButtonText: '确认',
        cancelButtonText: '取消',
        type: props.danger ? 'warning' : 'info',
      }
    )
    emit('confirm')
  } catch {
    // User cancelled
  }
}
</script>
