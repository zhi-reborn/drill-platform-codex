import type { StepInstance, StepStatus } from '@/types'

export type MonitorStepTreeNode = StepInstance & {
  children?: MonitorStepTreeNode[]
  _isGroup?: boolean
  _groupType?: 'phase' | 'phaseStep'
}

export type VisibleMonitorStepRow = MonitorStepTreeNode & {
  children?: undefined
}

export type StepTreeMeta = {
  depth: number
  siblingIndex: number
  isParent: boolean
  statusText: { text: string; isParent: boolean }
}

export type MonitorStepPatch = {
  step_id?: number
  stepId?: number
  new_status?: StepStatus
  newStatus?: StepStatus
  start_time?: string
  end_time?: string
  timeout_at?: string | null
  assignee_names?: string
  remark?: string
  issue_desc?: string
}

export type MonitorStepPatchResult = {
  steps: StepInstance[]
  previous: StepInstance
  updated: StepInstance
}

const TERMINAL_STATUSES: StepStatus[] = ['completed', 'skipped', 'issue']
const terminalStatusSet = new Set<StepStatus>(TERMINAL_STATUSES)

function createGroupNode(
  id: string,
  name: string,
  seq: number,
  groupType: 'phase' | 'phaseStep'
): MonitorStepTreeNode {
  return {
    id: id as any,
    name,
    seq,
    status: '' as StepStatus,
    step_type: groupType === 'phase' ? 'phase' : 'phaseStep',
    timeout_minutes: 0,
    default_assignee_role: '',
    estimated_duration_minutes: 0,
    attributes: {},
    start_time: null,
    end_time: null,
    timeout_at: null,
    assignee_names: '',
    remark: '',
    issue_desc: '',
    executor_team: '',
    step_template_id: 0,
    drill_instance_id: 0,
    assignee_ids: '',
    actual_operator: null,
    created_at: '',
    children: [],
    _isGroup: true,
    _groupType: groupType,
  }
}

export function createMonitorStepTreeBuilder() {
  const stepNodeCache = new Map<number, { source: StepInstance; node: MonitorStepTreeNode }>()
  const groupNodeCache = new Map<string, MonitorStepTreeNode>()

  function getStepNode(step: StepInstance): MonitorStepTreeNode {
    const cached = stepNodeCache.get(step.id)
    if (cached?.source === step) {
      cached.node.children = undefined
      return cached.node
    }

    const node: MonitorStepTreeNode = { ...step }
    stepNodeCache.set(step.id, { source: step, node })
    return node
  }

  function getGroupNode(id: string, name: string, seq: number, groupType: 'phase' | 'phaseStep') {
    const cached = groupNodeCache.get(id)
    if (cached) {
      cached.name = name
      cached.seq = seq
      cached.children = []
      return cached
    }
    const node = createGroupNode(id, name, seq, groupType)
    groupNodeCache.set(id, node)
    return node
  }

  function build(steps: StepInstance[]): MonitorStepTreeNode[] {
    const stepMap = new Map<number, MonitorStepTreeNode>()
    const liveStepIds = new Set<number>()

    for (const step of steps) {
      liveStepIds.add(step.id)
      stepMap.set(step.id, getStepNode(step))
    }
    for (const id of stepNodeCache.keys()) {
      if (!liveStepIds.has(id)) stepNodeCache.delete(id)
    }

    const childIds = new Set<number>()
    for (const step of steps) {
      if (step.parent_step_id && stepMap.has(step.parent_step_id)) {
        const parent = stepMap.get(step.parent_step_id)!
        if (!parent.children) parent.children = []
        parent.children.push(stepMap.get(step.id)!)
        childIds.add(step.id)
      }
    }

    const rootSteps = steps.filter(s => !childIds.has(s.id)).map(s => stepMap.get(s.id)!)
    if (childIds.size > 0) {
      return rootSteps
    }

    const phaseMap = new Map<string, MonitorStepTreeNode>()
    const roots: MonitorStepTreeNode[] = []
    let virtualSeq = -1

    for (const step of rootSteps) {
      const phase = step.phase || '未分类'
      const phaseId = `__phase__${phase}`
      let phaseNode = phaseMap.get(phase)
      if (!phaseNode) {
        phaseNode = getGroupNode(phaseId, phase, virtualSeq--, 'phase')
        phaseMap.set(phase, phaseNode)
        roots.push(phaseNode)
      }

      const phaseStep = step.phase_step || '默认'
      const phaseStepId = `__phaseStep__${phase}__${phaseStep}`
      let phaseStepNode = phaseNode.children!.find(
        node => node._isGroup && node._groupType === 'phaseStep' && node.name === phaseStep
      )
      if (!phaseStepNode) {
        phaseStepNode = getGroupNode(phaseStepId, phaseStep, virtualSeq--, 'phaseStep')
        phaseNode.children!.push(phaseStepNode)
      }

      phaseStepNode.children!.push(step)
    }

    return roots
  }

  return { build }
}

export function patchMonitorStepList(
  steps: StepInstance[],
  payload: MonitorStepPatch
): MonitorStepPatchResult | null {
  const stepId = payload.step_id || payload.stepId
  if (!stepId) return null

  const idx = steps.findIndex(step => step.id === stepId)
  if (idx === -1) return null

  const newStatus = payload.new_status || payload.newStatus
  if (!newStatus) return null

  const previous = steps[idx]
  const updated = { ...previous }
  updated.status = newStatus === 'timeout' ? 'running' : newStatus
  if (payload.start_time) updated.start_time = payload.start_time
  if (payload.end_time) updated.end_time = payload.end_time
  if ('timeout_at' in payload) updated.timeout_at = payload.timeout_at ?? null
  if (payload.assignee_names) updated.assignee_names = payload.assignee_names
  if (payload.remark) updated.remark = payload.remark
  if (payload.issue_desc) updated.issue_desc = payload.issue_desc

  const nextSteps = [...steps]
  nextSteps[idx] = updated

  return { steps: nextSteps, previous, updated }
}

export function getDefaultExpandedStepKeys(steps: StepInstance[]): Array<number | string> {
  const keys = new Set<number | string>()
  const stepById = new Map<number, StepInstance>()
  const parentIds = new Set<number>()

  for (const step of steps) {
    stepById.set(step.id, step)
    if (step.parent_step_id) {
      parentIds.add(step.parent_step_id)
    }
    if (!step.parent_step_id) {
      keys.add(step.id)
    }
  }

  for (const step of steps) {
    if (step.status !== 'running' && step.status !== 'issue') continue
    let current = step
    while (current.parent_step_id) {
      const parent = stepById.get(current.parent_step_id)
      if (!parent) break
      keys.add(parent.id)
      current = parent
    }
    if (parentIds.has(step.id)) {
      keys.add(step.id)
    }
  }

  return steps
    .map(step => step.id)
    .filter(id => keys.has(id))
}

export function flattenVisibleStepTree(
  nodes: MonitorStepTreeNode[],
  expandedKeys: Array<number | string>
): VisibleMonitorStepRow[] {
  const expandedKeySet = new Set(expandedKeys.map(String))
  const rows: VisibleMonitorStepRow[] = []

  function visit(items: MonitorStepTreeNode[]) {
    for (const node of items) {
      const { children: _children, ...row } = node
      rows.push(row as VisibleMonitorStepRow)
      if (_children?.length && expandedKeySet.has(String(node.id))) {
        visit(_children)
      }
    }
  }

  visit(nodes)
  return rows
}

export function buildStepTreeMetaMap(nodes: MonitorStepTreeNode[]): Map<number | string, StepTreeMeta> {
  const metaMap = new Map<number | string, StepTreeMeta>()

  function visit(items: MonitorStepTreeNode[], depth: number) {
    let total = 0
    let completed = 0
    let running = 0

    items.forEach((node, index) => {
      const hasChildren = Boolean(node.children?.length)
      const childSummary = hasChildren ? visit(node.children!, depth + 1) : null
      const nodeTotal = childSummary?.total ?? 1
      const nodeCompleted = childSummary?.completed ?? (terminalStatusSet.has(node.status) ? 1 : 0)
      const nodeRunning = childSummary?.running ?? (node.status === 'running' ? 1 : 0)
      const isParent = Boolean(node._isGroup || hasChildren)
      const statusText = isParent
        ? {
            text: nodeRunning > 0
              ? `${nodeCompleted}/${nodeTotal} 已完成 · ${nodeRunning} 进行中`
              : `${nodeCompleted}/${nodeTotal} ${node._isGroup ? '已完成' : '子任务已完成'}`,
            isParent: true,
          }
        : { text: node.status, isParent: false }

      metaMap.set(node.id, {
        depth,
        siblingIndex: index + 1,
        isParent,
        statusText,
      })

      total += nodeTotal
      completed += nodeCompleted
      running += nodeRunning
    })

    return { total, completed, running }
  }

  visit(nodes, 0)
  return metaMap
}
