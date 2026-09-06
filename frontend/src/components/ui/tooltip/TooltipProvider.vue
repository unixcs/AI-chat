<script setup>
import { TooltipProvider } from 'reka-ui'
const props = defineProps({ delayDuration: { type: Number, default: 200 } })
</script>
<template>
  <TooltipProvider :delay-duration="props.delayDuration"><slot /></TooltipProvider>
</template>
