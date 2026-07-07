<template>
  <div :class="$style.container">
    <NavigationSelectorItem
      v-for="item in entries"
      :key="item.type"
      :class="$style.item"
      :is-selected="currentNavigation === item.type"
      :has-notification="item.hasNotification"
      :icon-mdi="item.iconMdi"
      :icon-name="item.iconName"
      :title="navigationLabel(item)"
      @click="onNavigationItemClick(item.type)"
    />
    <div v-if="showSeparator" :class="$style.separator" />
    <NavigationSelectorItem
      v-for="item in ephemeralEntries"
      :key="item.type"
      :class="$style.item"
      :is-selected="currentEphemeralNavigation === item.type"
      :icon-mdi="item.iconMdi"
      :icon-name="item.iconName"
      :color-claim="item.colorClaim"
      :title="ephemeralNavigationLabel(item)"
      @click="onEphemeralNavigationItemClick(item.type)"
    />
  </div>
</template>

<script lang="ts" setup>
import { computed, watch } from 'vue'

import NavigationSelectorItem from '/@/components/Main/NavigationBar/NavigationSelectorItem.vue'
import type {
  EphemeralNavigationItemType,
  NavigationItemType
} from '/@/components/Main/NavigationBar/composables/useNavigationConstructor'
import {
  ephemeralNavigationTypeNameMap,
  navigationTypeNameMap,
  useEphemeralNavigationSelectorItem,
  useNavigationSelectorItem
} from '/@/components/Main/NavigationBar/composables/useNavigationConstructor'

import type {
  EphemeralNavigationSelectorEntry,
  NavigationSelectorEntry
} from './composables/useNavigationSelectorEntry'
import useNavigationSelectorEntry from './composables/useNavigationSelectorEntry'

withDefaults(
  defineProps<{
    currentNavigation?: NavigationItemType
    currentEphemeralNavigation?: EphemeralNavigationItemType
  }>(),
  {
    currentNavigation: 'home' as const
  }
)

const emit = defineEmits<{
  (e: 'navigationChange', _type: NavigationItemType): void
  (e: 'ephemeralNavigationChange', _type: EphemeralNavigationItemType): void
  (e: 'ephemeralEntryRemove', _entry: EphemeralNavigationSelectorEntry): void
  (e: 'ephemeralEntryAdd', _entry: EphemeralNavigationSelectorEntry): void
}>()

const { onNavigationItemClick } = useNavigationSelectorItem(emit)
const { onNavigationItemClick: onEphemeralNavigationItemClick } =
  useEphemeralNavigationSelectorItem(emit)
const { entries, ephemeralEntries } = useNavigationSelectorEntry()
const showSeparator = computed(() => ephemeralEntries.value.length > 0)
const navigationLabel = (item: NavigationSelectorEntry) =>
  navigationTypeNameMap[item.type]
const ephemeralNavigationLabel = (item: EphemeralNavigationSelectorEntry) =>
  item.type ? ephemeralNavigationTypeNameMap[item.type] : undefined

watch(ephemeralEntries, (entries, prevEntries) => {
  prevEntries
    ?.filter(e => !entries.includes(e))
    .forEach(e => {
      emit('ephemeralEntryRemove', e)
    })
  entries
    ?.filter(e => !prevEntries?.includes(e))
    .forEach(e => {
      emit('ephemeralEntryAdd', e)
    })
})
</script>

<style lang="scss" module>
.container {
  @include color-ui-primary;
  @include background-secondary;
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 0 12px;
  overflow-x: auto;
  overflow-y: hidden;
  scrollbar-width: none;
  -webkit-overflow-scrolling: touch;

  &::-webkit-scrollbar {
    display: none;
  }
}
.item {
  flex: 0 0 auto;
  margin: 10px 0;
}
.separator {
  @include background-tertiary;
  flex: 0 0 1px;
  align-self: stretch;
  margin: 10px 4px;
}
</style>
