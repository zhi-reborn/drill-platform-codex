<template>
  <pre class="code-block"><code class="code-content">{{ code }}</code>
    <button v-if="showCopy" class="copy-btn" @click="copyCode">
      {{ copied ? '已复制' : '复制' }}
    </button>
  </pre>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface Props {
  code: string
  showCopy?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  showCopy: false
})

const copied = ref(false)

async function copyCode() {
  await navigator.clipboard.writeText(props.code)
  copied.value = true
  setTimeout(() => copied.value = false, 2000)
}
</script>

<style lang="scss" scoped>
@use '@/styles/variables' as *;

.code-block {
  position: relative;
  background: $bg-tertiary;
  border: 1px solid $border-color;
  border-radius: $radius-base;
  padding: $spacing-base;
  overflow-x: auto;
  
  .code-content {
    font-family: $font-family-mono;
    font-size: $font-size-sm;
    color: $text-primary;
    line-height: 1.7;
    white-space: pre-wrap;
    word-break: break-all;
  }
  
  .copy-btn {
    position: absolute;
    top: $spacing-sm;
    right: $spacing-sm;
    font-size: $font-size-xs;
    padding: 2px 8px;
    background: $bg-secondary;
    border: 1px solid $border-color;
    border-radius: $radius-sm;
    color: $text-secondary;
    cursor: pointer;
    
    &:hover {
      color: $color-primary;
      border-color: $color-primary;
    }
  }
}
</style>
