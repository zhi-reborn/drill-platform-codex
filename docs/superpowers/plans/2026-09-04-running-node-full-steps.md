# Running Node Full Step List Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Keep the running node's steps in one readable column, show up to ten without clipping, and scroll only the list when more steps exist.

**Architecture:** Keep the existing running-node data selection and focus transform. Add a scroll-state binding at the SFC boundary, cap only lists longer than ten items, and reserve enough viewport space for the transformed ten-row panel and its glow.

**Tech Stack:** Vue 3, TypeScript, Vitest, Vue SFC compiler, Vite

---

### Task 1: Select Visible Steps by Node Status

**Files:**
- Modify: `web/src/views/director/screenPhaseFlow.ts`
- Modify: `web/src/views/director/screenPhaseFlow.test.ts`
- Modify: `web/src/views/director/ScreenView2.vue:127-143,237,529-530`
- Test: `web/src/views/director/screenPhaseFlow.test.ts`
- Test: `web/src/views/director/ScreenView2.test.ts`

- [ ] **Step 1: Write the failing helper tests**

Add `getVisibleNodeSteps` to the import list in `screenPhaseFlow.test.ts`, then add these cases under `describe('selected phase nodes', ...)`:

```ts
it('shows every step for the running node', () => {
  const steps = Array.from({ length: 5 }, (_, index) => ({ id: String(index + 1) }))
  expect(getVisibleNodeSteps({ status: 'running', steps }, 3)).toEqual(steps)
})

it('keeps non-running nodes within the compact step limit', () => {
  const steps = Array.from({ length: 5 }, (_, index) => ({ id: String(index + 1) }))
  expect(getVisibleNodeSteps({ status: 'done', steps }, 3)).toEqual(steps.slice(0, 3))
  expect(getVisibleNodeSteps({ status: 'pending', steps }, 3)).toEqual(steps.slice(0, 3))
})
```

- [ ] **Step 2: Run the focused test and verify RED**

Run:

```bash
cd web && npm test -- src/views/director/screenPhaseFlow.test.ts
```

Expected: FAIL because `getVisibleNodeSteps` is not exported by `screenPhaseFlow.ts`.

- [ ] **Step 3: Add the minimal pure helper**

Add this export to `screenPhaseFlow.ts` after `FlowStepDetail`:

```ts
export function getVisibleNodeSteps<T>(node: { status: string; steps: T[] }, limit: number): T[] {
  return node.status === 'running' ? node.steps : node.steps.slice(0, limit)
}
```

- [ ] **Step 4: Run the focused test and verify GREEN**

Run:

```bash
cd web && npm test -- src/views/director/screenPhaseFlow.test.ts
```

Expected: PASS with both new cases passing.

- [ ] **Step 5: Wire the helper into the screen template**

Add `getVisibleNodeSteps` to the existing import from `./screenPhaseFlow`, then change the node-step rendering block to:

```vue
<li
  v-for="step in getVisibleNodeSteps(node, NODE_STEP_LIMIT)"
  :key="step.id"
  class="node-step"
  :class="'is-' + step.status"
>
  <i class="step-ico" aria-hidden="true" />
  <span class="step-name" :title="step.name">{{ step.name }}</span>
  <i v-if="step.status === 'done'" class="step-check" aria-hidden="true">
    <svg viewBox="0 0 12 12"><path d="M2.4 6.4 L5 9 L9.6 3.4" /></svg>
  </i>
</li>
<li v-if="node.status !== 'running' && node.steps.length > NODE_STEP_LIMIT" class="node-step is-more">
  <i class="step-ico" aria-hidden="true" />
  <span class="step-name">另有 {{ node.steps.length - NODE_STEP_LIMIT }} 个步骤…</span>
</li>
```

- [ ] **Step 6: Add a template regression assertion**

Add this case to `ScreenView2.test.ts`:

```ts
it('expands every task for the running node while compacting other nodes', () => {
  expect(template).toContain('v-for="step in getVisibleNodeSteps(node, NODE_STEP_LIMIT)"')
  expect(template).toContain("node.status !== 'running' && node.steps.length > NODE_STEP_LIMIT")
  expect(source).toContain('getVisibleNodeSteps')
})
```

- [ ] **Step 7: Run screen flow tests**

Run:

```bash
cd web && npm test -- src/views/director/screenPhaseFlow.test.ts src/views/director/ScreenView2.test.ts
```

Expected: PASS with no failed tests.

- [ ] **Step 8: Build the frontend**

Run:

```bash
cd web && npm run build
```

Expected: `vue-tsc` and Vite finish with exit code 0.

- [ ] **Step 9: Verify the live Chrome layout**

Reload `http://localhost:5173/admin/screen/90` at a 1159 × 863 viewport and inspect the running node “应急回滚方案准备”. Verify all four real steps are present, “另有 1 个步骤…” is absent from that node, the list remains within the flow viewport, and it does not overlap the overview or execution-log panels.

- [ ] **Step 10: Commit the implementation**

```bash
git add web/src/views/director/screenPhaseFlow.ts web/src/views/director/screenPhaseFlow.test.ts web/src/views/director/ScreenView2.vue web/src/views/director/ScreenView2.test.ts
git commit -m "fix: expand running screen node steps"
```

### Task 2: Constrain the Single-Column Running List

**Files:**
- Modify: `web/src/views/director/ScreenView2.vue:127,529-530,3796-3805,4035-4052,4181-4189`
- Modify: `web/src/views/director/ScreenView2.test.ts`
- Test: `web/src/views/director/ScreenView2.test.ts`

- [x] **Step 1: Write failing SFC regression assertions**

Assert that the template applies `is-scrollable` only when a running node has more than `RUNNING_NODE_VISIBLE_LIMIT` steps, that the limit is ten, and that the scroll-state CSS uses a single column with an internal vertical scroller and a styled scrollbar. Update the existing viewport-clearance assertion to require `clamp(104px, 15.5vh, 132px)`.

- [x] **Step 2: Run the focused test and verify RED**

Run:

```bash
cd web && npm test -- src/views/director/ScreenView2.test.ts
```

Expected: FAIL because the limit, class binding, scroll-state styles, and expanded clearance do not exist yet.

- [x] **Step 3: Add the template state and limit**

Add `RUNNING_NODE_VISIBLE_LIMIT = 10` beside `NODE_STEP_LIMIT` and bind `is-scrollable` on `.node-steps` only for a running node whose step count exceeds that limit.

- [x] **Step 4: Add the single-column scrolling treatment**

Keep the natural list height through ten entries. For longer running lists, use a ten-row maximum height, `overflow-y: auto`, `overscroll-behavior: contain`, a stable scrollbar gutter, and a narrow amber scrollbar. Increase `.flow-viewport` bottom padding to `clamp(104px, 15.5vh, 132px)` so the focused list border and glow remain visible after the `1.3` transform.

- [x] **Step 5: Run focused and full frontend verification**

Run:

```bash
cd web && npm test -- src/views/director/ScreenView2.test.ts src/views/director/screenPhaseFlow.test.ts
cd web && npm run build
git diff --check
```

Expected: both Vitest files pass, the Vue/TypeScript build exits zero, and the diff check is clean.

- [x] **Step 6: Verify Chrome geometry**

Reload `http://localhost:5173/admin/screen/90` in the user's Chrome tab. Confirm the running node's fourth row, bottom border, and glow are visibly inside `.flow-viewport` at the current short viewport; confirm the list stays single-column and does not overlap the side panels.
