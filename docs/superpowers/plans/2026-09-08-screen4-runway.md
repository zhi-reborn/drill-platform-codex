# 大屏4独立跑道 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将大屏4环节卡片流程替换为独立的大屏1风格蛇形能量跑道，同时保持大屏1、大屏2和大屏4实现互不影响。

**Architecture:** 新建纯 TypeScript 跑道几何模块和大屏4专属 Vue SVG 组件；`ScreenView4.vue` 只负责把现有实时 `flowNodes` 映射为组件 props。跑道不请求数据、不导入 `PhaseRing.vue`，大屏1和大屏2文件保持不变。

**Tech Stack:** Vue 3、TypeScript、SVG、SCSS、Vitest、Vue Compiler SFC

---

### Task 1: 跑道几何与状态算法

**Files:**
- Create: `web/src/views/director/screen4Runway.ts`
- Create: `web/src/views/director/screen4Runway.test.ts`

- [ ] **Step 1: 写失败测试**

```ts
import { describe, expect, it } from 'vitest'
import { buildScreen4RunwayLayout, getScreen4RunwayProgress, splitScreen4RunwayName } from './screen4Runway'

describe('screen4 runway geometry', () => {
  it('lays out a short runway on one forward lane with a milestone', () => {
    const layout = buildScreen4RunwayLayout(3)
    expect(layout.rowCount).toBe(1)
    expect(layout.points).toHaveLength(4)
    expect(layout.points.map(point => point.row)).toEqual([0, 0, 0, 0])
    expect(layout.points[0].x).toBeLessThan(layout.points[3].x)
  })

  it('folds a longer runway into alternating lanes', () => {
    const layout = buildScreen4RunwayLayout(7)
    expect(layout.rowCount).toBe(2)
    expect(layout.points).toHaveLength(8)
    expect(layout.points[4].direction).toBe('left')
    expect(layout.points[4].x).toBeGreaterThan(layout.points[5].x)
    expect(layout.trackPath).toContain('Q 998')
  })

  it('counts terminal task states without exceeding the total', () => {
    expect(getScreen4RunwayProgress(['done', 'skipped', 'issue', 'running', 'pending']))
      .toEqual({ completed: 3, total: 5, ratio: 0.6 })
  })

  it('keeps runway labels to two concise lines', () => {
    expect(splitScreen4RunwayName('缓存与消息队列切换')).toEqual(['缓存与消息', '队列切换'])
    expect(splitScreen4RunwayName('非常非常长的环节节点名称需要截断')).toEqual(['非常非常长的', '环节节点名…'])
  })
})
```

- [ ] **Step 2: 运行测试并确认模块缺失导致失败**

Run: `cd web && npm test -- src/views/director/screen4Runway.test.ts`

Expected: FAIL，提示无法解析 `./screen4Runway`。

- [ ] **Step 3: 实现纯算法模块**

```ts
export type Screen4RunwayStatus = 'done' | 'skipped' | 'running' | 'issue' | 'pending'
export type Screen4RunwayDirection = 'left' | 'right'

export interface Screen4RunwayPoint {
  index: number
  row: number
  x: number
  y: number
  direction: Screen4RunwayDirection
}

export interface Screen4RunwayLayout {
  rowCount: number
  viewBoxHeight: number
  laneY: number[]
  points: Screen4RunwayPoint[]
  trackPath: string
}

export interface Screen4RunwayNode {
  id: string
  name: string
  status: Screen4RunwayStatus
  completed: number
  total: number
}

export function buildScreen4RunwayLayout(nodeCount: number): Screen4RunwayLayout
export function getScreen4RunwayProgress(statuses: string[]): { completed: number; total: number; ratio: number }
export function splitScreen4RunwayName(name: string): string[]
export function getScreen4RunwayPath(points: Screen4RunwayPoint[], from: number, to: number): string
```

实现固定 1040 SVG 宽度、每行最多 4 个节点、真实节点后追加里程碑、奇数行反向和二次贝塞尔折返。终态为 `done`、`skipped`、`issue`，完成数限制在 `0..total`。

- [ ] **Step 4: 运行算法测试并确认通过**

Run: `cd web && npm test -- src/views/director/screen4Runway.test.ts`

Expected: PASS，4 项测试通过。

- [ ] **Step 5: 提交算法**

```bash
git add web/src/views/director/screen4Runway.ts web/src/views/director/screen4Runway.test.ts
git commit -m "feat: add screen4 runway geometry"
```

### Task 2: 大屏4专属 SVG 跑道组件

**Files:**
- Create: `web/src/views/director/Screen4Runway.vue`
- Create: `web/src/views/director/Screen4Runway.test.ts`

- [ ] **Step 1: 写失败模板测试**

```ts
import { describe, expect, it } from 'vitest'
import { compileTemplate, parse } from '@vue/compiler-sfc'
import source from './Screen4Runway.vue?raw'

const { descriptor } = parse(source)
const template = descriptor.template!.content
const styles = descriptor.styles.map(style => style.content).join('\n')

describe('screen4 runway component', () => {
  it('compiles an accessible serpentine energy runway', () => {
    expect(compileTemplate({ source: template, filename: 'Screen4Runway.vue', id: 'screen4-runway' }).errors).toEqual([])
    expect(template).toContain('aria-label="当前阶段环节能量跑道"')
    expect(template).toContain('class="runway-svg"')
    expect(template).toContain(':d="layout.trackPath"')
    expect(template).toContain('class="runway-node-count"')
    expect(template).toContain('里程碑')
  })

  it('owns its implementation and never renders task names', () => {
    expect(source).not.toContain('PhaseRing')
    expect(source).not.toContain('ScreenView2')
    expect(template).not.toContain('node-steps')
    expect(template).not.toContain('step.name')
  })

  it('provides distinct status colors and reduced motion', () => {
    expect(styles).toContain('.runway-node.is-completed')
    expect(styles).toContain('.runway-node.is-running')
    expect(styles).toContain('.runway-node.is-issue')
    expect(styles).toContain('@media (prefers-reduced-motion: reduce)')
  })
})
```

- [ ] **Step 2: 运行测试并确认组件缺失导致失败**

Run: `cd web && npm test -- src/views/director/Screen4Runway.test.ts`

Expected: FAIL，提示无法解析 `Screen4Runway.vue?raw`。

- [ ] **Step 3: 实现独立跑道组件**

组件 props 固定为：

```ts
const props = defineProps<{
  phaseName: string
  phaseStatus: string
  nodes: Screen4RunwayNode[]
}>()
```

组件使用 `buildScreen4RunwayLayout` 和 `getScreen4RunwayPath` 绘制以下 SVG 层：完整冷蓝跑道、已完成荧光绿路径、进行中琥珀路径、折返指示灯、普通状态节点、运行能量棒、名称双行文本、完成数胶囊和末端里程碑靶心。样式全部 scoped，并提供加载入场、能量流动、呼吸环、靶心扫描及 reduced-motion 降级。

- [ ] **Step 4: 运行组件与算法测试**

Run: `cd web && npm test -- src/views/director/Screen4Runway.test.ts src/views/director/screen4Runway.test.ts`

Expected: PASS，7 项测试通过。

- [ ] **Step 5: 提交组件**

```bash
git add web/src/views/director/Screen4Runway.vue web/src/views/director/Screen4Runway.test.ts
git commit -m "feat: build independent screen4 energy runway"
```

### Task 3: 替换大屏4环节流程区域

**Files:**
- Modify: `web/src/views/director/ScreenView4.vue:91-169`
- Modify: `web/src/views/director/ScreenView4.test.ts`

- [ ] **Step 1: 先扩展集成测试**

```ts
expect(screen4Source).toContain("import Screen4Runway from './Screen4Runway.vue'")
expect(screen4Source).toContain('<Screen4Runway')
expect(screen4Source).toContain(':phase-name="currentPhaseData?.name || \'当前阶段\'"')
expect(screen4Source).toContain(':nodes="runwayNodes"')
expect(screen4Source).not.toContain('ref="flowTrackRef"')
expect(screen4Source).not.toContain('class="node-steps"')
expect(screen4Source).not.toContain("from '@/components/screen/PhaseRing.vue'")
```

- [ ] **Step 2: 运行集成测试并确认旧流程图导致失败**

Run: `cd web && npm test -- src/views/director/ScreenView4.test.ts`

Expected: FAIL，提示缺少 `Screen4Runway` 导入和渲染。

- [ ] **Step 3: 替换模板并映射节点数据**

```vue
<Screen4Runway
  v-show="flowNodes.length"
  :phase-name="currentPhaseData?.name || '当前阶段'"
  :phase-status="selectedPhaseStatus"
  :nodes="runwayNodes"
/>
```

```ts
const runwayNodes = computed(() => flowNodes.value.map(node => {
  const progress = getScreen4RunwayProgress(node.steps.map(step => step.status))
  return {
    id: node.id,
    name: node.name,
    status: node.status as Screen4RunwayStatus,
    completed: progress.completed,
    total: progress.total,
  }
}))
```

移除只服务于旧横向卡片的 `flowViewportRef`、`flowTrackRef`、聚焦缩放、箭头几何、ResizeObserver、生命周期调用、相关 helper imports，以及 `.flow-viewport` 到虚拟端帽之间的旧样式；保留 `flow-board`、标题、网格、底部信息和信号灯样式。

- [ ] **Step 4: 运行大屏4和大屏1回归测试**

Run: `cd web && npm test -- src/views/director/ScreenView4.test.ts src/views/director/Screen4Runway.test.ts src/views/director/screen4Runway.test.ts src/components/screen/PhaseRing.test.ts src/views/director/ScreenView2.test.ts`

Expected: PASS，全部测试通过。

- [ ] **Step 5: 提交集成**

```bash
git add web/src/views/director/ScreenView4.vue web/src/views/director/ScreenView4.test.ts
git commit -m "feat: replace screen4 flow cards with runway"
```

### Task 4: 完整验证与浏览器检查

**Files:**
- Verify: `web/src/views/director/ScreenView4.vue`
- Verify: `web/src/views/director/Screen4Runway.vue`
- Verify: `web/src/views/director/screen4Runway.ts`

- [ ] **Step 1: 运行全量前端测试**

Run: `cd web && npm test`

Expected: PASS，零失败。

- [ ] **Step 2: 运行类型检查和生产构建**

Run: `cd web && npm run typecheck`

Run: `cd web && npm run build`

Expected: 两个命令均 exit 0。

- [ ] **Step 3: 检查改动边界**

Run: `git diff --check && ! git diff HEAD~3 -- web/src/views/ScreenView.vue web/src/components/screen/PhaseRing.vue web/src/views/director/ScreenView2.vue`

Expected: exit 0，且无输出。

- [ ] **Step 4: 浏览器验证**

在 `/director/screen4/90` 检查：顶部阶段切换不变；标记区域显示蛇形跑道；节点仅有环节名和完成数；实施阶段显示已完成绿、当前橙、待执行蓝与里程碑；底部概览及日志不变；浏览器应用错误日志为空。
