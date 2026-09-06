<script setup>
import { cn } from '@/lib/utils'
const props = defineProps({ class: { type: null, default: '' } })
</script>
<template>
  <h3 :class="cn('text-lg leading-tight font-semibold text-card-foreground', props.class)"><slot /></h3>
</template>
