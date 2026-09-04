import { computed, ref, watch, type Ref } from 'vue'

interface PhaseSummary {
  name: string
  status: string
}

export function useScreenPhaseSelection(phases: Ref<PhaseSummary[]>, drillId: Ref<number>) {
  const selectedName = ref<string | null>(null)
  let previousDrillId = drillId.value
  let lastRunningName: string | null = null

  watch([() => phases.value.map(p => ({ name: p.name, status: p.status })), drillId], ([items, id]) => {
    if (id !== previousDrillId) {
      previousDrillId = id
      selectedName.value = null
      lastRunningName = null
    }
    const running = items.find(p => p.status === 'running')
    if (running && running.name !== lastRunningName) {
      selectedName.value = running.name
    } else if (!items.some(p => p.name === selectedName.value)) {
      selectedName.value = (running ?? items.find(p => p.status !== 'done') ?? items[items.length - 1])?.name ?? null
    }
    // A pause/resume must not count as advancing to a new phase.
    if (running) lastRunningName = running.name
  }, { immediate: true })

  const selectedPhaseIdx = computed({
    get: () => phases.value.findIndex(p => p.name === selectedName.value),
    set: (index: number) => {
      if (phases.value[index]) selectedName.value = phases.value[index].name
    },
  })

  return { selectedPhaseIdx }
}

export interface FlowStepDetail {
  id: string
  name: string
  status: string
}

export function getPhaseFlowNodes<T extends { name: string }>(
  phase: { name: string; phaseSteps: T[] } | null,
  statusOf: (link: T) => string,
  stepsOf?: (link: T) => FlowStepDetail[],
) {
  if (!phase) return []
  return phase.phaseSteps.filter(link => link.name !== phase.name).map((link, index) => ({
    id: `${phase.name}:${index}:${link.name}`,
    name: link.name,
    index: index + 1,
    status: statusOf(link),
    steps: stepsOf ? stepsOf(link) : [],
  }))
}

export function getFlowFocusIndex(nodes: { status: string }[]): number {
  const running = nodes.findIndex(node => node.status === 'running')
  if (running >= 0) return running
  const pending = nodes.findIndex(node => node.status === 'pending')
  return pending >= 0 ? pending : nodes.length - 1
}

export function getPhaseStripScrollLeft(scrollLeft: number, viewportWidth: number, cardLeft: number, cardWidth: number): number {
  if (cardLeft < scrollLeft) return cardLeft
  if (cardLeft + cardWidth > scrollLeft + viewportWidth) return cardLeft + cardWidth - viewportWidth
  return scrollLeft
}

export function getPhaseChamberPath(
  width: number,
  height: number,
  boardTop: number,
  tab?: { left: number; right: number; top: number },
): string {
  const x = width - 1
  const y = height - 1
  const top = boardTop + 1
  let raisedTab = ''
  if (tab) {
    const left = Math.max(20, tab.left)
    const right = Math.min(width - 20, tab.right)
    if (right - left >= 24) {
      // 绕过选中页签，底部不闭合，与内容区域共用一条外轮廓。
      raisedTab = `H ${left - 6} Q ${left} ${top} ${left} ${top - 6} V ${tab.top + 10} Q ${left} ${tab.top} ${left + 10} ${tab.top} H ${right - 10} Q ${right} ${tab.top} ${right} ${tab.top + 10} V ${top - 6} Q ${right} ${top} ${right + 6} ${top}`
    }
  }
  return `M 15 ${top} ${raisedTab} H ${x - 14} Q ${x} ${top} ${x} ${top + 14} V ${y - 14} Q ${x} ${y} ${x - 14} ${y} H 15 Q 1 ${y} 1 ${y - 14} V ${top + 14} Q 1 ${top} 15 ${top} Z`
}
