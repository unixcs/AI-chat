<script setup>
import { PopoverPortal, PopoverContent } from 'reka-ui'
import { cn } from '@/lib/utils'

const props = defineProps({
  side: { type: String, default: 'bottom' },
  sideOffset: { type: Number, default: 8 },
  align: { type: String, default: 'center' },
  alignOffset: { type: Number, default: undefined },
  class: { type: null, default: '' },
})
</script>

<template>
  <PopoverPortal>
    <PopoverContent
      :side="props.side"
      :side-offset="props.sideOffset"
      :align="props.align"
      :align-offset="props.alignOffset"
      :class="cn('z-50 max-h-[min(70vh,480px)] w-72 origin-(--reka-popover-content-transform-origin) overflow-y-auto rounded-xl border border-border bg-popover p-4 text-popover-foreground shadow-[var(--shadow-overlay)] outline-none data-[state=open]:animate-in data-[state=open]:fade-in-0 data-[state=open]:zoom-in-95 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95', props.class)"
    >
      <slot />
    </PopoverContent>
  </PopoverPortal>
</template>
