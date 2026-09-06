<script setup>
import { Label } from 'reka-ui'
import { cn } from '@/lib/utils'

const props = defineProps({
  for: { type: String, default: undefined },
  class: { type: null, default: '' },
})
</script>

<template>
  <Label
    :for="props.for"
    :class="cn('block text-[13px] font-semibold text-muted-foreground', props.class)"
  >
    <slot />
  </Label>
</template>
