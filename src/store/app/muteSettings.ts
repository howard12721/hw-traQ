import type { Message } from '@traptitech/traq'

import { computed, reactive, toRefs } from 'vue'

import { acceptHMRUpdate, defineStore } from 'pinia'

import useLocalStorageValue from '/@/composables/storage/useLocalStorage'
import { convertToRefsStore } from '/@/store/utils/convertToRefsStore'
import type { UserId } from '/@/types/entity-ids'

type State = {
  mutedUserIds: UserId[]
  mutedKeywords: string[]
}

const normalizeKeyword = (keyword: string) => keyword.trim()

const dedupe = <T>(values: readonly T[]) => [...new Set(values)]
const dedupeKeywords = (keywords: readonly string[]) => {
  const lowerKeywordSet = new Set<string>()
  const result: string[] = []

  for (const keyword of keywords.map(normalizeKeyword).filter(Boolean)) {
    const lowerKeyword = keyword.toLowerCase()
    if (lowerKeywordSet.has(lowerKeyword)) continue

    lowerKeywordSet.add(lowerKeyword)
    result.push(keyword)
  }
  return result
}

const useMuteSettingsPinia = defineStore('app/muteSettings', () => {
  const initialValue: State = {
    mutedUserIds: [],
    mutedKeywords: []
  }

  const [state] = useLocalStorageValue(
    'store/app/muteSettings',
    1,
    {},
    initialValue
  )

  const normalizedMutedKeywords = computed(() =>
    dedupeKeywords(state.mutedKeywords)
  )
  const mutedUserIdSet = computed(() => new Set(state.mutedUserIds))

  const addMutedUserId = (userId: UserId) => {
    state.mutedUserIds = dedupe([...state.mutedUserIds, userId])
  }

  const removeMutedUserId = (userId: UserId) => {
    state.mutedUserIds = state.mutedUserIds.filter(id => id !== userId)
  }

  const addMutedKeyword = (keyword: string) => {
    const normalizedKeyword = normalizeKeyword(keyword)
    if (!normalizedKeyword) return
    state.mutedKeywords = dedupeKeywords([
      ...state.mutedKeywords,
      normalizedKeyword
    ])
  }

  const removeMutedKeyword = (keyword: string) => {
    state.mutedKeywords = state.mutedKeywords.filter(
      item => normalizeKeyword(item).toLowerCase() !== keyword.toLowerCase()
    )
  }

  const isMessageMuted = (
    message: Pick<Message, 'content' | 'userId'>
  ): boolean => {
    if (mutedUserIdSet.value.has(message.userId)) return true

    const lowerContent = message.content.toLowerCase()
    return normalizedMutedKeywords.value.some(keyword =>
      lowerContent.includes(keyword.toLowerCase())
    )
  }

  return {
    ...toRefs(state),
    config: reactive(state),
    normalizedMutedKeywords,
    addMutedUserId,
    removeMutedUserId,
    addMutedKeyword,
    removeMutedKeyword,
    isMessageMuted
  }
})

export const useMuteSettings = convertToRefsStore(useMuteSettingsPinia)

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useMuteSettingsPinia, import.meta.hot))
}
