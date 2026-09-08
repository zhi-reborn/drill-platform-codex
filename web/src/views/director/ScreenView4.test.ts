import { describe, expect, it } from 'vitest'
import { existsSync, readFileSync } from 'node:fs'
import { fileURLToPath, URL } from 'node:url'

const screen4Path = fileURLToPath(new URL('./ScreenView4.vue', import.meta.url))
const monitorPath = fileURLToPath(new URL('./MonitorView.vue', import.meta.url))
const routerPath = fileURLToPath(new URL('../../router/index.ts', import.meta.url))

describe('independent screen4 entry', () => {
  it('owns an independent page implementation', () => {
    expect(existsSync(screen4Path)).toBe(true)
    const source = readFileSync(screen4Path, 'utf8')
    expect(source).toContain('class="screen-root cyber-command-screen"')
    expect(source).toContain("from '@/api/modules/drill'")
    expect(source).not.toMatch(/import\s+ScreenView2|<ScreenView2/)
  })

  it('places the screen4 button after screen3 and opens the independent route', () => {
    const source = readFileSync(monitorPath, 'utf8')
    expect(source).toMatch(/@click="viewScreen3"[\s\S]*?大屏3[\s\S]*?@click="viewScreen4"[\s\S]*?大屏4/)
    expect(source).toContain('window.open(`/director/screen4/${drillId.value}`, \'_blank\')')
    expect(source).toContain('.screen-entry-aqua')
  })

  it('routes screen4 to its own component', () => {
    const source = readFileSync(routerPath, 'utf8')
    expect(source).toMatch(/path: 'screen4\/:id\(\\\\d\+\)'[\s\S]*?name: 'DirectorScreen4'[\s\S]*?ScreenView4\.vue/)
  })
})
