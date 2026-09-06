<script setup>
import { TabsContent } from 'reka-ui'
import { cn } from '@/lib/utils'
const props = defineProps({ value: { type: null, required: true }, class: { type: null, default: '' } })
</script>
<template>
  <TabsContent :value="props.value" :class="cn('mt-3 focus-visible:outline-none', props.class)"><slot /></TabsContent>
</template>
