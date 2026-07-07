import { ChannelViewState } from '@traptitech/traq'

import { computed, watchEffect } from 'vue'
import { useRoute } from 'vue-router'

import { useEventListener } from '@vueuse/core'

import { changeViewState, setTimelineStreamingState } from '/@/lib/websocket'
import { RouteName } from '/@/router'
import { useBrowserSettings } from '/@/store/app/browserSettings'
import { useViewStateSenderStore } from '/@/store/domain/viewStateSenderStore'
import { useMainViewStore } from '/@/store/ui/mainView'

const useViewStateSender = () => {
  const route = useRoute()
  const { primaryView } = useMainViewStore()
  const { shouldReceiveLatestMessages, isTyping } = useViewStateSenderStore()
  const { activityMode, stealthMode } = useBrowserSettings()

  const currentChannelId = computed(() => {
    // ルートがチャンネルでないときは閲覧チャンネルをnullにするため
    if (
      route.name !== RouteName.Channel &&
      route.name !== RouteName.User &&
      route.name !== RouteName.File
    ) {
      return undefined
    }
    if (
      primaryView.value.type === 'channel' ||
      primaryView.value.type === 'dm'
    ) {
      return primaryView.value.channelId
    }
    return undefined
  })

  const state = computed(() => {
    if (!shouldReceiveLatestMessages.value) return ChannelViewState.StaleViewing
    // 最新メッセージ閲覧中でない場合はタイピング中でもEditingにしてはいけない
    // (Editingにすると未読に追加されなくなるため)
    return isTyping.value
      ? ChannelViewState.Editing
      : ChannelViewState.Monitoring
  })

  const sendCurrentViewState = () => {
    if (!currentChannelId.value || stealthMode.value) {
      changeViewState(null)
      return
    }

    if (
      document.visibilityState !== 'visible' ||
      !document.hasFocus()
    ) {
      changeViewState(currentChannelId.value, ChannelViewState.None)
      return
    }

    changeViewState(currentChannelId.value, state.value)
  }

  watchEffect(() => {
    sendCurrentViewState()
  })

  watchEffect(() => {
    setTimelineStreamingState(stealthMode.value || activityMode.value.all)
  })

  const visibilitychangeListener = () => {
    sendCurrentViewState()
  }
  const focusListener = () => {
    sendCurrentViewState()
  }

  const blurListener = () => {
    sendCurrentViewState()
  }

  useEventListener(document, 'visibilitychange', visibilitychangeListener)
  useEventListener('focus', focusListener)
  useEventListener('blur', blurListener)
}

export default useViewStateSender
