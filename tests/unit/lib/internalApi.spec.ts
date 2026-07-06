import {
  INTERNAL_PING_PATH,
  checkInternalBackend,
  pingInternalBackend
} from '/@/lib/internalApi'

describe('internalApi', () => {
  const fetchMock = vi.fn()

  beforeEach(() => {
    vi.stubGlobal('fetch', fetchMock)
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    fetchMock.mockReset()
  })

  it('pings internal backend', async () => {
    fetchMock.mockResolvedValue(
      new Response(JSON.stringify({ message: 'pong' }), {
        status: 200,
        headers: {
          'Content-Type': 'application/json'
        }
      })
    )

    await expect(pingInternalBackend()).resolves.toEqual({ message: 'pong' })
    expect(fetchMock).toHaveBeenCalledWith(INTERNAL_PING_PATH, {
      cache: 'no-store',
      credentials: 'same-origin',
      headers: {
        Accept: 'application/json'
      }
    })
  })

  it('accepts pong as a successful connectivity check', async () => {
    fetchMock.mockResolvedValue(
      new Response(JSON.stringify({ message: 'pong' }), {
        status: 200,
        headers: {
          'Content-Type': 'application/json'
        }
      })
    )

    await expect(checkInternalBackend()).resolves.toBeUndefined()
  })

  it('rejects unexpected responses', async () => {
    fetchMock.mockResolvedValue(
      new Response(JSON.stringify({ message: 'ok' }), {
        status: 200,
        headers: {
          'Content-Type': 'application/json'
        }
      })
    )

    await expect(checkInternalBackend()).rejects.toThrow(
      'unexpected internal backend response'
    )
  })

  it('rejects failed responses', async () => {
    fetchMock.mockResolvedValue(new Response('', { status: 503 }))

    await expect(checkInternalBackend()).rejects.toThrow(
      'internal backend ping failed'
    )
  })
})
