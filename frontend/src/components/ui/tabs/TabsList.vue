<script setup>
import { TabsList } from 'reka-ui'
import { cn } from '@/lib/utils'
const props = defineProps({ class: { type: null, default: '' } })
</script>
<template>
  <TabsList :class="cn('inline-flex h-11 items-center justify-center gap-1 rounded-lg bg-muted p-1 text-muted-foreground', props.class)"><slot /></TabsList>
</template>
