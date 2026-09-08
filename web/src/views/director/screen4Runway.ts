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

const VIEWBOX_WIDTH = 1040
const LANE_CAPACITY = 4
const LANE_LEFT = 90
const LANE_RIGHT = 950
const LANE_PADDING = 70
const TURN_LEFT = 42
const TURN_RIGHT = 998
const FIRST_LANE_Y = 120
const LANE_GAP = 190

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
    const middleY = (previous.y + current.y) / 2
    path += ` L ${turnX} ${previous.y} Q ${turnX} ${middleY} ${turnX} ${current.y} L ${current.x} ${current.y}`
  }

  return path
}

export function buildScreen4RunwayLayout(nodeCount: number): Screen4RunwayLayout {
  const safeNodeCount = Math.max(0, Math.floor(nodeCount))
  const itemCount = safeNodeCount + 1
  const rowCount = Math.max(1, Math.ceil(itemCount / LANE_CAPACITY))
  const laneCounts = distributeAcrossLanes(itemCount, rowCount)
  const laneY = Array.from({ length: rowCount }, (_, row) => (
    rowCount === 1 ? 170 : FIRST_LANE_Y + row * LANE_GAP
  ))
  const points: Screen4RunwayPoint[] = []

  laneCounts.forEach((count, row) => {
    const direction: Screen4RunwayDirection = row % 2 === 0 ? 'right' : 'left'
    const usableLeft = LANE_LEFT + LANE_PADDING
    const usableRight = LANE_RIGHT - LANE_PADDING
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
    viewBoxHeight: laneY[laneY.length - 1] + 130,
    laneY,
    points,
    trackPath: getScreen4RunwayPath(points, 0, points.length - 1),
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

export function splitScreen4RunwayName(name: string): string[] {
  const normalizedName = name.trim() || '未命名环节'
  const characters = Array.from(normalizedName)

  if (characters.length <= 6) return [normalizedName]

  const firstLineLength = characters.length <= 12 ? Math.ceil(characters.length / 2) : 6
  const firstLine = characters.slice(0, firstLineLength).join('')
  const remaining = characters.slice(firstLineLength)
  const secondLine = remaining.length > 5
    ? `${remaining.slice(0, 5).join('')}…`
    : remaining.join('')

  return [firstLine, secondLine]
}
