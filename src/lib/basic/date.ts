export const getMonthString = (date: Readonly<Date>) =>
  (date.getMonth() + 1).toString().padStart(2, '0')

export const getDateString = (date: Readonly<Date>) =>
  date.getDate().toString().padStart(2, '0')

export const getHoursString = (date: Readonly<Date>) =>
  date.getHours().toString().padStart(2, '0')

export const getMinutesString = (date: Readonly<Date>) =>
  date.getMinutes().toString().padStart(2, '0')

export const getSecondsString = (date: Readonly<Date>) =>
  date.getSeconds().toString().padStart(2, '0')

export const getTimeString = (date: Readonly<Date>) =>
  `${getHoursString(date)}:${getMinutesString(date)}`

export const getDayString = (date: Readonly<Date>) =>
  `${getMonthString(date)}/${getDateString(date)}`

export const getFullDayString = (date: Readonly<Date>) =>
  `${date.getFullYear()}/${getDayString(date)}`

export const getFullDayWithTimeString = (date: Readonly<Date>) =>
  `${getFullDayString(date)} ${getTimeString(date)}`

export const getISOFullDayString = (date: Readonly<Date>) =>
  date.toISOString().split('T')[0]

const JST_OFFSET_MINUTES = 9 * 60
const JST_OFFSET_MS = JST_OFFSET_MINUTES * 60 * 1000

const getJstDate = (date: Readonly<Date>) =>
  new Date(date.getTime() + JST_OFFSET_MS)

const getJstYearString = (date: Readonly<Date>) =>
  getJstDate(date).getUTCFullYear().toString().padStart(4, '0')

const getJstMonthString = (date: Readonly<Date>) =>
  (getJstDate(date).getUTCMonth() + 1).toString().padStart(2, '0')

const getJstDateString = (date: Readonly<Date>) =>
  getJstDate(date).getUTCDate().toString().padStart(2, '0')

const getJstHoursString = (date: Readonly<Date>) =>
  getJstDate(date).getUTCHours().toString().padStart(2, '0')

const getJstMinutesString = (date: Readonly<Date>) =>
  getJstDate(date).getUTCMinutes().toString().padStart(2, '0')

export const getJstDateTimeLocalString = (date: Readonly<Date>) =>
  `${getJstYearString(date)}-${getJstMonthString(date)}-${getJstDateString(
    date
  )}T${getJstHoursString(date)}:${getJstMinutesString(date)}`

export const getJstFullDayWithTimeString = (date: Readonly<Date>) =>
  `${getJstYearString(date)}/${getJstMonthString(date)}/${getJstDateString(
    date
  )} ${getJstHoursString(date)}:${getJstMinutesString(date)}`

export const parseJstDateTimeLocalString = (dateTime: string) => {
  const match =
    /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})$/.exec(dateTime)
  if (!match) return undefined

  const [, year, month, day, hours, minutes] = match
  const parsed = new Date(
    Date.UTC(
      Number(year),
      Number(month) - 1,
      Number(day),
      Number(hours),
      Number(minutes)
    ) - JST_OFFSET_MS
  )

  if (getJstDateTimeLocalString(parsed) !== dateTime) {
    return undefined
  }
  return parsed
}

export const getCurrentTimeString = () => getTimeString(new Date())

/**
 * 2つの日時を比べ、差異がない部分については省略したものを出力する
 * @param ofDate 出力する日時
 * @param fromDate 比較する日時
 */
export const getDateRepresentationWithoutSameDate = (
  ofDate: Readonly<Date>,
  fromDate: Readonly<Date>
) => {
  const timeString = getTimeString(ofDate)
  if (fromDate.getFullYear() !== ofDate.getFullYear()) {
    return getFullDayString(ofDate) + ' ' + timeString
  }
  if (
    fromDate.getDate() !== ofDate.getDate() ||
    fromDate.getMonth() !== ofDate.getMonth()
  ) {
    return `${getDayString(ofDate)} ${timeString}`
  }
  return timeString
}

export const getDateRepresentation = (date: Readonly<Date> | string) => {
  const displayDate = new Date(date)
  if (Number.isNaN(displayDate.getTime())) {
    return ''
  }
  const today = new Date()
  const timeString = getTimeString(displayDate)
  const yesterday = new Date(today.getTime() - 1000 * 60 * 60 * 24)

  if (
    displayDate.getFullYear() === today.getFullYear() &&
    displayDate.getMonth() === today.getMonth() &&
    displayDate.getDate() === today.getDate()
  ) {
    return `今日 ${timeString}`
  }
  if (
    displayDate.getFullYear() === yesterday.getFullYear() &&
    displayDate.getMonth() === yesterday.getMonth() &&
    displayDate.getDate() === yesterday.getDate()
  ) {
    return `昨日 ${timeString}`
  }
  if (displayDate.getFullYear() === today.getFullYear()) {
    return `${getDayString(displayDate)} ${timeString}`
  } else {
    return `${getFullDayString(displayDate)} ${timeString}`
  }
}

export const compareDate = (date1: Date, date2: Date, inverse = false) => {
  const _inv = inverse ? -1 : 1
  const _t1 = date1.getTime()
  const _t2 = date2.getTime()
  return _t1 < _t2 ? -_inv : _t1 > _t2 ? _inv : 0
}

export const compareDateString = (
  str1: string,
  str2: string,
  inverse = false
) => compareDate(new Date(str1), new Date(str2), inverse)
