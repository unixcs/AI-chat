<script setup>
import { DropdownMenuLabel } from 'reka-ui'
import { cn } from '@/lib/utils'
const props = defineProps({ class: { type: null, default: '' } })
</script>
<template>
  <DropdownMenuLabel :class="cn('px-2.5 py-1.5 text-xs font-semibold text-muted-foreground', props.class)"><slot /></DropdownMenuLabel>
</template>
