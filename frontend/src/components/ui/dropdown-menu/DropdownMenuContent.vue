<script setup>
import { DropdownMenuPortal, DropdownMenuContent } from 'reka-ui'
import { cn } from '@/lib/utils'

const props = defineProps({
  sideOffset: { type: Number, default: 6 },
  align: { type: String, default: 'end' },
  class: { type: null, default: '' },
})
</script>

<template>
  <DropdownMenuPortal>
    <DropdownMenuContent
      :side-offset="props.sideOffset"
      :align="props.align"
      :class="cn('z-50 min-w-40 overflow-hidden rounded-lg border border-border bg-popover p-1.5 text-popover-foreground shadow-[var(--shadow-overlay)] data-[state=open]:animate-in data-[state=open]:fade-in-0 data-[state=open]:zoom-in-95 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95', props.class)"
    >
      <slot />
    </DropdownMenuContent>
  </DropdownMenuPortal>
</template>
