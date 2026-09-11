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

type Screen4RunwayCompletionNode = Pick<Screen4RunwayNode, 'status' | 'completed' | 'total'>

const VIEWBOX_WIDTH = 1040
const LANE_CAPACITY = 5
const LANE_LEFT = 90
const LANE_RIGHT = 950
const LANE_PADDING = 70
const SUMMARY_CLEARANCE = 120
const TURN_LEFT = 42
const TURN_RIGHT = 998
const TURN_RADIUS = 28
const FIRST_LANE_Y = 240
const LANE_GAP = 160
const BOTTOM_PADDING = 120

function distributeAcrossLanes(itemCount: number, rowCount: number): number[] {
  const baseCount = Math.floor(itemCount / rowCount)
  const remainder = itemCount % rowCount

  return Array.from({ length: rowCount }, (_, row) => baseCount + (row < remainder ? 1 : 0))
}

export function getScreen4RunwayPath(
  points: Screen4RunwayPoint[],
  from: number,
  to: number,
): string {
  if (!points.length || from < 0 || to < from || from >= points.length) return ''

  const pathPoints = points.slice(from, Math.min(to, points.length - 1) + 1)
  if (!pathPoints.length) return ''

  let path = `M ${pathPoints[0].x} ${pathPoints[0].y}`

  for (let index = 1; index < pathPoints.length; index += 1) {
    const previous = pathPoints[index - 1]
    const current = pathPoints[index]

    if (previous.row === current.row) {
      path += ` L ${current.x} ${current.y}`
      continue
    }

    const turnX = previous.direction === 'right' ? TURN_RIGHT : TURN_LEFT
    const turnDirection = previous.direction === 'right' ? 1 : -1
    const radius = Math.min(TURN_RADIUS, Math.abs(current.y - previous.y) / 2)
    const curveX = turnX - turnDirection * radius
    const verticalDirection = Math.sign(current.y - previous.y) || 1
    const curveEntryY = previous.y + verticalDirection * radius
    const curveExitY = current.y - verticalDirection * radius

    path += ` L ${curveX} ${previous.y}`
    path += ` Q ${turnX} ${previous.y} ${turnX} ${curveEntryY}`
    path += ` L ${turnX} ${curveExitY}`
    path += ` Q ${turnX} ${current.y} ${curveX} ${current.y}`
    path += ` L ${current.x} ${current.y}`
  }

  return path
}

export function buildScreen4RunwayLayout(nodeCount: number): Screen4RunwayLayout {
  const safeNodeCount = Math.max(0, Math.floor(nodeCount))
  const itemCount = safeNodeCount + 1
  const rowCount = Math.max(1, Math.ceil(itemCount / LANE_CAPACITY))
  const laneCounts = distributeAcrossLanes(itemCount, rowCount)
  const laneY = Array.from({ length: rowCount }, (_, row) => FIRST_LANE_Y + row * LANE_GAP)
  const points: Screen4RunwayPoint[] = []

  laneCounts.forEach((count, row) => {
    const direction: Screen4RunwayDirection = row % 2 === 0 ? 'right' : 'left'
    const usableLeft = LANE_LEFT + LANE_PADDING
    const reservesSummarySpace = row === rowCount - 1 && direction === 'right' && rowCount > 1
    const usableRight = LANE_RIGHT - LANE_PADDING - (reservesSummarySpace ? SUMMARY_CLEARANCE : 0)
    const span = usableRight - usableLeft

    for (let laneIndex = 0; laneIndex < count; laneIndex += 1) {
      const ratio = count === 1 ? 0.5 : laneIndex / (count - 1)
      const forwardX = usableLeft + span * ratio
      points.push({
        index: points.length,
        row,
        x: Math.round(direction === 'right' ? forwardX : VIEWBOX_WIDTH - forwardX),
        y: laneY[row],
        direction,
      })
    }
  })

  return {
    rowCount,
    viewBoxHeight: laneY[laneY.length - 1] + BOTTOM_PADDING,
    laneY,
    points,
    trackPath: getScreen4RunwayPath(points, 0, points.length - 1),
  }
}

export function isScreen4RunwayNodeCompleted(node: Screen4RunwayCompletionNode): boolean {
  const status = String(node.status).toLowerCase()
  if (status === 'issue' || status === 'timeout') return false
  return status === 'done'
    || status === 'completed'
    || status === 'skipped'
    || (node.total > 0 && node.completed >= node.total)
}

export function getScreen4RunwayCompletedPathEnd(nodes: Screen4RunwayCompletionNode[]): number | null {
  let completedEndIndex = -1

  for (const [index, node] of nodes.entries()) {
    if (!isScreen4RunwayNodeCompleted(node)) break
    completedEndIndex = index
  }

  return completedEndIndex >= 0
    ? Math.min(completedEndIndex + 1, nodes.length)
    : null
}

export interface Screen4RunwayHopCursor {
  x: number
  y: number
}

export interface Screen4RunwayHopProgress {
  strokeLength: number
  cursor: Screen4RunwayHopCursor
  angle: number
}

interface Screen4RunwaySegment {
  x1: number
  y1: number
  x2: number
  y2: number
  length: number
}

// 弧线按二次贝塞尔采样为短折线，保证进度端点与渲染路径严格重合。
const ARC_SAMPLE_STEPS = 10

function pushScreen4RunwayLine(segments: Screen4RunwaySegment[], x1: number, y1: number, x2: number, y2: number) {
  const length = Math.hypot(x2 - x1, y2 - y1)
  if (length > 0) segments.push({ x1, y1, x2, y2, length })
}

function pushScreen4RunwayQuadratic(
  segments: Screen4RunwaySegment[],
  x0: number,
  y0: number,
  cx: number,
  cy: number,
  x1: number,
  y1: number,
) {
  let previousX = x0
  let previousY = y0
  for (let step = 1; step <= ARC_SAMPLE_STEPS; step += 1) {
    const t = step / ARC_SAMPLE_STEPS
    const rest = 1 - t
    const x = rest * rest * x0 + 2 * t * rest * cx + t * t * x1
    const y = rest * rest * y0 + 2 * t * rest * cy + t * t * y1
    pushScreen4RunwayLine(segments, previousX, previousY, x, y)
    previousX = x
    previousY = y
  }
}

function buildScreen4RunwayHopSegments(
  previous: Screen4RunwayPoint,
  current: Screen4RunwayPoint,
): Screen4RunwaySegment[] {
  const segments: Screen4RunwaySegment[] = []

  if (previous.row === current.row) {
    pushScreen4RunwayLine(segments, previous.x, previous.y, current.x, current.y)
    return segments
  }

  // 与 getScreen4RunwayPath 一致的折返几何。
  const turnX = previous.direction === 'right' ? TURN_RIGHT : TURN_LEFT
  const turnDirection = previous.direction === 'right' ? 1 : -1
  const radius = Math.min(TURN_RADIUS, Math.abs(current.y - previous.y) / 2)
  const curveX = turnX - turnDirection * radius
  const verticalDirection = Math.sign(current.y - previous.y) || 1
  const curveEntryY = previous.y + verticalDirection * radius
  const curveExitY = current.y - verticalDirection * radius

  pushScreen4RunwayLine(segments, previous.x, previous.y, curveX, previous.y)
  pushScreen4RunwayQuadratic(segments, curveX, previous.y, turnX, previous.y, turnX, curveEntryY)
  pushScreen4RunwayLine(segments, turnX, curveEntryY, turnX, curveExitY)
  pushScreen4RunwayQuadratic(segments, turnX, curveExitY, turnX, current.y, curveX, current.y)
  pushScreen4RunwayLine(segments, curveX, current.y, current.x, current.y)
  return segments
}

export function getScreen4RunwayHopProgress(
  points: Screen4RunwayPoint[],
  from: number,
  to: number,
  ratio: number,
): Screen4RunwayHopProgress | null {
  const startIndex = Math.max(0, from)
  const endIndex = Math.min(to, points.length - 1)
  if (startIndex >= endIndex || !points[startIndex]) return null

  const segments: Screen4RunwaySegment[] = []
  for (let index = startIndex; index < endIndex; index += 1) {
    segments.push(...buildScreen4RunwayHopSegments(points[index], points[index + 1]))
  }
  if (!segments.length) return null

  const totalLength = segments.reduce((sum, segment) => sum + segment.length, 0)
  const targetLength = totalLength * Math.min(1, Math.max(0, ratio))

  let walked = 0
  let cursor = { x: points[startIndex].x, y: points[startIndex].y }
  let angle = Math.atan2(
    segments[0].y2 - segments[0].y1,
    segments[0].x2 - segments[0].x1,
  ) * 180 / Math.PI
  for (const segment of segments) {
    if (walked + segment.length >= targetLength) {
      const ratioInSegment = (targetLength - walked) / segment.length
      cursor = {
        x: segment.x1 + (segment.x2 - segment.x1) * ratioInSegment,
        y: segment.y1 + (segment.y2 - segment.y1) * ratioInSegment,
      }
      angle = Math.atan2(segment.y2 - segment.y1, segment.x2 - segment.x1) * 180 / Math.PI
      break
    }
    walked += segment.length
  }

  return {
    strokeLength: Math.round(targetLength * 100) / 100,
    cursor: {
      x: Math.round(cursor.x * 100) / 100,
      y: Math.round(cursor.y * 100) / 100,
    },
    angle: Math.round(angle * 100) / 100,
  }
}

export function getScreen4RunwayProgress(statuses: string[]): { completed: number; total: number; ratio: number } {
  const total = statuses.length
  const completed = Math.min(
    total,
    statuses.filter(status => status === 'done' || status === 'skipped' || status === 'issue').length,
  )

  return {
    completed,
    total,
    ratio: total ? completed / total : 0,
  }
}

export function getScreen4PhaseStepProgress(
  nodes: Array<Pick<Screen4RunwayNode, 'completed' | 'total'>>,
): { completed: number; total: number; percent: number } {
  const total = nodes.reduce((sum, node) => sum + Math.max(0, node.total), 0)
  const completed = Math.min(
    total,
    nodes.reduce((sum, node) => sum + Math.max(0, Math.min(node.completed, node.total)), 0),
  )

  return {
    completed,
    total,
    percent: total ? Math.round((completed / total) * 100) : 0,
  }
}

export function truncateScreen4RunwayText(text: string, maxLength = 12): string {
  const characters = Array.from(text.trim())
  return characters.length > maxLength
    ? `${characters.slice(0, maxLength).join('')}…`
    : characters.join('')
}

export function splitScreen4RunwayName(name: string): string[] {
  const normalizedName = name.trim() || '未命名环节'
  const characters = Array.from(truncateScreen4RunwayText(normalizedName))

  if (characters.length <= 6) return [normalizedName]

  const firstLineLength = characters.length <= 12 ? Math.ceil(characters.length / 2) : 6
  const firstLine = characters.slice(0, firstLineLength).join('')
  const secondLine = characters.slice(firstLineLength).join('')

  return [firstLine, secondLine]
}
