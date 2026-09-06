<script setup>
import { DropdownMenuItem } from 'reka-ui'
import { cn } from '@/lib/utils'

const props = defineProps({
  disabled: { type: Boolean, default: false },
  class: { type: null, default: '' },
})
</script>

<template>
  <DropdownMenuItem
    :disabled="props.disabled"
    :class="cn('relative flex min-h-9 cursor-pointer items-center gap-2 rounded-md px-2.5 py-2 text-sm outline-none transition-colors select-none focus:bg-accent focus:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50 [&_svg]:size-4 [&_svg]:shrink-0', props.class)"
  >
    <slot />
  </DropdownMenuItem>
</template>
