import assert from 'node:assert/strict'
import fs from 'node:fs'
import path from 'node:path'
import vm from 'node:vm'
import ts from 'typescript'

const root = path.resolve(import.meta.dirname, '..')
const sourcePath = path.join(root, 'src/views/director/monitorStepTree.ts')
const source = fs.readFileSync(sourcePath, 'utf8')
const compiled = ts.transpileModule(source, {
  compilerOptions: {
    module: ts.ModuleKind.CommonJS,
    target: ts.ScriptTarget.ES2020,
  },
})

const sandbox = { exports: {} }
vm.runInNewContext(compiled.outputText, sandbox, { filename: sourcePath })

const {
  createMonitorStepTreeBuilder,
  buildStepTreeMetaMap,
  patchMonitorStepList,
  getDefaultExpandedStepKeys,
  flattenVisibleStepTree,
} = sandbox.exports

function step(id, parentStepId, status = 'pending') {
  return {
    id,
    name: `步骤${id}`,
    seq: id,
    parent_step_id: parentStepId,
    status,
    step_type: 'serial',
    phase: '阶段',
    phase_step: '环节',
    timeout_minutes: 0,
    default_assignee_role: 'director',
    estimated_duration_minutes: 0,
    attributes: {},
    start_time: null,
    end_time: null,
    timeout_at: null,
    assignee_names: '',
    remark: '',
    issue_desc: '',
    executor_team: '',
    step_template_id: id,
    drill_instance_id: 1,
  }
}

const builder = createMonitorStepTreeBuilder()
const initialSteps = [
  step(1, null, 'running'),
  step(2, 1, 'running'),
  step(3, 1, 'pending'),
  step(4, null, 'pending'),
]

const firstTree = builder.build(initialSteps)
const firstRoot = firstTree[0]
const firstChangedChild = firstRoot.children[0]
const firstUnchangedChild = firstRoot.children[1]

const nextSteps = [...initialSteps]
nextSteps[1] = { ...nextSteps[1], status: 'completed' }
const nextTree = builder.build(nextSteps)
const metaMap = buildStepTreeMetaMap(nextTree)

assert.equal(nextTree[0], firstRoot)
assert.equal(nextTree[0].children[1], firstUnchangedChild)
assert.notEqual(nextTree[0].children[0], firstChangedChild)
assert.equal(metaMap.get(1).statusText.text, '1/2 子任务已完成')

const patched = patchMonitorStepList(initialSteps, {
  step_id: 2,
  new_status: 'completed',
  end_time: '2026-06-23T12:00:00.000Z',
})

assert.ok(patched)
assert.equal(patched.previous, initialSteps[1])
assert.notEqual(patched.steps, initialSteps)
assert.equal(patched.steps[0], initialSteps[0])
assert.notEqual(patched.steps[1], initialSteps[1])
assert.equal(patched.steps[1].status, 'completed')
assert.equal(patched.steps[1].end_time, '2026-06-23T12:00:00.000Z')

const expandedKeys = getDefaultExpandedStepKeys([
  step(10, null, 'pending'),
  step(11, 10, 'pending'),
  step(12, 11, 'running'),
  step(20, null, 'pending'),
  step(21, 20, 'pending'),
])
assert.deepEqual(expandedKeys, [10, 11, 20])

const flatVisible = flattenVisibleStepTree(firstTree, [1])
assert.equal(JSON.stringify(flatVisible.map((node) => node.id)), JSON.stringify([1, 2, 3, 4]))
assert.equal(flatVisible[0].children, undefined)

console.log('monitor step tree ok')
