<script setup>
import { AlertDialogTitle } from 'reka-ui'
import { cn } from '@/lib/utils'
const props = defineProps({ class: { type: null, default: '' } })
</script>
<template>
  <AlertDialogTitle :class="cn('text-lg font-semibold', props.class)"><slot /></AlertDialogTitle>
</template>
