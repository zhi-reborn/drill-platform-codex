# 大屏4当前环节待完成任务舱 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 用“主任务 + 候场矩阵”替换大屏4底部演练概览和执行日志，实时展示当前环节尚未完成的任务。

**Architecture:** 新建大屏4专属纯函数模块，根据所选阶段的业务环节与叶子任务状态生成目标环节、主任务、候场任务和空状态。`ScreenView4.vue` 只负责把现有响应式 `steps` 数据传入选择器并渲染任务舱，不修改大屏2、公共流程模块或跑道组件。

**Tech Stack:** Vue 3 Composition API、TypeScript、CSS、Vitest、Vite

---

## 文件结构

- Create: `web/src/views/director/screen4PendingTasks.ts` — 大屏4待完成任务选择与排序。
- Create: `web/src/views/director/screen4PendingTasks.test.ts` — 纯函数状态覆盖。
- Modify: `web/src/views/director/ScreenView4.vue` — 替换底部模板、接入选择器、增加任务舱样式并清理旧日志/概览逻辑。
- Modify: `web/src/views/director/ScreenView4.test.ts` — 固化页面独立性和新任务舱结构。

### Task 1: 大屏4专属任务选择器

**Files:**
- Create: `web/src/views/director/screen4PendingTasks.ts`
- Create: `web/src/views/director/screen4PendingTasks.test.ts`

- [ ] **Step 1: 写失败的行为测试**

使用最小 `StepInstance` 工厂构造阶段占位节点、完成环节、执行中环节和待执行环节，断言：

```ts
const result = buildScreen4PendingTaskPanel('实施阶段', phaseSteps, () => true)
expect(result).toMatchObject({
  state: 'ready',
  phaseStepName: '主数据库切换',
  primaryTask: { id: 21, status: 'running' },
})
expect(result.tasks.map(task => task.id)).toEqual([21, 22, 23])
expect(result.queuedTasks.map(task => task.id)).toEqual([22, 23])
```

补充断言：无 `running` 时回退第一个含 `pending` 的环节；终态任务被过滤；多个 `running` 按 `seq` 排在 `pending` 前；无业务叶子任务返回 `empty`；无待完成任务返回 `complete`。

- [ ] **Step 2: 运行测试并确认失败**

Run: `cd web && npm test -- src/views/director/screen4PendingTasks.test.ts`

Expected: FAIL，原因是 `screen4PendingTasks` 模块尚不存在。

- [ ] **Step 3: 实现纯函数**

```ts
import type { StepInstance } from '@/types/instance'

export interface Screen4PhaseStep {
  name: string
  stepNodes: StepInstance[]
}

export interface Screen4PendingTaskPanel {
  state: 'ready' | 'empty' | 'complete'
  phaseStepName: string
  tasks: StepInstance[]
  primaryTask: StepInstance | null
  queuedTasks: StepInstance[]
}

export function buildScreen4PendingTaskPanel(
  phaseName: string,
  phaseSteps: Screen4PhaseStep[],
  isLeafStep: (step: StepInstance) => boolean,
): Screen4PendingTaskPanel {
  const businessSteps = phaseSteps.filter(item => item.name !== phaseName)
  const candidates = businessSteps.map(item => ({
    ...item,
    tasks: item.stepNodes.filter(isLeafStep),
  }))
  const hasConfiguredTasks = candidates.some(item => item.tasks.length > 0)
  if (!hasConfiguredTasks) return { state: 'empty', phaseStepName: '', tasks: [], primaryTask: null, queuedTasks: [] }

  const target = candidates.find(item => item.tasks.some(task => task.status === 'running'))
    ?? candidates.find(item => item.tasks.some(task => task.status === 'pending'))
  if (!target) return { state: 'complete', phaseStepName: '', tasks: [], primaryTask: null, queuedTasks: [] }

  const tasks = target.tasks
    .filter(task => task.status === 'running' || task.status === 'pending')
    .sort((left, right) => Number(right.status === 'running') - Number(left.status === 'running') || left.seq - right.seq)
  return {
    state: 'ready',
    phaseStepName: target.name,
    tasks,
    primaryTask: tasks[0] ?? null,
    queuedTasks: tasks.slice(1),
  }
}
```

- [ ] **Step 4: 运行选择器测试**

Run: `cd web && npm test -- src/views/director/screen4PendingTasks.test.ts`

Expected: PASS，六类状态全部通过。

- [ ] **Step 5: 提交**

```bash
git add web/src/views/director/screen4PendingTasks.ts web/src/views/director/screen4PendingTasks.test.ts
git commit -m "feat: select screen4 pending tasks"
```

### Task 2: 替换底部信息区为任务舱

**Files:**
- Modify: `web/src/views/director/ScreenView4.vue`
- Modify: `web/src/views/director/ScreenView4.test.ts`

- [ ] **Step 1: 写失败的页面结构测试**

新增测试，要求页面导入选择器并包含以下稳定结构，同时不再包含旧区域：

```ts
expect(screen4Source).toContain("from './screen4PendingTasks'")
expect(screen4Source).toContain('class="pending-task-panel"')
expect(screen4Source).toContain('class="primary-task-card"')
expect(screen4Source).toContain('class="task-queue-grid"')
expect(screen4Source).toContain('当前阶段待完成任务已清零')
expect(screen4Source).not.toContain('class="flow-brief"')
expect(screen4Source).not.toContain('class="flow-log-panel"')
```

- [ ] **Step 2: 运行页面测试并确认失败**

Run: `cd web && npm test -- src/views/director/ScreenView4.test.ts`

Expected: FAIL，原因是新任务舱尚未渲染且旧区域仍存在。

- [ ] **Step 3: 接入响应式派生数据**

在 `ScreenView4.vue` 中导入 `buildScreen4PendingTaskPanel`，增加：

```ts
const pendingTaskPanel = computed(() => buildScreen4PendingTaskPanel(
  currentPhaseData.value?.name ?? '',
  currentPhaseData.value?.phaseSteps ?? [],
  isLeafStep,
))

function taskOwner(step: StepInstance): string {
  return step.executor_team || step.assignee_names || '待分配'
}

function taskStatusText(step: StepInstance): string {
  return step.status === 'running' ? '执行中' : '待执行'
}
```

- [ ] **Step 4: 替换模板**

将 `.flow-information` 整块替换为 `.pending-task-panel`：标题展示目标环节名和任务数；主任务卡展示序号、状态、名称、负责人；候场矩阵使用 `v-for="(task, index) in pendingTaskPanel.queuedTasks"`；`empty` 与 `complete` 分别显示“该阶段暂无任务配置”和“当前阶段待完成任务已清零”。所有任务名称保留 `title`，列表使用 `aria-live="polite"`。

- [ ] **Step 5: 完成任务舱视觉样式并清理旧逻辑**

实现青蓝全宽舱体、约 `0.72fr 1.8fr` 的主任务/候场布局、琥珀执行态、青蓝待执行态、局部滚动、任务入场动画和 `prefers-reduced-motion`。删除旧概览与日志的模板专属计算、滚动观察器、日志拉取/缓存及对应样式；WebSocket 仍刷新 `drill` 和 `steps`，任务完成弹窗继续由步骤事件触发。

- [ ] **Step 6: 运行相关测试**

Run: `cd web && npm test -- src/views/director/ScreenView4.test.ts src/views/director/screen4PendingTasks.test.ts src/views/director/screen4Runway.test.ts src/views/director/screenPhaseFlow.test.ts`

Expected: PASS，无失败测试。

- [ ] **Step 7: 提交**

```bash
git add web/src/views/director/ScreenView4.vue web/src/views/director/ScreenView4.test.ts
git commit -m "feat: add screen4 pending task panel"
```

### Task 3: 全量与浏览器验证

**Files:**
- Verify: `web/src/views/director/ScreenView4.vue`
- Verify: `web/src/views/director/screen4PendingTasks.ts`

- [ ] **Step 1: 运行完整前端验证**

Run: `cd web && npm test && npm run typecheck && npm run build`

Expected: 所有测试通过、`vue-tsc` 无错误、Vite 构建退出码为 0。

- [ ] **Step 2: 清理构建生成噪声并检查差异**

恢复构建自动更新但与需求无关的 `web/src/types/components.d.ts` 与 `web/tsconfig.tsbuildinfo`，然后运行：

```bash
git diff --check
git status --short
```

Expected: 无空白错误，工作区没有未提交变更。

- [ ] **Step 3: 浏览器实页验证**

打开 `http://localhost:5173/director/screen4/90` 并检查：当前“主数据库切换”环节显示执行中任务主卡和两项候场任务；旧演练概览与执行日志消失；任务名、状态、人员和计数没有截断或溢出；上方跑道与阶段数字不回归。

- [ ] **Step 4: 检查独立性边界**

确认本功能提交未修改 `ScreenView2.vue`、`screenPhaseFlow.ts` 或 `Screen4Runway.vue`，并记录最终提交号与验证结果。
