<script setup>
import { cva } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const props = defineProps({
  variant: { type: String, default: 'default' },
  class: { type: null, default: '' },
})

const badgeVariants = cva(
  'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold transition-colors',
  {
    variants: {
      variant: {
        default: 'bg-accent text-accent-foreground',
        secondary: 'bg-secondary text-secondary-foreground',
        destructive: 'bg-destructive/10 text-destructive',
        outline: 'border border-input text-foreground',
      },
    },
    defaultVariants: { variant: 'default' },
  },
)
</script>

<template>
  <span :class="cn(badgeVariants({ variant }), props.class)">
    <slot />
  </span>
</template>
