<script setup>
import { cn } from '@/lib/utils'

const props = defineProps({ class: { type: null, default: '' } })
const model = defineModel({ type: String, default: '' })
</script>

<template>
  <textarea
    v-model="model"
    :class="cn('flex min-h-20 w-full rounded-md border border-input bg-card/70 px-3.5 py-2.5 text-[15px] text-foreground shadow-sm transition-colors duration-150 placeholder:text-faint focus-visible:border-primary/50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-50', props.class)"
  />
</template>
