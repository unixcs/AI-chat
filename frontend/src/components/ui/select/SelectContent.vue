<script setup>
import { SelectPortal, SelectContent, SelectViewport, SelectScrollUpButton, SelectScrollDownButton } from 'reka-ui'
import { ChevronDown, ChevronUp } from 'lucide-vue-next'
import { cn } from '@/lib/utils'

const props = defineProps({
  position: { type: String, default: 'popper' },
  sideOffset: { type: Number, default: 6 },
  class: { type: null, default: '' },
})
</script>

<template>
  <SelectPortal>
    <SelectContent
      :position="props.position"
      :side-offset="props.sideOffset"
      :class="cn('z-50 max-h-96 min-w-[var(--reka-select-trigger-width)] overflow-hidden rounded-lg border border-border bg-popover text-popover-foreground shadow-[var(--shadow-overlay)] data-[state=open]:animate-in data-[state=open]:fade-in-0 data-[state=open]:zoom-in-95 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95', props.position === 'popper' && 'data-[side=bottom]:translate-y-1 data-[side=top]:-translate-y-1', props.class)"
    >
      <SelectScrollUpButton class="flex h-6 items-center justify-center text-muted-foreground">
        <ChevronUp class="size-4" />
      </SelectScrollUpButton>
      <SelectViewport :class="cn('p-1.5', props.position === 'popper' && 'h-[var(--reka-select-trigger-height)] w-full min-w-[var(--reka-select-trigger-width)]')">
        <slot />
      </SelectViewport>
      <SelectScrollDownButton class="flex h-6 items-center justify-center text-muted-foreground">
        <ChevronDown class="size-4" />
      </SelectScrollDownButton>
    </SelectContent>
  </SelectPortal>
</template>
