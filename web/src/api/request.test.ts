import axios, { AxiosError, type AxiosAdapter } from 'axios'
import { beforeEach, afterEach, describe, expect, it, vi } from 'vitest'

const messages = vi.hoisted(() => ({ error: vi.fn(), success: vi.fn() }))
vi.mock('element-plus', () => ({ ElMessage: messages }))

describe('apiRequest error options', () => {
  const storage = new Map<string, string>()
  const adapter = vi.fn<AxiosAdapter>()
  let apiRequest: typeof import('./request').apiRequest

  beforeEach(async () => {
    vi.resetModules()
    vi.clearAllMocks()
    storage.clear()
    storage.set('drill_auth', JSON.stringify({ access_token: 'session-token' }))
    vi.stubGlobal('localStorage', {
      getItem: (key: string) => storage.get(key) ?? null,
      removeItem: (key: string) => storage.delete(key),
    })
    vi.stubGlobal('window', { location: { href: '/admin' } })
    axios.defaults.adapter = adapter
    apiRequest = (await import('./request')).apiRequest
  })

  afterEach(() => vi.unstubAllGlobals())

  function respond(status: number, code = 0) {
    adapter.mockImplementation(async config => {
      const response = { status, statusText: '', headers: {}, config, data: { code, message: 'backend failure', data: { count: 0 } } }
      if (status >= 400) throw new AxiosError('failure', 'ERR_BAD_RESPONSE', config, undefined, response)
      return response
    })
  }

  it('unwraps genuine zero and sends bearer auth with client-only silent option', async () => {
    respond(200)
    await expect(apiRequest({ url: '/v1/dashboard/team', method: 'GET', silentError: true })).resolves.toEqual({ count: 0 })
    const config = adapter.mock.calls[0][0]
    expect(config.headers.Authorization).toBe('Bearer session-token')
    expect(config.headers.has('silentError')).toBe(false)
  })

  it.each([403, 500])('suppresses a silent HTTP %s toast but still rejects', async status => {
    respond(status)
    await expect(apiRequest({ url: '/v1/dashboard/team', method: 'GET', silentError: true })).rejects.toThrow('backend failure')
    expect(messages.error).not.toHaveBeenCalled()
  })

  it('suppresses silent business errors while keeping ordinary business errors visible', async () => {
    respond(200, 1001)
    await expect(apiRequest({ url: '/v1/dashboard/team', method: 'GET', silentError: true })).rejects.toThrow('backend failure')
    expect(messages.error).not.toHaveBeenCalled()
    await expect(apiRequest({ url: '/v1/drills', method: 'GET' })).rejects.toThrow('backend failure')
    expect(messages.error).toHaveBeenCalledWith('backend failure')
  })

  it.each([403, 500])('preserves ordinary HTTP %s toast behavior', async status => {
    respond(status)
    await expect(apiRequest({ url: '/v1/drills', method: 'GET' })).rejects.toThrow('backend failure')
    expect(messages.error).toHaveBeenCalledWith('backend failure')
  })

  it('preserves authentication clearing and redirect for silent 401', async () => {
    respond(401)
    await expect(apiRequest({ url: '/v1/auth/heartbeat', method: 'POST', silentError: true })).rejects.toThrow('backend failure')
    expect(storage.has('drill_auth')).toBe(false)
    expect(window.location.href).toBe('/login')
  })

  it('keeps the login 401 exception', async () => {
    respond(401)
    await expect(apiRequest({ url: '/v1/auth/login', method: 'POST', silentError: true })).rejects.toThrow('backend failure')
    expect(storage.has('drill_auth')).toBe(true)
    expect(window.location.href).toBe('/admin')
  })
})
