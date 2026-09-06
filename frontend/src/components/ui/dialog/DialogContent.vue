<script setup>
import { DialogPortal, DialogOverlay, DialogContent, DialogClose } from 'reka-ui'
import { X } from 'lucide-vue-next'
import { cn } from '@/lib/utils'

const props = defineProps({ class: { type: null, default: '' } })
</script>

<template>
  <DialogPortal>
    <DialogOverlay
      class="fixed inset-0 z-50 bg-[rgba(20,30,22,0.4)] backdrop-blur-[2px] transition-opacity duration-200 data-[state=open]:animate-in data-[state=open]:fade-in-0 data-[state=closed]:animate-out data-[state=closed]:fade-out-0"
    />
    <DialogContent
      :class="cn('fixed top-1/2 left-1/2 z-50 grid max-h-[calc(100vh-32px)] w-[calc(100vw-32px)] max-w-lg -translate-x-1/2 -translate-y-1/2 gap-4 overflow-y-auto rounded-xl border border-border bg-popover p-6 text-popover-foreground shadow-[var(--shadow-overlay)] duration-200 data-[state=open]:animate-in data-[state=open]:fade-in-0 data-[state=open]:zoom-in-95 data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=closed]:zoom-out-95 sm:w-full', props.class)"
    >
      <slot />
      <DialogClose
        class="absolute top-3.5 right-3.5 rounded-md p-1.5 text-muted-foreground opacity-70 transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring"
      >
        <X class="size-4" />
        <span class="sr-only">关闭</span>
      </DialogClose>
    </DialogContent>
  </DialogPortal>
</template>
