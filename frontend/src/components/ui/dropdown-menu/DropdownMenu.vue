<script setup>
import { DropdownMenuRoot } from 'reka-ui'
const props = defineProps({ defaultOpen: { type: Boolean, default: false }, open: { type: Boolean, default: undefined } })
const emit = defineEmits(['update:open'])
</script>
<template>
  <DropdownMenuRoot :open="props.open" :default-open="props.defaultOpen" @update:open="emit('update:open', $event)">
    <slot />
  </DropdownMenuRoot>
</template>
