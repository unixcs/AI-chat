<script setup>
import { PopoverRoot } from 'reka-ui'
const props = defineProps({ open: { type: Boolean, default: undefined }, defaultOpen: { type: Boolean, default: false } })
const emit = defineEmits(['update:open'])
</script>
<template>
  <PopoverRoot :open="props.open" :default-open="props.defaultOpen" @update:open="emit('update:open', $event)">
    <slot />
  </PopoverRoot>
</template>
