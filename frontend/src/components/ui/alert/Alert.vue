<script setup>
import { cva } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const props = defineProps({
  variant: { type: String, default: 'default' },
  class: { type: null, default: '' },
})

const alertVariants = cva(
  'relative w-full rounded-lg border px-4 py-3 text-sm [&>svg]:absolute [&>svg]:top-3.5 [&>svg]:left-4 [&>svg]:size-4 [&>svg+div]:translate-y-[-2px] [&>svg~*]:pl-7',
  {
    variants: {
      variant: {
        default: 'bg-card text-card-foreground border-border',
        destructive: 'border-destructive/30 bg-destructive/10 text-destructive',
        warning: 'border-warning/30 bg-warning/10 text-warning',
        success: 'border-primary/30 bg-accent text-accent-foreground',
      },
    },
    defaultVariants: { variant: 'default' },
  },
)
</script>

<template>
  <div role="alert" :class="cn(alertVariants({ variant }), props.class)">
    <slot />
  </div>
</template>
