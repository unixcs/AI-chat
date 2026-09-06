<script setup>
import { TabsTrigger } from 'reka-ui'
import { cn } from '@/lib/utils'
const props = defineProps({ value: { type: null, required: true }, class: { type: null, default: '' } })
</script>
<template>
  <TabsTrigger :value="props.value" :class="cn('inline-flex h-9 cursor-pointer items-center justify-center gap-1.5 rounded-md px-3.5 text-sm font-medium whitespace-nowrap transition-all duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 data-[state=active]:bg-card data-[state=active]:text-foreground data-[state=active]:shadow-sm', props.class)"><slot /></TabsTrigger>
</template>
