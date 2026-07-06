import {
  GAZER_NOTIFICATION_PREFIX,
  parseGazerNotification
} from '/@/lib/gazerNotification'

const payload = {
  messageId: 'message-id',
  channelId: 'channel-id',
  authorId: 'author-id',
  content: '障害対応お願いします\n二行目です',
  pattern: '障害|deploy',
  createdAt: '2026-07-06T12:34:56.000Z'
}

const encodePayload = (value: unknown) =>
  Buffer.from(JSON.stringify(value), 'utf8').toString('base64url')

describe('gazerNotification', () => {
  it('parses gazer notification payload', () => {
    const content = `${GAZER_NOTIFICATION_PREFIX}${encodePayload(
      payload
    )} -->\nGazer matched`

    expect(parseGazerNotification(content)).toEqual(payload)
  })

  it('ignores normal messages', () => {
    expect(parseGazerNotification('hello')).toBeUndefined()
  })

  it('ignores invalid payloads', () => {
    expect(
      parseGazerNotification(`${GAZER_NOTIFICATION_PREFIX}invalid -->`)
    ).toBeUndefined()
  })
})
