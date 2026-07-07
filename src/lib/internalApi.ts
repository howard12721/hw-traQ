export const INTERNAL_PING_PATH = '/internal/v1/ping'
export const INTERNAL_ME_PATH = '/internal/v1/me'
export const INTERNAL_GAZER_PATH = '/internal/v1/gazer'
export const INTERNAL_GAZER_TOKEN_PATH = '/internal/v1/gazer/token'
export const INTERNAL_GAZER_OAUTH_CLIENT_PATH =
  '/internal/v1/gazer/oauth-client'
export const INTERNAL_GAZER_NOTIFICATIONS_PATH =
  '/internal/v1/gazer/notifications'
export const INTERNAL_GAZER_NOTIFICATIONS_READ_PATH =
  '/internal/v1/gazer/notifications/read'
export const INTERNAL_SCHEDULED_MESSAGES_PATH =
  '/internal/v1/scheduled-messages'

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
  displayName: string
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
  botUserId?: string
}

export type GazerResponse = {
  setting: GazerSetting
  status: GazerStatus
}

export type GazerNotificationItem = {
  id: number
  messageId: string
  channelId: string
  authorId: string
  content: string
  pattern: string
  displayName: string
  createdAt: string
  notifiedAt: string
  read: boolean
}

export type GazerNotificationsResponse = {
  notifications: GazerNotificationItem[]
  botUserId?: string
}

export type ScheduledMessageItem = {
  id: string
  channelId: string
  content: string
  scheduledAt: string
  createdAt: string
  retryAt?: string
  lastError?: string
  failedAttempts?: number
}

export type ScheduledMessagesResponse = {
  messages: ScheduledMessageItem[]
}

export type ScheduledMessageResponse = {
  message: ScheduledMessageItem
}

export type CreateScheduledMessageRequest = {
  channelId: string
  content: string
  scheduledAt: string
}

export type GazerTokenRequest = {
  code: string
  codeVerifier: string
  redirectUri: string
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

export const putGazerToken = async (request: GazerTokenRequest) => {
  const res = await fetch(INTERNAL_GAZER_TOKEN_PATH, {
    method: 'PUT',
    cache: 'no-store',
    credentials: 'same-origin',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(request)
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

export const getGazerNotifications = async () => {
  const res = await fetch(INTERNAL_GAZER_NOTIFICATIONS_PATH, {
    cache: 'no-store',
    credentials: 'same-origin',
    headers: {
      Accept: 'application/json'
    }
  })

  if (!res.ok) {
    throw new Error('failed to get gazer notifications')
  }

  return (await res.json()) as GazerNotificationsResponse
}

export const markGazerNotificationsRead = async () => {
  const res = await fetch(INTERNAL_GAZER_NOTIFICATIONS_READ_PATH, {
    method: 'POST',
    cache: 'no-store',
    credentials: 'same-origin'
  })

  if (!res.ok) {
    throw new Error('failed to mark gazer notifications as read')
  }
}

export const getScheduledMessages = async () => {
  const res = await fetch(INTERNAL_SCHEDULED_MESSAGES_PATH, {
    cache: 'no-store',
    credentials: 'same-origin',
    headers: {
      Accept: 'application/json'
    }
  })

  if (!res.ok) {
    throw new Error('failed to get scheduled messages')
  }

  return (await res.json()) as ScheduledMessagesResponse
}

export const postScheduledMessage = async (
  request: CreateScheduledMessageRequest
) => {
  const res = await fetch(INTERNAL_SCHEDULED_MESSAGES_PATH, {
    method: 'POST',
    cache: 'no-store',
    credentials: 'same-origin',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(request)
  })

  if (!res.ok) {
    throw new Error('failed to create scheduled message')
  }

  return (await res.json()) as ScheduledMessageResponse
}

export const deleteScheduledMessage = async (id: string) => {
  const res = await fetch(`${INTERNAL_SCHEDULED_MESSAGES_PATH}/${id}`, {
    method: 'DELETE',
    cache: 'no-store',
    credentials: 'same-origin'
  })

  if (!res.ok) {
    throw new Error('failed to delete scheduled message')
  }
}
