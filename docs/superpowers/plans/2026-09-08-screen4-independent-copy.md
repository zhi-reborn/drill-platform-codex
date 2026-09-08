# 大屏4独立副本 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在指挥监控页增加大屏4入口，并提供与大屏2初始效果一致、前端代码双向隔离的独立页面。

**Architecture:** `ScreenView4.vue` 是 `ScreenView2.vue` 在实施时的完整文件副本，两者不互相导入；它们仅独立调用同一组后端 API 与 WebSocket 数据源。指挥端新增独立路由，监控页新增独立按钮和打开函数。

**Tech Stack:** Vue 3、TypeScript、Vue Router、Element Plus、Vitest、Vite

---

### Task 1: 用测试锁定入口、路由和页面隔离

**Files:**
- Create: `web/src/views/director/ScreenView4.test.ts`
- Test: `web/src/views/director/ScreenView4.test.ts`

- [ ] **Step 1: 写失败测试**

```ts
import { describe, expect, it } from 'vitest'
import { existsSync, readFileSync } from 'node:fs'
import { fileURLToPath, URL } from 'node:url'

const screen4Path = fileURLToPath(new URL('./ScreenView4.vue', import.meta.url))
const monitorPath = fileURLToPath(new URL('./MonitorView.vue', import.meta.url))
const routerPath = fileURLToPath(new URL('../../router/index.ts', import.meta.url))

describe('independent screen4 entry', () => {
  it('owns an independent page implementation', () => {
    expect(existsSync(screen4Path)).toBe(true)
    const source = readFileSync(screen4Path, 'utf8')
    expect(source).toContain('class="screen-root cyber-command-screen"')
    expect(source).toContain("from '@/api/modules/drill'")
    expect(source).not.toMatch(/import\s+ScreenView2|<ScreenView2/)
  })

  it('places the screen4 button after screen3 and opens the independent route', () => {
    const source = readFileSync(monitorPath, 'utf8')
    expect(source).toMatch(/@click="viewScreen3"[\s\S]*?大屏3[\s\S]*?@click="viewScreen4"[\s\S]*?大屏4/)
    expect(source).toContain('window.open(`/director/screen4/${drillId.value}`, \'_blank\')')
    expect(source).toContain('.screen-entry-aqua')
  })

  it('routes screen4 to its own component', () => {
    const source = readFileSync(routerPath, 'utf8')
    expect(source).toMatch(/path: 'screen4\/:id\(\\\\d\+\)'[\s\S]*?name: 'DirectorScreen4'[\s\S]*?ScreenView4\.vue/)
  })
})
```

- [ ] **Step 2: 运行测试并确认因功能缺失而失败**

Run: `cd web && npm test -- src/views/director/ScreenView4.test.ts`

Expected: FAIL，首个断言显示 `ScreenView4.vue` 不存在。

- [ ] **Step 3: 提交失败测试**

```bash
git add web/src/views/director/ScreenView4.test.ts
git commit -m "test: define independent screen4 behavior"
```

### Task 2: 创建与大屏2双向隔离的页面副本

**Files:**
- Create: `web/src/views/director/ScreenView4.vue`
- Reference: `web/src/views/director/ScreenView2.vue`

- [ ] **Step 1: 将当前大屏2文件完整复制为独立文件**

创建 `ScreenView4.vue`，内容逐字复制自实施时的 `ScreenView2.vue`。新文件不得导入、渲染或扩展 `ScreenView2.vue`；模板、脚本及两个 style 块全部保存在新文件中。

- [ ] **Step 2: 保持数据调用独立但来源相同**

确认新文件自身包含以下调用，不通过大屏2组件代理：

```ts
import { drillApi } from '@/api/modules/drill'

await drillApi.getDetail(drillId.value)
await drillApi.getSteps(drillId.value)
await drillApi.getLogs(drillId.value, 200)
```

- [ ] **Step 3: 暂不运行完整测试**

入口和路由尚未实现，Task 1 的测试仍应保持失败；进入 Task 3 完成最小实现。

### Task 3: 增加独立路由与美观入口

**Files:**
- Modify: `web/src/router/index.ts`
- Modify: `web/src/views/director/MonitorView.vue`
- Test: `web/src/views/director/ScreenView4.test.ts`

- [ ] **Step 1: 增加指挥端独立路由**

在 `DirectorScreen` 路由后添加：

```ts
{
  path: 'screen4/:id(\\d+)',
  name: 'DirectorScreen4',
  component: () => import('@/views/director/ScreenView4.vue'),
  meta: { title: '数据大屏4', icon: 'DataBoard', hidden: true },
},
```

- [ ] **Step 2: 在大屏3右侧增加入口**

```vue
<el-button class="screen-entry-button screen-entry-aqua" @click="viewScreen4">
  <el-icon><DataBoard /></el-icon>
  大屏4
</el-button>
```

- [ ] **Step 3: 增加合法演练 ID 防护和新标签页跳转**

```ts
function viewScreen4() {
  if (!isValidDrill.value) return
  window.open(`/director/screen4/${drillId.value}`, '_blank')
}
```

- [ ] **Step 4: 增加青绿色独立按钮样式**

```scss
.screen-entry-aqua {
  color: #08979c;
  border: 1px solid rgba(8, 151, 156, 0.42);
  background: linear-gradient(135deg, rgba(8, 151, 156, 0.07), rgba(54, 207, 201, 0.04));

  &:hover,
  &:focus {
    color: #ffffff;
    border-color: #08979c;
    background: linear-gradient(135deg, #08979c, #13c2c2);
    box-shadow: 0 4px 14px rgba(8, 151, 156, 0.3);
    transform: translateY(-1px);
  }

  &:active {
    transform: translateY(0);
    background: #006d75;
  }
}
```

- [ ] **Step 5: 运行测试并确认通过**

Run: `cd web && npm test -- src/views/director/ScreenView4.test.ts src/views/director/ScreenView2.test.ts`

Expected: PASS，两个测试文件全部通过。

- [ ] **Step 6: 提交实现**

```bash
git add web/src/router/index.ts web/src/views/director/MonitorView.vue web/src/views/director/ScreenView4.vue
git commit -m "feat: add independent screen4 view"
```

### Task 4: 完整验证

**Files:**
- Verify: `web/src/views/director/ScreenView4.vue`
- Verify: `web/src/views/director/ScreenView2.vue`
- Verify: `web/src/views/director/MonitorView.vue`
- Verify: `web/src/router/index.ts`

- [ ] **Step 1: 运行前端测试**

Run: `cd web && npm test`

Expected: PASS，零失败。

- [ ] **Step 2: 运行类型检查**

Run: `cd web && npm run typecheck`

Expected: exit 0。

- [ ] **Step 3: 运行生产构建**

Run: `cd web && npm run build`

Expected: exit 0。

- [ ] **Step 4: 检查补丁完整性与隔离性**

Run: `git diff --check && ! rg -n "ScreenView2" web/src/views/director/ScreenView4.vue`

Expected: exit 0，且无输出。

- [ ] **Step 5: 浏览器验证**

打开 `/director/monitor/90`，确认“大屏4”紧邻“大屏3”右侧；点击后确认新标签页 URL 为 `/director/screen4/90`，页面正常渲染，并与大屏2初始视觉及数据一致。
