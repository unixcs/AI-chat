<script setup>
import { cn } from '@/lib/utils'
const props = defineProps({ class: { type: null, default: '' } })
</script>
<template>
  <td :class="cn('px-3 py-3 align-middle whitespace-nowrap', props.class)"><slot /></td>
</template>
