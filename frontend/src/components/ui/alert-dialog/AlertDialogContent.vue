<script setup>
import { AlertDialogPortal, AlertDialogOverlay, AlertDialogContent } from 'reka-ui'
import { cn } from '@/lib/utils'

const props = defineProps({ class: { type: null, default: '' } })
</script>

<template>
  <AlertDialogPortal>
    <AlertDialogOverlay
      class="fixed inset-0 z-50 bg-[rgba(20,30,22,0.4)] backdrop-blur-[2px] transition-opacity duration-200 data-[state=open]:animate-in data-[state=open]:fade-in-0 data-[state=closed]:animate-out data-[state=closed]:fade-out-0"
    />
    <AlertDialogContent
      :class="cn('fixed top-1/2 left-1/2 z-50 grid w-[calc(100vw-32px)] max-w-md -translate-x-1/2 -translate-y-1/2 gap-4 rounded-xl border border-border bg-popover p-6 text-popover-foreground shadow-[var(--shadow-overlay)] duration-200 data-[state=open]:animate-in data-[state=open]:fade-in-0 data-[state=open]:zoom-in-95 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95 sm:w-full', props.class)"
    >
      <slot />
    </AlertDialogContent>
  </AlertDialogPortal>
</template>
