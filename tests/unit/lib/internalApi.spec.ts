import {
  INTERNAL_GAZER_OAUTH_CLIENT_PATH,
  INTERNAL_GAZER_PATH,
  INTERNAL_GAZER_TOKEN_PATH,
  INTERNAL_ME_PATH,
  INTERNAL_PING_PATH,
  checkInternalBackend,
  getGazer,
  getGazerOAuthClient,
  getInternalMe,
  pingInternalBackend,
  putGazer,
  putGazerToken
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

  it('gets authenticated internal user', async () => {
    fetchMock.mockResolvedValue(
      new Response(
        JSON.stringify({
          user: {
            id: 'user-id',
            name: 'howard127',
            displayName: 'Howard',
            iconFileId: 'icon-file-id',
            bot: false,
            state: 1
          }
        }),
        {
          status: 200,
          headers: {
            'Content-Type': 'application/json'
          }
        }
      )
    )

    await expect(getInternalMe()).resolves.toEqual({
      user: {
        id: 'user-id',
        name: 'howard127',
        displayName: 'Howard',
        iconFileId: 'icon-file-id',
        bot: false,
        state: 1
      }
    })
    expect(fetchMock).toHaveBeenCalledWith(INTERNAL_ME_PATH, {
      cache: 'no-store',
      credentials: 'same-origin',
      headers: {
        Accept: 'application/json'
      }
    })
  })

  it('rejects failed authenticated internal user responses', async () => {
    fetchMock.mockResolvedValue(new Response('', { status: 401 }))

    await expect(getInternalMe()).rejects.toThrow(
      'internal backend auth failed'
    )
  })

  it('gets gazer setting', async () => {
    fetchMock.mockResolvedValue(
      new Response(
        JSON.stringify({
          setting: {
            entries: [
              {
                id: 1,
                pattern: 'foo',
                includeSelf: false,
                includeBots: true
              }
            ],
            enabled: true
          },
          status: {
            running: true,
            tokenConfigured: true
          }
        }),
        {
          status: 200,
          headers: {
            'Content-Type': 'application/json'
          }
        }
      )
    )

    await expect(getGazer()).resolves.toEqual({
      setting: {
        entries: [
          {
            id: 1,
            pattern: 'foo',
            includeSelf: false,
            includeBots: true
          }
        ],
        enabled: true
      },
      status: {
        running: true,
        tokenConfigured: true
      }
    })
    expect(fetchMock).toHaveBeenCalledWith(INTERNAL_GAZER_PATH, {
      cache: 'no-store',
      credentials: 'same-origin',
      headers: {
        Accept: 'application/json'
      }
    })
  })

  it('gets gazer oauth client', async () => {
    fetchMock.mockResolvedValue(
      new Response(JSON.stringify({ clientId: 'client-id' }), {
        status: 200,
        headers: {
          'Content-Type': 'application/json'
        }
      })
    )

    await expect(getGazerOAuthClient()).resolves.toEqual({
      clientId: 'client-id'
    })
    expect(fetchMock).toHaveBeenCalledWith(INTERNAL_GAZER_OAUTH_CLIENT_PATH, {
      cache: 'no-store',
      credentials: 'same-origin',
      headers: {
        Accept: 'application/json'
      }
    })
  })

  it('saves gazer setting', async () => {
    fetchMock.mockResolvedValue(
      new Response(
        JSON.stringify({
          setting: {
            entries: [
              {
                id: 1,
                pattern: 'foo',
                includeSelf: true,
                includeBots: false
              },
              {
                id: 2,
                pattern: 'bar',
                includeSelf: false,
                includeBots: true
              }
            ],
            enabled: true
          },
          status: {
            running: true,
            tokenConfigured: true
          }
        }),
        {
          status: 200,
          headers: {
            'Content-Type': 'application/json'
          }
        }
      )
    )

    await expect(
      putGazer({
        entries: [
          {
            pattern: 'foo',
            includeSelf: true,
            includeBots: false
          },
          {
            pattern: 'bar',
            includeSelf: false,
            includeBots: true
          }
        ]
      })
    ).resolves.toEqual({
      setting: {
        entries: [
          {
            id: 1,
            pattern: 'foo',
            includeSelf: true,
            includeBots: false
          },
          {
            id: 2,
            pattern: 'bar',
            includeSelf: false,
            includeBots: true
          }
        ],
        enabled: true
      },
      status: {
        running: true,
        tokenConfigured: true
      }
    })
    expect(fetchMock).toHaveBeenCalledWith(INTERNAL_GAZER_PATH, {
      method: 'PUT',
      cache: 'no-store',
      credentials: 'same-origin',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        entries: [
          {
            pattern: 'foo',
            includeSelf: true,
            includeBots: false
          },
          {
            pattern: 'bar',
            includeSelf: false,
            includeBots: true
          }
        ]
      })
    })
  })

  it('saves gazer access token', async () => {
    fetchMock.mockResolvedValue(
      new Response(
        JSON.stringify({
          setting: {
            entries: [
              {
                id: 1,
                pattern: 'foo',
                includeSelf: true,
                includeBots: false
              }
            ],
            enabled: true
          },
          status: {
            running: true,
            tokenConfigured: true
          }
        }),
        {
          status: 200,
          headers: {
            'Content-Type': 'application/json'
          }
        }
      )
    )

    await expect(putGazerToken('access-token')).resolves.toEqual({
      setting: {
        entries: [
          {
            id: 1,
            pattern: 'foo',
            includeSelf: true,
            includeBots: false
          }
        ],
        enabled: true
      },
      status: {
        running: true,
        tokenConfigured: true
      }
    })
    expect(fetchMock).toHaveBeenCalledWith(INTERNAL_GAZER_TOKEN_PATH, {
      method: 'PUT',
      cache: 'no-store',
      credentials: 'same-origin',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        accessToken: 'access-token'
      })
    })
  })
})
