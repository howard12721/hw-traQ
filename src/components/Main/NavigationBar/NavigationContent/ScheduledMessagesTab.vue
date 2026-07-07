<template>
  <div :class="$style.container">
    <div :class="$style.toolbar">
      <button
        title="再読み込み"
        :class="$style.refreshButton"
        :disabled="loading"
        @click="refresh"
      >
        <AIcon name="history" mdi :size="18" />
      </button>
    </div>

    <transition-group name="timeline" tag="div">
      <template v-if="scheduledMessages.length > 0">
        <div
          v-for="message in scheduledMessages"
          :key="message.id"
          :class="$style.item"
        >
          <router-link :to="messageLink(message)" :class="$style.link">
            <div :class="$style.channel">
              {{ channelLabel(message.channelId) }}
            </div>
            <div :class="$style.separator" />
            <RenderContent
              :content="message.content"
              line-clamp-content
              :class="$style.content"
            />
            <div :class="$style.footer">
              <AIcon name="clock-outline" mdi :size="16" />
              <span>{{ formatScheduledAt(message.scheduledAt) }}</span>
              <span
                v-if="message.failedAttempts"
                :class="$style.retrying"
              >
                再試行待ち
              </span>
            </div>
          </router-link>
          <button
            title="予約をキャンセル"
            :class="$style.cancelButton"
            :disabled="cancelingIds.has(message.id)"
            @click="cancelScheduledMessage(message.id)"
          >
            <AIcon name="delete" mdi :size="18" />
          </button>
        </div>
      </template>
      <EmptyState v-else>予約投稿はありません</EmptyState>
    </transition-group>
  </div>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue'

import AIcon from '/@/components/UI/AIcon.vue'
import EmptyState from '/@/components/UI/EmptyState.vue'
import RenderContent from '/@/components/UI/MessagePanel/RenderContent.vue'
import useChannelPath from '/@/composables/useChannelPath'
import { getJstFullDayWithTimeString } from '/@/lib/basic/date'
import { useChannelsStore } from '/@/store/entities/channels'
import {
  type ScheduledMessage,
  useScheduledMessageStore
} from '/@/store/ui/scheduledMessageStore'
import { useToastStore } from '/@/store/ui/toast'
import type { ChannelId, DMChannelId } from '/@/types/entity-ids'

const {
  scheduledMessages,
  loading,
  fetchScheduledMessages,
  removeScheduledMessage
} = useScheduledMessageStore()
const { fetchChannels } = useChannelsStore()
const { channelIdToLink, channelIdToPathString } = useChannelPath()
const { addInfoToast, addErrorToast } = useToastStore()

const cancelingIds = ref(new Set<string>())

const refresh = async () => {
  await Promise.all([fetchScheduledMessages(), fetchChannels()])
}

const channelLabel = (channelId: ChannelId | DMChannelId) =>
  channelIdToPathString(channelId, true) ?? '# unknown'

const messageLink = (message: ScheduledMessage) =>
  channelIdToLink(message.channelId) ?? ''

const formatScheduledAt = (scheduledAt: string) =>
  getJstFullDayWithTimeString(new Date(scheduledAt))

const setCanceling = (id: string, canceling: boolean) => {
  const next = new Set(cancelingIds.value)
  if (canceling) {
    next.add(id)
  } else {
    next.delete(id)
  }
  cancelingIds.value = next
}

const cancelScheduledMessage = async (id: string) => {
  if (!confirm('この予約投稿をキャンセルしますか？')) return

  try {
    setCanceling(id, true)
    await removeScheduledMessage(id)
    addInfoToast('予約投稿をキャンセルしました')
  } catch (e) {
    addErrorToast('予約投稿をキャンセルできませんでした')
    // eslint-disable-next-line no-console
    console.error(e)
  } finally {
    setCanceling(id, false)
  }
}

onMounted(() => {
  void refresh().catch(e => {
    addErrorToast('予約投稿一覧を取得できませんでした')
    // eslint-disable-next-line no-console
    console.error(e)
  })
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
  @include background-primary;
  position: relative;
  display: flex;
  align-items: stretch;
  margin: 16px 0;
  border-radius: 4px;
}

.link {
  min-width: 0;
  flex: 1;
  padding: 8px 48px 8px 16px;
}

.channel {
  @include color-ui-primary;
  @include size-body1;
  margin: 4px 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: bold;
}

.separator {
  @include background-secondary;
  width: 100%;
  height: 2px;
  margin: 4px 0;
}

.content {
  margin: 4px 0;
}

.footer {
  @include color-ui-secondary;
  @include size-body2;
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 8px;
}

.retrying {
  color: $theme-accent-error-default;
  margin-left: auto;
}

.cancelButton {
  @include color-ui-secondary;
  position: absolute;
  top: 8px;
  right: 8px;
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
</style>
