<script setup>
import { Toaster as Sonner } from 'vue-sonner'

const props = defineProps({
  position: { type: String, default: 'top-center' },
  duration: { type: Number, default: 2200 },
})
</script>

<template>
  <Sonner
    class="toaster group"
    :position="props.position"
    :duration="props.duration"
    v-bind="$attrs"
    :style="{
      '--normal-bg': 'var(--popover)',
      '--normal-text': 'var(--popover-foreground)',
      '--normal-border': 'var(--border)',
    }"
  />
</template>
