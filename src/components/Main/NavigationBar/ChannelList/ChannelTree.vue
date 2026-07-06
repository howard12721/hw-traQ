<template>
  <div>
    <template v-for="channel in channels" :key="channel.id">
      <template v-if="channel.children.length > 0">
        <ChannelFolderElement
          :class="$style.element"
          :channel="channel"
          :is-opened="childrenShownChannels.has(channel.id)"
          :show-shortened-path="showShortenedPath"
          @click="toggleChildren"
        />
        <SlideDown :is-open="childrenShownChannels.has(channel.id)">
          <div :class="$style.children">
            <ChannelElement
              :class="$style.element"
              :channel="rootChannel(channel)"
              :show-topic="showTopic"
            />
            <channel-tree
              :channels="channel.children"
              :show-topic="showTopic && !preventChildTopic"
            />
          </div>
        </SlideDown>
      </template>
      <ChannelElement
        v-else
        :class="$style.element"
        :channel="channel"
        :show-shortened-path="showShortenedPath"
        :show-topic="showTopic"
      />
    </template>
  </div>
</template>

<script lang="ts" setup>
import { type HTMLAttributes, ref } from 'vue'

import SlideDown from '/@/components/UI/SlideDown.vue'
import type { ChannelTreeNode } from '/@/lib/channelTree'
import type { ChannelId } from '/@/types/entity-ids'

import ChannelElement from './ChannelElement.vue'
import ChannelFolderElement from './ChannelFolderElement.vue'

interface Props extends /* @vue-ignore */ HTMLAttributes {
  channels: ReadonlyArray<ChannelTreeNode>
  showShortenedPath?: boolean
  showTopic?: boolean
  preventChildTopic?: boolean
}

withDefaults(defineProps<Props>(), {
  showShortenedPath: false,
  showTopic: false,
  preventChildTopic: false
})

const childrenShownChannels = ref(new Set<ChannelId>())
const toggleChildren = (channelId: ChannelId) => {
  if (childrenShownChannels.value.has(channelId)) {
    childrenShownChannels.value.delete(channelId)
  } else {
    childrenShownChannels.value.add(channelId)
  }
}

const rootChannel = (channel: ChannelTreeNode): ChannelTreeNode => ({
  ...channel,
  name: '(root)',
  children: [],
  skippedAncestorNames: undefined
})
</script>

<style lang="scss" module>
.element {
  margin: 4px 0;
}

$childrenIndent: 12px;
$childrenContentIndent: 16px;
$connectorLeft: 11px;
$connectorWidth: 2px;

.children {
  position: relative;
  margin-left: $childrenIndent;
  padding-left: $childrenContentIndent;
  &::before {
    content: '';
    position: absolute;
    top: 0;
    bottom: 0;
    left: $connectorLeft;
    width: $connectorWidth;
    background: $theme-ui-primary-background;
  }
}
</style>
