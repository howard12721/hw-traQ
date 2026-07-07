import { computed, ref } from 'vue'

import { acceptHMRUpdate, defineStore } from 'pinia'

import {
  deleteScheduledMessage,
  getScheduledMessages,
  postScheduledMessage,
  type ScheduledMessageItem
} from '/@/lib/internalApi'
import { convertToRefsStore } from '/@/store/utils/convertToRefsStore'
import type { ChannelId, DMChannelId } from '/@/types/entity-ids'

export type ScheduledMessage = ScheduledMessageItem & {
  channelId: ChannelId | DMChannelId
}

const sortScheduledMessages = (messages: ScheduledMessage[]) =>
  [...messages].sort((a, b) => {
    const aRunnableAt = new Date(a.retryAt ?? a.scheduledAt).getTime()
    const bRunnableAt = new Date(b.retryAt ?? b.scheduledAt).getTime()
    return aRunnableAt - bRunnableAt
  })

const useScheduledMessageStorePinia = defineStore(
  'ui/scheduledMessageStore',
  () => {
    const messages = ref<ScheduledMessage[]>([])
    const loading = ref(false)

    const scheduledMessages = computed(() =>
      sortScheduledMessages(messages.value)
    )

    const fetchScheduledMessages = async () => {
      loading.value = true
      try {
        const res = await getScheduledMessages()
        messages.value = res.messages as ScheduledMessage[]
      } finally {
        loading.value = false
      }
    }

    const addScheduledMessage = async ({
      channelId,
      content,
      scheduledAt
    }: {
      channelId: ChannelId | DMChannelId
      content: string
      scheduledAt: Date
    }) => {
      const res = await postScheduledMessage({
        channelId,
        content,
        scheduledAt: scheduledAt.toISOString()
      })
      const message = res.message as ScheduledMessage
      messages.value = sortScheduledMessages([...messages.value, message])
      return message
    }

    const removeScheduledMessage = async (id: string) => {
      await deleteScheduledMessage(id)
      messages.value = messages.value.filter(message => message.id !== id)
    }

    return {
      scheduledMessages,
      loading,
      fetchScheduledMessages,
      addScheduledMessage,
      removeScheduledMessage
    }
  }
)

export const useScheduledMessageStore = convertToRefsStore(
  useScheduledMessageStorePinia
)

if (import.meta.hot) {
  import.meta.hot.accept(
    acceptHMRUpdate(useScheduledMessageStorePinia, import.meta.hot)
  )
}
