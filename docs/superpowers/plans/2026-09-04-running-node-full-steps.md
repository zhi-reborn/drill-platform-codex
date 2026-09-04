# Running Node Full Step List Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Show every step for the currently running flow node while keeping completed and pending nodes limited to three visible steps plus a remainder summary.

**Architecture:** Add one pure display-selection helper beside the existing screen phase-flow helpers, cover it with focused Vitest cases, then use it from `ScreenView2.vue`. The existing flow-node data, status calculation, layout, and styling remain unchanged.

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
