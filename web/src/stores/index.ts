import { createPinia } from 'pinia'
import type { App } from 'vue'

// 创建 pinia 实例
const pinia = createPinia()

// 安装 pinia
export function setupStore(app: App) {
  app.use(pinia)
}

// 导出 store
export * from './user'
export * from './websocket'

export default pinia
