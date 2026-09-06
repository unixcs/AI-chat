<script setup>
import { TooltipPortal, TooltipContent } from 'reka-ui'
import { cn } from '@/lib/utils'
const props = defineProps({
  side: { type: String, default: 'top' },
  sideOffset: { type: Number, default: 6 },
  class: { type: null, default: '' },
})
</script>
<template>
  <TooltipPortal>
    <TooltipContent :side="props.side" :side-offset="props.sideOffset" :class="cn('z-50 rounded-md bg-foreground px-3 py-1.5 text-xs text-background shadow-md data-[state=delayed-open]:animate-in data-[state=delayed-open]:fade-in-0 data-[state=delayed-open]:zoom-in-95', props.class)">
      <slot />
    </TooltipContent>
  </TooltipPortal>
</template>
