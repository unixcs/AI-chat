<script setup>
import { cn } from '@/lib/utils'
const props = defineProps({ class: { type: null, default: '' } })
</script>
<template>
  <th :class="cn('h-11 px-3 text-left align-middle text-xs font-bold tracking-wider text-muted-foreground uppercase whitespace-nowrap', props.class)"><slot /></th>
</template>
