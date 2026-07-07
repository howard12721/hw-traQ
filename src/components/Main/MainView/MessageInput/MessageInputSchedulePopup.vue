<template>
  <ClickOutside @click-outside="emit('close')">
    <form :class="$style.container" @submit.prevent="scheduleMessage">
      <div :class="$style.header">
        <AIcon mdi name="clock-outline" :size="20" />
        <span :class="$style.title">予約投稿</span>
      </div>
      <FormInput
        v-model="scheduledAtInput"
        type="datetime-local"
        label="日時"
        :min="minScheduledAtInput"
        step="60"
        focus-on-mount
      />
      <p v-if="errorMessage" :class="$style.error">{{ errorMessage }}</p>
      <div :class="$style.actions">
        <button type="button" :class="$style.cancelButton" @click="emit('close')">
          キャンセル
        </button>
        <button type="submit" :class="$style.scheduleButton" :disabled="scheduling">
          <AIcon mdi name="clock-outline" :size="18" />
          予約
        </button>
      </div>
    </form>
  </ClickOutside>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from 'vue'

import ClickOutside from '/@/components/UI/ClickOutside'
import AIcon from '/@/components/UI/AIcon.vue'
import FormInput from '/@/components/UI/FormInput.vue'
import { formatResizeError } from '/@/lib/apis'
import {
  getJstDateTimeLocalString,
  parseJstDateTimeLocalString
} from '/@/lib/basic/date'
import type { MessageInputState } from '/@/store/ui/messageInputStateStore'
import { useScheduledMessageStore } from '/@/store/ui/scheduledMessageStore'
import { useToastStore } from '/@/store/ui/toast'
import type { ChannelId, DMChannelId } from '/@/types/entity-ids'

import { usePostMessageSender } from './composables/usePostMessage'

const DEFAULT_DELAY_MS = 10 * 60 * 1000

const props = defineProps<{
  channelId: ChannelId | DMChannelId
  messageState: MessageInputState
  canSchedule: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'scheduled'): void
}>()

const { addScheduledMessage } = useScheduledMessageStore()
const { addInfoToast, addErrorToast } = useToastStore()
const { prepareMessageInputContent, confirmForcePostIfNeeded } =
  usePostMessageSender()

const createDefaultScheduledAtInput = () =>
  getJstDateTimeLocalString(new Date(Date.now() + DEFAULT_DELAY_MS))

const scheduledAtInput = ref(createDefaultScheduledAtInput())
const minScheduledAtInput = computed(() => getJstDateTimeLocalString(new Date()))
const errorMessage = ref('')
const scheduling = ref(false)

watch(
  () => props.channelId,
  () => {
    scheduledAtInput.value = createDefaultScheduledAtInput()
    errorMessage.value = ''
  }
)

const scheduleMessage = async () => {
  errorMessage.value = ''

  if (!props.canSchedule) {
    errorMessage.value = 'メッセージを入力してください'
    return
  }

  const scheduledAt = parseJstDateTimeLocalString(scheduledAtInput.value)
  if (!scheduledAt) {
    errorMessage.value = '日時を正しく入力してください'
    return
  }
  if (scheduledAt.getTime() <= Date.now()) {
    errorMessage.value = '未来の日時を指定してください'
    return
  }

  if (!confirmForcePostIfNeeded(props.channelId)) {
    return
  }

  try {
    scheduling.value = true
    const content = await prepareMessageInputContent(
      props.channelId,
      props.messageState
    )
    if (content === undefined) return

    await addScheduledMessage({
      channelId: props.channelId,
      content,
      scheduledAt
    })
    addInfoToast('予約しました')
    emit('scheduled')
  } catch (e) {
    addErrorToast(formatResizeError(e, '予約投稿に失敗しました'))
    // eslint-disable-next-line no-console
    console.error(e)
  } finally {
    scheduling.value = false
  }
}
</script>

<style lang="scss" module>
.container {
  @include background-secondary;
  @include color-text-primary;
  position: absolute;
  right: 0;
  bottom: 36px;
  width: min(320px, calc(100vw - 32px));
  padding: 16px;
  border: 1px solid $theme-ui-secondary-default;
  border-radius: 8px;
  box-shadow: 0 12px 32px rgb(0 0 0 / 24%);
  z-index: $z-index-stamp-picker;
}

.header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.title {
  @include size-body1;
  font-weight: bold;
}

.error {
  @include size-body2;
  color: $theme-accent-error-default;
  margin-top: 8px;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 16px;
}

.cancelButton,
.scheduleButton {
  height: 34px;
  padding: 0 14px;
  border-radius: 4px;
  font-weight: bold;
  cursor: pointer;
}

.cancelButton {
  @include color-ui-secondary;
  border: 1px solid $theme-ui-secondary-default;
}

.scheduleButton {
  @include color-common-text-white-primary;
  @include background-accent-primary;
  display: flex;
  align-items: center;
  gap: 6px;

  &:disabled {
    @include background-secondary;
    @include color-ui-secondary-inactive;
    cursor: wait;
  }
}

</style>
