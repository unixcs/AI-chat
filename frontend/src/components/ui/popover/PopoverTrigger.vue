<script setup>
import { PopoverTrigger } from 'reka-ui'
const props = defineProps({ asChild: { type: Boolean, default: false } })
</script>
<template>
  <PopoverTrigger :as-child="props.asChild"><slot /></PopoverTrigger>
</template>
