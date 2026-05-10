export const darkTheme = {
  color: ['#55C3D3', '#2EA043', '#B8860B', '#DA3633', '#7D7D7D', '#646464', '#7DD4E0', '#40C057'],
  backgroundColor: 'transparent',
  textStyle: {
    color: '#8B949E',
    fontFamily: 'Fira Sans, sans-serif',
  },
  title: {
    textStyle: { color: '#E0E6ED' },
    subtextStyle: { color: '#8B949E' },
  },
  lineStyle: { width: 2 },
  splitLine: {
    lineStyle: { color: '#1A1F2E', type: 'dashed' as const },
  },
  axisLine: {
    lineStyle: { color: '#30363D' },
  },
  axisLabel: {
    color: '#8B949E',
    fontSize: 12,
  },
  tooltip: {
    backgroundColor: '#161B22',
    borderColor: '#30363D',
    textStyle: { color: '#E0E6ED' },
    axisPointer: {
      lineStyle: { color: '#55C3D3' },
    },
  },
  legend: {
    textStyle: { color: '#8B949E' },
  },
}

export function applyDarkTheme(option: Record<string, unknown>): Record<string, unknown> {
  return {
    ...option,
    textStyle: { ...darkTheme.textStyle, ...(option.textStyle || {}) },
    grid: {
      top: 40,
      right: 20,
      bottom: 40,
      left: 50,
      containLabel: true,
      ...(option.grid || {}),
    },
  }
}
