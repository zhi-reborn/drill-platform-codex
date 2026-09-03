import { defineComponent, h } from 'vue'

const Container = defineComponent({
  setup(_props, { slots }) { return () => h('div', [slots.header?.(), slots.default?.()]) },
})
const Empty = defineComponent({ setup() { return () => null } })

export const ElRow = Container
export const ElCol = Container
export const ElCard = Container
export const ElTable = Empty
export const ElTableColumn = Empty
export const ElTag = Container
export const ElProgress = Empty
export const ElButton = Container
export const ElIcon = Container
export const ElTooltip = Container
export const ElConfigProvider = Container
