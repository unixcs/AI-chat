<script setup>
import { SelectItem, SelectItemText, SelectItemIndicator } from 'reka-ui'
import { Check } from 'lucide-vue-next'
import { cn } from '@/lib/utils'

const props = defineProps({
  value: { type: null, required: true },
  disabled: { type: Boolean, default: false },
  class: { type: null, default: '' },
})
</script>

<template>
  <SelectItem
    :value="props.value"
    :disabled="props.disabled"
    :class="cn('relative flex w-full cursor-pointer items-center rounded-md py-2 pr-8 pl-2.5 text-sm outline-none select-none focus:bg-accent focus:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50', props.class)"
  >
    <span class="absolute right-2 flex h-4 w-4 items-center justify-center">
      <SelectItemIndicator>
        <Check class="size-4 text-primary" />
      </SelectItemIndicator>
    </span>
    <SelectItemText><slot /></SelectItemText>
  </SelectItem>
</template>
