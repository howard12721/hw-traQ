import {
  getJstDateTimeLocalString,
  getJstFullDayWithTimeString,
  parseJstDateTimeLocalString
} from '/@/lib/basic/date'

describe('date', () => {
  describe('JST datetime helpers', () => {
    it('formats a Date as a datetime-local value in JST', () => {
      const date = new Date('2026-07-06T15:01:00.000Z')

      expect(getJstDateTimeLocalString(date)).toBe('2026-07-07T00:01')
    })

    it('formats a Date as a display string in JST', () => {
      const date = new Date('2026-07-07T03:34:00.000Z')

      expect(getJstFullDayWithTimeString(date)).toBe('2026/07/07 12:34')
    })

    it('parses a datetime-local value as JST', () => {
      const date = parseJstDateTimeLocalString('2026-07-07T09:30')

      expect(date?.toISOString()).toBe('2026-07-07T00:30:00.000Z')
    })

    it('rejects invalid datetime-local values', () => {
      expect(parseJstDateTimeLocalString('2026-02-30T10:00')).toBeUndefined()
      expect(parseJstDateTimeLocalString('invalid')).toBeUndefined()
    })
  })
})
