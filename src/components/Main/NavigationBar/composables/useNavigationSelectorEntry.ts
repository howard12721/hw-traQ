import { computed, reactive } from 'vue'

import { useQall } from '/@/composables/qall/useQall'
import { isDefined } from '/@/lib/basic/array'
import type { ThemeClaim } from '/@/lib/styles'
import { useGazerNotificationsStore } from '/@/store/domain/gazerNotifications'
import { useSubscriptionStore } from '/@/store/domain/subscription'
import { useChannelsStore } from '/@/store/entities/channels'
import { useAudioController } from '/@/store/ui/audioController'
import { useMainViewStore } from '/@/store/ui/mainView'
import { useMessageInputStateStore } from '/@/store/ui/messageInputStateStore'

import type {
  EphemeralNavigationItemType,
  NavigationItemType
} from './useNavigationConstructor'

export type NavigationSelectorEntry = {
  type: NavigationItemType
  iconName: string
  iconMdi?: boolean
  hasNotification?: boolean
}

export type EphemeralNavigationSelectorEntry = {
  type: EphemeralNavigationItemType
  iconName: string
  iconMdi?: boolean
  colorClaim?: ThemeClaim<string> // 色
  selectOnAdd?: boolean
}

export const createItems = (notificationState: {
  channel: boolean
  dm: boolean
  gazer: boolean
}): NavigationSelectorEntry[] => [
  {
    type: 'home',
    iconName: 'home',
    iconMdi: true,
    hasNotification: notificationState.channel
  },
  {
    type: 'channels',
    iconName: 'hash'
  },
  {
    type: 'activity',
    iconName: 'activity'
  },
  {
    type: 'users',
    iconName: 'user',
    hasNotification: notificationState.dm
  },
  {
    type: 'gazer',
    iconName: 'eye-outline',
    iconMdi: true,
    hasNotification: notificationState.gazer
  },
  {
    type: 'clips',
    iconName: 'bookmark',
    iconMdi: true
  }
]
export const ephemeralItems: Record<
  NonNullable<EphemeralNavigationItemType>,
  EphemeralNavigationSelectorEntry
> = {
  qallController: {
    type: 'qallController',
    iconName: 'phone',
    iconMdi: true,
    colorClaim: (_, common) => common.ui.qall,
    selectOnAdd: true
  },
  draftList: {
    type: 'draftList',
    iconName: 'pencil',
    iconMdi: true
  },
  audioController: {
    type: 'audioController',
    iconName: 'music-note',
    iconMdi: true
  }
}

const useNavigationSelectorEntry = () => {
  const { unreadChannelsMap } = useSubscriptionStore()
  const { channelsMap, dmChannelsMap } = useChannelsStore()
  const { isGatewayUserId, unreadCount } = useGazerNotificationsStore()
  const { hasInputChannel } = useMessageInputStateStore()
  const { fileId } = useAudioController()
  const { getQallingState } = useQall()
  const { primaryView } = useMainViewStore()

  const unreadChannels = computed(() => [...unreadChannelsMap.value.values()])
  const notificationState = reactive({
    channel: computed(() =>
      unreadChannels.value.some(c => channelsMap.value.has(c.channelId))
    ),
    dm: computed(() =>
      unreadChannels.value.some(c => {
        const dmChannel = dmChannelsMap.value.get(c.channelId)
        return !!dmChannel && !isGatewayUserId(dmChannel.userId)
      })
    ),
    gazer: computed(
      () =>
        unreadCount.value > 0 ||
        unreadChannels.value.some(c => {
          const dmChannel = dmChannelsMap.value.get(c.channelId)
          return !!dmChannel && isGatewayUserId(dmChannel.userId)
        })
    )
  })
  const entries = computed(() => createItems(notificationState))

  const hasActiveQallSession = computed(
    () =>
      primaryView.value.type === 'channel' &&
      getQallingState(primaryView.value.channelId) === 'subView'
  )
  const ephemeralEntries = computed(() =>
    [
      hasActiveQallSession.value ? ephemeralItems.qallController : undefined,
      hasInputChannel.value ? ephemeralItems.draftList : undefined,
      fileId.value ? ephemeralItems.audioController : undefined
    ].filter(isDefined)
  )

  return {
    entries,
    ephemeralEntries
  }
}
export default useNavigationSelectorEntry
