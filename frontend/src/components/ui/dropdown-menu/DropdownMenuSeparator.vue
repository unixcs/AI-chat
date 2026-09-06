<script setup>
import { DropdownMenuSeparator } from 'reka-ui'
import { cn } from '@/lib/utils'
const props = defineProps({ class: { type: null, default: '' } })
</script>
<template>
  <DropdownMenuSeparator :class="cn('-mx-1.5 my-1.5 h-px bg-border', props.class)" />
</template>
