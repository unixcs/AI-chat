<script setup>
import { cn } from '@/lib/utils'
const props = defineProps({ class: { type: null, default: '' } })
</script>
<template>
  <h5 :class="cn('mb-1 font-semibold leading-none', props.class)"><slot /></h5>
</template>
