# 大屏2/大屏4环节统计口径修正 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 使大屏2和大屏4顶部阶段卡片只统计真实业务环节，并在保持两页独立的前提下精修数字展示。

**Architecture:** 大屏2和大屏4分别在自身 `phaseCards` 计算中过滤名称与阶段名相同的占位节点，不增加共享统计实现。两页使用相同但独立的模板和 CSS 精修，让分子、分隔符、分母和单位形成明确层级。

**Tech Stack:** Vue 3、TypeScript、CSS、Vitest、Vue Compiler SFC

---

### Task 1: 独立修正大屏2统计和数字层级

**Files:**
- Modify: `web/src/views/director/ScreenView2.vue:77-80,473-496,3387-3436,4655-4670`
- Test: `web/src/views/director/ScreenView2.test.ts`

- [ ] **Step 1: 写失败回归测试**

在 `ScreenView2.test.ts` 中增加：

```ts
it('counts only visible business phase steps and renders stable numeric hierarchy', () => {
  expect(source).toContain('const visiblePhaseSteps = phase.phaseSteps.filter(ps => ps.name !== phase.name)')
  expect(source).toContain("visiblePhaseSteps.filter(ps => getPhaseStepStatus(ps) === 'done').length")
  expect(source).toContain('const totalPhaseSteps = visiblePhaseSteps.length')
  expect(source).not.toContain('const totalPhaseSteps = phase.phaseSteps.length || 1')
  expect(template).toContain('class="stat-divider"')
  expect(template).toContain('class="stat-total"')

  const styles = descriptor.styles.map(style => style.content).join('\n')
  expect(styles).toMatch(/\.phase-stats\s*\{[^}]*font-variant-numeric:\s*tabular-nums/)
  expect(styles).toContain('.stat-divider')
  expect(styles).toContain('.stat-total')
})
```

- [ ] **Step 2: 运行测试并确认旧口径导致失败**

Run: `cd web && npm test -- src/views/director/ScreenView2.test.ts`

Expected: FAIL，提示缺少 `visiblePhaseSteps`、`stat-divider` 和等宽数字样式。

- [ ] **Step 3: 修正大屏2统计口径**

在 `phaseCards` 内部替换环节统计：

```ts
const visiblePhaseSteps = phase.phaseSteps.filter(ps => ps.name !== phase.name)
const completedPhaseSteps = visiblePhaseSteps.filter(ps => getPhaseStepStatus(ps) === 'done').length
const totalPhaseSteps = visiblePhaseSteps.length
```

叶子步骤相关的 `completedSteps` 和 `totalSteps` 保持不变。

- [ ] **Step 4: 精修大屏2阶段统计模板和样式**

将两个统计单元改为：

```vue
<span class="stat-cell" :aria-label="`${phase.completedPhaseSteps}/${phase.totalPhaseSteps} 环节`">
  <b>{{ phase.completedPhaseSteps }}</b>
  <span class="stat-divider" aria-hidden="true">/</span>
  <span class="stat-total">{{ phase.totalPhaseSteps }}</span>
  <em>环节</em>
</span>
<span class="stat-cell" :aria-label="`${phase.completedSteps}/${phase.totalSteps} 步骤`">
  <b>{{ phase.completedSteps }}</b>
  <span class="stat-divider" aria-hidden="true">/</span>
  <span class="stat-total">{{ phase.totalSteps }}</span>
  <em>步骤</em>
</span>
```

在 `.phase-stats` 增加 `font-variant-numeric: tabular-nums;`，将原 `.phase-stats span` 和响应式同名选择器改为 `.phase-stats > .stat-cell`，并增加：

```css
.stat-divider {
  margin: 0 2px;
  color: rgba(197, 232, 248, 0.48);
  font-size: 0.86em;
}

.stat-total {
  color: rgba(233, 248, 255, 0.82);
  font-size: 0.94em;
  font-weight: 700;
}
```

- [ ] **Step 5: 运行大屏2测试并确认通过**

Run: `cd web && npm test -- src/views/director/ScreenView2.test.ts src/views/director/screenPhaseFlow.test.ts`

Expected: PASS，大屏2测试和共享环节过滤回归测试全部通过。

- [ ] **Step 6: 提交大屏2修正**

```bash
git add web/src/views/director/ScreenView2.vue web/src/views/director/ScreenView2.test.ts
git commit -m "fix: align screen2 phase counts"
```

### Task 2: 独立修正大屏4统计和数字层级

**Files:**
- Modify: `web/src/views/director/ScreenView4.vue:77-80,420-443,3248-3300,3828-3843`
- Test: `web/src/views/director/ScreenView4.test.ts`

- [ ] **Step 1: 写失败回归测试**

在 `ScreenView4.test.ts` 中增加：

```ts
it('counts only visible business phase steps with its own presentation', () => {
  expect(screen4Source).toContain('const visiblePhaseSteps = phase.phaseSteps.filter(ps => ps.name !== phase.name)')
  expect(screen4Source).toContain("visiblePhaseSteps.filter(ps => getPhaseStepStatus(ps) === 'done').length")
  expect(screen4Source).toContain('const totalPhaseSteps = visiblePhaseSteps.length')
  expect(screen4Source).not.toContain('const totalPhaseSteps = phase.phaseSteps.length || 1')
  expect(screen4Source).toContain('class="stat-divider"')
  expect(screen4Source).toContain('class="stat-total"')
  expect(screen4Source).toMatch(/\.phase-stats\s*\{[^}]*font-variant-numeric:\s*tabular-nums/)
  expect(screen4Source).toContain('.stat-divider')
  expect(screen4Source).toContain('.stat-total')
})
```

- [ ] **Step 2: 运行测试并确认旧口径导致失败**

Run: `cd web && npm test -- src/views/director/ScreenView4.test.ts`

Expected: FAIL，提示大屏4仍使用全量 `phase.phaseSteps` 计数且缺少新数字样式。

- [ ] **Step 3: 在大屏4内独立修正统计与样式**

在 `ScreenView4.vue` 的 `phaseCards` 内部增加：

```ts
const visiblePhaseSteps = phase.phaseSteps.filter(ps => ps.name !== phase.name)
const completedPhaseSteps = visiblePhaseSteps.filter(ps => getPhaseStepStatus(ps) === 'done').length
const totalPhaseSteps = visiblePhaseSteps.length
```

将两个统计单元改为：

```vue
<span class="stat-cell" :aria-label="`${phase.completedPhaseSteps}/${phase.totalPhaseSteps} 环节`">
  <b>{{ phase.completedPhaseSteps }}</b>
  <span class="stat-divider" aria-hidden="true">/</span>
  <span class="stat-total">{{ phase.totalPhaseSteps }}</span>
  <em>环节</em>
</span>
<span class="stat-cell" :aria-label="`${phase.completedSteps}/${phase.totalSteps} 步骤`">
  <b>{{ phase.completedSteps }}</b>
  <span class="stat-divider" aria-hidden="true">/</span>
  <span class="stat-total">{{ phase.totalSteps }}</span>
  <em>步骤</em>
</span>
```

在 `.phase-stats` 增加 `font-variant-numeric: tabular-nums;`，将原 `.phase-stats span` 和响应式同名选择器改为 `.phase-stats > .stat-cell`，并增加：

```css
.stat-divider {
  margin: 0 2px;
  color: rgba(197, 232, 248, 0.48);
  font-size: 0.86em;
}

.stat-total {
  color: rgba(233, 248, 255, 0.82);
  font-size: 0.94em;
  font-weight: 700;
}
```

不导入 `ScreenView2.vue` 的任何实现，不修改 `screenPhaseFlow.ts`。

- [ ] **Step 4: 运行大屏4和隔离回归测试**

Run: `cd web && npm test -- src/views/director/ScreenView4.test.ts src/views/director/ScreenView2.test.ts src/views/director/screenPhaseFlow.test.ts`

Expected: PASS，两页统计测试均通过，大屏4仍不导入大屏2。

- [ ] **Step 5: 提交大屏4修正**

```bash
git add web/src/views/director/ScreenView4.vue web/src/views/director/ScreenView4.test.ts
git commit -m "fix: align screen4 phase counts"
```

### Task 3: 完整验证与双页浏览器检查

**Files:**
- Verify: `web/src/views/director/ScreenView2.vue`
- Verify: `web/src/views/director/ScreenView4.vue`

- [ ] **Step 1: 运行全量前端测试**

Run: `cd web && npm test`

Expected: PASS，零失败。

- [ ] **Step 2: 运行类型检查和生产构建**

Run: `cd web && npm run typecheck && npm run build`

Expected: exit 0，无 TypeScript 或 Vue 编译错误。

- [ ] **Step 3: 检查改动边界**

Run: `git diff --check && git status --short`

Expected: `git diff --check` 无输出；只允许计划内的两个页面和两个测试文件发生变化，构建生成文件应还原。

- [ ] **Step 4: 浏览器验证大屏2**

打开 `/director/screen2/90`，确认实施阶段显示 `1/6 环节`和 `6/22 步骤`；数字层级清晰，卡片宽度无抖动、无截断。

- [ ] **Step 5: 浏览器验证大屏4**

打开 `/director/screen4/90`，确认顶部实施阶段和跑道汇总均显示 `1/6 环节`，且步骤仍为 `6/22`。
