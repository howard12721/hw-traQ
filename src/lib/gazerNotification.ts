export const GAZER_NOTIFICATION_PREFIX = '<!-- hw-traq-gazer:'

export type GazerNotification = {
  messageId: string
  channelId: string
  authorId: string
  content: string
  pattern: string
  createdAt: string
}

const decoder = new TextDecoder()

const decodeBase64URL = (encoded: string) => {
  const base64 =
    encoded.replaceAll('-', '+').replaceAll('_', '/') +
    '='.repeat((4 - (encoded.length % 4)) % 4)
  const binary = atob(base64)
  const bytes = Uint8Array.from(binary, char => char.charCodeAt(0))
  return decoder.decode(bytes)
}

export const parseGazerNotification = (
  content: string
): GazerNotification | undefined => {
  if (!content.startsWith(GAZER_NOTIFICATION_PREFIX)) return undefined

  const end = content.indexOf(' -->')
  if (end === -1) return undefined

  const encoded = content.slice(GAZER_NOTIFICATION_PREFIX.length, end)
  try {
    const json = decodeBase64URL(encoded)
    const parsed = JSON.parse(json) as Partial<GazerNotification>
    if (
      typeof parsed.messageId !== 'string' ||
      typeof parsed.channelId !== 'string' ||
      typeof parsed.authorId !== 'string' ||
      typeof parsed.content !== 'string' ||
      typeof parsed.pattern !== 'string' ||
      typeof parsed.createdAt !== 'string'
    ) {
      return undefined
    }
    return parsed as GazerNotification
  } catch {
    return undefined
  }
}
