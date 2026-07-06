import { computed, ref } from 'vue'

import { acceptHMRUpdate, defineStore } from 'pinia'

import {
  getGazerNotifications,
  markGazerNotificationsRead
} from '/@/lib/internalApi'
import type {
  GazerNotificationItem,
  GazerNotificationsResponse,
  GazerResponse
} from '/@/lib/internalApi'
import { messageMitt } from '/@/store/entities/messages'
import { useToastStore } from '/@/store/ui/toast'
import { convertToRefsStore } from '/@/store/utils/convertToRefsStore'
import type { UserId } from '/@/types/entity-ids'

const useGazerNotificationsStorePinia = defineStore(
  'domain/gazerNotifications',
  () => {
    const notifications = ref<GazerNotificationItem[]>([])
    const gatewayUserId = ref<UserId>()
    const loading = ref(false)

    const unreadCount = computed(
      () =>
        notifications.value.filter(notification => !notification.read).length
    )

    const applyNotificationResponse = (
      response: GazerNotificationsResponse
    ) => {
      notifications.value = response.notifications
      gatewayUserId.value = response.botUserId
    }

    const applyGazerResponse = (response: GazerResponse) => {
      gatewayUserId.value = response.status.botUserId
    }

    const fetchNotifications = async () => {
      loading.value = true
      try {
        applyNotificationResponse(await getGazerNotifications())
      } finally {
        loading.value = false
      }
    }

    const markRead = async () => {
      if (unreadCount.value === 0) return

      await markGazerNotificationsRead()
      notifications.value = notifications.value.map(notification => ({
        ...notification,
        read: true
      }))
    }

    const isGatewayUserId = (userId: UserId) =>
      gatewayUserId.value !== undefined && gatewayUserId.value === userId

    const fetchNotificationsForGatewayMessage = async () => {
      const knownIds = new Set(
        notifications.value.map(notification => notification.id)
      )
      try {
        await fetchNotifications()
      } catch (e) {
        // eslint-disable-next-line no-console
        console.warn('Failed to fetch gazer notifications', { cause: e })
        return
      }

      const latest = notifications.value.find(
        notification => !knownIds.has(notification.id)
      )
      if (!latest) return

      useToastStore().addToast({
        type: 'info',
        text: `Gazer: ${latest.displayName}`,
        timeout: 8000
      })
    }

    messageMitt.on('addMessage', ({ message }) => {
      if (!isGatewayUserId(message.userId)) return
      void fetchNotificationsForGatewayMessage()
    })

    return {
      notifications,
      unreadCount,
      gatewayUserId,
      loading,
      applyGazerResponse,
      fetchNotifications,
      markRead,
      isGatewayUserId
    }
  }
)

export const useGazerNotificationsStore = convertToRefsStore(
  useGazerNotificationsStorePinia
)

if (import.meta.hot) {
  import.meta.hot.accept(
    acceptHMRUpdate(useGazerNotificationsStorePinia, import.meta.hot)
  )
}
