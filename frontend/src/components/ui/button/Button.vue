<script setup>
import { Primitive } from 'reka-ui'
import { cva } from 'class-variance-authority'
import { cn } from '@/lib/utils'

const props = defineProps({
  variant: { type: String, default: 'default' },
  size: { type: String, default: 'default' },
  class: { type: null, default: '' },
  asChild: { type: Boolean, default: false },
  as: { type: null, default: 'button' },
})

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-lg text-sm font-semibold transition-all duration-150 disabled:pointer-events-none disabled:opacity-50 outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-1 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0 cursor-pointer",
  {
    variants: {
      variant: {
        default: 'bg-primary text-primary-foreground shadow-sm hover:bg-primary-hover',
        destructive: 'bg-destructive text-destructive-foreground shadow-sm hover:opacity-90',
        outline: 'border border-input bg-card/60 hover:bg-accent hover:text-accent-foreground',
        secondary: 'bg-secondary text-secondary-foreground hover:opacity-90',
        ghost: 'hover:bg-accent hover:text-accent-foreground',
        link: 'text-primary underline-offset-4 hover:underline',
      },
      size: {
        default: 'h-11 px-5',
        sm: 'h-9 px-3.5 text-[13px]',
        lg: 'h-12 px-7 text-base',
        icon: 'h-11 w-11',
        iconSm: 'h-9 w-9',
      },
    },
    defaultVariants: { variant: 'default', size: 'default' },
  },
)
</script>

<template>
  <Primitive
    :as="as"
    :as-child="asChild"
    :class="cn(buttonVariants({ variant, size }), props.class)"
  >
    <slot />
  </Primitive>
</template>
