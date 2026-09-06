<script setup>
import { Separator } from 'reka-ui'
import { cn } from '@/lib/utils'

const props = defineProps({
  orientation: { type: String, default: 'horizontal' },
  decorative: { type: Boolean, default: true },
  class: { type: null, default: '' },
})
</script>

<template>
  <Separator
    :orientation="props.orientation"
    :decorative="props.decorative"
    :class="cn('shrink-0 bg-border', props.orientation === 'vertical' ? 'w-px h-full' : 'h-px w-full', props.class)"
  />
</template>
