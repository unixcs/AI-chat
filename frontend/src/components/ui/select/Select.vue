<script setup>
import { SelectRoot } from 'reka-ui'
const props = defineProps({
  modelValue: { type: null, default: undefined },
  defaultValue: { type: null, default: undefined },
  placeholder: { type: String, default: '' },
  disabled: { type: Boolean, default: false },
})
const emit = defineEmits(['update:modelValue'])
</script>
<template>
  <SelectRoot
    :model-value="props.modelValue"
    :default-value="props.defaultValue"
    :placeholder="props.placeholder"
    :disabled="props.disabled"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <slot />
  </SelectRoot>
</template>
