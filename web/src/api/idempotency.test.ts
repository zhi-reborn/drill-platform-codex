import { describe, it, expect } from 'vitest'
import { generateKey, getKeyForAction, clearKey } from './idempotency'

describe('idempotency', () => {
  it('reuses the same key when a mutation retry receives a network failure', () => {
    // 用户点击"开始演练"，生成 actionId
    const actionId = 'drill:1:start'

    // 首次尝试获取幂等键
    const firstKey = getKeyForAction(actionId)

    // 模拟网络失败后重试：同一个 actionId 应复用同一把键
    const retryKey = getKeyForAction(actionId)

    expect(retryKey).toBe(firstKey)
  })

  it('generates a new key for a new user action', () => {
    // 第一次用户操作
    const actionId1 = 'drill:2:start'
    const key1 = getKeyForAction(actionId1)

    // 另一个全新的用户操作
    const actionId2 = 'drill:3:start'
    const key2 = getKeyForAction(actionId2)

    expect(key2).not.toBe(key1)
  })

  it('generates a new key after an action is cleared and a new action begins', () => {
    // 一次操作完成
    const actionId = 'drill:4:pause'
    const firstKey = getKeyForAction(actionId)
    clearKey(actionId)

    // 同一目标的新一次用户操作（例如先暂停后又再次暂停）应使用新键
    const newKey = getKeyForAction(actionId)

    expect(newKey).not.toBe(firstKey)
  })

  it('produces uuid-shaped keys via generateKey', () => {
    const key1 = generateKey()
    const key2 = generateKey()

    expect(key1).not.toBe(key2)
    expect(key1).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i
    )
  })
})
