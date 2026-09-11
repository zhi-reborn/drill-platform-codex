# Screen4 Low-GPU Motion Optimization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Keep Screen4's key energy-flow animations visible on ordinary machines while reducing continuous filter work, broad repaints, and hidden-tab animation cost.

**Architecture:** Keep the existing Vue component and SVG runway structure. Add one document-visibility state at the component boundary, then use CSS containment and static-glow/moving-core separation so decorative motion remains visually rich without repeatedly animating filtered layers.

**Tech Stack:** Vue 3 Composition API, TypeScript, scoped SCSS, SVG, Vitest, Vue SFC compiler

---

### Task 1: Pause decorative motion when the page is hidden

**Files:**
- Modify: `web/src/views/director/Screen4Runway.vue:1-2,277-279,461-550,553-565,1567-1582`
- Test: `web/src/views/director/screen4RunwayComponent.test.ts`

- [ ] **Step 1: Write the failing component-source test**

Add this test to `screen4RunwayComponent.test.ts`:

```ts
it('pauses decorative motion while the document is hidden', () => {
  expect(template).toContain("'is-motion-paused': motionPaused")
  expect(source).toContain('document.addEventListener(\'visibilitychange\', syncMotionVisibility)')
  expect(source).toContain('document.removeEventListener(\'visibilitychange\', syncMotionVisibility)')
  expect(source).toContain('if (!root || !dial || motionPaused.value || prefersReducedMotion()) return')
  expect(styles).toMatch(/\.screen4-runway\.is-motion-paused[\s\S]*\*::after[\s\S]*animation-play-state: paused !important;/)
})
```

- [ ] **Step 2: Run the focused test and verify RED**

Run: `cd web && npm test -- screen4RunwayComponent.test.ts`

Expected: FAIL because the visibility listener, `motionPaused`, root class, and pause style do not exist.

- [ ] **Step 3: Add minimal visibility lifecycle state**

Update the Vue import and root class:

```vue
<div
  ref="rootRef"
  class="screen4-runway"
  :class="{ 'is-complete': phaseComplete, 'is-motion-paused': motionPaused }"
>
```

```ts
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'

const motionPaused = ref(typeof document !== 'undefined' && document.hidden)

function syncMotionVisibility() {
  motionPaused.value = document.hidden
}

onMounted(() => {
  syncMotionVisibility()
  document.addEventListener('visibilitychange', syncMotionVisibility)
})
```

Reject hidden-tab flyer creation and clean up the listener:

```ts
if (!root || !dial || motionPaused.value || prefersReducedMotion()) return
```

```ts
onUnmounted(() => {
  document.removeEventListener('visibilitychange', syncMotionVisibility)
  if (dialAbsorbTimer) clearTimeout(dialAbsorbTimer)
})
```

Add the scoped pause rule immediately before the reduced-motion media query:

```scss
.screen4-runway.is-motion-paused *,
.screen4-runway.is-motion-paused *::before,
.screen4-runway.is-motion-paused *::after {
  animation-play-state: paused !important;
}
```

- [ ] **Step 4: Run the focused test and verify GREEN**

Run: `cd web && npm test -- screen4RunwayComponent.test.ts`

Expected: PASS with zero failures.

### Task 2: Separate static glow from moving energy layers

**Files:**
- Modify: `web/src/views/director/Screen4Runway.vue:553-565,929-955,1010-1115,1170-1252,1295-1310,1420-1530,1567-1582`
- Test: `web/src/views/director/screen4RunwayComponent.test.ts`

- [ ] **Step 1: Write failing performance-contract tests**

Add this test:

```ts
it('keeps motion visible with bounded low-cost paint layers', () => {
  expect(styles).toMatch(/\.screen4-runway\s*\{[^}]*contain: paint style;/s)
  expect(styles).toMatch(/\.runway-svg\s*\{[^}]*contain: paint;/s)
  expect(styles).toMatch(/\.runway-complete-flow\s*\{[^}]*filter: none;/s)
  expect(styles).toMatch(/\.runway-active-path\s*\{[^}]*filter: none;/s)
  expect(styles).toMatch(/\.progress-head-core\s*\{[^}]*filter: none;/s)
  expect(styles).toMatch(/\.runway-baton\s*\{[^}]*filter: none;/s)
  expect(styles).toMatch(/\.runway-progress-head\s*\{[^}]*will-change: transform, opacity;/s)
  expect(styles).toMatch(/\.milestone-scan\s*\{[^}]*will-change: transform;/s)
  expect(styles).not.toMatch(/@media \(max-width: 1280px\)[\s\S]*animation: none !important;/)
})
```

- [ ] **Step 2: Run the focused test and verify RED**

Run: `cd web && npm test -- screen4RunwayComponent.test.ts`

Expected: FAIL on missing containment/compositor hints, filtered moving layers, and the existing narrow-screen blanket animation shutdown.

- [ ] **Step 3: Bound repaints and remove filters from moving layers**

Add paint boundaries:

```scss
.screen4-runway {
  contain: paint style;
}

.runway-svg,
.runway-deck {
  contain: paint;
}
```

Keep the existing filtered `runway-rail`, `runway-complete-path`, and active-progress underlays as static atmosphere. Update only continuously moving layers:

```scss
.runway-complete-flow,
.runway-active-path,
.progress-head-core,
.runway-baton {
  filter: none;
}

.runway-progress-head {
  will-change: transform, opacity;
}

.milestone-scan {
  transform-box: fill-box;
  transform-origin: center;
  will-change: transform;
}
```

Replace the `max-width: 1280px` animation shutdown list with slower but still-live motion:

```scss
@media (max-width: 1280px) {
  .runway-dashes { animation-duration: 4.8s; }
  .runway-complete-flow { animation-duration: 2.4s; }
  .runway-active-path { animation-duration: 2.6s; }
  .runway-node.is-completed .node-orbit-outer { animation-duration: 18s; }
  .ticker-standby { animation-duration: 4.2s; }
  .ticker-standby::after { animation-duration: 5.2s; }
}
```

Retain the existing spacing adjustments in that media query.

- [ ] **Step 4: Run Screen4 tests and verify GREEN**

Run: `cd web && npm test -- screen4Runway.test.ts screen4RunwayComponent.test.ts`

Expected: both test files pass with zero failures.

### Task 3: Verify production behavior and visual quality

**Files:**
- Verify: `web/src/views/director/Screen4Runway.vue`
- Verify: `web/src/views/director/screen4RunwayComponent.test.ts`

- [ ] **Step 1: Run the full frontend production build**

Run: `cd web && npm run build`

Expected: `vue-tsc -b && vite build` exits 0.

- [ ] **Step 2: Check the complete diff**

Run: `git diff --check && git diff -- web/src/views/director/Screen4Runway.vue web/src/views/director/screen4RunwayComponent.test.ts`

Expected: no whitespace errors; every new line maps to visibility pausing, paint containment, moving-layer cost, or tests. Existing task-name truncation changes remain intact.

- [ ] **Step 3: Inspect the live page**

Open or refresh `http://localhost:5173/director/screen4/90` and verify:

- the page layout and status colors are unchanged;
- green completed-path beads, orange active-path motion, the current-node pulse, and milestone scan remain visible;
- narrow viewport motion remains active at a slower cadence;
- switching away from the tab pauses decorative CSS animations and switching back resumes them;
- no new browser console errors appear.

- [ ] **Step 4: Commit implementation files only**

```bash
git add web/src/views/director/Screen4Runway.vue web/src/views/director/screen4RunwayComponent.test.ts
git commit -m "perf: optimize screen4 motion for low gpu devices"
```
