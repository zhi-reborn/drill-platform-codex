# Screen Flow Rail Cap Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the card-like virtual start/end nodes with restrained rail caps and make the selected flow node fade progressively toward both viewport edges without changing centering behavior.

**Architecture:** Keep the existing virtual wrapper elements so layout width and the `real index + 1` DOM mapping remain stable. Move distance-to-presentation math into pure helpers in `screenPhaseFlow.ts`, consume those helpers from `ScreenView2.vue`, and verify the visual contract with Vitest plus browser checks for first, middle, and last focus positions.

**Tech Stack:** Vue 3 SFC, TypeScript, CSS, Vitest, Vue compiler-sfc, Vite

---

## File Structure

- Modify `web/src/views/director/screenPhaseFlow.ts`: expose pure focus-presentation and virtual-index helpers.
- Modify `web/src/views/director/screenPhaseFlow.test.ts`: test distance tiers and real-to-DOM index mapping.
- Modify `web/src/views/director/ScreenView2.vue`: render rail caps, consume focus helpers, and replace virtual badge styling.
- Modify `web/src/views/director/ScreenView2.test.ts`: lock the rail-cap template and reduced-motion contract.

The existing files remain in place; no component extraction or unrelated refactor is required.

### Task 1: Make focus tiers and virtual index mapping testable

**Files:**
- Modify: `web/src/views/director/screenPhaseFlow.test.ts:1-170`
- Modify: `web/src/views/director/screenPhaseFlow.ts:60-65`
- Modify: `web/src/views/director/ScreenView2.vue:241,556-640`

- [ ] **Step 1: Write failing tests for the visual distance tiers and virtual wrapper offset**

Extend the import in `screenPhaseFlow.test.ts`:

```ts
import {
  getFlowFocusIndex,
  getFlowFocusPresentation,
  getFlowTargetItemIndex,
  getPhaseChamberPath,
  getPhaseFlowNodes,
  getPhaseStripScrollLeft,
  useScreenPhaseSelection,
} from './screenPhaseFlow'
```

Add these tests inside `describe('selected phase nodes', ...)`:

```ts
it('steps focus presentation down from the selected node toward both edges', () => {
  expect(getFlowFocusPresentation(2, 2)).toEqual({ scale: 1.3, opacity: 1, zIndex: 30 })
  expect(getFlowFocusPresentation(1, 2)).toEqual({ scale: 1, opacity: 0.76, zIndex: 29 })
  expect(getFlowFocusPresentation(0, 2)).toEqual({ scale: 0.84, opacity: 0.52, zIndex: 28 })
  expect(getFlowFocusPresentation(-1, 2)).toEqual({ scale: 0.68, opacity: 0.32, zIndex: 27 })
  expect(getFlowFocusPresentation(7, 2)).toEqual({ scale: 0.58, opacity: 0.24, zIndex: 25 })
})

it('maps real focus indices past the leading virtual wrapper', () => {
  expect(getFlowTargetItemIndex(0, 7)).toBe(1)
  expect(getFlowTargetItemIndex(4, 7)).toBe(5)
  expect(getFlowTargetItemIndex(-1, 7)).toBe(-1)
  expect(getFlowTargetItemIndex(6, 7)).toBe(-1)
})
```

- [ ] **Step 2: Run the focused test and verify it fails**

Run:

```bash
cd web && npm test -- src/views/director/screenPhaseFlow.test.ts
```

Expected: FAIL because `getFlowFocusPresentation` and `getFlowTargetItemIndex` are not exported.

- [ ] **Step 3: Implement the pure focus helpers**

Add after `getFlowFocusIndex` in `screenPhaseFlow.ts`:

```ts
export interface FlowFocusPresentation {
  scale: number
  opacity: number
  zIndex: number
}

export function getFlowFocusPresentation(index: number, focusedIndex: number): FlowFocusPresentation {
  const distance = focusedIndex < 0 ? 0 : Math.abs(index - focusedIndex)
  const scales = [1.3, 1, 0.84, 0.68, 0.58]
  const opacities = [1, 0.76, 0.52, 0.32, 0.24]
  const tier = Math.min(distance, scales.length - 1)
  return {
    scale: scales[tier],
    opacity: opacities[tier],
    zIndex: 30 - distance,
  }
}

export function getFlowTargetItemIndex(focusedIndex: number, itemCount: number): number {
  if (focusedIndex < 0) return -1
  const targetIndex = focusedIndex + 1
  return targetIndex < itemCount ? targetIndex : -1
}
```

- [ ] **Step 4: Use the helpers in `ScreenView2.vue`**

Add both helper names to the existing import from `./screenPhaseFlow`. Replace local scale/opacity calculation with:

```ts
function focusPresentation(index: number) {
  return getFlowFocusPresentation(index, focusedNodeIndex.value)
}

function focusStyle(index: number): CSSProperties {
  const presentation = focusPresentation(index)
  return {
    transform: `scale(${presentation.scale.toFixed(3)})`,
    opacity: presentation.opacity.toFixed(3),
    zIndex: String(presentation.zIndex),
  }
}
```

Update arrow calculations to use `focusPresentation(index).scale` and `.opacity`. In `recomputeFocusShift`, replace the inline `r + 1` bounds check with:

```ts
const targetItemIndex = getFlowTargetItemIndex(r, items.length)
if (targetItemIndex < 0) {
  focusShift.value = 0
  return
}
const target = items[targetItemIndex]
```

- [ ] **Step 5: Run the helper tests and typecheck**

Run:

```bash
cd web && npm test -- src/views/director/screenPhaseFlow.test.ts && npm run typecheck
```

Expected: the focused Vitest file passes and `vue-tsc` exits with code 0.

### Task 2: Replace virtual badges with rail caps

**Files:**
- Modify: `web/src/views/director/ScreenView2.test.ts:1-55`
- Modify: `web/src/views/director/ScreenView2.vue:105-168,603-618,4291-4410,4467-4490`

- [ ] **Step 1: Write a failing component contract test**

Add to `ScreenView2.test.ts`:

```ts
it('renders restrained rail caps for virtual endpoints', () => {
  expect(template).toContain('class="rail-cap"')
  expect(template).toContain('<span class="virtual-name">起点</span>')
  expect(template).toContain('<span class="virtual-name">终点</span>')
  expect(template).toContain("index === flowNodes.length - 1 ? virtualArrowStyle('end') : arrowStyle(index)")
  expect(template).not.toContain('virtual-badge')
  expect(template).not.toContain('virtual-glyph')

  const styles = descriptor.styles.map(style => style.content).join('\n')
  expect(styles).toContain('.rail-cap')
  expect(styles).toContain('.rail-cap-core')
  expect(styles).toMatch(/prefers-reduced-motion[\s\S]*?\.rail-cap-core/)
})
```

- [ ] **Step 2: Run the component contract test and verify it fails**

Run:

```bash
cd web && npm test -- src/views/director/ScreenView2.test.ts
```

Expected: FAIL because the template still contains `virtual-badge` and `virtual-glyph`.

- [ ] **Step 3: Replace the virtual-node markup**

Use this content for both virtual wrappers, retaining their existing positions and `focusStyle` indices:

```vue
<div class="flow-node-wrap is-virtual" :style="focusStyle(-1)">
  <div class="flow-node is-virtual-start">
    <span class="rail-cap" aria-hidden="true"><i class="rail-cap-core" /></span>
    <span class="virtual-name">起点</span>
  </div>
</div>
```

```vue
<div class="flow-node-wrap is-virtual" :style="focusStyle(flowNodes.length)">
  <div class="flow-node is-virtual-end">
    <span class="rail-cap" aria-hidden="true"><i class="rail-cap-core" /></span>
    <span class="virtual-name">终点</span>
  </div>
</div>
```

For the arrow emitted after each real node, use the endpoint-aware style:

```vue
:style="index === flowNodes.length - 1 ? virtualArrowStyle('end') : arrowStyle(index)"
```

- [ ] **Step 4: Align virtual-arrow geometry with the narrow cap**

Rename the ratio to `VIRTUAL_CAP_RATIO` and set it to `0.18`. In `virtualArrowStyle`, use `focusPresentation(...).scale` and `.opacity` to calculate the cap inset, real-card extension, and adjacent minimum opacity. This keeps the full wrapper width for centering while joining the visible connector to the narrow cap.

- [ ] **Step 5: Replace the virtual badge CSS with rail-cap CSS**

Replace the `.virtual-badge` and `.virtual-glyph*` block with:

```css
.flow-node.is-virtual-start,
.flow-node.is-virtual-end {
  --cap-color: 81, 230, 255;
  height: var(--node-tag-h);
  display: grid;
  place-items: center;
}

.flow-node.is-virtual-end { --cap-color: 75, 231, 173; }

.rail-cap {
  position: relative;
  width: 18%;
  height: 68%;
  display: grid;
  place-items: center;
  color: rgb(var(--cap-color));
}

.rail-cap-core {
  width: 3px;
  height: 78%;
  border-radius: 3px;
  background: linear-gradient(180deg, transparent, rgba(var(--cap-color), 0.95), transparent);
  box-shadow: -8px 0 rgba(var(--cap-color), 0.13), 8px 0 rgba(var(--cap-color), 0.13), 0 0 14px rgba(var(--cap-color), 0.48);
  animation: rail-cap-breathe 3.2s ease-in-out infinite;
}

.rail-cap::after {
  content: "";
  position: absolute;
  width: 20px;
  height: 20px;
  border-top: 1px solid rgba(var(--cap-color), 0.48);
  border-right: 1px solid rgba(var(--cap-color), 0.48);
  transform: rotate(45deg);
}

.flow-node.is-virtual-end .rail-cap::after { transform: rotate(-135deg); }

.virtual-name {
  position: absolute;
  top: calc(100% + 7px);
  left: 50%;
  transform: translateX(-50%);
  margin-left: 0.24em;
  color: rgba(var(--cap-color), 0.58);
  font-size: clamp(9px, 0.72vw, 11px);
  font-weight: 600;
  letter-spacing: 0.28em;
  white-space: nowrap;
  text-shadow: 0 0 9px rgba(var(--cap-color), 0.24);
}

@keyframes rail-cap-breathe {
  0%, 100% { opacity: 0.52; filter: brightness(0.82); }
  50% { opacity: 0.92; filter: brightness(1.18); }
}
```

Keep the existing subdued `.flow-arrow.is-virtual` palette. Replace the reduced-motion references to `.virtual-badge` with `.rail-cap-core` so the cap animation is disabled when requested.

- [ ] **Step 6: Run focused tests and inspect the patch**

Run:

```bash
cd web && npm test -- src/views/director/ScreenView2.test.ts src/views/director/screenPhaseFlow.test.ts
git diff --check
git diff -- web/src/views/director/ScreenView2.vue web/src/views/director/ScreenView2.test.ts web/src/views/director/screenPhaseFlow.ts web/src/views/director/screenPhaseFlow.test.ts
```

Expected: both Vitest files pass, `git diff --check` reports no errors, and the diff contains no changes outside the four named files.

### Task 3: Verify the completed flow strip

**Files:**
- Verify only: `web/src/views/director/ScreenView2.vue`

- [ ] **Step 1: Run the full frontend verification suite**

Run:

```bash
cd web && npm test && npm run typecheck && npm run build
```

Expected: all Vitest suites pass, `vue-tsc` exits with code 0, and Vite completes a production build.

- [ ] **Step 2: Verify the live middle-focus state**

Open `http://localhost:5173/admin/screen/90`. Confirm that the running node is centered, gold, and dominant; left and right nodes step down symmetrically; the rail caps do not resemble cards; and no node steps overlap the arrows.

- [ ] **Step 3: Verify first- and last-focus states using phase previews**

Select a pending phase whose focus falls on its first real node and confirm that the first node is centered with the start cap visible to its left. Select a completed phase whose focus falls on its last real node and confirm that the last node is centered with the end cap visible to its right. Return to the running phase after verification.

- [ ] **Step 4: Verify reduced motion and responsive clipping**

Emulate `prefers-reduced-motion: reduce` and confirm the rail cap and arrow particles stop animating. Check the current desktop viewport and one narrower desktop width; the viewport mask must fade edges without clipping the centered active card.

- [ ] **Step 5: Final working-tree review**

Run:

```bash
git status --short
git diff --check
```

Expected: only the already-existing `ScreenView2` work plus the four explicitly modified implementation/test files remain uncommitted, with no whitespace errors. Do not commit these shared dirty files unless the user explicitly requests a code commit.
