import { ref } from 'vue'

import { acceptHMRUpdate, defineStore } from 'pinia'

import { convertToRefsStore } from '/@/store/utils/convertToRefsStore'

export const stealthModeState = ref(false)

/**
 * /@/views/composables/useViewStateSenderで利用する情報を格納することを想定
 */
const useViewStateSenderStorePinia = defineStore(
  'domain/viewStateSenderStore',
  () => {
    /**
     * 最新のメッセージを受け取る必要があるかどうか
     *
     * messageFetcherの`isReachedLatest`と同期する必要がある
     */
    const shouldReceiveLatestMessages = ref(false)
    /** タイピング中かどうか */
    const isTyping = ref(false)
    /** ステルスモード中かどうか */
    const isStealthMode = stealthModeState

    return { shouldReceiveLatestMessages, isTyping, isStealthMode }
  }
)

export const useViewStateSenderStore = convertToRefsStore(
  useViewStateSenderStorePinia
)

if (import.meta.hot) {
  import.meta.hot.accept(
    acceptHMRUpdate(useViewStateSenderStorePinia, import.meta.hot)
  )
}
