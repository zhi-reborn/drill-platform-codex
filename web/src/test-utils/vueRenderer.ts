import { compile, createRenderer, ssrContextKey, type Component, type ComponentOptions } from 'vue'
import { compileScript, parse } from '@vue/compiler-sfc'

export interface TestNode {
  type: string
  text: string
  props: Record<string, unknown>
  children: TestNode[]
  parent: TestNode | null
}

function node(type: string, text = ''): TestNode {
  return { type, text, props: {}, children: [], parent: null }
}

const renderer = createRenderer<TestNode, TestNode>({
  createElement: type => node(type),
  createText: text => node('#text', text),
  createComment: text => node('#comment', text),
  setText: (target, text) => { target.text = text },
  setElementText: (target, text) => { target.text = text; target.children = [] },
  parentNode: target => target.parent,
  nextSibling: target => {
    const siblings = target.parent?.children ?? []
    return siblings[siblings.indexOf(target) + 1] ?? null
  },
  patchProp: (target, key, _previous, value) => { target.props[key] = value },
  insert: (target, parent, anchor = null) => {
    if (target.parent) {
      target.parent.children.splice(target.parent.children.indexOf(target), 1)
    }
    target.parent = parent
    const index = anchor ? parent.children.indexOf(anchor) : -1
    if (index < 0) parent.children.push(target)
    else parent.children.splice(index, 0, target)
  },
  remove: target => {
    target.parent?.children.splice(target.parent.children.indexOf(target), 1)
    target.parent = null
  },
})

export function createTestApp(component: Component) {
  const root = node('root')
  const app = renderer.createApp(component)
  app.provide(ssrContextKey, {})
  return { app, root, mount: () => app.mount(root) }
}

// Vitest's Node environment compiles SFCs for SSR. Reuse the real setup and
// compile its original template for the in-memory renderer and lifecycle tests.
export function withClientTemplate(component: Component, source: string): Component {
  const { descriptor } = parse(source)
  const { bindings } = compileScript(descriptor, { id: 'test-component' })
  ;(component as ComponentOptions).render = compile(descriptor.template!.content, { bindingMetadata: bindings, prefixIdentifiers: true })
  return component
}

export function nodesWithClass(root: TestNode, className: string): TestNode[] {
  const matches = String(root.props.class ?? '').split(' ').includes(className) ? [root] : []
  return matches.concat(root.children.flatMap(child => nodesWithClass(child, className)))
}

export function nodeText(root: TestNode): string {
  if (root.type === '#comment') return ''
  return root.text + root.children.map(nodeText).join('')
}

export function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason: unknown) => void
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise
    reject = rejectPromise
  })
  return { promise, resolve, reject }
}
