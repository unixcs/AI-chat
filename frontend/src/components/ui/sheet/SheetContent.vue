<script setup>
import { DialogPortal, DialogOverlay, DialogContent, DialogClose } from 'reka-ui'
import { X } from 'lucide-vue-next'
import { cva } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const props = defineProps({
  side: { type: String, default: 'right' },
  class: { type: null, default: '' },
})

const sheetVariants = cva(
  'fixed z-50 flex flex-col gap-4 border border-border bg-popover text-popover-foreground shadow-[var(--shadow-overlay)] transition ease-in-out duration-200 data-[state=open]:animate-in data-[state=closed]:animate-out',
  {
    variants: {
      side: {
        top: 'inset-x-0 top-0 border-b data-[state=open]:slide-in-from-top data-[state=closed]:slide-out-to-top',
        bottom: 'inset-x-0 bottom-0 border-t data-[state=open]:slide-in-from-bottom data-[state=closed]:slide-out-to-bottom',
        left: 'inset-y-0 left-0 h-full w-4/5 max-w-sm border-r data-[state=open]:slide-in-from-left data-[state=closed]:slide-out-to-left',
        right: 'inset-y-0 right-0 h-full w-4/5 max-w-sm border-l data-[state=open]:slide-in-from-right data-[state=closed]:slide-out-to-right',
      },
    },
    defaultVariants: { side: 'right' },
  },
)
</script>

<template>
  <DialogPortal>
    <DialogOverlay
      class="fixed inset-0 z-50 bg-[rgba(20,30,22,0.4)] backdrop-blur-[2px] transition-opacity duration-200 data-[state=open]:animate-in data-[state=open]:fade-in-0 data-[state=closed]:animate-out data-[state=closed]:fade-out-0"
    />
    <DialogContent :class="cn(sheetVariants({ side: props.side }), props.class)">
      <slot />
      <DialogClose class="absolute top-3.5 right-3.5 rounded-md p-1.5 text-muted-foreground opacity-70 transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring">
        <X class="size-4" />
        <span class="sr-only">关闭</span>
      </DialogClose>
    </DialogContent>
  </DialogPortal>
</template>
