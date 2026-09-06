<script setup>
import { AlertDialogAction } from 'reka-ui'
import { cn } from '@/lib/utils'
const props = defineProps({ class: { type: null, default: '' } })
</script>
<template>
  <AlertDialogAction :class="cn('inline-flex h-11 items-center justify-center rounded-lg bg-primary px-5 text-sm font-semibold text-primary-foreground transition-colors hover:bg-primary-hover cursor-pointer', props.class)">
    <slot />
  </AlertDialogAction>
</template>
