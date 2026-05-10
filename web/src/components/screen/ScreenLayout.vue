<template>
  <div
    ref="containerRef"
    class="screen-layout"
    :class="{ 'cursor-hidden': cursorHidden }"
    @mousemove="handleMouseMove"
    @mouseleave="cursorHidden = true"
  >
    <slot />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'

const cursorHidden = ref(false)
const containerRef = ref<HTMLElement>()
let hideTimer: ReturnType<typeof setTimeout> | null = null

function resetHideTimer() {
  cursorHidden.value = false
  if (hideTimer) clearTimeout(hideTimer)
  hideTimer = setTimeout(() => { cursorHidden.value = true }, 3000)
}

function handleMouseMove() {
  resetHideTimer()
}

onMounted(() => resetHideTimer())
onBeforeUnmount(() => { if (hideTimer) clearTimeout(hideTimer) })

// ESC to exit fullscreen
function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    document.exitFullscreen?.()
  }
  if (e.key === 'F' || e.key === 'f') {
    containerRef.value?.requestFullscreen?.()
  }
}

onMounted(() => window.addEventListener('keydown', handleKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', handleKeydown))
</script>

<style scoped>
.screen-layout {
  width: 100vw;
  height: 100vh;
  background: #0D1117;
  overflow: hidden;
  position: relative;
  cursor: default;
}

.screen-layout.cursor-hidden {
  cursor: none;
}
</style>
