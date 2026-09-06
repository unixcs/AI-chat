<script setup>
import { AlertDialogRoot } from 'reka-ui'
const props = defineProps({ open: { type: Boolean, default: undefined }, defaultOpen: { type: Boolean, default: false } })
const emit = defineEmits(['update:open'])
</script>
<template>
  <AlertDialogRoot :open="props.open" :default-open="props.defaultOpen" @update:open="emit('update:open', $event)">
    <slot />
  </AlertDialogRoot>
</template>
