<script setup>
import { TabsRoot } from 'reka-ui'
const props = defineProps({
  modelValue: { type: null, default: undefined },
  defaultValue: { type: null, default: undefined },
})
const emit = defineEmits(['update:modelValue'])
</script>
<template>
  <TabsRoot :model-value="props.modelValue" :default-value="props.defaultValue" @update:model-value="emit('update:modelValue', $event)">
    <slot />
  </TabsRoot>
</template>
