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
        <router-link
          v-for="notification in notifications"
          :key="notification.id"
          :to="constructMessagesPath(notification.messageId)"
          :class="$style.item"
          :data-unread="$boolAttr(!notification.read)"
        >
          <UserIcon
            :class="$style.icon"
            :user-id="notification.authorId"
            :size="28"
          />
          <div :class="$style.body">
            <div :class="$style.meta">
              <span :class="$style.pattern">{{ notification.pattern }}</span>
              <time :class="$style.time" :datetime="notification.notifiedAt">
                {{ formatDate(notification.notifiedAt) }}
              </time>
            </div>
            <p :class="$style.content">{{ notification.content }}</p>
          </div>
        </router-link>
      </template>
      <EmptyState v-else> Gazer通知はありません </EmptyState>
    </transition-group>
  </div>
</template>

<script lang="ts" setup>
import { onMounted, watch } from 'vue'

import AIcon from '/@/components/UI/AIcon.vue'
import EmptyState from '/@/components/UI/EmptyState.vue'
import UserIcon from '/@/components/UI/UserIcon.vue'
import { getDateRepresentation } from '/@/lib/basic/date'
import { constructMessagesPath } from '/@/router'
import { useGazerNotificationsStore } from '/@/store/domain/gazerNotifications'
import { useSubscriptionStore } from '/@/store/domain/subscription'
import { useChannelsStore } from '/@/store/entities/channels'

const { notifications, loading, gatewayUserId, unreadCount } =
  useGazerNotificationsStore()
const { fetchNotifications, markRead } = useGazerNotificationsStore()
const { userIdToDmChannelIdMap } = useChannelsStore()
const { deleteUnreadChannelWithSend } = useSubscriptionStore()

const formatDate = (value: string) => getDateRepresentation(value)

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
  await fetchNotifications()
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
  @include background-secondary;
  display: grid;
  grid-template-columns: 28px minmax(0, 1fr);
  gap: 10px;
  padding: 10px;
  margin: 10px 0;
  border: 1px solid transparent;
  border-radius: 6px;

  &[data-unread] {
    border-color: $theme-accent-primary-default;
  }
}

.icon {
  margin-top: 2px;
}

.body {
  min-width: 0;
}

.meta {
  display: flex;
  align-items: baseline;
  gap: 8px;
  min-width: 0;
}

.pattern {
  @include color-accent-primary;
  @include size-caption;
  flex: 1 1 auto;
  min-width: 0;
  font-weight: bold;
  overflow-wrap: anywhere;
}

.time {
  @include color-ui-secondary;
  @include size-caption;
  flex: 0 0 auto;
}

.content {
  @include color-ui-primary;
  display: -webkit-box;
  margin-top: 4px;
  overflow: hidden;
  overflow-wrap: anywhere;
  white-space: pre-wrap;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
}
</style>
