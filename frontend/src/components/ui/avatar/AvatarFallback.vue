<script setup>
import { AvatarFallback } from 'reka-ui'
import { cn } from '@/lib/utils'
const props = defineProps({ class: { type: null, default: '' } })
</script>
<template>
  <AvatarFallback :class="cn('flex h-full w-full items-center justify-center rounded-full bg-accent text-sm font-semibold text-accent-foreground', props.class)">
    <slot />
  </AvatarFallback>
</template>
