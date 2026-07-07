import type { ChannelViewer } from '@traptitech/traq'
import { ChannelViewState } from '@traptitech/traq'

import type { Ref } from 'vue'
import { computed, ref, watch } from 'vue'

import useMittListener from '/@/composables/utils/useMittListener'
import { wsListener } from '/@/lib/websocket'
import { useBrowserSettings } from '/@/store/app/browserSettings'
import { useMeStore } from '/@/store/domain/me'
import type { ChannelId } from '/@/types/entity-ids'

const useCurrentViewers = (channelId: Ref<ChannelId>) => {
  const meStore = useMeStore()
  const { stealthMode } = useBrowserSettings()

  /** チャンネルを見ている人の一覧(古い順) */
  const currentViewers = ref<ChannelViewer[]>([])

  /**
   * チャンネルを見ている人(入力中を含む、バックグラウンド表示を除く)のIDの一覧(古い順)
   */
  const activeViewingUsers = computed(() => {
    if (stealthMode.value) return []
    return currentViewers.value
      .filter(
        v =>
          v.state === ChannelViewState.Monitoring ||
          v.state === ChannelViewState.Editing ||
          v.state === ChannelViewState.StaleViewing
      )
      .map(v => v.userId)
  })

  const inactiveViewingUsers = computed(() => {
    if (stealthMode.value) return []
    return currentViewers.value
      .filter(v => v.state === ChannelViewState.None)
      .map(v => v.userId)
  })

  /**
   * チャンネルで入力中の人のIDの一覧(新しい順)
   */
  const typingUsers = computed(() => {
    if (stealthMode.value) return []
    const myId = meStore.myId.value
    return currentViewers.value
      .filter(v => v.state === ChannelViewState.Editing && v.userId !== myId)
      .map(v => v.userId)
      .reverse()
  })

  useMittListener(wsListener, 'CHANNEL_VIEWERS_CHANGED', ({ id, viewers }) => {
    if (stealthMode.value) return
    if (channelId.value === id) {
      currentViewers.value = viewers
    }
  })
  watch(stealthMode, isStealthMode => {
    if (isStealthMode) {
      currentViewers.value = []
    }
  })
  // NOTE: 再接続時にはCHANNEL_VIEWERS_CHANGEDが送られてくる

  return { activeViewingUsers, typingUsers, inactiveViewingUsers }
}

export default useCurrentViewers
