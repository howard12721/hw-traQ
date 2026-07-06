<template>
  <div
    :class="$style.container"
    :data-is-selected="$boolAttr(isSelected && !isOpened)"
    :data-is-inactive="$boolAttr(!channel.active)"
  >
    <button
      type="button"
      :class="$style.folderContainer"
      :aria-expanded="isOpened"
      :data-is-inactive="$boolAttr(!channel.active)"
      :aria-label="
        isOpened ? `${pathToShow} フォルダを閉じる` : `${pathToShow} フォルダを開く`
      "
      @click="onClick"
      @mouseenter="onMouseEnter"
      @mouseleave="onMouseLeave"
      @focus="onFocus"
      @blur="onBlur"
    >
      <span
        :class="$style.folderIconWrapper"
        :data-is-opened="$boolAttr(isOpened)"
        :data-is-inactive="$boolAttr(!channel.active)"
      >
        <AIcon
          :name="isOpened ? 'folder-open-outline' : 'folder-outline'"
          :class="$style.folderIcon"
          :size="20"
          mdi
        />
        <span v-if="hasNotification" :class="$style.indicator">
          <NotificationIndicator :border-width="2" />
        </span>
      </span>
      <span :class="$style.name" :title="pathTooltip">
        {{ pathToShow }}
      </span>
    </button>

    <div
      v-if="(isSelected && !isOpened) || isHovered || isFocused"
      :class="$style.selectedBg"
      :data-is-hovered="$boolAttr(isHovered)"
      :data-is-focused="$boolAttr(isFocused)"
    />
  </div>
</template>

<script lang="ts" setup>
import { computed, toRef } from 'vue'

import AIcon from '/@/components/UI/AIcon.vue'
import NotificationIndicator from '/@/components/UI/NotificationIndicator.vue'
import useFocus from '/@/composables/dom/useFocus'
import useHover from '/@/composables/dom/useHover'
import type { ChannelTreeNode } from '/@/lib/channelTree'
import { useMainViewStore } from '/@/store/ui/mainView'
import type { ChannelId } from '/@/types/entity-ids'

import useNotificationState from '../composables/useNotificationState'
import type { TypedProps } from './composables/usePath'
import { usePath } from './composables/usePath'

const props = withDefaults(
  defineProps<{
    channel: ChannelTreeNode
    isOpened?: boolean
    showShortenedPath?: boolean
  }>(),
  {
    isOpened: false,
    showShortenedPath: false
  }
)

const emit = defineEmits<{
  (e: 'click', channelId: ChannelId): void
}>()

const { primaryView } = useMainViewStore()

const isSelected = computed(
  () =>
    primaryView.value.type === 'channel' &&
    props.channel.id === primaryView.value.channelId
)

const notificationState = useNotificationState(toRef(props, 'channel'))
const hasNotification = computed(
  () =>
    notificationState.hasNotification ||
    notificationState.hasNotificationOnChild
)

const { pathToShow, pathTooltip } = usePath(props as TypedProps)
const { isHovered, onMouseEnter, onMouseLeave } = useHover()
const { isFocused, onFocus, onBlur } = useFocus()

const onClick = () => {
  emit('click', props.channel.id)
}
</script>

<style lang="scss" module>
$elementHeight: 32px;
$bgHeight: 36px;
$bgLeftShift: 8px;

.container {
  @include color-ui-primary;
  display: block;
  user-select: none;
  position: relative;
  contain: layout;
  &[data-is-inactive] {
    @include color-ui-secondary;
  }
  &[data-is-selected] {
    @include color-accent-primary;
  }
}
.folderContainer {
  @include color-ui-primary;
  position: relative;
  display: flex;
  align-items: center;
  width: calc(100% - #{$bgLeftShift});
  height: $elementHeight;
  padding-left: 24px;
  padding-right: 12px;
  margin-left: $bgLeftShift;
  cursor: pointer;
  text-align: left;
  z-index: 0;
  &[data-is-inactive] {
    @include color-ui-secondary;
  }
}
.folderIconWrapper {
  position: absolute;
  left: 3px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border-radius: 4px;
  color: $theme-ui-primary-default;
  &[data-is-opened] {
    color: var(--specific-channel-hash-opened);
    background: $theme-ui-primary-background;
  }
  &[data-is-inactive] {
    color: $theme-ui-secondary-default;
  }
}
.folderIcon {
  flex-shrink: 0;
}
.name {
  @include size-body1;
  display: block;
  min-width: 0;
  max-width: 100%;
  padding: 0 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.indicator {
  position: absolute;
  top: -2px;
  right: -2px;
}
.selectedBg {
  position: absolute;
  width: calc(100% + #{$bgLeftShift});
  height: $bgHeight;
  top: -1 * math.div($bgHeight - $elementHeight, 2);
  left: 0;
  z-index: 0;
  border-top-left-radius: 100vw;
  border-bottom-left-radius: 100vw;
  opacity: 0.1;
  pointer-events: none;

  display: none;
  .container[data-is-selected] > & {
    @include background-accent-primary;
    display: block;
  }
  &[data-is-hovered],
  &[data-is-focused] {
    display: block;
    background: $theme-ui-primary-background;
  }
}
</style>
