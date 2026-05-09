import { createApp } from 'vue'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'

import App from './App.vue'
import { setupStore } from './stores'
import router from './router'
import { loadFonts } from './styles'

// 创建应用实例
const app = createApp(App)

// 安装 Pinia
setupStore(app)

// 安装 Vue Router
app.use(router)

// 安装 Element Plus
app.use(ElementPlus, {
  locale: zhCn,
  size: 'default',
})

// 注册 Element Plus 图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

// 加载字体
loadFonts()

// 挂载应用
app.mount('#app')

// 导出 app 实例 (用于测试)
export default app
