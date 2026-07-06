export const INTERNAL_PING_PATH = '/internal/v1/ping'

type InternalPingResponse = {
  message: string
}

export const pingInternalBackend = async () => {
  const res = await fetch(INTERNAL_PING_PATH, {
    cache: 'no-store',
    credentials: 'same-origin',
    headers: {
      Accept: 'application/json'
    }
  })

  if (!res.ok) {
    throw new Error('internal backend ping failed')
  }

  return (await res.json()) as InternalPingResponse
}

export const checkInternalBackend = async () => {
  const res = await pingInternalBackend()
  if (res.message !== 'pong') {
    throw new Error('unexpected internal backend response')
  }
}
