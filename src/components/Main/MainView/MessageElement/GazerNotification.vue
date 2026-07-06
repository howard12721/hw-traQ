<template>
  <div :class="$style.container">
    <div :class="$style.header">
      <UserIcon :user-id="notification.authorId" :size="28" />
      <div :class="$style.title">
        <span :class="$style.label">Gazer</span>
        <span :class="$style.pattern">{{ notification.pattern }}</span>
      </div>
      <router-link :class="$style.link" :to="messagePath">開く</router-link>
    </div>
    <p :class="$style.content">{{ notification.content }}</p>
  </div>
</template>

<script lang="ts" setup>
import { computed } from 'vue'

import UserIcon from '/@/components/UI/UserIcon.vue'
import type { GazerNotification } from '/@/lib/gazerNotification'
import { constructMessagesPath } from '/@/router'

const props = defineProps<{
  notification: GazerNotification
}>()

const messagePath = computed(() =>
  constructMessagesPath(props.notification.messageId)
)
</script>

<style lang="scss" module>
.container {
  @include background-secondary;
  border-left: 4px solid $theme-accent-primary-default;
  border-radius: 4px;
  padding: 12px;
}

.header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.title {
  display: flex;
  align-items: baseline;
  gap: 8px;
  min-width: 0;
}

.label {
  @include color-accent-primary;
  font-weight: bold;
}

.pattern {
  @include color-ui-secondary;
  @include size-caption;
  overflow-wrap: anywhere;
}

.link {
  @include color-accent-primary;
  margin-left: auto;
  font-weight: bold;
}

.content {
  @include color-ui-primary;
  margin-top: 8px;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}
</style>
