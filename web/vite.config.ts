import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
import { resolve } from 'path'
import { fileURLToPath, pathToFileURL } from 'url'
import fs from 'fs'

// 自定义 SCSS importer 用于解析@别名
const scssAliasImporter = {
  canonicalize(url: string) {
    if (url.startsWith('@/')) {
      const resolved = resolve(__dirname, 'src', url.slice(2))
      return new URL(pathToFileURL(resolved).href)
    }
    return null
  },
  load(canonicalUrl: URL) {
    const contents = fs.readFileSync(fileURLToPath(canonicalUrl), 'utf-8')
    return { contents, syntax: 'scss' as const }
  },
}

export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      imports: ['vue', 'pinia', 'vue-router'],
      resolvers: [
        ElementPlusResolver(),
        (name) => {
          if (name === 'ElMessage') return { name: 'ElMessage', from: 'element-plus' }
          if (name === 'ElMessageBox') return { name: 'ElMessageBox', from: 'element-plus' }
          if (name === 'ElNotification') return { name: 'ElNotification', from: 'element-plus' }
          return undefined
        },
      ],
      dts: 'src/types/auto-imports.d.ts',
    }),
    Components({
      resolvers: [
        ElementPlusResolver(),
      ],
      dirs: ['src/components'],
      dts: 'src/types/components.d.ts',
    }),
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://localhost:8081',
        ws: true,
      },
    },
  },
  css: {
    preprocessorOptions: {
      scss: {
        // 使用自定义 importer 解析@别名
        importer: scssAliasImporter,
      },
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['vue', 'vue-router', 'pinia', 'echarts'],
          'element-plus': ['element-plus'],
        },
      },
    },
  },
})
