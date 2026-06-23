import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import vm from 'node:vm'
import ts from 'typescript'

const root = path.resolve(import.meta.dirname, '..')
const sourcePath = path.join(root, 'src/views/director/monitorRefreshPolicy.ts')
const source = fs.readFileSync(sourcePath, 'utf8')
const compiled = ts.transpileModule(source, {
  compilerOptions: {
    module: ts.ModuleKind.CommonJS,
    target: ts.ScriptTarget.ES2020,
  },
})

const sandbox = { exports: {} }
vm.runInNewContext(compiled.outputText, sandbox, { filename: sourcePath })

const { getMonitorRefreshPlan } = sandbox.exports

function assertPlan(reason, expected) {
  assert.equal(JSON.stringify(getMonitorRefreshPlan(reason)), JSON.stringify(expected))
}

assertPlan('initial-load', {
  detail: true,
  steps: true,
  logs: true,
})

assertPlan('fallback-poll', {
  detail: false,
  steps: true,
  logs: true,
})

assertPlan('step-action-fallback', {
  detail: false,
  steps: true,
  logs: true,
})

assertPlan('drill-event', {
  detail: true,
  steps: false,
  logs: true,
})

console.log('monitor refresh policy ok')
