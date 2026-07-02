import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'
import assert from 'node:assert/strict'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const router = readFileSync(resolve(root, 'src/router/index.ts'), 'utf8')
const screen2 = readFileSync(resolve(root, 'src/views/director/ScreenView2.vue'), 'utf8')
const screen3 = readFileSync(resolve(root, 'src/views/ScreenView3.vue'), 'utf8')
const adminDashboard = readFileSync(resolve(root, 'src/views/admin/DashboardView.vue'), 'utf8')

assert.ok(router.includes("path: '/screen/:id(\\\\d+)'"), '大屏路由必须保留 /screen/:id')
assert.ok(router.includes("component: () => import('@/views/ScreenView.vue')"), '大屏路由必须继续使用 ScreenView.vue')
assert.ok(router.includes("path: '/screen3/:id(\\\\d+)'"), '大屏3路由必须保留 /screen3/:id')
assert.ok(router.includes("component: () => import('@/views/ScreenView3.vue')"), '大屏3路由必须使用独立 ScreenView3.vue')
assert.match(router, /name:\s*'AdminScreen'[\s\S]*?component:\s*\(\)\s*=>\s*import\('@\/views\/director\/ScreenView2\.vue'\)/, 'Admin 大屏2 路由必须指向 ScreenView2.vue')
assert.ok(router.includes("path: 'screen/:id(\\\\d+)'"), '大屏2 子路由必须保留 screen/:id')
assert.ok(router.includes("component: () => import('@/views/director/ScreenView2.vue')"), '大屏2 路由必须继续使用 ScreenView2.vue')

for (const api of ['getDetail', 'getSteps', 'getLogs']) {
  assert.match(screen2, new RegExp(`drillApi\\.${api}\\(drillId\\.value\\)`), `ScreenView2 必须复用 drillApi.${api}`)
  assert.match(screen3, new RegExp(`drillApi\\.${api}\\(drillId\\.value\\)`), `ScreenView3 必须复用 drillApi.${api}`)
}

assert.match(screen2, /class="[^"]*\bcyber-command-screen\b[^"]*"/, 'ScreenView2 应呈现新版指挥中心大屏外层')
assert.match(screen2, /class="phase-card-strip"/, 'ScreenView2 应呈现参考图中的阶段卡片区')
assert.match(screen2, /class="execution-carousel"/, 'ScreenView2 应呈现参考图中的执行中步骤横向卡片区')
assert.match(adminDashboard, /大屏2[\s\S]*?viewScreen3\(drill\.id\)[\s\S]*?大屏3/, 'Admin 系统概览必须在大屏2后提供大屏3入口')
