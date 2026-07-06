<template>
  <div :class="$style.container">
    <AIcon mdi name="comment-off-outline" :size="20" :class="$style.icon" />
    <span :class="$style.text">ミュートしたメッセージです</span>
    <button
      :class="$style.button"
      type="button"
      @click.prevent.stop="emit('reveal')"
    >
      表示
    </button>
  </div>
</template>

<script lang="ts" setup>
import AIcon from '/@/components/UI/AIcon.vue'

const emit = defineEmits<{
  (e: 'reveal'): void
}>()
</script>

<style lang="scss" module>
.container {
  @include color-ui-secondary;
  @include background-secondary;
  @include size-body2;
  display: flex;
  align-items: center;
  gap: 8px;
  width: fit-content;
  max-width: 100%;
  min-height: 32px;
  padding: 4px 8px;
  border-radius: 4px;
}

.icon {
  flex-shrink: 0;
}

.text {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.button {
  @include color-accent-primary;
  flex-shrink: 0;
  padding: 2px 6px;
  border-radius: 4px;
  font-weight: bold;
  cursor: pointer;

  &:hover,
  &:focus-visible {
    @include background-tertiary;
  }
}
</style>
