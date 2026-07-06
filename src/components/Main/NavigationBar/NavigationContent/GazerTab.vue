<template>
  <div :class="$style.container">
    <div :class="$style.toolbar">
      <button
        :class="$style.refreshButton"
        :disabled="loading"
        @click="refresh"
      >
        <AIcon name="history" mdi :size="18" />
      </button>
    </div>

    <transition-group name="timeline" tag="div">
      <template v-if="notifications.length > 0">
        <MessagePanel
          v-for="notification in notifications"
          :key="notification.id"
          :class="$style.item"
          :data-unread="$boolAttr(!notification.read)"
          title-type="user"
          line-clamp-content
          :message="toActivityMessage(notification)"
          :to="constructMessagesPath(notification.messageId)"
        >
          <template #footer>
            <div :class="$style.gazerSeparator" />
            <div :class="$style.gazerName">{{ notification.displayName }}</div>
          </template>
        </MessagePanel>
      </template>
      <EmptyState v-else> Gazer通知はありません </EmptyState>
    </transition-group>
  </div>
</template>

<script lang="ts" setup>
import type { ActivityTimelineMessage } from '@traptitech/traq'

import { onMounted, watch } from 'vue'

import AIcon from '/@/components/UI/AIcon.vue'
import EmptyState from '/@/components/UI/EmptyState.vue'
import MessagePanel from '/@/components/UI/MessagePanel/MessagePanel.vue'
import type { GazerNotificationItem } from '/@/lib/internalApi'
import { constructMessagesPath } from '/@/router'
import { useGazerNotificationsStore } from '/@/store/domain/gazerNotifications'
import { useSubscriptionStore } from '/@/store/domain/subscription'
import { useChannelsStore } from '/@/store/entities/channels'

const { notifications, loading, gatewayUserId, unreadCount } =
  useGazerNotificationsStore()
const { fetchNotifications, markRead } = useGazerNotificationsStore()
const { userIdToDmChannelIdMap } = useChannelsStore()
const { fetchChannels } = useChannelsStore()
const { deleteUnreadChannelWithSend } = useSubscriptionStore()

const toActivityMessage = (
  notification: GazerNotificationItem
): ActivityTimelineMessage => ({
  id: notification.messageId,
  userId: notification.authorId,
  channelId: notification.channelId,
  content: notification.content,
  createdAt: notification.createdAt,
  updatedAt: notification.createdAt
})

const markGatewayRead = async () => {
  await markRead()
  const dmChannelId = gatewayUserId.value
    ? userIdToDmChannelIdMap.value.get(gatewayUserId.value)
    : undefined
  if (dmChannelId) {
    await deleteUnreadChannelWithSend(dmChannelId)
  }
}

const refresh = async () => {
  await Promise.all([fetchNotifications(), fetchChannels()])
  await markGatewayRead()
}

onMounted(refresh)

watch(unreadCount, count => {
  if (count > 0) {
    void markGatewayRead()
  }
})
</script>

<style lang="scss" module>
.container {
  padding: 0 16px 0 0;
}

.toolbar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 12px;
}

.refreshButton {
  @include color-ui-secondary;
  display: grid;
  place-items: center;
  width: 32px;
  height: 32px;
  border-radius: 4px;
  cursor: pointer;

  &:hover,
  &:focus-visible {
    @include background-secondary;
    @include color-ui-primary;
  }

  &:disabled {
    cursor: wait;
    opacity: 0.5;
  }
}

.item {
  display: block;
  margin: 16px 0;
  border-radius: 6px;

  &[data-unread] {
    box-shadow:
      inset 4px 0 0 $theme-accent-primary-default,
      0 0 0 1px $theme-accent-primary-default;
  }
}

.gazerSeparator {
  @include background-secondary;
  width: 100%;
  height: 2px;
  margin: 8px 0 4px;
}

.gazerName {
  @include color-ui-secondary;
  @include size-caption;
  overflow-wrap: anywhere;
}
</style>
