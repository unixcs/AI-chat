<script setup>
import { SelectTrigger, SelectIcon } from 'reka-ui'
import { ChevronDown } from 'lucide-vue-next'
import { cn } from '@/lib/utils'

const props = defineProps({ class: { type: null, default: '' } })
</script>

<template>
  <SelectTrigger
    :class="cn('flex h-11 w-full items-center justify-between gap-2 rounded-md border border-input bg-card/70 px-3.5 py-2 text-[15px] text-foreground shadow-sm transition-colors duration-150 placeholder:text-faint focus-visible:border-primary/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-50 [&>span]:truncate', props.class)"
  >
    <slot />
    <SelectIcon as-child>
      <ChevronDown class="size-4 shrink-0 text-muted-foreground" />
    </SelectIcon>
  </SelectTrigger>
</template>
