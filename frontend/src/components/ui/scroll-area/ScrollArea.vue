<script setup>
import { ScrollAreaRoot, ScrollAreaViewport, ScrollAreaScrollbar, ScrollAreaThumb, ScrollAreaCorner } from 'reka-ui'
import { cn } from '@/lib/utils'

const props = defineProps({
  orientation: { type: String, default: 'vertical' },
  class: { type: null, default: '' },
})
</script>

<template>
  <ScrollAreaRoot :class="cn('relative overflow-hidden', props.class)">
    <ScrollAreaViewport class="h-full w-full rounded-[inherit] outline-none">
      <slot />
    </ScrollAreaViewport>
    <ScrollAreaScrollbar
      :orientation="props.orientation"
      class="flex touch-none select-none p-0.5 transition-colors data-[orientation=vertical]:w-2.5 data-[orientation=horizontal]:h-2.5"
    >
      <ScrollAreaThumb class="relative flex-1 rounded-full bg-border hover:bg-border/80" />
    </ScrollAreaScrollbar>
    <ScrollAreaCorner />
  </ScrollAreaRoot>
</template>
