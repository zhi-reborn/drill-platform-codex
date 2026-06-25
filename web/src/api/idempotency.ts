// 幂等键管理：同一用户操作（actionId）在重试时复用同一把键，
// 操作完成后清除，新一次操作生成新键。

const actionKeys = new Map<string, string>()

function randomUUID(): string {
  if (typeof globalThis.crypto?.randomUUID === 'function') {
    return globalThis.crypto.randomUUID()
  }
  // 旧环境兜底
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0
    const v = c === 'x' ? r : (r & 0x3) | 0x8
    return v.toString(16)
  })
}

export function generateKey(): string {
  return randomUUID()
}

export function getKeyForAction(actionId: string): string {
  let key = actionKeys.get(actionId)
  if (!key) {
    key = generateKey()
    actionKeys.set(actionId, key)
  }
  return key
}

export function clearKey(actionId: string): void {
  actionKeys.delete(actionId)
}
