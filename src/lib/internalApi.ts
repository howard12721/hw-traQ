export const INTERNAL_PING_PATH = '/internal/v1/ping'
export const INTERNAL_ME_PATH = '/internal/v1/me'
export const INTERNAL_GAZER_PATH = '/internal/v1/gazer'
export const INTERNAL_GAZER_TOKEN_PATH = '/internal/v1/gazer/token'
export const INTERNAL_GAZER_OAUTH_CLIENT_PATH =
  '/internal/v1/gazer/oauth-client'

type InternalPingResponse = {
  message: string
}

export type InternalUser = {
  id: string
  name: string
  displayName: string
  iconFileId: string
  bot: boolean
  state: number
  permissions?: string[]
  groups?: string[]
}

type InternalMeResponse = {
  user: InternalUser
}

export type GazerEntry = {
  id?: number
  pattern: string
  includeSelf: boolean
  includeBots: boolean
}

export type GazerSetting = {
  entries: GazerEntry[]
  enabled: boolean
}

export type GazerStatus = {
  running: boolean
  tokenConfigured: boolean
}

export type GazerResponse = {
  setting: GazerSetting
  status: GazerStatus
}

type GazerOAuthClientResponse = {
  clientId: string
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

export const getInternalMe = async () => {
  const res = await fetch(INTERNAL_ME_PATH, {
    cache: 'no-store',
    credentials: 'same-origin',
    headers: {
      Accept: 'application/json'
    }
  })

  if (!res.ok) {
    throw new Error('internal backend auth failed')
  }

  return (await res.json()) as InternalMeResponse
}

export const getGazer = async () => {
  const res = await fetch(INTERNAL_GAZER_PATH, {
    cache: 'no-store',
    credentials: 'same-origin',
    headers: {
      Accept: 'application/json'
    }
  })

  if (!res.ok) {
    throw new Error('failed to get gazer setting')
  }

  return (await res.json()) as GazerResponse
}

export const putGazer = async (setting: { entries: GazerEntry[] }) => {
  const res = await fetch(INTERNAL_GAZER_PATH, {
    method: 'PUT',
    cache: 'no-store',
    credentials: 'same-origin',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(setting)
  })

  if (!res.ok) {
    throw new Error('failed to save gazer setting')
  }

  return (await res.json()) as GazerResponse
}

export const putGazerToken = async (accessToken: string) => {
  const res = await fetch(INTERNAL_GAZER_TOKEN_PATH, {
    method: 'PUT',
    cache: 'no-store',
    credentials: 'same-origin',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ accessToken })
  })

  if (!res.ok) {
    throw new Error('failed to save gazer token')
  }

  return (await res.json()) as GazerResponse
}

export const getGazerOAuthClient = async () => {
  const res = await fetch(INTERNAL_GAZER_OAUTH_CLIENT_PATH, {
    cache: 'no-store',
    credentials: 'same-origin',
    headers: {
      Accept: 'application/json'
    }
  })

  if (!res.ok) {
    throw new Error('failed to get gazer oauth client')
  }

  return (await res.json()) as GazerOAuthClientResponse
}
