<template>
  <div :class="$style.container" :data-is-mobile="$boolAttr(isMobile)">
    <MessageInputInsertStampButton @click="emit('clickStamp')" />
    <button
      :class="$style.scheduleButton"
      title="予約投稿"
      :data-has-scheduled-messages="$boolAttr(hasScheduledMessages)"
      data-testid="message-schedule-button"
      @click="emit('clickSchedule')"
    >
      <AIcon mdi name="clock-outline" />
    </button>
    <button
      :class="$style.sendButton"
      title="送信する"
      :disabled="!canPostMessage"
      data-testid="message-send-button"
      @click="onClickSendButton"
    >
      <transition name="post">
        <AIcon v-if="!isPosting" mdi name="send" />
      </transition>
    </button>
  </div>
</template>

<script lang="ts" setup>
import AIcon from '/@/components/UI/AIcon.vue'
import useResponsive from '/@/composables/useResponsive'

import MessageInputInsertStampButton from './MessageInputInsertStampButton.vue'

const props = withDefaults(
  defineProps<{
    canPostMessage?: boolean
    isPosting?: boolean
    hasScheduledMessages?: boolean
  }>(),
  {
    canPostMessage: false,
    isPosting: false,
    hasScheduledMessages: false
  }
)

const emit = defineEmits<{
  (e: 'clickSend'): void
  (e: 'clickStamp'): void
  (e: 'clickSchedule'): void
}>()

const { isMobile } = useResponsive()

const onClickSendButton = () => {
  if (props.canPostMessage) {
    emit('clickSend')
  }
}
</script>

<style lang="scss" module>
.container {
  @include color-ui-secondary;
  display: flex;
}
.scheduleButton,
.sendButton {
  height: 24px;
  width: 24px;
  cursor: pointer;

  margin: 0 8px;
  .container[data-is-mobile] & {
    margin: 0 8px;
  }

  &:first-child:first-child {
    margin-left: 0;
  }
  &:last-child:last-child {
    margin-right: 0;
  }

  transform: scale(1);
  transition: transform 0.1s;
  &:hover {
    transform: scale(1.1);
  }
}

.scheduleButton {
  @include color-ui-secondary;

  &[data-has-scheduled-messages] {
    @include color-accent-primary;
  }
}

.sendButton {
  @include color-accent-primary;

  &[disabled] {
    @include color-ui-secondary-inactive;
    cursor: not-allowed;
  }
}
</style>
