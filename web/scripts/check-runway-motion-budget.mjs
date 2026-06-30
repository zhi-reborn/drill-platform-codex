import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'

const here = dirname(fileURLToPath(import.meta.url))
const source = readFileSync(resolve(here, '../src/components/screen/PhaseRing.vue'), 'utf8')

const forbidden = [
  ['SVG animateMotion', /<animateMotion\b/],
  ['animated SVG lane path', /class="[^"]*\blane-energy\b/],
  ['animated SVG circuit lane', /lane-flow-circuit/],
  ['runway stroke-dashoffset keyframes', /@keyframes\s+energy-run\b|@keyframes\s+dash-move\b|stroke-dashoffset/],
  ['large rotating runway sweep', /\.relay-runway::before[\s\S]*?animation:\s*runway-sweep/],
  ['filtered runway flow elements', /\.runway-flow-[^{]+{[^}]*filter\s*:/],
  ['moving runway flow beams', /runway-svg-flow-beam/],
]

const failures = forbidden.filter(([, pattern]) => pattern.test(source))
const required = [
  ['runway short center dashes', /stroke-dasharray="22 28"[\s\S]*class="lane-dash"/],
  ['svg-aligned runway turn short dashes', /<rect[\s\S]*class="runway-svg-turn-pip"/],
  ['milestone beacon rings', /finish-beacon-ring/],
  ['milestone sparks', /finish-sparks/],
  ['milestone crown', /finish-crown/],
  ['milestone anchored below last node', /x:\s*lastPoint\.x[\s\S]*y:\s*lastPoint\.y\s*\+\s*108/],
  ['transform-only turn animation', /@keyframes\s+runway-svg-turn-pip[\s\S]*transform/],
]
const missing = required.filter(([, pattern]) => !pattern.test(source))

if (failures.length > 0 || missing.length > 0) {
  console.error('Runway motion budget exceeded:')
  for (const [name] of failures) console.error(`- ${name}`)
  for (const [name] of missing) console.error(`- missing ${name}`)
  process.exit(1)
}

console.log('Runway motion budget ok')
