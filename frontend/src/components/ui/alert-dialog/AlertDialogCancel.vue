<script setup>
import { AlertDialogCancel } from 'reka-ui'
import { cn } from '@/lib/utils'
const props = defineProps({ class: { type: null, default: '' } })
</script>
<template>
  <AlertDialogCancel :class="cn('inline-flex h-11 items-center justify-center rounded-lg border border-input bg-card/60 px-5 text-sm font-semibold transition-colors hover:bg-accent cursor-pointer', props.class)">
    <slot />
  </AlertDialogCancel>
</template>
